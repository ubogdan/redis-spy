package main

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/europelee/redis-spy/utils"
)

const (
	cacheDefaultSize    = 1024
	cacheDefaultTimeOut = 0
)

//Cache default for slave redis-py
type Cache struct {
	cacheInitSize int
	cacheTimeOut  int
	cacheBuffer   []utils.CmdRecord
	utils.RecordInf
}

func init() {

}

//Init initial some config
func (p *Cache) Init() error {
	p.cacheInitSize = cacheDefaultSize
	p.cacheTimeOut = cacheDefaultTimeOut
	p.cacheBuffer = nil
	return nil
}

//SetCache set cache config
func (p *Cache) SetCache(timeout int) error {
	p.cacheTimeOut = timeout
	if p.cacheTimeOut > 0 {
		p.cacheBuffer = make([]utils.CmdRecord, p.cacheInitSize)
	}
	return nil
}

//GetCache get cache data
func (p *Cache) GetCache() (*[]utils.CmdRecord, error) {
	if p.cacheBuffer == nil {
		return nil, errors.New("cachebuffer is nil")
	}
	p.updateCache()
	return &(p.cacheBuffer), nil
}

//ClearCache clear cachebuffer
func (p *Cache) ClearCache() error {
	p.cacheBuffer = nil
	return nil
}

//InputCommand process command string from redis-spy
func (p *Cache) InputCommand(record utils.CmdRecord) error {
	p.updateCache()
	p.cacheBuffer = append(p.cacheBuffer, record)
	return nil
}

func (p *Cache) updateCache() bool {
	curTimeStamp := time.Now().Unix()
	idx := sort.Search(len(p.cacheBuffer), func(i int) bool {
		return p.cacheBuffer[i].TimeStamp >= curTimeStamp-int64(p.cacheTimeOut)
	})
	if idx >= len(p.cacheBuffer) {
		p.cacheBuffer = nil
	} else {
		fmt.Println("%v", p.cacheBuffer)
		fmt.Println("idx:%d, %v", idx, curTimeStamp)
		p.cacheBuffer = p.cacheBuffer[idx:len(p.cacheBuffer)]
		fmt.Println("%v", p.cacheBuffer)
	}
	return true
}

//RecordInst Cache plugin
var RecordInst = Cache{}
