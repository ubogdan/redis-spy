package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	addrSepFlag     = ":"
	addrListSepFlag = ","
)

//NetAddr <ip, port>
type NetAddr struct {
	ip   string
	port uint16
}

//NetAddrList []NetAddr
type NetAddrList struct {
	NetAddrs []NetAddr
}

//Set parse <netaddr>,<netaddr>
func (rp *NetAddrList) Set(s string) error {
	varSlice := strings.Split(s, addrListSepFlag)
	for _, netAddrStr := range varSlice {
		var redisAddrInst NetAddr
		err := redisAddrInst.Set(netAddrStr)
		if err != nil {
			return err
		}
		rp.NetAddrs = append(rp.NetAddrs, redisAddrInst)
	}
	return nil
}

func (rp *NetAddrList) String() string {
	return fmt.Sprintf("%+v", *rp)
}

//Set parse <ip>:<port>
func (p *NetAddr) Set(s string) error {
	valSlice := strings.Split(s, addrSepFlag)
	if len(valSlice) != 2 {
		return errors.New("invalid netAddr")
	}
	p.ip = valSlice[0]
	port, err := strconv.ParseUint(valSlice[1], 10, 16)
	if err != nil {
		return errors.New("invalid netAddr")
	}
	p.port = uint16(port)
	return nil
}

func (p *NetAddr) String() string {
	return fmt.Sprintf("%v%v%v", p.ip, addrSepFlag, p.port)
}

//New return a NetAddr instance
func New(ipStr string, port uint16) *NetAddr {
	return &NetAddr{
		ip:   ipStr,
		port: port}
}
