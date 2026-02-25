package storage



type DataNode interface {
	isDataNode()
}

type StringType struct {
	Value    string

}

func (st *StringType) isDataNode() {}
