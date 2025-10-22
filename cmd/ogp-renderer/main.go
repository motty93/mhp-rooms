package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/nfnt/resize"
	"golang.org/x/image/font/opentype"
	_ "golang.org/x/image/webp"

	"golang.org/x/image/font"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/infrastructure/persistence"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/view"
)

const (
	// OGP画像サイズ（最終出力サイズ）
	OGPWidth  = 1200
	OGPHeight = 630

	// 内部レンダリング倍率（高解像度で描画→等倍に縮小）
	RenderScale = 2 // 2～3がおすすめ

	// レイアウト設定（Zenn風）
	Padding        = 50.0
	BorderWidth    = 16.0 // 枠の太さ（さらに太く）
	BorderRadius   = 20.0 // 枠の角丸
	ContentPadding = 40.0 // 枠内の余白
	LogoIconSize   = 95.0 // MonHubアイコンサイズ
	MaxTitleLines  = 3    // タイトル最大行数

	// フォント設定
	TitleFontSize       = 64.0 // タイトル
	LogoFontSize        = 36.0 // MonHub
	GameVersionFontSize = 36.0 // ゲームバージョン
	FontPath            = "cmd/ogp-renderer/assets/fonts/NotoSansCJKjp-Bold.otf"

	// アセット設定
	IconImagePath = "cmd/ogp-renderer/assets/images/icon.webp"
	HeroImagePath = "static/images/hero.webp"
)

func main() {
	startTime := time.Now()

	// .envファイルのロード
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数を使用します。")
	}

	// Configの初期化
	config.Init()
	cfg := config.AppConfig

	// 環境変数の取得
	roomIDStr := os.Getenv("ROOM_ID")
	ogBucket := os.Getenv("OG_BUCKET")
	ogPrefix := os.Getenv("OG_PREFIX")

	if roomIDStr == "" {
		log.Fatal("必須の環境変数が設定されていません: ROOM_ID")
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
	dbAdapter, err := persistence.NewDBAdapter(cfg)
	if err != nil {
		log.Fatalf("データベース接続失敗: %v", err)
	}

	// 部屋情報の取得
	var room models.Room
	if err := dbAdapter.GetConn().Preload("GameVersion").Preload("Host").First(&room, roomID).Error; err != nil {
		log.Fatalf("部屋情報の取得失敗: %v", err)
	}

	log.Printf("部屋情報取得完了: name=%s, game_version=%s", room.Name, room.GameVersion.Code)

	// 配色の決定
	pal := view.GetPalette(room.GameVersion.Code)
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

// ------------------------------
// フォント: HintingNone + truetype
// ------------------------------
func mustLoadFaceTTF(path string, size float64) font.Face {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("フォント読み込み失敗: %v", err)
	}
	f, err := opentype.Parse(b)
	if err != nil {
		log.Fatalf("フォント解析失敗: %v", err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatalf("フォントフェイス作成失敗: %v", err)
	}

	return face
}

// saveToLocal ローカルファイルシステムに画像を保存
func saveToLocal(img image.Image, ogPrefix string, roomID uuid.UUID) error {
	// パス: tmp/images/{env}/ogp/rooms/{id}.png
	dirPath := filepath.Join("tmp", "images", ogPrefix, "ogp", "rooms")
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

	// オブジェクトパス: %s/ogp/rooms/%s.png
	objectPath := fmt.Sprintf("%s/ogp/rooms/%s.png", ogPrefix, roomID)
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

// generateOGPImage OGP画像を生成（Zenn風デザイン）
// 内部では RenderScale 倍のキャンバスに描画し、最後に等倍へ縮小します。
func generateOGPImage(room *models.Room, pal view.GameVersionPalette) (image.Image, error) {
	scale := float64(RenderScale)
	W := int(float64(OGPWidth) * scale)
	H := int(float64(OGPHeight) * scale)

	dc := gg.NewContext(W, H)

	// 背景: 白
	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	dc.Clear()

	// グラデーション枠を描画
	if err := drawGradientBorder(dc, pal, scale); err != nil {
		return nil, fmt.Errorf("枠描画失敗: %w", err)
	}

	// 左上: 部屋名
	if err := drawTitleTopLeft(dc, room.Name, scale); err != nil {
		return nil, fmt.Errorf("タイトル描画失敗: %w", err)
	}

	// 左下: ゲームバージョン
	if err := drawGameVersionBottomLeft(dc, room.GameVersion.Code, scale); err != nil {
		return nil, fmt.Errorf("ゲームバージョン描画失敗: %w", err)
	}

	// 右下: MonHubロゴ
	if err := drawMonHubLogoBottomRight(dc, scale); err != nil {
		return nil, fmt.Errorf("ロゴ描画失敗: %w", err)
	}

	// 高解像度→等倍へ縮小（Lanczos3）
	hi := dc.Image()
	lo := resize.Resize(uint(OGPWidth), uint(OGPHeight), hi, resize.Lanczos3)
	return lo, nil
}

// drawGradientBorder グラデーション枠を描画（Zenn風）
func drawGradientBorder(dc *gg.Context, pal view.GameVersionPalette, s float64) error {
	p := Padding * s
	bw := BorderWidth * s
	br := BorderRadius * s

	// 左上から右下へのグラデーション
	gradient := gg.NewLinearGradient(0, 0, float64(dc.Width()), float64(dc.Height()))
	gradient.AddColorStop(0, pal.TopColor)
	gradient.AddColorStop(1, pal.BottomColor)

	// 外側の枠を描画（角丸）
	dc.SetFillStyle(gradient)
	dc.DrawRoundedRectangle(p, p, float64(dc.Width())-p*2, float64(dc.Height())-p*2, br)
	dc.Fill()

	// 内側を白で塗りつぶし（枠だけ残す・角丸）
	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	dc.DrawRoundedRectangle(
		p+bw,
		p+bw,
		float64(dc.Width())-p*2-bw*2,
		float64(dc.Height())-p*2-bw*2,
		br,
	)
	dc.Fill()

	return nil
}

// drawTitleTopLeft 部屋名を左上に描画
func drawTitleTopLeft(dc *gg.Context, title string, s float64) error {
	face := mustLoadFaceTTF(FontPath, TitleFontSize*s)
	dc.SetFontFace(face)

	// テキストを折り返し
	maxWidth := float64(dc.Width()) - (Padding+BorderWidth+ContentPadding)*2*s - 100*s
	lines := wrapText(dc, title, maxWidth, MaxTitleLines)

	x := (Padding + BorderWidth + ContentPadding) * s
	y := (Padding+BorderWidth+ContentPadding)*s + TitleFontSize*s

	// 黒色で描画
	dc.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	lineHeight := TitleFontSize*s + 10*s
	for _, line := range lines {
		dc.DrawString(line, x, y)
		y += lineHeight
	}
	return nil
}

// drawGameVersionBottomLeft ゲームバージョンを左下に描画
func drawGameVersionBottomLeft(dc *gg.Context, gameCode string, s float64) error {
	face := mustLoadFaceTTF(FontPath, GameVersionFontSize*s)
	dc.SetFontFace(face)

	x := (Padding + BorderWidth + ContentPadding) * s
	y := float64(dc.Height()) - (Padding+BorderWidth+ContentPadding)*s

	// 黒色で描画
	dc.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	dc.DrawString(gameCode, x, y)
	return nil
}

// drawMonHubLogoBottomRight MonHubロゴを右下に描画
func drawMonHubLogoBottomRight(dc *gg.Context, s float64) error {
	// アイコン画像を読み込み
	iconImg, err := gg.LoadImage(IconImagePath)
	if err != nil {
		log.Printf("アイコン画像の読み込みに失敗: %v", err)
		return nil
	}

	// アイコン画像をリサイズ
	iconSize := uint(LogoIconSize * s)
	resizedIcon := resize.Resize(iconSize, iconSize, iconImg, resize.Lanczos3)

	// フォント設定
	dc.SetFontFace(mustLoadFaceTTF(FontPath, LogoFontSize*s))

	// テキスト幅を取得
	text := "MonHub"
	textWidth, _ := dc.MeasureString(text)

	// game_versionと同じベースラインに揃える
	baselineY := float64(dc.Height()) - (Padding+BorderWidth+ContentPadding)*s

	// テキストのベースラインがbaselineYになるように配置
	textY := baselineY

	// アイコンの底辺をベースラインに揃える
	// アイコンの上端Y座標 = baselineY（底辺） - iconSize（高さ）
	// しかし、アイコンを少し下げてテキストと視覚的に揃える
	iconY := baselineY - float64(iconSize)*0.65

	// 右端から配置
	totalWidth := float64(iconSize) + textWidth
	baseX := float64(dc.Width()) - (Padding+BorderWidth+ContentPadding)*s - totalWidth

	// アイコンを描画
	dc.DrawImage(resizedIcon, int(baseX), int(iconY))

	// MonHubテキストを描画
	textX := baseX + float64(iconSize)
	dc.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	dc.DrawString(text, textX, textY)

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
