package main

import (
	"testing"
)

// TestGenerateSiteOGPImage サイト用OGPがアセット込みで描画でき、規定サイズで出力されることを確認する
func TestGenerateSiteOGPImage(t *testing.T) {
	img, err := generateSiteOGPImage(
		"assets/fonts/NotoSansCJKjp-Bold.otf",
		"assets/images/icon.webp",
	)
	if err != nil {
		t.Fatalf("generateSiteOGPImage: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != OGPWidth || bounds.Dy() != OGPHeight {
		t.Errorf("サイズ = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), OGPWidth, OGPHeight)
	}
}
