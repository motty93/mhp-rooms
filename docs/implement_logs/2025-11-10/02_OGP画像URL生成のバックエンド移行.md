# OGP画像URL生成のバックエンド移行

## 実装日時
- 開始: 2025-11-10
- 完了: 2025-11-10
- 所要時間: 約30分

## 概要
`/rooms/${roomId}/join` と `/rooms/${roomId}` の両方で、OGP画像URLをテンプレート側からバックエンド側に移行し、コードの保守性と一貫性を向上させました。

## 問題の詳細

### 1. GitHub issue #99の対応
- `/rooms/${roomId}/join` でOGP画像が表示されていない問題
- テンプレートに `og:image` メタタグが設定されていなかった

### 2. コードの重複
- OGP URL生成ロジックが複数のテンプレートに重複
  - `templates/layouts/room_detail.tmpl`
  - `templates/pages/room_join.tmpl`
- テンプレート側で環境変数を直接参照していた

### 3. 保守性の問題
- URL形式変更時に複数箇所の修正が必要
- ビジネスロジックとプレゼンテーション層の責任が曖昧
- テンプレート内のロジックはユニットテストが困難

## 実装内容

### 1. OGP URL生成ヘルパー関数の作成

**ファイル**: `internal/handlers/helpers.go`（新規作成）

```go
package handlers

import (
	"fmt"
	"mhp-rooms/internal/config"
	"github.com/google/uuid"
)

// BuildOGPImageURL OGP画像URLを生成
func BuildOGPImageURL(roomID uuid.UUID, ogVersion int) string {
	ogBucket := config.GetEnv("OG_BUCKET", "")
	ogPrefix := config.GetEnv("OG_PREFIX", "dev")

	if ogBucket != "" {
		return fmt.Sprintf(
			"https://storage.googleapis.com/%s/%s/ogp/rooms/%s.png?v=%d",
			ogBucket, ogPrefix, roomID, ogVersion,
		)
	}

	siteURL := config.GetEnv("SITE_URL", "http://localhost:8080")
	return fmt.Sprintf(
		"%s/tmp/images/%s/ogp/rooms/%s.png?v=%d",
		siteURL, ogPrefix, roomID, ogVersion,
	)
}
```

#### 特徴
- 環境変数（`OG_BUCKET`, `OG_PREFIX`, `SITE_URL`）を使用
- GCS使用時とローカル環境の両方に対応
- `ogVersion` パラメータでキャッシュバスティング

### 2. データ構造の拡張

#### 2.1 `RoomJoinPageData` 構造体
**ファイル**: `internal/handlers/room_join.go`

```go
type RoomJoinPageData struct {
	Room        *RoomBasicInfo `json:"room"`
	IsJoined    bool           `json:"is_joined"`
	IsHost      bool           `json:"is_host"`
	HasPassword bool           `json:"has_password"`
	OGImageURL  string         `json:"og_image_url"`  // 追加
}
```

#### 2.2 `RoomBasicInfo` 構造体
**ファイル**: `internal/handlers/room_join.go`

```go
type RoomBasicInfo struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	RoomCode    string             `json:"room_code"`
	GameVersion models.GameVersion `json:"game_version"`
	Host        models.User        `json:"host"`
	MaxPlayers  int                `json:"max_players"`
	HasPassword bool               `json:"has_password"`
	OGVersion   int                `json:"og_version"`  // 追加
}
```

#### 2.3 `RoomDetailPageData` 構造体
**ファイル**: `internal/handlers/room_detail.go`

```go
type RoomDetailPageData struct {
	Room        *models.Room         `json:"room"`
	Members     []*models.RoomMember `json:"members"`
	Logs        []models.RoomLog     `json:"logs"`
	MemberCount int                  `json:"member_count"`
	IsHost      bool                 `json:"is_host"`
	OGImageURL  string               `json:"og_image_url"`  // 追加
}
```

### 3. ハンドラーの修正

#### 3.1 `RoomJoinHandler.RoomJoinPage`
**ファイル**: `internal/handlers/room_join.go`

```go
basicInfo := &RoomBasicInfo{
	ID:          room.ID,
	Name:        room.Name,
	RoomCode:    room.RoomCode,
	GameVersion: room.GameVersion,
	Host:        room.Host,
	MaxPlayers:  room.MaxPlayers,
	HasPassword: room.HasPassword(),
	OGVersion:   room.OGVersion,  // 追加
}

ogImageURL := BuildOGPImageURL(room.ID, room.OGVersion)  // OGP URL生成

data := TemplateData{
	Title:   room.Name + " - 部屋参加",
	HasHero: false,
	User:    r.Context().Value("user"),
	PageData: RoomJoinPageData{
		Room:        basicInfo,
		IsJoined:    isJoined,
		IsHost:      isHost,
		HasPassword: room.HasPassword(),
		OGImageURL:  ogImageURL,  // 追加
	},
}
```

#### 3.2 `RoomDetailHandler.RoomDetail`
**ファイル**: `internal/handlers/room_detail.go`

```go
ogImageURL := BuildOGPImageURL(room.ID, room.OGVersion)  // OGP URL生成

data := TemplateData{
	Title:   room.Name + " - 部屋詳細",
	HasHero: false,
	User:    r.Context().Value("user"),
	SSEHost: config.AppConfig.Server.SSEHost,
	PageData: RoomDetailPageData{
		Room:        room,
		Members:     memberSlots,
		Logs:        logs,
		MemberCount: memberCount,
		IsHost:      isHost,
		OGImageURL:  ogImageURL,  // 追加
	},
}
```

### 4. テンプレートの修正

#### 4.1 `room_join.tmpl`
**ファイル**: `templates/pages/room_join.tmpl`

**変更前**:
```html
{{ $ogBucket := getEnv "OG_BUCKET" "" }}
{{ $ogPrefix := getEnv "OG_PREFIX" "dev" }}
{{ $siteURL := getEnv "SITE_URL" "http://localhost:8080" }}
{{ $ogImageURL := "" }}
{{ if $ogBucket }}
  {{ $ogImageURL = printf "https://storage.googleapis.com/%s/%s/ogp/rooms/%s.png" $ogBucket $ogPrefix .PageData.Room.ID }}
{{ else }}
  {{ $ogImageURL = printf "%s/tmp/images/%s/ogp/rooms/%s.png" $siteURL $ogPrefix .PageData.Room.ID }}
{{ end }}

<meta property="og:image" content="{{ $ogImageURL }}" />
```

**変更後**:
```html
{{ $siteURL := getEnv "SITE_URL" "http://localhost:8080" }}

<meta property="og:image" content="{{ .PageData.OGImageURL }}" />
```

#### 4.2 `room_detail.tmpl`
**ファイル**: `templates/layouts/room_detail.tmpl`

**変更前**:
```html
{{ $ogBucket := getEnv "OG_BUCKET" "" }}
{{ $ogPrefix := getEnv "OG_PREFIX" "dev" }}
{{ $siteURL := getEnv "SITE_URL" "http://localhost:8080" }}
{{ $ogImageURL := "" }}
{{ if $ogBucket }}
  {{ $ogImageURL = printf "https://storage.googleapis.com/%s/%s/ogp/rooms/%s.png?v=%d" $ogBucket $ogPrefix .PageData.Room.ID .PageData.Room.OGVersion }}
{{ else }}
  {{ $ogImageURL = printf "%s/tmp/images/%s/ogp/rooms/%s.png?v=%d" $siteURL $ogPrefix .PageData.Room.ID .PageData.Room.OGVersion }}
{{ end }}

<meta property="og:image" content="{{ $ogImageURL }}" />
```

**変更後**:
```html
{{ $siteURL := getEnv "SITE_URL" "http://localhost:8080" }}

<meta property="og:image" content="{{ .PageData.OGImageURL }}" />
```

## 変更ファイル一覧

- **新規作成**: `internal/handlers/helpers.go`
- **修正**: `internal/handlers/room_join.go`
- **修正**: `internal/handlers/room_detail.go`
- **修正**: `templates/pages/room_join.tmpl`
- **修正**: `templates/layouts/room_detail.tmpl`

## 期待される効果

### 1. 保守性の向上
- OGP URL生成ロジックが1箇所（`helpers.go`）に集約
- URL形式変更時の修正箇所が1箇所のみ

### 2. 一貫性の向上
- バックエンドで統一的にURL生成
- 複数のページで同じロジックが使用される

### 3. テスタビリティの向上
- Go関数としてユニットテストが可能
- テンプレートのロジックが減り、テストが簡単

### 4. 責任の分離
- **テンプレートの責任**: データの表示のみ
- **ハンドラーの責任**: ビジネスロジック（OGP URL生成含む）
- 明確な責任分担により、コードの見通しが向上

## 動作確認

### ビルド確認
```bash
go build ./...
```
✅ ビルド成功

### 確認項目（本番環境デプロイ後）

1. `/rooms/${roomId}/join` のOGP画像が表示されること
2. `/rooms/${roomId}` のOGP画像が引き続き表示されること
3. Twitter/Discord等でシェア時に部屋のOGP画像が表示されること

## 技術的な詳細

### OGP画像URLの形式

#### 本番環境（GCS使用時）
```
https://storage.googleapis.com/{OG_BUCKET}/{OG_PREFIX}/ogp/rooms/{roomID}.png?v={ogVersion}
```

#### 開発環境
```
http://localhost:8080/tmp/images/{OG_PREFIX}/ogp/rooms/{roomID}.png?v={ogVersion}
```

### キャッシュバスティング
- `ogVersion` パラメータでキャッシュ無効化
- 部屋情報更新時に `ogVersion` をインクリメントすることで、ブラウザキャッシュを回避

## 今後の改善案

### さらなる共通化の可能性
- デフォルトOGP画像（`base.tmpl`）の生成もバックエンドに移行
- OGP関連のメタタグ生成を共通コンポーネント化

### ユニットテストの追加
```go
func TestBuildOGPImageURL(t *testing.T) {
	// GCS環境のテスト
	os.Setenv("OG_BUCKET", "test-bucket")
	os.Setenv("OG_PREFIX", "stg")

	roomID := uuid.New()
	ogVersion := 1

	url := BuildOGPImageURL(roomID, ogVersion)
	expected := fmt.Sprintf(
		"https://storage.googleapis.com/test-bucket/stg/ogp/rooms/%s.png?v=1",
		roomID,
	)

	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}
```

## 参考資料
- GitHub issue #99: OGP画像 共有URLの`/join`で生成された部屋詳細のOGP画像が表示されていないので改善する
- Open Graph Protocol: https://ogp.me/
