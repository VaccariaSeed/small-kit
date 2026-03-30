package butor

import (
	"errors"
	"sync"
)

var NotFoundBucket = errors.New("bucket not found")

func newDataBucket(snap bool, sub map[uint64][]string, subBucket []uint64) *dataBucket {
	bucket := &dataBucket{subs: make(map[string][]uint64), snap: snap, values: make(map[string]any)}
	for id, values := range sub {
		bucket.subValues(values, id)
	}
	bucket.subBuckets = append(bucket.subBuckets, subBucket...)
	return bucket
}

// 数据桶
type dataBucket struct {
	subs       map[string][]uint64
	snap       bool
	lock       sync.RWMutex
	subBuckets []uint64
	values     map[string]any
}

// 获取某个值的快照值
func (d *dataBucket) obtain(key string) (any, error) {
	if !d.snap {
		return nil, errors.New("bucket not snap")
	}
	d.lock.RLock()
	defer d.lock.RUnlock()
	value, ok := d.values[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return value, nil
}

// 取消订阅某些值
func (d *dataBucket) unsubscribeValue(keys []string, id uint64) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, key := range keys {
		if cliIds, ok := d.subs[key]; ok {
			d.subs[key] = d.flushCli(cliIds, id)
		}
	}
}

// 取消订阅
func (d *dataBucket) unsubscribe(id uint64) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for value, ids := range d.subs {
		d.subs[value] = d.flushCli(ids, id)
	}
	d.subBuckets = d.flushCli(d.subBuckets, id)
}

func (d *dataBucket) flushCli(slice []uint64, target uint64) []uint64 {
	for i, v := range slice {
		if v == target {
			// 使用 append 删除第 i 个元素
			newSlice := append(slice[:i], slice[i+1:]...)
			return newSlice
		}
	}
	return slice
}

// 获取所有相关的cli
func (d *dataBucket) allMark() []uint64 {
	nums := append([]uint64{}, d.subBuckets...)
	for _, ids := range d.subs {
		nums = append(nums, ids...)
	}
	seen := make(map[uint64]struct{})
	result := make([]uint64, 0, len(nums))
	for _, num := range nums {
		if _, exists := seen[num]; !exists {
			seen[num] = struct{}{}
			result = append(result, num)
		}
	}
	return result

}

// 刷新值
func (d *dataBucket) flush(name string, value any) []uint64 {
	if d.snap {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.values[name] = value
	}
	return append(d.subs[name], d.subBuckets...) //获取订阅客户端
}

// 绑定订阅信息
func (d *dataBucket) subValues(values []string, id uint64) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, value := range values {
		d.subs[value] = append(d.subs[value], id)
	}
}

func (d *dataBucket) subBucket(id uint64) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.subBuckets = append(d.subBuckets, id)
}
