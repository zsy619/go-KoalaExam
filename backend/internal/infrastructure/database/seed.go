package database

import (
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/entity"
	"github.com/your-team/koala-exam-backend/pkg/encrypt"
)

// SeedDev 使用 GORM 创建完整的开发数据
// 通过 cmd/migrate 的 seed 操作触发
func SeedDev(db *gorm.DB) error {
	log.Println("开始播种开发数据...")

	hashed, err := encrypt.BcryptPassword("koala123")
	if err != nil {
		return err
	}

	// ============ 1. 组织/院系 ============
	departments := []entity.Department{
		{Name: "计算机学院", ParentID: 0, Sort: 1},
		{Name: "软件工程系", ParentID: 1, Sort: 1},
		{Name: "人工智能系", ParentID: 1, Sort: 2},
		{Name: "外语学院", ParentID: 0, Sort: 2},
		{Name: "经济管理学院", ParentID: 0, Sort: 3},
	}
	for i := range departments {
		db.FirstOrCreate(&departments[i], "name = ?", departments[i].Name)
	}

	// ============ 2. 班级 ============
	classes := []entity.Class{
		{Name: "软工2024-1班", Grade: "2024级", DepartmentID: &departments[1].ID, TeacherID: nil, StudentCnt: 40},
		{Name: "软工2024-2班", Grade: "2024级", DepartmentID: &departments[1].ID, TeacherID: nil, StudentCnt: 38},
		{Name: "AI2024-1班", Grade: "2024级", DepartmentID: &departments[2].ID, TeacherID: nil, StudentCnt: 35},
	}
	for i := range classes {
		db.FirstOrCreate(&classes[i], "name = ?", classes[i].Name)
	}

	// ============ 3. 用户（1 超管 + 2 教师 + 5 学员）============
	users := []entity.User{
		{Username: "admin", Password: hashed, Nickname: "超级管理员", Email: "admin@koala.com", Phone: "13800000001", Role: consts.RoleSuperAdmin, Gender: 1, Status: 1},
		{Username: "teacher", Password: hashed, Nickname: "教师小明", Email: "teacher@koala.com", Phone: "13800000002", Role: consts.RoleTeacher, Gender: 1, Status: 1, DepartmentID: &departments[1].ID},
		{Username: "teacher2", Password: hashed, Nickname: "教师小红", Email: "teacher2@koala.com", Phone: "13800000003", Role: consts.RoleTeacher, Gender: 2, Status: 1, DepartmentID: &departments[2].ID},
		{Username: "student", Password: hashed, Nickname: "学员小考", Email: "student@koala.com", Phone: "13800000004", Role: consts.RoleStudent, Gender: 1, Status: 1, ClassID: &classes[0].ID, DepartmentID: &departments[1].ID},
		{Username: "student2", Password: hashed, Nickname: "学员小拉", Email: "student2@koala.com", Phone: "13800000005", Role: consts.RoleStudent, Gender: 2, Status: 1, ClassID: &classes[0].ID, DepartmentID: &departments[1].ID},
		{Username: "student3", Password: hashed, Nickname: "学员小狮", Email: "student3@koala.com", Phone: "13800000006", Role: consts.RoleStudent, Gender: 1, Status: 1, ClassID: &classes[1].ID, DepartmentID: &departments[1].ID},
		{Username: "student4", Password: hashed, Nickname: "学员小虎", Email: "student4@koala.com", Phone: "13800000007", Role: consts.RoleStudent, Gender: 1, Status: 1, ClassID: &classes[2].ID, DepartmentID: &departments[2].ID},
		{Username: "student5", Password: hashed, Nickname: "学员小鹿", Email: "student5@koala.com", Phone: "13800000008", Role: consts.RoleStudent, Gender: 2, Status: 1, ClassID: &classes[0].ID, DepartmentID: &departments[1].ID},
	}
	for i := range users {
		db.FirstOrCreate(&users[i], "username = ?", users[i].Username)
	}
	admin := users[0]

	// ============ 4. 题库分类 ============
	categories := []entity.QuestionCategory{
		{Name: "计算机基础", Code: "CS-BASIC", Sort: 1, CreatorID: admin.ID},
		{Name: "前端开发", Code: "FE", Sort: 2, CreatorID: admin.ID},
		{Name: "后端开发", Code: "BE", Sort: 3, CreatorID: admin.ID},
		{Name: "算法与数据结构", Code: "ALGO", Sort: 4, CreatorID: admin.ID},
		{Name: "数据库", Code: "DATABASE", Sort: 5, CreatorID: admin.ID},
		{Name: "网络与安全", Code: "NET", Sort: 6, CreatorID: admin.ID},
	}
	for i := range categories {
		db.FirstOrCreate(&categories[i], "name = ?", categories[i].Name)
	}

	// ============ 5. 题目（20 道，含 6 种题型）============
	options4 := `[{"key":"A","text":"4位"},{"key":"B","text":"8位"},{"key":"C","text":"16位"},{"key":"D","text":"32位"}]`
	ansB := `["B"]`
	questions := []entity.Question{
		// 单选
		{CategoryID: categories[0].ID, Type: consts.QuestionTypeSingle, Difficulty: 1, Title: "一个字节（byte）由多少个二进制位组成？", Options: options4, Answer: ansB, Analysis: "1 byte = 8 bits", Tags: "计算机基础", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[5].ID, Type: consts.QuestionTypeSingle, Difficulty: 1, Title: "HTTP 协议默认端口是？", Options: `[{"key":"A","text":"21"},{"key":"B","text":"23"},{"key":"C","text":"80"},{"key":"D","text":"443"}]`, Answer: `["C"]`, Analysis: "HTTP 默认 80，HTTPS 默认 443", Tags: "网络", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[0].ID, Type: consts.QuestionTypeSingle, Difficulty: 2, Title: "下列哪个不是操作系统的内核？", Options: `[{"key":"A","text":"Linux"},{"key":"B","text":"Windows NT"},{"key":"C","text":"macOS"},{"key":"D","text":"Oracle"}]`, Answer: `["D"]`, Analysis: "Oracle 是数据库", Tags: "操作系统", Score: 3, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[2].ID, Type: consts.QuestionTypeSingle, Difficulty: 2, Title: "Go 语言中，下列哪个关键字用于延迟执行？", Options: `[{"key":"A","text":"go"},{"key":"B","text":"defer"},{"key":"C","text":"await"},{"key":"D","text":"yield"}]`, Answer: `["B"]`, Analysis: "defer 用于延迟函数调用", Tags: "Go", Score: 3, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[2].ID, Type: consts.QuestionTypeSingle, Difficulty: 2, Title: "Hertz 是哪个公司开源的 HTTP 框架？", Options: `[{"key":"A","text":"Google"},{"key":"B","text":"字节跳动"},{"key":"C","text":"Meta"},{"key":"D","text":"Microsoft"}]`, Answer: `["B"]`, Analysis: "字节跳动（cloudwego）", Tags: "Go", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[4].ID, Type: consts.QuestionTypeSingle, Difficulty: 2, Title: "MySQL 中哪种存储引擎支持事务？", Options: `[{"key":"A","text":"MyISAM"},{"key":"B","text":"Memory"},{"key":"C","text":"InnoDB"},{"key":"D","text":"CSV"}]`, Answer: `["C"]`, Analysis: "InnoDB 支持 ACID 事务", Tags: "MySQL", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[4].ID, Type: consts.QuestionTypeSingle, Difficulty: 2, Title: "SQL 中去除重复行的关键字是？", Options: `[{"key":"A","text":"UNIQUE"},{"key":"B","text":"DISTINCT"},{"key":"C","text":"GROUP"},{"key":"D","text":"ORDER"}]`, Answer: `["B"]`, Analysis: "SELECT DISTINCT", Tags: "SQL", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[0].ID, Type: consts.QuestionTypeSingle, Difficulty: 1, Title: "二进制 1010 等于十进制的多少？", Options: `[{"key":"A","text":"8"},{"key":"B","text":"10"},{"key":"C","text":"12"},{"key":"D","text":"14"}]`, Answer: `["B"]`, Analysis: "1010 = 8+2 = 10", Tags: "二进制", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[4].ID, Type: consts.QuestionTypeSingle, Difficulty: 2, Title: "以下哪个不是关系型数据库？", Options: `[{"key":"A","text":"MySQL"},{"key":"B","text":"PostgreSQL"},{"key":"C","text":"MongoDB"},{"key":"D","text":"Oracle"}]`, Answer: `["C"]`, Analysis: "MongoDB 是文档型 NoSQL", Tags: "数据库", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[5].ID, Type: consts.QuestionTypeSingle, Difficulty: 2, Title: "下列哪个不是常见的对称加密算法？", Options: `[{"key":"A","text":"AES"},{"key":"B","text":"DES"},{"key":"C","text":"RSA"},{"key":"D","text":"3DES"}]`, Answer: `["C"]`, Analysis: "RSA 是非对称加密", Tags: "安全", Score: 2, CreatorID: admin.ID, Status: 1},
		// 多选
		{CategoryID: categories[1].ID, Type: consts.QuestionTypeMultiple, Difficulty: 2, Title: "以下哪些是 Vue 3 的新特性？", Options: `[{"key":"A","text":"组合式 API"},{"key":"B","text":"Fragment"},{"key":"C","text":"Teleport"},{"key":"D","text":"Mixin"}]`, Answer: `["A","C"]`, Analysis: "Vue 3 引入组合式 API 和 Teleport", Tags: "Vue", Score: 4, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[0].ID, Type: consts.QuestionTypeMultiple, Difficulty: 2, Title: "以下哪些是面向对象的特性？", Options: `[{"key":"A","text":"封装"},{"key":"B","text":"继承"},{"key":"C","text":"多态"},{"key":"D","text":"并发"}]`, Answer: `["A","B","C"]`, Analysis: "OOP 三大特性：封装、继承、多态", Tags: "OOP", Score: 4, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[3].ID, Type: consts.QuestionTypeMultiple, Difficulty: 3, Title: "以下哪些排序算法时间复杂度是 O(n log n)？", Options: `[{"key":"A","text":"快速排序"},{"key":"B","text":"归并排序"},{"key":"C","text":"堆排序"},{"key":"D","text":"冒泡排序"}]`, Answer: `["A","B","C"]`, Analysis: "冒泡排序是 O(n²)", Tags: "算法", Score: 4, CreatorID: admin.ID, Status: 1},
		// 判断
		{CategoryID: categories[5].ID, Type: consts.QuestionTypeJudge, Difficulty: 1, Title: "TCP 是面向连接的可靠传输协议。", Answer: "[true]", Analysis: "TCP 三次握手、四次挥手", Tags: "网络", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[1].ID, Type: consts.QuestionTypeJudge, Difficulty: 1, Title: "JavaScript 是强类型语言。", Answer: "[false]", Analysis: "JS 是弱类型/动态类型语言", Tags: "JS", Score: 2, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[4].ID, Type: consts.QuestionTypeJudge, Difficulty: 1, Title: "主键不允许重复但允许为空。", Answer: "[false]", Analysis: "主键不允许重复，也不允许为空", Tags: "SQL", Score: 2, CreatorID: admin.ID, Status: 1},
		// 填空
		{CategoryID: categories[2].ID, Type: consts.QuestionTypeFill, Difficulty: 2, Title: "Go 语言中用于并发编程的关键字是 ____。", Answer: `["goroutine"]`, Analysis: "Go 使用 goroutine 实现并发", Tags: "Go", Score: 3, CreatorID: admin.ID, Status: 1},
		{CategoryID: categories[1].ID, Type: consts.QuestionTypeFill, Difficulty: 2, Title: "CSS 中用于设置元素外边距的属性是 ____。", Answer: `["margin"]`, Analysis: "margin 控制外边距", Tags: "CSS", Score: 3, CreatorID: admin.ID, Status: 1},
		// 不定项
		{CategoryID: categories[1].ID, Type: consts.QuestionTypeUncertain, Difficulty: 3, Title: "关于 React Hooks 的描述，正确的是？", Options: `[{"key":"A","text":"只能在函数组件顶层调用"},{"key":"B","text":"可以在条件语句中使用"},{"key":"C","text":"useEffect 副作用清理函数返回 undefined 会报错"},{"key":"D","text":"useState 的 setter 是异步的"}]`, Answer: `["A","D"]`, Analysis: "Hook 不能在条件/循环中调用", Tags: "React", Score: 4, CreatorID: admin.ID, Status: 1},
		// 编程
		{CategoryID: categories[3].ID, Type: consts.QuestionTypeCode, Difficulty: 3, Title: "实现一个函数，输入 n，返回斐波那契数列第 n 项。", Answer: "[\"code\"]", Analysis: "经典动态题或递归", Tags: "算法", Score: 10, CreatorID: admin.ID, Status: 1},
	}
	for i := range questions {
		db.FirstOrCreate(&questions[i], "title = ?", questions[i].Title)
	}

	// ============ 6. 试卷 ============
	qids := func(ids ...int) string { b, _ := json.Marshal(ids); return string(b) }
	papers := []entity.Paper{
		{Title: "计算机基础测试卷（A）", Description: "考察计算机基础、网络、数据库等综合知识", Strategy: consts.StrategyFixed, TotalScore: 100, Duration: 60, PassScore: 60, Status: 1, CreatorID: admin.ID, QuestionIDs: qids(1, 2, 3, 14, 15, 17, 21)},
		{Title: "前端开发综合卷", Description: "Vue、React、HTML、CSS 综合题", Strategy: consts.StrategyFixed, TotalScore: 100, Duration: 60, PassScore: 60, Status: 1, CreatorID: admin.ID, QuestionIDs: qids(2, 11, 18, 19)},
		{Title: "算法与数据结构", Description: "考察排序、动态规划、递归", Strategy: consts.StrategyFixed, TotalScore: 100, Duration: 90, PassScore: 60, Status: 1, CreatorID: admin.ID, QuestionIDs: qids(13, 20)},
		{Title: "Go 后端开发卷（随机组卷）", Description: "随机抽取 Go/数据库/网络题", Strategy: consts.StrategyRandom, TotalScore: 100, Duration: 60, PassScore: 60, Status: 1, CreatorID: admin.ID, ConfigRule: `{"rules":[{"type":1,"difficulty":2,"count":3,"score":10},{"type":3,"difficulty":1,"count":2,"score":5}]}`},
		{Title: "入门考试", Description: "新手入门测试，10 分钟快速测试", Strategy: consts.StrategyFixed, TotalScore: 50, Duration: 10, PassScore: 30, Status: 1, CreatorID: admin.ID, QuestionIDs: qids(1, 14, 16, 21)},
	}
	for i := range papers {
		db.FirstOrCreate(&papers[i], "title = ?", papers[i].Title)
	}

	// ============ 7. 试卷-题目关联（按试卷1 为例）============
	pqList := []entity.PaperQuestion{
		{PaperID: papers[0].ID, QuestionID: 1, Type: 1, Score: 10, Sort: 1, Section: "一、单选题"},
		{PaperID: papers[0].ID, QuestionID: 2, Type: 1, Score: 10, Sort: 2, Section: "一、单选题"},
		{PaperID: papers[0].ID, QuestionID: 3, Type: 1, Score: 10, Sort: 3, Section: "一、单选题"},
		{PaperID: papers[0].ID, QuestionID: 14, Type: 3, Score: 10, Sort: 4, Section: "二、判断题"},
		{PaperID: papers[0].ID, QuestionID: 15, Type: 3, Score: 10, Sort: 5, Section: "二、判断题"},
		{PaperID: papers[0].ID, QuestionID: 17, Type: 4, Score: 20, Sort: 6, Section: "三、填空题"},
		{PaperID: papers[0].ID, QuestionID: 21, Type: 1, Score: 30, Sort: 7, Section: "四、应用题"},
		{PaperID: papers[1].ID, QuestionID: 2, Type: 1, Score: 20, Sort: 1, Section: "一、单选"},
		{PaperID: papers[1].ID, QuestionID: 11, Type: 2, Score: 30, Sort: 2, Section: "二、多选"},
		{PaperID: papers[1].ID, QuestionID: 18, Type: 4, Score: 25, Sort: 3, Section: "三、填空"},
		{PaperID: papers[1].ID, QuestionID: 19, Type: 5, Score: 25, Sort: 4, Section: "四、不定项"},
		{PaperID: papers[2].ID, QuestionID: 13, Type: 2, Score: 40, Sort: 1, Section: "一、选择题"},
		{PaperID: papers[2].ID, QuestionID: 20, Type: 6, Score: 60, Sort: 2, Section: "二、编程题"},
	}
	for _, pq := range pqList {
		db.FirstOrCreate(&pq, "paper_id = ? AND question_id = ?", pq.PaperID, pq.QuestionID)
	}

	// ============ 8. 考试（5 场）============
	now := time.Now()
	exams := []entity.Exam{
		{Title: "2024春季计算机基础期中考试", Description: "面向全校计算机基础课", PaperID: papers[0].ID, StartTime: now.AddDate(0, -1, 0), EndTime: now.AddDate(1, 0, 0), Duration: 60, MaxAttempts: 1, ShuffleQ: true, ShuffleOpt: true, AntiCheat: true, Status: consts.ExamStatusRunning, CreatorID: admin.ID},
		{Title: "前端开发期末考试", Description: "面向软工专业", PaperID: papers[1].ID, StartTime: now.AddDate(0, -1, 0), EndTime: now.AddDate(1, 0, 0), Duration: 90, MaxAttempts: 2, ShuffleQ: true, ShuffleOpt: true, AntiCheat: true, Status: consts.ExamStatusRunning, CreatorID: admin.ID},
		{Title: "算法能力测试（随时考）", Description: "算法专项测试，可重复参加", PaperID: papers[2].ID, StartTime: now.AddDate(-1, 0, 0), EndTime: now.AddDate(1, 0, 0), Duration: 90, MaxAttempts: 5, ShuffleQ: true, ShuffleOpt: true, AntiCheat: true, Status: consts.ExamStatusRunning, CreatorID: admin.ID},
		{Title: "Go 后端开发认证考试", Description: "企业认证考试", PaperID: papers[3].ID, StartTime: now.AddDate(0, 0, -1), EndTime: now.AddDate(1, 0, 0), Duration: 60, MaxAttempts: 1, ShuffleQ: true, ShuffleOpt: true, AntiCheat: true, Status: consts.ExamStatusRunning, CreatorID: admin.ID},
		{Title: "新手入门摸底测试", Description: "10 分钟快速摸底", PaperID: papers[4].ID, StartTime: now.AddDate(-1, 0, 0), EndTime: now.AddDate(1, 0, 0), Duration: 10, MaxAttempts: 99, ShuffleQ: false, ShuffleOpt: false, AntiCheat: false, Status: consts.ExamStatusRunning, CreatorID: admin.ID},
	}
	for i := range exams {
		db.FirstOrCreate(&exams[i], "title = ?", exams[i].Title)
	}

	// ============ 9. 学员收藏夹 ============
	studentIDs := []int64{users[3].ID, users[4].ID, users[5].ID, users[6].ID, users[7].ID}
	for _, sid := range studentIDs {
		db.FirstOrCreate(&entity.FavoriteFolder{UserID: sid, Name: "我的收藏", IsSystem: true, Color: "#409eff", Icon: "star", Sort: 1}, "user_id = ? AND name = ? AND is_system = ?", sid, "我的收藏", true)
		db.FirstOrCreate(&entity.FavoriteFolder{UserID: sid, Name: "我的错题本", IsSystem: true, Color: "#f56c6c", Icon: "warning", Sort: 2}, "user_id = ? AND name = ? AND is_system = ?", sid, "我的错题本", true)
	}
	// 给 student 额外建自定义夹
	db.FirstOrCreate(&entity.FavoriteFolder{UserID: users[3].ID, Name: "高频错题", IsSystem: false, Color: "#e6a23c", Icon: "folder", Sort: 3}, "user_id = ? AND name = ?", users[3].ID, "高频错题")
	db.FirstOrCreate(&entity.FavoriteFolder{UserID: users[3].ID, Name: "必背大题", IsSystem: false, Color: "#67c23a", Icon: "folder", Sort: 4}, "user_id = ? AND name = ?", users[3].ID, "必背大题")

	// ============ 10. 收藏记录（主动 + 错题自动）============
	var stuFolder, stuWrongFolder entity.FavoriteFolder
	db.Where("user_id = ? AND name = ?", users[3].ID, "我的收藏").First(&stuFolder)
	db.Where("user_id = ? AND name = ?", users[3].ID, "我的错题本").First(&stuWrongFolder)
	favs := []entity.Favorite{
		{UserID: users[3].ID, TargetType: consts.TargetTypeQuestion, TargetID: 13, FolderID: &stuFolder.ID, SourceType: consts.FavoriteSourceManual, Difficulty: 3, Note: "排序算法对比"},
		{UserID: users[3].ID, TargetType: consts.TargetTypeQuestion, TargetID: 20, FolderID: &stuFolder.ID, SourceType: consts.FavoriteSourceManual, Difficulty: 3, Note: "经典动态题"},
		{UserID: users[3].ID, TargetType: consts.TargetTypeQuestion, TargetID: 11, FolderID: &stuFolder.ID, SourceType: consts.FavoriteSourceManual, Difficulty: 2, Note: "Vue3 新特性"},
		{UserID: users[3].ID, TargetType: consts.TargetTypeQuestion, TargetID: 3, FolderID: &stuWrongFolder.ID, SourceType: consts.FavoriteSourceAuto, Difficulty: 2},
		{UserID: users[3].ID, TargetType: consts.TargetTypeQuestion, TargetID: 7, FolderID: &stuWrongFolder.ID, SourceType: consts.FavoriteSourceAuto, Difficulty: 2},
	}
	for i := range favs {
		db.FirstOrCreate(&favs[i], "user_id = ? AND target_type = ? AND target_id = ?", favs[i].UserID, favs[i].TargetType, favs[i].TargetID)
	}

	// ============ 11. 错题日志 ============
	wrongs := []entity.WrongAnswerLog{
		{UserID: users[3].ID, QuestionID: 3, ExamID: exams[0].ID, UserAnswer: "[\"A\"]", CorrectAnswer: "[\"D\"]", WrongCount: 2, LastWrongAt: now.AddDate(0, 0, -5), IsReviewed: false, MasteryLevel: 2},
		{UserID: users[3].ID, QuestionID: 7, ExamID: exams[0].ID, UserAnswer: "[\"A\"]", CorrectAnswer: "[\"B\"]", WrongCount: 1, LastWrongAt: now.AddDate(0, 0, -5), IsReviewed: true, MasteryLevel: 4},
		{UserID: users[3].ID, QuestionID: 15, ExamID: exams[0].ID, UserAnswer: "[true]", CorrectAnswer: "[false]", WrongCount: 1, LastWrongAt: now.AddDate(0, 0, -5), IsReviewed: false, MasteryLevel: 3},
		{UserID: users[3].ID, QuestionID: 17, ExamID: exams[0].ID, UserAnswer: "[\"channel\"]", CorrectAnswer: "[\"goroutine\"]", WrongCount: 1, LastWrongAt: now.AddDate(0, 0, -5), IsReviewed: false, MasteryLevel: 2},
		{UserID: users[4].ID, QuestionID: 1, ExamID: exams[0].ID, UserAnswer: "[\"C\"]", CorrectAnswer: "[\"B\"]", WrongCount: 1, LastWrongAt: now.AddDate(0, 0, -4), IsReviewed: true, MasteryLevel: 5},
		{UserID: users[4].ID, QuestionID: 12, ExamID: exams[1].ID, UserAnswer: "[\"A\",\"B\"]", CorrectAnswer: "[\"A\",\"B\",\"C\"]", WrongCount: 1, LastWrongAt: now.AddDate(0, 0, -3), IsReviewed: false, MasteryLevel: 2},
		{UserID: users[5].ID, QuestionID: 13, ExamID: exams[2].ID, UserAnswer: "[\"A\",\"B\"]", CorrectAnswer: "[\"A\",\"B\",\"C\"]", WrongCount: 1, LastWrongAt: now.AddDate(0, 0, -2), IsReviewed: false, MasteryLevel: 1},
	}
	for i := range wrongs {
		db.FirstOrCreate(&wrongs[i], "user_id = ? AND question_id = ?", wrongs[i].UserID, wrongs[i].QuestionID)
	}

	// ============ 12. 考试记录（历史已完成）============
	records := []entity.ExamRecord{
		{ExamID: exams[0].ID, UserID: users[3].ID, Status: consts.RecordStatusGraded, StartTime: now.AddDate(0, 0, -5), SubmitTime: ptrTime(now.AddDate(0, 0, -5).Add(time.Hour)), Duration: 3600, TotalScore: 76, ObjectiveScore: 76, SubjectiveScore: 0, Passed: true, ScoreHash: "sha256-stu1-exam1", TabSwitchCnt: 1},
		{ExamID: exams[0].ID, UserID: users[4].ID, Status: consts.RecordStatusGraded, StartTime: now.AddDate(0, 0, -4), SubmitTime: ptrTime(now.AddDate(0, 0, -4).Add(time.Hour)), Duration: 3600, TotalScore: 92, ObjectiveScore: 92, SubjectiveScore: 0, Passed: true, ScoreHash: "sha256-stu2-exam1", TabSwitchCnt: 0},
		{ExamID: exams[1].ID, UserID: users[4].ID, Status: consts.RecordStatusGraded, StartTime: now.AddDate(0, 0, -3), SubmitTime: ptrTime(now.AddDate(0, 0, -3).Add(90 * time.Minute)), Duration: 5400, TotalScore: 80, ObjectiveScore: 80, SubjectiveScore: 0, Passed: true, ScoreHash: "sha256-stu2-exam2", TabSwitchCnt: 2},
		{ExamID: exams[2].ID, UserID: users[5].ID, Status: consts.RecordStatusGraded, StartTime: now.AddDate(0, 0, -2), SubmitTime: ptrTime(now.AddDate(0, 0, -2).Add(90 * time.Minute)), Duration: 5400, TotalScore: 60, ObjectiveScore: 60, SubjectiveScore: 0, Passed: true, ScoreHash: "sha256-stu3-exam3", TabSwitchCnt: 0},
		{ExamID: exams[4].ID, UserID: users[3].ID, Status: consts.RecordStatusGraded, StartTime: now.AddDate(0, 0, -7), SubmitTime: ptrTime(now.AddDate(0, 0, -7).Add(10 * time.Minute)), Duration: 600, TotalScore: 40, ObjectiveScore: 40, SubjectiveScore: 0, Passed: true, ScoreHash: "sha256-stu1-exam5", TabSwitchCnt: 0},
		{ExamID: exams[4].ID, UserID: users[4].ID, Status: consts.RecordStatusGraded, StartTime: now.AddDate(0, 0, -7), SubmitTime: ptrTime(now.AddDate(0, 0, -7).Add(10 * time.Minute)), Duration: 600, TotalScore: 50, ObjectiveScore: 50, SubjectiveScore: 0, Passed: true, ScoreHash: "sha256-stu2-exam5", TabSwitchCnt: 0},
	}
	for i := range records {
		db.FirstOrCreate(&records[i], "exam_id = ? AND user_id = ?", records[i].ExamID, records[i].UserID)
	}

	log.Println("✓ 播种完成：1 超管 + 2 教师 + 5 学员 + 6 题库分类 + 20 题目 + 5 试卷 + 5 考试 + 12 收藏夹 + 5 收藏 + 7 错题 + 6 考试记录")
	return nil
}

// ptrTime 辅助函数
func ptrTime(t time.Time) *time.Time { return &t }
