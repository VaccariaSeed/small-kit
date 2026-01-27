# small-kit
development kit

#### read_buf
读数据类，只用来拆byte数组的协议包
###### 使用教程
```go
buf := NewReadBuf(byte数组)
//复用这个对象，但是复归所有数据
buf.Flush(byte数组)
```

#### cor 责任链
###### 使用教程
```go
//创建一个责任链
cor := NewCor("名字")
//添加一个责任链调用
cor.AppendChain("test1", test1Handle, 1000 * time.Second)
//添加一个参数
cor.AppendParam("value1", 1)
//运行，某个handle发生错误时会中断运行并返回错误
result, errChName, err := cor.Run()
```

#### 数据监督器
###### 使用教程
1. 新建一个数据监督器
```go
//新建一个数据监督器
dcs := NewDCS[float32](数据队列最大长度, 过期毫秒值)
```
2. 新建一个回调
```go
var _ DcsCaller = (*dcsCallBack)(nil)

type dcsCallBack struct{}

func (d dcsCallBack) Trigger(value float64) {
	fmt.Println("触发", value)
}

func (d dcsCallBack) Regress(value float64) {
	fmt.Println("复归", value)
}
```
3. 绑定回调
```go
err := dcs.AppendMethodCallBack("数据类型", "比较运算符", 1, dcsCallBack{})

//例如 平均值 >= 1时触发，dcsCallBack
err := dcs.AppendMethodCallBack(AVG, GE, 1, dcsCallBack{})
```
4. 绑定一个广泛回调
```go
dcs.AppendCallBack(callBack)
// 会传入所有的有效值，只有当新增的时候才会调用
```
5. 方法
```go
	//判定是否不为空
	flag := dcs.IsNotEntry()
	//添加一个数据
	dcs.Append(1.6)
	//获取队列长度
	size := dcs.Size()
	//获取队列有效长度（没有过期的数据有多少）
	length := dcs.Length()
	//获取指定索引的数据
	value, err := dcs.Load(0)
	//获取所有的未过期的数据
	values := dcs.Effective()
	//获取所有有效值的平均值
	avg, err := dcs.Avg()
	//获取指定时间的平均值
	avg, err = dcs.AvgByValidity(time.Now().Add(-time.Duration(3000) * time.Millisecond).UnixMilli())
	//获取最大值
	maxValue, err := dcs.Max()
	//获取指定最大值
	maxValue, err = dcs.MaxByValidity(time.Now().Add(-time.Duration(3000) * time.Millisecond).UnixMilli())
	//获取最小值
	minValue, err := dcs.Min()
	//获取指定最小值
	minValue, err = dcs.MinByValidity(time.Now().Add(-time.Duration(3000) * time.Millisecond).UnixMilli())
	//获取所有有效值的和
	sumValue, err := dcs.Sum()
	//获取指定和
	sumValue, err = dcs.SumByValidity(time.Now().Add(-time.Duration(3000) * time.Millisecond).UnixMilli())
	//获取最新一次更新时间
	lastTs, err := dcs.LatestTs()
	//获取所有值，包括过期的和有效的
	allValues := dcs.LoadAll()
```

#### 基础数据包装类
```go
var u16 Uint16 = 0x1234
var i16 Int16 = -12345
var u32 Uint32 = 0x12345678
var i32 Int32 = -123456789
var f32 Float32 = 3.14159
var f64 Float64 = 3.141592653589793
```

#### 数据桶，用于程序内部数据的交互
###### 使用教程
1. 创建一个数据分发桶
```go
tor := NewDataDistributor(消息通知的队列长度)
```
2. 创建一个handler
```go
type Handle1 struct {
call ReleaseFunc
}

func (h *Handle1) Register(register *DataRegister) {
//订阅某个桶里的某些数据
register.Append("jojo2_1", "handle2_1", "handle2_2", "handle2_3")
register.Append("jojo2_2", "handle2_1", "handle2_2", "handle2_3")
}

func (h *Handle1) Received(bucket string, name string, value any) {
fmt.Println("----------------handle1-------------------")
fmt.Println(fmt.Sprintf("Handle1 收到: %s %s %d", bucket, name, value.(int64)))
}

handle1 := &Handle1{}
//申报handler
call, _ := tor.Register(handle1) //返回一个发布数据的方法
handle1.call = call
```

#### 程序内部数据交互的消息总线
###### 使用教程
```go
//创建一个消息总线
bus := NewMessageBus[Msg]()
//申报一个订阅，返回一个只读channel和一个发送数据的方法
serverChannel, sendFunc1 := bus.Apply("server", 10)
//启动监听
for value := range serverChannel {
fmt.Println(value.toString())
}
```
###### 案例
```go

type Msg struct {
	name string
	msg  string
}

func (m *Msg) toString() string {
	return fmt.Sprintf("{name:%s, msg:%s}", m.name, m.msg)
}

func TestBus(t *testing.T) {
	bus := NewMessageBus[Msg]()
	//申请一个channel
	serverChannel, sendFunc1 := bus.Apply("server", 10)
	go serverMonitor(serverChannel, sendFunc1)
	clientChannel, sendFunc2 := bus.Apply("client", 10)
	go clientMonitor(clientChannel, sendFunc2)
	time.Sleep(1 * time.Minute)
}

// 客户端
func clientMonitor(channel <-chan Msg, sendFunc func(string, Msg) error) {
	go func() {
		for {
			msg := Msg{"client", time.Now().String()}
			err := sendFunc("server", msg)
			if err != nil {
				fmt.Println("client err:", err.Error())
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
	for value := range channel {
		fmt.Println(value.toString())
	}
}

// 服务端
func serverMonitor(channel <-chan Msg, sendFunc func(string, Msg) error) {
	go func() {
		for {
			msg := Msg{"server", time.Now().String()}
			err := sendFunc("client", msg)
			if err != nil {
				fmt.Println("server err:", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	for value := range channel {
		fmt.Println(value.toString())
	}
}

```

#### 一个等待size个协程执行的通讯体
###### 使用案例
```go

func TestChanMutual_WaitByTimeout(t *testing.T) {
	mut := NewChanMutual[string](3)
	go handle1(mut)
	go handle2(mut)
	go Handle3(mut)
	err := mut.WaitByTimeout(time.Second * 8)
	fmt.Println(err)
	fmt.Println("--------------end--------------")
	values := mut.Values()
	fmt.Println(values)
}

func TestMutual(t *testing.T) {
	mut := NewChanMutual[string](3)
	go handle1(mut)
	go handle2(mut)
	go Handle3(mut)
	mut.Wait()
	fmt.Println("--------------end--------------")
	values := mut.Values()
	fmt.Println(values)
}

func Handle3(mut *ChanMutual[string]) {
	time.Sleep(6 * time.Second)
	value, err := mut.Value("handle2")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(value)
	fmt.Println("Handle3")
	mut.DoneValue("Handle3", "result_3")
}

func handle2(mut *ChanMutual[string]) {
	time.Sleep(3 * time.Second)
	fmt.Println("handle2")
	mut.DoneValue("handle2", "result_2")
}

func handle1(mut *ChanMutual[string]) {
	fmt.Println("handle1")
	mut.Done()
}

```
