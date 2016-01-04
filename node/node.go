package node

type NodeInfo struct {
	Name string
	Host string
}

type NodeMetadata map[string]string

var Info = NodeInfo{}
var Metadata = NodeMetadata{}
