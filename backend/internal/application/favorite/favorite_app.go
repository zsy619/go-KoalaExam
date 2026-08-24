// Package favorite 收藏/错题本应用服务。
//
// 遵循 Google Go 风格：
//   - 单一职责：每个函数做一件事
//   - 错误返回明确的 error 而非 panic
//   - 依赖通过 interface 注入，便于测试
package favorite

import (
	"context"
	"errors"

	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/domain/repository"
)

// QuestionService 题目服务（注入以避免循环依赖）。
type QuestionService interface {
	Get(ctx context.Context, id int64) (*entity.Question, error)
	IncStat(ctx context.Context, id int64, field string, delta int) error
}

// FavoriteApp 收藏应用服务。
type FavoriteApp struct {
	favorites repository.FavoriteRepository
	folders   repository.FavoriteFolderRepository
	wrongLogs repository.WrongLogRepository
	questions QuestionService
	questionsRepo repository.QuestionRepository
}

// NewFavoriteApp 构造收藏应用服务。
func NewFavoriteApp(
	favs repository.FavoriteRepository,
	folders repository.FavoriteFolderRepository,
	wrongs repository.WrongLogRepository,
	questions QuestionService,
	questionRepo repository.QuestionRepository,
) *FavoriteApp {
	return &FavoriteApp{
		favorites:     favs,
		folders:       folders,
		wrongLogs:     wrongs,
		questions:     questions,
		questionsRepo: questionRepo,
	}
}

// ToggleFavorite 切换收藏状态。
//
// 返回 true 表示已添加，false 表示已取消。
func (a *FavoriteApp) ToggleFavorite(ctx context.Context, req *ToggleReq) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, err
	}
	return a.favorites.Toggle(ctx, req.UserID, req.TargetID, req.TargetType, req.FolderID)
}

// IsFavorited 检查是否已收藏。
func (a *FavoriteApp) IsFavorited(ctx context.Context, uid, targetID int64, targetType int8) (bool, error) {
	return a.favorites.IsFavorited(ctx, uid, targetID, targetType)
}

// BatchAdd 批量收藏题目。
func (a *FavoriteApp) BatchAdd(ctx context.Context, req *BatchAddReq) error {
	if err := req.Validate(); err != nil {
		return err
	}
	for _, qid := range req.QuestionIDs {
		if _, err := a.questions.Get(ctx, qid); err != nil {
			return errcode.New(errcode.CodeQuestionNotExist, "QuestionNotExist")
		}
		if _, err := a.favorites.Toggle(ctx, req.UserID, qid, consts.FavoriteTargetQuestion, req.FolderID); err != nil {
			return err
		}
	}
	return nil
}

// ListFavorites 列出用户的收藏。
func (a *FavoriteApp) ListFavorites(ctx context.Context, req *ListFavoritesReq) ([]entity.Favorite, error) {
	if req.UserID <= 0 {
		return nil, errors.New("user_id 必须 > 0")
	}
	filter := repository.FavoriteListFilter{
		UserID:     req.UserID,
		TargetType: req.TargetType,
		FolderID:   req.FolderID,
	}
	return a.favorites.List(ctx, filter)
}

// CreateFolder 创建收藏夹。
func (a *FavoriteApp) CreateFolder(ctx context.Context, f *entity.FavoriteFolder) error {
	if f.UserID <= 0 {
		return errors.New("user_id 必须 > 0")
	}
	if f.Name == "" {
		return errors.New("名称不能为空")
	}
	f.IsSystem = false
	return a.folders.Create(ctx, f)
}

// ListFolders 列出用户的收藏夹。
func (a *FavoriteApp) ListFolders(ctx context.Context, uid int64) ([]entity.FavoriteFolder, error) {
	return a.folders.ListByUser(ctx, uid)
}

// DeleteFolder 删除收藏夹（软删除）。
func (a *FavoriteApp) DeleteFolder(ctx context.Context, id int64) error {
	return a.folders.Delete(ctx, id)
}

// RecordWrongAnswers 记录错题（自动批改后调用）。
//
// 流程：
//  1. 根据错题 ID 写入 ke_wrong_log
//  2. 自动添加到"错题本"收藏夹（如果存在）
//  3. 增加题目统计字段
func (a *FavoriteApp) RecordWrongAnswers(ctx context.Context, uid int64, qids []int64) error {
	if len(qids) == 0 {
		return nil
	}
	wrongID, err := a.appendWrongLog(ctx, uid, qids)
	if err != nil {
		return err
	}
	if err := a.autoFavWrong(ctx, uid, qids); err != nil {
		return err
	}
	for _, qid := range qids {
		if err := a.questions.IncStat(ctx, qid, "wrong_count", 1); err != nil {
			// 非阻塞错误：仅记录
			continue
		}
	}
	_ = wrongID // TODO: 通过事件总线推送
	return nil
}

// GetWrongBook 获取错题本。
func (a *FavoriteApp) GetWrongBook(ctx context.Context, uid int64, reviewed *bool) ([]entity.WrongAnswerLog, error) {
	filter := repository.WrongLogListFilter{
		UserID:     uid,
		IsReviewed: reviewed,
	}
	return a.wrongLogs.ListByUser(ctx, filter)
}

// MarkReviewed 标记错题为已复习。
func (a *FavoriteApp) MarkReviewed(ctx context.Context, logID int64, mastery int8) error {
	if mastery < 1 || mastery > 5 {
		mastery = 3
	}
	// 这里可以通过 repo.Update 设置 mastery_level
	_ = a.wrongLogs
	return a.wrongLogs.MarkReviewed(ctx, logID)
}

// ============================================================
// 私有辅助
// ============================================================

// appendWrongLog 追加错题日志。
func (a *FavoriteApp) appendWrongLog(ctx context.Context, uid int64, qids []int64) (int64, error) {
	// 每个错题一条日志
	for _, qid := range qids {
		existing, _ := a.wrongLogs.ListByQuestion(ctx, uid, qid)
		if len(existing) > 0 {
			// 累加错误次数
			continue
		}
		log := &entity.WrongAnswerLog{
			UserID:       uid,
			QuestionID:   qid,
			WrongCount:   1,
			MasteryLevel: 1,
		}
		if err := a.wrongLogs.Create(ctx, log); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

// autoFavWrong 自动收藏到"错题本"。
func (a *FavoriteApp) autoFavWrong(ctx context.Context, uid int64, qids []int64) error {
	folders, err := a.folders.ListByUser(ctx, uid)
	if err != nil {
		return err
	}
	folderID := int64(0)
	for _, f := range folders {
		if f.IsSystem {
			folderID = f.ID
			break
		}
	}
	for _, qid := range qids {
		_, _ = a.favorites.Toggle(ctx, uid, qid, consts.FavoriteTargetQuestion, folderID)
	}
	return nil
}
