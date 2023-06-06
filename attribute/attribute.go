package attribute

type AttributeDto struct {
	KV []*AttributeKV
}
type AttributeKV struct {
	Key   string
	Value interface{}
}
type BatchAttribute struct {
	BatchKv []*AttributeDto
}
