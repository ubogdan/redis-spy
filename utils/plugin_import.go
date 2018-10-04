package utils

import (
	"fmt"
	"plugin"
)

//GetPlgInst get plugin instance from plugin so file
func GetPlgInst(filePath string) (RecordInf, error) {
	plg, err := plugin.Open(filePath)
	if err != nil {
		return nil, err
	}
	symRecordInst, err := plg.Lookup("RecordInst")
	if err != nil {
		return nil, err
	}
	var inst RecordInf
	inst, ok := symRecordInst.(RecordInf)
	if !ok {
		return nil, fmt.Errorf("%+v invalid RecordInf interface", symRecordInst)
	}
	return inst, nil
}
