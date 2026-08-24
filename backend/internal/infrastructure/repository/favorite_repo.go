package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository { return &FavoriteRepository{db: db} }

// Add 添加收藏（已存在则忽略，返回是否新增）
func (r *FavoriteRepository) Add(ctx context.Context, f *entity.Favorite) (bool, error) {
	var existing entity.Favorite
	err := r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", f.UserID, f.TargetType, f.TargetID).First(&existing).Error
	if err == nil {
		// 已存在，更新 folder/note
		existing.FolderID = f.FolderID
		existing.Note = f.Note
		existing.SourceType = f.SourceType
		return false, r.db.WithContext(ctx).Save(&existing).Error
	}
	if err != gorm.ErrRecordNotFound {
		return false, err
	}
	return true, r.db.WithContext(ctx).Create(f).Error
}

// BatchAdd 批量添加（带事务 + Upsert）
func (r *FavoriteRepository) BatchAdd(ctx context.Context, items []entity.Favorite) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, f := range items {
			var existing entity.Favorite
			err := tx.Where("user_id = ? AND target_type = ? AND target_id = ?",
				f.UserID, f.TargetType, f.TargetID).First(&existing).Error
			if err == nil {
				existing.FolderID = f.FolderID
				existing.SourceType = f.SourceType
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			} else if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&f).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return nil
	})
}

// Remove 取消收藏
func (r *FavoriteRepository) Remove(ctx context.Context, userID int64, targetType int8, targetID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, targetType, targetID).Delete(&entity.Favorite{}).Error
}

// IsFavorited 是否已收藏
func (r *FavoriteRepository) IsFavorited(ctx context.Context, userID int64, targetType int8, targetID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Favorite{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).Count(&count).Error
	return count > 0, err
}

// ListByUser 列出用户收藏
func (r *FavoriteRepository) ListByUser(ctx context.Context, userID int64, targetType int8, page, size int) ([]entity.Favorite, int64, error) {
	var list []entity.Favorite
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.Favorite{}).Where("user_id = ?", userID)
	if targetType > 0 {
		q = q.Where("target_type = ?", targetType)
	}
	q.Count(&total)
	q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list)
	return list, total, q.Error
}

// WrongLogRepository 错题日志仓储
type WrongLogRepository struct {
	db *gorm.DB
}

func NewWrongLogRepository(db *gorm.DB) *WrongLogRepository { return &WrongLogRepository{db: db} }

// Upsert 写入或累加错题
func (r *WrongLogRepository) Upsert(ctx context.Context, log *entity.WrongAnswerLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing entity.WrongAnswerLog
		err := tx.Where("user_id = ? AND question_id = ?", log.UserID, log.QuestionID).First(&existing).Error
		if err == nil {
			existing.WrongCount++
			existing.LastWrongAt = log.LastWrongAt
			existing.UserAnswer = log.UserAnswer
			existing.MasteryLevel = log.MasteryLevel
			existing.IsReviewed = false
			return tx.Save(&existing).Error
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		return tx.Create(log).Error
	})
}

// ListByUser 分页查询错题
func (r *WrongLogRepository) ListByUser(ctx context.Context, userID int64, page, size int, masteryLevel int8) ([]entity.WrongAnswerLog, int64, error) {
	var list []entity.WrongAnswerLog
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.WrongAnswerLog{}).Where("user_id = ?", userID)
	if masteryLevel > 0 {
		q = q.Where("mastery_level <= ?", masteryLevel)
	}
	q.Count(&total)
	q.Order("last_wrong_at DESC").Offset((page - 1) * size).Limit(size).Find(&list)
	return list, total, q.Error
}

// MarkReviewed 标记已复习
func (r *WrongLogRepository) MarkReviewed(ctx context.Context, id int64, masteryLevel int8) error {
	return r.db.WithContext(ctx).Model(&entity.WrongAnswerLog{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"is_reviewed": true, "mastery_level": masteryLevel}).Error
}

// FolderRepository 收藏夹仓储
type FolderRepository struct {
	db *gorm.DB
}

func NewFolderRepository(db *gorm.DB) *FolderRepository { return &FolderRepository{db: db} }

func (r *FolderRepository) ListByUser(ctx context.Context, userID int64) ([]entity.FavoriteFolder, error) {
	var list []entity.FavoriteFolder
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *FolderRepository) Create(ctx context.Context, f *entity.FavoriteFolder) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *FolderRepository) Update(ctx context.Context, f *entity.FavoriteFolder) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *FolderRepository) Delete(ctx context.Context, id int64, userID int64) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&entity.FavoriteFolder{}).Error
}

func (r *FolderRepository) GetOrCreateSystemFolder(ctx context.Context, userID int64, name string) (*entity.FavoriteFolder, error) {
	var f entity.FavoriteFolder
	err := r.db.WithContext(ctx).Where("user_id = ? AND name = ? AND is_system = ?", userID, name, true).First(&f).Error
	if err == nil {
		return &f, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	f = entity.FavoriteFolder{UserID: userID, Name: name, IsSystem: true, Color: "#ff7875", Icon: "book"}
	if err := r.db.WithContext(ctx).Create(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}
