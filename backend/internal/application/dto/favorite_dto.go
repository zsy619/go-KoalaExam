package dto

// FavoriteReq 单个收藏/取消
type FavoriteReq struct {
	TargetType int8   `json:"target_type" binding:"required,min=1,max=3"`
	TargetID   int64  `json:"target_id" binding:"required"`
	FolderID   *int64 `json:"folder_id"`
	Note       string `json:"note"`
}

// BatchFavoriteReq 批量收藏（错题自动入库）
type BatchFavoriteReq struct {
	TargetType int8    `json:"target_type" binding:"required"`
	TargetIDs  []int64 `json:"target_ids" binding:"required"`
	FolderID   *int64  `json:"folder_id"`
	SourceType int8    `json:"source_type"`
}

// CreateFolderReq 创建收藏夹
type CreateFolderReq struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// WrongQuestionItem 错题本条目（联合查询）
type WrongQuestionItem struct {
	LogID         int64        `json:"log_id"`
	Question      QuestionResp `json:"question"`
	WrongCount    int          `json:"wrong_count"`
	LastWrongAt   string       `json:"last_wrong_at"`
	IsReviewed    bool         `json:"is_reviewed"`
	MasteryLevel  int8         `json:"mastery_level"`
	UserAnswer    interface{}  `json:"user_answer"`
	CorrectAnswer interface{}  `json:"correct_answer"`
}
