package utils

import (
	"errors"
	"fmt"
	"strings"
)

const (
	plgListSepFlag = ","
	plgSepFlag     = ":"
	plgRoleNum     = 2
)

//PlgRole identify plugin role
type PlgRole string

const (
	//MasterRole for master plugin
	MasterRole PlgRole = "Leader"
	//SlaveRole for slave plugin nnm
	SlaveRole PlgRole = "Follower"
)

//SpyPlugin plugin setting
type SpyPlugin struct {
	Key      PlgRole
	FilePath string
}

//SpyPluginList SpyPlugin list
type SpyPluginList struct {
	SpyPlugins []SpyPlugin
}

//Set pares []PluginConfig
func (spl *SpyPluginList) Set(s string) error {
	spl.SpyPlugins = nil
	plgStrList := strings.Split(s, plgListSepFlag)
	for _, plgStr := range plgStrList {
		var plgConfig SpyPlugin
		err := plgConfig.Set(plgStr)
		if err != nil {
			return err
		}
		spl.SpyPlugins = append(spl.SpyPlugins, plgConfig)
	}
	if len(spl.SpyPlugins) != plgRoleNum {
		return fmt.Errorf("wrong plugin config num[%d]", len(spl.SpyPlugins))
	}
	return nil
}

func (spl *SpyPluginList) String() string {
	var ret string
	for i, e := range spl.SpyPlugins {
		ret = ret + e.String()
		if i < len(spl.SpyPlugins)-1 {
			ret = fmt.Sprintf("%s%s", ret, plgListSepFlag)
		}
	}
	return ret
}

//Set parse <plugin role>:<plugin so file path>
func (p *SpyPlugin) Set(s string) error {
	valSlice := strings.Split(s, plgSepFlag)
	if len(valSlice) != 2 {
		return errors.New("invalid plugin setting")
	}
	if PlgRole(valSlice[0]) == MasterRole || PlgRole(valSlice[0]) == SlaveRole {
		p.Key = PlgRole(valSlice[0])
		p.FilePath = valSlice[1]
	} else {
		return fmt.Errorf("invalid plugin role config[%s]", valSlice[0])
	}
	return nil
}

func (p *SpyPlugin) String() string {
	return fmt.Sprintf("%v%v%v", p.Key, plgSepFlag, p.FilePath)
}

//NewPlgConfig return a plugin setting
func NewPlgConfig(key string, filePath string) *SpyPlugin {
	return &SpyPlugin{
		Key:      PlgRole(key),
		FilePath: filePath}
}
