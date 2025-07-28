package main

import (
	"net/http"

	"mhp-rooms/internal/middleware"

	"github.com/gorilla/mux"
)

func (app *Application) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

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

func (app *Application) setupPageRoutes(r *mux.Router) {
	ph := app.pageHandler

	// 各ページルートに個別にミドルウェアを適用
	if app.authMiddleware != nil {
		r.HandleFunc("/", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.Home)).ServeHTTP).Methods("GET")
		r.HandleFunc("/terms", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.Terms)).ServeHTTP).Methods("GET")
		r.HandleFunc("/privacy", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.Privacy)).ServeHTTP).Methods("GET")
		r.HandleFunc("/contact", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.Contact)).ServeHTTP).Methods("GET", "POST")
		r.HandleFunc("/faq", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.FAQ)).ServeHTTP).Methods("GET")
		r.HandleFunc("/guide", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.Guide)).ServeHTTP).Methods("GET")
		r.HandleFunc("/hello", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.Hello)).ServeHTTP).Methods("GET")
		r.HandleFunc("/sitemap.xml", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(ph.Sitemap)).ServeHTTP).Methods("GET")
	} else {
		r.HandleFunc("/", ph.Home).Methods("GET")
		r.HandleFunc("/terms", ph.Terms).Methods("GET")
		r.HandleFunc("/privacy", ph.Privacy).Methods("GET")
		r.HandleFunc("/contact", ph.Contact).Methods("GET", "POST")
		r.HandleFunc("/faq", ph.FAQ).Methods("GET")
		r.HandleFunc("/guide", ph.Guide).Methods("GET")
		r.HandleFunc("/hello", ph.Hello).Methods("GET")
		r.HandleFunc("/sitemap.xml", ph.Sitemap).Methods("GET")
	}
}

func (app *Application) setupRoomRoutes(r *mux.Router) {
	rr := r.PathPrefix("/rooms").Subrouter()
	rh := app.roomHandler
	rdh := app.roomDetailHandler
	rmh := app.roomMessageHandler

	// 認証不要なルート
	rr.HandleFunc("", rh.Rooms).Methods("GET")
	rr.HandleFunc("/{id}", rdh.RoomDetail).Methods("GET")

	// 認証が必要なルート
	if app.authMiddleware != nil {
		protected := rr.PathPrefix("").Subrouter()
		protected.Use(app.authMiddleware.Middleware)

		protected.HandleFunc("", rh.CreateRoom).Methods("POST")
		protected.HandleFunc("/{id}/join", rh.JoinRoom).Methods("POST")
		protected.HandleFunc("/{id}/leave", rh.LeaveRoom).Methods("POST")
		protected.HandleFunc("/{id}/toggle-closed", rh.ToggleRoomClosed).Methods("PUT")

		// メッセージ関連
		protected.HandleFunc("/{id}/messages", rmh.SendMessage).Methods("POST")
		protected.HandleFunc("/{id}/messages", rmh.GetMessages).Methods("GET")
		protected.HandleFunc("/{id}/sse-token", app.sseTokenHandler.GenerateSSEToken).Methods("POST")

		// SSEストリーム（一時トークン認証）
		rr.HandleFunc("/{id}/messages/stream", rmh.StreamMessages).Methods("GET")
	} else {
		// 認証ミドルウェアがない場合（開発環境など）
		rr.HandleFunc("", rh.CreateRoom).Methods("POST")
		rr.HandleFunc("/{id}/join", rh.JoinRoom).Methods("POST")
		rr.HandleFunc("/{id}/leave", rh.LeaveRoom).Methods("POST")
		rr.HandleFunc("/{id}/toggle-closed", rh.ToggleRoomClosed).Methods("PUT")

		// メッセージ関連
		rr.HandleFunc("/{id}/messages", rmh.SendMessage).Methods("POST")
		rr.HandleFunc("/{id}/messages", rmh.GetMessages).Methods("GET")
		rr.HandleFunc("/{id}/sse-token", app.sseTokenHandler.GenerateSSEToken).Methods("POST")
		rr.HandleFunc("/{id}/messages/stream", rmh.StreamMessages).Methods("GET")
	}
}

func (app *Application) setupAuthRoutes(r *mux.Router) {
	ar := r.PathPrefix("/auth").Subrouter()
	ah := app.authHandler

	// 認証ページルート（レート制限は緩め）
	ar.HandleFunc("/login", ah.LoginPage).Methods("GET")
	ar.HandleFunc("/register", ah.RegisterPage).Methods("GET")
	ar.HandleFunc("/password-reset", ah.PasswordResetPage).Methods("GET")
	ar.HandleFunc("/password-reset/confirm", ah.PasswordResetConfirmPage).Methods("GET")
	ar.HandleFunc("/callback", ah.AuthCallback).Methods("GET")
	ar.HandleFunc("/complete-profile", ah.CompleteProfilePage).Methods("GET")

	// 認証アクションルート（厳しいレート制限）
	authActionRoutes := ar.PathPrefix("").Subrouter()
	authActionRoutes.Use(middleware.AuthRateLimitMiddleware(app.authLimiter))

	authActionRoutes.HandleFunc("/login", ah.Login).Methods("POST")
	authActionRoutes.HandleFunc("/register", ah.Register).Methods("POST")
	authActionRoutes.HandleFunc("/logout", ah.Logout).Methods("POST")
	authActionRoutes.HandleFunc("/password-reset", ah.PasswordResetRequest).Methods("POST")
	authActionRoutes.HandleFunc("/password-reset/confirm", ah.PasswordResetConfirm).Methods("POST")
	authActionRoutes.HandleFunc("/google", ah.GoogleAuth).Methods("GET")
	authActionRoutes.HandleFunc("/google/callback", ah.GoogleCallback).Methods("GET")
	authActionRoutes.HandleFunc("/complete-profile", ah.CompleteProfile).Methods("POST")
}

func (app *Application) setupAPIRoutes(r *mux.Router) {
	apiRoutes := r.PathPrefix("/api").Subrouter()

	// 認証不要なAPIエンドポイント
	apiRoutes.HandleFunc("/config/supabase", app.configHandler.GetSupabaseConfig).Methods("GET")
	apiRoutes.HandleFunc("/health", app.healthCheck).Methods("GET")

	if app.authMiddleware != nil {
		// 認証関連API（厳しいレート制限 + 認証必須）
		authAPIRoutes := apiRoutes.PathPrefix("/auth").Subrouter()
		authAPIRoutes.Use(middleware.AuthRateLimitMiddleware(app.authLimiter))
		authAPIRoutes.Use(app.authMiddleware.Middleware)

		authAPIRoutes.HandleFunc("/sync", app.authHandler.SyncUser).Methods("POST")
		authAPIRoutes.HandleFunc("/psn-id", app.authHandler.UpdatePSNId).Methods("PUT")

		// 認証必須のAPIエンドポイント
		apiRoutes.HandleFunc("/user/current", app.authMiddleware.Middleware(http.HandlerFunc(app.authHandler.CurrentUser)).ServeHTTP).Methods("GET")

		// リアクション関連API（認証必須）
		apiRoutes.HandleFunc("/messages/{messageId}/reactions", app.authMiddleware.Middleware(http.HandlerFunc(app.reactionHandler.AddReaction)).ServeHTTP).Methods("POST")
		apiRoutes.HandleFunc("/messages/{messageId}/reactions/{reactionType}", app.authMiddleware.Middleware(http.HandlerFunc(app.reactionHandler.RemoveReaction)).ServeHTTP).Methods("DELETE")

		// 認証オプションのAPIエンドポイント
		apiRoutes.HandleFunc("/rooms", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(app.roomHandler.GetAllRoomsAPI)).ServeHTTP).Methods("GET")
		apiRoutes.HandleFunc("/messages/{messageId}/reactions", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(app.reactionHandler.GetMessageReactions)).ServeHTTP).Methods("GET")
		apiRoutes.HandleFunc("/reactions/types", app.authMiddleware.OptionalMiddleware(http.HandlerFunc(app.reactionHandler.GetAvailableReactions)).ServeHTTP).Methods("GET")
	} else {
		// 認証ミドルウェアがない場合（開発環境）
		apiRoutes.HandleFunc("/user/current", app.authHandler.CurrentUser).Methods("GET")
		apiRoutes.HandleFunc("/auth/sync", app.authHandler.SyncUser).Methods("POST")
		apiRoutes.HandleFunc("/auth/psn-id", app.authHandler.UpdatePSNId).Methods("PUT")
		apiRoutes.HandleFunc("/rooms", app.roomHandler.GetAllRoomsAPI).Methods("GET")

		// リアクション関連API（認証なし環境）
		apiRoutes.HandleFunc("/messages/{messageId}/reactions", app.reactionHandler.AddReaction).Methods("POST")
		apiRoutes.HandleFunc("/messages/{messageId}/reactions/{reactionType}", app.reactionHandler.RemoveReaction).Methods("DELETE")
		apiRoutes.HandleFunc("/messages/{messageId}/reactions", app.reactionHandler.GetMessageReactions).Methods("GET")
		apiRoutes.HandleFunc("/reactions/types", app.reactionHandler.GetAvailableReactions).Methods("GET")
	}
}

func (app *Application) setupStaticRoutes(r *mux.Router) {
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))),
	)
}

func (app *Application) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"monhub"}`))
}
