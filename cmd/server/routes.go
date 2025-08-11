package main

import (
	"net/http"
	"os"

	"mhp-rooms/internal/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// isProductionEnv 本番環境かどうかを判定
func isProductionEnv() bool {
	env := os.Getenv("ENV")
	return env == "production"
}

// hasAuthMiddleware 認証ミドルウェアが有効かどうかを判定
func (app *Application) hasAuthMiddleware() bool {
	return app.authMiddleware != nil
}

// withAuth 認証ミドルウェアを適用するヘルパー関数
func (app *Application) withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.authMiddleware.Middleware(handler).ServeHTTP(w, r)
	}
}

// withOptionalAuth オプショナル認証ミドルウェアを適用するヘルパー関数
func (app *Application) withOptionalAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.authMiddleware.OptionalMiddleware(handler).ServeHTTP(w, r)
	}
}

func (app *Application) SetupRoutes() chi.Router {
	r := chi.NewRouter()

	// Chi標準ミドルウェア
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Logger)

	// グローバルミドルウェアの適用
	r.Use(middleware.SecurityHeaders(app.securityConfig))
	r.Use(middleware.RateLimitMiddleware(app.generalLimiter))

	// 静的ファイルを最初に設定（追加のミドルウェアを適用しない）
	app.setupStaticRoutes(r)

	app.setupPageRoutes(r)
	app.setupRoomRoutes(r)
	app.setupAuthRoutes(r)
	app.setupAPIRoutes(r)

	return r
}

func (app *Application) setupPageRoutes(r chi.Router) {
	ph := app.pageHandler

	// 本番環境では認証情報をオプションで取得、開発環境では認証なしでアクセス可能
	if app.hasAuthMiddleware() {
		r.Get("/", app.withOptionalAuth(ph.Home))
		r.Get("/terms", app.withOptionalAuth(ph.Terms))
		r.Get("/privacy", app.withOptionalAuth(ph.Privacy))
		r.HandleFunc("/contact", app.withOptionalAuth(ph.Contact))
		r.Get("/faq", app.withOptionalAuth(ph.FAQ))
		r.Get("/guide", app.withOptionalAuth(ph.Guide))
		r.Get("/hello", app.withOptionalAuth(ph.Hello))
		r.Get("/sitemap.xml", app.withOptionalAuth(ph.Sitemap))
	} else {
		r.Get("/", ph.Home)
		r.Get("/terms", ph.Terms)
		r.Get("/privacy", ph.Privacy)
		r.HandleFunc("/contact", ph.Contact)
		r.Get("/faq", ph.FAQ)
		r.Get("/guide", ph.Guide)
		r.Get("/hello", ph.Hello)
		r.Get("/sitemap.xml", ph.Sitemap)
	}
}

func (app *Application) setupRoomRoutes(r chi.Router) {
	r.Route("/rooms", func(rr chi.Router) {
		rh := app.roomHandler
		rdh := app.roomDetailHandler
		rmh := app.roomMessageHandler

		// 部屋一覧・詳細（本番環境では認証情報をオプションで取得、開発環境では認証なし）
		if app.hasAuthMiddleware() {
			rr.Get("/", app.withOptionalAuth(rh.Rooms))
			rr.Get("/{id}", app.withOptionalAuth(rdh.RoomDetail))
		} else {
			rr.Get("/", rh.Rooms)
			rr.Get("/{id}", rdh.RoomDetail)
		}

		// 部屋操作・メッセージ機能（本番環境では認証必須、開発環境では認証なし）
		if app.hasAuthMiddleware() {
			rr.Group(func(protected chi.Router) {
				protected.Use(app.authMiddleware.Middleware)

				protected.Post("/", rh.CreateRoom)
				protected.Put("/{id}", rh.UpdateRoom)
				protected.Delete("/{id}", rh.DismissRoom)
				protected.Post("/{id}/join", rh.JoinRoom)
				protected.Post("/{id}/leave", rh.LeaveRoom)
				protected.Put("/{id}/toggle-closed", rh.ToggleRoomClosed)

				// メッセージ関連
				protected.Post("/{id}/messages", rmh.SendMessage)
				protected.Get("/{id}/messages", rmh.GetMessages)
				protected.Post("/{id}/sse-token", app.sseTokenHandler.GenerateSSEToken)
			})

			// SSEストリーム（一時トークン認証）
			rr.Get("/{id}/messages/stream", rmh.StreamMessages)
		} else {
			rr.Post("/", rh.CreateRoom)
			rr.Put("/{id}", rh.UpdateRoom)
			rr.Delete("/{id}", rh.DismissRoom)
			rr.Post("/{id}/join", rh.JoinRoom)
			rr.Post("/{id}/leave", rh.LeaveRoom)
			rr.Put("/{id}/toggle-closed", rh.ToggleRoomClosed)

			// メッセージ関連
			rr.Post("/{id}/messages", rmh.SendMessage)
			rr.Get("/{id}/messages", rmh.GetMessages)
			rr.Post("/{id}/sse-token", app.sseTokenHandler.GenerateSSEToken)
			rr.Get("/{id}/messages/stream", rmh.StreamMessages)
		}
	})
}

func (app *Application) setupAuthRoutes(r chi.Router) {
	r.Route("/auth", func(ar chi.Router) {
		ah := app.authHandler

		// 認証ページルート（レート制限は緩め）
		ar.Get("/login", ah.LoginPage)
		ar.Get("/register", ah.RegisterPage)
		ar.Get("/password-reset", ah.PasswordResetPage)
		ar.Get("/password-reset/confirm", ah.PasswordResetConfirmPage)
		ar.Get("/callback", ah.AuthCallback)
		ar.Get("/complete-profile", ah.CompleteProfilePage)

		// 認証アクションルート（厳しいレート制限）
		ar.Group(func(arr chi.Router) {
			arr.Use(middleware.AuthRateLimitMiddleware(app.authLimiter))

			arr.Post("/login", ah.Login)
			arr.Post("/register", ah.Register)
			arr.Post("/logout", ah.Logout)
			arr.Post("/password-reset", ah.PasswordResetRequest)
			arr.Post("/password-reset/confirm", ah.PasswordResetConfirm)
			arr.Get("/google", ah.GoogleAuth)
			arr.Get("/google/callback", ah.GoogleCallback)
			arr.Post("/complete-profile", ah.CompleteProfile)
		})
	})
}

func (app *Application) setupAPIRoutes(r chi.Router) {
	r.Route("/api", func(ar chi.Router) {
		// 認証不要なAPIエンドポイント
		ar.Get("/config/supabase", app.configHandler.GetSupabaseConfig)
		ar.Get("/health", app.healthCheck)
		ar.Get("/game-versions/active", app.gameVersionHandler.GetActiveGameVersionsAPI)

		if app.hasAuthMiddleware() {
			// 認証関連API（厳しいレート制限 + 認証必須）
			ar.Route("/auth", func(apr chi.Router) {
				apr.Use(middleware.AuthRateLimitMiddleware(app.authLimiter))
				apr.Use(app.authMiddleware.Middleware)

				apr.Post("/sync", app.authHandler.SyncUser)
				apr.Put("/psn-id", app.authHandler.UpdatePSNId)
			})

			// 認証必須のAPIエンドポイント
			ar.Get("/user/current", app.withAuth(app.authHandler.CurrentUser))
			ar.Get("/user/current-room", app.withAuth(app.roomHandler.GetCurrentRoom))
			ar.Get("/user/current/room-status", app.withAuth(app.roomHandler.GetUserRoomStatus))
			ar.Post("/leave-current-room", app.withAuth(app.roomHandler.LeaveCurrentRoom))

			// リアクション関連API（認証必須）
			ar.Post("/messages/{messageId}/reactions", app.withAuth(app.reactionHandler.AddReaction))
			ar.Delete("/messages/{messageId}/reactions/{reactionType}", app.withAuth(app.reactionHandler.RemoveReaction))

			// 認証オプションのAPIエンドポイント
			ar.Get("/rooms", app.withOptionalAuth(app.roomHandler.GetAllRoomsAPI))
			ar.Get("/messages/{messageId}/reactions", app.withOptionalAuth(app.reactionHandler.GetMessageReactions))
			ar.Get("/reactions/types", app.withOptionalAuth(app.reactionHandler.GetAvailableReactions))
		} else {
			// 開発環境では認証なしですべてのAPIにアクセス可能
			ar.Get("/user/current", app.authHandler.CurrentUser)
			ar.Get("/user/current-room", app.roomHandler.GetCurrentRoom)
			ar.Get("/user/current/room-status", app.roomHandler.GetUserRoomStatus)
			ar.Post("/leave-current-room", app.roomHandler.LeaveCurrentRoom)
			ar.Post("/auth/sync", app.authHandler.SyncUser)
			ar.Put("/auth/psn-id", app.authHandler.UpdatePSNId)
			ar.Get("/rooms", app.roomHandler.GetAllRoomsAPI)

			// リアクション関連API
			ar.Post("/messages/{messageId}/reactions", app.reactionHandler.AddReaction)
			ar.Delete("/messages/{messageId}/reactions/{reactionType}", app.reactionHandler.RemoveReaction)
			ar.Get("/messages/{messageId}/reactions", app.reactionHandler.GetMessageReactions)
			ar.Get("/reactions/types", app.reactionHandler.GetAvailableReactions)
		}
	})
}

func (app *Application) setupStaticRoutes(r chi.Router) {
	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))
}

func (app *Application) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"monhub"}`))
}
