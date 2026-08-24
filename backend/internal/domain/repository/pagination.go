package repository

// PageResult 通用分页结果。
// 遵循 Google Go 风格：Result 是个包含列表和总数的值类型。
type PageResult[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// NewPageResult 构造分页结果。
func NewPageResult[T any](list []T, total int64, q PageQuery) PageResult[T] {
	return PageResult[T]{
		List:  list,
		Total: total,
		Page:  q.Page,
		Size:  q.Limit(),
	}
}
