# 部屋詳細メッセージ機能設計書（htmx × SSE）

## 概要
部屋詳細ページにリアルタイムメッセージ機能を実装する。htmxとServer-Sent Events（SSE）を使用して、リアルタイムでメッセージの送受信を行う。

## 技術構成
- **フロントエンド**: htmx + SSE（Alpine.js併用）
- **バックエンド**: Go（Gorilla Mux）
- **リアルタイム通信**: Server-Sent Events（SSE）
- **データ永続化**: PostgreSQL（既存のRoomMessageテーブル）

## データベース設計

### 既存のRoomMessageテーブル（変更なし）
```sql
-- すでに定義済み
CREATE TABLE room_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id),
    user_id UUID NOT NULL REFERENCES users(id),
    message TEXT NOT NULL,
    message_type VARCHAR(20) NOT NULL DEFAULT 'chat',
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### メッセージタイプ
- `chat`: 通常のチャットメッセージ
- `system`: システムメッセージ（入退室など）
- `action`: アクション通知（準備完了など）

## API設計

### 1. 部屋詳細ページ
```
GET /rooms/{id}
```
- 部屋情報とメンバー一覧を表示
- メッセージ履歴（最新20件）を表示
- SSE接続用のエンドポイント情報を含む

### 2. SSEエンドポイント
```
GET /rooms/{id}/messages/stream
```
- Server-Sent Eventsでリアルタイムメッセージを配信
- 認証必須（部屋のメンバーのみ接続可能）
- イベントタイプ：
  - `message`: 新しいメッセージ
  - `member_join`: メンバー参加
  - `member_leave`: メンバー退室
  - `room_update`: 部屋情報更新

### 3. メッセージ送信
```
POST /rooms/{id}/messages
Content-Type: application/x-www-form-urlencoded

message=テストメッセージ
```
- htmxのhx-postで送信
- レスポンスはHTMLフラグメント（htmxで部分更新）

### 4. メッセージ履歴取得
```
GET /rooms/{id}/messages?before={message_id}&limit=20
```
- 無限スクロール用のページネーション
- htmxのhx-triggerで自動読み込み

## フロントエンド実装

### 部屋詳細ページのHTML構造
```html
<!-- templates/pages/room_detail.tmpl -->
<div class="room-detail" data-room-id="{{ .Room.ID }}">
  <!-- 部屋情報ヘッダー -->
  <div class="room-header">
    <h1>{{ .Room.Name }}</h1>
    <div class="room-info">
      <span>{{ .Room.GameVersion.Name }}</span>
      <span>{{ .Room.CurrentPlayers }}/{{ .Room.MaxPlayers }}人</span>
    </div>
  </div>

  <!-- メッセージエリア -->
  <div class="message-area">
    <!-- メッセージリスト（SSE接続） -->
    <div id="message-list" 
         hx-ext="sse" 
         sse-connect="/rooms/{{ .Room.ID }}/messages/stream"
         sse-swap="message">
      <!-- 既存メッセージ -->
      {{ range .Messages }}
        {{ template "message_item" . }}
      {{ end }}
    </div>

    <!-- メッセージ入力フォーム -->
    <form hx-post="/rooms/{{ .Room.ID }}/messages"
          hx-target="#message-list"
          hx-swap="beforeend"
          hx-on::after-request="this.reset()">
      <input type="text" 
             name="message" 
             placeholder="メッセージを入力..."
             required>
      <button type="submit">送信</button>
    </form>
  </div>

  <!-- メンバーリスト -->
  <div class="member-list">
    <h3>参加メンバー</h3>
    <ul id="member-list">
      {{ range .Room.Members }}
        {{ template "member_item" . }}
      {{ end }}
    </ul>
  </div>
</div>
```

### メッセージアイテムテンプレート
```html
<!-- templates/components/message_item.tmpl -->
<div class="message-item" data-message-id="{{ .ID }}">
  <div class="message-header">
    <span class="username">{{ .User.Username }}</span>
    <time>{{ .CreatedAt.Format "15:04" }}</time>
  </div>
  <div class="message-content">{{ .Message }}</div>
</div>
```

## バックエンド実装概要

### SSEハンドラー
```go
// internal/handlers/room_messages.go
func (h *RoomMessageHandler) StreamMessages(w http.ResponseWriter, r *http.Request) {
    // SSE設定
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    
    // クライアント接続管理
    // メッセージブロードキャスト
    // 接続エラー処理
}
```

### メッセージ送信ハンドラー
```go
func (h *RoomMessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
    // メッセージ検証
    // DB保存
    // SSEでブロードキャスト
    // HTMLフラグメント返却
}
```

## セキュリティ考慮事項
1. **認証・認可**
   - 部屋のメンバーのみメッセージ送受信可能
   - SSE接続時の認証チェック

2. **入力検証**
   - XSS対策（HTMLエスケープ）
   - メッセージ長制限（1000文字）
   - レート制限（1秒に1メッセージ）

3. **リソース管理**
   - SSE接続数の制限
   - 自動切断機能（30分無活動）
   - メモリリーク対策

## パフォーマンス最適化
1. **メッセージキャッシュ**
   - Redis使用（オプション）
   - 最新20件をキャッシュ

2. **接続管理**
   - 部屋ごとの接続プール
   - 効率的なブロードキャスト

3. **フロントエンド最適化**
   - Virtual Scrolling（大量メッセージ対応）
   - 遅延読み込み

## 実装手順
1. 部屋詳細ページの基本実装
2. SSEエンドポイント実装
3. メッセージ送信機能実装
4. リアルタイム更新機能実装
5. メンバー入退室通知実装
6. エラーハンドリング・再接続処理
7. テスト実装