package utils

//CmdRecord a command record unit
//TimeStamp Unix time
type CmdRecord struct {
	TimeStamp int64
	Command   string
}

//RecordInf interface for meeting command operations monitor
type RecordInf interface {
	Init() error
	InputCommand(record CmdRecord) error
	SetCache(timout int) error
	GetCache() (*[]CmdRecord, error)
	ClearCache() error
}
