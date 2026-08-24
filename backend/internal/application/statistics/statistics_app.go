package statistics

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
)

// StatisticsApp 统计分析服务
type StatisticsApp struct {
	db         *gorm.DB
	recordRepo *repository.ExamRecordRepository
	favRepo    *repository.FavoriteRepository
	wrongRepo  *repository.WrongLogRepository
}

func NewStatisticsApp(db *gorm.DB, r *repository.ExamRecordRepository, f *repository.FavoriteRepository, w *repository.WrongLogRepository) *StatisticsApp {
	return &StatisticsApp{db: db, recordRepo: r, favRepo: f, wrongRepo: w}
}

// ExamOverview 考试概览
type ExamOverview struct {
	ExamID      int64            `json:"exam_id"`
	AvgScore    float64          `json:"avg_score"`
	MaxScore    float64          `json:"max_score"`
	MinScore    float64          `json:"min_score"`
	TotalCount  int64            `json:"total_count"`
	PassedCount int64            `json:"passed_count"`
	PassRate    float64          `json:"pass_rate"`
	ScoreRange  map[string]int64 `json:"score_range"`
}

func (a *StatisticsApp) ExamOverview(ctx context.Context, examID int64) (*ExamOverview, error) {
	overview := &ExamOverview{ExamID: examID, ScoreRange: map[string]int64{"0-59": 0, "60-69": 0, "70-79": 0, "80-89": 0, "90-100": 0}}
	type row struct {
		Avg, Max, Min float64
		Cnt, Passed   int64
	}
	var r row
	err := a.db.WithContext(ctx).
		Model(struct{}{}). // 表模型将在 repository 中处理
		Where("exam_id = ? AND status = 2", examID).
		Select("AVG(total_score) AS avg, MAX(total_score) AS max, MIN(total_score) AS min, COUNT(*) AS cnt, SUM(CASE WHEN passed=1 THEN 1 ELSE 0 END) AS passed").
		Scan(&r).Error
	if err != nil { return nil, err }
	overview.AvgScore = r.Avg
	overview.MaxScore = r.Max
	overview.MinScore = r.Min
	overview.TotalCount = r.Cnt
	overview.PassedCount = r.Passed
	if r.Cnt > 0 { overview.PassRate = float64(r.Passed) / float64(r.Cnt) * 100 }
	// 分数段
	type scoreRow struct { Score float64 }
	var scores []scoreRow
	a.db.WithContext(ctx).Table("ke_exam_record").Where("exam_id = ? AND status = 2", examID).Select("total_score AS score").Scan(&scores)
	for _, s := range scores {
		switch {
		case s.Score < 60: overview.ScoreRange["0-59"]++
		case s.Score < 70: overview.ScoreRange["60-69"]++
		case s.Score < 80: overview.ScoreRange["70-79"]++
		case s.Score < 90: overview.ScoreRange["80-89"]++
		default: overview.ScoreRange["90-100"]++
		}
	}
	return overview, nil
}

// UserLearningStats 用户学习统计
type UserLearningStats struct {
	UserID         int64            `json:"user_id"`
	TotalFavorites int64            `json:"total_favorites"`
	WrongTotal     int64            `json:"wrong_total"`
	MasteryDist    map[string]int64 `json:"mastery_dist"`
	ExamsCompleted int64            `json:"exams_completed"`
	AvgScore       float64          `json:"avg_score"`
	ReviewedCount  int64            `json:"reviewed_count"`
}

func (a *StatisticsApp) UserLearningStats(ctx context.Context, userID int64) (*UserLearningStats, error) {
	stats := &UserLearningStats{UserID: userID, MasteryDist: map[string]int64{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0}}
	// 收藏数
	a.db.WithContext(ctx).Table("ke_favorite").Where("user_id = ? AND deleted_at IS NULL", userID).Count(&stats.TotalFavorites)
	// 错题 + 掌握度
	type masteryRow struct { MasteryLevel int8; Cnt int64 }
	var rows []masteryRow
	a.db.WithContext(ctx).Table("ke_wrong_log").Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("mastery_level, COUNT(*) AS cnt").Group("mastery_level").Scan(&rows)
	for _, r := range rows {
		if r.MasteryLevel >= 1 && r.MasteryLevel <= 5 { stats.MasteryDist[strconv.Itoa(int(r.MasteryLevel))] = r.Cnt; stats.WrongTotal += r.Cnt }
	}
	// 已考考试 + 平均分
	type examRow struct { Cnt int64; Avg float64 }
	var er examRow
	a.db.WithContext(ctx).Table("ke_exam_record").Where("user_id = ? AND status IN (1,2)", userID).
		Select("COUNT(*) AS cnt, COALESCE(AVG(total_score),0) AS avg").Scan(&er)
	stats.ExamsCompleted = er.Cnt
	stats.AvgScore = er.Avg
	// 已复习错题
	a.db.WithContext(ctx).Table("ke_wrong_log").Where("user_id = ? AND is_reviewed = 1 AND deleted_at IS NULL", userID).Count(&stats.ReviewedCount)
	return stats, nil
}

// DashboardSummary 仪表盘汇总（超管首页）
type DashboardSummary struct {
	UserTotal    int64 `json:"user_total"`
	StudentTotal int64 `json:"student_total"`
	TeacherTotal int64 `json:"teacher_total"`
	QuestionTotal int64 `json:"question_total"`
	PaperTotal   int64 `json:"paper_total"`
	ExamTotal    int64 `json:"exam_total"`
	RecordTotal  int64 `json:"record_total"`
	WrongTotal   int64 `json:"wrong_total"`
	TodayRecords int64 `json:"today_records"`
}

func (a *StatisticsApp) DashboardSummary(ctx context.Context) (*DashboardSummary, error) {
	d := &DashboardSummary{}
	a.db.WithContext(ctx).Table("ke_user").Where("deleted_at IS NULL").Count(&d.UserTotal)
	a.db.WithContext(ctx).Table("ke_user").Where("role = ? AND deleted_at IS NULL", 3).Count(&d.StudentTotal)
	a.db.WithContext(ctx).Table("ke_user").Where("role = ? AND deleted_at IS NULL", 2).Count(&d.TeacherTotal)
	a.db.WithContext(ctx).Table("ke_question").Where("deleted_at IS NULL").Count(&d.QuestionTotal)
	a.db.WithContext(ctx).Table("ke_paper").Where("deleted_at IS NULL").Count(&d.PaperTotal)
	a.db.WithContext(ctx).Table("ke_exam").Where("deleted_at IS NULL").Count(&d.ExamTotal)
	a.db.WithContext(ctx).Table("ke_exam_record").Where("deleted_at IS NULL").Count(&d.RecordTotal)
	a.db.WithContext(ctx).Table("ke_wrong_log").Where("deleted_at IS NULL").Count(&d.WrongTotal)
	today := time.Now().Format("2006-01-02")
	a.db.WithContext(ctx).Table("ke_exam_record").Where("DATE(start_time) = ?", today).Count(&d.TodayRecords)
	return d, nil
}
