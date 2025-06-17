package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"

	"mhp-rooms/internal/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// MIMEタイプを設定
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".png", "image/png")
	mime.AddExtensionType(".jpg", "image/jpeg")
	mime.AddExtensionType(".ico", "image/x-icon")

	r := mux.NewRouter()

	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/hello", handlers.HelloHandler).Methods("GET")
	
	// ヘルスチェックエンドポイント（デバッグ用）
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	// ファイル存在確認用デバッグエンドポイント
	r.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		// 現在の作業ディレクトリを表示
		if wd, err := os.Getwd(); err == nil {
			w.Write([]byte("Working directory: " + wd + "\n"))
		}
		
		// staticディレクトリの確認
		if _, err := os.Stat("static"); os.IsNotExist(err) {
			w.Write([]byte("static directory NOT found\n"))
		} else {
			w.Write([]byte("static directory found\n"))
			
			// staticディレクトリの中身をリスト
			if files, err := os.ReadDir("static"); err == nil {
				w.Write([]byte("static directory contents:\n"))
				for _, file := range files {
					w.Write([]byte("  " + file.Name() + "\n"))
				}
			}
		}
		
		// static/cssディレクトリの確認
		if _, err := os.Stat("static/css"); os.IsNotExist(err) {
			w.Write([]byte("static/css directory NOT found\n"))
		} else {
			w.Write([]byte("static/css directory found\n"))
			
			// static/cssディレクトリの中身をリスト
			if files, err := os.ReadDir("static/css"); err == nil {
				w.Write([]byte("static/css directory contents:\n"))
				for _, file := range files {
					w.Write([]byte("  " + file.Name() + "\n"))
				}
			}
		}
		
		// style.cssファイルの確認
		if _, err := os.Stat("static/css/style.css"); os.IsNotExist(err) {
			w.Write([]byte("style.css NOT found\n"))
		} else {
			w.Write([]byte("style.css found\n"))
		}
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
