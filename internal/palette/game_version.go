package palette

import "image/color"

type GameVersionPalette struct {
	TopColor    color.RGBA
	BottomColor color.RGBA
	AccentColor color.RGBA
}

var GameVersionPalettes = map[string]GameVersionPalette{
	"MHP": {
		TopColor:    color.RGBA{R: 210, G: 160, B: 130, A: 255}, // 明るい茶色
		BottomColor: color.RGBA{R: 140, G: 100, B: 70, A: 255},  // 濃い茶色
		AccentColor: color.RGBA{R: 230, G: 190, B: 165, A: 255},
	},
	"MHP2": {
		TopColor:    color.RGBA{R: 165, G: 210, B: 245, A: 255}, // 明るい青
		BottomColor: color.RGBA{R: 90, G: 140, B: 180, A: 255},  // 濃い青
		AccentColor: color.RGBA{R: 195, G: 230, B: 255, A: 255},
	},
	"MHP2G": {
		TopColor:    color.RGBA{R: 165, G: 225, B: 185, A: 255}, // 明るい緑
		BottomColor: color.RGBA{R: 90, G: 150, B: 110, A: 255},  // 濃い緑
		AccentColor: color.RGBA{R: 195, G: 245, B: 215, A: 255},
	},
	"MHP3": {
		TopColor:    color.RGBA{R: 250, G: 215, B: 155, A: 255}, // 明るいゴールド
		BottomColor: color.RGBA{R: 180, G: 145, B: 85, A: 255},  // 濃いゴールド
		AccentColor: color.RGBA{R: 255, G: 235, B: 185, A: 255},
	},
}

func GetPalette(gameVersionCode string) GameVersionPalette {
	if palette, exists := GameVersionPalettes[gameVersionCode]; exists {
		return palette
	}

	return GameVersionPalette{
		TopColor:    color.RGBA{R: 107, G: 114, B: 128, A: 255},
		BottomColor: color.RGBA{R: 55, G: 65, B: 81, A: 255},
		AccentColor: color.RGBA{R: 209, G: 213, B: 219, A: 255},
	}
}
