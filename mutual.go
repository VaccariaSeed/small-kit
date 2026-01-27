package small_kit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// NewChanMutual 创建一个等待size个协程执行的通讯体
func NewChanMutual[T any](size int) *ChanMutual[T] {
	m := &ChanMutual[T]{
		size:   size,
		values: make(map[string]T),
	}
	if m.size < 1 {
		m.size = 1
	}
	m.snapSize = m.size
	m.wg = sync.WaitGroup{}
	m.wg.Add(m.size)
	m.ctx, m.cancel = context.WithCancel(context.Background())
	return m
}

// ChanMutual 协程消息通讯
type ChanMutual[T any] struct {
	Name     string //名字，用来区分的，一般用不到
	wg       sync.WaitGroup
	size     int
	lock     sync.Mutex
	snapSize int

	ctx    context.Context
	cancel context.CancelFunc
	//以下是协程的运行值
	values map[string]T
}

// Wait 等待执行完毕
func (m *ChanMutual[T]) Wait() {
	m.wg.Wait()
}

// WaitByTimeout 等待执行完毕或超时
func (m *ChanMutual[T]) WaitByTimeout(timeout time.Duration) error {
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.ctx.Done():
			return nil
		}
	}
	<-m.ctx.Done()
	return nil

}

func (m *ChanMutual[T]) Done() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.snapSize <= 0 {
		return errors.New("mutual already consumed")
	}
	m.wg.Done()
	m.snapSize--
	if m.snapSize <= 0 {
		m.cancel()
	}
	return nil
}

// DoneValue 一个协程执行完毕后调用
func (m *ChanMutual[T]) DoneValue(key string, value T) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.snapSize <= 0 {
		return errors.New("mutual already consumed")
	}
	m.values[key] = value
	m.wg.Done()
	m.snapSize--
	if m.snapSize <= 0 {
		m.cancel()
	}
	return nil
}

// Values 获取所有结果值
func (m *ChanMutual[T]) Values() map[string]T {
	return m.values
}

// Value 获取某个结果值
func (m *ChanMutual[T]) Value(key string) (T, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.values != nil && len(m.values) > 0 {
		if v, ok := m.values[key]; ok {
			return v, nil
		}
	}
	var result T
	return result, fmt.Errorf("not found by %s", key)
}

// ReFlush 刷新 方便第二次使用
func (m *ChanMutual[T]) ReFlush(size int) {
	if size < 1 {
		size = 1
	}
	m.size = size
	m.snapSize = m.size
	m.wg = sync.WaitGroup{}
	m.wg.Add(m.size)
	m.ctx, m.cancel = context.WithCancel(context.Background())
}
