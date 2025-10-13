package app

import (
	"context"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gradeflow/internal/config"
	"gradeflow/internal/controllers"
	edb "gradeflow/internal/migrations"
	"gradeflow/pkg/middleware"
	"gradeflow/pkg/utils"
)

type App struct {
	Cfg    config.Config
	DB     *gorm.DB
	RDB    *redis.Client
	Engine *gin.Engine
	Tokens utils.TokenStore
	MinIO  *minio.Client
}

func New(pgURL string) *App {
	cfg := config.Load()
	gdb, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("gorm open: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := edb.AutoMigrateUp(ctx, gdb); err != nil {
		log.Fatalf("migrate up: %v", err)
	}

	// Redis (optional)
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	tokenStore := utils.NewRedisTokenStore(rdb)

	r := gin.New()
	r.Use(gin.Recovery(), cors.Default())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public API (no auth) and Private API (JWT)
	apiPublic := r.Group("/api")
	api := r.Group("/api")
	api.Use(middleware.JWT(cfg, gdb))
	auth := controllers.NewAuthController(gdb, cfg, tokenStore)
	// Auth endpoints manage their own protection for /me and /logout
	auth.RegisterRoutes(apiPublic.Group("/auth"))

	// Departments (admin or dean for mutations; already enforced inside controllers if any)
	dept := controllers.NewDepartmentController(gdb, cfg)
	dept.RegisterRoutes(api)

	prog := controllers.NewProgramController(gdb, cfg)
	prog.RegisterRoutes(api)

	subj := controllers.NewSubjectController(gdb, cfg)
	subj.RegisterRoutes(api)

	grp := controllers.NewGroupController(gdb, cfg)
	grp.RegisterRoutes(api)

	stu := controllers.NewStudentController(gdb, cfg)
	stu.RegisterRoutes(api)

	course := controllers.NewCourseController(gdb, cfg)
	course.RegisterRoutes(api)

	enr := controllers.NewEnrollmentController(gdb, cfg)
	enr.RegisterRoutes(api)

	lesson := controllers.NewLessonController(gdb, cfg)
	lesson.RegisterRoutes(api)

	att := controllers.NewAttendanceController(gdb, cfg)
	att.RegisterRoutes(api)

	asm := controllers.NewAssessmentController(gdb, cfg)
	asm.RegisterRoutes(api)

	gr := controllers.NewAssessmentGradeController(gdb, cfg)
	gr.RegisterRoutes(api)

	acs := controllers.NewAcademicSessionController(gdb, cfg)
	acs.RegisterRoutes(api)

	// Schedule and Journal
	sch := controllers.NewScheduleController(gdb, cfg)
	sch.RegisterRoutes(api)

	jr := controllers.NewJournalController(gdb, cfg)
	jr.RegisterRoutes(api)

	// Teachers, Staff, Admins
	teacher := controllers.NewTeacherController(gdb, cfg)
	teacher.RegisterRoutes(api)

	staff := controllers.NewStaffController(gdb, cfg)
	staff.RegisterRoutes(api)

	admin := controllers.NewAdminController(gdb, cfg)
	admin.RegisterRoutes(api)

	// Practices and Exam Sessions / Attempts / Credits
	pr := controllers.NewPracticeController(gdb, cfg)
	pr.RegisterRoutes(api)

	exs := controllers.NewExamSessionController(gdb, cfg)
	exs.RegisterRoutes(api)

	ea := controllers.NewExamAttemptController(gdb, cfg)
	ea.RegisterRoutes(api)

	cr := controllers.NewCreditController(gdb, cfg)
	cr.RegisterRoutes(api)

	// MinIO (optional)
	var mc *minio.Client
	if cfg.MinIO.Endpoint != "" && cfg.MinIO.AccessKey != "" {
		client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
			Secure: cfg.MinIO.UseSSL,
		})
		if err != nil {
			log.Printf("minio init error: %v", err)
		} else {
			mc = client
			// Ensure buckets
			for _, b := range []string{cfg.MinIO.BucketAvatars, cfg.MinIO.BucketReports} {
				if b == "" {
					continue
				}
				exists, err := mc.BucketExists(ctx, b)
				if err == nil && !exists {
					if err := mc.MakeBucket(ctx, b, minio.MakeBucketOptions{}); err != nil {
						log.Printf("minio make bucket %s: %v", b, err)
					}
				}
			}
		}
	}

	return &App{Cfg: cfg, DB: gdb, RDB: rdb, Engine: r, Tokens: tokenStore, MinIO: mc}
}
