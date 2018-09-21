package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
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

type redisClient struct {
	redisConn net.Conn
}

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

func (r *redisClient) init_conn(redisAddrParam redisAddr) bool {
	var err error
	addr_str := fmt.Sprintf("%v%v%v", redisAddrParam.ip, addrSepFlag, redisAddrParam.port)
	fmt.Println("addr_str:", addr_str)
	r.redisConn, err = net.Dial("tcp", addr_str)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (r *redisClient) fin_conn() {
	if r.redisConn != nil {
		r.redisConn.Close()
	}
}

func (r *redisClient) send_watch_request() {
	var buffer bytes.Buffer
	if r.redisConn == nil {
		return
	}
	monitor_msg := "monitor\r\n"
	err := binary.Write(&buffer, binary.BigEndian, []byte(monitor_msg))
	if err != nil {
		fmt.Println(err)
	}
	r.redisConn.Write(buffer.Bytes())
	buffer.Reset()
}

func (r *redisClient) receive_watch_response() bool {
	if r.redisConn == nil {
		return false
	}
	buf := make([]byte, 128)
	c, err := r.redisConn.Read(buf)
	if err != nil {
		fmt.Println("read error:", err.Error())
		if err == io.EOF {
			return false
		}
	}
	fmt.Println(string(buf[0:c]))
	return true
}

func main() {
	var redisAddrParam = redisAddr{ip: "127.0.0.1", port: 6379}
	flag.Var(&redisAddrParam, "redisAddr", "set redis address, eg 127.0.0.1:6379")
	flag.Parse()
	var redisClientInst = redisClient{redisConn: nil}
	redisClientInst.init_conn(redisAddrParam)
	redisClientInst.send_watch_request()
	for {
		ret := redisClientInst.receive_watch_response()
		if ret == false {
			break
		}
	}
	redisClientInst.fin_conn()
}
