package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/database"
	"mhp-rooms/internal/handlers"
	"mhp-rooms/internal/repository"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数を使用します。")
	}

	// 設定を初期化
	log.Println("設定を初期化中...")
	config.Init()

	log.Println("データベース接続を待機中...")
	if err := database.WaitForDB(config.AppConfig, 30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	// データベース接続を作成
	log.Println("データベース接続を初期化中...")
	db, err := database.NewDB(config.AppConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// マイグレーション実行（環境変数で制御）
	if config.AppConfig.Migration.AutoRun {
		log.Println("データベースマイグレーションを実行中...")
		if err := db.Migrate(); err != nil {
			log.Fatalf("マイグレーションに失敗しました: %v", err)
		}
		log.Println("マイグレーション完了")
	} else {
		log.Println("マイグレーションはスキップされました（RUN_MIGRATION=trueで有効化）")
	}

	// 依存関係を構築
	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo)

	// MIME type
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".ico", "image/x-icon")

	r := mux.NewRouter()

	r.HandleFunc("/", handler.HomeHandler).Methods("GET")
	r.HandleFunc("/rooms", handler.RoomsHandler).Methods("GET")
	r.HandleFunc("/rooms", handler.CreateRoomHandler).Methods("POST")
	r.HandleFunc("/rooms/{id}/join", handler.JoinRoomHandler).Methods("POST")
	r.HandleFunc("/rooms/{id}/leave", handler.LeaveRoomHandler).Methods("POST")
	r.HandleFunc("/rooms/{id}/toggle-closed", handler.ToggleRoomClosedHandler).Methods("PUT")
	r.HandleFunc("/terms", handler.TermsHandler).Methods("GET")
	r.HandleFunc("/privacy", handler.PrivacyHandler).Methods("GET")
	r.HandleFunc("/contact", handler.ContactHandler).Methods("GET", "POST")
	r.HandleFunc("/hello", handlers.HelloHandler).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	addr := config.AppConfig.GetServerAddr()
	fmt.Printf("サーバーを起動しています... %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
