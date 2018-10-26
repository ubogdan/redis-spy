package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/europelee/redis-spy/election"
	"github.com/europelee/redis-spy/utils"
)

type redisClient struct {
	redisConn net.Conn
	reader    *bufio.Reader
	isLeader  bool
	commandCh chan string
	errorCh   chan error
	outPlgs   map[utils.PlgRole]utils.RecordInf
}

var redisAddrParam = *utils.New("127.0.0.1", 6379)
var raftBindAddr = *utils.New("127.0.0.1", 1000)
var raftDataDir = "/tmp/raft_data"
var raftPeers utils.NetAddrList
var plgConfigList = utils.
	SpyPluginList{SpyPlugins: []utils.
	SpyPlugin{{Key: "Leader", FilePath: "plugins/primary.so"}, {Key: "Follower", FilePath: "plugins/cache.so"}}}

func (r *redisClient) initOutPlgs(plgCfg utils.SpyPluginList) bool {
	r.outPlgs = make(map[utils.PlgRole]utils.RecordInf, len(plgCfg.SpyPlugins))
	for _, sp := range plgCfg.SpyPlugins {
		fmt.Println(sp.FilePath)
		inst, err := utils.GetPlgInst(sp.FilePath)
		if err != nil {
			fmt.Println(err)
			return false
		}
		inst.Init()
		inst.SetCache(5)
		r.outPlgs[sp.Key] = inst

	}
	return true
}
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
	r.isLeader = false
	r.commandCh = make(chan string)
	r.errorCh = make(chan error)
	return true
}

func (r *redisClient) finConn() {
	if r.redisConn != nil {
		r.redisConn.Close()
	}
	close(r.commandCh)
	close(r.errorCh)
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

func (r *redisClient) recvWatchResponse(ch <-chan election.NodeState) bool {
	if r.redisConn == nil {
		r.errorCh <- errors.New("redis conn nil")
		return false
	}
	line, err := r.reader.ReadString('\n')
	if err != nil {
		fmt.Println("read error:", err.Error())
		if err == io.EOF {
			r.errorCh <- errors.New("io.EOF")
			return false
		}
	}

	r.commandCh <- line
	return true
}

func (r *redisClient) loopRecv(ch <-chan election.NodeState) error {
	go func() {
		for {
			ret := r.recvWatchResponse(ch)
			if ret == false {
				break
			}
		}
	}()
	for {
		select {
		case nodeStat := <-ch:
			fmt.Println("nodeStat:", nodeStat)
			if nodeStat == election.Leader {
				r.isLeader = true
				pBuf, err := r.outPlgs[utils.SlaveRole].GetCache()
				if err != nil {
					fmt.Println(err)
				} else {
					for _, record := range *pBuf {
						r.outPlgs[utils.MasterRole].InputCommand(record)
					}
					r.outPlgs[utils.SlaveRole].ClearCache()
				}
			}
			if nodeStat == election.Follower {
				r.isLeader = false
				r.outPlgs[utils.MasterRole].ClearCache()
			}
		case command := <-r.commandCh:
			fmt.Println("command:", command)
			ts := time.Now().Unix()
			record := utils.CmdRecord{TimeStamp: ts, Command: command}
			if r.isLeader {
				if err := r.outPlgs[utils.MasterRole].InputCommand(record); err != nil {
					fmt.Println(err)
				}
			} else {
				if err := r.outPlgs[utils.SlaveRole].InputCommand(record); err != nil {
					fmt.Println(err)
				}
			}
		case err := <-r.errorCh:
			fmt.Println(err)
			goto ForEnd
		}
	}
ForEnd:
	return nil
}

func init() {
	flag.Var(&redisAddrParam, "redisAddr", "set redis address")
	flag.Var(&raftBindAddr, "raftBindAddr", "set raft bind address")
	flag.StringVar(&raftDataDir, "raftDataDir", raftDataDir, "set raft data directory")
	flag.Var(&raftPeers, "raftPeers", "set raft peers, default null")
	flag.Var(&plgConfigList, "plugins", "plugin settings")
}

func main() {
	flag.Parse()
	//log.Panicf("%s", raftPeers.String())
	electionInst := election.New(raftBindAddr, raftDataDir, raftPeers)
	go electionInst.Start()
	var redisClientInst = redisClient{redisConn: nil}
	redisClientInst.initOutPlgs(plgConfigList)
	fmt.Println(redisClientInst.outPlgs)
	redisClientInst.initConn(redisAddrParam)
	redisClientInst.sendWatchRequest()
	redisClientInst.loopRecv(electionInst.NodeStatCh)
	redisClientInst.finConn()
}
