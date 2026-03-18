package small_kit

import (
	"errors"
	"fmt"
	"slices"
	"sync"
)

// ReleaseFunc 发布一个数据
// bucket 数据桶
// name 数据名
// value 数据
type ReleaseFunc func(bucket string, name string, value any)

type DataRegister struct {
	buckets     map[string][]string
	registerAll bool
}

// RegisterAll 订阅所有
func (d *DataRegister) RegisterAll() {
	d.registerAll = true
}

// Append 添加关注的信息
// bucket 数据桶
// name 数据名
func (d *DataRegister) Append(bucket string, name ...string) {
	d.buckets[bucket] = append(d.buckets[bucket], name...)
}

type client struct {
	dc chan []any
	as DataAssociator
}

func (c *client) flush() {
	for data := range c.dc {
		c.as.Received(data[0].(string), data[1].(string), data[2])
	}
}

// DataAssociator 客户机
type DataAssociator interface {
	// Register 注册需要关注的信息
	Register(register *DataRegister)
	// Received 收到了信息
	Received(bucket string, name string, value any)
}

// NewDataDistributor 创建一个数据总线与数据分发器的整合体
// bufferSize 消息通知的队列长度
func NewDataDistributor(bufferSize int) *DataDistributor {
	if bufferSize <= 0 {
		bufferSize = 10
	}
	return &DataDistributor{buckets: make(map[string]*bucket), clients: make(map[uint64]*client), bufferSize: bufferSize, allRegisters: make(map[uint64]struct{})}
}

type bucket struct {
	name    string
	value   map[string]any
	marking map[string][]uint64
	lock    sync.Mutex
}

// 将数据刷入桶中
func (b *bucket) flushValue(name string, value any) []uint64 {
	b.lock.Lock()
	b.value[name] = value
	b.lock.Unlock()
	//通知客户机
	return b.marking[name]
}

func (b *bucket) mark(id uint64, names ...string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	for _, name := range names {
		b.marking[name] = append(b.marking[name], id)
	}
}

// DataDistributor 数据总线与数据分发器的整合体
type DataDistributor struct {
	lock         sync.Mutex
	clientId     uint64
	buckets      map[string]*bucket
	clients      map[uint64]*client
	bufferSize   int
	allRegisters map[uint64]struct{}
}

// ClearClients 清空所有客户端
func (d *DataDistributor) ClearClients() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.clients = make(map[uint64]*client)
	d.allRegisters = make(map[uint64]struct{})
	for _, kit := range d.buckets {
		kit.lock.Lock()
		kit.marking = make(map[string][]uint64)
		kit.lock.Unlock()
	}
}

// ClearClient 清空一个客户端
func (d *DataDistributor) ClearClient(clientId uint64) {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.clients, clientId)
	delete(d.allRegisters, clientId)
	for _, kit := range d.buckets {
		for key, res := range kit.marking {
			if slices.Contains(res, clientId) {
				kit.marking[key] = slices.DeleteFunc(res, func(v uint64) bool {
					return v == clientId
				})
			}
		}
	}
}

// Register 注册一个客户机
// 返回一个发布数据的方法
func (d *DataDistributor) Register(as DataAssociator) (send ReleaseFunc, clientId uint64, err error) {
	if as == nil {
		return nil, 0, errors.New("DataAssociator is nil")
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	dr := &DataRegister{buckets: make(map[string][]string)}
	as.Register(dr)
	if len(dr.buckets) == 0 && dr.registerAll == false {
		return nil, 0, fmt.Errorf("no buckets registered")
	}
	clientId = d.flushClientId()
	if dr.registerAll == false {
		//开始打点
		for bucketName, pointNames := range dr.buckets {
			if _, ok := d.buckets[bucketName]; !ok {
				d.buckets[bucketName] = &bucket{name: bucketName, value: make(map[string]any), marking: make(map[string][]uint64)}
			}
			d.buckets[bucketName].mark(clientId, pointNames...)
		}
	} else {
		d.allRegisters[clientId] = struct{}{}
	}
	cli := &client{dc: make(chan []any, d.bufferSize), as: as}
	go cli.flush()
	d.clients[clientId] = cli
	return d.release, clientId, nil
}

func (d *DataDistributor) flushClientId() uint64 {
	defer func() {
		d.clientId++
	}()
	return d.clientId
}

// ObtainReleaseFunc 获取写数据的方法
func (d *DataDistributor) ObtainReleaseFunc() ReleaseFunc {
	return d.release
}

// ObtainValue 获取一条数据
func (d *DataDistributor) ObtainValue(bucket string, name string) any {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.buckets[bucket]; !ok {
		return nil
	}
	return d.buckets[bucket].value[name]
}

// ObtainBucketValues 获取一个桶里的所有数据
func (d *DataDistributor) ObtainBucketValues(bucket string) map[string]any {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.buckets[bucket]; !ok {
		return nil
	}
	result := make(map[string]any)
	for k, v := range d.buckets[bucket].value {
		result[k] = v
	}
	return result
}

// 发送一条信息
func (d *DataDistributor) release(bucketName string, name string, value any) {
	d.lock.Lock()
	if _, ok := d.buckets[bucketName]; !ok {
		d.buckets[bucketName] = &bucket{name: bucketName, value: make(map[string]any), marking: make(map[string][]uint64)}
		//没有桶，绑定全局订阅的客户端
		for clientId := range d.allRegisters {
			d.buckets[bucketName].mark(clientId, name)
		}
	}
	d.lock.Unlock()
	clients := d.buckets[bucketName].flushValue(name, value)
	if len(clients) == 0 {
		return
	}
	//要用缓冲队列，用来应对
	for _, id := range clients {
		if cli, ok := d.clients[id]; ok {
			cli.dc <- []any{bucketName, name, value}
		}
	}
}
