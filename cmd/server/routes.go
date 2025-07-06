package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *Application) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	app.setupPageRoutes(r)
	app.setupRoomRoutes(r)
	app.setupAuthRoutes(r)
	app.setupAPIRoutes(r)
	app.setupStaticRoutes(r)

	return r
}

func (app *Application) setupPageRoutes(r *mux.Router) {
	pr := r.PathPrefix("").Subrouter()
	ph := app.pageHandler

	pr.HandleFunc("/", ph.Home).Methods("GET")
	pr.HandleFunc("/terms", ph.Terms).Methods("GET")
	pr.HandleFunc("/privacy", ph.Privacy).Methods("GET")
	pr.HandleFunc("/contact", ph.Contact).Methods("GET", "POST")
	pr.HandleFunc("/faq", ph.FAQ).Methods("GET")
	pr.HandleFunc("/guide", ph.Guide).Methods("GET")
	pr.HandleFunc("/hello", ph.Hello).Methods("GET")
	pr.HandleFunc("/sitemap.xml", ph.Sitemap).Methods("GET")
}

func (app *Application) setupRoomRoutes(r *mux.Router) {
	rr := r.PathPrefix("/rooms").Subrouter()
	rh := app.roomHandler

	rr.HandleFunc("", rh.Rooms).Methods("GET")
	rr.HandleFunc("", rh.CreateRoom).Methods("POST")
	rr.HandleFunc("/{id}/join", rh.JoinRoom).Methods("POST")
	rr.HandleFunc("/{id}/leave", rh.LeaveRoom).Methods("POST")
	rr.HandleFunc("/{id}/toggle-closed", rh.ToggleRoomClosed).Methods("PUT")
}

func (app *Application) setupAuthRoutes(r *mux.Router) {
	ar := r.PathPrefix("/auth").Subrouter()
	ah := app.authHandler

	ar.HandleFunc("/login", ah.LoginPage).Methods("GET")
	ar.HandleFunc("/login", ah.Login).Methods("POST")
	ar.HandleFunc("/register", ah.RegisterPage).Methods("GET")
	ar.HandleFunc("/register", ah.Register).Methods("POST")
	ar.HandleFunc("/logout", ah.Logout).Methods("POST")

	ar.HandleFunc("/password-reset", ah.PasswordResetPage).Methods("GET")
	ar.HandleFunc("/password-reset", ah.PasswordResetRequest).Methods("POST")
	ar.HandleFunc("/password-reset/confirm", ah.PasswordResetConfirmPage).Methods("GET")
	ar.HandleFunc("/password-reset/confirm", ah.PasswordResetConfirm).Methods("POST")

	ar.HandleFunc("/callback", ah.AuthCallback).Methods("GET")
	ar.HandleFunc("/google", ah.GoogleAuth).Methods("GET")
	ar.HandleFunc("/google/callback", ah.GoogleCallback).Methods("GET")

	ar.HandleFunc("/complete-profile", ah.CompleteProfilePage).Methods("GET")
	ar.HandleFunc("/complete-profile", ah.CompleteProfile).Methods("POST")
}

func (app *Application) setupAPIRoutes(r *mux.Router) {
	apiRoutes := r.PathPrefix("/api").Subrouter()

	apiRoutes.HandleFunc("/config/supabase", app.configHandler.GetSupabaseConfig).Methods("GET")
	apiRoutes.HandleFunc("/health", app.healthCheck).Methods("GET")
	
	if app.authMiddleware != nil {
		protected := apiRoutes.PathPrefix("").Subrouter()
		protected.Use(app.authMiddleware.Middleware)
		
		protected.HandleFunc("/user/current", app.authHandler.CurrentUser).Methods("GET")
		protected.HandleFunc("/auth/sync", app.authHandler.SyncUser).Methods("POST")
		protected.HandleFunc("/auth/psn-id", app.authHandler.UpdatePSNId).Methods("PUT")
		
		optional := apiRoutes.PathPrefix("").Subrouter()
		optional.Use(app.authMiddleware.OptionalMiddleware)
		
		optional.HandleFunc("/rooms", app.roomHandler.GetAllRoomsAPI).Methods("GET")
	} else {
		apiRoutes.HandleFunc("/user/current", app.authHandler.CurrentUser).Methods("GET")
		apiRoutes.HandleFunc("/auth/sync", app.authHandler.SyncUser).Methods("POST")
		apiRoutes.HandleFunc("/auth/psn-id", app.authHandler.UpdatePSNId).Methods("PUT")
		apiRoutes.HandleFunc("/rooms", app.roomHandler.GetAllRoomsAPI).Methods("GET")
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
