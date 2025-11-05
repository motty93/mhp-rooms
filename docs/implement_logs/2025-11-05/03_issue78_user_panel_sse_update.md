# 部屋詳細画面 - 入退室時のユーザーパネル更新（SSE実装）

**実装日**: 2025-11-05
**実装時間**: 約60分
**対応Issue**: #78

## 概要

部屋詳細画面で、ユーザーが入退室した際にシステムメッセージは表示されるものの、ユーザーパネルが更新されない問題を解決しました。SSEを使用して、入退室時にユーザーパネルをリアルタイムで更新する機能を実装しました。

## 問題点

- 入室時にシステムメッセージ「○○さんが入室しました」は表示される
- しかし、ユーザーパネルは空のまま更新されない
- ページをリロードすると正しく表示される

## 実装内容

### 1. バックエンド: SSEイベントの追加

**ファイル**: `internal/handlers/rooms.go`

#### JoinRoomハンドラー (行526-542)

既存の `system_message` イベントに加えて、新しく `member_update` イベントを追加しました。

```go
// メンバー更新イベント（ユーザーパネル用）
members, err := h.repo.Room.GetRoomMembers(roomID)
if err != nil {
    log.Printf("メンバー情報取得エラー: %v", err)
    members = []models.RoomMember{} // エラー時は空配列
}

memberUpdateEvent := sse.Event{
    ID:   uuid.New().String(),
    Type: "member_update",
    Data: map[string]interface{}{
        "action":  "join",
        "members": members,
        "count":   len(members),
    },
}
h.hub.BroadcastToRoom(roomID, memberUpdateEvent)
```

#### LeaveRoomハンドラー (行615-631)

退室時も同様に `member_update` イベントを追加しました。

```go
// メンバー更新イベント（ユーザーパネル用）
members, err := h.repo.Room.GetRoomMembers(roomID)
if err != nil {
    log.Printf("メンバー情報取得エラー: %v", err)
    members = []models.RoomMember{} // エラー時は空配列
}

memberUpdateEvent := sse.Event{
    ID:   uuid.New().String(),
    Type: "member_update",
    Data: map[string]interface{}{
        "action":  "leave",
        "members": members,
        "count":   len(members),
    },
}
h.hub.BroadcastToRoom(roomID, memberUpdateEvent)
```

### 2. フロントエンド: Alpine.jsデータ構造の拡張

**ファイル**: `templates/components/room_detail_script.tmpl`

#### データ構造追加 (行55-57)

```javascript
// ユーザーパネル用メンバー配列（4人分のスロット）
members: [],
memberCount: {{ .PageData.MemberCount }},
```

#### 初期化メソッド (行930-954)

ページロード時に初期メンバー情報を読み込むメソッドを実装しました。

```javascript
loadInitialMembers() {
    const memberSlots = [null, null, null, null];
    let count = 0;

    {{ range $member := .PageData.Members }}
    {{ if $member }}
    const playerNum = {{ $member.PlayerNumber }};
    if (playerNum >= 1 && playerNum <= 4) {
        memberSlots[playerNum - 1] = {
            id: '{{ $member.User.ID }}',
            supabase_user_id: '{{ $member.User.SupabaseUserID }}',
            display_name: '{{ jsEscape $member.DisplayName }}',
            avatar_url: '{{ jsEscapePtr $member.User.AvatarURL }}' || '/static/images/default-avatar.webp',
            is_host: {{ $member.IsHost }},
            player_number: {{ $member.PlayerNumber }}
        };
        count++;
    }
    {{ end }}
    {{ end }}

    this.members = memberSlots;
    this.memberCount = count;
}
```

#### SSEイベントハンドラー追加 (行341-343)

既存のSSEイベントハンドラーに `member_update` の処理を追加しました。

```javascript
} else if (type === 'member_update') {
    this.handleMemberUpdate(json.data);
}
```

#### メンバー更新ハンドラー (行956-982)

SSEイベント受信時にメンバー配列を更新するメソッドを実装しました。

```javascript
handleMemberUpdate(data) {
    const memberSlots = [null, null, null, null];
    let count = 0;

    if (data.members && Array.isArray(data.members)) {
        data.members.forEach(member => {
            if (member.player_number >= 1 && member.player_number <= 4) {
                memberSlots[member.player_number - 1] = {
                    id: member.user.id,
                    supabase_user_id: member.user.supabase_user_id,
                    display_name: member.user.display_name || member.user.username,
                    avatar_url: member.user.avatar_url || '/static/images/default-avatar.webp',
                    is_host: member.is_host,
                    player_number: member.player_number
                };
                count++;
            }
        });
    }

    this.members = memberSlots;
    this.memberCount = count;

    // Alpine.jsの反応を促す
    this.$nextTick();
}
```

### 3. テンプレート: ユーザーパネルの動的化

**ファイル**: `templates/pages/room_detail.tmpl`

#### デスクトップ版ユーザーパネル (行387-429)

静的HTMLから Alpine.js の `x-for` ループによる動的レンダリングに変更しました。

```html
<div class="flex-1 px-4 pt-4 space-y-3 overflow-y-auto">
    <template x-for="(member, index) in members" :key="index">
        <div>
            <!-- メンバーが存在する場合 -->
            <template x-if="member">
                <a :href="'/users/' + member.id"
                   class="w-full flex items-center space-x-3 p-2 bg-gray-100 hover:bg-gray-200 rounded border-2 transition-colors cursor-pointer"
                   :class="{'border-blue-500': currentUserId === member.supabase_user_id, 'border-gray-200': currentUserId !== member.supabase_user_id}">
                    <img :src="member.avatar_url || '/static/images/default-avatar.webp'"
                         class="w-8 h-8 rounded-full object-cover"
                         :alt="member.display_name + 'のアバター'" />
                    <span class="text-gray-800 text-sm font-medium" x-text="member.display_name"></span>
                    <span x-show="member.is_host"
                          class="text-xs bg-yellow-100 text-yellow-700 px-2 py-0.5 rounded">ホスト</span>
                </a>
            </template>

            <!-- 空きスロット -->
            <template x-if="!member">
                <div class="flex items-center space-x-3 p-2 bg-gray-50 rounded border border-gray-200 opacity-60">
                    <div class="w-6 h-8 bg-gray-300 rounded flex items-center justify-center">
                        <div class="w-4 h-6 bg-gray-500 rounded-sm"></div>
                    </div>
                    <span class="text-gray-400 text-sm">－</span>
                </div>
            </template>
        </div>
    </template>
</div>
```

#### モバイルヘッダー参加者数 (行21)

参加者数表示も動的化しました。

```html
<div class="text-gray-500 text-sm">
    参加者 <span x-text="memberCount"></span>/4
</div>
```

## 実装のポイント

### 1. イベント分離

**重要**: メッセージSSEとユーザーパネルSSEを分離することで、無駄なリクエストを防止しました。

- **system_message**: チャットエリアに表示（メッセージ送信時も動作）
- **member_update**: ユーザーパネルのみ更新（入退室時のみ動作）

この設計により、メッセージが送信されるたびにメンバー情報を取得・更新する無駄を回避できました。

### 2. データ一貫性の保証

- メンバー配列は常に4要素（null埋め）
- player_number (1-4) をインデックス (0-3) に変換
- メンバー数（count）も同時に更新

### 3. エラーハンドリング

- メンバー情報取得失敗時は空配列を送信
- フロントエンドでは既存表示を維持
- SSEイベント処理エラーはコンソールログのみ

### 4. リアクティビティの確保

Alpine.js のリアクティブシステムを活用し、`$nextTick()` で確実に再レンダリングを発火させました。

## 動作確認

### 確認項目

- ✅ ページロード時に初期メンバー情報が正しく表示
- ✅ ユーザーが入室すると、ユーザーパネルにメンバーが追加表示
- ✅ チャットエリアに「○○さんが入室しました」と表示
- ✅ 参加者数が更新される
- ✅ ユーザーが退室すると、ユーザーパネルからメンバーが削除（空きスロット表示）
- ✅ チャットエリアに「○○さんが退室しました」と表示
- ✅ 参加者数が更新される
- ✅ 複数ブラウザで同時に開いた場合、全てのクライアントで同期して更新

## 技術的な学び

### SSEイベントの設計

1つの機能（入退室）に対して複数のSSEイベントを送信することで、異なるUI要素を独立して更新できることを学びました。

**Before (問題)**:
- system_messageのみ → チャットエリアは更新されるが、ユーザーパネルは更新されない

**After (解決)**:
- system_message → チャットエリア更新
- member_update → ユーザーパネル更新

### Alpine.js x-forとテンプレート

Go テンプレートと Alpine.js テンプレートを混在させる際の注意点：
- 初期値はGo テンプレートから注入
- 動的更新は Alpine.js の `x-for` で処理
- `x-if` と `x-for` を組み合わせて条件付きレンダリング

## 改善の余地

### 1. パフォーマンス

現在は入退室ごとに全メンバー情報を取得していますが、差分のみを送信する方式も検討できます。

### 2. エラー通知

メンバー情報取得失敗時、ユーザーに通知する仕組みがあると良いかもしれません。

### 3. アニメーション

メンバー追加・削除時にアニメーションを追加すると、UXが向上します。

## まとめ

SSEを活用して、部屋詳細画面のユーザーパネルをリアルタイム更新する機能を実装しました。イベントを適切に分離することで、無駄なリクエストを防ぎつつ、必要な箇所のみを効率的に更新できるようになりました。

この実装により、ユーザーは入退室状況を即座に把握でき、より良いユーザー体験を提供できるようになりました。
