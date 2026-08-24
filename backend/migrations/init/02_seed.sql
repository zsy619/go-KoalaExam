-- ============================================================
-- KoalaExam 种子数据（测试数据）
-- ============================================================

-- 1. 组织/院系
INSERT INTO ke_department (id, name, parent_id, sort) VALUES
  (1, '计算机学院', 0, 1),
  (2, '软件工程系', 1, 1),
  (3, '人工智能系', 1, 2),
  (4, '外语学院', 0, 2),
  (5, '经济管理学院', 0, 3);

-- 2. 班级
INSERT INTO ke_class (id, name, grade, department_id, teacher_id, student_cnt) VALUES
  (1, '软工2024-1班', '2024级', 2, 2, 40),
  (2, '软工2024-2班', '2024级', 2, 2, 38),
  (3, 'AI2024-1班', '2024级', 3, 2, 35);

-- 3. 用户（密码：koala123，使用 bcrypt 哈希 $2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK）
INSERT INTO ke_user (id, username, password, nickname, email, phone, role, gender, status, class_id, department_id) VALUES
  (1, 'admin', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '超级管理员', 'admin@koala.com', '13800000001', 1, 1, 1, NULL, NULL),
  (2, 'teacher', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '教师小明', 'teacher@koala.com', '13800000002', 2, 1, 1, NULL, 2),
  (3, 'teacher2', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '教师小红', 'teacher2@koala.com', '13800000003', 2, 2, 1, NULL, 3),
  (4, 'student', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '学员小考', 'student@koala.com', '13800000004', 3, 1, 1, 1, 2),
  (5, 'student2', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '学员小拉', 'student2@koala.com', '13800000005', 3, 2, 1, 1, 2),
  (6, 'student3', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '学员小狮', 'student3@koala.com', '13800000006', 3, 1, 1, 2, 2),
  (7, 'student4', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '学员小虎', 'student4@koala.com', '13800000007', 3, 1, 1, 3, 3),
  (8, 'student5', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '学员小鹿', 'student5@koala.com', '13800000008', 3, 2, 1, 1, 2);

-- 4. 题库分类
INSERT INTO ke_question_category (id, parent_id, name, code, sort, creator_id) VALUES
  (1, 0, '计算机基础', 'CS-BASIC', 1, 1),
  (2, 0, '前端开发', 'FE', 2, 1),
  (3, 0, '后端开发', 'BE', 3, 1),
  (4, 0, '算法与数据结构', 'ALGO', 4, 1),
  (5, 0, '数据库', 'DATABASE', 5, 1),
  (6, 0, '网络与安全', 'NET', 6, 1);

-- 5. 题目
INSERT INTO ke_question (id, category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
  (1, 1, 1, 1, '一个字节（byte）由多少个二进制位组成？', '[{"key":"A","text":"4位"},{"key":"B","text":"8位"},{"key":"C","text":"16位"},{"key":"D","text":"32位"}]', '["B"]', '1 byte = 8 bits', '计算机基础', 2, 1, 1),
  (2, 1, 1, 1, 'HTTP 协议默认端口是？', '[{"key":"A","text":"21"},{"key":"B","text":"23"},{"key":"C","text":"80"},{"key":"D","text":"443"}]', '["C"]', 'HTTP 默认 80，HTTPS 默认 443', '网络', 2, 1, 1),
  (3, 1, 1, 2, '下列哪个不是操作系统的内核？', '[{"key":"A","text":"Linux"},{"key":"B","text":"Windows NT"},{"key":"C","text":"macOS"},{"key":"D","text":"Oracle"}]', '["D"]', 'Oracle 是数据库', '操作系统', 3, 1, 1),
  (4, 3, 1, 2, 'Go 语言中，下列哪个关键字用于延迟执行？', '[{"key":"A","text":"go"},{"key":"B","text":"defer"},{"key":"C","text":"await"},{"key":"D","text":"yield"}]', '["B"]', 'defer 用于延迟函数调用', 'Go', 3, 1, 1),
  (5, 3, 1, 2, 'Hertz 是哪个公司开源的 HTTP 框架？', '[{"key":"A","text":"Google"},{"key":"B","text":"字节跳动"},{"key":"C","text":"Meta"},{"key":"D","text":"Microsoft"}]', '["B"]', '字节跳动（cloudwego）', 'Go', 2, 1, 1),
  (6, 5, 1, 2, 'MySQL 中哪种存储引擎支持事务？', '[{"key":"A","text":"MyISAM"},{"key":"B","text":"Memory"},{"key":"C","text":"InnoDB"},{"key":"D","text":"CSV"}]', '["C"]', 'InnoDB 支持 ACID 事务', 'MySQL', 2, 1, 1),
  (7, 5, 1, 2, 'SQL 中去除重复行的关键字是？', '[{"key":"A","text":"UNIQUE"},{"key":"B","text":"DISTINCT"},{"key":"C","text":"GROUP"},{"key":"D","text":"ORDER"}]', '["B"]', 'SELECT DISTINCT', 'SQL', 2, 1, 1),
  (8, 2, 2, 2, '以下哪些是 Vue 3 的新特性？', '[{"key":"A","text":"组合式 API"},{"key":"B","text":"Fragment"},{"key":"C","text":"Teleport"},{"key":"D","text":"Mixin"}]', '["A","C"]', 'Vue 3 引入组合式 API 和 Teleport', 'Vue', 4, 1, 1),
  (9, 1, 2, 2, '以下哪些是面向对象的特性？', '[{"key":"A","text":"封装"},{"key":"B","text":"继承"},{"key":"C","text":"多态"},{"key":"D","text":"并发"}]', '["A","B","C"]', 'OOP 三大特性：封装、继承、多态', 'OOP', 4, 1, 1),
  (10, 4, 2, 3, '以下哪些排序算法时间复杂度是 O(n log n)？', '[{"key":"A","text":"快速排序"},{"key":"B","text":"归并排序"},{"key":"C","text":"堆排序"},{"key":"D","text":"冒泡排序"}]', '["A","B","C"]', '冒泡排序是 O(n²)', '算法', 4, 1, 1),
  (11, 1, 3, 1, 'TCP 是面向连接的可靠传输协议。', NULL, '[true]', 'TCP 三次握手、四次挥手', '网络', 2, 1, 1),
  (12, 1, 3, 1, 'JavaScript 是强类型语言。', NULL, '[false]', 'JS 是弱类型/动态类型语言', 'JS', 2, 1, 1),
  (13, 5, 3, 1, '主键不允许重复但允许为空。', NULL, '[false]', '主键不允许重复，也不允许为空', 'SQL', 2, 1, 1),
  (14, 3, 4, 2, 'Go 语言中用于并发编程的关键字是 ____。', NULL, '["goroutine"]', 'Go 使用 goroutine 实现并发', 'Go', 3, 1, 1),
  (15, 1, 4, 2, 'CSS 中用于设置元素外边距的属性是 ____。', NULL, '["margin"]', 'margin 控制外边距', 'CSS', 3, 1, 1),
  (16, 2, 5, 3, '关于 React Hooks 的描述，正确的是？', '[{"key":"A","text":"只能在函数组件顶层调用"},{"key":"B","text":"可以在条件语句中使用"},{"key":"C","text":"useEffect 副作用清理函数返回 undefined 会报错"},{"key":"D","text":"useState 的 setter 是异步的"}]', '["A","D"]', 'Hook 不能在条件/循环中调用', 'React', 4, 1, 1),
  (17, 4, 6, 3, '实现一个函数，输入 n，返回斐波那契数列第 n 项。', NULL, '["code"]', '经典动态题或递归', '算法', 10, 1, 1),
  (18, 1, 1, 1, '二进制 1010 等于十进制的多少？', '[{"key":"A","text":"8"},{"key":"B","text":"10"},{"key":"C","text":"12"},{"key":"D","text":"14"}]', '["B"]', '1010 = 8+2 = 10', '二进制', 2, 1, 1),
  (19, 5, 1, 2, '以下哪个不是关系型数据库？', '[{"key":"A","text":"MySQL"},{"key":"B","text":"PostgreSQL"},{"key":"C","text":"MongoDB"},{"key":"D","text":"Oracle"}]', '["C"]', 'MongoDB 是文档型 NoSQL', '数据库', 2, 1, 1),
  (20, 6, 1, 2, '下列哪个不是常见的对称加密算法？', '[{"key":"A","text":"AES"},{"key":"B","text":"DES"},{"key":"C","text":"RSA"},{"key":"D","text":"3DES"}]', '["C"]', 'RSA 是非对称加密', '安全', 2, 1, 1);

-- 6. 试卷
INSERT INTO ke_paper (id, title, description, strategy, total_score, duration, pass_score, status, creator_id, config_rule, question_ids) VALUES
  (1, '计算机基础测试卷（A）', '考察计算机基础、网络、数据库等综合知识', 1, 100, 60, 60, 1, 1, NULL, '[1,2,3,11,12,14,18]'),
  (2, '前端开发综合卷', 'Vue、React、HTML、CSS 综合题', 1, 100, 60, 60, 1, 1, NULL, '[2,8,15,16]'),
  (3, '算法与数据结构', '考察排序、动态题、递归', 1, 100, 90, 60, 1, 1, NULL, '[10,17]'),
  (4, 'Go 后端开发卷（随机组卷）', '随机抽取 Go/数据库/网络题', 2, 100, 60, 60, 1, 1, '{"rules":[{"type":1,"difficulty":2,"count":3,"score":10},{"type":3,"difficulty":1,"count":2,"score":5}]}', NULL),
  (5, '入门考试', '新手入门测试，10 分钟快速测试', 1, 50, 10, 30, 1, 1, NULL, '[1,11,13,18]');

-- 7. 试卷-题目关联
INSERT INTO ke_paper_question (paper_id, question_id, type, score, sort, section) VALUES
  (1, 1, 1, 10, 1, '一、选择题'),
  (1, 2, 1, 10, 2, '一、选择题'),
  (1, 3, 1, 10, 3, '一、选择题'),
  (1, 11, 3, 10, 4, '二、判断题'),
  (1, 12, 3, 10, 5, '二、判断题'),
  (1, 14, 4, 20, 6, '三、填空题'),
  (1, 18, 1, 30, 7, '四、应用题'),
  (2, 2, 1, 20, 1, '一、单选'),
  (2, 8, 2, 30, 2, '二、多选'),
  (2, 15, 4, 25, 3, '三、填空'),
  (2, 16, 5, 25, 4, '四、不定项'),
  (3, 10, 2, 40, 1, '一、选择题'),
  (3, 17, 6, 60, 2, '二、编程题'),
  (5, 1, 1, 10, 1, '一、单选'),
  (5, 11, 3, 10, 2, '二、判断'),
  (5, 13, 3, 10, 3, '三、判断'),
  (5, 18, 1, 20, 4, '四、单选');

-- 8. 考试
INSERT INTO ke_exam (id, title, description, paper_id, start_time, end_time, duration, max_attempts, shuffle_q, shuffle_opt, anti_cheat, status, creator_id) VALUES
  (1, '2024春季计算机基础期中考试', '面向全校计算机基础课', 1, '2024-03-01 09:00:00', '2099-12-31 23:59:59', 60, 1, 1, 1, 1, 1, 1),
  (2, '前端开发期末考试', '面向软工专业', 2, '2024-06-15 14:00:00', '2099-12-31 23:59:59', 90, 2, 1, 1, 1, 1, 1),
  (3, '算法能力测试（随时考）', '算法专项测试，可重复参加', 3, '2024-01-01 00:00:00', '2099-12-31 23:59:59', 90, 5, 1, 1, 1, 1, 1),
  (4, 'Go 后端开发认证考试', '企业认证考试', 4, '2024-05-01 10:00:00', '2099-12-31 23:59:59', 60, 1, 1, 1, 1, 1, 1),
  (5, '新手入门摸底测试', '10 分钟快速摸底', 5, '2024-01-01 00:00:00', '2099-12-31 23:59:59', 10, 99, 0, 0, 0, 1, 1);

-- 9. 收藏夹（每个学员预置系统夹）
INSERT INTO ke_favorite_folder (user_id, name, is_system, color, icon, sort) VALUES
  (4, '我的收藏', 1, '#409eff', 'star', 1),
  (4, '我的错题本', 1, '#f56c6c', 'warning', 2),
  (5, '我的收藏', 1, '#409eff', 'star', 1),
  (5, '我的错题本', 1, '#f56c6c', 'warning', 2),
  (6, '我的收藏', 1, '#409eff', 'star', 1),
  (6, '我的错题本', 1, '#f56c6c', 'warning', 2),
  (4, '高频错题', 0, '#e6a23c', 'folder', 3),
  (4, '必背大题', 0, '#67c23a', 'folder', 4);

-- 10. 收藏
INSERT INTO ke_favorite (user_id, target_type, target_id, folder_id, source_type, note) VALUES
  (4, 1, 10, 1, 1, '排序算法对比'),
  (4, 1, 17, 1, 1, '经典动态题'),
  (4, 1, 8, 1, 1, 'Vue3 新特性'),
  (4, 1, 3, 7, 2, '操作系统容易混淆'),
  (4, 1, 7, 7, 2, 'SQL 关键字'),
  (5, 1, 9, 5, 1, 'OOP 重要'),
  (6, 1, 4, 9, 1, 'Go defer 用法'),
  (6, 1, 5, 9, 1, 'Hertz 框架');

-- 11. 错题日志
INSERT INTO ke_wrong_log (user_id, question_id, exam_id, exam_record_id, user_answer, correct_answer, wrong_count, last_wrong_at, is_reviewed, mastery_level) VALUES
  (4, 3, 1, 1, '["A"]', '["D"]', 2, '2024-03-15 10:30:00', 0, 2),
  (4, 7, 1, 1, '["A"]', '["B"]', 1, '2024-03-15 10:35:00', 1, 4),
  (4, 12, 1, 1, '[true]', '[false]', 1, '2024-03-15 10:40:00', 0, 3),
  (4, 14, 1, 1, '["channel"]', '["goroutine"]', 1, '2024-03-15 10:45:00', 0, 2),
  (5, 1, 1, 2, '["C"]', '["B"]', 1, '2024-03-16 09:20:00', 1, 5),
  (5, 9, 2, 3, '["A","B"]', '["A","B","C"]', 1, '2024-06-15 14:30:00', 0, 2),
  (6, 10, 3, 4, '["A","B"]', '["A","B","C"]', 1, '2024-04-10 16:00:00', 0, 1),
  (6, 16, 2, 5, '["A"]', '["A","D"]', 1, '2024-06-16 15:00:00', 0, 3);

-- 12. 考试记录
INSERT INTO ke_exam_record (id, exam_id, user_id, status, start_time, submit_time, duration, total_score, objective_score, subjective_score, passed, score_hash, tab_switch_cnt) VALUES
  (1, 1, 4, 2, '2024-03-15 10:00:00', '2024-03-15 11:00:00', 3600, 76, 76, 0, 1, 'sha256-placeholder-1', 1),
  (2, 1, 5, 2, '2024-03-16 09:00:00', '2024-03-16 10:00:00', 3600, 92, 92, 0, 1, 'sha256-placeholder-2', 0),
  (3, 2, 5, 2, '2024-06-15 14:00:00', '2024-06-15 15:30:00', 5400, 80, 80, 0, 1, 'sha256-placeholder-3', 2),
  (4, 3, 6, 2, '2024-04-10 15:00:00', '2024-04-10 16:30:00', 5400, 60, 60, 0, 1, 'sha256-placeholder-4', 0),
  (5, 2, 6, 2, '2024-06-16 14:00:00', '2024-06-16 15:00:00', 3600, 75, 75, 0, 1, 'sha256-placeholder-5', 1),
  (6, 5, 4, 2, '2024-03-01 08:00:00', '2024-03-01 08:10:00', 600, 40, 40, 0, 1, 'sha256-placeholder-6', 0),
  (7, 5, 5, 2, '2024-03-02 10:00:00', '2024-03-02 10:10:00', 600, 50, 50, 0, 1, 'sha256-placeholder-7', 0);



-- ============================================================
-- 4.1 子分类补充题目
-- ============================================================

INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(101, 1, 2, 'CPU 中用于暂存指令的寄存器是？', '[{"key":"A","text":"PC"},{"key":"B","text":"IR"},{"key":"C","text":"ACC"},{"key":"D","text":"MAR"}]', '["B"]', 'IR 指令寄存器暂存当前指令', '组成原理', 3, 1, 1),
(101, 1, 3, '冯·诺依曼体系结构的核心思想是？', '[{"key":"A","text":"存储程序"},{"key":"B","text":"并行计算"},{"key":"C","text":"分布式"},{"key":"D","text":"流水线"}]', '["A"]', '存储程序原理：程序和数据一样存在内存中', '组成原理', 3, 1, 1),
(101, 1, 2, '下列哪个不是 CPU 的组成部分？', '[{"key":"A","text":"ALU"},{"key":"B","text":"CU"},{"key":"C","text":"Cache"},{"key":"D","text":"ROM"}]', '["D"]', 'ROM 是内存，不属于 CPU', '组成原理', 3, 1, 1);

-- 操作系统 (id=102)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(102, 1, 2, '进程与线程的根本区别是？', '[{"key":"A","text":"进程是程序，线程是函数"},{"key":"B","text":"进程是资源分配单位，线程是CPU调度单位"},{"key":"C","text":"进程更快"},{"key":"D","text":"线程占用更多内存"}]', '["B"]', '进程拥有独立资源，线程共享进程资源', '操作系统', 3, 1, 1),
(102, 1, 2, '死锁的四个必要条件不包括？', '[{"key":"A","text":"互斥"},{"key":"B","text":"持有并等待"},{"key":"C","text":"可抢占"},{"key":"D","text":"循环等待"}]', '["C"]', '死锁需要不可抢占，不是可抢占', '操作系统', 3, 1, 1),
(102, 1, 3, '虚拟内存的主要作用是？', '[{"key":"A","text":"提高 CPU 速度"},{"key":"B","text":"扩大可用内存"},{"key":"C","text":"减少磁盘IO"},{"key":"D","text":"加快网络"}]', '["B"]', '虚拟内存让程序使用比物理内存更大的地址空间', '操作系统', 3, 1, 1);

-- 计算机网络基础 (id=103)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(103, 1, 1, 'OSI 七层模型中最底层是？', '[{"key":"A","text":"物理层"},{"key":"B","text":"数据链路层"},{"key":"C","text":"网络层"},{"key":"D","text":"传输层"}]', '["A"]', '物理层是最底层', '网络', 2, 1, 1),
(103, 1, 2, 'IP 地址 192.168.1.1 属于哪类网络？', '[{"key":"A","text":"A 类"},{"key":"B","text":"B 类"},{"key":"C","text":"C 类"},{"key":"D","text":"D 类"}]', '["C"]', '192 开头是 C 类（192-223）', '网络', 2, 1, 1);

-- 数据结构基础 (id=104)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(104, 1, 1, '栈的特点是？', '[{"key":"A","text":"先进先出"},{"key":"B","text":"后进先出"},{"key":"C","text":"随机存取"},{"key":"D","text":"循环访问"}]', '["B"]', 'LIFO', '数据结构', 2, 1, 1),
(104, 1, 2, '二叉树前序遍历的第一个节点一定是？', '[{"key":"A","text":"左子树的根"},{"key":"B","text":"右子树的根"},{"key":"C","text":"根节点"},{"key":"D","text":"叶子节点"}]', '["C"]', '前序遍历：根→左→右', '数据结构', 3, 1, 1);

-- HTML/CSS (id=201)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(201, 1, 1, 'HTML 中用于换行的标签是？', '[{"key":"A","text":"<br>"},{"key":"B","text":"<hr>"},{"key":"C","text":"<p>"},{"key":"D","text":"<div>"}]', '["A"]', '<br> 是换行标签', 'HTML', 2, 1, 1),
(201, 1, 2, 'CSS 中 position: absolute 的定位参考是？', '[{"key":"A","text":"浏览器窗口"},{"key":"B","text":"body 元素"},{"key":"C","text":"最近的已定位父元素"},{"key":"D","text":"文档流"}]', '["C"]', 'absolute 相对于最近的已定位父元素', 'CSS', 3, 1, 1),
(201, 1, 2, 'Flex 布局中 justify-content: center 表示？', '[{"key":"A","text":"垂直居中"},{"key":"B","text":"水平居中"},{"key":"C","text":"两端对齐"},{"key":"D","text":"环绕对齐"}]', '["B"]', 'justify 控制主轴（默认水平）', 'CSS', 3, 1, 1);

-- JavaScript 基础 (id=202)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(202, 1, 1, 'JavaScript 中定义常量的关键字是？', '[{"key":"A","text":"var"},{"key":"B","text":"let"},{"key":"C","text":"const"},{"key":"D","text":"static"}]', '["C"]', 'const 用于定义常量', 'JS', 2, 1, 1),
(202, 1, 2, '下列哪个不是 JS 的原始类型？', '[{"key":"A","text":"string"},{"key":"B","text":"number"},{"key":"C","text":"boolean"},{"key":"D","text":"array"}]', '["D"]', 'array 是引用类型', 'JSON', 2, 1, 1),
(202, 2, 2, '下列哪些是 JS 的循环语句？', '[{"key":"A","text":"for"},{"key":"B","text":"while"},{"key":"C","text":"foreach"},{"key":"D","text":"loop"}]', '["A","B","C"]', 'JS 没有 loop 关键字', 'JS', 4, 1, 1);

-- Vue.js (id=203)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(203, 1, 2, 'Vue 中实现双向绑定的指令是？', '[{"key":"A","text":"v-bind"},{"key":"B","text":"v-model"},{"key":"C","text":"v-on"},{"key":"D","text":"v-if"}]', '["B"]', 'v-model 用于双向绑定', 'Vue', 3, 1, 1),
(203, 1, 2, 'Vue 3 中响应式 API 的核心是？', '[{"key":"A","text":"Object.defineProperty"},{"key":"B","text":"Proxy"},{"key":"C","text":"Reflect"},{"key":"D","text":"Symbol"}]', '["B"]', 'Vue 3 用 Proxy 替代 defineProperty', 'Vue', 3, 1, 1),
(203, 1, 3, 'Pinia 是 Vue 官方推荐的？', '[{"key":"A","text":"路由库"},{"key":"B","text":"状态管理库"},{"key":"C","text":"UI 组件库"},{"key":"D","text":"构建工具"}]', '["B"]', 'Pinia 替代 Vuex 作为状态管理', 'Vue', 3, 1, 1);

-- React (id=204)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(204, 1, 2, 'React 中用于函数组件接收参数的语法是？', '[{"key":"A","text":"arguments"},{"key":"B","text":"props"},{"key":"C","text":"state"},{"key":"D","text":"context"}]', '["B"]', 'props 是函数组件的参数', 'React', 3, 1, 1),
(204, 1, 2, 'React 18 引入的新并发特性是？', '[{"key":"A","text":"Fiber"},{"key":"B","text":"Suspense"},{"key":"C","text":"useTransition"},{"key":"D","text":"useMemo"}]', '["C"]', 'useTransition 用于标记非紧急更新', 'React', 3, 1, 1);

-- 前端工程化 (id=205)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(205, 1, 1, 'webpack 中用于处理 CSS 的 loader 是？', '[{"key":"A","text":"css-loader"},{"key":"B","text":"style-loader"},{"key":"C","text":"css-loader 和 style-loader"},{"key":"D","text":"file-loader"}]', '["C"]', '两个 loader 配合使用', 'Webpack', 3, 1, 1),
(205, 1, 2, 'Vite 比 Webpack 快的主要原因是？', '[{"key":"A","text":"多线程"},{"key":"B","text":"使用 ESM 和 esbuild"},{"key":"C","text":"更好的压缩"},{"key":"D","text":"更小的依赖"}]', '["B"]', 'Vite 利用浏览器原生 ESM', 'Vite', 3, 1, 1);

-- Go 语言 (id=301)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(301, 1, 2, 'Go 中 channel 的方向有几种？', '[{"key":"A","text":"1"},{"key":"B","text":"2"},{"key":"C","text":"3"},{"key":"D","text":"4"}]', '["B"]', '双向 channel 和单向 channel', 'Go', 3, 1, 1),
(301, 1, 2, 'Go 中用于并发安全访问的机制是？', '[{"key":"A","text":"sync.Mutex"},{"key":"B","text":"channel"},{"key":"C","text":"sync.WaitGroup"},{"key":"D","text":"以上都是"}]', '["D"]', '三种都是并发控制工具', 'Go', 3, 1, 1),
(301, 1, 3, 'Go 中 context 的主要用途是？', '[{"key":"A","text":"传递上下文和取消信号"},{"key":"B","text":"存储全局变量"},{"key":"C","text":"替代 channel"},{"key":"D","text":"错误处理"}]', '["A"]', 'context 用于传递截止时间和取消信号', 'Go', 3, 1, 1),
(301, 4, 2, 'Go 语言中用于打印输出的函数是 ____。', NULL, '["fmt.Println"]', 'fmt.Println 标准输出', 'Go', 3, 1, 1),
(301, 2, 3, '下列哪些是 Go 的关键字？', '[{"key":"A","text":"func"},{"key":"B","text":"defer"},{"key":"C","text":"class"},{"key":"D","text":"async"}]', '["A","B"]', 'Go 没有 class 和 async', 'Go', 4, 1, 1);

-- MySQL (id=501)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(501, 1, 2, 'MySQL 中 B+ 树索引的 B 是指？', '[{"key":"A","text":"Binary"},{"key":"B","text":"Balanced"},{"key":"C","text":"Block"},{"key":"D","text":"Base"}]', '["B"]', 'B+ 树是平衡多路查找树', 'MySQL', 3, 1, 1),
(501, 1, 2, 'MySQL 中 EXPLAIN 用于？', '[{"key":"A","text":"执行 SQL"},{"key":"B","text":"分析查询执行计划"},{"key":"C","text":"导出数据"},{"key":"D","text":"权限管理"}]', '["B"]', 'EXPLAIN 分析执行计划', 'MySQL', 3, 1, 1),
(501, 1, 3, 'MySQL 中事务隔离级别最高的是？', '[{"key":"A","text":"Read Uncommitted"},{"key":"B","text":"Read Committed"},{"key":"C","text":"Repeatable Read"},{"key":"D","text":"Serializable"}]', '["D"]', 'Serializable 是最高隔离级别', 'MySQL', 3, 1, 1),
(501, 2, 3, 'MySQL 中支持事务的存储引擎包括？', '[{"key":"A","text":"InnoDB"},{"key":"B","text":"MyISAM"},{"key":"C","text":"NDB"},{"key":"D","text":"Memory"}]', '["A","C"]', 'InnoDB 和 NDB 支持事务', 'MySQL', 4, 1, 1);

-- Redis (id=503)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(503, 1, 1, 'Redis 的默认端口是？', '[{"key":"A","text":"3306"},{"key":"B","text":"6379"},{"key":"C","text":"8080"},{"key":"D","text":"5432"}]', '["B"]', 'Redis 默认端口 6379', 'Redis', 2, 1, 1),
(503, 1, 2, 'Redis 中实现分布式锁的命令是？', '[{"key":"A","text":"SET"},{"key":"B","text":"SETNX"},{"key":"C","text":"GETSET"},{"key":"D","text":"MSET"}]', '["B"]', 'SETNX 实现分布式锁', 'Redis', 3, 1, 1),
(503, 1, 2, 'Redis 的持久化机制不包括？', '[{"key":"A","text":"RDB"},{"key":"B","text":"AOF"},{"key":"C","text":"WAL"},{"key":"D","text":"混合模式"}]', '["C"]', 'WAL 是数据库日志机制，不是 Redis 的', 'Redis', 3, 1, 1);

-- SQL 语言 (id=505)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(505, 1, 1, 'SQL 中查询数据的关键字是？', '[{"key":"A","text":"GET"},{"key":"B","text":"SELECT"},{"key":"C","text":"FETCH"},{"key":"D","text":"READ"}]', '["B"]', 'SELECT 用于查询', 'SQL', 2, 1, 1),
(505, 1, 2, 'SQL 中 LEFT JOIN 的结果是？', '[{"key":"A","text":"两表的交集"},{"key":"B","text":"左表全部 + 右表匹配"},{"key":"C","text":"右表全部 + 左表匹配"},{"key":"D","text":"两表并集"}]', '["B"]', 'LEFT JOIN 保留左表所有记录', 'SQL', 3, 1, 1),
(202, 4, 2, 'JS 中用于解析 JSON 字符串的函数是 ____。', NULL, '["JSON.parse"]', 'JSON.parse 解析为对象', 'JS', 3, 1, 1);

-- TCP/IP 协议 (id=601)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(601, 1, 1, 'TCP 三次握手的目的是？', '[{"key":"A","text":"建立可靠连接"},{"key":"B","text":"传输数据"},{"key":"C","text":"断开连接"},{"key":"D","text":"错误检测"}]', '["A"]', '三次握手建立可靠连接', 'TCP', 2, 1, 1),
(601, 1, 2, 'TCP 四次挥手用于？', '[{"key":"A","text":"建立连接"},{"key":"B","text":"传输数据"},{"key":"C","text":"断开连接"},{"key":"D","text":"加密"}]', '["C"]', '四次挥手用于断开连接', 'TCP', 3, 1, 1),
(601, 1, 2, 'TIME_WAIT 状态持续时间通常是？', '[{"key":"A","text":"30秒"},{"key":"B","text":"1分钟"},{"key":"C","text":"2分钟"},{"key":"D","text":"5分钟"}]', '["C"]', 'TIME_WAIT 通常为 2MSL', 'TCP', 3, 1, 1),
(601, 3, 1, 'UDP 是面向连接的协议。', NULL, '[false]', 'UDP 是无连接的', 'TCP', 2, 1, 1),
(601, 3, 2, 'TCP 的滑动窗口用于流量控制。', NULL, '[true]', '滑动窗口实现流量控制', 'TCP', 2, 1, 1);

-- HTTP/HTTPS (id=602)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(602, 1, 1, 'HTTP 状态码 404 表示？', '[{"key":"A","text":"成功"},{"key":"B","text":"服务器错误"},{"key":"C","text":"未找到资源"},{"key":"D","text":"重定向"}]', '["C"]', '404 Not Found', 'HTTP', 2, 1, 1),
(602, 1, 2, 'HTTP 协议是？', '[{"key":"A","text":"无状态"},{"key":"B","text":"有状态"},{"key":"C","text":"长连接"},{"key":"D","text":"面向连接"}]', '["A"]', 'HTTP 是无状态协议', 'HTTP', 3, 1, 1),
(602, 1, 2, 'HTTPS 默认端口是？', '[{"key":"A","text":"80"},{"key":"B","text":"443"},{"key":"C","text":"8080"},{"key":"D","text":"8443"}]', '["B"]', 'HTTPS 默认 443', 'HTTPS', 2, 1, 1),
(602, 3, 1, 'HTTP 状态码 500 表示客户端错误。', NULL, '[false]', '500 是服务器错误', 'HTTP', 2, 1, 1);

-- 排序算法 (id=401)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(401, 1, 2, '快速排序的最坏时间复杂度是？', '[{"key":"A","text":"O(n)"},{"key":"B","text":"O(n log n)"},{"key":"C","text":"O(n²)"},{"key":"D","text":"O(2^n)"}]', '["C"]', '快排最坏 O(n²)，平均 O(n log n)', '算法', 3, 1, 1),
(401, 1, 2, '归并排序的空间复杂度是？', '[{"key":"A","text":"O(1)"},{"key":"B","text":"O(log n)"},{"key":"C","text":"O(n)"},{"key":"D","text":"O(n²)"}]', '["C"]', '归并需要 O(n) 额外空间', '算法', 3, 1, 1);

-- 动态规划 (id=403)
INSERT INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(403, 1, 3, '动态规划的核心思想是？', '[{"key":"A","text":"分治"},{"key":"B","text":"记忆化 + 状态转移"},{"key":"C","text":"贪心"},{"key":"D","text":"递归"}]', '["B"]', 'DP = 状态转移方程 + 记忆化', 'DP', 4, 1, 1),
(403, 1, 3, '0-1 背包问题的时间复杂度是？', '[{"key":"A","text":"O(n)"},{"key":"B","text":"O(n log n)"},{"key":"C","text":"O(n × W)"},{"key":"D","text":"O(2^n)"}]', '["C"]', 'W 是背包容量', 'DP', 4, 1, 1);


-- ============================================================
-- 全状态覆盖补充数据
-- 覆盖每张表的所有枚举值（status、role、type、strategy 等）
-- ============================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 1. 用户多状态覆盖 (role, status, gender)
-- ============================================================
INSERT IGNORE INTO ke_user (username, password, nickname, email, phone, role, gender, status, class_id, department_id) VALUES
('disabled_user', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '已禁用用户', 'disabled@koala.com', '13900000000', 3, 1, 0, 1, 1),
('teacher_li',    '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '李老师',   'li@koala.com',     '13900000001', 2, 2, 1, NULL, 2),
('teacher_wang',  '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '王老师',   'wang@koala.com',   '13900000002', 2, 1, 1, NULL, 3),
('teacher_zhang', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '张老师',   'zhang@koala.com',  '13900000003', 2, 2, 1, NULL, 4),
('teacher_chen',  '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '陈老师',   'chen@koala.com',   '13900000004', 2, 0, 1, NULL, 5),
('student_anna',  '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '安娜',     'anna@koala.com',   '13900000005', 3, 2, 1, 1, 1),
('student_bob',   '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '鲍勃',     'bob@koala.com',    '13900000006', 3, 1, 1, 2, 2),
('student_cathy', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '凯西',     'cathy@koala.com',  '13900000007', 3, 2, 1, 3, 3),
('student_david', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '大卫',     'david@koala.com',  '13900000008', 3, 1, 1, 1, 1),
('student_emily', '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '艾米丽',   'emily@koala.com',  '13900000009', 3, 0, 1, 2, 2),
('student_fred',  '$2a$10$NhbhLMr5VwOnBGdblXlG8emJo/8FTsZlUFdVDF//M5FI6RzzwuEsK', '弗雷德',   'fred@koala.com',   '13900000010', 3, 1, 0, 3, 4);

-- ============================================================
-- 2. 题目状态覆盖 (status: 0/1/2, type: 1-6 全类型, difficulty: 1-5)
-- ============================================================
-- 多选题补充
INSERT IGNORE INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(301, 2, 2, 'Go 语言中以下哪些是值类型？', '[{"key":"A","text":"int"},{"key":"B","text":"slice"},{"key":"C","text":"struct"},{"key":"D","text":"map"}]', '["A","C"]', 'int 和 struct 是值类型', 'Go', 4, 1, 1),
(202, 2, 2, '下列哪些是 ES6 引入的新特性？', '[{"key":"A","text":"箭头函数"},{"key":"B","text":"class"},{"key":"C","text":"Promise"},{"key":"D","text":"var"}]', '["A","B","C"]', 'ES6 引入箭头函数、class、Promise', 'ES6', 4, 1, 1),
(501, 2, 3, 'MySQL 索引可以提升哪些场景？', '[{"key":"A","text":"WHERE 条件"},{"key":"B","text":"JOIN 关联"},{"key":"C","text":"ORDER BY 排序"},{"key":"D","text":"INSERT 插入"}]', '["A","B","C"]', '索引不提升 INSERT 性能', 'MySQL', 4, 1, 1),
(602, 2, 2, 'HTTP 协议中哪些是安全的请求方法？', '[{"key":"A","text":"GET"},{"key":"B","text":"POST"},{"key":"C","text":"HEAD"},{"key":"D","text":"OPTIONS"}]', '["A","C","D"]', 'POST 会修改服务端状态', 'HTTP', 4, 1, 1),
(203, 2, 3, 'Vue 3 中以下哪些是组合式 API 的钩子函数？', '[{"key":"A","text":"onMounted"},{"key":"B","text":"useState"},{"key":"C","text":"computed"},{"key":"D","text":"ref"}]', '["A","C"]', 'onMounted 是生命周期钩子，computed 是响应式计算', 'Vue', 4, 1, 1),
(302, 2, 3, 'Java 中以下哪些是线程安全的集合？', '[{"key":"A","text":"ConcurrentHashMap"},{"key":"B","text":"HashMap"},{"key":"C","text":"CopyOnWriteArrayList"},{"key":"D","text":"ArrayList"}]', '["A","C"]', 'ConcurrentHashMap 和 CopyOnWriteArrayList 是线程安全的', 'Java', 4, 1, 1);

-- 判断题补充
INSERT IGNORE INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(301, 3, 1, 'Go 语言中 defer 用于延迟函数执行。', NULL, '[true]', 'defer 用于延迟调用', 'Go', 2, 1, 1),
(202, 3, 1, 'JavaScript 中 undefined == null。', NULL, '[true]', '松散相等下 undefined 和 null 相等', 'JS', 2, 1, 1),
(203, 3, 1, 'Vue 中 v-show 通过 display:none 隐藏元素。', NULL, '[true]', 'v-show 控制 display 属性', 'Vue', 2, 1, 1),
(501, 3, 2, 'MySQL 中 COUNT(*) 不会统计 NULL 行。', NULL, '[false]', 'COUNT(*) 统计所有行，包括 NULL', 'MySQL', 2, 1, 1),
(503, 3, 2, 'Redis 支持事务回滚。', NULL, '[false]', 'Redis 事务不支持回滚', 'Redis', 2, 1, 1),
(602, 3, 1, 'HTTP 是基于 TCP 的应用层协议。', NULL, '[true]', 'HTTP 建立在 TCP 之上', 'HTTP', 2, 1, 1);

-- 填空题补充
INSERT IGNORE INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(101, 4, 1, '1 KB 等于 ____ 字节。', NULL, '["1024"]', '1KB = 1024B', '计算机基础', 2, 1, 1),
(103, 4, 2, 'OSI 参考模型分为 ____ 层。', NULL, '["7"]', 'OSI 是 7 层模型', '网络', 2, 1, 1),
(601, 4, 2, 'IP 地址 IPv4 共 ____ 位二进制。', NULL, '["32"]', 'IPv4 是 32 位', '网络', 2, 1, 1),
(505, 4, 2, 'SQL 中用于过滤的关键字是 ____。', NULL, '["WHERE"]', 'WHERE 用于过滤', 'SQL', 2, 1, 1),
(301, 4, 2, 'Go 中用于声明变量的关键字是 ____。', NULL, '["var"]', 'var 用于声明变量', 'Go', 2, 1, 1);

-- 不定项选择题 (type=5)
INSERT IGNORE INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(301, 5, 3, '下列哪些技术可以用于 Go 服务的性能优化？', '[{"key":"A","text":"pprof 性能分析"},{"key":"B","text":"连接池复用"},{"key":"C","text":"内存对齐"},{"key":"D","text":"泛型"}]', '["A","B","C"]', '泛型不影响运行性能', 'Go', 5, 1, 1),
(501, 5, 3, '下列哪些会影响 MySQL 查询性能？', '[{"key":"A","text":"索引失效"},{"key":"B","text":"数据量过大"},{"key":"C","text":"锁竞争"},{"key":"D","text":"字段编码"}]', '["A","B","C","D"]', '都会影响查询性能', 'MySQL', 5, 1, 1),
(601, 5, 3, '下列哪些措施可以提升网络安全？', '[{"key":"A","text":"HTTPS 加密"},{"key":"B","text":"SQL 参数化"},{"key":"C","text":"XSS 过滤"},{"key":"D","text":"CSRF Token"}]', '["A","B","C","D"]', '都是常见安全措施', '安全', 5, 1, 1),
(202, 5, 3, 'JS 异步处理方式包括哪些？', '[{"key":"A","text":"回调函数"},{"key":"B","text":"Promise"},{"key":"C","text":"async/await"},{"key":"D","text":"Generator"}]', '["A","B","C","D"]', '四种都是 JS 异步方式', 'JS', 5, 1, 1),
(403, 5, 4, 'DP 优化的常见方向包括？', '[{"key":"A","text":"状态压缩"},{"key":"B","text":"滚动数组"},{"key":"C","text":"单调队列"},{"key":"D","text":"斜率优化"}]', '["A","B","C","D"]', '都是 DP 优化方法', 'DP', 5, 1, 1);

-- 编程题 (type=6)
INSERT IGNORE INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(301, 6, 3, '编写 Go 函数：计算两个整数的最大公约数（GCD）。要求使用欧几里得算法。', NULL, '["func gcd(a, b int) int { for b != 0 { a, b = b, a%b } return a }"]', '辗转相除法', 'Go', 10, 1, 1),
(303, 6, 3, '编写 Python 函数：判断字符串是否为回文。', NULL, '["def is_palindrome(s): return s == s[::-1]"]', '反转比较', 'Python', 10, 1, 1),
(302, 6, 4, '编写 Java 方法：实现单例模式（双重检查锁）。', NULL, '["public class Singleton { private static volatile Singleton instance; private Singleton(){} public static Singleton getInstance(){ if(instance==null){ synchronized(Singleton.class){ if(instance==null){ instance=new Singleton(); }}} return instance; }}"]', '双重检查锁', 'Java', 12, 1, 1),
(406, 6, 4, '编写算法：二叉树的中序遍历。', NULL, '["func inorder(root *TreeNode) []int { res := []int{}; if root != nil { res = append(res, inorder(root.Left)...); res = append(res, root.Val); res = append(res, inorder(root.Right)...) }; return res }"]', '递归中序', '算法', 12, 1, 1),
(401, 6, 3, '实现快速排序算法。', NULL, '["func quicksort(arr []int) []int { if len(arr) <= 1 { return arr }; pivot := arr len(arr)/2]; left := []int{}; right := []int{}; for _, v := range arr { if v < pivot { left = append(left, v) } else if v > pivot { right = append(right, v) } }; return append(append(quicksort(left), pivot), quicksort(right)...)}"]', '快排', '算法', 12, 1, 1);

-- 难度 4 和 5 补充
INSERT IGNORE INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(403, 1, 4, '区间 DP 的时间复杂度通常是？', '[{"key":"A","text":"O(n)"},{"key":"B","text":"O(n²)"},{"key":"C","text":"O(n³)"},{"key":"D","text":"O(2^n)"}]', '["C"]', '区间 DP 通常为 O(n³)', 'DP', 4, 1, 1),
(405, 1, 4, 'Dijkstra 算法不能处理哪种边？', '[{"key":"A","text":"正权边"},{"key":"B","text":"负权边"},{"key":"C","text":"零权边"},{"key":"D","text":"自环边"}]', '["B"]', 'Dijkstra 不支持负权', '图论', 4, 1, 1),
(301, 1, 5, 'GMP 模型中的 G 代表什么？', '[{"key":"A","text":"Goroutine"},{"key":"B","text":"Goroutine 的协程对象"},{"key":"C","text":"Go Scheduler"},{"key":"D","text":"Garbage Collector"}]', '["B"]', 'G 是 Goroutine 的运行时对象', 'Go', 5, 1, 1),
(602, 1, 5, 'HTTPS 中 TLS 1.3 默认的密钥交换是？', '[{"key":"A","text":"RSA"},{"key":"B","text":"DH"},{"key":"C","text":"ECDHE"},{"key":"D","text":"PSK"}]', '["C"]', 'TLS 1.3 默认 ECDHE', 'HTTPS', 5, 1, 1),
(501, 1, 4, 'InnoDB 的 Buffer Pool 默认大小由谁决定？', '[{"key":"A","text":"innodb_buffer_pool_size"},{"key":"B","text":"max_connections"},{"key":"C","text":"key_buffer_size"},{"key":"D","text":"query_cache_size"}]', '["A"]', 'innodb_buffer_pool_size', 'MySQL', 4, 1, 1);

-- 题目状态：禁用 (0) 和 审核中 (2)
INSERT IGNORE INTO ke_question (category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
(101, 1, 1, '[旧题] 计算机第一台诞生于哪一年？', '[{"key":"A","text":"1940"},{"key":"B","text":"1946"},{"key":"C","text":"1950"},{"key":"D","text":"1955"}]', '["B"]', 'ENIAC 1946', '历史', 2, 1, 0),
(202, 1, 1, '[草稿] JavaScript 数组去重方法？', '[{"key":"A","text":"Set"},{"key":"B","text":"filter"},{"key":"C","text":"for循环"},{"key":"D","text":"以上都是"}]', '["D"]', '三种都可以', 'JS', 2, 1, 2),
(301, 1, 2, '[审核] Go 的 panic 可以被 recover 捕获？', '[{"key":"A","text":"能"},{"key":"B","text":"不能"}]', '["A"]', 'recover 捕获 panic', 'Go', 2, 1, 2);

-- ============================================================
-- 3. 试卷策略与状态覆盖
-- ============================================================
INSERT IGNORE INTO ke_paper (title, description, strategy, total_score, duration, pass_score, status, creator_id, config_rule) VALUES
('Java 综合笔试卷',           '覆盖 Java 基础到高级的完整笔试卷',    1, 100, 120, 60, 1, 1, '{"single":10,"multi":5,"judge":10,"fill":5}'),
('Python 编程实战',           'Python 编程能力测试（自动随机抽题）', 2, 80,  90,  48, 1, 1, '{"pool_size":40,"select":20}'),
('算法能力测试 (混合组卷)',   '基于遗传算法的智能组卷',               3, 100, 150, 60, 1, 2, '{"ga":{"pop":30,"iter":50,"mutation":0.1}}'),
('数据库原理 [草稿]',          '待完善',                               1, 50,  60,  30, 0, 2, '{}'),
('2024 春季模拟卷 [归档]',    '历史归档试卷',                         1, 100, 90,  60, 2, 1, '{}'),
('前端面试题 (固定组卷)',     'HTML/CSS/JS/Vue 完整覆盖',              1, 60,  45,  36, 1, 2, '{"single":20,"multi":5}'),
('系统架构模拟卷',            '高难度架构综合测试',                   3, 100, 120, 60, 1, 1, '{"ga":{"pop":50,"iter":100}}');

-- 试卷题目关联
INSERT IGNORE INTO ke_paper_question (paper_id, question_id, score, sort) VALUES
(6, 22, 5, 1), (6, 23, 5, 2), (6, 24, 5, 3), (6, 25, 5, 4), (6, 26, 5, 5),
(7, 27, 4, 1), (7, 28, 4, 2), (7, 29, 4, 3), (7, 30, 4, 4), (7, 31, 4, 5),
(8, 32, 5, 1), (8, 33, 5, 2), (8, 34, 5, 3), (8, 35, 5, 4), (8, 36, 5, 5);

-- ============================================================
-- 4. 考试状态全覆盖
-- ============================================================
INSERT IGNORE INTO ke_exam (title, description, paper_id, start_time, end_time, duration, max_attempts, shuffle_q, shuffle_opt, anti_cheat, status, creator_id) VALUES
('【草稿】新员工入职考试',     '尚未发布的考试模板',                  6, '2026-09-01 09:00:00', '2026-12-31 18:00:00', 60, 1, 1, 1, 1, 0, 1),
('【归档】2023春季期末考试',   '历史归档',                            5, '2023-06-01 09:00:00', '2023-06-15 18:00:00', 120, 1, 1, 1, 1, 2, 1),
('Python 编程周末考',         'Python 实战能力测试',                  7, '2024-01-01 09:00:00', '2099-12-31 23:59:59', 90, 2, 0, 1, 0, 1, 2),
('Java 全栈开发认证考试',     '官方认证',                            6, '2024-01-01 09:00:00', '2099-12-31 23:59:59', 120, 1, 1, 1, 1, 1, 1),
('【限时】算法竞赛选拔',       '高难度限时考试',                       8, '2024-01-01 09:00:00', '2099-12-31 23:59:59', 150, 1, 1, 1, 1, 1, 1);

-- ============================================================
-- 5. 考试记录状态全覆盖 (status: 0/1/2/3, passed: 0/1)
-- ============================================================
INSERT IGNORE INTO ke_exam_record (exam_id, user_id, paper_snapshot, answers, status, start_time, submit_time, duration, total_score, objective_score, subjective_score, passed, score_hash, tab_switch_cnt, audit_log) VALUES
(1, 12, '{}', '{}', 0, '2024-04-01 10:00:00', NULL, 0, 0, 0, 0, 0, '', 0, ''),
(2, 13, '{}', '{}', 0, '2024-04-01 11:00:00', NULL, 0, 0, 0, 0, 0, '', 0, ''),
(3, 14, '{"questions":[]}', '{"1":["A"]}', 1, '2024-04-02 09:00:00', '2024-04-02 10:30:00', 5400, 0, 0, 0, 0, '', 1, '[]'),
(4, 15, '{"questions":[]}', '{"1":["B"]}', 1, '2024-04-02 14:00:00', '2024-04-02 15:20:00', 4800, 0, 0, 0, 0, '', 0, '[]'),
(1, 16, '{"questions":[]}', '{"1":["B"]}', 2, '2024-03-15 09:00:00', '2024-03-15 10:00:00', 3600, 85, 80, 5, 1, 'sha256-abc123', 0, '[]'),
(2, 17, '{"questions":[]}', '{"1":["A"]}', 2, '2024-03-20 09:00:00', '2024-03-20 10:15:00', 4500, 42, 40, 2, 0, 'sha256-def456', 3, '[{"event":"tab_switch","time":3}]'),
(3, 18, '{"questions":[]}', '{"1":["B"]}', 2, '2024-03-25 09:00:00', '2024-03-25 10:00:00', 3600, 100, 90, 10, 1, 'sha256-ghi789', 0, '[]'),
(4, 19, '{"questions":[]}', '{"1":["B"]}', 3, '2024-03-30 09:00:00', '2024-03-30 09:30:00', 1800, 0, 0, 0, 0, '', 0, '[{"event":"timeout","reason":"网络中断"}]');


-- ============================================================
-- 6. 收藏覆盖 (target_type: 1/2/3, source_type: 1/2/3)
-- ============================================================
-- ============================================================
INSERT IGNORE INTO ke_favorite_folder (user_id, name, icon, color, sort) VALUES
(12, 'Python 学习笔记', 'Star',     '#67C23A', 1),
(12, '面试准备',         'Briefcase', '#409EFF', 2),
(13, '错题本',           'WarningFilled', '#F56C6C', 1),
(14, '推荐收藏',         'StarFilled', '#E6A23C', 1),
(15, '高频考点',         'Trophy',     '#909399', 1);

INSERT IGNORE INTO ke_favorite (user_id, target_type, target_id, folder_id, source_type, difficulty, note) VALUES
(12, 1, 1,  9, 1, 1, '重要基础概念'),
(13, 1, 8,  10, 2, 3, '错题自动收藏'),
(14, 1, 22, 11, 3, 2, '系统推荐'),
(15, 1, 33, 12, 1, 2, '面试必考'),
(12, 2, 6,  9,  1, 0, 'Java 综合卷'),
(13, 2, 7,  10, 3, 0, 'Python 实战'),
(14, 2, 8,  11, 3, 0, '算法测试'),
(12, 3, 301, 9,  1, 0, 'Go 语言重点'),
(13, 3, 203, 10, 2, 0, 'Vue.js 收藏'),
(14, 3, 501, 11, 3, 0, '数据库推荐');

-- ============================================================
-- 7. 错题日志覆盖 (mastery_level: 1-5 全覆盖, is_reviewed: 0/1)
-- ============================================================
INSERT IGNORE INTO ke_wrong_log (user_id, question_id, wrong_count, last_wrong_at, is_reviewed, mastery_level, user_answer, correct_answer) VALUES
(12, 22, 1, '2024-04-05 14:30:00', 1, 5, '["C"]', '["B"]'),
(13, 23, 2, '2024-04-05 15:00:00', 1, 4, '["C"]', '["A"]'),
(14, 24, 3, '2024-04-05 16:00:00', 0, 3, '["B"]', '["A"]'),
(15, 25, 2, '2024-04-05 17:00:00', 0, 2, '["A"]', '["C"]'),
(16, 26, 4, '2024-04-06 09:00:00', 0, 1, '["D"]', '["B"]'),
(17, 27, 1, '2024-04-06 10:00:00', 1, 5, '["A","C","D"]', '["A","B","C"]'),
(18, 28, 2, '2024-04-06 11:00:00', 0, 4, '["A","B"]', '["B","C"]'),
(19, 29, 1, '2024-04-06 14:00:00', 0, 3, '["func"]', '["function"]'),
(12, 30, 1, '2024-04-06 15:00:00', 1, 4, '["True"]', '["False"]'),
(13, 31, 2, '2024-04-06 16:00:00', 0, 2, '["OSI"]', '["7"]');

SET FOREIGN_KEY_CHECKS = 1;
