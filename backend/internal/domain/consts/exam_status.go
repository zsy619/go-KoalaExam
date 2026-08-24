package consts

// 考试状态
const (
	ExamStatusDraft    int8 = 0 // 未发布
	ExamStatusRunning  int8 = 1 // 进行中
	ExamStatusFinished int8 = 2 // 已结束
	ExamStatusArchived int8 = 3 // 已归档
)

// 考试记录状态
const (
	RecordStatusOngoing   int8 = 0 // 进行中
	RecordStatusSubmitted int8 = 1 // 已交卷
	RecordStatusGraded    int8 = 2 // 已批改
	RecordStatusAbnormal  int8 = 3 // 异常（作弊/超时）
)

// 组卷策略
const (
	StrategyFixed  int8 = 1 // 固定
	StrategyRandom int8 = 2 // 随机
	StrategyGA     int8 = 3 // 遗传算法
)
