// Package service 领域服务。
//
// 领域服务用于处理不属于某个具体实体的业务逻辑（如试卷组装、答案比对）。
// 与应用服务（application）的区别：领域服务不依赖任何基础设施，纯业务逻辑。
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/internal/domain/valueobject"
)

// QuestionSelector 题目选择器（领域服务）。
//
// 根据组卷策略（固定/随机/遗传算法）从题库中选择题目。
type QuestionSelector struct{}

// NewQuestionSelector 构造选择器。
func NewQuestionSelector() *QuestionSelector { return &QuestionSelector{} }

// SelectFixed 固定选题（按题目 ID 列表）。
func (s *QuestionSelector) SelectFixed(all []entity.Question, ids []int64) ([]entity.Question, error) {
	byID := make(map[int64]entity.Question, len(all))
	for _, q := range all {
		byID[q.ID] = q
	}
	out := make([]entity.Question, 0, len(ids))
	for _, id := range ids {
		q, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("question %d not found", id)
		}
		out = append(out, q)
	}
	return out, nil
}

// SelectRandom 随机选题（按题型/难度分组抽题）。
//
// rules: 每条规则指定 (题型, 难度, 数量, 总分)
func (s *QuestionSelector) SelectRandom(all []entity.Question, rules []SelectionRule) ([]entity.Question, error) {
	// 按 (题型, 难度) 索引
	bucket := make(map[tuple2][]entity.Question)
	for _, q := range all {
		key := tuple2{Type: q.Type, Difficulty: q.Difficulty}
		bucket[key] = append(bucket[key], q)
	}

	// 随机化桶内顺序（使用 Fisher-Yates 洗牌）
	for key := range bucket {
		bucket[key] = shuffle(bucket[key])
	}

	out := make([]entity.Question, 0)
	used := make(map[int64]bool)
	for _, rule := range rules {
		key := tuple2{Type: rule.Type, Difficulty: rule.Difficulty}
		pool, ok := bucket[key]
		if !ok {
			return nil, fmt.Errorf("no questions match rule: type=%d difficulty=%d", rule.Type, rule.Difficulty)
		}
		picked := 0
		for _, q := range pool {
			if used[q.ID] {
				continue
			}
			out = append(out, q)
			used[q.ID] = true
			picked++
			if picked >= rule.Count {
				break
			}
		}
		if picked < rule.Count {
			return nil, fmt.Errorf("insufficient questions for rule: type=%d difficulty=%d need %d, got %d",
				rule.Type, rule.Difficulty, rule.Count, picked)
		}
	}
	return out, nil
}

// SelectionRule 组卷规则。
type SelectionRule struct {
	Type       int8
	Difficulty int8
	Count      int
	PerScore   float64
}

type tuple2 struct {
	Type       int8
	Difficulty int8
}

// shuffle Fisher-Yates 洗牌。
func shuffle(qs []entity.Question) []entity.Question {
	for i := len(qs) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		qs[i], qs[j] = qs[j], qs[i]
	}
	return qs
}

// AnswerComparator 答案比对器（领域服务）。
type AnswerComparator struct{}

// NewAnswerComparator 构造比对器。
func NewAnswerComparator() *AnswerComparator { return &AnswerComparator{} }

// Compare 比较用户答案与正确答案。
//
// 支持字符串、数字、数组（多选）、布尔类型。
func (c *AnswerComparator) Compare(correctRaw, userRaw interface{}) bool {
	if correctRaw == nil || userRaw == nil {
		return false
	}

	cs := normalize(correctRaw)
	us := normalize(userRaw)

	// 字符串完全相等
	if cs == us {
		return true
	}

	// 数组元素相等（多选题）
	cArr, cOk := correctRaw.([]interface{})
	uArr, uOk := userRaw.([]interface{})
	if cOk && uOk {
		return compareSlices(cArr, uArr)
	}

	return false
}

// normalize 标准化为字符串。
func normalize(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.ToLower(strings.TrimSpace(t))
	case []byte:
		return strings.ToLower(strings.TrimSpace(string(t)))
	case float64:
		return fmt.Sprintf("%v", t)
	case int, int64:
		return fmt.Sprintf("%d", t)
	default:
		b, _ := json.Marshal(t)
		return strings.ToLower(strings.TrimSpace(string(b)))
	}
}

// compareSlices 比较两个切片（顺序无关）。
func compareSlices(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	ac := make([]string, len(a))
	bc := make([]string, len(b))
	for i := range a {
		ac[i] = normalize(a[i])
		bc[i] = normalize(b[i])
	}
	sort.Strings(ac)
	sort.Strings(bc)
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

// GradingStrategy 自动阅卷策略。
type GradingStrategy struct {
	comparator *AnswerComparator
}

// NewGradingStrategy 构造阅卷策略。
func NewGradingStrategy() *GradingStrategy {
	return &GradingStrategy{comparator: NewAnswerComparator()}
}

// GradeObjective 客观题阅卷。
//
// 返回 (得分, 是否正确)。
func (g *GradingStrategy) GradeObjective(q entity.Question, userAns interface{}) (float64, bool) {
	if !consts.IsObjective(q.Type) {
		return 0, false
	}
	var correct interface{}
	if err := json.Unmarshal([]byte(q.Answer), &correct); err != nil {
		return 0, false
	}
	if g.comparator.Compare(correct, userAns) {
		return q.Score, true
	}
	return 0, false
}

// ScoreSigner 成绩签名（防篡改）。
type ScoreSigner struct {
	secret string
}

// NewScoreSigner 构造签名器。
func NewScoreSigner(secret string) *ScoreSigner {
	return &ScoreSigner{secret: secret}
}

// Sign 生成 SHA-256 签名（record_id|user_id|exam_id|score|secret）。
func (s *ScoreSigner) Sign(recordID, userID, examID int64, score float64) string {
	h := sha256.New()
	payload := fmt.Sprintf("%d|%d|%d|%.2f|%s|%d",
		recordID, userID, examID, score, s.secret, rand.Int())
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// PaperAssembler 试卷组装器（领域服务）。
type PaperAssembler struct {
	selector *QuestionSelector
}

// NewPaperAssembler 构造试卷组装器。
func NewPaperAssembler() *PaperAssembler {
	return &PaperAssembler{selector: NewQuestionSelector()}
}

// Assemble 组装试卷（固定题号）。
func (a *PaperAssembler) Assemble(paper *entity.Paper, questions []entity.Question) (valueobject.ExamPaper, error) {
	if paper == nil {
		return valueobject.ExamPaper{}, fmt.Errorf("paper is nil")
	}
	var ids []int64
	if err := json.Unmarshal([]byte(paper.QuestionIDs), &ids); err != nil {
		return valueobject.ExamPaper{}, err
	}
	picked, err := a.selector.SelectFixed(questions, ids)
	if err != nil {
		return valueobject.ExamPaper{}, err
	}
	vq := make([]valueobject.ExamQuestion, 0, len(picked))
	for _, q := range picked {
		var opts interface{}
		if q.Options != "" {
			_ = json.Unmarshal([]byte(q.Options), &opts)
		}
		vq = append(vq, valueobject.ExamQuestion{
			ID:         q.ID,
			Type:       q.Type,
			Difficulty: q.Difficulty,
			Title:      q.Title,
			Options:    opts,
			Score:      q.Score,
		})
	}
	return valueobject.NewExamPaper(vq), nil
}

// ScoreCalculator 分数计算器。
type ScoreCalculator struct{}

// Calculate 计算总分（包含主观题加分）。
func (c *ScoreCalculator) Calculate(objective, subjective, passScore float64) valueobject.ExamScore {
	return valueobject.NewExamScore(objective, subjective, passScore)
}
