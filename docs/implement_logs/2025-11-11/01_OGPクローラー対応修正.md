# OGPクローラー対応修正

## 実装時間
- 開始: 会話開始時
- 完了: 実装ログ作成時
- 所要時間: 約30分

## 実装概要

`/rooms/{roomId}/join` ルートでOGP画像が表示されない問題を修正しました。

### 問題の原因

- `/rooms/{id}/join` ルートに `withAuth`（認証必須）ミドルウェアが適用されていた
- OGPクローラー（Facebook, Twitter, Slackなど）は認証情報を持たないため、ログインページへリダイレクトされていた
- その結果、ログインページのOGPタグ（「ログイン - HuntersHub」）が読み込まれていた

### 修正内容

#### 1. ルーティングの変更 (`cmd/server/routes.go:120`)

```go
// 修正前
rr.Get("/{id}/join", app.withAuth(rjh.RoomJoinPage))

// 修正後（認証オプショナルに変更）
rr.Get("/{id}/join", app.withOptionalAuth(rjh.RoomJoinPage))
```

#### 2. クローラー検出機能の追加 (`internal/handlers/room_join.go`)

**新規追加: `isCrawler` 関数**
- User-Agentから一般的なOGPクローラーを検出
- 対応クローラー: Facebook, Twitter, Slack, Discord, Telegram, LinkedIn, WhatsApp, Google, Bingなど

```go
func isCrawler(userAgent string) bool {
    userAgent = strings.ToLower(userAgent)
    crawlerKeywords := []string{
        "bot", "crawler", "spider", "facebookexternalhit", "twitterbot",
        "slackbot", "discordbot", "telegrambot", "linkedinbot", "whatsapp",
        "headlesschrome", "lighthouse", "googlebot", "bingbot",
    }
    for _, keyword := range crawlerKeywords {
        if strings.Contains(userAgent, keyword) {
            return true
        }
    }
    return false
}
```

#### 3. ハンドラーロジックの改善 (`RoomJoinPage`)

**認証チェックの条件分岐**
```go
// クローラー判定
userAgent := r.Header.Get("User-Agent")
isBot := isCrawler(userAgent)

dbUser, exists := middleware.GetDBUserFromContext(r.Context())

// 通常のユーザー（クローラーでない）かつ未認証の場合はログインへリダイレクト
if !isBot && (!exists || dbUser == nil) {
    redirectURL := "/auth/login?redirect=" + r.URL.Path
    http.Redirect(w, r, redirectURL, http.StatusFound)
    return
}
```

**参加状態チェックの条件追加**
```go
// 認証済みユーザーの場合のみ参加状態チェックとリダイレクト
var isJoined, isHost bool
if exists && dbUser != nil {
    isJoined = h.repo.Room.IsUserJoinedRoom(roomID, dbUser.ID)
    if isJoined {
        http.Redirect(w, r, "/rooms/"+roomID.String(), http.StatusFound)
        return
    }

    isHost = dbUser.ID == room.HostUserID
    if isHost {
        http.Redirect(w, r, "/rooms/"+roomID.String(), http.StatusFound)
        return
    }
}
```

#### 4. 未認証クローラー向け限定表示

- クローラー判定かつ未認証の場合は `IsLimitedView` フラグを立て、テンプレート側で参加フォームやJSを描画しないようにしました。
- OGPメタ情報は維持しつつ、画面上には「OGPプレビュー専用ページ」のみを表示します。

## 変更ファイル

1. `cmd/server/routes.go`
   - `/rooms/{id}/join` のミドルウェアを `withAuth` → `withOptionalAuth` に変更

2. `internal/handlers/room_join.go`
   - `strings` パッケージをインポート
   - `isCrawler` 関数、`IsLimitedView` フラグを追加
   - `RoomJoinPage` ハンドラーにクローラー検出ロジックを追加
   - 認証チェックとリダイレクトの条件を改善、未認証クローラーは限定表示

3. `templates/pages/room_join.tmpl`
   - `IsLimitedView` が true の場合は参加UIを隠し、メタ情報取得専用のメッセージだけを表示
   - 参加フォームとクライアントJSは認証済みユーザーにのみ描画

## 動作確認

### クローラーからのアクセス
- 認証なしでもOGPタグ付きHTMLが返される
- 正しい部屋のOGP画像が表示される

### 未認証の一般ユーザー
- ログインページへリダイレクト（従来通り）

### 認証済みユーザー
- 通常通り参加状態をチェック
- 既に参加している場合は部屋詳細ページへリダイレクト
- ホストの場合は部屋詳細ページへリダイレクト

## 特に注意した点

1. **セキュリティ**: クローラーにはOGP情報のみを提供し、実際の参加機能は認証済みユーザーのみに制限
2. **後方互換性**: 既存の認証済みユーザーの動作に影響を与えない
3. **クローラー検出**: 主要なSNSプラットフォームのクローラーを網羅的にカバー

## 今後の改善案

- クローラー検出キーワードリストの外部設定化（環境変数やデータベース）
- アクセスログへのクローラー判定情報の記録
- より詳細なUser-Agent解析ライブラリの導入検討
