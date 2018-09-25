package main

import (
	"bufio"
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
	reader    *bufio.Reader
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
	r.reader = bufio.NewReader(r.redisConn)
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
	line, err := r.reader.ReadString('\n')
	if err != nil {
		fmt.Println("read error:", err.Error())
		if err == io.EOF {
			return false
		}
	}
	fmt.Println(line)
	return true
}

func (r *redisClient) loopRecv() error {
	for {
		ret := r.recvWatchResponse()
		if ret == false {
			break
		}
	}
	return nil
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
	go func() {
		for {
			select {
			case nodeStat := <-electionInst.NodeStatCh:
				fmt.Println("nodeStat:", nodeStat)
				//todo
			}
		}
	}()
	var redisClientInst = redisClient{redisConn: nil}
	redisClientInst.initConn(redisAddrParam)
	redisClientInst.sendWatchRequest()
	redisClientInst.loopRecv()
	redisClientInst.finConn()
}
