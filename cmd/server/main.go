package main

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mhp-rooms/internal/config"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数を使用します。")
	}

	config.Init()
	setupMimeTypes()
}

func main() {
	app, err := NewApplication(config.AppConfig)
	if err != nil {
		log.Fatalf("アプリケーションの初期化に失敗しました: %v", err)
	}
	defer app.Close()

	handler := app.SetupRoutes()

	addr := config.AppConfig.GetServerAddr()
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("サーバーを起動しています... %s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("サーバーの起動に失敗しました: %v", err)
		}
	}()

	<-shutdown
	log.Println("シャットダウンを開始しています...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("サーバーのシャットダウンに失敗しました: %v", err)
	}

	log.Println("サーバーを終了しました")
}

func setupMimeTypes() {
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".ico", "image/x-icon")
}
