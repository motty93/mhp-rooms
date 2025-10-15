package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mhp-rooms/internal/models"
	"mhp-rooms/internal/palette"
)

const (
	// OGP画像サイズ
	OGPWidth  = 1200
	OGPHeight = 630

	// レイアウト設定
	Padding       = 60.0
	LogoIconSize  = 48.0   // MonHubアイコンサイズ
	LogoY         = 60.0   // ロゴY位置
	TitleY        = 315.0  // 中央（630/2 = 315）
	SubtitleY     = 380.0  // タイトル下
	GameVersionX  = 1080.0 // 右下X位置（1200 - 60 - 60 = 1080）
	GameVersionY  = 540.0  // 右下Y位置（630 - 90 = 540）
	MaxTitleLines = 2

	// フォント設定
	TitleFontSize       = 72.0
	SubtitleFontSize    = 32.0
	LogoFontSize        = 36.0
	GameVersionFontSize = 48.0
	FontPathBold        = "./assets/fonts/NotoSansJP-Bold.ttf"
	FontPathRegular     = "./assets/fonts/NotoSansJP-Regular.ttf"

	// アセット設定
	IconImagePath = "./assets/images/icon.webp"
)

func main() {
	startTime := time.Now()

	// 環境変数の取得
	roomIDStr := os.Getenv("ROOM_ID")
	ogBucket := os.Getenv("OG_BUCKET")
	ogPrefix := os.Getenv("OG_PREFIX")
	databaseURL := os.Getenv("DATABASE_URL")

	if roomIDStr == "" || databaseURL == "" {
		log.Fatal("必須の環境変数が設定されていません: ROOM_ID, DATABASE_URL")
	}

	if ogPrefix == "" {
		ogPrefix = "dev" // デフォルト
	}

	// ローカルモード判定（OG_BUCKETが空の場合）
	isLocalMode := ogBucket == ""
	if isLocalMode {
		log.Printf("ローカルモード: tmp/images/og/ に保存します")
	}

	// RoomIDのパース
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		log.Fatalf("無効なROOM_ID: %v", err)
	}

	log.Printf("OGP画像生成開始: room_id=%s, bucket=%s, prefix=%s", roomID, ogBucket, ogPrefix)

	// データベース接続
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("データベース接続失敗: %v", err)
	}

	// 部屋情報の取得
	var room models.Room
	if err := db.Preload("GameVersion").Preload("Host").First(&room, roomID).Error; err != nil {
		log.Fatalf("部屋情報の取得失敗: %v", err)
	}

	log.Printf("部屋情報取得完了: name=%s, game_version=%s", room.Name, room.GameVersion.Code)

	// 配色の決定
	pal := palette.GetPalette(room.GameVersion.Code)
	log.Printf("配色決定: game_version=%s", room.GameVersion.Code)

	// OGP画像の生成
	img, err := generateOGPImage(&room, pal)
	if err != nil {
		log.Fatalf("OGP画像生成失敗: %v", err)
	}

	log.Printf("OGP画像生成完了")

	// 保存先の決定とアップロード
	ctx := context.Background()
	if isLocalMode {
		// ローカルファイルシステムに保存
		if err := saveToLocal(img, ogPrefix, roomID); err != nil {
			log.Fatalf("ローカル保存失敗: %v", err)
		}
	} else {
		// GCSへのアップロード
		if err := uploadToGCS(ctx, img, ogBucket, ogPrefix, roomID); err != nil {
			log.Fatalf("GCSアップロード失敗: %v", err)
		}
	}

	duration := time.Since(startTime).Milliseconds()
	log.Printf("OGP画像保存完了: duration_ms=%d", duration)
}

// saveToLocal ローカルファイルシステムに画像を保存
func saveToLocal(img image.Image, ogPrefix string, roomID uuid.UUID) error {
	// パス: tmp/images/og/{env}/rooms/{id}.png
	dirPath := filepath.Join("tmp", "images", "og", ogPrefix, "rooms")
	filePath := filepath.Join(dirPath, fmt.Sprintf("%s.png", roomID))

	// ディレクトリを作成
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("ディレクトリ作成失敗: %w", err)
	}

	// ファイルを作成
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ファイル作成失敗: %w", err)
	}
	defer file.Close()

	// PNG画像をエンコード
	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("画像エンコード失敗: %w", err)
	}

	log.Printf("ローカル保存完了: path=%s", filePath)
	return nil
}

// uploadToGCS GCSに画像をアップロード
func uploadToGCS(ctx context.Context, img image.Image, ogBucket, ogPrefix string, roomID uuid.UUID) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("GCSクライアント作成失敗: %w", err)
	}
	defer client.Close()

	// オブジェクトパス: og/{env}/rooms/{id}.png
	objectPath := fmt.Sprintf("og/%s/rooms/%s.png", ogPrefix, roomID)
	bucket := client.Bucket(ogBucket)
	obj := bucket.Object(objectPath)

	// アップロード
	w := obj.NewWriter(ctx)
	w.ContentType = "image/png"
	w.CacheControl = "public, max-age=31536000, immutable"

	if err := png.Encode(w, img); err != nil {
		return fmt.Errorf("画像エンコード失敗: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("アップロード失敗: %w", err)
	}

	log.Printf("GCSアップロード完了: path=%s", objectPath)
	return nil
}

// generateOGPImage OGP画像を生成
func generateOGPImage(room *models.Room, pal palette.GameVersionPalette) (image.Image, error) {
	dc := gg.NewContext(OGPWidth, OGPHeight)

	// 背景グラデーション（左上から右下）
	drawGradientBackground(dc, pal)

	// 左上: MonHubロゴ
	if err := drawMonHubLogo(dc); err != nil {
		return nil, fmt.Errorf("ロゴ描画失敗: %w", err)
	}

	// 中央: 部屋名
	if err := drawTitle(dc, room.Name); err != nil {
		return nil, fmt.Errorf("タイトル描画失敗: %w", err)
	}

	// 中央サブ: サブタイトル
	if err := drawSubtitle(dc); err != nil {
		return nil, fmt.Errorf("サブタイトル描画失敗: %w", err)
	}

	// 右下: ゲームバージョン
	if err := drawGameVersion(dc, room.GameVersion.Code, room.GameVersion.Name); err != nil {
		return nil, fmt.Errorf("ゲームバージョン描画失敗: %w", err)
	}

	return dc.Image(), nil
}

// drawGradientBackground グラデーション背景を描画（左上から右下）
func drawGradientBackground(dc *gg.Context, pal palette.GameVersionPalette) {
	// 左上から右下へのグラデーション（対角線）
	gradient := gg.NewLinearGradient(0, 0, OGPWidth, OGPHeight)
	gradient.AddColorStop(0, pal.TopColor)
	gradient.AddColorStop(1, pal.BottomColor)
	dc.SetFillStyle(gradient)
	dc.DrawRectangle(0, 0, OGPWidth, OGPHeight)
	dc.Fill()
}

// drawMonHubLogo MonHubロゴを左上に描画
func drawMonHubLogo(dc *gg.Context) error {
	x := Padding
	y := LogoY

	// アイコン画像を読み込み
	iconImg, err := gg.LoadImage(IconImagePath)
	if err != nil {
		log.Printf("アイコン画像の読み込みに失敗、代替表示を使用: %v", err)
		// 代替として円形アイコンを描画
		dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
		dc.DrawCircle(x+LogoIconSize/2, y+LogoIconSize/2, LogoIconSize/2)
		dc.Fill()
	} else {
		// アイコン画像を描画（リサイズして配置）
		dc.DrawImageAnchored(iconImg, int(x+LogoIconSize/2), int(y+LogoIconSize/2), 0.5, 0.5)
	}

	// MonHubテキストを右に配置
	if err := dc.LoadFontFace(FontPathBold, LogoFontSize); err != nil {
		return fmt.Errorf("フォント読み込み失敗: %w", err)
	}
	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	_, logoHeight := dc.MeasureString("MonHub")
	dc.DrawString("MonHub", x+LogoIconSize+20, y+LogoIconSize/2+logoHeight/2-8)

	return nil
}

// drawTitle タイトル（部屋名）を中央に描画
func drawTitle(dc *gg.Context, title string) error {
	if err := dc.LoadFontFace(FontPathBold, TitleFontSize); err != nil {
		return fmt.Errorf("フォント読み込み失敗: %w", err)
	}

	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// タイトルを中央揃えで描画
	maxWidth := OGPWidth - Padding*2
	lines := wrapText(dc, title, maxWidth, MaxTitleLines)

	// 複数行の場合は上にずらす
	startY := TitleY
	if len(lines) > 1 {
		startY -= TitleFontSize / 2
	}

	y := startY
	for _, line := range lines {
		textWidth, _ := dc.MeasureString(line)
		x := (OGPWidth - textWidth) / 2 // 中央揃え
		dc.DrawString(line, x, y)
		y += TitleFontSize + 10
	}

	return nil
}

// drawSubtitle サブタイトルを中央に描画
func drawSubtitle(dc *gg.Context) error {
	if err := dc.LoadFontFace(FontPathRegular, SubtitleFontSize); err != nil {
		return fmt.Errorf("フォント読み込み失敗: %w", err)
	}

	subtitle := "モンハンパーティ募集"
	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 200})

	textWidth, _ := dc.MeasureString(subtitle)
	x := (OGPWidth - textWidth) / 2 // 中央揃え
	dc.DrawString(subtitle, x, SubtitleY)

	return nil
}

// drawGameVersion ゲームバージョンを右下に描画
func drawGameVersion(dc *gg.Context, gameCode, gameName string) error {
	if err := dc.LoadFontFace(FontPathBold, GameVersionFontSize); err != nil {
		return fmt.Errorf("フォント読み込み失敗: %w", err)
	}

	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// ゲームバージョンコードを右下に配置
	textWidth, textHeight := dc.MeasureString(gameCode)
	x := OGPWidth - Padding - textWidth
	y := OGPHeight - Padding - textHeight + textHeight // ベースライン調整

	dc.DrawString(gameCode, x, y)

	return nil
}

// wrapText テキストを指定幅で折り返し
func wrapText(dc *gg.Context, text string, maxWidth float64, maxLines int) []string {
	var lines []string
	words := []rune(text)

	var currentLine []rune
	for _, r := range words {
		testLine := append(currentLine, r)
		w, _ := dc.MeasureString(string(testLine))

		if w > maxWidth {
			if len(currentLine) > 0 {
				lines = append(lines, string(currentLine))
				currentLine = []rune{r}
			} else {
				// 1文字でも幅を超える場合はそのまま追加
				lines = append(lines, string(r))
				currentLine = []rune{}
			}

			if len(lines) >= maxLines {
				break
			}
		} else {
			currentLine = testLine
		}
	}

	if len(currentLine) > 0 && len(lines) < maxLines {
		lines = append(lines, string(currentLine))
	}

	return lines
}
