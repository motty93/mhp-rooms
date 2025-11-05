# 部屋詳細画面 - 入退室時のユーザーパネル更新（SSE実装）

**実装日**: 2025-11-05
**実装時間**: 約180分（初期実装60分 + バグ修正・改善120分）
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

## バグ修正と追加改善

### Issue 1: SSEイベント形式のミスマッチ

**問題**: ユーザーパネルが更新されない根本原因を発見。

**原因**:
- SSE Hub が `event: member_update` ヘッダーで送信
- フロントエンドは `event: message` のみをリスニング

**修正**: `internal/infrastructure/sse/hub.go` の `SerializeEvent` 関数 (117-120行目)

```go
// すべてのイベントを "message" イベントとして送信し、
// データ内の type フィールドで区別する
return fmt.Sprintf("id: %s\nevent: message\ndata: %s\n\n",
    event.ID, string(data)), nil
```

SSEの仕様上、`event:` ヘッダーで分けるのではなく、全て `message` イベントとして送信し、JSONの `type` フィールドで区別するように統一しました。

### Issue 2: DisplayNameが表示されない

**問題**: 退出後、残ったユーザーの `display_name` が空になる

**原因**: `GetRoomMembers` 関数で DisplayName の設定処理が抜けていた

**修正**: `internal/repository/room_repository.go` (534-542行目)

```go
// DisplayNameを設定（display_name > username の優先順位）
for i := range members {
    displayName := members[i].User.DisplayName
    // display_nameが空の場合はusernameを使用
    if displayName == "" && members[i].User.Username != nil && *members[i].User.Username != "" {
        displayName = *members[i].User.Username
    }
    members[i].DisplayName = displayName
}
```

### Issue 3: システムメッセージの永続化

**問題**: 入退室時のシステムメッセージがページ更新で消える

**背景**:
- `room_messages` テーブルは元々ユーザーチャット専用（2025-07-28実装）
- システムメッセージはSSE経由のみで、`room_logs` に統計用として保存（2025-11-05, Issue #72）

**修正**: システムメッセージも `room_messages` テーブルに保存

`internal/handlers/rooms.go` の JoinRoom/LeaveRoom/LeaveCurrentRoom ハンドラーに追加:

```go
joinMessage := models.RoomMessage{
    BaseModel: models.BaseModel{
        ID: uuid.New(),
    },
    RoomID:      roomID,
    UserID:      userID,
    Message:     fmt.Sprintf("%sさんが入室しました", displayName),
    MessageType: "system",  // ← 重要: system/chat で区別
}
joinMessage.User = *dbUser

// DBに保存
if err := h.repo.RoomMessage.CreateMessage(&joinMessage); err != nil {
    log.Printf("入室メッセージの保存に失敗: %v", err)
}
```

### Issue 4: システムメッセージのスタイル崩れ

**問題**: ページ更新後、システムメッセージがユーザーの吹き出しとして表示される

**原因**: `loadInitialMessages` が全メッセージを `type: 'user'` として扱っていた

**修正**: `templates/components/room_detail_script.tmpl` (230-259行目)

```javascript
messages.forEach(msg => {
    // message_typeをチェックしてシステムメッセージとユーザーメッセージを区別
    if (msg.message_type === 'system') {
        // システムメッセージ（入退室など）
        const isLeave = msg.message && msg.message.includes('退室しました');
        const subtype = isLeave ? 'leave' : 'join';
        this.messages.push({
            id: msg.id,
            type: 'system',
            subtype: subtype,
            content: msg.message,
            timestamp: new Date(msg.created_at)
        });
    } else {
        // ユーザーメッセージ（チャット）
        this.messages.push({
            id: msg.id,
            type: 'user',
            content: msg.message,
            userName: msg.user.display_name || msg.user.username,
            userAvatar: msg.user.avatar_url || '/static/images/default-avatar.webp',
            isOwn: isOwn,
            timestamp: new Date(msg.created_at)
        });
    }
});
```

### Issue 5: モバイルでメッセージ入力フォームが消える

**問題**: モバイル表示でメッセージ入力フォームが一瞬表示されてすぐ消える

**原因**: `.mobile-chat-area` に固定高さ `calc(100vh - 200px)` が設定され、flexboxの自動調整が上書きされていた

**修正**: `static/css/style.css` (281-299行目)

```css
/* モバイル版のメッセージフォームを画面下部に固定 */
.mobile-message-form {
    position: sticky;
    bottom: 0;
    z-index: 10;
    background-color: white;
}
```

`templates/pages/room_detail.tmpl` (572行目)

```html
<div class="bg-white border-t border-gray-200 p-4 flex-shrink-0 md:relative mobile-message-form">
```

- モバイル: `position: sticky` で画面下部に固定
- デスクトップ: `md:relative` で通常のフロー配置

### Issue 6: モバイルのplaceholder

**問題**: モバイルでも「Shift+Enterで改行」という不要な説明が表示される

**修正**: `templates/pages/room_detail.tmpl` (598行目)

```html
<textarea
    id="message-input"
    name="message"
    :placeholder="window.innerWidth >= 768 ? 'メッセージを入力（Shift+Enterで改行）...' : 'メッセージを入力...'"
    ...
></textarea>
```

Alpine.js の `:placeholder` バインディングで画面幅に応じて動的に切り替え。

### Issue 7: ユーザーパネルのスロット詰め問題

**問題**: 2番目のユーザーが退出 → 3番目が2番目に移動 → 再入室時に上書きされる

**試行錯誤の過程**:

1. **初回実装**: `GetRoomMembers` でメモリ上で詰める
2. **問題発覚**: 再入室時の上書き問題なし、しかし更新後にスロットが空く
3. **誤った修正**: `LeaveRoom` でDBの `player_number` を更新
4. **新たな問題**: 再入室時にDBレベルで上書きが発生

**最終解決策**:

- **DB**: 実際のスロット番号（1〜4）を保持（空きスロットあり）
- **表示**: `GetRoomMembers` でメモリ上で詰める（空きスロットなし）

`internal/repository/room_repository.go` (567-571行目)

```go
// player_numberを再割り当て（空いたスロットを詰める）
// 注意: DBは更新せず、表示用にメモリ上でのみ詰める
for i := range members {
    members[i].PlayerNumber = i + 1
}
```

**重要ポイント**:
- `LeaveRoom` ではDBの `player_number` を更新しない
- `GetRoomMembers` の呼び出し時にのみメモリ上で詰める
- SSEでもAPI応答でも一貫した表示になる

### デバッグログの削除

Cloud Logging の負荷を軽減するため、全てのデバッグログを削除:

- `internal/handlers/rooms.go` の全デバッグログ
- `templates/components/room_detail_script.tmpl` の全 `console.log('[DEBUG] ...')`

## まとめ

SSEを活用して、部屋詳細画面のユーザーパネルをリアルタイム更新する機能を実装しました。イベントを適切に分離することで、無駄なリクエストを防ぎつつ、必要な箇所のみを効率的に更新できるようになりました。

実装後に発見された複数のバグを修正し、特に以下の点を改善しました：

1. **SSEイベント形式の統一**: `event: message` に統一し、JSONの `type` フィールドで区別
2. **システムメッセージの永続化**: `room_messages` テーブルに `message_type='system'` として保存
3. **モバイルUI**: `position: sticky` による入力フォームの固定表示
4. **スロット管理**: DBは実際のスロット番号を保持し、表示時のみ詰める

この実装により、ユーザーは入退室状況を即座に把握でき、デスクトップ・モバイル共に快適なユーザー体験を提供できるようになりました。
