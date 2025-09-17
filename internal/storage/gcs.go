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
	"path"
	"regexp"
	"strings"

	"mhp-rooms/internal/config"

	"cloud.google.com/go/storage"
)

// GCSUploader Google Cloud Storageアップローダー
type GCSUploader struct {
	client *storage.Client
	config *config.GCSConfig
}

// NewGCSUploader 新しいGCSアップローダーを作成
func NewGCSUploader(ctx context.Context) (*GCSUploader, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCSクライアントの初期化に失敗しました: %w", err)
	}

	return &GCSUploader{
		client: client,
		config: &config.AppConfig.GCS,
	}, nil
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
