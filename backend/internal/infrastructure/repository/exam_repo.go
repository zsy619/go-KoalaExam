package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/domain/entity"
)

type ExamRepository struct {
	db *gorm.DB
}

func NewExamRepository(db *gorm.DB) *ExamRepository { return &ExamRepository{db: db} }

func (r *ExamRepository) Create(ctx context.Context, e *entity.Exam) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *ExamRepository) Update(ctx context.Context, e *entity.Exam) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *ExamRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Exam{}, id).Error
}

func (r *ExamRepository) GetByID(ctx context.Context, id int64) (*entity.Exam, error) {
	var e entity.Exam
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExamRepository) List(ctx context.Context, page, size int, status int8, keyword string) ([]entity.Exam, int64, error) {
	var list []entity.Exam
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.Exam{})
	if status > 0 {
		q = q.Where("status = ?", status)
	}
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

func (r *ExamRepository) ListAvailableForUser(ctx context.Context, userID int64) ([]entity.Exam, error) {
	var list []entity.Exam
	err := r.db.WithContext(ctx).Where("status = 1 AND end_time > ?", time.Now()).Find(&list).Error
	return list, err
}

// ExamRecordRepository 考试记录仓储
type ExamRecordRepository struct {
	db *gorm.DB
}

func NewExamRecordRepository(db *gorm.DB) *ExamRecordRepository { return &ExamRecordRepository{db: db} }

func (r *ExamRecordRepository) Create(ctx context.Context, rec *entity.ExamRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *ExamRecordRepository) Update(ctx context.Context, rec *entity.ExamRecord) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *ExamRecordRepository) GetByExamAndUser(ctx context.Context, examID, userID int64) (*entity.ExamRecord, error) {
	var rec entity.ExamRecord
	if err := r.db.WithContext(ctx).Where("exam_id = ? AND user_id = ?", examID, userID).First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *ExamRecordRepository) GetByID(ctx context.Context, id int64) (*entity.ExamRecord, error) {
	var rec entity.ExamRecord
	if err := r.db.WithContext(ctx).First(&rec, id).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *ExamRecordRepository) ListByExam(ctx context.Context, examID int64, page, size int) ([]entity.ExamRecord, int64, error) {
	var list []entity.ExamRecord
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.ExamRecord{}).Where("exam_id = ?", examID)
	q.Count(&total)
	q.Order("total_score DESC").Offset((page - 1) * size).Limit(size).Find(&list)
	return list, total, q.Error
}

func (r *ExamRecordRepository) ListByUser(ctx context.Context, userID int64, page, size int) ([]entity.ExamRecord, int64, error) {
	var list []entity.ExamRecord
	var total int64
	q := r.db.WithContext(ctx).Model(&entity.ExamRecord{}).Where("user_id = ?", userID)
	q.Count(&total)
	q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list)
	return list, total, q.Error
}

// StatsAvgScore 统计平均分
func (r *ExamRecordRepository) StatsAvgScore(ctx context.Context, examID int64) (float64, error) {
	var avg float64
	err := r.db.WithContext(ctx).Model(&entity.ExamRecord{}).
		Where("exam_id = ? AND status >= 2", examID).
		Select("AVG(total_score)").Scan(&avg).Error
	return avg, err
}
