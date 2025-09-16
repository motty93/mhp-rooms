package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
)

// Config GCSの設定
type Config struct {
	Bucket         string
	BaseURL        string
	MaxUploadBytes int64
	AllowedMIMEs   map[string]struct{}
	AssetPrefix    string
}

// GCSUploader Google Cloud Storageアップローダー
type GCSUploader struct {
	client *storage.Client
	config *Config
}

// NewGCSUploader 新しいGCSアップローダーを作成
func NewGCSUploader(ctx context.Context) (*GCSUploader, error) {
	config := mustConfig()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCSクライアントの初期化に失敗しました: %w", err)
	}

	return &GCSUploader{
		client: client,
		config: config,
	}, nil
}

// mustConfig 必須の環境変数から設定を読み込む
func mustConfig() *Config {
	return &Config{
		Bucket:         mustGetenv("GCS_BUCKET"),
		BaseURL:        mustGetenv("BASE_PUBLIC_ASSET_URL"),
		MaxUploadBytes: envInt64("MAX_UPLOAD_BYTES", 10<<20), // デフォルト10MB
		AllowedMIMEs:   parseAllowed(os.Getenv("ALLOW_CONTENT_TYPES"), []string{"image/jpeg", "image/png", "image/webp"}),
		AssetPrefix:    cleanPrefix(os.Getenv("ASSET_PREFIX")),
	}
}

// mustGetenv 必須の環境変数を取得
func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("環境変数 %s が設定されていません", key))
	}
	return v
}

// envInt64 環境変数から数値を取得
func envInt64(key string, def int64) int64 {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("環境変数 %s が無効な数値です: %v", key, err))
	}
	return v
}

// parseAllowed 許可されたMIMEタイプをパース
func parseAllowed(env string, defaults []string) map[string]struct{} {
	list := defaults
	if env != "" {
		list = strings.Split(env, ",")
	}
	m := make(map[string]struct{})
	for _, s := range list {
		m[strings.TrimSpace(s)] = struct{}{}
	}
	return m
}

// cleanPrefix プレフィックスをクリーンアップ（dev/stg/prodなどのみ許可）
func cleanPrefix(s string) string {
	s = strings.TrimSpace(strings.Trim(s, "/"))
	if s == "" {
		return "dev" // デフォルトをdevに
	}
	re := regexp.MustCompile(`^[a-z0-9._-]+$`)
	if !re.MatchString(s) {
		panic(fmt.Sprintf("無効なASSET_PREFIX: %s", s))
	}
	return s
}

// UploadResult アップロード結果
type UploadResult struct {
	URL         string `json:"url"`
	ObjectPath  string `json:"object_path"`
	ContentType string `json:"content_type"`
}

// UploadAvatar アバター画像をアップロード
func (u *GCSUploader) UploadAvatar(ctx context.Context, userID string, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// ファイルサイズチェック
	if header.Size > u.config.MaxUploadBytes {
		return nil, fmt.Errorf("ファイルサイズが制限を超えています（最大 %d MB）", u.config.MaxUploadBytes/(1<<20))
	}

	// ファイルを読み込み
	buf := bytes.NewBuffer(nil)
	if _, err := io.CopyN(buf, file, u.config.MaxUploadBytes+1); err != nil && err != io.EOF {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}

	// Content-Type判定
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		bufBytes := buf.Bytes()
		detectSize := 512
		if len(bufBytes) < detectSize {
			detectSize = len(bufBytes)
		}
		contentType = http.DetectContentType(bufBytes[:detectSize])
	}

	// MIMEタイプチェック
	if _, ok := u.config.AllowedMIMEs[contentType]; !ok {
		return nil, errors.New("許可されていないファイル形式です")
	}

	// 拡張子取得
	ext := getExtension(header.Filename, contentType)

	// ハッシュ計算（重複防止）
	h := md5.New()
	h.Write(buf.Bytes())
	hash12 := hex.EncodeToString(h.Sum(nil))[:12]

	// ベースネーム取得
	base := baseNameSansExt(header.Filename)
	if base == "" {
		base = "avatar"
	}
	base = sanitizeName(base)

	// オブジェクトパス生成（環境別フォルダ付き）
	objectPath := path.Join(
		u.config.AssetPrefix, // dev/prod などを先頭に
		"avatars",
		userID,
		fmt.Sprintf("%s-%s%s", base, hash12, ext),
	)

	// GCSにアップロード
	bucket := u.client.Bucket(u.config.Bucket)
	obj := bucket.Object(objectPath)
	writer := obj.NewWriter(ctx)

	// メタデータ設定
	writer.CacheControl = "public, max-age=31536000, immutable"
	writer.ContentType = contentType

	// 書き込み
	if _, err := writer.Write(buf.Bytes()); err != nil {
		return nil, fmt.Errorf("GCS書き込みエラー: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("GCSクローズエラー: %w", err)
	}

	// 公開URL生成
	publicURL := u.PublicURL(objectPath)

	return &UploadResult{
		URL:         publicURL,
		ObjectPath:  objectPath,
		ContentType: contentType,
	}, nil
}

// PublicURL オブジェクトパスから公開URLを生成
func (u *GCSUploader) PublicURL(objectPath string) string {
	base := strings.TrimRight(u.config.BaseURL, "/")
	return base + "/" + strings.TrimLeft(objectPath, "/")
}

// Close クライアントをクローズ
func (u *GCSUploader) Close() error {
	return u.client.Close()
}

// getExtension ファイル名またはMIMEタイプから拡張子を取得
func getExtension(filename, mimeType string) string {
	// ファイル名から拡張子を取得
	if ext := path.Ext(filename); ext != "" {
		return strings.ToLower(ext)
	}

	// MIMEタイプから拡張子を推定
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg" // デフォルト
	}
}

// baseNameSansExt 拡張子を除いたベース名を取得
func baseNameSansExt(filename string) string {
	base := path.Base(filename)
	ext := path.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}

// sanitizeName ファイル名をサニタイズ
func sanitizeName(name string) string {
	// 英数字、ハイフン、アンダースコアのみ許可
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	sanitized := re.ReplaceAllString(name, "")

	// 長さ制限
	if len(sanitized) > 64 {
		sanitized = sanitized[:64]
	}

	// 空の場合はデフォルト
	if sanitized == "" {
		sanitized = "file"
	}

	return sanitized
}
