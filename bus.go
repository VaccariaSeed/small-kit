package small_kit

import (
	"fmt"
	"sync"
)

func NewMessageBus[T any]() *MessageBus[T] {
	return &MessageBus[T]{clients: make(map[string]chan T)}
}

// MessageBus 程序内部通讯，以channel进行通知
type MessageBus[T any] struct {
	clients map[string]chan T
	lock    sync.Mutex
}

// Broadcast 广播发送给所有客户端
// 参数：
// msg 消息
func (b *MessageBus[T]) Broadcast(msg T) {
	b.lock.Lock()
	defer b.lock.Unlock()
	for _, ch := range b.clients {
		if ch != nil {
			ch <- msg
		}
	}
}

// Send 向执行客户端发送一条数据
// 参数：
// ident 唯一标志
// msg 消息
func (b *MessageBus[T]) Send(ident string, msg T) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if nel, ok := b.clients[ident]; ok {
		if nel == nil {
			return fmt.Errorf("%s channel is nil", ident)
		}
		select {
		case nel <- msg:
			return nil
		default:
			return fmt.Errorf("%s channel is full or unable to write", ident) // channel 已满或无法写入
		}
	}
	return fmt.Errorf("%s no corresponding channel found", ident)
}

// Apply 申请一个通道
// 参数：
// ident 唯一标志
// size channel大小
// 返回值
// 只读channel
// 发布数据的方法
func (b *MessageBus[T]) Apply(ident string, size byte) (<-chan T, func(string, T) error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if size == 0 {
		size = 1
	}
	nel := make(chan T, size)
	b.clients[ident] = nel
	return nel, b.Send
}

// Cancel 注销对应的channel
// 参数：
// ident 唯一标志
func (b *MessageBus[T]) Cancel(ident string) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	if nel, ok := b.clients[ident]; ok {
		delete(b.clients, ident)
		close(nel)
		return nil
	}
	return fmt.Errorf("%s no corresponding channel found", ident)
}
