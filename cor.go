package small_kit

import (
	"errors"
	"time"
)

type ChainsHandler func(ch map[string]interface{}) (interface{}, error)

// NewCor 创建一个责任链
func NewCor(name string) *Cor {
	return &Cor{
		Name:     name,
		chains:   make(map[string]ChainsHandler),
		sequence: nil,
		chs:      make(map[string]interface{}),
		pauses:   map[string]time.Duration{},
	}
}

// Cor 责任链
type Cor struct {
	Name     string                   // 责任链的名字
	chains   map[string]ChainsHandler //责任链
	sequence []string                 //责任链的顺序
	pauses   map[string]time.Duration //某个handle运行前暂停的时间
	chs      map[string]interface{}   //运行中的参数
}

// Run 运行责任链
// result 每个责任链的运行结果
// 第一个发生错误的责任链名字
// 第一个发生的错误
func (cor *Cor) Run() (result map[string]interface{}, errorChainName string, err error) {
	if len(cor.sequence) == 0 {
		return nil, "", errors.New("no sequence")
	}
	result = make(map[string]interface{})
	for _, chainName := range cor.sequence {
		if ch, ok := cor.chains[chainName]; ok {
			if pause, pk := cor.pauses[chainName]; pk {
				if pause > 0 {
					time.Sleep(pause)
				}
			}
			value, chErr := ch(cor.chs)
			if chErr != nil {
				return nil, chainName, chErr
			} else {
				result[chainName] = value
			}
		} else {
			return nil, chainName, errors.New("chain " + chainName + " not found")
		}
	}
	return
}

// FlushParams 清空掉运行参数
func (cor *Cor) FlushParams() {
	clear(cor.chs)
}

// AppendParam 添加运行参数，他会在代码运行中流转
func (cor *Cor) AppendParam(k string, v interface{}) {
	cor.chs[k] = v
}

// AppendParams 添加多个运行参数，他会在代码运行中流转
func (cor *Cor) AppendParams(params map[string]interface{}) {
	for k, v := range params {
		cor.chs[k] = v
	}
}

// AppendChain 给责任链添加一个回调方法
// name 某个handle的名字
// chain handle
// pause handle运行前暂停的时间
func (cor *Cor) AppendChain(name string, chain ChainsHandler, pause time.Duration) {
	cor.chains[name] = chain
	cor.pauses[name] = pause
	cor.sequence = append(cor.sequence, name)
}

// DelChain 删除一个责任链
func (cor *Cor) DelChain(name string) {
	delete(cor.chains, name)
	for i, item := range cor.sequence {
		if item == name {
			// 将后面的元素向前移动
			copy(cor.sequence[i:], cor.sequence[i+1:])
			cor.sequence = cor.sequence[:len(cor.sequence)-1]
		}
	}
	delete(cor.pauses, name)
}
