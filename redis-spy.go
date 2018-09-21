package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	addrSepFlag = ":"
)

type redisAddr struct {
	ip   string
	port uint16
}

var redisConn net.Conn
var redisAddrParam = redisAddr{ip: "127.0.0.1", port: 6379}

func (p *redisAddr) Set(s string) error {
	valSlice := strings.Split(s, addrSepFlag)
	if len(valSlice) != 2 {
		return errors.New("invalid redisAddr")
	}
	p.ip = valSlice[0]
	port, err := strconv.ParseUint(valSlice[1], 10, 16)
	if err != nil {
		return errors.New("invalid redisAddr")
	}
	p.port = uint16(port)
	return nil
}

func (p *redisAddr) String() string {
	return fmt.Sprintf("%+v", *p)
}

func init_conn() bool {
	var err error
	addr_str := fmt.Sprintf("%v%v%v", redisAddrParam.ip, addrSepFlag, redisAddrParam.port)
	fmt.Println("addr_str:", addr_str)
	redisConn, err = net.Dial("tcp", addr_str)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func fin_conn() {
	if redisConn != nil {
		redisConn.Close()
	}
}

func send_watch_request() {
	var buffer bytes.Buffer
	if redisConn == nil {
		return
	}
	monitor_msg := "monitor\r\n"
	err := binary.Write(&buffer, binary.BigEndian, []byte(monitor_msg))
	if err != nil {
		fmt.Println(err)
	}
	redisConn.Write(buffer.Bytes())
	buffer.Reset()
}

func receive_watch_response() {
	if redisConn == nil {
		return
	}
	buf := make([]byte, 128)
	c, err := redisConn.Read(buf)
	if err != nil {
		fmt.Println("read error:", err.Error())
	}
	fmt.Println(string(buf[0:c]))
}

func init() {
	flag.Var(&redisAddrParam, "redisAddr", "set redis address, eg 127.0.0.1:6379")
	flag.Parse()
}

func main() {
	fmt.Println(redisAddrParam)
	init_conn()
	send_watch_request()
	for {
		receive_watch_response()
	}
	fin_conn()
}
