# 🐨 KoalaExam Frontend

基于 **Vue 3 + TypeScript + Vite + Pinia + Element Plus** 的考拉智测前端。

## 📂 目录

```
src/
├── api/           # Axios 封装 + 各模块 API
├── assets/        # 静态资源
├── components/    # 公共组件 / 业务组件
├── composables/   # 组合式函数（hooks）
│   ├── useExamTimer.ts    # 考试倒计时
│   ├── useAntiCheat.ts    # 防作弊（切屏/复制粘贴）
│   └── useFavorite.ts     # 收藏乐观更新
├── layouts/       # 布局（管理端/学员端）
├── router/        # 路由 + 守卫
├── store/         # Pinia 状态管理
│   ├── modules/user.ts       # 用户
│   ├── modules/favorite.ts   # 收藏（核心）\n│   └── modules/exam.ts       # 考试进行中
├── types/         # TS 类型
├── utils/         # 工具
└── views/         # 页面
    ├── auth/      # 登录
    ├── admin/     # 管理端
    ├── student/   # 学员端
    └── error/     # 403/404
```

## 🚀 启动

```bash
pnpm install     # 或 npm install
pnpm dev         # 启动开发服务器 http://localhost:5173
pnpm build       # 生产构建
```

## 🔑 默认账号

-  超管 admin / koala123
-  教师 teacher / koala123
-  学员 student / koala123

## 🐳 Docker

```bash
docker build -t koala-exam-frontend -f Dockerfile .
```
