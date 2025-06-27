package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"time"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/handlers"
	"mhp-rooms/internal/infrastructure/auth/supabase"
	"mhp-rooms/internal/infrastructure/persistence/postgres"
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
	if err := postgres.WaitForDB(config.AppConfig, 30, 2*time.Second); err != nil {
		log.Fatalf("データベース接続待機に失敗しました: %v", err)
	}

	// データベース接続を作成
	log.Println("データベース接続を初期化中...")
	db, err := postgres.NewDB(config.AppConfig)
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

	// Supabaseクライアントを初期化
	log.Println("Supabaseクライアントを初期化中...")
	if err := supabase.Init(); err != nil {
		log.Fatalf("Supabaseクライアントの初期化に失敗しました: %v", err)
	}

	// 依存関係を構築
	repo := repository.NewRepository(db)
	h := handlers.NewHandler(repo, supabase.GetClient())

	// MIME type
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".ico", "image/x-icon")

	r := mux.NewRouter()

	r.HandleFunc("/", h.Home).Methods("GET")
	r.HandleFunc("/rooms", h.Rooms).Methods("GET")
	r.HandleFunc("/rooms", h.CreateRoom).Methods("POST")
	r.HandleFunc("/rooms/{id}/join", h.JoinRoom).Methods("POST")
	r.HandleFunc("/rooms/{id}/leave", h.LeaveRoom).Methods("POST")
	r.HandleFunc("/rooms/{id}/toggle-closed", h.ToggleRoomClosed).Methods("PUT")
	r.HandleFunc("/terms", h.Terms).Methods("GET")
	r.HandleFunc("/privacy", h.Privacy).Methods("GET")
	r.HandleFunc("/contact", h.Contact).Methods("GET", "POST")
	r.HandleFunc("/faq", h.FAQ).Methods("GET")
	r.HandleFunc("/guide", h.Guide).Methods("GET")

	// 認証関連
	r.HandleFunc("/auth/login", h.LoginPage).Methods("GET")
	r.HandleFunc("/auth/login", h.Login).Methods("POST")
	r.HandleFunc("/auth/register", h.RegisterPage).Methods("GET")
	r.HandleFunc("/auth/register", h.Register).Methods("POST")
	r.HandleFunc("/auth/logout", h.Logout).Methods("POST")

	// パスワードリセット
	r.HandleFunc("/auth/password-reset", h.PasswordResetPage).Methods("GET")
	r.HandleFunc("/auth/password-reset", h.PasswordResetRequest).Methods("POST")
	r.HandleFunc("/auth/password-reset/confirm", h.PasswordResetConfirmPage).Methods("GET")
	r.HandleFunc("/auth/password-reset/confirm", h.PasswordResetConfirm).Methods("POST")

	// Google OAuth（準備中）
	r.HandleFunc("/auth/google", h.GoogleAuth).Methods("GET")
	r.HandleFunc("/auth/google/callback", h.GoogleCallback).Methods("GET")

	// プロフィール補完
	r.HandleFunc("/auth/complete-profile", h.CompleteProfilePage).Methods("GET")
	r.HandleFunc("/auth/complete-profile", h.CompleteProfile).Methods("POST")

	// API
	r.HandleFunc("/api/user/current", h.CurrentUser).Methods("GET")
	r.HandleFunc("/api/rooms", h.GetAllRoomsAPI).Methods("GET")

	r.HandleFunc("/hello", handlers.Hello).Methods("GET")
	r.HandleFunc("/sitemap.xml", h.Sitemap).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	addr := config.AppConfig.GetServerAddr()
	fmt.Printf("サーバーを起動しています... %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
