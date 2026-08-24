package question

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
)

// QuestionApp 题目应用服务
type QuestionApp struct {
	repo    *repository.QuestionRepository
	catRepo *repository.CategoryRepository
}

func NewQuestionApp(r *repository.QuestionRepository, c *repository.CategoryRepository) *QuestionApp {
	return &QuestionApp{repo: r, catRepo: c}
}

// Create 创建题目
func (a *QuestionApp) Create(ctx context.Context, req *dto.CreateQuestionReq, creatorID int64) (int64, error) {
	q := &entity.Question{
		CategoryID: req.CategoryID, Type: req.Type, Difficulty: req.Difficulty,
		Title: req.Title, Tags: req.Tags, Score: req.Score, Analysis: req.Analysis,
		CreatorID: creatorID, Status: 1,
	}
	opts, _ := json.Marshal(req.Options)
	q.Options = string(opts)
	ans, _ := json.Marshal(req.Answer)
	q.Answer = string(ans)
	if err := a.repo.Create(ctx, q); err != nil {
		return 0, err
	}
	return q.ID, nil
}

// Update
func (a *QuestionApp) Update(ctx context.Context, id int64, req *dto.CreateQuestionReq) error {
	old, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return errcode.New(errcode.CodeQuestionNotExist, "QuestionNotExist")
	}
	old.CategoryID = req.CategoryID
	old.Type = req.Type
	old.Difficulty = req.Difficulty
	old.Title = req.Title
	old.Tags = req.Tags
	old.Score = req.Score
	old.Analysis = req.Analysis
	opts, _ := json.Marshal(req.Options)
	old.Options = string(opts)
	ans, _ := json.Marshal(req.Answer)
	old.Answer = string(ans)
	return a.repo.Update(ctx, old)
}

// Delete
func (a *QuestionApp) Delete(ctx context.Context, id int64) error {
	return a.repo.Delete(ctx, id)
}

// Get 获取题目详情
func (a *QuestionApp) Get(ctx context.Context, id int64) (*dto.QuestionResp, error) {
	q, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errcode.New(errcode.CodeQuestionNotExist, "QuestionNotExist")
	}
	return toResp(q, true), nil
}

// List 列表
func (a *QuestionApp) List(ctx context.Context, page, size int, categoryID, qtype, difficulty int64, keyword string) ([]entity.Question, int64, error) {
	return a.repo.List(ctx, page, size, categoryID, qtype, difficulty, keyword)
}

// BatchImport 批量导入（Excel）— 实际可调用 excelize 解析
func (a *QuestionApp) BatchImport(ctx context.Context, req *dto.BatchImportReq, creatorID int64) (int, error) {
	// TODO: 调用 excelize 解析 Excel，按行构建 Question
	// 这里返回 0 表示未实现，请使用 ListCreate 手动传入
	return 0, errors.New("batch import: please use frontend to parse xlsx, then call ListCreate")
}

// ListCreate 批量创建（前端解析后调用）
func (a *QuestionApp) ListCreate(ctx context.Context, items []dto.CreateQuestionReq, creatorID int64) (int, error) {
	qs := make([]entity.Question, 0, len(items))
	for _, req := range items {
		opts, _ := json.Marshal(req.Options)
		ans, _ := json.Marshal(req.Answer)
		qs = append(qs, entity.Question{
			CategoryID: req.CategoryID, Type: req.Type, Difficulty: req.Difficulty,
			Title: req.Title, Tags: req.Tags, Score: req.Score, Analysis: req.Analysis,
			Options: string(opts), Answer: string(ans),
			CreatorID: creatorID, Status: 1,
		})
	}
	if err := a.repo.BatchCreate(ctx, qs); err != nil {
		return 0, err
	}
	return len(qs), nil
}

// ListCategories
func (a *QuestionApp) ListCategories(ctx context.Context) ([]entity.QuestionCategory, error) {
	return a.catRepo.List(ctx)
}

// CreateCategory
func (a *QuestionApp) CreateCategory(ctx context.Context, c *entity.QuestionCategory) error {
	return a.catRepo.Create(ctx, c)
}


// UpdateCategory 更新分类。
func (a *QuestionApp) UpdateCategory(ctx context.Context, cat *entity.QuestionCategory) error {
	return a.catRepo.Update(ctx, cat)
}

// DeleteCategory 删除分类。
func (a *QuestionApp) DeleteCategory(ctx context.Context, id int64) error {
	return a.catRepo.Delete(ctx, id)
}

// toResp 转换为响应 DTO（学员端隐藏答案）
func toResp(q *entity.Question, withAnswer bool) *dto.QuestionResp {
	var opts []dto.QuestionOption
	if q.Options != "" {
		_ = json.Unmarshal([]byte(q.Options), &opts)
	}
	var ans interface{}
	if withAnswer && q.Answer != "" {
		_ = json.Unmarshal([]byte(q.Answer), &ans)
	}
	return &dto.QuestionResp{
		ID: q.ID, Type: q.Type, Difficulty: q.Difficulty, Title: q.Title,
		Options: opts, Score: q.Score, Answer: ans, Analysis: q.Analysis,
	}
}

// ToResp 暴露给外部
func ToResp(q *entity.Question, withAnswer bool) *dto.QuestionResp {
	return toResp(q, withAnswer)
}
