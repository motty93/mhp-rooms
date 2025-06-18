package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"mhp-rooms/internal/database"
	"mhp-rooms/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// データベースが利用可能になるまで待機
	log.Println("データベース接続を待機中...")
	if err := database.WaitForDB(30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	// データベース接続を初期化
	if err := database.InitDB(); err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer database.CloseDB()

	// マイグレーションを実行
	log.Println("データベースマイグレーションを実行中...")
	if err := database.Migrate(); err != nil {
		log.Fatalf("マイグレーションに失敗しました: %v", err)
	}
	log.Println("マイグレーション完了")

	// MIMEタイプを設定
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".ico", "image/x-icon")

	r := mux.NewRouter()

	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/rooms", handlers.RoomsHandler).Methods("GET")
	r.HandleFunc("/hello", handlers.HelloHandler).Methods("GET")

	// ヘルスチェックエンドポイント
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// 静的ファイルの配信
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// ポート設定（Fly.ioの環境変数に対応）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("サーバーを起動しています... :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
