package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	Pool     redis.Pool     //连接池
	oSynWait sync.WaitGroup //互斥锁
)

const (
	OnMaxRun   = 20000                     //单线程执行命令数
	AllMaxRun  = 50                        //并发数
	TimeFormat = "2006-01-02 15:04:05.999" //打印时间
	Type       = "set"                     //指令类型
)

func main() {
	fmt.Println("--------------------------redis压力测试--------------------------")
	GetPool(&Pool)
	OldTime := time.Now()
	fmt.Println("开始", OldTime.Format(TimeFormat))

	oSynWait.Add(AllMaxRun)
	for i := 0; i < AllMaxRun; i++ {
		go ReadWriteInfo(i)
		// ReadWriteInfo()
	}
	oSynWait.Wait()

	EndTime := time.Now()
	fmt.Println("结束", EndTime.Format(TimeFormat))
	fmt.Printf("并发数: %d ;总数据量：%d ;耗时: %.2fs ;\n类型：%s ;key：1;value：1;\nqps：%.0f\n",
		AllMaxRun, AllMaxRun*OnMaxRun, EndTime.Sub(OldTime).Seconds(), Type, AllMaxRun*OnMaxRun/time.Now().Sub(OldTime).Seconds())
	fmt.Println("-----------------------------测试结束-----------------------------")
}

//取得一个连接池
func GetPool(this *redis.Pool) {
	this.MaxActive = 10
	this.MaxIdle = 10
	this.Wait = true
	this.IdleTimeout = 100 * time.Second
	this.Dial = func() (conn redis.Conn, err error) {
		conn, err = redis.Dial("tcp", "127.0.0.1:6379")
		if err != nil {
			fmt.Println("连接失败：", err)
		}
		return
	}
}

//存放数据
func ReadWriteInfo(num int) {
	defer oSynWait.Done()

	for i := 0; i < OnMaxRun; i++ {
		//存
		conn := Pool.Get()
		// _, err := conn.Do(Type, i+OnMaxRun*num, i+OnMaxRun*num)
		// if err != nil {
		// 	fmt.Println("存放数据失败", err)
		// 	return
		// }
		//取
		_, err := conn.Do("get", i+OnMaxRun*num)
		if err != nil {
			fmt.Println("取出数据失败", err)
			return
		}
		//关
		err = conn.Close()
		if err != nil {
			fmt.Println("关闭连接失败 ", err)
			return
		}
	}
}
