package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/handlers"
	"mhp-rooms/internal/infrastructure/persistence"
	redisinfra "mhp-rooms/internal/infrastructure/redis"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/sse"
	"mhp-rooms/internal/storage"
)

type Application struct {
	config             *config.Config
	db                 persistence.DBAdapter
	redisClient        redisinfra.RedisClient
	repo               *repository.Repository
	authHandler        *handlers.AuthHandler
	roomHandler        *handlers.RoomHandler
	roomDetailHandler  *handlers.RoomDetailHandler
	roomMessageHandler *handlers.RoomMessageHandler
	sseTokenHandler    *handlers.SSETokenHandler
	pageHandler        *handlers.PageHandler
	configHandler      *handlers.ConfigHandler
	reactionHandler    *handlers.ReactionHandler
	gameVersionHandler *handlers.GameVersionHandler
	profileHandler     *handlers.ProfileHandler
	userHandler        *handlers.UserHandler
	followHandler      *handlers.FollowHandler
	reportHandler      *handlers.ReportHandler
	authMiddleware     *middleware.JWTAuth
	securityConfig     *middleware.SecurityConfig
	generalLimiter     *middleware.RateLimiter
	authLimiter        *middleware.RateLimiter
	sseHub             *sse.Hub
	eventBus           sse.EventBus
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

func (app *Application) initRedis() error {
	if !app.config.Redis.Enabled {
		log.Println("Redis無効: インメモリモードで動作します")
		return nil
	}

	upstashToken := app.config.GetEnv("UPSTASH_TOKEN", "")

	client, err := redisinfra.NewUpstashClient(app.config.Redis.URL, upstashToken)
	if err != nil {
		return fmt.Errorf("Redis接続に失敗しました: %w", err)
	}

	app.redisClient = client
	log.Printf("Redis接続成功: %s", app.config.Redis.Host)

	return nil
}

func (app *Application) initSSE() error {
	// SSE Hubを初期化
	app.sseHub = sse.NewHub()
	go app.sseHub.Run()

	// トークンマネージャーの初期化
	if app.config.Redis.Enabled && app.redisClient != nil {
		sse.InitializeTokenManagerWithClient(app.redisClient, app.config.SSE.TokenTTL)
	} else {
		sse.InitializeTokenManagerWithClient(nil, app.config.SSE.TokenTTL)
	}

	// イベントバスの初期化
	if app.config.Redis.Enabled && app.redisClient != nil {
		sse.InitializeEventBusWithClient(app.redisClient, app.sseHub)
	} else {
		sse.InitializeEventBusWithClient(nil, app.sseHub)
	}

	app.eventBus = sse.GetEventBus()

	return nil
}

func (app *Application) initHandlers() error {
	app.repo = repository.NewRepository(app.db)

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
	app.roomMessageHandler = handlers.NewRoomMessageHandler(app.repo, app.sseHub)
	app.sseTokenHandler = handlers.NewSSETokenHandler(app.repo)
	app.pageHandler = handlers.NewPageHandler(app.repo)
	app.configHandler = handlers.NewConfigHandler()
	app.reactionHandler = handlers.NewReactionHandler(app.repo)
	app.gameVersionHandler = handlers.NewGameVersionHandler(app.repo)
	app.profileHandler = handlers.NewProfileHandler(app.repo, app.authMiddleware)
	app.userHandler = handlers.NewUserHandler(app.repo)
	app.followHandler = handlers.NewFollowHandler(app.repo)
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

	app.authHandler.SetAuthMiddleware(authMiddleware)

	return nil
}

func (app *Application) Close() {
	if app.db != nil {
		app.db.Close()
	}
}
