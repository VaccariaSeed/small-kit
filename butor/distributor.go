package butor

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"
	"sync/atomic"
)

const (
	defaultSize = 10
)

// 创建一个cli
func newCli(bufferSize byte, sche Scheduler) *cli {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &cli{dc: make(chan []any, bufferSize), sche: sche, ctx: ctx, cancel: cancelFunc}
}

type cli struct {
	dc     chan []any
	sche   Scheduler
	ctx    context.Context
	cancel context.CancelFunc
	*Register
}

// 通知
func (c *cli) notice(bucketName, point string, value any, isChanged bool) {
	c.dc <- []any{bucketName, point, value, isChanged}
}

// 监听
func (c *cli) monitor() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case data := <-c.dc:
			if data[3].(bool) {
				c.sche.Switches(data[0].(string), data[1].(string), data[2])
			} else {
				c.sche.Received(data[0].(string), data[1].(string), data[2])
			}
		}
	}
}

// NewDistributor 创建一个数据分发器
func NewDistributor(bufferSize byte) *Distributor {
	if bufferSize == 0 {
		bufferSize = defaultSize
	}
	return &Distributor{idGenerator: &idGenerator{}, valuesSub: make(map[string]map[uint64][]string), switches: make(map[string]map[string]*dataSwitch), bufferSize: bufferSize, transit: newTransit()}
}

// Distributor 数据分发器
type Distributor struct {
	*idGenerator                                   //id生成器
	*transit                                       //订阅中转
	buckets      sync.Map                          //数据桶
	valuesSub    map[string]map[uint64][]string    // map[桶名]map[调度员id][]订阅字段
	switches     map[string]map[string]*dataSwitch //变值。一般这种值都比较重要，创建时不与桶强绑定
	sches        sync.Map                          //调度员
	bufferSize   byte
	isClosed     atomic.Bool
}

//关闭
func (d *Distributor) Close() {
	d.lock.Lock()
	defer d.lock.Unlock()
	// 遍历所有键值对
	d.sches.Range(func(key, value interface{}) bool {
		d.closeCli(key.(uint64))
		value.(*cli).sche.ButorClosed()
		return true // 返回 true 继续遍历，false 则停止
	})
	clear(d.valuesSub)
	clear(d.switches)
	d.isClosed.Store(true)
}

// 取消订阅某些变更值
func (d *Distributor) unsubscribeSwitch(id uint64, bucketName string, keys ...string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.switches[bucketName]; !ok {
		return
	}
	for _, key := range keys {
		if ds, ok := d.switches[bucketName][key]; ok {
			ds.removeCli(id)
		}
	}
}

// 取消订阅某些值
func (d *Distributor) unsubscribeValue(id uint64, bucketName string, key ...string) {
	if bucket, ok := d.buckets.Load(bucketName); ok {
		//存在桶
		bucket.(*dataBucket).unsubscribeValue(key, id)
	}
}

// 取消订阅
func (d *Distributor) unsubscribe(id uint64, bucketName ...string) {
	for _, name := range bucketName {
		if bucket, ok := d.buckets.Load(name); ok {
			bucket.(*dataBucket).unsubscribe(id)
		}
	}
}

// 关闭某个客户端
func (d *Distributor) closeCli(id uint64) {
	if cl, ok := d.sches.Load(id); ok {
		snap := cl.(*cli)
		snap.cancel()
		close(snap.dc)
		//关闭相关桶
		d.closeBucket(slices.Collect(maps.Keys(snap.creater))...)
		d.sches.Delete(id)
	}
}

// 关闭桶
func (d *Distributor) closeBucket(buckets ...string) {
	for _, bucketName := range buckets {
		if bucket, ok := d.buckets.Load(bucketName); ok {
			//存在桶
			ids := bucket.(*dataBucket).allMark()
			for _, id := range ids {
				if cl, hasCli := d.sches.Load(id); hasCli {
					//存在对应客户端
					cl.(*cli).sche.SourceCloed(bucketName)
				}
			}
			d.buckets.Delete(bucketName)
		}
	}
}

// 发送一个数据
func (d *Distributor) release(bucketName string, name string, value any) error {
	if d.isClosed.Load() == true {
		return errors.New("distributor closed")
	}
	//处理变化值
	if d.switches[bucketName] != nil && d.switches[bucketName][name] != nil {
		if sches, ok := d.switches[bucketName][name].flush(value); ok {
			for _, id := range sches {
				if sche, ok := d.sches.Load(id); ok {
					sche.(*cli).notice(bucketName, name, value, true)
				}
			}
		}
	}
	//入值
	if bucket, ok := d.buckets.Load(bucketName); !ok {
		return NotFoundBucket
	} else {
		sches := bucket.(*dataBucket).flush(name, value)
		for _, id := range sches {
			if sche, ok := d.sches.Load(id); ok {
				sche.(*cli).notice(bucketName, name, value, false)
			}
		}
	}
	return nil
}

// Register 注册一个调度员
func (d *Distributor) Register(sche Scheduler) error {
	if sche == nil {
		return errors.New("sche is empty")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	reg := newRegister()
	sche.Register(reg)
	id := d.nextID()
	//创建某些桶，bool表示是否保存最新值快照
	//订阅某个桶的某些值
	//订阅某些桶的所有数据
	//只有当某个桶的某些值发生变更或者第一次加入时就会发起通知
	_ = d.createBucket(reg.creater).subValues(reg.subscribe, id).subBucket(reg.subBucket, id).subSwitch(reg.switches, id)
	sche.Syntony(d.newOperator(id, reg.creater))
	cliSnap := newCli(d.bufferSize, sche)
	cliSnap.Register = reg
	d.sches.Store(id, cliSnap)
	go cliSnap.monitor()
	return nil
}

// 创建一个操作器
func (d *Distributor) newOperator(id uint64, creater map[string]bool) *Operator {
	return &Operator{release: d.release, Id: id, close: d.closeCli, creater: creater, unsubscribe: d.unsubscribe, unsubscribeValue: d.unsubscribeValue, unsubscribeSwitch: d.unsubscribeSwitch, closeBucket: d.closeBucket, obtain: d.obtain}
}

// 获取某个值的快照值
func (d *Distributor) obtain(bucketName, key string) (any, error) {
	if bucket, ok := d.buckets.Load(bucketName); ok {
		return bucket.(*dataBucket).obtain(key)
	}
	return nil, NotFoundBucket
}

// 订阅一个变化值
func (d *Distributor) subSwitch(switches map[string][]string, id uint64) *Distributor {
	for bucketName, values := range switches {
		if _, ok := d.switches[bucketName]; !ok {
			d.switches[bucketName] = make(map[string]*dataSwitch)
		}
		for _, value := range values {
			if _, ok := d.switches[bucketName][value]; !ok {
				d.switches[bucketName][value] = &dataSwitch{isFirst: true}
			}
			d.switches[bucketName][value].subSwitch(id)
		}
	}
	return d
}

func (d *Distributor) subBucket(buckets []string, id uint64) *Distributor {
	for _, bucketName := range buckets {
		if bucket, ok := d.buckets.Load(bucketName); ok {
			//桶存在
			bucket.(*dataBucket).subBucket(id)
		} else {
			//桶不存在
			d.bucketTransit(bucketName, id)
		}
	}
	return d
}

// 订阅某些桶的某些值
func (d *Distributor) subValues(subscribe map[string][]string, id uint64) *Distributor {
	for bucketName, values := range subscribe {
		if bucket, ok := d.buckets.Load(bucketName); ok {
			bucket.(*dataBucket).subValues(values, id)
		} else {
			//不存在这个桶，加入到快照之中，创建时绑定
			d.subTransit(bucketName, values, id)
		}
	}
	return d
}

// 创建桶
func (d *Distributor) createBucket(creater map[string]bool) *Distributor {
	for bucketName, snap := range creater {
		if bucket, ok := d.buckets.Load(bucketName); !ok {
			//桶不存在
			d.buckets.Store(bucketName, newDataBucket(snap, d.subs[bucketName], d.subBuckets[bucketName]))
		} else {
			//桶存在
			bucket.(*dataBucket).snap = snap
		}
	}
	return d
}
