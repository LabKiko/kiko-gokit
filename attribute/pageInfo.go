package attribute

type Page struct {
	PageNo   int64    //页数
	PageSize int64    // 一页多少条
	S        []Sort   // 排序规则
	Field    []string // 返回字段
}
type Sort struct {
	Key  string // 排序的 key
	Desc bool   // 是否升序 ，true 是升序 false 是降序
}
