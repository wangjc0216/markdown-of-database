package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"gopkg.in/mgo.v2"
)

func main() {

}

func standAloneWrite() {

}

func ReadHq(conn net.Conn, c *mgo.Collection, Code string) {

	var buf [28]byte
	var x Stock
	var y Kline_Day
	for i := 0; ; i++ {
		_, err := conn.Read(buf[0:28])
		if err == io.EOF {
			fmt.Println("此个文件传输结束")
			break
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		b_buf := bytes.NewBuffer(buf[0:28])

		binary.Read(b_buf, binary.LittleEndian, &x) //binary.LittleEndian  是内存中的字节序的概念，就是把低字节的放到了后面。网络传输一般用BigEndian，内存字节序和cpu有关，编程时要转化。
		y.Code = Code
		y.A = x.A
		y.C = x.C
		y.Date = x.Date
		y.H = x.H
		y.L = x.L
		y.O = x.O
		y.V = x.V
		//fmt.Println(y)
		err = c.Insert(&y)
		if err != nil {
			panic(err)
		}

	}

	return
}
