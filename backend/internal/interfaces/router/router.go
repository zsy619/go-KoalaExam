package router

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/your-team/koala-exam-backend/internal/application/exam"
	"github.com/your-team/koala-exam-backend/internal/application/favorite"
	"github.com/your-team/koala-exam-backend/internal/application/grading"
	"github.com/your-team/koala-exam-backend/internal/application/question"
	"github.com/your-team/koala-exam-backend/internal/application/statistics"
	"github.com/your-team/koala-exam-backend/internal/application/user"
	"github.com/your-team/koala-exam-backend/internal/domain/consts"
	"github.com/your-team/koala-exam-backend/internal/domain/event"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/cache"
	"github.com/your-team/koala-exam-backend/internal/infrastructure/repository"
	"github.com/your-team/koala-exam-backend/internal/interfaces/handler"
	"github.com/your-team/koala-exam-backend/internal/interfaces/middleware"
	"github.com/your-team/koala-exam-backend/pkg/jwt"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(h *server.Hertz, db *gorm.DB, rdb *redis.Client, jwtHelper *jwt.Helper) {
	// 初始化仓储
	userRepo := repository.NewUserRepository(db)
	qRepo := repository.NewQuestionRepository(db)
	catRepo := repository.NewCategoryRepository(db)
	paperRepo := repository.NewPaperRepository(db)
	examRepo := repository.NewExamRepository(db)
	recordRepo := repository.NewExamRecordRepository(db)
	favRepo := repository.NewFavoriteRepository(db)
	folderRepo := repository.NewFolderRepository(db)
	wrongRepo := repository.NewWrongLogRepository(db)

	// 初始化事件总线、限流器
	bus := event.NewBus()
	loginLimiter := cache.NewRedisLoginLimiter(rdb, 5, 5*time.Minute)

	// 初始化应用服务
	tokenSvc := user.NewJwtTokenAdapter(jwtHelper)
	userApp := user.NewUserApp(userRepo, tokenSvc, bus, loginLimiter)
	qApp := question.NewQuestionApp(qRepo, catRepo)
	paperApp := question.NewPaperApp(paperRepo, qRepo)
	examApp := exam.NewExamApp(examRepo, recordRepo, paperApp, rdb, bus)
	favApp := favorite.NewFavoriteApp(db, favRepo, folderRepo, wrongRepo, nil, bus)
	gradingApp := grading.NewGradingApp(recordRepo, qRepo, examApp, favApp)

	// 初始化 Handler
	userH := handler.NewUserHandler(userApp)
	qH := handler.NewQuestionHandler(qApp)
	pH := handler.NewPaperHandler(paperApp)
	eH := handler.NewExamHandler(examApp, gradingApp, db)
	fH := handler.NewFavoriteHandler(favApp)
	statsApp := statistics.NewStatisticsApp(db, recordRepo, favRepo, wrongRepo)
	statsH := handler.NewStatisticsHandler(statsApp)

	// 全局中间件
	h.Use(middleware.CORS())
	h.Use(middleware.Recovery())
	h.Use(middleware.Audit())

	// 健康检查
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	// API v1
	v1 := h.Group("/api/v1")
	{
		// 公开接口
		v1.POST("/auth/login", userH.Login)
		v1.POST("/auth/refresh", userH.RefreshToken)
		v1.POST("/auth/logout", userH.Logout)

		// 鉴权后接口
		auth := v1.Group("", middleware.Auth(jwtHelper))
		auth.GET("/user/profile", userH.Profile)
		auth.PUT("/user/profile", userH.UpdateProfile)
		auth.PUT("/user/password", userH.ChangePassword)


		// 管理员（用户管理）
		admin := auth.Group("/admin", middleware.RequireRole(consts.RoleSuperAdmin))
		admin.POST("/users", userH.Create)
		admin.GET("/users", userH.List)
		admin.POST("/users/:id/reset-password", userH.ResetPassword)

		admin.GET("/users/:id", userH.GetByID)
		admin.PUT("/users/:id", userH.Update)
		admin.PUT("/users/:id/status", userH.ToggleStatus)
		admin.DELETE("/users/:id", userH.Delete)

		// 教师/超管（题库 + 试卷 + 考试）
		ta := auth.Group("", middleware.TeacherOrAdmin())
		ta.POST("/questions", qH.Create)
		ta.PUT("/questions/:id", qH.Update)
		ta.DELETE("/questions/:id", qH.Delete)
		ta.GET("/questions/:id", qH.Get)
		ta.GET("/questions", qH.List)
		ta.POST("/questions/import", qH.BatchImport)
		ta.GET("/question-categories", qH.ListCategories)
		ta.POST("/question-categories", qH.CreateCategory)
		ta.PUT("/question-categories/:id", qH.UpdateCategory)
		ta.DELETE("/question-categories/:id", qH.DeleteCategory)

		ta.POST("/papers", pH.Create)
		ta.PUT("/papers/:id", pH.Update)
		ta.GET("/papers", pH.List)
		ta.GET("/papers/:id", pH.Get)
		ta.DELETE("/papers/:id", pH.Delete)

		ta.POST("/exams", eH.Create)
		ta.GET("/admin/exam-records", eH.AdminListRecords)
		ta.GET("/exams", eH.List)
		ta.GET("/exams/:id", eH.Get)
		ta.PUT("/exams/:id", eH.Update)
		ta.DELETE("/exams/:id", eH.Delete)
		ta.GET("/exams/:id/records", eH.Records)
		ta.POST("/grading/subjective", eH.GradeSubjective)
		ta.POST("/grading/subjective/batch", eH.GradeSubjectiveBatch)

		// 统计
		auth.GET("/stats/me", statsH.MyLearning)
		admin.GET("/stats/dashboard", statsH.Dashboard)
		ta.GET("/stats/exam/:id", statsH.ExamOverview)

		// 学员端
		auth.GET("/exams/available", eH.Available)
		auth.POST("/exams/:id/start", eH.Start)
		auth.POST("/exams/answer", eH.SaveAnswer)
		auth.POST("/exams/audit", eH.Audit)
		auth.POST("/exams/submit", eH.Submit)
		auth.GET("/exam-records/mine", eH.MyRecords)
		auth.GET("/exam-records/:id", eH.GetRecord)

		// 收藏 / 错题本（深度收藏核心）
		auth.POST("/favorites/toggle", fH.Toggle)
		auth.POST("/favorites/batch", fH.BatchAdd)
		auth.GET("/favorites/check", fH.IsFavorited)
		auth.GET("/favorites", fH.List)
		auth.GET("/favorite-folders", fH.ListFolders)
		auth.POST("/favorite-folders", fH.CreateFolder)
		auth.DELETE("/favorite-folders/:id", fH.DeleteFolder)
		auth.GET("/wrong-book", fH.GetWrongBook)
		auth.GET("/wrong-book/distribution", fH.MasteryDistribution)
		auth.GET("/favorites/stats", fH.GetStats)
		auth.POST("/wrong-log/:id/reviewed", fH.MarkReviewed)
	}
}
