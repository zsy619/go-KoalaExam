package consts

// 收藏来源
const (
	FavoriteSourceManual    int8 = 1 // 主动收藏
	FavoriteSourceAuto      int8 = 2 // 错题自动
	FavoriteSourceRecommend int8 = 3 // 智能推荐
)

// 收藏目标类型
const (
	TargetTypeQuestion  int8 = 1 // 题目
	TargetTypePaper     int8 = 2 // 试卷
	TargetTypeKnowledge int8 = 3 // 知识点
)
