package question

import (
	"context"
	"encoding/json"

	"github.com/your-team/koala-exam-backend/internal/application/dto"
	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/errcode"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
)

// PaperApp 试卷应用服务
type PaperApp struct {
	paperRepo *repository.PaperRepository
	qRepo     *repository.QuestionRepository
}

func NewPaperApp(p *repository.PaperRepository, q *repository.QuestionRepository) *PaperApp {
	return &PaperApp{paperRepo: p, qRepo: q}
}

// Create 创建试卷
func (a *PaperApp) Create(ctx context.Context, req *dto.CreatePaperReq, creatorID int64) (int64, error) {
	p := &entity.Paper{
		Title: req.Title, Description: req.Description, Strategy: req.Strategy,
		TotalScore: req.TotalScore, Duration: req.Duration, PassScore: req.PassScore,
		CreatorID: creatorID, Status: 1,
	}

	switch req.Strategy {
	case consts.StrategyFixed:
		if len(req.QuestionIDs) == 0 {
			return 0, errcode.New(errcode.CodeQuestionEmpty, "QuestionEmpty")
		}
		ids, _ := json.Marshal(req.QuestionIDs)
		p.QuestionIDs = string(ids)
		// 初始化 ConfigRule 为合法 JSON
		p.ConfigRule = "{}"
	case consts.StrategyRandom, consts.StrategyGA:
		cfg, _ := json.Marshal(req.ConfigRule)
		p.ConfigRule = string(cfg)
	}

	if err := a.paperRepo.Create(ctx, p); err != nil {
		return 0, err
	}
	if req.Strategy == consts.StrategyFixed && len(req.QuestionIDs) > 0 {
		items := make([]entity.PaperQuestion, 0, len(req.QuestionIDs))
		qs, _ := a.qRepo.ListByIDs(ctx, req.QuestionIDs)
		qmap := map[int64]entity.Question{}
		for _, q := range qs {
			qmap[q.ID] = q
		}
		for i, qid := range req.QuestionIDs {
			score := 1.0
			if q, ok := qmap[qid]; ok && q.Score > 0 {
				score = q.Score
			}
			items = append(items, entity.PaperQuestion{
				PaperID: p.ID, QuestionID: qid, Sort: i, Score: score,
			})
		}
		if err := a.paperRepo.AddQuestions(ctx, p.ID, items); err != nil {
			return 0, err
		}
	}
	return p.ID, nil
}

// Assemble 组装试卷题目（处理随机/遗传算法）
func (a *PaperApp) Assemble(ctx context.Context, paperID int64) ([]entity.Question, error) {
	p, err := a.paperRepo.GetByID(ctx, paperID)
	if err != nil {
		return nil, errcode.New(errcode.CodePaperNotExist, "PaperNotExist")
	}
	switch p.Strategy {
	case consts.StrategyFixed:
		pq, _ := a.paperRepo.GetQuestionsByPaper(ctx, paperID)
		ids := make([]int64, 0, len(pq))
		for _, item := range pq {
			ids = append(ids, item.QuestionID)
		}
		return a.qRepo.ListByIDs(ctx, ids)
	case consts.StrategyRandom:
		return a.assembleRandom(ctx, p)
	case consts.StrategyGA:
		return a.assembleGA(ctx, p)
	}
	return nil, errcode.New(errcode.CodeSystemError, "未知的组卷策略")
}

// assembleRandom 随机组卷
func (a *PaperApp) assembleRandom(ctx context.Context, p *entity.Paper) ([]entity.Question, error) {
	var cfg dto.RandomConfig
	if err := json.Unmarshal([]byte(p.ConfigRule), &cfg); err != nil {
		return nil, err
	}
	all := []entity.Question{}
	for _, r := range cfg.Rules {
		qs, _ := a.qRepo.RandomByTypeAndDifficulty(ctx, r.Type, r.Difficulty, r.Count)
		all = append(all, qs...)
	}
	return all, nil
}

// assembleGA 遗传算法组卷（基于适应度：难度/章节/知识点均衡）
func (a *PaperApp) assembleGA(ctx context.Context, p *entity.Paper) ([]entity.Question, error) {
	// 简版：退化为随机组卷（生产可接入完整 GA 算法）
	return a.assembleRandom(ctx, p)
}

// GetDetail
func (a *PaperApp) GetDetail(ctx context.Context, id int64) (*entity.Paper, []entity.Question, error) {
	p, err := a.paperRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, errcode.New(errcode.CodePaperNotExist, "PaperNotExist")
	}
	qs, _ := a.Assemble(ctx, id)
	return p, qs, nil
}

// List
func (a *PaperApp) List(ctx context.Context, page, size int, keyword string) ([]entity.Paper, int64, error) {
	return a.paperRepo.List(ctx, page, size, keyword)
}


// Update 更新试卷（含题目绑定）
func (a *PaperApp) Update(ctx context.Context, id int64, req *dto.CreatePaperReq) error {
	// 检查试卷是否存在
	existing, err := a.paperRepo.GetByID(ctx, id)
	if err != nil {
		return errcode.New(errcode.CodePaperNotExist, "PaperNotExist")
	}

	// 更新基础字段
	existing.Title = req.Title
	existing.Description = req.Description
	existing.Strategy = req.Strategy
	existing.TotalScore = req.TotalScore
	existing.Duration = req.Duration
	existing.PassScore = req.PassScore

	switch req.Strategy {
	case consts.StrategyFixed:
		if len(req.QuestionIDs) == 0 {
			return errcode.New(errcode.CodeQuestionEmpty, "QuestionEmpty")
		}
		ids, _ := json.Marshal(req.QuestionIDs)
		existing.QuestionIDs = string(ids)
		// 清空 ConfigRule（使用合法 JSON 避免 MySQL JSON 列错误）
		existing.ConfigRule = "{}"
	case consts.StrategyRandom, consts.StrategyGA:
		cfg, _ := json.Marshal(req.ConfigRule)
		existing.ConfigRule = string(cfg)
		// 清空 QuestionIDs（使用合法 JSON 避免 MySQL JSON 列错误）
		existing.QuestionIDs = "[]"
	}

	if err := a.paperRepo.Update(ctx, existing); err != nil {
		return err
	}

	// 固定策略：重新绑定题目
	if req.Strategy == consts.StrategyFixed && len(req.QuestionIDs) > 0 {
		items := make([]entity.PaperQuestion, 0, len(req.QuestionIDs))
		qs, _ := a.qRepo.ListByIDs(ctx, req.QuestionIDs)
		qmap := map[int64]entity.Question{}
		for _, q := range qs {
			qmap[q.ID] = q
		}
		for i, qid := range req.QuestionIDs {
			score := 1.0
			if q, ok := qmap[qid]; ok && q.Score > 0 {
				score = q.Score
			}
			items = append(items, entity.PaperQuestion{
				PaperID: id, QuestionID: qid, Sort: i, Score: score,
			})
		}
		if err := a.paperRepo.AddQuestions(ctx, id, items); err != nil {
			return err
		}
	}

	return nil
}

// Delete
func (a *PaperApp) Delete(ctx context.Context, id int64) error {
	return a.paperRepo.Delete(ctx, id)
}
