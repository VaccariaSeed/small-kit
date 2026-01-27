package small_kit

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

type Uint16 uint16
type Int16 int16
type Uint32 uint32
type Int32 int32
type Uint64 uint64
type Int64 int64
type Float32 float32
type Float64 float64

// 检查位索引是否有效
func checkBitIndex(bitIndex, totalBits int) error {
	if bitIndex < 0 || bitIndex >= totalBits {
		return fmt.Errorf("bit index %d out of range [0, %d]", bitIndex, totalBits-1)
	}
	return nil
}

// 检查位区间是否有效
func checkBitRange(start, end, totalBits int) error {
	if start < 0 || start >= totalBits {
		return fmt.Errorf("start bit index %d out of range [0, %d]", start, totalBits-1)
	}
	if end < 0 || end >= totalBits {
		return fmt.Errorf("end bit index %d out of range [0, %d]", end, totalBits-1)
	}
	if start > end {
		return fmt.Errorf("start bit index %d cannot be greater than end bit index %d", start, end)
	}
	return nil
}

// 将字节数组转换为二进制字符串
func bytesToBinaryString(data []byte) string {
	var sb strings.Builder
	for i, b := range data {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%08b", b))
	}
	return sb.String()
}

type BasicWraper interface {
	ToBytesBE() []byte                      // 转换为大端序字节数组
	ToBytesLE() []byte                      // 转换为小端序字节数组
	FromBytesBE(data []byte) error          // 从大端序字节数组转换
	FromBytesLE(data []byte) error          // 从小端序字节数组转换
	GetBit(bitIndex int) (uint8, error)     // 获取指定位的值
	GetBits(start, end int) (uint64, error) // 获取位区间的值
	ToBinaryString() string                 // 转换为二进制字符串
	ToFloat32() float32                     // 转换为 float32
	ToFloat64() float64                     // 转换为 float64
}

// ==================== Uint16 实现 ====================

func (n *Uint16) ToBytesBE() []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(*n))
	return buf
}

func (n *Uint16) ToBytesLE() []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(*n))
	return buf
}

func (n *Uint16) FromBytesBE(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data for Uint16")
	}
	*n = Uint16(binary.BigEndian.Uint16(data))
	return nil
}

func (n *Uint16) FromBytesLE(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data for Uint16")
	}
	*n = Uint16(binary.LittleEndian.Uint16(data))
	return nil
}

func (n *Uint16) GetBit(bitIndex int) (uint8, error) {
	if err := checkBitIndex(bitIndex, 16); err != nil {
		return 0, err
	}
	return uint8((uint16(*n) >> (15 - bitIndex)) & 1), nil
}

func (n *Uint16) GetBits(start, end int) (uint64, error) {
	if err := checkBitRange(start, end, 16); err != nil {
		return 0, err
	}
	length := end - start + 1
	mask := (uint16(1) << length) - 1
	return uint64((uint16(*n) >> (15 - end)) & mask), nil
}

func (n *Uint16) ToBinaryString() string {
	return bytesToBinaryString(n.ToBytesBE())
}

func (n *Uint16) ToFloat32() float32 {
	return float32(*n)
}

func (n *Uint16) ToFloat64() float64 {
	return float64(*n)
}

// ==================== Int16 实现 ====================

func (n *Int16) ToBytesBE() []byte {
	u := Uint16(uint16(*n))
	return u.ToBytesBE()
}

func (n *Int16) ToBytesLE() []byte {
	u := Uint16(uint16(*n))
	return u.ToBytesLE()
}

func (n *Int16) FromBytesBE(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data for Int16")
	}
	*n = Int16(int16(binary.BigEndian.Uint16(data)))
	return nil
}

func (n *Int16) FromBytesLE(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("insufficient data for Int16")
	}
	*n = Int16(int16(binary.LittleEndian.Uint16(data)))
	return nil
}

func (n *Int16) GetBit(bitIndex int) (uint8, error) {
	u := Uint16(uint16(*n))
	return u.GetBit(bitIndex)
}

func (n *Int16) GetBits(start, end int) (uint64, error) {
	u := Uint16(uint16(*n))
	return u.GetBits(start, end)
}

func (n *Int16) ToBinaryString() string {
	u := Uint16(uint16(*n))
	return u.ToBinaryString()
}

func (n *Int16) ToFloat32() float32 {
	return float32(*n)
}

func (n *Int16) ToFloat64() float64 {
	return float64(*n)
}

// ==================== Uint32 实现 ====================

func (n *Uint32) ToBytesBE() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(*n))
	return buf
}

func (n *Uint32) ToBytesLE() []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(*n))
	return buf
}

func (n *Uint32) FromBytesBE(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for Uint32")
	}
	*n = Uint32(binary.BigEndian.Uint32(data))
	return nil
}

func (n *Uint32) FromBytesLE(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for Uint32")
	}
	*n = Uint32(binary.LittleEndian.Uint32(data))
	return nil
}

func (n *Uint32) GetBit(bitIndex int) (uint8, error) {
	if err := checkBitIndex(bitIndex, 32); err != nil {
		return 0, err
	}
	return uint8((uint32(*n) >> (31 - bitIndex)) & 1), nil
}

func (n *Uint32) GetBits(start, end int) (uint64, error) {
	if err := checkBitRange(start, end, 32); err != nil {
		return 0, err
	}
	length := end - start + 1
	mask := (uint32(1) << length) - 1
	return uint64((uint32(*n) >> (31 - end)) & mask), nil
}

func (n *Uint32) ToBinaryString() string {
	return bytesToBinaryString(n.ToBytesBE())
}

func (n *Uint32) ToFloat32() float32 {
	return float32(*n)
}

func (n *Uint32) ToFloat64() float64 {
	return float64(*n)
}

// ==================== Int32 实现 ====================

func (n *Int32) ToBytesBE() []byte {
	u := Uint32(uint32(*n))
	return u.ToBytesBE()
}

func (n *Int32) ToBytesLE() []byte {
	u := Uint32(uint32(*n))
	return u.ToBytesLE()
}

func (n *Int32) FromBytesBE(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for Int32")
	}
	*n = Int32(int32(binary.BigEndian.Uint32(data)))
	return nil
}

func (n *Int32) FromBytesLE(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for Int32")
	}
	*n = Int32(int32(binary.LittleEndian.Uint32(data)))
	return nil
}

func (n *Int32) GetBit(bitIndex int) (uint8, error) {
	u := Uint32(uint32(*n))
	return u.GetBit(bitIndex)
}

func (n *Int32) GetBits(start, end int) (uint64, error) {
	u := Uint32(uint32(*n))
	return u.GetBits(start, end)
}

func (n *Int32) ToBinaryString() string {
	u := Uint32(uint32(*n))
	return u.ToBinaryString()
}

func (n *Int32) ToFloat32() float32 {
	return float32(*n)
}

func (n *Int32) ToFloat64() float64 {
	return float64(*n)
}

// ==================== Uint64 实现 ====================

func (n *Uint64) ToBytesBE() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(*n))
	return buf
}

func (n *Uint64) ToBytesLE() []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(*n))
	return buf
}

func (n *Uint64) FromBytesBE(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for Uint64")
	}
	*n = Uint64(binary.BigEndian.Uint64(data))
	return nil
}

func (n *Uint64) FromBytesLE(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for Uint64")
	}
	*n = Uint64(binary.LittleEndian.Uint64(data))
	return nil
}

func (n *Uint64) GetBit(bitIndex int) (uint8, error) {
	if err := checkBitIndex(bitIndex, 64); err != nil {
		return 0, err
	}
	return uint8((uint64(*n) >> (63 - bitIndex)) & 1), nil
}

func (n *Uint64) GetBits(start, end int) (uint64, error) {
	if err := checkBitRange(start, end, 64); err != nil {
		return 0, err
	}
	length := end - start + 1
	mask := (uint64(1) << length) - 1
	return (uint64(*n) >> (63 - end)) & mask, nil
}

func (n *Uint64) ToBinaryString() string {
	return bytesToBinaryString(n.ToBytesBE())
}

func (n *Uint64) ToFloat32() float32 {
	return float32(*n)
}

func (n *Uint64) ToFloat64() float64 {
	return float64(*n)
}

// ==================== Int64 实现 ====================

func (n *Int64) ToBytesBE() []byte {
	u := Uint64(uint64(*n))
	return u.ToBytesBE()
}

func (n *Int64) ToBytesLE() []byte {
	u := Uint64(uint64(*n))
	return u.ToBytesLE()
}

func (n *Int64) FromBytesBE(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for Int64")
	}
	*n = Int64(int64(binary.BigEndian.Uint64(data)))
	return nil
}

func (n *Int64) FromBytesLE(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for Int64")
	}
	*n = Int64(int64(binary.LittleEndian.Uint64(data)))
	return nil
}

func (n *Int64) GetBit(bitIndex int) (uint8, error) {
	u := Uint64(uint64(*n))
	return u.GetBit(bitIndex)
}

func (n *Int64) GetBits(start, end int) (uint64, error) {
	u := Uint64(uint64(*n))
	return u.GetBits(start, end)
}

func (n *Int64) ToBinaryString() string {
	u := Uint64(uint64(*n))
	return u.ToBinaryString()
}

func (n *Int64) ToFloat32() float32 {
	return float32(*n)
}

func (n *Int64) ToFloat64() float64 {
	return float64(*n)
}

// ==================== Float32 实现 ====================

func (n *Float32) ToBytesBE() []byte {
	u := Uint32(math.Float32bits(float32(*n)))
	return u.ToBytesBE()
}

func (n *Float32) ToBytesLE() []byte {
	u := Uint32(math.Float32bits(float32(*n)))
	return u.ToBytesLE()
}

func (n *Float32) FromBytesBE(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for Float32")
	}
	uintVal := binary.BigEndian.Uint32(data)
	*n = Float32(math.Float32frombits(uintVal))
	return nil
}

func (n *Float32) FromBytesLE(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("insufficient data for Float32")
	}
	uintVal := binary.LittleEndian.Uint32(data)
	*n = Float32(math.Float32frombits(uintVal))
	return nil
}

func (n *Float32) GetBit(bitIndex int) (uint8, error) {
	u := Uint32(math.Float32bits(float32(*n)))
	return u.GetBit(bitIndex)
}

func (n *Float32) GetBits(start, end int) (uint64, error) {
	u := Uint32(math.Float32bits(float32(*n)))
	return u.GetBits(start, end)
}

func (n *Float32) ToBinaryString() string {
	u := Uint32(math.Float32bits(float32(*n)))
	return u.ToBinaryString()
}

func (n *Float32) ToFloat32() float32 {
	return float32(*n)
}

func (n *Float32) ToFloat64() float64 {
	return float64(*n)
}

// ==================== Float64 实现 ====================

func (n *Float64) ToBytesBE() []byte {
	u := Uint64(math.Float64bits(float64(*n)))
	return u.ToBytesBE()
}

func (n *Float64) ToBytesLE() []byte {
	u := Uint64(math.Float64bits(float64(*n)))
	return u.ToBytesLE()
}

func (n *Float64) FromBytesBE(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for Float64")
	}
	uintVal := binary.BigEndian.Uint64(data)
	*n = Float64(math.Float64frombits(uintVal))
	return nil
}

func (n *Float64) FromBytesLE(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("insufficient data for Float64")
	}
	uintVal := binary.LittleEndian.Uint64(data)
	*n = Float64(math.Float64frombits(uintVal))
	return nil
}

func (n *Float64) GetBit(bitIndex int) (uint8, error) {
	u := Uint64(math.Float64bits(float64(*n)))
	return u.GetBit(bitIndex)
}

func (n *Float64) GetBits(start, end int) (uint64, error) {
	u := Uint64(math.Float64bits(float64(*n)))
	return u.GetBits(start, end)
}

func (n *Float64) ToBinaryString() string {
	u := Uint64(math.Float64bits(float64(*n)))
	return u.ToBinaryString()
}

func (n *Float64) ToFloat32() float32 {
	return float32(*n)
}

func (n *Float64) ToFloat64() float64 {
	return float64(*n)
}
