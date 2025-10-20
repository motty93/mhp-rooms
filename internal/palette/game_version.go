package palette

import "image/color"

type GameVersionPalette struct {
	TopColor    color.RGBA
	BottomColor color.RGBA
	AccentColor color.RGBA
}

var GameVersionPalettes = map[string]GameVersionPalette{
	"MHP": {
		TopColor:    color.RGBA{R: 190, G: 140, B: 110, A: 255}, // 明るめの茶色（少し濃く）
		BottomColor: color.RGBA{R: 165, G: 120, B: 95, A: 255},  // やや濃い茶色
		AccentColor: color.RGBA{R: 210, G: 170, B: 145, A: 255},
	},
	"MHP2": {
		TopColor:    color.RGBA{R: 145, G: 190, B: 225, A: 255}, // 明るめの青（少し濃く）
		BottomColor: color.RGBA{R: 120, G: 165, B: 205, A: 255}, // やや濃い青
		AccentColor: color.RGBA{R: 175, G: 210, B: 240, A: 255},
	},
	"MHP2G": {
		TopColor:    color.RGBA{R: 145, G: 205, B: 165, A: 255}, // 明るめの緑（少し濃く）
		BottomColor: color.RGBA{R: 120, G: 180, B: 140, A: 255}, // やや濃い緑
		AccentColor: color.RGBA{R: 175, G: 225, B: 195, A: 255},
	},
	"MHP3": {
		TopColor:    color.RGBA{R: 230, G: 195, B: 135, A: 255}, // 明るめのゴールド（少し濃く）
		BottomColor: color.RGBA{R: 210, G: 175, B: 115, A: 255}, // やや濃いゴールド
		AccentColor: color.RGBA{R: 245, G: 215, B: 165, A: 255},
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
