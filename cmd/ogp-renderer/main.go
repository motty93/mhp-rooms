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
	LogoIconSize   = 95.0 // HuntersHubアイコンサイズ
	MaxTitleLines  = 3    // タイトル最大行数

	// フォント設定
	TitleFontSize       = 64.0 // タイトル
	LogoFontSize        = 36.0 // HuntersHub
	GameVersionFontSize = 36.0 // ゲームバージョン

	// サイト用OGP（デフォルトカード）のレイアウト設定
	SiteIconSize        = 120.0 // 中央ロゴのアイコンサイズ
	SiteNameFontSize    = 84.0  // サービス名
	SiteTaglineFontSize = 38.0  // タグライン
	SiteBadgeFontSize   = 28.0  // ゲームバッジ
)

// siteGameCodes サイト用OGPに並べる対応タイトル（表示順）
var siteGameCodes = []string{"MHP", "MHP2", "MHP2G", "MHP3", "MHXX"}

func main() {
	startTime := time.Now()

	// .envファイルのロード
	if err := godotenv.Load(); err != nil {
		log.Println(".envファイルが見つかりません。環境変数を使用します。")
	}

	// 環境変数の取得
	roomIDStr := os.Getenv("ROOM_ID")
	ogBucket := os.Getenv("OG_BUCKET")
	ogPrefix := os.Getenv("OG_PREFIX")

	// アセットパスの取得（環境変数で上書き可能）
	fontPath := os.Getenv("FONT_PATH")
	if fontPath == "" {
		fontPath = "cmd/ogp-renderer/assets/fonts/NotoSansCJKjp-Bold.otf" // ローカル開発用デフォルト
	}

	iconImagePath := os.Getenv("ICON_IMAGE_PATH")
	if iconImagePath == "" {
		iconImagePath = "cmd/ogp-renderer/assets/images/icon.webp" // ローカル開発用デフォルト
	}

	if ogPrefix == "" {
		ogPrefix = "dev" // デフォルト
	}

	// ローカルモード判定（OG_BUCKETが空の場合）
	isLocalMode := ogBucket == ""
	if isLocalMode {
		log.Printf("ローカルモード: tmp/images/ に保存します")
	}

	// サイト用OGP（デフォルトカード）の生成モード。DB接続・ROOM_ID は不要
	if os.Getenv("OGP_TARGET") == "site" {
		log.Printf("サイト用OGP画像生成開始: bucket=%s, prefix=%s", ogBucket, ogPrefix)

		img, err := generateSiteOGPImage(fontPath, iconImagePath)
		if err != nil {
			log.Fatalf("サイト用OGP画像生成失敗: %v", err)
		}

		objectPath := "ogp/og_image.png"
		if isLocalMode {
			if err := saveToLocal(img, ogPrefix, objectPath); err != nil {
				log.Fatalf("ローカル保存失敗: %v", err)
			}
		} else {
			// 同一URLを上書きするため immutable にしない
			if err := uploadToGCS(context.Background(), img, ogBucket, ogPrefix, objectPath, "public, max-age=3600"); err != nil {
				log.Fatalf("GCSアップロード失敗: %v", err)
			}
		}

		log.Printf("サイト用OGP画像保存完了: duration_ms=%d", time.Since(startTime).Milliseconds())
		return
	}

	if roomIDStr == "" {
		log.Fatal("必須の環境変数が設定されていません: ROOM_ID")
	}

	// RoomIDのパース
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		log.Fatalf("無効なROOM_ID: %v", err)
	}

	log.Printf("OGP画像生成開始: room_id=%s, bucket=%s, prefix=%s", roomID, ogBucket, ogPrefix)

	// OGP Renderer用の最小限のConfig設定
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:           os.Getenv("DB_TYPE"),
			TursoURL:       os.Getenv("TURSO_DATABASE_URL"),
			TursoAuthToken: os.Getenv("TURSO_AUTH_TOKEN"),
		},
	}

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
	img, err := generateOGPImage(&room, pal, fontPath, iconImagePath)
	if err != nil {
		log.Fatalf("OGP画像生成失敗: %v", err)
	}

	log.Printf("OGP画像生成完了")

	// 保存先の決定とアップロード
	ctx := context.Background()
	objectPath := fmt.Sprintf("ogp/rooms/%s.png", roomID)
	if isLocalMode {
		// ローカルファイルシステムに保存
		if err := saveToLocal(img, ogPrefix, objectPath); err != nil {
			log.Fatalf("ローカル保存失敗: %v", err)
		}
	} else {
		// GCSへのアップロード
		if err := uploadToGCS(ctx, img, ogBucket, ogPrefix, objectPath, "public, max-age=31536000, immutable"); err != nil {
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
func saveToLocal(img image.Image, ogPrefix string, objectPath string) error {
	// パス: tmp/images/{env}/{objectPath}
	filePath := filepath.Join("tmp", "images", ogPrefix, filepath.FromSlash(objectPath))

	// ディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
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
func uploadToGCS(ctx context.Context, img image.Image, ogBucket, ogPrefix string, objectPath string, cacheControl string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("GCSクライアント作成失敗: %w", err)
	}
	defer client.Close()

	// オブジェクトパス: {prefix}/{objectPath}
	fullPath := fmt.Sprintf("%s/%s", ogPrefix, objectPath)
	bucket := client.Bucket(ogBucket)
	obj := bucket.Object(fullPath)

	// アップロード
	w := obj.NewWriter(ctx)
	w.ContentType = "image/png"
	w.CacheControl = cacheControl

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
func generateOGPImage(room *models.Room, pal view.GameVersionPalette, fontPath, iconImagePath string) (image.Image, error) {
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
	if err := drawTitleTopLeft(dc, room.Name, scale, fontPath); err != nil {
		return nil, fmt.Errorf("タイトル描画失敗: %w", err)
	}

	// 左下: ゲームバージョン
	if err := drawGameVersionBottomLeft(dc, room.GameVersion.Code, scale, fontPath); err != nil {
		return nil, fmt.Errorf("ゲームバージョン描画失敗: %w", err)
	}

	// 右下: HuntersHubロゴ
	if err := drawHuntersHubLogoBottomRight(dc, scale, fontPath, iconImagePath); err != nil {
		return nil, fmt.Errorf("ロゴ描画失敗: %w", err)
	}

	// 高解像度→等倍へ縮小（Lanczos3）
	hi := dc.Image()
	lo := resize.Resize(uint(OGPWidth), uint(OGPHeight), hi, resize.Lanczos3)
	return lo, nil
}

// generateSiteOGPImage サイト共通のデフォルトOGP画像を生成する。
// 部屋OGPと同じデザイン言語（グラデーション枠 + 白カード）で、
// 中央にロゴ・タグライン・対応タイトルのバッジ列を配置する
func generateSiteOGPImage(fontPath, iconImagePath string) (image.Image, error) {
	scale := float64(RenderScale)
	W := int(float64(OGPWidth) * scale)
	H := int(float64(OGPHeight) * scale)

	dc := gg.NewContext(W, H)

	// 背景: 白
	dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	dc.Clear()

	// 枠はサイト共通のグレー（未定義コードで返る既定パレット）
	if err := drawGradientBorder(dc, view.GetPalette(""), scale); err != nil {
		return nil, fmt.Errorf("枠描画失敗: %w", err)
	}

	centerX := float64(W) / 2

	// 中央: アイコン + サービス名
	dc.SetFontFace(mustLoadFaceTTF(fontPath, SiteNameFontSize*scale))
	name := "HuntersHub"
	nameWidth, _ := dc.MeasureString(name)

	iconSize := SiteIconSize * scale
	iconGap := 24 * scale
	rowWidth := iconSize + iconGap + nameWidth
	rowX := centerX - rowWidth/2
	nameBaselineY := 255 * scale

	if iconImg, err := gg.LoadImage(iconImagePath); err != nil {
		log.Printf("アイコン画像の読み込みに失敗: %v", err)
	} else {
		resizedIcon := resize.Resize(uint(iconSize), uint(iconSize), iconImg, resize.Lanczos3)
		// テキストのベースラインに視覚的に揃える
		iconY := nameBaselineY - iconSize*0.82
		dc.DrawImage(resizedIcon, int(rowX), int(iconY))
	}

	dc.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	dc.DrawString(name, rowX+iconSize+iconGap, nameBaselineY)

	// タグライン
	dc.SetFontFace(mustLoadFaceTTF(fontPath, SiteTaglineFontSize*scale))
	tagline := "モンハンシリーズのパーティ募集掲示板"
	taglineWidth, _ := dc.MeasureString(tagline)
	dc.SetColor(color.RGBA{R: 75, G: 85, B: 99, A: 255})
	dc.DrawString(tagline, centerX-taglineWidth/2, 350*scale)

	// 対応タイトルのバッジ列（各ゲームのパレット色）
	dc.SetFontFace(mustLoadFaceTTF(fontPath, SiteBadgeFontSize*scale))
	badgeHeight := 54 * scale
	badgePadX := 24 * scale
	badgeGap := 16 * scale
	badgeCenterY := 445 * scale

	badgeWidths := make([]float64, len(siteGameCodes))
	totalBadgeWidth := 0.0
	for i, code := range siteGameCodes {
		w, _ := dc.MeasureString(code)
		badgeWidths[i] = w + badgePadX*2
		totalBadgeWidth += badgeWidths[i]
	}
	totalBadgeWidth += badgeGap * float64(len(siteGameCodes)-1)

	x := centerX - totalBadgeWidth/2
	for i, code := range siteGameCodes {
		dc.SetColor(view.GetPalette(code).BottomColor)
		dc.DrawRoundedRectangle(x, badgeCenterY-badgeHeight/2, badgeWidths[i], badgeHeight, badgeHeight/2)
		dc.Fill()

		dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
		dc.DrawStringAnchored(code, x+badgeWidths[i]/2, badgeCenterY, 0.5, 0.37)

		x += badgeWidths[i] + badgeGap
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
func drawTitleTopLeft(dc *gg.Context, title string, s float64, fontPath string) error {
	face := mustLoadFaceTTF(fontPath, TitleFontSize*s)
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
func drawGameVersionBottomLeft(dc *gg.Context, gameCode string, s float64, fontPath string) error {
	face := mustLoadFaceTTF(fontPath, GameVersionFontSize*s)
	dc.SetFontFace(face)

	x := (Padding + BorderWidth + ContentPadding) * s
	y := float64(dc.Height()) - (Padding+BorderWidth+ContentPadding)*s

	// 黒色で描画
	dc.SetColor(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	dc.DrawString(gameCode, x, y)
	return nil
}

// drawHuntersHubLogoBottomRight HuntersHubロゴを右下に描画
func drawHuntersHubLogoBottomRight(dc *gg.Context, s float64, fontPath, iconImagePath string) error {
	// アイコン画像を読み込み
	iconImg, err := gg.LoadImage(iconImagePath)
	if err != nil {
		log.Printf("アイコン画像の読み込みに失敗: %v", err)
		return nil
	}

	// アイコン画像をリサイズ
	iconSize := uint(LogoIconSize * s)
	resizedIcon := resize.Resize(iconSize, iconSize, iconImg, resize.Lanczos3)

	// フォント設定
	dc.SetFontFace(mustLoadFaceTTF(fontPath, LogoFontSize*s))

	// テキスト幅を取得
	text := "HuntersHub"
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

	// HuntersHubテキストを描画
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
