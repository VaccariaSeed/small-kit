package small_kit

import (
	"errors"
	"sync"
	"time"
)

type CBT string

type OPE string

const (
	AVG CBT = "AVG"
	MAX CBT = "MAX"
	MIN CBT = "MIN"
	SUM CBT = "SUM"

	GE    OPE = ">="
	LE    OPE = "<="
	Big   OPE = ">"
	LT    OPE = "<"
	NE    OPE = "!="
	EQUAL OPE = "="
)

// NewDCS 创建数据缓存监督器
// size 队列大小
// expired 数据过期周期，单位：毫秒；为0表示没有过期时间
func NewDCS[T Calculable](size int, expired int64) *DataCacher[T] {
	return &DataCacher[T]{
		data:    make([]*dataBox[T], 0, size),
		expired: expired,
		size:    size,
	}
}

type Calculable interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}
type dataBox[T Calculable] struct {
	data T
	ts   int64
}

// DataCacher 数据缓存监督器
type DataCacher[T Calculable] struct {
	lock             sync.Mutex
	data             []*dataBox[T]
	expired          int64 //数据过期时间，单位：毫秒
	size             int   //队列长度
	Satisfy          int   //当有效数据的数量大于等于某个数值时才会调用各种回调
	lastTs           int64
	avgCallBack      []*dcsCallBackSnap
	maxCallBack      []*dcsCallBackSnap
	minCallBack      []*dcsCallBackSnap
	sumCallBack      []*dcsCallBackSnap
	templateCallBack TemplateCallBacker[T]
}

// Append 新增一个数据,使用系统毫秒级时间戳
func (d *DataCacher[T]) Append(value T) {
	d.AppendByTs(value, time.Now().UnixMilli())
}

// AppendByTs 新增一个数据，自定义一个时间戳
func (d *DataCacher[T]) AppendByTs(value T, ts int64) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if len(d.data) >= d.size {
		box := d.data[0]
		box.data = value
		box.ts = ts
		copy(d.data, d.data[1:])
		d.data[len(d.data)-1] = box
	} else {
		d.data = append(d.data, &dataBox[T]{data: value, ts: ts})
	}
	d.lastTs = ts
	//调用回调
	allData := d.effective()
	if d.Satisfy > 0 {
		if len(allData) < d.Satisfy {
			return
		}
	}
	if d.avgCallBack != nil && len(d.avgCallBack) > 0 {
		val, err := d.avg(allData)
		if err == nil {
			for _, v := range d.avgCallBack {
				v.run(val)
			}
		}
	}
	if d.maxCallBack != nil && len(d.maxCallBack) > 0 {
		maxValue, err := d.max(allData)
		if err == nil {
			for _, v := range d.maxCallBack {
				v.run(float64(maxValue))
			}
		}
	}
	if d.minCallBack != nil && len(d.minCallBack) > 0 {
		minValue, err := d.min(allData)
		if err == nil {
			for _, v := range d.minCallBack {
				v.run(float64(minValue))
			}
		}
	}
	if d.sumCallBack != nil && len(d.sumCallBack) > 0 {
		sumValue, err := d.sum(allData)
		if err == nil {
			for _, v := range d.sumCallBack {
				v.run(sumValue)
			}
		}
	}
	if d.templateCallBack != nil {
		d.templateCallBack(allData)
	}
}

// Load 获取指定索引的数据
func (d *DataCacher[T]) Load(index int) (T, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if len(d.data) == 0 || index >= len(d.data) {
		var zero T
		return zero, errors.New("index out of range")
	}
	return d.data[index].data, nil
}

// Length 获取现在有多少个数据
func (d *DataCacher[T]) Length() int {
	d.lock.Lock()
	defer d.lock.Unlock()
	return len(d.data)
}

// Size 获取队列长度
func (d *DataCacher[T]) Size() int {
	return d.size
}

// Expired 获取过期周期，单位：毫秒
func (d *DataCacher[T]) Expired() int64 {
	return d.expired
}

// IsNotEntry 判定是否不为空，没有任何数据
func (d *DataCacher[T]) IsNotEntry() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return len(d.data) > 0
}

// Effective 获取所有的有效数据
func (d *DataCacher[T]) Effective() []T {
	if d.expired > 0 {
		ts := time.Now().Add(-time.Duration(d.expired) * time.Millisecond).UnixMilli()
		return d.Validity(ts)
	} else {
		//获取所有的数据
		return d.Validity(0)
	}
}

func (d *DataCacher[T]) effective() []T {
	var ts int64
	if d.expired > 0 {
		ts = time.Now().Add(-time.Duration(d.expired) * time.Millisecond).UnixMilli()
	}
	var result []T
	for i := 0; i < len(d.data); i++ {
		if ts >= 0 {
			//判定过期时间
			if ts <= d.data[i].ts {
				result = append(result, d.data[i].data)
			}
		} else {
			result = append(result, d.data[i].data)
		}
	}
	return result
}

// Validity 获取指定毫秒级时间戳之后时间的所有的有效数据
func (d *DataCacher[T]) Validity(ts int64) []T {
	d.lock.Lock()
	defer d.lock.Unlock()
	var result []T
	for i := 0; i < len(d.data); i++ {
		if ts >= 0 {
			//判定过期时间
			if ts <= d.data[i].ts {
				result = append(result, d.data[i].data)
			}
		} else {
			result = append(result, d.data[i].data)
		}
	}
	return result
}

// Avg 获取有效数据的平均值
func (d *DataCacher[T]) Avg() (float64, error) {
	values := d.Effective()
	return d.avg(values)
}

// AvgByValidity 获取指定时间戳之后的数据并求取平均值
func (d *DataCacher[T]) AvgByValidity(ts int64) (float64, error) {
	values := d.Validity(ts)
	return d.avg(values)
}

func (d *DataCacher[T]) avg(values []T) (float64, error) {
	if values == nil || len(values) == 0 {
		return 0, errors.New("cache is empty")
	}
	var sum float64
	for _, v := range values {
		sum += float64(v)
	}
	return sum / float64(len(values)), nil
}

// Max 获取有效值的最大值
func (d *DataCacher[T]) Max() (T, error) {
	values := d.Effective()
	return d.max(values)
}

// MaxByValidity 获取指定时间戳之后的所有有效值的最大值
func (d *DataCacher[T]) MaxByValidity(ts int64) (T, error) {
	values := d.Validity(ts)
	return d.max(values)
}

func (d *DataCacher[T]) max(values []T) (T, error) {
	if values == nil || len(values) == 0 {
		return 0, errors.New("cache is empty")
	}
	maxValue := values[0]
	for index := 1; index < len(values); index++ {
		if maxValue < values[index] {
			maxValue = values[index]
		}
	}
	return maxValue, nil
}

// Min 获取所有有效值的平均值
func (d *DataCacher[T]) Min() (T, error) {
	values := d.Effective()
	return d.min(values)
}

// MinByValidity 获取指定时间戳之后的所有有效值的最小值
func (d *DataCacher[T]) MinByValidity(ts int64) (T, error) {
	values := d.Validity(ts)
	return d.min(values)
}

func (d *DataCacher[T]) min(values []T) (T, error) {
	if values == nil || len(values) == 0 {
		return 0, errors.New("cache is empty")
	}
	minValue := values[0]
	for index := 1; index < len(values); index++ {
		if minValue > values[index] {
			minValue = values[index]
		}
	}
	return minValue, nil
}

// Sum 获取所有有效值的和
func (d *DataCacher[T]) Sum() (float64, error) {
	values := d.Effective()
	return d.sum(values)
}

// SumByValidity 获取指定时间戳之后的所有有效数据的和
func (d *DataCacher[T]) SumByValidity(ts int64) (float64, error) {
	values := d.Validity(ts)
	return d.sum(values)
}

func (d *DataCacher[T]) sum(values []T) (float64, error) {
	if values == nil || len(values) == 0 {
		return 0, errors.New("cache is empty")
	}
	var sum float64
	for _, v := range values {
		sum += float64(v)
	}
	return sum, nil
}

func (d *DataCacher[T]) LatestTs() (int64, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.lastTs > 0 {
		return d.lastTs, nil
	}
	return 0, errors.New("maybe cache is empty")
}

// LoadAll 获取所有数据，包括有效的和失效的，如果返回nil则表示没有数据
func (d *DataCacher[T]) LoadAll() map[int64][]T {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.data == nil || len(d.data) == 0 {
		return nil
	}
	result := make(map[int64][]T, len(d.data))
	for _, v := range d.data {
		result[v.ts] = append(result[v.ts], v.data)
	}
	return result
}

type DcsCaller interface {
	Trigger(value float64) //触发
	Regress(value float64) //复归
}

type TemplateCallBacker[T Calculable] func(values []T)

type dcsCallBackSnap struct {
	cbType        CBT //类型
	opeType       OPE
	templateValue float64
	cb            DcsCaller
}

func (d *dcsCallBackSnap) run(value float64) {
	var flag bool
	switch d.opeType {
	case GE:
		flag = d.templateValue >= value
	case LE:
		flag = d.templateValue <= value
	case Big:
		flag = d.templateValue > value
	case LT:
		flag = d.templateValue < value
	case NE:
		flag = d.templateValue != value
	case EQUAL:
		flag = d.templateValue == value
	default:
		return
	}
	if flag {
		d.cb.Trigger(value)
	} else {
		d.cb.Regress(value)
	}
}

// AppendMethodCallBack 添加 函数回调
func (d *DataCacher[T]) AppendMethodCallBack(cbType CBT, opeType OPE, templateValue float64, cb DcsCaller) error {
	if cb == nil {
		return errors.New("cb is nil")
	}
	if opeType == "" {
		return errors.New("opeType is empty")
	}
	if opeType != GE && opeType != LE && opeType != Big && opeType != LT && opeType != NE && opeType != EQUAL {
		return errors.New("opeType is invalid")
	}
	dcbs := &dcsCallBackSnap{cbType: cbType, opeType: opeType, templateValue: templateValue, cb: cb}
	switch cbType {
	case AVG:
		d.avgCallBack = append(d.avgCallBack, dcbs)
	case MAX:
		d.maxCallBack = append(d.maxCallBack, dcbs)
	case MIN:
		d.minCallBack = append(d.minCallBack, dcbs)
	case SUM:
		d.sumCallBack = append(d.sumCallBack, dcbs)
	default:
		return errors.New("cbType is invalid")
	}
	return nil
}

// AppendCallBack 添加自定义回调
func (d *DataCacher[T]) AppendCallBack(cb TemplateCallBacker[T]) error {
	if cb == nil {
		return errors.New("cb is nil")
	}
	d.templateCallBack = cb
	return nil
}
