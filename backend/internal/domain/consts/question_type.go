package consts

// 题型
const (
	QuestionTypeSingle    int8 = 1 // 单选
	QuestionTypeMultiple  int8 = 2 // 多选
	QuestionTypeJudge     int8 = 3 // 判断
	QuestionTypeFill      int8 = 4 // 填空
	QuestionTypeUncertain int8 = 5 // 不定项
	QuestionTypeCode      int8 = 6 // 编程
)

// QuestionTypeText
func QuestionTypeText(t int8) string {
	switch t {
	case QuestionTypeSingle:
		return "单选题"
	case QuestionTypeMultiple:
		return "多选题"
	case QuestionTypeJudge:
		return "判断题"
	case QuestionTypeFill:
		return "填空题"
	case QuestionTypeUncertain:
		return "不定项"
	case QuestionTypeCode:
		return "编程题"
	}
	return "未知"
}

// IsObjective 判断是否客观题（自动阅卷）
func IsObjective(t int8) bool {
	return t == QuestionTypeSingle || t == QuestionTypeMultiple || t == QuestionTypeJudge || t == QuestionTypeUncertain
}
