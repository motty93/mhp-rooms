package palette

import "image/color"

type GameVersionPalette struct {
	TopColor    color.RGBA
	BottomColor color.RGBA
	AccentColor color.RGBA
}

var GameVersionPalettes = map[string]GameVersionPalette{
	"MHP": {
		TopColor:    color.RGBA{R: 139, G: 69, B: 19, A: 255},
		BottomColor: color.RGBA{R: 101, G: 50, B: 14, A: 255},
		AccentColor: color.RGBA{R: 180, G: 120, B: 80, A: 255},
	},
	"MHP2": {
		TopColor:    color.RGBA{R: 70, G: 130, B: 180, A: 255},
		BottomColor: color.RGBA{R: 50, G: 90, B: 140, A: 255},
		AccentColor: color.RGBA{R: 135, G: 206, B: 250, A: 255},
	},
	"MHP2G": {
		TopColor:    color.RGBA{R: 34, G: 139, B: 34, A: 255},
		BottomColor: color.RGBA{R: 24, G: 100, B: 24, A: 255},
		AccentColor: color.RGBA{R: 144, G: 238, B: 144, A: 255},
	},
	"MHP3": {
		TopColor:    color.RGBA{R: 218, G: 165, B: 32, A: 255},
		BottomColor: color.RGBA{R: 184, G: 134, B: 11, A: 255},
		AccentColor: color.RGBA{R: 255, G: 215, B: 100, A: 255},
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
