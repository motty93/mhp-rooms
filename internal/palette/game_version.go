package palette

import "image/color"

type GameVersionPalette struct {
	TopColor    color.RGBA
	BottomColor color.RGBA
	AccentColor color.RGBA
}

var GameVersionPalettes = map[string]GameVersionPalette{
	"MHP": {
		TopColor:    color.RGBA{R: 230, G: 180, B: 150, A: 255}, // より明るい茶色
		BottomColor: color.RGBA{R: 120, G: 80, B: 50, A: 255},   // より濃い茶色
		AccentColor: color.RGBA{R: 250, G: 210, B: 185, A: 255},
	},
	"MHP2": {
		TopColor:    color.RGBA{R: 185, G: 230, B: 255, A: 255}, // より明るい青
		BottomColor: color.RGBA{R: 70, G: 120, B: 160, A: 255},  // より濃い青
		AccentColor: color.RGBA{R: 215, G: 250, B: 255, A: 255},
	},
	"MHP2G": {
		TopColor:    color.RGBA{R: 185, G: 245, B: 205, A: 255}, // より明るい緑
		BottomColor: color.RGBA{R: 70, G: 130, B: 90, A: 255},   // より濃い緑
		AccentColor: color.RGBA{R: 215, G: 255, B: 235, A: 255},
	},
	"MHP3": {
		TopColor:    color.RGBA{R: 255, G: 235, B: 175, A: 255}, // より明るいゴールド
		BottomColor: color.RGBA{R: 160, G: 125, B: 65, A: 255},  // より濃いゴールド
		AccentColor: color.RGBA{R: 255, G: 245, B: 205, A: 255},
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
