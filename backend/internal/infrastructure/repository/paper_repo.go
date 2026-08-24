package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

type PaperRepository struct {
	db *gorm.DB
}

func NewPaperRepository(db *gorm.DB) *PaperRepository { return &PaperRepository{db: db} }

func (r *PaperRepository) Create(ctx context.Context, p *entity.Paper) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PaperRepository) Update(ctx context.Context, p *entity.Paper) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *PaperRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Paper{}, id).Error
}

func (r *PaperRepository) GetByID(ctx context.Context, id int64) (*entity.Paper, error) {
	var p entity.Paper
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaperRepository) List(ctx context.Context, page, size int, keyword string) ([]entity.Paper, int64, error) {
	var list []entity.Paper
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.Paper{})
	if keyword != "" {
		q = q.Where("title LIKE ?", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AddQuestions 关联题目到试卷
func (r *PaperRepository) AddQuestions(ctx context.Context, paperID int64, items []entity.PaperQuestion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("paper_id = ?", paperID).Delete(&entity.PaperQuestion{}).Error; err != nil {
			return err
		}
		return tx.CreateInBatches(items, 100).Error
	})
}

func (r *PaperRepository) GetQuestionsByPaper(ctx context.Context, paperID int64) ([]entity.PaperQuestion, error) {
	var list []entity.PaperQuestion
	err := r.db.WithContext(ctx).Where("paper_id = ?", paperID).Order("sort ASC").Find(&list).Error
	return list, err
}
