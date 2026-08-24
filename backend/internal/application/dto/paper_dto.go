package dto

// CreatePaperReq
type CreatePaperReq struct {
	Title       string      `json:"title" binding:"required"`
	Description string      `json:"description"`
	Strategy    int8        `json:"strategy" binding:"required"`
	TotalScore  float64     `json:"total_score"`
	Duration    int         `json:"duration"`
	PassScore   float64     `json:"pass_score"`
	QuestionIDs []int64     `json:"question_ids"` // 固定策略
	ConfigRule  interface{} `json:"config_rule"`  // 随机/遗传配置
}

// RandomRule 单题型规则
type RandomRule struct {
	Type       int8    `json:"type"` // 题型
	Difficulty int8    `json:"difficulty"`
	Count      int     `json:"count"` // 抽取数量
	Score      float64 `json:"score"` // 每题分值
}

// RandomConfig 随机组卷配置
type RandomConfig struct {
	Rules      []RandomRule `json:"rules"`
	TotalScore float64      `json:"total_score"`
}
