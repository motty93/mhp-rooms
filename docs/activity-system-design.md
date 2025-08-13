# アクティビティシステム設計書

## 概要

ユーザーの行動履歴を記録・表示するアクティビティシステムの設計書です。
プロフィール画面での「最近の活動履歴」タブで、ユーザーの様々な活動を時系列で表示します。

## 現状の問題点

### 現在の実装
- `getMockActivities()`で固定のモックデータを返している
- 実際のユーザー行動は記録されていない
- データベースから動的に取得していない

### 既存DB設計での課題
1. **データの散在**: アクティビティ情報が複数テーブルに分散
   - 部屋作成: `rooms`テーブル
   - 部屋参加: `room_members`テーブル  
   - フォロー: `user_follows`テーブル
   - メッセージ: `room_messages`テーブル

2. **パフォーマンス問題**: 毎回複数テーブルをJOINして時系列ソートが必要

3. **データ永続性の問題**: 関連エンティティが削除されるとアクティビティ履歴も消失

4. **詳細情報の不足**: アクティビティの文脈や詳細が記録されない

## 設計方針

### 基本方針
1. **専用テーブルでの一元管理**: アクティビティ専用テーブルを新設
2. **非正規化アプローチ**: パフォーマンス重視で必要な情報をアクティビティテーブルに保存
3. **イベントドリブン記録**: 各種アクションの実行時にアクティビティを自動記録
4. **柔軟なメタデータ**: JSONBフィールドで拡張可能な詳細情報を保存

### 技術方針
- **記録方式**: 同期的にアクティビティを記録（将来的には非同期化も検討）
- **保存期間**: 基本的に永続保存（将来的にアーカイブ機能を検討）
- **パフォーマンス**: インデックス最適化とページネーション対応

## データベース設計

### 新規テーブル: `user_activities`

```sql
CREATE TABLE user_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    activity_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    related_entity_type VARCHAR(50), -- 'room', 'user', 'message'など
    related_entity_id UUID,
    metadata JSONB DEFAULT '{}',
    icon VARCHAR(100), -- Font Awesomeアイコンクラス
    icon_color VARCHAR(50), -- Tailwind CSSカラークラス
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_user_activities_user_id ON user_activities(user_id);
CREATE INDEX idx_user_activities_type ON user_activities(activity_type);
CREATE INDEX idx_user_activities_created_at ON user_activities(created_at DESC);
CREATE INDEX idx_user_activities_user_created ON user_activities(user_id, created_at DESC);
```

### アクティビティタイプ定義

```go
const (
    // 部屋関連
    ActivityRoomCreate    = "room_create"    // 部屋作成
    ActivityRoomJoin      = "room_join"      // 部屋参加
    ActivityRoomLeave     = "room_leave"     // 部屋退出
    ActivityRoomClose     = "room_close"     // 部屋終了
    ActivityRoomUpdate    = "room_update"    // 部屋設定変更
    
    // フォロー関連
    ActivityFollowAdd     = "follow_add"     // フォロー開始
    ActivityFollowAccept  = "follow_accept"  // フォロー承認
    ActivityFollowRemove  = "follow_remove"  // フォロー解除
    
    // メッセージ関連
    ActivityMessageSend   = "message_send"   // メッセージ送信
    
    // プロフィール関連
    ActivityProfileUpdate = "profile_update" // プロフィール更新
    
    // システム関連
    ActivityUserJoin      = "user_join"      // ユーザー登録
)
```

## 実装設計

### 1. データモデル

```go
// models/user_activity.go
type UserActivity struct {
    ID                uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    UserID            uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
    ActivityType      string     `json:"activity_type" gorm:"size:50;not null;index"`
    Title             string     `json:"title" gorm:"size:255;not null"`
    Description       *string    `json:"description"`
    RelatedEntityType *string    `json:"related_entity_type" gorm:"size:50"`
    RelatedEntityID   *uuid.UUID `json:"related_entity_id"`
    Metadata          datatypes.JSON `json:"metadata" gorm:"type:jsonb;default:'{}'"`
    Icon              string     `json:"icon" gorm:"size:100"`
    IconColor         string     `json:"icon_color" gorm:"size:50"`
    CreatedAt         time.Time  `json:"created_at" gorm:"index"`
    UpdatedAt         time.Time  `json:"updated_at"`
    
    // リレーション
    User User `json:"user" gorm:"foreignKey:UserID"`
}
```

### 2. リポジトリインターface

```go
// repository/interfaces.go
type UserActivityRepository interface {
    CreateActivity(activity *models.UserActivity) error
    GetUserActivities(userID uuid.UUID, limit, offset int) ([]models.UserActivity, error)
    GetUserActivitiesByType(userID uuid.UUID, activityType string, limit, offset int) ([]models.UserActivity, error)
    CountUserActivities(userID uuid.UUID) (int64, error)
    DeleteOldActivities(olderThan time.Time) error
}
```

### 3. アクティビティ記録サービス

```go
// services/activity_service.go
type ActivityService struct {
    repo *repository.Repository
}

func (s *ActivityService) RecordRoomCreate(userID uuid.UUID, room *models.Room) error {
    activity := &models.UserActivity{
        UserID:            userID,
        ActivityType:      ActivityRoomCreate,
        Title:             "【部屋作成】" + room.Name,
        Description:       buildRoomDescription(room),
        RelatedEntityType: strPtr("room"),
        RelatedEntityID:   &room.ID,
        Icon:              "fa-door-open",
        IconColor:         "text-green-500",
        Metadata: map[string]interface{}{
            "game_version": room.GameVersion.Code,
            "max_players":  room.MaxPlayers,
            "target_monster": room.TargetMonster,
        },
    }
    
    return s.repo.UserActivity.CreateActivity(activity)
}

func (s *ActivityService) RecordRoomJoin(userID uuid.UUID, room *models.Room, hostUser *models.User) error {
    activity := &models.UserActivity{
        UserID:            userID,
        ActivityType:      ActivityRoomJoin,
        Title:             "【部屋参加】" + room.Name,
        Description:       "ホスト: " + hostUser.DisplayName,
        RelatedEntityType: strPtr("room"),
        RelatedEntityID:   &room.ID,
        Icon:              "fa-right-to-bracket",
        IconColor:         "text-blue-500",
        Metadata: map[string]interface{}{
            "game_version": room.GameVersion.Code,
            "host_user_id": room.HostUserID,
        },
    }
    
    return s.repo.UserActivity.CreateActivity(activity)
}

func (s *ActivityService) RecordFollow(followerID, followingID uuid.UUID, followingUser *models.User) error {
    activity := &models.UserActivity{
        UserID:            followerID,
        ActivityType:      ActivityFollowAdd,
        Title:             followingUser.DisplayName + "さんをフォローしました",
        Description:       nil,
        RelatedEntityType: strPtr("user"),
        RelatedEntityID:   &followingID,
        Icon:              "fa-user-plus",
        IconColor:         "text-yellow-500",
        Metadata: map[string]interface{}{
            "following_user_id": followingID,
        },
    }
    
    return s.repo.UserActivity.CreateActivity(activity)
}
```

### 4. ハンドラーでの利用

```go
// handlers/profile.go
func (ph *ProfileHandler) Activity(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromContext(r.Context()) // URLパラメータまたは認証ユーザー
    
    activities, err := ph.repo.UserActivity.GetUserActivities(userID, 20, 0)
    if err != nil {
        ph.logger.Printf("アクティビティ取得エラー: %v", err)
        // フォールバック: 空の配列を返す
        activities = []models.UserActivity{}
    }
    
    // models.UserActivityをActivity構造体に変換
    displayActivities := make([]Activity, len(activities))
    for i, activity := range activities {
        displayActivities[i] = Activity{
            Type:        activity.ActivityType,
            Title:       activity.Title,
            Description: getStringValue(activity.Description),
            TimeAgo:     formatRelativeTime(activity.CreatedAt),
            Icon:        activity.Icon,
            IconColor:   activity.IconColor,
        }
    }
    
    data := struct {
        Activities []Activity
    }{
        Activities: displayActivities,
    }
    
    if err := renderPartialTemplate(w, "profile_activity.tmpl", data); err != nil {
        ph.logger.Printf("テンプレートレンダリングエラー: %v", err)
        http.Error(w, "テンプレートの描画に失敗しました", http.StatusInternalServerError)
        return
    }
}
```

## 実装手順

### Phase 1: DB準備
1. `user_activities`テーブルのマイグレーション作成
2. `UserActivity`モデルの実装
3. `UserActivityRepository`の実装

### Phase 2: 基本機能
1. `ActivityService`の実装
2. 部屋作成・参加時のアクティビティ記録
3. プロフィール画面での表示機能

### Phase 3: 機能拡張
1. フォロー関連アクティビティの記録
2. その他アクションのアクティビティ化
3. アクティビティのフィルタリング機能

### Phase 4: 最適化
1. パフォーマンス最適化
2. 古いアクティビティのアーカイブ機能
3. 非同期記録への移行検討

## パフォーマンス考慮事項

### インデックス戦略
- `(user_id, created_at DESC)`: ユーザー別時系列取得用
- `activity_type`: タイプ別フィルタリング用
- `created_at DESC`: 全体の時系列ソート用

### ページネーション
- 基本的に20件ずつ表示
- `LIMIT`と`OFFSET`を使用した実装
- 将来的にはカーソルベースページネーションも検討

### キャッシュ戦略
- 頻繁にアクセスされるユーザーのアクティビティはRedisにキャッシュ
- アクティビティ追加時にキャッシュを無効化

## セキュリティ・プライバシー考慮事項

### データ保護
- ユーザーのプライバシー設定に応じてアクティビティの表示制御
- 機密情報はメタデータに含めない

### アクセス制御
- 自分のアクティビティのみ表示可能
- 管理者は全ユーザーのアクティビティを参照可能

## 今後の拡張可能性

### 通知との連携
- フォローユーザーのアクティビティを通知
- 興味のある部屋作成の通知

### ソーシャル機能
- アクティビティへのいいね・コメント
- アクティビティの共有機能

### 分析機能
- ユーザー行動の分析
- 人気アクティビティの把握

## 関連ファイル

- `docs/er.md` - 現在のデータベース設計
- `internal/handlers/profile.go` - プロフィール関連ハンドラー  
- `templates/components/profile_activity.tmpl` - アクティビティ表示テンプレート

## 更新履歴

- 2025-08-13: 初版作成