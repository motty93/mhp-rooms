package utils

// GetGameVersionColor はゲームバージョンに応じた色を返す
func GetGameVersionColor(code string) string {
	switch code {
	case "MHP":
		return "bg-amber-800" // 茶色系に近い色
	case "MHP2":
		return "bg-sky-600" // 青色系
	case "MHP2G":
		return "bg-green-700" // 緑色系
	case "MHP3":
		return "bg-yellow-600" // 金色系
	default:
		return "bg-gray-600"
	}
}

// GetGameVersionIcon はゲームバージョンに応じたアイコンSVGを返す
func GetGameVersionIcon(code string) string {
	switch code {
	case "MHP":
		// 1を表すアイコン
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-xl font-bold">1</text>
		</svg>`
	case "MHP2":
		// 2を表すアイコン
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-xl font-bold">2</text>
		</svg>`
	case "MHP2G":
		// 2Gを表すアイコン
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-sm font-bold">2G</text>
		</svg>`
	case "MHP3":
		// 3を表すアイコン
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<text x="50%" y="50%" text-anchor="middle" dy=".35em" class="text-xl font-bold">3</text>
		</svg>`
	default:
		// デフォルトアイコン（家のアイコン）
		return `<svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
			<path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" />
		</svg>`
	}
}

// GetGameVersionAbbreviation はゲームバージョンの略称を返す
func GetGameVersionAbbreviation(code string) string {
	switch code {
	case "MHP":
		return "1"
	case "MHP2":
		return "2"
	case "MHP2G":
		return "2G"
	case "MHP3":
		return "3"
	default:
		return "?"
	}
}
