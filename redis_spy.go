package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/europelee/redis-spy/election"
	"github.com/europelee/redis-spy/utils"
)

type redisClient struct {
	redisConn net.Conn
}

var redisAddrParam = *utils.New("127.0.0.1", 6379)
var raftBindAddr = *utils.New("127.0.0.1", 1000)
var raftDataDir = "/tmp/raft_data"
var raftPeers utils.NetAddrList

func (r *redisClient) initConn(redisAddrParam utils.NetAddr) bool {
	var err error
	addrStr := redisAddrParam.String()
	fmt.Println("addrStr:", addrStr)
	r.redisConn, err = net.Dial("tcp", addrStr)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (r *redisClient) finConn() {
	if r.redisConn != nil {
		r.redisConn.Close()
	}
}

func (r *redisClient) sendWatchRequest() {
	var buffer bytes.Buffer
	if r.redisConn == nil {
		return
	}
	monMsg := "monitor\r\n"
	err := binary.Write(&buffer, binary.BigEndian, []byte(monMsg))
	if err != nil {
		fmt.Println(err)
	}
	r.redisConn.Write(buffer.Bytes())
	buffer.Reset()
}

func (r *redisClient) recvWatchResponse() bool {
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

func init() {
	flag.Var(&redisAddrParam, "redisAddr", "set redis address")
	flag.Var(&raftBindAddr, "raftBindAddr", "set raft bind address")
	flag.StringVar(&raftDataDir, "raftDataDir", raftDataDir, "set raft data directory")
	flag.Var(&raftPeers, "raftPeers", "set raft peers, default null")
}

func main() {
	flag.Parse()
	//log.Panicf("%s", raftPeers.String())
	electionInst := election.New(raftBindAddr, raftDataDir, raftPeers)
	go electionInst.Start()
	var redisClientInst = redisClient{redisConn: nil}
	redisClientInst.initConn(redisAddrParam)
	redisClientInst.sendWatchRequest()
	for {
		ret := redisClientInst.recvWatchResponse()
		if ret == false {
			break
		}
	}
	redisClientInst.finConn()
}
