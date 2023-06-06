package attribute

type PageResult struct {
	Total int64                    //总页数
	Data  []map[string]interface{} //结果数组
}
type Result struct {
	Id        string
	RowAffect int64
}
type UpdateResult struct {
	Id        []string
	RowAffect int64
}
