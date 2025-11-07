package main

import (
	"encoding/json"
	"net/http"
	"os"

	"mhp-rooms/internal/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

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
	if app.authMiddleware != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			app.authMiddleware.Middleware(handler).ServeHTTP(w, r)
		}
	}
	// 認証ミドルウェアが利用できない場合は認証エラーを返す
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "認証システムが初期化されていません。SUPABASE_JWT_SECRETが設定されていることを確認してください。",
		})
	}
}

// withOptionalAuth オプショナル認証ミドルウェアを適用するヘルパー関数
func (app *Application) withOptionalAuth(handler http.HandlerFunc) http.HandlerFunc {
	if app.authMiddleware != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			app.authMiddleware.OptionalMiddleware(handler).ServeHTTP(w, r)
		}
	}
	// 認証ミドルウェアが利用できない場合は認証なしで継続（オプショナルなため）
	return handler
}

func (app *Application) SetupRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Logger)
	r.Use(middleware.SecurityHeaders(app.securityConfig))
	r.Use(middleware.RateLimitMiddleware(app.generalLimiter))

	// SERVICE_MODEに基づいてルーティングを分岐
	if app.config.ServiceMode == "sse" {
		// SSE専用モード：最小限のルーティング
		app.setupSSEOnlyRoutes(r)
	} else {
		// 通常モード：全ルーティング
		// 静的ファイルを最初に設定（追加のミドルウェアを適用しない）
		app.setupStaticRoutes(r)

		app.setupPageRoutes(r)
		app.setupRoomRoutes(r)
		app.setupAuthRoutes(r)
		app.setupAPIRoutes(r)
	}

	return r
}

func (app *Application) setupPageRoutes(r chi.Router) {
	ph := app.pageHandler
	profileHandler := app.profileHandler
	infoHandler := app.infoHandler
	roadmapHandler := app.roadmapHandler
	operatorHandler := app.operatorHandler

	r.Get("/", app.withOptionalAuth(ph.Home))
	r.Get("/terms", app.withOptionalAuth(ph.Terms))
	r.Get("/privacy", app.withOptionalAuth(ph.Privacy))
	r.Get("/contact", app.withOptionalAuth(ph.Contact))
	r.With(middleware.ContactRateLimitMiddleware(app.contactLimiter)).Post("/contact", app.withOptionalAuth(ph.Contact))
	r.Get("/faq", app.withOptionalAuth(ph.FAQ))
	r.Get("/guide", app.withOptionalAuth(ph.Guide))
	r.Get("/hello", app.withOptionalAuth(ph.Hello))
	r.Get("/sitemap.xml", app.withOptionalAuth(ph.Sitemap))
	r.Get("/profile", app.withAuth(profileHandler.Profile))
	r.Get("/profile/edit", app.withAuth(profileHandler.EditForm))
	r.Get("/profile/view", app.withAuth(profileHandler.ViewProfile))
	r.Get("/users/{uuid}", app.withOptionalAuth(app.userHandler.Show))

	// 更新情報・ロードマップ（完全静的のため認証ミドルウェアを適用しない）
	r.Get("/info", infoHandler.List)
	r.Get("/info/{slug}", infoHandler.Detail)
	r.Get("/info-feed.xml", infoHandler.Feed)
	r.Get("/info-atom.xml", infoHandler.AtomFeed)
	r.Get("/roadmap", roadmapHandler.Index)
	r.Get("/operator", operatorHandler.Index)
}

func (app *Application) setupRoomRoutes(r chi.Router) {
	r.Route("/rooms", func(rr chi.Router) {
		rh := app.roomHandler
		rdh := app.roomDetailHandler
		rjh := app.roomJoinHandler
		rmh := app.roomMessageHandler

		// 部屋一覧・詳細（本番環境では認証情報をオプションで取得、開発環境では認証なし）
		if app.hasAuthMiddleware() {
			rr.Get("/", app.withOptionalAuth(rh.Rooms))
			rr.Get("/{id}", app.withOptionalAuth(rdh.RoomDetail))
			// 部屋参加ページ（スケルトン、認証必須）
			rr.Get("/{id}/join", app.withAuth(rjh.RoomJoinPage))
		} else {
			rr.Get("/", rh.Rooms)
			rr.Get("/{id}", rdh.RoomDetail)
			// 部屋参加ページ（開発環境では認証なし）
			rr.Get("/{id}/join", rjh.RoomJoinPage)
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
		})
	})
}

func (app *Application) setupAPIRoutes(r chi.Router) {
	r.Route("/api", func(ar chi.Router) {
		// 認証不要なAPIエンドポイント
		ar.Get("/config/supabase", app.configHandler.GetSupabaseConfig)
		ar.Get("/health", app.healthCheck)
		ar.Get("/game-versions/active", app.gameVersionHandler.GetActiveGameVersionsAPI)

		// 他のユーザーのプロフィール関連API（認証オプション）
		ar.Get("/users/{uuid}", app.withOptionalAuth(app.userHandler.GetUserProfile))
		ar.Get("/users/{uuid}/profile-card", app.withOptionalAuth(app.userHandler.GetProfileCard))
		ar.Get("/users/{uuid}/rooms", app.withOptionalAuth(app.userHandler.Rooms))
		ar.Get("/users/{uuid}/activity", app.withOptionalAuth(app.profileHandler.Activity))
		ar.Get("/users/{uuid}/followers", app.withOptionalAuth(app.profileHandler.Followers))
		ar.Get("/users/{uuid}/following", app.withOptionalAuth(app.profileHandler.Following))

		// 認証関連API（厳しいレート制限 + 認証必須）
		ar.Route("/auth", func(apr chi.Router) {
			if app.authMiddleware != nil {
				apr.Use(middleware.AuthRateLimitMiddleware(app.authLimiter))
				apr.Use(app.authMiddleware.Middleware)
			}
			apr.Post("/sync", app.authHandler.SyncUser)
			apr.Put("/psn-id", app.authHandler.UpdatePSNId)
		})

		// 認証必須のAPIエンドポイント
		ar.Get("/user/current", app.withAuth(app.authHandler.CurrentUser))
		ar.Get("/user/me", app.withAuth(app.authHandler.GetCurrentUser))
		ar.Get("/user/current-room", app.withAuth(app.roomHandler.GetCurrentRoom))
		ar.Get("/user/current/room-status", app.withAuth(app.roomHandler.GetUserRoomStatus))
		ar.Post("/leave-current-room", app.withAuth(app.roomHandler.LeaveCurrentRoom))

		// プロフィール関連API（認証必須）
		ar.Get("/profile/edit-form", app.withAuth(app.profileHandler.EditForm))
		ar.Get("/profile/view", app.withAuth(app.profileHandler.ViewProfile))
		ar.Post("/profile/update", app.withAuth(app.profileHandler.UpdateProfile))
		ar.Post("/profile/upload-avatar", app.withAuth(app.profileHandler.UploadAvatar))
		ar.Get("/profile/activity", app.withAuth(app.profileHandler.Activity))
		ar.Get("/profile/rooms", app.withAuth(app.profileHandler.Rooms))
		ar.Get("/profile/followers", app.withAuth(app.profileHandler.Followers))
		ar.Get("/profile/following", app.withAuth(app.profileHandler.Following))

		// フォロー関連API（認証必須）
		ar.Post("/users/{userID}/follow", app.withAuth(app.followHandler.FollowUser))
		ar.Delete("/users/{userID}/unfollow", app.withAuth(app.followHandler.UnfollowUser))
		ar.Get("/users/{userID}/follow-status", app.withAuth(app.followHandler.GetFollowStatus))

		// リアクション関連API（認証必須）
		ar.Post("/messages/{messageId}/reactions", app.withAuth(app.reactionHandler.AddReaction))
		ar.Delete("/messages/{messageId}/reactions/{reactionType}", app.withAuth(app.reactionHandler.RemoveReaction))

		// 通報関連API（認証必須）
		ar.Post("/users/{id}/report", app.withAuth(app.reportHandler.CreateReport))
		ar.Post("/reports/{id}/upload", app.withAuth(app.reportHandler.UploadAttachment))
		ar.Get("/report/reasons", app.withAuth(app.reportHandler.GetReportReasons))

		// 認証オプションのAPIエンドポイント
		ar.Get("/rooms", app.withOptionalAuth(app.roomHandler.GetAllRoomsAPI))
		ar.Get("/messages/{messageId}/reactions", app.withOptionalAuth(app.reactionHandler.GetMessageReactions))
		ar.Get("/reactions/types", app.withOptionalAuth(app.reactionHandler.GetAvailableReactions))
	})
}

func (app *Application) setupStaticRoutes(r chi.Router) {
	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// ローカル環境のOGP画像配信（OG_BUCKETが空の場合のみ）
	if os.Getenv("OG_BUCKET") == "" {
		tmpFileServer := http.FileServer(http.Dir("tmp"))
		r.Handle("/tmp/*", http.StripPrefix("/tmp/", tmpFileServer))
	}
}

func (app *Application) setupSSEOnlyRoutes(r chi.Router) {
	// SSE専用モード：メッセージストリーミング機能のみ提供
	r.Route("/rooms", func(rr chi.Router) {
		rmh := app.roomMessageHandler

		// SSE関連のルート（認証必須）
		if app.hasAuthMiddleware() {
			rr.Group(func(protected chi.Router) {
				protected.Use(app.authMiddleware.Middleware)

				// SSEトークン生成とメッセージストリーミング
				protected.Post("/{id}/sse-token", app.sseTokenHandler.GenerateSSEToken)
				protected.Get("/{id}/messages/stream", rmh.StreamMessages)
			})
		} else {
			// 開発環境での認証なしアクセス
			rr.Post("/{id}/sse-token", app.sseTokenHandler.GenerateSSEToken)
			rr.Get("/{id}/messages/stream", rmh.StreamMessages)
		}
	})

	// ヘルスチェック
	r.Get("/health", app.healthCheck)

	// 最小限のAPI
	r.Route("/api", func(ar chi.Router) {
		ar.Get("/health", app.healthCheck)
	})
}

func (app *Application) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"monhub"}`))
}
