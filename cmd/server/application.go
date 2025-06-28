package main

import (
	"fmt"
	"log"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/handlers"
	"mhp-rooms/internal/infrastructure/auth/supabase"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
	"mhp-rooms/internal/repository"
)

type Application struct {
	config      *config.Config
	db          *postgres.DB
	repo        *repository.Repository
	authHandler *handlers.AuthHandler
	roomHandler *handlers.RoomHandler
	pageHandler *handlers.PageHandler
}

func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		config: cfg,
	}

	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("データベース初期化エラー: %w", err)
	}

	if err := supabase.Init(); err != nil {
		app.Close()
		return nil, fmt.Errorf("Supabaseクライアントの初期化に失敗しました: %w", err)
	}

	app.initHandlers()

	return app, nil
}

func (app *Application) initDatabase() error {
	if err := postgres.WaitForDB(app.config, 30, 2); err != nil {
		return fmt.Errorf("データベース接続待機に失敗しました: %w", err)
	}

	db, err := postgres.NewDB(app.config)
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

func (app *Application) initHandlers() {
	app.repo = repository.NewRepository(app.db)

	sc := supabase.GetClient()
	app.authHandler = handlers.NewAuthHandler(app.repo, sc)
	app.roomHandler = handlers.NewRoomHandler(app.repo, sc)
	app.pageHandler = handlers.NewPageHandler(app.repo, sc)
}

func (app *Application) Close() {
	if app.db != nil {
		app.db.Close()
	}
}
