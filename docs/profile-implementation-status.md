# プロフィール機能 実装状況ドキュメント

## 概要

モンハンルーム管理アプリのプロフィール機能の実装状況と今後の開発計画をまとめたドキュメントです。

## 現在の実装状況

### ✅ 完全実装済み

#### データベース設計
- **usersテーブル拡張**：
  - `favorite_games` カラム（JSONB形式）- お気に入りゲーム保存
  - `play_times` カラム（JSONB形式）- プレイ時間帯保存
  - プラットフォームID関連フィールド追加：
    - `psn_online_id` (varchar(16)) - PlayStation Network ID
    - `nintendo_network_id` (varchar(16)) - Nintendo Network ID
    - `nintendo_switch_id` (varchar(20)) - Nintendo Switch フレンドコード
    - `pretendo_network_id` (varchar(16)) - Pretendo Network ID
    - `twitter_id` (varchar(15)) - Twitter/X アカウントID
- **user_followsテーブル**：
  - フォロー・フォロワー関係管理
  - ステータス管理（pending/accepted）

#### UI/UX実装
- プロフィール画面の基本レイアウト
- レスポンシブデザイン（モバイル・デスクトップ対応）
- タブ切り替え機能（htmx + Alpine.js）
- アバター表示機能
- フォロワー数表示
- お気に入りゲーム・プレイ時間帯の条件付き表示
- プロフィール編集ボタン

#### ルーティング
- `/profile` - 自分のプロフィール画面
- `/users/{uuid}` - 他ユーザーのプロフィール画面
- `/api/profile/edit-form` - プロフィール編集フォーム
- `/api/profile/activity` - アクティビティタブ
- `/api/profile/rooms` - 作成した部屋タブ
- `/api/profile/followers` - フォロワータブ
- `/api/profile/following` - フォロー中タブ
- `/api/users/{uuid}` - ユーザー情報API

### 🔄 モック実装（暫定実装）

以下の機能は現在、固定HTMLを返すモック実装となっています：

#### プロフィールタブコンテンツ
```go
// 現在の実装例
func (ph *ProfileHandler) Followers(w http.ResponseWriter, r *http.Request) {
    html := `
    <div>
        <h3 class="text-xl font-bold mb-4 text-gray-800">フォロワーリスト</h3>
        <!-- 固定のHTMLコンテンツ -->
    </div>`
    w.Write([]byte(html))
}
```

#### モック対象機能
- **アクティビティタブ**：固定の活動履歴を表示
- **作成した部屋タブ**：固定の部屋情報を表示（一部動的データあり）
- **フォロワータブ**：固定のフォロワーリストを表示
- **フォロー中タブ**：固定のフォロー中ユーザーリストを表示
- **プロフィール編集**：編集フォームは表示されるが機能は未実装

### ✅ 開発環境対応実装

#### 認証バイパス機能
- **開発環境での認証回避**：
  - `ENV != "production"` の場合、認証ミドルウェアを無効化
  - テストユーザー（`hunter1@example.com`）を自動取得
  - プロフィール画面への直接アクセス可能

#### ユーザー情報取得の改善
```go
// getUserFromContext関数の実装
func getUserFromContext(ctx context.Context) *models.User {
    // 認証ミドルウェアからユーザー情報を取得
    if user, ok := ctx.Value(middleware.DBUserContextKey).(*models.User); ok {
        return user
    }
    return nil
}
```

#### データ表示機能
- **JSONBフィールドの適切な処理**：
  - `GetFavoriteGames()` - お気に入りゲーム配列の取得
  - `GetPlayTimes()` - プレイ時間帯オブジェクトの取得
  - `SetFavoriteGames()` - お気に入りゲーム設定
  - `SetPlayTimes()` - プレイ時間帯設定

## 今後の実装予定

### 🚧 Phase 1: 動的データ表示

#### 1. 部分テンプレートの作成
```
templates/components/
├── profile_activity.tmpl     # アクティビティタブ用
├── profile_rooms.tmpl        # 作成した部屋タブ用
├── profile_followers.tmpl    # フォロワータブ用
└── profile_following.tmpl    # フォロー中タブ用
```

#### 2. リポジトリメソッドの完全実装
```go
// UserFollowRepository
func (r *UserFollowRepository) GetFollowers(userID uuid.UUID) ([]UserFollow, error)
func (r *UserFollowRepository) GetFollowing(userID uuid.UUID) ([]UserFollow, error)
func (r *UserFollowRepository) GetFollowerCount(userID uuid.UUID) (int64, error)

// UserRepository  
func (r *UserRepository) GetUserActivities(userID uuid.UUID) ([]Activity, error)
func (r *UserRepository) GetUserRooms(userID uuid.UUID) ([]Room, error)
```

#### 3. ハンドラーの動的実装
```go
// 動的実装例
func (ph *ProfileHandler) Followers(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context())
    followers, err := ph.repo.UserFollow.GetFollowers(userID)
    if err != nil {
        http.Error(w, "フォロワー情報の取得に失敗しました", http.StatusInternalServerError)
        return
    }
    
    data := struct {
        Followers []models.UserFollow `json:"followers"`
    }{Followers: followers}
    
    renderPartialTemplate(w, "profile_followers.tmpl", data)
}
```

### 🚧 Phase 2: インタラクション機能

#### フォロー・アンフォロー機能
- `POST /api/users/{uuid}/follow` - ユーザーをフォロー
- `DELETE /api/users/{uuid}/follow` - フォローを解除
- フォローボタンの状態管理
- リアルタイムフォロワー数更新

#### プロフィール編集機能
- お気に入りゲームの設定・編集
- プレイ時間帯の設定・編集
- 自己紹介の編集
- アバター画像のアップロード

### 🚧 Phase 3: 高度な機能

#### 通知機能
- フォロー通知
- 新規フォロワー通知

#### 検索・発見機能
- ユーザー検索機能
- おすすめユーザー表示

## 技術的な注意点

### 1. モック実装の識別
現在のモック実装は以下の特徴があります：
- ハンドラーで直接HTMLを文字列として定義
- 固定データを使用
- エラーハンドリングが簡素

### 2. 開発環境の特殊仕様
- **認証ミドルウェア無効化**：開発環境では`SUPABASE_JWT_SECRET`未設定時に認証をバイパス
- **テストユーザー自動取得**：`hunter1@example.com`のユーザーを自動で使用
- **JSONBデータ処理**：`favorite_games`と`play_times`の適切な読み書き処理を実装

### 3. JSONBフィールドの実装詳細
```go
// お気に入りゲーム設定例
user.SetFavoriteGames([]string{"MHP2G", "MHP3"})

// プレイ時間帯設定例
playTimes := &models.PlayTimes{
    Weekday: "19:00-23:00",
    Weekend: "13:00-24:00",
}
user.SetPlayTimes(playTimes)
```

### 4. 動的実装への移行手順
1. 部分テンプレートファイルの作成
2. リポジトリメソッドの実装
3. ハンドラーの書き換え（モック → 動的）
4. エラーハンドリングの追加
5. テストケースの追加

### 5. パフォーマンス考慮事項
- フォロワー数が多い場合のページネーション実装
- キャッシュ機能の検討
- N+1問題の回避（適切なJOIN使用）

## 実装優先度

1. **高優先度**：フォロー・アンフォロー機能の実装
2. **中優先度**：プロフィール編集機能の完全実装
3. **低優先度**：アクティビティタブの動的化
4. **将来対応**：通知機能・検索機能

## 最近の実装完了事項

### 2025年8月13日
- ✅ **認証ミドルウェアの開発環境対応**
- ✅ **プロフィール画面の表示確認**
- ✅ **getUserFromContext関数の実装**
- ✅ **JSONBフィールド処理の修正**
- ✅ **開発環境でのテストユーザー自動取得**

## 参考情報

- データベースマイグレーション: `scripts/migration_user_profile.sql`
- モデル定義: `internal/models/user.go`, `internal/models/user_follow.go`
- 現在のハンドラー実装: `internal/handlers/profile.go`
- テンプレート: `templates/pages/profile.tmpl`