package dto

// QuestionOption 选项
type QuestionOption struct {
	Key  string `json:"key"` // A/B/C/D
	Text string `json:"text"`
}

// CreateQuestionReq 创建题目
type CreateQuestionReq struct {
	CategoryID int64            `json:"category_id" binding:"required"`
	Type       int8             `json:"type" binding:"required,min=1,max=6"`
	Difficulty int8             `json:"difficulty"`
	Title      string           `json:"title" binding:"required"`
	Options    []QuestionOption `json:"options"`
	Answer     interface{}      `json:"answer" binding:"required"`
	Analysis   string           `json:"analysis"`
	Tags       string           `json:"tags"`
	Score      float64          `json:"score"`
}

// QuestionResp 题目响应（学员端隐藏答案）
type QuestionResp struct {
	ID         int64            `json:"id"`
	Type       int8             `json:"type"`
	Difficulty int8             `json:"difficulty"`
	Title      string           `json:"title"`
	Options    []QuestionOption `json:"options"`
	Score      float64          `json:"score"`
	Answer     interface{}      `json:"answer,omitempty"`
	Analysis   string           `json:"analysis,omitempty"`
}

// BatchImportReq 批量导入题目（Excel）
type BatchImportReq struct {
	CategoryID int64  `json:"category_id" binding:"required"`
	FileURL    string `json:"file_url" binding:"required"`
}
