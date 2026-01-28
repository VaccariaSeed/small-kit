package small_kit

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"sync"
	"unsafe"
)

type Step bool //决定了读数据后会不会推进读索引

const (
	StepOn  Step = true  //读数据时步进（推进读索引）
	StepOff Step = false //读数据时不步进，就拿个快照（不推进读索引）
)

var IndexOutError = errors.New("array index out of bounds")

func NewReadBuf(sli []byte) *ReadBuf {
	buf := &ReadBuf{}
	buf.Flush(sli)
	return buf
}

// ReadBuf 字节序
type ReadBuf struct {
	snap     []byte
	maxIndex int //最大索引
	index    int //当前读索引
	lock     sync.Mutex
}


// Int8 获取一个int8 int8   范围 -128 到 127
func (b *ReadBuf) Int8(step Step) (int8, error) {
	value, err := b.Byte(step)
	if err != nil {
		return 0, err
	}
	return int8(value), nil
}


func (b *ReadBuf) flushArraySize(value []byte, size int) []byte {
	currentLen := len(value)
	if currentLen >= size {
		return value[:size]
	}
	// 一次性分配足够的空间
	result := make([]byte, size)
	copy(result, value)
	// 不需要显式填充0，因为make已经初始化了
	return result
}

func (b *ReadBuf) flushArray(step Step, size int, typeSize int) ([]byte, error) {
	value, err := b.Bytes(step, size)
	if err != nil {
		return nil, err
	}
	value = b.flushArraySize(value, typeSize)
	return value, nil
}

//处理不是特定长度的int16,就是长度可能是3，可能是1，然后将结果转成int16
func (b *ReadBuf) IrregularInt16(step Step, order binary.ByteOrder, size int) (int16, error) {
	value, err := b.flushArray(step, size, 2)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt16(value, order)
}

// SkipInt16 获取一个int16 读索引推进2
func (b *ReadBuf) SkipInt16(order binary.ByteOrder) (int16, error) {
	value, err := b.Bytes(StepOn, 2)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt16(value, order)
}

// Int16 获取一个int16，读索引不推进
func (b *ReadBuf) Int16(order binary.ByteOrder) (int16, error) {
	value, err := b.Bytes(StepOff, 2)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt16(value, order)
}

func (b *ReadBuf) sliceToInt16(value []byte, order binary.ByteOrder) (int16, error) {
	if order == binary.BigEndian {
		//大端
		return int16(value[0])<<8 | int16(value[1]), nil
	}
	//小端
	return int16(value[1])<<8 | int16(value[0]), nil
}

//处理不是特定长度的int32,就是长度可能是3，可能是5，然后将结果转成int32
func (b *ReadBuf) IrregularInt32(step Step, order binary.ByteOrder, size int) (int32, error) {
	value, err := b.flushArray(step, size, 4)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt32(value, order)
}

// SkipInt32 获取一个int32，读索引推进4
func (b *ReadBuf) SkipInt32(order binary.ByteOrder) (int32, error) {
	value, err := b.Bytes(StepOn, 4)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt32(value, order)
}

// Int32 获取一个int32，读索引不推进
func (b *ReadBuf) Int32(order binary.ByteOrder) (int32, error) {
	value, err := b.Bytes(StepOff, 4)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt32(value, order)
}

func (b *ReadBuf) sliceToInt32(value []byte, order binary.ByteOrder) (int32, error) {
	if order == binary.BigEndian {
		//大端
		return int32(value[0])<<24 | int32(value[1])<<16 | int32(value[2])<<8 | int32(value[3]), nil
	}
	//小端
	return int32(value[3])<<24 | int32(value[2])<<16 | int32(value[1])<<8 | int32(value[0]), nil
}

//处理不是特定长度的int64,就是长度可能是3，可能是5，然后将结果转成int64
func (b *ReadBuf) IrregularInt64(step Step, order binary.ByteOrder, size int) (int64, error) {
	value, err := b.flushArray(step, size, 8)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt64(value, order)
}

// SkipInt64 获取一个int64, 读索引推进8
func (b *ReadBuf) SkipInt64(order binary.ByteOrder) (int64, error) {
	value, err := b.Bytes(StepOn, 8)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt64(value, order)
}

// Int64 获取一个int64, 读索引不推进
func (b *ReadBuf) Int64(order binary.ByteOrder) (int64, error) {
	value, err := b.Bytes(StepOff, 8)
	if err != nil {
		return 0, err
	}
	return b.sliceToInt64(value, order)
}

func (b *ReadBuf) sliceToInt64(value []byte, order binary.ByteOrder) (int64, error) {
	if order == binary.BigEndian {
		//大端
		return int64(value[0])<<56 | int64(value[1])<<48 | int64(value[2])<<40 | int64(value[3])<<32 |
			int64(value[4])<<24 | int64(value[5])<<16 | int64(value[6])<<8 | int64(value[7]), nil
	}
	//小端
	return int64(value[7])<<56 | int64(value[6])<<48 | int64(value[5])<<40 | int64(value[4])<<32 |
		int64(value[3])<<24 | int64(value[2])<<16 | int64(value[1])<<8 | int64(value[0]), nil
}

//处理不是特定长度的Uint16,就是长度可能是3，可能是5，然后将结果转成int16
func (b *ReadBuf) IrregularUint16(step Step, order binary.ByteOrder, size int) (uint16, error) {
	value, err := b.flushArray(step, size, 2)
	if err != nil {
		return 0, err
	}
	return order.Uint16(value[:]), nil
}

// SkipUint16 获取一个uint16；读索引推进2
func (b *ReadBuf) SkipUint16(order binary.ByteOrder) (uint16, error) {
	value, err := b.Bytes(StepOn, 2)
	if err != nil {
		return 0, err
	}
	return order.Uint16(value[:]), nil
}

// Uint16 获取一个uint16；读索引不推进
func (b *ReadBuf) Uint16(order binary.ByteOrder) (uint16, error) {
	value, err := b.Bytes(StepOff, 2)
	if err != nil {
		return 0, err
	}
	return order.Uint16(value[:]), nil
}

//处理不是特定长度的Uint32,就是长度可能是3，可能是5，然后将结果转成int32
func (b *ReadBuf) IrregularUint32(step Step, order binary.ByteOrder, size int) (uint32, error) {
	value, err := b.flushArray(step, size, 4)
	if err != nil {
		return 0, err
	}
	return order.Uint32(value[:]), nil
}

// SkipUint32 uint32，读索引推进4
func (b *ReadBuf) SkipUint32(order binary.ByteOrder) (uint32, error) {
	value, err := b.Bytes(StepOn, 4)
	if err != nil {
		return 0, err
	}
	return order.Uint32(value[:]), nil
}

// Uint32 uint32，读索引不推进
func (b *ReadBuf) Uint32(order binary.ByteOrder) (uint32, error) {
	value, err := b.Bytes(StepOff, 4)
	if err != nil {
		return 0, err
	}
	return order.Uint32(value[:]), nil
}

//处理不是特定长度的Uint64,就是长度可能是3，可能是5，然后将结果转成int64
func (b *ReadBuf) IrregularUint64(step Step, order binary.ByteOrder, size int) (uint64, error) {
	value, err := b.flushArray(step, size, 8)
	if err != nil {
		return 0, err
	}
	return order.Uint64(value[:]), nil
}

// SkipUint64 uint64，读索引推进8
func (b *ReadBuf) SkipUint64(order binary.ByteOrder) (uint64, error) {
	value, err := b.Bytes(StepOn, 8)
	if err != nil {
		return 0, err
	}
	return order.Uint64(value[:]), nil
}

// Uint64 uint64，读索引不推进
func (b *ReadBuf) Uint64(order binary.ByteOrder) (uint64, error) {
	value, err := b.Bytes(StepOff, 8)
	if err != nil {
		return 0, err
	}
	return order.Uint64(value[:]), nil
}

// SkipFloat32 float32，读索引推进4
func (b *ReadBuf) SkipFloat32(order binary.ByteOrder) (float32, error) {
	value, err := b.Bytes(StepOn, 4)
	if err != nil {
		return 0, err
	}
	bits := order.Uint32(value[:])
	return *(*float32)(unsafe.Pointer(&bits)), nil
}

// Float32 float32，读索引不推进
func (b *ReadBuf) Float32(order binary.ByteOrder) (float32, error) {
	value, err := b.Bytes(StepOn, 4)
	if err != nil {
		return 0, err
	}
	bits := order.Uint32(value[:])
	return *(*float32)(unsafe.Pointer(&bits)), nil
}

// SkipFloat64 float64，读索引推进8
func (b *ReadBuf) SkipFloat64(order binary.ByteOrder) (float64, error) {
	value, err := b.Bytes(StepOn, 8)
	if err != nil {
		return 0, err
	}
	bits := order.Uint64(value[:])
	return *(*float64)(unsafe.Pointer(&bits)), nil
}

// Float64 float64，读索引不推进
func (b *ReadBuf) Float64(order binary.ByteOrder) (float64, error) {
	value, err := b.Bytes(StepOff, 8)
	if err != nil {
		return 0, err
	}
	bits := order.Uint64(value[:])
	return *(*float64)(unsafe.Pointer(&bits)), nil
}

// 读取指定长度的字符串
func (b *ReadBuf) HexString(step Step, size int) (string, error) {
	value, err := b.Bytes(step, size)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

// 读取指定长度的字符串
func (b *ReadBuf) String(step Step, size int) (string, error) {
	hex, err := b.Bytes(step, size)
	if err != nil {
		return "", err
	}
	return string(hex), nil
}

// Bytes Reset 获取指定长度的byte
func (b *ReadBuf) Bytes(step Step, size int) ([]byte, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if size <= 0 {
		return nil, errors.New("size must be greater than zero")
	}
	if b.index+size > b.maxIndex {
		return nil, IndexOutError
	}
	value := make([]byte, size)
	copy(value, b.snap[b.index:b.index+size])
	// 根据步进标志更新索引
	b.step(step, size)
	return value, nil
}

// Byte 获取一个byte
func (b *ReadBuf) Byte(step Step) (byte, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.index > b.maxIndex {
		//数组越界
		return 0, IndexOutError
	}
	//可以正常获取到值
	defer b.step(step, 1)
	return b.snap[b.index], nil
}

// 步进指定长度
func (b *ReadBuf) step(stepFlag Step, stepNum int) {
	if stepFlag == StepOn {
		b.index = b.index + stepNum
	}
}

// Skip 跳过指定长度
func (b *ReadBuf) Skip(size int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.step(StepOn, size)
}


// Flush 重置
func (b *ReadBuf) Flush(data []byte) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.snap = append([]byte(nil), data...)
	b.maxIndex = len(data) - 1
	b.index = 0
}
