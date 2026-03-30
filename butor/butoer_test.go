package butor

import (
	"fmt"
	"testing"
	"time"
)

func TestButor(t *testing.T) {
	distributor := NewDistributor(10)
	//创建第一个
	fs := &FirstScheduler{}
	if err := distributor.Register(fs); err != nil {
		t.Fatalf("register first scheduler failed: %v", err)
		return
	}
	ss := &SecondScheduler{}
	if err := distributor.Register(ss); err != nil {
		t.Fatalf("register second scheduler failed: %v", err)
		return
	}
	ts := &ThirdScheduler{}
	if err := distributor.Register(ts); err != nil {
		t.Fatalf("register third scheduler failed: %v", err)
		return
	}
	go func() {
		for {
			fmt.Printf("firstClient send time:%s\n", time.Now().Format("2006-01-02 15:04:05.000"))
			if err := fs.Release("first_bucket", "first_bucket_key1", "value1"); err != nil {
				fmt.Printf("firstClient release failed: %v\n", err)
			}
			if err := fs.Release("first_bucket", "first_bucket_key2", "value2"); err != nil {
				fmt.Printf("firstClient release failed: %v\n", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			fmt.Printf("secondClient send time:%s\n", time.Now().Format("2006-01-02 15:04:05.000"))
			if err := ss.Release("second_bucket", "second_bucket_key1", "value1"); err != nil {
				fmt.Printf("secondClient release failed: %v\n", err)
			}
			if err := ss.Release("second_bucket", "second_bucket_key2", "value2"); err != nil {
				fmt.Printf("secondClient release failed: %v\n", err)
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	go func() {
		//rand.Seed(time.Now().UnixNano())
		for {

			//randomUint64 := rand.Uint64()
			fmt.Printf("发送随机值:%s  值：%d\n", time.Now().Format("2006-01-02 15:04:05.000"), 6659)
			if err := ts.Release("third_bucket_switch", "jojo", 6659); err != nil {
				fmt.Printf("third_bucket_switch failed: %v\n", err)
			}
			time.Sleep(200 * time.Millisecond)
		}

	}()

	go func() {
		time.Sleep(10 * time.Second)
		fmt.Println("关闭switch订阅")
		fs.UnsubscribeSwitch("third_bucket_switch", "jojo")
		time.Sleep(10 * time.Second)
		//ss.CloseBucket("second_bucket")
		//fs.UnsubscribeValue("second_bucket", "second_bucket_key1")
	}()
	time.Sleep(1 * time.Minute)
}

var _ Scheduler = (*FirstScheduler)(nil)

type FirstScheduler struct {
	*Operator
}

func (f *FirstScheduler) Register(register *Register) {
	register.Create("first_bucket", false)
	register.Subscribe("second_bucket", "second_bucket_key1", "second_bucket_key2")
	register.SubscribeBucket("third_bucket")
	register.Switch("third_bucket_switch", "jojo")
}

func (f *FirstScheduler) Received(bucketName string, name string, value any) {
	fmt.Printf("FirstClient Received bucket:%s point:%s value:%v time:%s\n", bucketName, name, value, time.Now().Format("2006-01-02 15:04:05.000"))
}

func (f *FirstScheduler) Syntony(operator *Operator) {
	f.Operator = operator
}

func (f *FirstScheduler) Switches(bucketName string, name string, value any) {
	fmt.Printf("收到一个随机值变更， bucket:%s point:%s value:%v time:%s\n", bucketName, name, value, time.Now().Format("2006-01-02 15:04:05.000"))
}

func (f *FirstScheduler) SourceCloed(bucketName string) {
	fmt.Println(bucketName + "closed")
}

/*------------------------------------*/

var _ Scheduler = (*SecondScheduler)(nil)

type SecondScheduler struct {
	*Operator
}

func (s *SecondScheduler) Register(register *Register) {
	register.Create("second_bucket", true)
	register.Subscribe("first_bucket", "first_bucket_key1", "first_bucket_key2")
	register.SubscribeBucket("third_bucket")
}

func (s *SecondScheduler) Received(bucketName string, name string, value any) {
	fmt.Printf("SecondClient Received bucket:%s point:%s value:%v time:%s\n", bucketName, name, value, time.Now().Format("2006-01-02 15:04:05.000"))
}

func (s *SecondScheduler) Syntony(operator *Operator) {
	s.Operator = operator
}

func (s *SecondScheduler) Switches(bucketName string, name string, value any) {
	//TODO implement me
	panic("implement me")
}

func (s *SecondScheduler) SourceCloed(bucketName string) {
	//TODO implement me
	panic("implement me")
}

/*---------------------------*/
var _ Scheduler = (*ThirdScheduler)(nil)

type ThirdScheduler struct {
	*Operator
}

func (t *ThirdScheduler) Register(register *Register) {
	register.Create("third_bucket_switch", false)
}

func (t *ThirdScheduler) Received(bucketName string, name string, value any) {

}

func (t *ThirdScheduler) Syntony(operator *Operator) {
	t.Operator = operator
}

func (t *ThirdScheduler) Switches(bucketName string, name string, value any) {
	//TODO implement me
	panic("implement me")
}

func (t *ThirdScheduler) SourceCloed(bucketName string) {
	//TODO implement me
	panic("implement me")
}
