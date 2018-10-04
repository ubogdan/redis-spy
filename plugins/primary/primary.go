package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/europelee/redis-spy/utils"
)

const (
	fileDefaultLineSize = 10240
	saveDefaultDir      = "/tmp/primary"
	datLineSep          = ","
)

//Primary default for master redis-py
type Primary struct {
	fileLineSize int
	curLineNum   int
	saveDir      string
	filePrefix   string
	fileHandler  *os.File
	utils.RecordInf
}

func init() {

}

//Init initial some config
func (p *Primary) Init() error {
	p.fileLineSize = fileDefaultLineSize
	p.curLineNum = 0
	p.fileHandler = nil
	p.saveDir = saveDefaultDir
	p.filePrefix = "primary"
	err := os.MkdirAll(p.saveDir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

//SetCache set cache config
func (p *Primary) SetCache(timeout int) error {
	return nil
}

//GetCache get cache data
func (p *Primary) GetCache() (*[]utils.CmdRecord, error) {
	return nil, nil
}

//ClearCache clear cachebuffer
func (p *Primary) ClearCache() error {
	if p.fileHandler == nil {
		return nil
	}
	if err := p.fileHandler.Close(); err != nil {
		log.Fatal(err)
	}
	p.curLineNum = 0
	p.fileHandler = nil
	return nil
}

//GetFileHandler get file handler
func (p *Primary) GetFileHandler() (*os.File, error) {
	if p.curLineNum > p.fileLineSize {
		p.ClearCache()
	}
	if p.fileHandler == nil {
		ts := time.Now().UnixNano() / int64(time.Millisecond)
		var err error
		fileName := fmt.Sprintf("%s/%s-%v.dat", p.saveDir, p.filePrefix, ts)
		p.fileHandler, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
	return p.fileHandler, nil
}

//InputCommand process command string from redis-spy
func (p *Primary) InputCommand(record utils.CmdRecord) error {
	line := fmt.Sprintf("%v%s%v", record.TimeStamp, datLineSep, record.Command)
	fHdl, err := p.GetFileHandler()
	if err != nil {
		return err
	}
	_, err = fHdl.WriteString(line)
	if err != nil {
		return err
	}
	p.curLineNum++
	return nil
}

//RecordInst primary record plugin
var RecordInst = Primary{}
