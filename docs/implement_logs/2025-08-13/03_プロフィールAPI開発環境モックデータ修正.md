# プロフィールAPI開発環境モックデータ修正

**実装時間**: 約45分 (16:00-16:45)

## 概要
プロフィール画面で「作成した部屋」をクリックした際、開発環境でハードコーディングされたモックデータが表示される問題を修正しました。

## 問題の詳細

### 発見された問題
- `internal/handlers/profile.go`の`GetUserProfile`メソッド（API用）でモックデータ（`getMockRooms()`）を使用していた
- このため、API経由でプロフィール情報を取得すると、実際のユーザーに関係なく固定の「テスト部屋（更新済み）」「古龍種連戦」が表示されていた
- プロフィールページの「作成した部屋」タブ自体は正しく実装されていたが、APIレスポンスでは誤った情報が返されていた

### 問題のあったコード箇所
```go
// internal/handlers/profile.go:217行目（修正前）
profileData := ProfileData{
    User:         user,
    IsOwnProfile: isOwnProfile,
    Activities:   ph.getMockActivities(),
    Rooms:        ph.getMockRooms(),    // ←この部分が問題
    Followers:    ph.getMockFollowers(),
}
```

## 修正内容

### 1. GetUserProfileメソッドの修正
`getMockRooms()`を実際のデータベースクエリに変更：

```go
// 修正後
// お気に入りゲームとプレイ時間帯を取得
favoriteGames, _ := user.GetFavoriteGames()
playTimes, _ := user.GetPlayTimes()

// フォロワー数を取得
var followerCount int64 = 0
if ph.repo != nil && ph.repo.UserFollow != nil {
    followers, err := ph.repo.UserFollow.GetFollowers(user.ID)
    if err == nil {
        followerCount = int64(len(followers))
    }
}

// 実際に作成した部屋を取得
rooms, err := ph.repo.Room.GetRoomsByHostUser(user.ID, 10, 0) // 最大10件取得
var roomSummaries []RoomSummary
if err == nil {
    for _, room := range rooms {
        roomSummaries = append(roomSummaries, roomToSummary(room))
    }
}

profileData := ProfileData{
    User:          user,
    IsOwnProfile:  isOwnProfile,
    Activities:    ph.getMockActivities(),
    Rooms:         roomSummaries,  // ←実際のデータに修正
    Followers:     ph.getMockFollowers(),
    FollowerCount: followerCount,
    FavoriteGames: favoriteGames,
    PlayTimes:     playTimes,
}
```

### 2. 既存の正しい実装の確認
- プロフィールページの初期表示（`Profile`メソッド）は既に正しく実装されていた
- プロフィール画面の「作成した部屋」タブ（`Rooms`メソッド）も正しく実装されていた
- 問題があったのは`GetUserProfile` APIメソッドのみだった

## 技術的な詳細

### データフロー
1. **ユーザーがプロフィール画面の「作成した部屋」タブをクリック**
2. **htmxがAPIリクエストを送信**（`/api/users/{uuid}`または`/api/profile/rooms`）
3. **サーバーがユーザーIDを特定**
4. **`GetRoomsByHostUser`でデータベースから実際の部屋を取得**
5. **`roomToSummary`で表示用データに変換**
6. **JSONレスポンスまたはHTMLテンプレートで返却**

### 使用されるデータベースクエリ
```sql
SELECT
    rooms.*,
    gv.name as game_version_name,
    gv.code as game_version_code,
    u.display_name as host_display_name,
    u.psn_online_id as host_psn_online_id,
    COUNT(DISTINCT rm.id) as current_players
FROM rooms
LEFT JOIN game_versions gv ON rooms.game_version_id = gv.id
LEFT JOIN users u ON rooms.host_user_id = u.id
LEFT JOIN room_members rm ON rooms.id = rm.room_id AND rm.status = 'active'
WHERE rooms.host_user_id = ?
GROUP BY rooms.id, gv.id, u.id
ORDER BY rooms.created_at DESC
LIMIT ? OFFSET ?
```

## 動作確認結果

### テスト1: 1番目のテストユーザー（ハンター太郎）
```bash
curl -X GET "http://localhost:8080/api/users/d56f582b-4490-4399-b698-29e57d3b208c"
```
**結果**: 8つの実際の部屋が表示（以前は2つの固定モックデータ）

### テスト2: 2番目のテストユーザー（猫好きハンター）  
```bash
curl -X GET "http://localhost:8080/api/users/9c40956c-88b2-4de6-a441-c8028b8566d7"
```
**結果**: 4つの実際の部屋が表示

### テスト3: プロフィールタブのHTMLレンダリング
```bash
curl -X GET "http://localhost:8080/api/users/d56f582b-4490-4399-b698-29e57d3b208c/rooms"
```
**結果**: HTMLテンプレートで8つの部屋が正しく表示

## 注意した点

### 開発環境と本番環境の統一
- CLAUDE.mdに記載されている「環境別実装の禁止」方針に従って修正
- 開発環境でもハードコーディングではなく、実際のデータベースから取得するように統一

### エラーハンドリング
- データベースクエリが失敗した場合でも空配列を返すことで画面が壊れないように対応
- フォロワー数取得でも同様のエラーハンドリングを追加

## 影響範囲

### 修正対象
- `internal/handlers/profile.go`の`GetUserProfile`メソッドのみ

### 影響を受けない部分  
- プロフィールページの初期表示機能
- プロフィールタブのhtmxによる動的読み込み機能
- 他のプロフィール関連API

## 今後の改善点

### まだモック実装が残っている部分
- Activities（アクティビティ）: `getMockActivities()`
- Followers（フォロワー）: `getMockFollowers()`

これらも将来的には実際のデータベースから取得する実装に変更する予定。

## まとめ
開発環境でハードコーディングされたモックデータが表示される問題を修正し、実際のデータベースから動的に部屋情報を取得するように変更しました。これにより、開発環境でも本番環境と同じ動作をするようになり、プロジェクトの方針に沿った実装となりました。