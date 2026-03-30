package butor

import (
	"errors"
	"reflect"
	"sync"
)

// ReleaseFunc 发布一个数据
// bucket 数据桶
// name 数据名
// value 数据
type ReleaseFunc func(bucket string, name string, value any) error

type dataSwitch struct {
	value   any
	isFirst bool //是否是第一次
	ids     []uint64
	lock    sync.Mutex
}

func (s *dataSwitch) removeCli(id uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for i, v := range s.ids {
		if v == id {
			// 使用 append 删除第 i 个元素
			s.ids = append(s.ids[:i], s.ids[i+1:]...)
			return
		}
	}
}

func (s *dataSwitch) subSwitch(id uint64) {
	s.ids = append(s.ids, id)
}

func (s *dataSwitch) flush(value any) ([]uint64, bool) {
	if value == nil {
		return nil, false
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return s.compare(value)
	default:
		return nil, false
	}
}

func (s *dataSwitch) compare(value any) ([]uint64, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isFirst {
		s.isFirst = false
		s.value = value
		return s.ids, true
	}
	flag := reflect.DeepEqual(s.value, value)
	if !flag {
		s.value = value
		return s.ids, true
	}
	return nil, false
}

// id生成器
type idGenerator struct {
	snap uint64
	lock sync.Mutex
}

// 生成一个id
func (id *idGenerator) nextID() uint64 {
	defer func() { id.snap++ }()
	return id.snap
}

// 创建一个订阅加载中间件
func newRegister() *Register {
	return &Register{
		subscribe: make(map[string][]string),
		creater:   make(map[string]bool),
		switches:  make(map[string][]string),
	}
}

// Register 注册器
type Register struct {
	subscribe map[string][]string //订阅某个桶的某些值
	subBucket []string            //订阅某些桶的所有数据
	creater   map[string]bool     //创建某些桶，bool表示是否保存最新值快照
	switches  map[string][]string //只有当某个桶的某些值发生变更或者第一次加入时就会发起通知
}

// Switch 订阅某一个桶里的某个数据，只有当这个值第一次加入或者发生变更时才会通知,只有基本数据类型有用
func (r *Register) Switch(bucketName string, key ...string) {
	r.switches[bucketName] = append(r.switches[bucketName], key...)
}

// Subscribe 订阅某些桶的某些数据
// bucketName 桶名称
// key 点位标识
func (r *Register) Subscribe(bucketName string, key ...string) {
	r.subscribe[bucketName] = append(r.subscribe[bucketName], key...)
}

// SubscribeBucket 订阅一整个桶
// bucketName 桶名称
func (r *Register) SubscribeBucket(bucketName ...string) {
	r.subBucket = append(r.subBucket, bucketName...)
}

// Create 创建一个桶
// bucketName 桶名称
// snap 该桶是否保留数据的最新值快照
func (r *Register) Create(bucketName string, snap bool) {
	r.creater[bucketName] = snap
}

// Operator 操作器
type Operator struct {
	isClosed bool //是否被关闭

	release           ReleaseFunc                                       //发送一个数据
	Id                uint64                                            //操作员id
	close             func(id uint64)                                   //关闭操作员
	creater           map[string]bool                                   //自己创建的桶
	unsubscribe       func(id uint64, bucketName ...string)             //取消订阅某个桶
	unsubscribeValue  func(id uint64, bucketName string, key ...string) // 关闭订阅某个桶的某些数据
	unsubscribeSwitch func(id uint64, bucketName string, key ...string) //取消订阅变更值
	closeBucket       func(bucketName ...string)                        //关停某个桶
	obtain            func(bucketName, key string) (any, error)         //获取某些值
}

// MandatoryObtain 强制获取某个值
func (o *Operator) MandatoryObtain(bucketName, key string) (any, error) {
	return o.obtain(bucketName, key)
}

// CloseBucket 关停某个桶
func (o *Operator) CloseBucket(bucketName string) error {
	if _, ok := o.creater[bucketName]; !ok {
		return errors.New("the target bucket was not created by yourself and cannot be deleted")
	}
	o.closeBucket(bucketName)
	return nil
}

// UnsubscribeSwitch 取消订阅变更值
func (o *Operator) UnsubscribeSwitch(bucketName string, key ...string) {
	o.unsubscribeSwitch(o.Id, bucketName, key...)
}

// UnsubscribeValue 取消订阅某个桶的某些数据
func (o *Operator) UnsubscribeValue(bucketName string, key ...string) {
	o.unsubscribeValue(o.Id, bucketName, key...)
}

// Unsubscribe 取消订阅某个桶
func (o *Operator) Unsubscribe(bucketName ...string) {
	o.unsubscribe(o.Id, bucketName...)
}

// Release 发送一个数据
func (o *Operator) Release(bucketName string, key string, value any) error {
	if o.isClosed {
		return errors.New("operator is closed")
	}
	if _, ok := o.creater[bucketName]; !ok {
		return errors.New("the target bucket was not created by ourselves and cannot be written to")
	}
	return o.release(bucketName, key, value)
}

// CloseOneSelf 关闭自己
func (o *Operator) CloseOneSelf() {
	o.close(o.Id)
	o.isClosed = true
}

// Scheduler 调度员接口，所有调度器都要实现它
type Scheduler interface {
	Register(register *Register) // Register 注册需要关注的信息

	Received(bucketName string, name string, value any) // 收到了信息

	Syntony(operator *Operator) //传入一个Scheduler的id和一个操作器

	Switches(bucketName string, name string, value any) //订阅的变更值发生了变更

	SourceCloed(bucketName string) // 产生数据订阅的桶被关闭了
}
