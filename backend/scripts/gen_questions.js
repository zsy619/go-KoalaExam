// 题库 SQL 生成器 - Node 版
const fs = require('fs');

function randInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

const CATS = [
  { id: 1, name: '计算机基础', tag: 'CS' },
  { id: 2, name: '前端开发', tag: 'FE' },
  { id: 3, name: '后端开发', tag: 'BE' },
  { id: 4, name: '算法与数据结构', tag: 'ALGO' },
  { id: 5, name: '数据库', tag: 'DB' },
  { id: 6, name: '网络与安全', tag: 'NET' },
];

const SINGLE_TITLES = [
  '关于 {S} 的描述，下列哪项是正确的？',
  '以下哪个选项最准确地描述了 {S} 的特性？',
  '在 {S} 领域中，下列说法正确的是？',
  '{S} 中常用的概念是以下哪一个？',
  '下列关于 {S} 的说法中，正确的是？',
  '下列哪个不属于 {S}？',
  '下列哪个是 {S} 领域的核心概念？',
  '关于 {S} 原理，下列哪项陈述最准确？',
];

const SINGLE_OPTIONS = [
  '支持事务 ACID 特性', '提供强类型检查', '运行在 JVM 上', '采用单线程模型',
  '基于事件驱动', '采用协程机制', '支持反射', '支持泛型', '拥有闭包特性',
  '提供依赖注入', '采用 MVC 架构', '支持热加载', '采用单例模式',
  '支持懒加载', '采用工厂模式', '使用观察者模式', '支持装饰器',
  '提供中间件机制', '采用 ORM 映射', '支持 SQL 注入防护', '采用连接池',
  '使用负载均衡', '采用微服务架构', '支持容器化部署', '使用分布式锁',
];

function escape(str) {
  return str.replace(/'/g, "''").replace(/\\/g, '\\\\');
}

function jsonStr(s) {
  return JSON.stringify(s).replace(/'/g, "''");
}

function singleQ(catName, idx, seed) {
  const tmpl = pick(SINGLE_TITLES);
  const title = tmpl.replace('{S}', catName);
  const correctIdx = randInt(0, 3);
  const usedOpts = new Set();
  const opts = [];
  const keys = ['A', 'B', 'C', 'D'];
  for (let i = 0; i < 4; i++) {
    let opt;
    let tries = 0;
    do {
      opt = pick(SINGLE_OPTIONS) + '（' + catName + '相关）';
      tries++;
    } while (usedOpts.has(opt) && tries < 10);
    usedOpts.add(opt);
    if (i === correctIdx) {
      opt = catName + '的核心特征之一（' + (seed % 1000) + ')';
    }
    opts.push({ key: keys[i], text: opt });
  }
  return {
    title,
    options: opts,
    answer: [keys[correctIdx]],
    analysis: '解析：本题考查' + catName + '相关知识点。正确答案' + keys[correctIdx] + '。',
  };
}

function multiQ(catName, idx, seed) {
  const correctCount = randInt(2, 3);
  const allIdx = [0, 1, 2, 3];
  // shuffle
  for (let i = allIdx.length - 1; i > 0; i--) {
    const j = randInt(0, i);
    [allIdx[i], allIdx[j]] = [allIdx[j], allIdx[i]];
  }
  const correctIdx = allIdx.slice(0, correctCount).sort((a, b) => a - b);
  const keys = ['A', 'B', 'C', 'D'];
  const opts = [];
  const correctSet = new Set(correctIdx);
  for (let i = 0; i < 4; i++) {
    let text;
    if (correctSet.has(i)) {
      text = catName + '的正确特性之一（' + (seed % 100) + '）';
    } else {
      text = '干扰项：' + pick(SINGLE_OPTIONS);
    }
    opts.push({ key: keys[i], text });
  }
  return {
    title: '以下哪些属于 ' + catName + ' 的相关特性？（多选）',
    options: opts,
    answer: correctIdx.map(i => keys[i]),
    analysis: '解析：本题为多选题，正确答案为 ' + correctIdx.map(i => keys[i]).join(',') + '。',
  };
}

function judgeQ(catName, idx, seed) {
  const isTrue = randInt(0, 1) === 1;
  const trueStatements = [
    'TCP 是面向连接的可靠传输协议。',
    'HTTP 协议默认使用 80 端口。',
    '主键不允许重复，也不允许为空。',
    'CSS 中 margin 控制元素外边距。',
    'InnoDB 存储引擎支持事务。',
    'JavaScript 是单线程语言。',
    'Vue 3 引入组合式 API。',
    'Go 语言支持 goroutine 并发。',
    'MySQL 使用 SQL 作为查询语言。',
    'HTTPS 默认端口是 443。',
  ];
  const falseStatements = [
    'Java 是弱类型语言。',
    'HTTP 是面向连接的状态协议。',
    'MyISAM 支持事务回滚。',
    '主键允许为空。',
    'TCP 三次握手是四次挥手。',
    'Go 不支持并发。',
    'SQL 中 DELETE 用于查询。',
    'HTTPS 比 HTTP 更不安全。',
    '数据库索引越多查询越快。',
    'IPv4 地址用 8 位表示。',
  ];
  const stmt = isTrue ? pick(trueStatements) : pick(falseStatements);
  return {
    title: '【判断】' + stmt,
    options: null,
    answer: [isTrue ? 'true' : 'false'],
    analysis: '解析：正确答案是' + (isTrue ? '正确' : '错误') + '。',
  };
}

function fillQ(catName, idx, seed) {
  const keywords = ['algorithm', 'function', 'variable', 'class', 'interface', 'module', 'package', 'method', 'array', 'pointer', 'coroutine', 'mutex', 'channel', 'struct', 'enum', 'lambda', 'closure', 'recursion', 'iterator', 'decorator', 'middleware', 'router', 'controller', 'service', 'repository', 'entity', 'dto', 'vo', 'bo', 'dao', 'po', 'orm', 'jwt', 'session', 'cookie', 'cache', 'queue', 'stack', 'heap', 'tree', 'graph', 'hash', 'binary', 'index', 'trigger', 'view', 'cursor', 'transaction', 'isolation', 'rollback', 'commit', 'lock', 'deadlock', 'sql', 'nosql', 'acid', 'cap', 'base'];
  const kw = pick(keywords);
  return {
    title: '【填空】在计算机相关领域中，关键术语英文是 ____（请填入英文单词）。',
    options: null,
    answer: [kw],
    analysis: '解析：答案为 ' + kw + '。',
  };
}

function essayQ(catName, idx, seed) {
  const topics = [
    '请简述相关概念及其典型应用场景。',
    '对比分析两种方案的优缺点及适用场景。',
    '描述其工作原理，并举例说明。',
    '分析常见问题及解决方案。',
    '总结最佳实践和注意事项。',
    '谈谈性能优化的常用手段。',
    '说明安全方面的风险和防护措施。',
    '从架构演进角度分析其发展。',
  ];
  return {
    title: '【简答】' + pick(topics),
    options: null,
    answer: ['open'],
    analysis: '主观题，由教师人工批改。',
  };
}

function codeQ(catName, idx, seed) {
  const problems = [
    '实现一个函数，输入数组，输出最大值。',
    '判断字符串是否为回文。',
    '实现二分查找算法。',
    '合并两个有序数组。',
    '实现快速排序。',
    '实现 LRU 缓存。',
    '实现单例模式。',
    '反转单链表。',
    '实现栈的数据结构。',
    '实现二叉树的前序遍历。',
    '实现两数之和（哈希表法）。',
    '求数组的最大子序和。',
    '判断链表是否有环。',
    '实现冒泡排序。',
    '实现选择排序。',
    '实现插入排序。',
    '字符串转整数 (atoi)。',
    '最长公共前缀。',
    '有效的括号匹配。',
    '爬楼梯问题。',
  ];
  return {
    title: '【编程】' + pick(problems),
    options: null,
    answer: ['code'],
    analysis: '编程题，需提交代码评测。',
  };
}

// ============ 生成 SQL ============
const ROWS = [];
let idCounter = 107; // 现有最大 106

function addRow(cat, type, q, tag, score, difficulty) {
  ROWS.push({
    id: idCounter++,
    category_id: cat.id,
    type,
    difficulty,
    title: q.title,
    options: q.options,
    answer: q.answer,
    analysis: q.analysis,
    tag,
    score,
    creator_id: 1,
    status: 1,
  });
}

// 1. 单选（每类约 80 题，共 ~500）
for (const cat of CATS) {
  for (let i = 0; i < 85; i++) {
    const q = singleQ(cat.name, i, idCounter);
    addRow(cat, 1, q, cat.tag, 2, randInt(1, 3));
  }
}

// 2. 多选（每类约 35 题，共 ~200）
for (const cat of CATS) {
  for (let i = 0; i < 35; i++) {
    const q = multiQ(cat.name, i, idCounter);
    addRow(cat, 2, q, cat.tag, 3, randInt(1, 3));
  }
}

// 3. 判断（每类约 20-40，共 150）
for (const cat of CATS) {
  for (let i = 0; i < 25; i++) {
    const q = judgeQ(cat.name, i, idCounter);
    addRow(cat, 3, q, '判断', 2, randInt(1, 2));
  }
}

// 4. 填空（每类 25-40，共 150）
for (const cat of CATS) {
  for (let i = 0; i < 30; i++) {
    const q = fillQ(cat.name, i, idCounter);
    addRow(cat, 4, q, '填空', 3, randInt(1, 3));
  }
}

// 5. 简答（每类 15，共 90）
for (const cat of CATS) {
  for (let i = 0; i < 15; i++) {
    const q = essayQ(cat.name, i, idCounter);
    addRow(cat, 5, q, '简答', 8, randInt(2, 3));
  }
}

// 6. 编程（80）
for (let i = 0; i < 50; i++) {
  const q = codeQ('后端', i, idCounter);
  addRow(CATS[2], 6, q, '编程', 15, 3);
}
for (let i = 0; i < 30; i++) {
  const q = codeQ('算法', i, idCounter);
  addRow(CATS[3], 6, q, '编程', 15, 3);
}

console.log('生成题目数: ' + ROWS.length);
console.log('ID 范围: ' + ROWS[0].id + ' - ' + ROWS[ROWS.length - 1].id);

// ============ 输出 SQL ============
const HEAD = `-- 自动生成的 1000+ 题库种子数据
-- 生成时间: ${new Date().toISOString()}
-- 共生成 ${ROWS.length} 题（不含原有 106 题）

INSERT INTO ke_question (id, category_id, type, difficulty, title, options, answer, analysis, tags, score, creator_id, status) VALUES
`;

const values = ROWS.map(r => {
  const options = r.options ? `'[${r.options.map(o => `{"key":"${o.key}","text":"${escape(o.text)}"}`).join(',')}]'` : 'NULL';
  const answer = `'[${r.answer.map(a => `"${escape(a)}"`).join(',')}]'`;
  const title = `'${escape(r.title)}'`;
  const analysis = `'${escape(r.analysis)}'`;
  const tag = `'${escape(r.tag)}'`;
  return `  (${r.id}, ${r.category_id}, ${r.type}, ${r.difficulty}, ${title}, ${options}, ${answer}, ${analysis}, ${tag}, ${r.score}, ${r.creator_id}, ${r.status})`;
}).join(',\n');

const SQL = HEAD + values + ';\n';

fs.writeFileSync(process.argv[2] || 'gen_questions.sql', SQL, 'utf-8');
console.log('SQL 已写入: ' + (process.argv[2] || 'gen_questions.sql'));
console.log('SQL 大小: ' + (SQL.length / 1024).toFixed(1) + ' KB');
