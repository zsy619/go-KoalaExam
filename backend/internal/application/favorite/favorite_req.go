package favorite

import "errors"

// ToggleReq 切换收藏请求。
type ToggleReq struct {
	UserID     int64 `json:"user_id"`
	TargetType int8  `json:"target_type"`
	TargetID   int64 `json:"target_id"`
	FolderID   int64 `json:"folder_id"`
}

// Validate 校验。
func (r *ToggleReq) Validate() error {
	if r.UserID <= 0 {
		return errors.New("user_id 必须 > 0")
	}
	if r.TargetID <= 0 {
		return errors.New("target_id 必须 > 0")
	}
	if r.TargetType < 1 || r.TargetType > 3 {
		return errors.New("target_type 必须为 1/2/3")
	}
	return nil
}

// BatchAddReq 批量收藏请求。
type BatchAddReq struct {
	UserID      int64   `json:"user_id"`
	QuestionIDs []int64 `json:"question_ids"`
	FolderID    int64   `json:"folder_id"`
}

// Validate 校验。
func (r *BatchAddReq) Validate() error {
	if r.UserID <= 0 {
		return errors.New("user_id 必须 > 0")
	}
	if len(r.QuestionIDs) == 0 {
		return errors.New("question_ids 不能为空")
	}
	if len(r.QuestionIDs) > 100 {
		return errors.New("单次最多收藏 100 题")
	}
	return nil
}

// ListFavoritesReq 列表请求。
type ListFavoritesReq struct {
	UserID     int64 `json:"user_id"`
	TargetType int8  `json:"target_type"`
	FolderID   int64 `json:"folder_id"`
}
