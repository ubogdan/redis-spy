package utils

import "fmt"

//CmdRecord interface for meeting command operations monitor
type CmdRecord struct {
	RecordType string
}

//InputCommand process command operation
func (p *CmdRecord) InputCommand(record string) error {
	fmt.Println(record)
	return nil
}
