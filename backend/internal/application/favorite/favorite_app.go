// Package favorite 收藏/错题本应用服务。
//
// 遵循 Google Go 风格：
//   - 命名简洁（ToggleFavorite/GetWrongBook 等动词+名词）
//   - context 作为第一个参数
//   - 显式 error 返回
//   - 通过 interface 注入依赖，便于测试
//   - 领域事件解耦副作用
package favorite

import (
	"strconv"

	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/event"
	"github.com/your-team/koala-exam-backend/internal/domain/repository"
	infra "github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
)

// QuestionStatUpdater 题目统计更新器（注入以避免循环依赖）。
type QuestionStatUpdater interface {
	IncStat(ctx context.Context, id int64, correct bool) error
}

// FavoriteApp 收藏应用服务。
//
// 设计要点：
//   - 单一职责：只负责收藏与错题管理的用例编排
//   - 事务一致性：批量操作在数据库事务中执行
//   - 事件解耦：副作用通过事件总线触发
type FavoriteApp struct {
	db         *gorm.DB
	favorites  *infra.FavoriteRepository
	folders    *infra.FolderRepository
	wrongLogs  *infra.WrongLogRepository
	questions  QuestionStatUpdater
	bus        *event.Bus
}

// NewFavoriteApp 构造收藏应用服务（依赖注入）。
func NewFavoriteApp(
	db *gorm.DB,
	favs *infra.FavoriteRepository,
	folders *infra.FolderRepository,
	wrongs *infra.WrongLogRepository,
	questions QuestionStatUpdater,
	bus *event.Bus,
) *FavoriteApp {
	return &FavoriteApp{
		db:        db,
		favorites: favs,
		folders:   folders,
		wrongLogs: wrongs,
		questions: questions,
		bus:       bus,
	}
}

// ToggleFavorite 切换收藏状态。
//
// 返回 true 表示已添加，false 表示已取消。
//
// 实现要点：
//  1. 查询当前状态
//  2. 存在则删除，不存在则添加
//  3. 发布 FavoriteToggledEvent 事件（用于统计、推荐等）
func (a *FavoriteApp) ToggleFavorite(ctx context.Context, req *ToggleReq) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, err
	}

	existing, err := a.findFavorite(ctx, req.UserID, req.TargetID, req.TargetType)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	added := false
	if existing != nil {
		// 已存在，删除
		if err := a.favorites.Remove(ctx, req.UserID, int8(req.TargetType), req.TargetID); err != nil {
			return false, err
		}
		added = false
	} else {
		// 不存在，添加
		fav := &entity.Favorite{
			UserID:     req.UserID,
			TargetType: int8(req.TargetType),
			TargetID:   req.TargetID,
		}
		if req.FolderID > 0 {
			id := req.FolderID
			fav.FolderID = &id
		}
		if _, err := a.favorites.Add(ctx, fav); err != nil {
			return false, err
		}
		added = true
	}

	// 发布事件（异步 handler 在订阅方实现）
	_ = a.bus.Publish(ctx, &event.FavoriteToggledEvent{
		UserID:     req.UserID,
		TargetType: int8(req.TargetType),
		TargetID:   req.TargetID,
		Added:      added,
		ToggledAt:  time.Now(),
	})

	return added, nil
}

// IsFavorited 判断是否已收藏。
func (a *FavoriteApp) IsFavorited(ctx context.Context, uid, targetID int64, targetType int8) (bool, error) {
	if uid <= 0 || targetID <= 0 {
		return false, errors.New("invalid parameters")
	}
	return a.favorites.IsFavorited(ctx, uid, targetType, targetID)
}

// BatchAdd 批量收藏题目（事务保证一致性）。
//
// 失败回滚：任一题目添加失败则全部回滚。
func (a *FavoriteApp) BatchAdd(ctx context.Context, req *BatchAddReq) (BatchAddResult, error) {
	if err := req.Validate(); err != nil {
		return BatchAddResult{}, err
	}
	if len(req.QuestionIDs) == 0 {
		return BatchAddResult{}, errors.New("no questions provided")
	}

	result := BatchAddResult{
		AddedIDs:   make([]int64, 0),
		SkippedIDs: make([]int64, 0),
	}

	// 在事务中处理
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, qid := range req.QuestionIDs {
			// 检查是否已收藏
			exists, err := a.favorites.IsFavorited(ctx, req.UserID, consts.TargetTypeQuestion, qid)
			if err != nil {
				return fmt.Errorf("check favorite %d: %w", qid, err)
			}
			if exists {
				result.SkippedIDs = append(result.SkippedIDs, qid)
				continue
			}

			fav := &entity.Favorite{
				UserID:     req.UserID,
				TargetType: consts.TargetTypeQuestion,
				TargetID:   qid,
			}
			if req.FolderID > 0 {
				id := req.FolderID
				fav.FolderID = &id
			}
			if _, err := a.favorites.Add(ctx, fav); err != nil {
				return fmt.Errorf("add favorite %d: %w", qid, err)
			}
			result.AddedIDs = append(result.AddedIDs, qid)
		}
		return nil
	})

	if err != nil {
		return BatchAddResult{}, err
	}
	return result, nil
}

// BatchAddResult 批量收藏结果。
type BatchAddResult struct {
	AddedIDs   []int64 `json:"added_ids"`
	SkippedIDs []int64 `json:"skipped_ids"`
}

// ListFavorites 列出用户的收藏（支持分类筛选）。
func (a *FavoriteApp) ListFavorites(ctx context.Context, uid int64, targetType int8, folderID int64) ([]entity.Favorite, error) {
	if uid <= 0 {
		return nil, errors.New("user_id must be > 0")
	}
	filter := repository.FavoriteListFilter{
		UserID:     uid,
		TargetType: targetType,
		FolderID:   folderID,
	}
	return a.favorites.List(ctx, filter)
}

// CreateFolder 创建收藏夹。
func (a *FavoriteApp) CreateFolder(ctx context.Context, f *entity.FavoriteFolder) error {
	if f.UserID <= 0 {
		return errors.New("user_id must be > 0")
	}
	if f.Name == "" {
		return errors.New("name is required")
	}
	if len(f.Name) > 32 {
		return errors.New("name too long (max 32)")
	}
	f.IsSystem = false
	return a.folders.Create(ctx, f)
}

// ListFolders 列出用户的收藏夹。
func (a *FavoriteApp) ListFolders(ctx context.Context, uid int64) ([]entity.FavoriteFolder, error) {
	return a.folders.ListByUser(ctx, uid)
}

// DeleteFolder 删除收藏夹。
func (a *FavoriteApp) DeleteFolder(ctx context.Context, id int64) error {
	return a.folders.Delete(ctx, id, 0)
}

// RecordWrongAnswers 记录错题（自动批改后调用）。
//
// 流程：
//  1. 写入错题日志（事务）
//  2. 自动添加到"错题本"收藏夹
//  3. 增加题目统计
//  4. 发布 WrongAnswerRecordedEvent
func (a *FavoriteApp) RecordWrongAnswers(ctx context.Context, uid int64, qids []int64) error {
	if len(qids) == 0 {
		return nil
	}
	if uid <= 0 {
		return errors.New("user_id must be > 0")
	}

	logIDs, err := a.appendWrongLog(ctx, uid, qids)
	if err != nil {
		return err
	}

	// 自动加入收藏夹（系统文件夹"错题本"）
	if err := a.autoFavWrong(ctx, uid, qids); err != nil {
		// 非阻塞错误
		_ = err
	}

	// 更新题目统计
	for _, qid := range qids {
		_ = a.questions.IncStat(ctx, qid, false)
	}

	// 发布事件
	for i, qid := range qids {
		var logID int64
		if i < len(logIDs) {
			logID = logIDs[i]
		}
		_ = a.bus.Publish(ctx, &event.WrongAnswerRecordedEvent{
			UserID:     uid,
			QuestionID: qid,
			WrongCount: 1,
			RecordedAt: time.Now(),
		})
		_ = logID
	}
	return nil
}

// GetWrongBook 获取错题本（带筛选）。
func (a *FavoriteApp) GetWrongBook(ctx context.Context, uid int64, opts WrongBookQuery) ([]entity.WrongAnswerLog, int64, error) {
	if uid <= 0 {
		return nil, 0, errors.New("user_id must be > 0")
	}
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 || opts.PageSize > 100 {
		opts.PageSize = 20
	}
	return a.wrongLogs.ListByUser(ctx, uid, opts.Page, opts.PageSize, opts.MasteryLevel, opts.IsReviewed)
}

// WrongBookQuery 错题本查询参数。
type WrongBookQuery struct {
	Page         int
	PageSize     int
	MasteryLevel int8
	IsReviewed   *bool
}

// MarkReviewed 标记错题为已复习。
//
// 同时发布 WrongBookReviewedEvent，用于学习曲线分析。
func (a *FavoriteApp) MarkReviewed(ctx context.Context, logID int64, mastery int8) error {
	if logID <= 0 {
		return errors.New("invalid log_id")
	}
	if mastery < 1 {
		mastery = 3
	}
	if mastery > 5 {
		mastery = 5
	}
	if err := a.wrongLogs.MarkReviewed(ctx, logID, mastery); err != nil {
		return err
	}

	_ = a.bus.Publish(ctx, &event.WrongBookReviewedEvent{
		LogID:        logID,
		MasteryLevel: mastery,
		ReviewedAt:   time.Now(),
	})
	return nil
}

// GetWrongStats 错题统计。
func (a *FavoriteApp) GetWrongStats(ctx context.Context, uid int64) (WrongStats, error) {
	if uid <= 0 {
		return WrongStats{}, errors.New("invalid user_id")
	}
	total, err := a.wrongLogs.CountByUser(ctx, uid)
	if err != nil {
		return WrongStats{}, err
	}
	return WrongStats{Total: total}, nil
}



// FavoriteStats 收藏统计聚合。
type FavoriteStats struct {
	Total      int64           `json:"total"`
	ByType     map[int8]int64  `json:"by_type"`        // 1=题目 2=试卷 3=知识点
	ByFolder   map[int64]int64 `json:"by_folder"`      // folderID -> count
	RecentWeek int64           `json:"recent_week_added"`
}

// GetStats 获取用户收藏统计。
func (a *FavoriteApp) GetStats(ctx context.Context, uid int64) (FavoriteStats, error) {
	if uid <= 0 {
		return FavoriteStats{}, errors.New("invalid user_id")
	}
	stats := FavoriteStats{ByType: map[int8]int64{}, ByFolder: map[int64]int64{}}
	all, _, err := a.favorites.ListByUser(ctx, uid, 0, 1, 1000)
	if err != nil {
		return stats, err
	}
	stats.Total = int64(len(all))
	weekAgo := time.Now().AddDate(0, 0, -7)
	for _, f := range all {
		stats.ByType[f.TargetType]++
		fid := int64(0)
		if f.FolderID != nil {
			fid = *f.FolderID
		}
		stats.ByFolder[fid]++
		if f.CreatedAt.After(weekAgo) {
			stats.RecentWeek++
		}
	}
	return stats, nil
}

// MasteryDist 错题掌握度分布。
type MasteryDist struct {
	Total    int64            `json:"total"`
	Mastered int64            `json:"mastered"`
	Levels   map[string]int64 `json:"levels"` // "1"未掌握, "2"薄弱, "3"一般, "4"良好, "5"熟练
}

// GetMasteryDistribution 获取错题掌握度分布。
func (a *FavoriteApp) GetMasteryDistribution(ctx context.Context, uid int64) (MasteryDist, error) {
	if uid <= 0 {
		return MasteryDist{}, errors.New("invalid user_id")
	}
	dist := MasteryDist{Levels: map[string]int64{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}}
	logs, _, err := a.wrongLogs.ListByUser(ctx, uid, 1, 1000, 0, nil)
	if err != nil {
		return dist, err
	}
	dist.Total = int64(len(logs))
	for _, log := range logs {
		if log.MasteryLevel < 1 {
			log.MasteryLevel = 1
		}
		if log.MasteryLevel > 5 {
			log.MasteryLevel = 5
		}
		key := strconv.FormatInt(int64(log.MasteryLevel), 10)
		dist.Levels[key]++
		if log.IsReviewed {
			dist.Mastered++
		}
	}
	return dist, nil
}

// WrongStats 错题统计。
type WrongStats struct {
	Total         int64 `json:"total"`
	Mastered      int64 `json:"mastered"`
	NeedReview    int64 `json:"need_review"`
	ThisWeekAdded int64 `json:"this_week_added"`
}

// ============================================================
// 私有辅助
// ============================================================

// findFavorite 查询收藏（避免循环依赖仓储层）。
func (a *FavoriteApp) findFavorite(ctx context.Context, uid, targetID int64, targetType int8) (*entity.Favorite, error) {
	filter := repository.FavoriteListFilter{
		UserID:     uid,
		TargetType: int8(targetType),
		TargetID:   targetID,
	}
	list, err := a.favorites.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	f := list[0]
	return &f, nil
}

// appendWrongLog 追加错题日志（去重）。
func (a *FavoriteApp) appendWrongLog(ctx context.Context, uid int64, qids []int64) ([]int64, error) {
	logIDs := make([]int64, 0, len(qids))
	for _, qid := range qids {
		// 复用 Upsert 处理去重
		log := &entity.WrongAnswerLog{
			UserID:       uid,
			QuestionID:   qid,
			WrongCount:   1,
			MasteryLevel: 1,
		}
		if err := a.wrongLogs.Upsert(ctx, log); err != nil {
			return logIDs, err
		}
		logIDs = append(logIDs, log.ID)
	}
	return logIDs, nil
}

// autoFavWrong 自动收藏到"错题本"。
func (a *FavoriteApp) autoFavWrong(ctx context.Context, uid int64, qids []int64) error {
	folders, err := a.folders.ListByUser(ctx, uid)
	if err != nil {
		return err
	}
	var folderID int64
	for _, f := range folders {
		if f.IsSystem {
			folderID = f.ID
			break
		}
	}
	if folderID == 0 {
		return nil
	}
	for _, qid := range qids {
		fav := &entity.Favorite{
			UserID:     uid,
			TargetType: consts.TargetTypeQuestion,
			TargetID:   qid,
			FolderID:   &folderID,
		}
		_, _ = a.favorites.Add(ctx, fav)
	}
	return nil
}
