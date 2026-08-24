package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository(db *gorm.DB) *QuestionRepository { return &QuestionRepository{db: db} }

func (r *QuestionRepository) Create(ctx context.Context, q *entity.Question) error {
	return r.db.WithContext(ctx).Create(q).Error
}

func (r *QuestionRepository) BatchCreate(ctx context.Context, qs []entity.Question) error {
	return r.db.WithContext(ctx).CreateInBatches(qs, 100).Error
}

func (r *QuestionRepository) Update(ctx context.Context, q *entity.Question) error {
	return r.db.WithContext(ctx).Save(q).Error
}

func (r *QuestionRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Question{}, id).Error
}

func (r *QuestionRepository) GetByID(ctx context.Context, id int64) (*entity.Question, error) {
	var q entity.Question
	if err := r.db.WithContext(ctx).First(&q, id).Error; err != nil { return nil, err }
	return &q, nil
}

func (r *QuestionRepository) ListByIDs(ctx context.Context, ids []int64) ([]entity.Question, error) {
	var qs []entity.Question
	if len(ids) == 0 { return qs, nil }
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&qs).Error; err != nil { return nil, err }
	return qs, nil
}

func (r *QuestionRepository) List(ctx context.Context, page, size int, categoryID, qtype, difficulty int64, keyword string) ([]entity.Question, int64, error) {
	var list []entity.Question
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.Question{})
	if categoryID > 0 { q = q.Where("category_id = ?", categoryID) }
	if qtype > 0 { q = q.Where("type = ?", qtype) }
	if difficulty > 0 { q = q.Where("difficulty = ?", difficulty) }
	if keyword != "" { q = q.Where("title LIKE ?", "%"+keyword+"%") }
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	if err := q.Order("id DESC").Offset((page-1)*size).Limit(size).Find(&list).Error; err != nil { return nil, 0, err }
	return list, total, nil
}

func (r *QuestionRepository) RandomByTypeAndDifficulty(ctx context.Context, qtype int8, difficulty int8, n int) ([]entity.Question, error) {
	var qs []entity.Question
	err := r.db.WithContext(ctx).Where("type = ? AND difficulty = ? AND status = 1", qtype, difficulty).
		Order("RAND()").Limit(n).Find(&qs).Error
	return qs, err
}

func (r *QuestionRepository) IncStat(ctx context.Context, id int64, correct bool) error {
	updates := map[string]interface{}{"use_count": gorm.Expr("use_count + 1")}
	if correct {
		updates["correct_count"] = gorm.Expr("correct_count + 1")
	} else {
		updates["wrong_count"] = gorm.Expr("wrong_count + 1")
	}
	return r.db.WithContext(ctx).Model(&entity.Question{}).Where("id = ?", id).Updates(updates).Error
}

// CategoryRepository 题库分类仓储
type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository { return &CategoryRepository{db: db} }

func (r *CategoryRepository) List(ctx context.Context) ([]entity.QuestionCategory, error) {
	var list []entity.QuestionCategory
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *CategoryRepository) Create(ctx context.Context, c *entity.QuestionCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepository) Update(ctx context.Context, c *entity.QuestionCategory) error {
	// 使用 map 更新避免 GORM 零值陷阱（特别是 created_at）
	return r.db.WithContext(ctx).Model(c).Updates(map[string]interface{}{
		"parent_id": c.ParentID,
		"name":      c.Name,
		"code":      c.Code,
		"sort":      c.Sort,
	}).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.QuestionCategory{}, id).Error
}
