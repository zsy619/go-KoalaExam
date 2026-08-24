package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 题目模板库
type Template struct {
	CategoryID int64
	Type       int // 1-6
	Title      string
	Options    []map[string]string
	Answer     []string
	Analysis   string
	Tag        string
	Score      int
}

var singleSubjects = []string{
	"计算机基础", "网络协议", "操作系统", "数据结构", "编程语言",
	"前端框架", "后端开发", "数据库", "算法", "安全",
}

func main() {
	rand.Seed(time.Now().UnixNano())
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/KoalaExam?charset=utf8mb4")
	if err != nil { fmt.Println("open:", err); os.Exit(1) }
	defer db.Close()

	var maxID int64
	db.QueryRow("SELECT COALESCE(MAX(id), 0) FROM ke_question").Scan(&maxID)
	fmt.Printf("现有最大 id: %d\n", maxID)

	var totalBefore int64
	db.QueryRow("SELECT COUNT(*) FROM ke_question").Scan(&totalBefore)
	fmt.Printf("现有题目数: %d\n", totalBefore)

	// 生成 1100 道题（覆盖各类）
	templates := buildTemplates()
	fmt.Printf("生成模板数: %d\n", len(templates))

	// 批量插入
	tx, err := db.Begin()
	if err != nil { fmt.Println("tx:", err); os.Exit(1) }

	stmt, err := tx.Prepare(`INSERT INTO ke_question
		(id, category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil { fmt.Println("prepare:", err); tx.Rollback(); os.Exit(1) }
	defer stmt.Close()

	inserted := int64(0)
	for i, t := range templates {
		id := maxID + int64(i+1)
		categoryID := t.CategoryID
		diff := rand.Intn(3) + 1 // 1-3
		creator := int64(1)
		status := 1

		var optionsStr, answerStr sql.NullString
		if t.Options != nil {
			b, _ := json.Marshal(t.Options)
			optionsStr = sql.NullString{String: string(b), Valid: true}
		}
		if t.Answer != nil {
			b, _ := json.Marshal(t.Answer)
			answerStr = sql.NullString{String: string(b), Valid: true}
		}

		_, err := stmt.Exec(id, categoryID, t.Type, diff, t.Title, optionsStr, answerStr, t.Analysis, t.Tag, t.Score, creator, status)
		if err != nil {
			fmt.Printf("insert id=%d err: %v\n", id, err)
			continue
		}
		inserted++
	}

	if err := tx.Commit(); err != nil { fmt.Println("commit:", err); os.Exit(1) }

	fmt.Printf("✅ 成功插入 %d 道题\n", inserted)

	var totalAfter int64
	db.QueryRow("SELECT COUNT(*) FROM ke_question").Scan(&totalAfter)
	fmt.Printf("插入后总数: %d\n", totalAfter)

	// 按类型统计
	rows, _ := db.Query("SELECT type, COUNT(*) FROM ke_question GROUP BY type ORDER BY type")
	defer rows.Close()
	fmt.Println("\n按类型统计:")
	for rows.Next() {
		var t int
		var c int64
		rows.Scan(&t, &c)
		fmt.Printf("  type=%d (%s): %d\n", t, typeName(t), c)
	}

	rows2, _ := db.Query("SELECT category_id, COUNT(*) FROM ke_question GROUP BY category_id ORDER BY category_id")
	defer rows2.Close()
	fmt.Println("\n按分类统计:")
	for rows2.Next() {
		var c int64
		var cnt int64
		rows2.Scan(&c, &cnt)
		fmt.Printf("  cat=%d: %d\n", c, cnt)
	}
}

func typeName(t int) string {
	switch t {
	case 1: return "单选"
	case 2: return "多选"
	case 3: return "判断"
	case 4: return "填空"
	case 5: return "简答"
	case 6: return "编程"
	}
	return "未知"
}

func buildTemplates() []Template {
	var ts []Template

	// ===== 单选题 (type=1) - 500道 =====
	ts = append(ts, generateSingle(1, "计算机基础", "CS", 80)...)
	ts = append(ts, generateSingle(2, "前端开发", "FE", 80)...)
	ts = append(ts, generateSingle(3, "后端开发", "BE", 80)...)
	ts = append(ts, generateSingle(4, "算法与数据结构", "ALGO", 80)...)
	ts = append(ts, generateSingle(5, "数据库", "DB", 90)...)
	ts = append(ts, generateSingle(6, "网络与安全", "NET", 90)...)
	_ = singleSubjects

	// ===== 多选题 (type=2) - 200道 =====
	ts = append(ts, generateMulti(1, "计算机基础", "CS", 35)...)
	ts = append(ts, generateMulti(2, "前端开发", "FE", 35)...)
	ts = append(ts, generateMulti(3, "后端开发", "BE", 35)...)
	ts = append(ts, generateMulti(5, "数据库", "DB", 45)...)
	ts = append(ts, generateMulti(6, "网络与安全", "NET", 50)...)

	// ===== 判断题 (type=3) - 150道 =====
	ts = append(ts, generateJudge(1, 25)...)
	ts = append(ts, generateJudge(2, 25)...)
	ts = append(ts, generateJudge(3, 25)...)
	ts = append(ts, generateJudge(5, 35)...)
	ts = append(ts, generateJudge(6, 40)...)

	// ===== 填空题 (type=4) - 150道 =====
	ts = append(ts, generateFill(1, 30)...)
	ts = append(ts, generateFill(3, 35)...)
	ts = append(ts, generateFill(5, 40)...)
	ts = append(ts, generateFill(2, 25)...)
	ts = append(ts, generateFill(6, 20)...)

	// ===== 简答题 (type=5) - 80道 =====
	ts = append(ts, generateEssay(1, 15)...)
	ts = append(ts, generateEssay(3, 20)...)
	ts = append(ts, generateEssay(4, 15)...)
	ts = append(ts, generateEssay(5, 15)...)
	ts = append(ts, generateEssay(2, 15)...)

	// ===== 编程题 (type=6) - 80道 =====
	ts = append(ts, generateCode(3, 25)...)
	ts = append(ts, generateCode(4, 35)...)
	ts = append(ts, generateCode(2, 20)...)

	return ts
}

// ============ 单选 ============
func generateSingle(cat int64, name, tag string, n int) []Template {
	rand.Seed(time.Now().UnixNano() + int64(n*cat))
	out := make([]Template, 0, n)
	for i := 0; i < n; i++ {
		opts, ans := randomSingleQA(name, i+1)
		out = append(out, Template{
			CategoryID: cat,
			Type:       1,
			Title:      fmt.Sprintf("[%s-单选-%04d] %s", name, i+1, singleTitles(name, i)),
			Options:    opts,
			Answer:     []string{ans},
			Analysis:   fmt.Sprintf("解析：本题考查 %s 知识点。正确答案 %s。", name, ans),
			Tag:        tag,
			Score:      2,
		})
	}
	return out
}

func singleTitles(name string, i int) string {
	templates := []string{
		fmt.Sprintf("关于 %s 的描述，下列哪项是正确的？", name),
		fmt.Sprintf("以下哪个选项最准确地描述了 %s 的特性？", name),
		fmt.Sprintf("在 %s 领域中，下列说法正确的是？", name),
		fmt.Sprintf("%s 中常用的概念是以下哪一个？", name),
		fmt.Sprintf("下列关于 %s 的说法中，正确的是？", name),
	}
	return templates[i%len(templates)]
}

func randomSingleQA(name string, seed int) ([]map[string]string, string) {
	correctIdx := rand.Intn(4)
	keys := []string{"A", "B", "C", "D"}
	corrects = []string{}
	for i, k := range keys {
		text := fmt.Sprintf("选项 %s：%s 相关描述 %d", k, name, seed+i)
		if i == correctIdx {
			text = fmt.Sprintf("选项 %s：%s 的正确描述 %d", k, name, seed)
		}
		opts = append(opts, map[string]string{"key": k, "text": text})
	}
	return opts, keys[correctIdx]
}

// ============ 多选 ============
func generateMulti(cat int64, name, tag string, n int) []Template {
	out := make([]Template, 0, n)
	for i := 0; i < n; i++ {
		opts, ans := randomMultiQA(name, i+1)
		out = append(out, Template{
			CategoryID: cat,
			Type:       2,
			Title:      fmt.Sprintf("[%s-多选-%04d] 以下哪些是 %s 的相关特性？（多选）", name, i+1, name),
			Options:    opts,
			Answer:     ans,
			Analysis:   fmt.Sprintf("解析：本题为多选题，答案为 %s。", strings.Join(ans, ",")),
			Tag:        tag,
			Score:      3,
		})
	}
	return out
}

func randomMultiQA(name string, seed int) ([]map[string]string, []string) {
	correctCount := 2 + rand.Intn(2) // 2-3 个正确答案
	selected := rand.Perm(4)[:correctCount]
	keySet := map[string]bool{}
	ans := []string{}
	for _, idx := range selected {
		keySet[[]"A","B","C","D"[idx]] = true
		ans = append(ans, []string{"A","B","C","D"}[idx])
	}
	keys := []string{"A", "B", "C", "D"}
	opts := make([]map[string]string, 0, 4)
	for i, k := range keys {
		text := fmt.Sprintf("特性 %s：%s 知识点 %d", k, name, seed+i)
		if keySet[k] {
			text = fmt.Sprintf("正确特性 %s：%s 描述 %d", k, name, seed)
		}
		opts = append(opts, map[string]string{"key": k, "text": text})
	}
	return opts, ans
}

// ============ 判断 ============
func generateJudge(cat int64, n int) []Template {
	out := make([]Template, 0, n)
	for i := 0; i < n; i++ {
		isTrue := rand.Intn(2) == 0
		ans := "false"
		if isTrue { ans = "true" }
		out = append(out, Template{
			CategoryID: cat,
			Type:       3,
			Title:      fmt.Sprintf("[判断-%04d] 这是关于计算机知识的一条陈述，请判断对错：相关描述 %d-%d。", i+1, cat, i),
			Answer:     []string{ans},
			Analysis:   fmt.Sprintf("解析：正确答案是 %s。", map[string]string{"true":"正确","false":"错误"}[ans]),
			Tag:        "判断",
			Score:      2,
		})
	}
	return out
}

// ============ 填空 ============
func generateFill(cat int64, n int) []Template {
	out := make([]Template, 0, n)
	for i := 0; i < n; i++ {
		keywords := []string{"algorithm", "function", "variable", "class", "interface", "module", "package", "method"}
		kw := keywords[rand.Intn(len(keywords))]
		out = append(out, Template{
			CategoryID: cat,
			Type:       4,
			Title:      fmt.Sprintf("[填空-%04d] 在计算机科学中，相关的关键术语是 ____ （请填入英文术语）。", i+1),
			Answer:     []string{kw},
			Analysis:   fmt.Sprintf("解析：答案为 %s。", kw),
			Tag:        "填空",
			Score:      3,
		})
	}
	return out
}

// ============ 简答 ============
func generateEssay(cat int64, n int) []Template {
	out := make([]Template, 0, n)
	topics := []string{
		"请简述相关概念及其应用场景。",
		"对比分析两种方案的优缺点。",
		"描述其工作原理，并举例说明。",
		"分析常见问题及解决方案。",
		"总结最佳实践。",
	}
	for i := 0; i < n; i++ {
		out = append(out, Template{
			CategoryID: cat,
			Type:       5,
			Title:      fmt.Sprintf("[简答-%04d] %s", i+1, topics[i%len(topics)]),
			Answer:     []string{"open"},
			Analysis:   "主观题，需人工批改。",
			Tag:        "简答",
			Score:      8,
		})
	}
	return out
}

// ============ 编程 ============
func generateCode(cat int64, n int) []Template {
	out := make([]Template, 0, n)
	problems := []string{
		"实现一个函数，输入数组，输出最大值。",
		"判断字符串是否为回文。",
		"实现二分查找。",
		"实现两个有序数组合并。",
		"实现快速排序。",
		"实现 LRU 缓存。",
		"实现单例模式。",
		"反转链表。",
	}
	for i := 0; i < n; i++ {
		out = append(out, Template{
			CategoryID: cat,
			Type:       6,
			Title:      fmt.Sprintf("[编程-%04d] %s", i+1, problems[i%len(problems)]),
			Answer:     []string{"code"},
			Analysis:   "编程题，需提交代码并自动/人工评测。",
			Tag:        "编程",
			Score:      15,
		})
	}
	return out
}
