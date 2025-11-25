package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/handlers"
	"mhp-rooms/internal/info"
	"mhp-rooms/internal/infrastructure/persistence"
	"mhp-rooms/internal/infrastructure/sse"
	"mhp-rooms/internal/infrastructure/storage"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/repository"
)

type Application struct {
	config             *config.Config
	db                 persistence.DBAdapter
	repo               *repository.Repository
	authHandler        *handlers.AuthHandler
	roomHandler        *handlers.RoomHandler
	roomDetailHandler  *handlers.RoomDetailHandler
	roomJoinHandler    *handlers.RoomJoinHandler
	roomMessageHandler *handlers.RoomMessageHandler
	sseTokenHandler    *handlers.SSETokenHandler
	pageHandler        *handlers.PageHandler
	reactionHandler    *handlers.ReactionHandler
	gameVersionHandler *handlers.GameVersionHandler
	profileHandler     *handlers.ProfileHandler
	userHandler        *handlers.UserHandler
	followHandler      *handlers.FollowHandler
	reportHandler      *handlers.ReportHandler
	infoHandler        *handlers.InfoHandler
	roadmapHandler     *handlers.RoadmapHandler
	operatorHandler    *handlers.OperatorHandler
	blogHandler        *handlers.BlogHandler
	guideHandler       *handlers.StaticPageHandler
	faqHandler         *handlers.StaticPageHandler
	termsHandler       *handlers.StaticPageHandler
	privacyHandler     *handlers.StaticPageHandler
	authMiddleware     *middleware.JWTAuth
	securityConfig     *middleware.SecurityConfig
	generalLimiter     *middleware.RateLimiter
	authLimiter        *middleware.RateLimiter
	contactLimiter     *middleware.RateLimiter
	sseHub             *sse.Hub
}

func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		config: cfg,
	}

	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("データベース初期化エラー: %w", err)
	}

	if err := app.initHandlers(); err != nil {
		return nil, fmt.Errorf("ハンドラー初期化エラー: %w", err)
	}

	return app, nil
}

func (app *Application) initDatabase() error {
	if err := persistence.WaitForDB(app.config, 30, 2*time.Second); err != nil {
		return fmt.Errorf("データベース接続待機に失敗しました: %w", err)
	}

	db, err := persistence.NewDBAdapter(app.config)
	if err != nil {
		return fmt.Errorf("データベース接続に失敗しました: %w", err)
	}
	app.db = db

	if app.config.Migration.AutoRun {
		if err := app.db.Migrate(); err != nil {
			return fmt.Errorf("マイグレーションに失敗しました: %w", err)
		}
	} else {
		log.Println("マイグレーションはスキップされました（RUN_MIGRATION=trueで有効化）")
	}

	return nil
}

func (app *Application) initHandlers() error {
	app.repo = repository.NewRepository(app.db)
	articleGenerator := info.NewGenerator("static/generated/info", info.DefaultContentSources())
	blogGenerator := info.NewGenerator("static/generated/blog", []info.ContentSource{
		{Dir: "content/blog", DefaultCategory: info.ArticleTypeBlogTechnical},
	})

	// SSE Hubを初期化
	app.sseHub = sse.NewHub()
	go app.sseHub.Run()

	// 認証ミドルウェアの初期化（他のハンドラーより先に初期化）
	authMiddleware, err := middleware.NewJWTAuth(app.repo)
	if err != nil {
		log.Printf("JWT認証ミドルウェアの初期化に失敗しました: %v", err)
		// 本番環境では認証ミドルウェアの初期化失敗は致命的エラーとして扱う
		if app.config.IsProduction() {
			return fmt.Errorf("本番環境では認証ミドルウェアが必須です: %w", err)
		}
		log.Printf("開発環境では認証ミドルウェアなしで継続しますが、認証が必要な機能は利用できません")
	}
	app.authMiddleware = authMiddleware

	app.authHandler = handlers.NewAuthHandler(app.repo)
	app.roomHandler = handlers.NewRoomHandler(app.repo, app.sseHub)
	app.roomDetailHandler = handlers.NewRoomDetailHandler(app.repo)
	app.roomJoinHandler = handlers.NewRoomJoinHandler(app.repo)
	app.roomMessageHandler = handlers.NewRoomMessageHandler(app.repo, app.sseHub)
	app.sseTokenHandler = handlers.NewSSETokenHandler(app.repo)
	app.pageHandler = handlers.NewPageHandler(app.repo, articleGenerator)
	app.reactionHandler = handlers.NewReactionHandler(app.repo)
	app.gameVersionHandler = handlers.NewGameVersionHandler(app.repo)
	app.profileHandler = handlers.NewProfileHandler(app.repo, app.authMiddleware)
	app.userHandler = handlers.NewUserHandler(app.repo)
	app.followHandler = handlers.NewFollowHandler(app.repo)
	app.infoHandler = handlers.NewInfoHandler(app.repo, articleGenerator)
	app.roadmapHandler = handlers.NewRoadmapHandler(app.repo, articleGenerator)
	app.operatorHandler = handlers.NewOperatorHandler(app.repo, articleGenerator)
	app.blogHandler = handlers.NewBlogHandler(app.repo, blogGenerator)
	app.guideHandler = handlers.NewGuideHandler(app.repo, articleGenerator)
	app.faqHandler = handlers.NewFAQHandler(app.repo, articleGenerator)
	app.termsHandler = handlers.NewTermsHandler(app.repo, articleGenerator)
	app.privacyHandler = handlers.NewPrivacyHandler(app.repo, articleGenerator)
	// GCSUploaderを初期化
	gcsUploader, err := storage.NewGCSUploader(context.Background())
	if err != nil {
		return fmt.Errorf("GCSアップローダーの初期化に失敗しました: %w", err)
	}

	app.reportHandler = handlers.NewReportHandler(app.repo.Report, app.repo.User, gcsUploader)

	// セキュリティ設定の初期化
	app.securityConfig = middleware.NewSecurityConfig()

	// レート制限器の初期化
	rateLimitConfig := middleware.DefaultRateLimitConfig()
	app.generalLimiter = middleware.NewRateLimiter(rateLimitConfig.General)
	app.authLimiter = middleware.NewRateLimiter(rateLimitConfig.Auth)
	app.contactLimiter = middleware.NewRateLimiter(rateLimitConfig.Contact)

	app.authHandler.SetAuthMiddleware(authMiddleware)

	return nil
}

func (app *Application) Close() {
	if app.db != nil {
		app.db.Close()
	}
}
