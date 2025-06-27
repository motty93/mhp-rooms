package supabase

import (
	"fmt"
	"log"
	"os"

	supa "github.com/supabase-community/supabase-go"
)

var Client *supa.Client

// Init initializes the Supabase client
func Init() error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}

	var err error
	Client, err = supa.NewClient(supabaseURL, supabaseKey, &supa.ClientOptions{})
	if err != nil {
		return fmt.Errorf("Supabaseクライアントの初期化に失敗しました: %w", err)
	}

	log.Println("Supabaseクライアントを初期化しました")
	return nil
}

// GetClient returns the initialized Supabase client
func GetClient() *supa.Client {
	if Client == nil {
		log.Fatal("Supabaseクライアントが初期化されていません")
	}
	return Client
}
