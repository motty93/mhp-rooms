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

	r.HandleFunc("/", h.HomeHandler).Methods("GET")
	r.HandleFunc("/rooms", h.RoomsHandler).Methods("GET")
	r.HandleFunc("/rooms", h.CreateRoomHandler).Methods("POST")
	r.HandleFunc("/rooms/{id}/join", h.JoinRoomHandler).Methods("POST")
	r.HandleFunc("/rooms/{id}/leave", h.LeaveRoomHandler).Methods("POST")
	r.HandleFunc("/rooms/{id}/toggle-closed", h.ToggleRoomClosedHandler).Methods("PUT")
	r.HandleFunc("/terms", h.TermsHandler).Methods("GET")
	r.HandleFunc("/privacy", h.PrivacyHandler).Methods("GET")
	r.HandleFunc("/contact", h.ContactHandler).Methods("GET", "POST")
	r.HandleFunc("/faq", h.FAQHandler).Methods("GET")
	r.HandleFunc("/guide", h.GuideHandler).Methods("GET")

	// 認証関連
	r.HandleFunc("/auth/login", h.LoginPageHandler).Methods("GET")
	r.HandleFunc("/auth/login", h.LoginHandler).Methods("POST")
	r.HandleFunc("/auth/register", h.RegisterPageHandler).Methods("GET")
	r.HandleFunc("/auth/register", h.RegisterHandler).Methods("POST")
	r.HandleFunc("/auth/logout", h.LogoutHandler).Methods("POST")

	// パスワードリセット
	r.HandleFunc("/auth/password-reset", h.PasswordResetPageHandler).Methods("GET")
	r.HandleFunc("/auth/password-reset", h.PasswordResetRequestHandler).Methods("POST")
	r.HandleFunc("/auth/password-reset/confirm", h.PasswordResetConfirmPageHandler).Methods("GET")
	r.HandleFunc("/auth/password-reset/confirm", h.PasswordResetConfirmHandler).Methods("POST")

	// Google OAuth（準備中）
	r.HandleFunc("/auth/google", h.GoogleAuthHandler).Methods("GET")
	r.HandleFunc("/auth/google/callback", h.GoogleCallbackHandler).Methods("GET")

	// プロフィール補完
	r.HandleFunc("/auth/complete-profile", h.CompleteProfilePageHandler).Methods("GET")
	r.HandleFunc("/auth/complete-profile", h.CompleteProfileHandler).Methods("POST")

	// API
	r.HandleFunc("/api/user/current", h.CurrentUserHandler).Methods("GET")

	r.HandleFunc("/hello", handlers.HelloHandler).Methods("GET")
	r.HandleFunc("/sitemap.xml", h.SitemapHandler).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	addr := config.AppConfig.GetServerAddr()
	fmt.Printf("サーバーを起動しています... %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
