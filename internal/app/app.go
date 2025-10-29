package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	"gradeflow/internal/domain/models"
	"gradeflow/internal/middleware"
	"gradeflow/internal/migrations"
	gormrepo "gradeflow/internal/repository/gorm"
	"gradeflow/internal/service"
	"gradeflow/pkg/cache"
	"gradeflow/pkg/logger"
	docs "gradeflow/pkg/swagger"
)

// App aggregates API dependencies.
type App struct {
	Cfg    config.Config
	Engine *gin.Engine
	DB     *gorm.DB
	MinIO  *minio.Client
	Redis  *redis.Client
	Cache  cache.Store
}

// New constructs the application with wired dependencies.
func New(cfg config.Config) (*App, error) {
	db, err := gorm.Open(postgres.Open(cfg.PGURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	adminBootstrap, err := migrations.Run(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	if adminBootstrap != nil {
		logger.Info("bootstrap admin credentials", "ins", adminBootstrap.INS, "password", adminBootstrap.Password)
	}

	minioClient, err := initMinIO(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init minio: %w", err)
	}

	var cacheStore cache.Store = cache.NewNoop()
	var redisClient *redis.Client
	if cfg.Redis.Addr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("redis unavailable, falling back to noop cache", "error", err)
		} else {
			cacheStore = cache.NewRedis(redisClient, 5*time.Minute)
		}
	}

	userRepo := gormrepo.NewUserRepository(db)
	groupRepo := gormrepo.NewGroupRepository(db)
	subjectRepo := gormrepo.NewSubjectRepository(db)
	sessionRepo := gormrepo.NewSessionRepository(db)
	gradeRepo := gormrepo.NewGradeRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	adminSvc := service.NewAdminService(userRepo, groupRepo, subjectRepo, sessionRepo, gradeRepo, cacheStore)
	deanSvc := service.NewDeanService(userRepo, groupRepo, subjectRepo, sessionRepo, gradeRepo, cacheStore)
	teacherSvc := service.NewTeacherService(userRepo, groupRepo, subjectRepo, sessionRepo, gradeRepo)
	studentSvc := service.NewStudentService(userRepo, groupRepo, subjectRepo, sessionRepo, gradeRepo)

	authCtrl := controllers.NewAuthController(authSvc, userRepo)
	adminCtrl := controllers.NewAdminController(adminSvc)
	deanCtrl := controllers.NewDeanController(deanSvc)
	teacherCtrl := controllers.NewTeacherController(teacherSvc)
	studentCtrl := controllers.NewStudentController(studentSvc)

	r := gin.New()
	cfgCors := cors.Config{
		AllowHeaders: []string{"Authorization", "Content-Type"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
	}
	if cfg.Cors == "*" || cfg.Cors == "" {
		cfgCors.AllowAllOrigins = true
	} else {
		cfgCors.AllowOrigins = strings.Split(cfg.Cors, ",")
	}
	r.Use(gin.Recovery(), cors.New(cfgCors))

	docs.SwaggerInfo.BasePath = cfg.APIBasePath

	r.GET("/healthz", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := r.Group(cfg.APIBasePath)
	authCtrl.RegisterPublicRoutes(public.Group("/auth"))

	private := r.Group(cfg.APIBasePath)
	private.Use(middleware.JWTAuth([]byte(cfg.JWTSecret)))
	authCtrl.RegisterPrivateRoutes(private.Group("/auth"))

	adminRoutes := private.Group("/admin")
	adminRoutes.Use(middleware.RequireRoles(string(models.UserRoleAdmin)))
	adminCtrl.RegisterRoutes(adminRoutes)

	deanRoutes := private.Group("/dean")
	deanRoutes.Use(middleware.RequireRoles(string(models.UserRoleDean)))
	deanCtrl.RegisterRoutes(deanRoutes)

	teacherRoutes := private.Group("/teacher")
	teacherRoutes.Use(middleware.RequireRoles(string(models.UserRoleTeacher)))
	teacherCtrl.RegisterRoutes(teacherRoutes)

	studentRoutes := private.Group("/student")
	studentRoutes.Use(middleware.RequireRoles(string(models.UserRoleStudent)))
	studentCtrl.RegisterRoutes(studentRoutes)

	return &App{Cfg: cfg, Engine: r, DB: db, MinIO: minioClient, Redis: redisClient, Cache: cacheStore}, nil
}

// Run starts the HTTP server.
func (a *App) Run() error {
	return a.Engine.Run(a.Cfg.HTTPAddr)
}

func initMinIO(ctx context.Context, cfg config.Config) (*minio.Client, error) {
	if cfg.MinIO.Endpoint == "" || cfg.MinIO.AccessKey == "" || cfg.MinIO.SecretKey == "" {
		return nil, nil
	}
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	for _, bucket := range []string{cfg.MinIO.BucketAvatars, cfg.MinIO.BucketReports} {
		if bucket == "" {
			continue
		}
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return nil, err
			}
		}
	}
	return client, nil
}
