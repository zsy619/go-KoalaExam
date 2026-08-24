package repository

import (
	"context"
)

// FavoriteFolderListFilter 收藏夹过滤
type FavoriteFolderListFilter struct {
	UserID int64
	IsSystem *bool // nil=全部
}

// FavoriteFolderRepository 收藏夹仓储接口
type IFavoriteFolderRepository interface {
	Create(ctx context.Context, f any) error
	Update(ctx context.Context, f any) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (any, error)
	ListByUser(ctx context.Context, userID int64) (any, error)
}

// FavoriteListFilter 收藏过滤
type FavoriteListFilter struct {
	UserID     int64
	TargetType int8  // 1=题目 2=试卷 3=知识点
	TargetID   int64 // 指定目标 ID 时筛选唯一
	FolderID   int64
}

// FavoriteRepository 收藏仓储接口
type IFavoriteRepository interface {
	Create(ctx context.Context, f any) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter FavoriteListFilter) (any, error)
	IsFavorited(ctx context.Context, uid, targetID int64, targetType int8) (bool, error)
	Toggle(ctx context.Context, uid, targetID int64, targetType int8, folderID int64) (bool, error)
	CountByUser(ctx context.Context, uid int64) (int64, error)
}

// WrongLogListFilter 错题日志过滤
type WrongLogListFilter struct {
	UserID      int64
	IsReviewed  *bool
	MasteryLevel int8
}

// WrongLogRepository 错题日志仓储接口
type IWrongLogRepository interface {
	Create(ctx context.Context, w any) error
	Update(ctx context.Context, w any) error
	GetByID(ctx context.Context, id int64) (any, error)
	ListByUser(ctx context.Context, filter WrongLogListFilter) (any, error)
	ListByQuestion(ctx context.Context, uid, qid int64) (any, error)
	MarkReviewed(ctx context.Context, id int64) error
	CountByUser(ctx context.Context, uid int64) (int64, error)
	MasteryDistribution(ctx context.Context, uid int64) (map[int8]int64, error)
}

// ExamRecordListFilter 考试记录过滤
type ExamRecordListFilter struct {
	PageQuery
	UserID  int64
	ExamID  int64
	Status  int8 // 0=进行中 1=已交卷 2=已批改 3=异常；-1=全部
}

// ExamRecordRepository 考试记录仓储接口
type IExamRecordRepository interface {
	Create(ctx context.Context, r any) error
	Update(ctx context.Context, r any) error
	GetByID(ctx context.Context, id int64) (any, error)
	GetByExamAndUser(ctx context.Context, examID, userID int64) (any, error)
	ListByExam(ctx context.Context, examID int64, page, size int) (any, int64, error)
	ListByUser(ctx context.Context, filter ExamRecordListFilter) (any, int64, error)

	// ScoreStats 单场考试分数统计
	ScoreStats(ctx context.Context, examID int64) (any, error)

	// TodayCount 今日新增记录数
	TodayCount(ctx context.Context) (int64, error)

	// CountTotal 总记录数
	CountTotal(ctx context.Context) (int64, error)
}
