package global

type GlobalInfo struct {
	Nodename   string
	Consulhost string
}

type GlobalMetadata map[string]string

var Info = GlobalInfo{}
var Metadata = GlobalMetadata{}
