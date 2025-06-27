# データベーススキーマ

## 概要
MonHubのデータベーススキーマ設計書です。PostgreSQL（開発環境）およびNeon（本番環境）を使用しています。

## テーブル定義

### users（ユーザー）
ユーザー認証と管理のためのテーブル。Supabase認証システムと連携。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| supabase_user_id | UUID | UNIQUE, NOT NULL | Supabase認証システムのユーザーID |
| email | VARCHAR(255) | UNIQUE, NOT NULL | メールアドレス |
| username | VARCHAR(50) | UNIQUE | ユーザー名（オプション） |
| display_name | VARCHAR(100) | NOT NULL | 表示名 |
| avatar_url | VARCHAR(255) | | アバター画像URL |
| bio | TEXT | | 自己紹介 |
| psn_online_id | VARCHAR(16) | | PlayStation NetworkオンラインID |
| twitter_id | VARCHAR(50) | | Twitter ID |
| is_active | BOOLEAN | NOT NULL, DEFAULT true | アクティブフラグ |
| role | VARCHAR(20) | NOT NULL, DEFAULT 'user' | ユーザー権限 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### game_versions（ゲームバージョン）
対応ゲームバージョンのマスターデータ。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| code | VARCHAR(10) | UNIQUE, NOT NULL | ゲームコード（MHP、MHP2、MHP2G、MHP3） |
| name | VARCHAR(100) | NOT NULL | ゲーム名 |
| short_name | VARCHAR(50) | NOT NULL | 略称 |
| display_order | INTEGER | NOT NULL | 表示順序 |
| is_active | BOOLEAN | NOT NULL, DEFAULT true | アクティブフラグ |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### rooms（ルーム）
ゲームルームの管理テーブル。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| room_code | VARCHAR(8) | UNIQUE, NOT NULL | ルームコード（8文字） |
| name | VARCHAR(100) | NOT NULL | ルーム名 |
| description | TEXT | | ルーム説明 |
| game_version_id | UUID | NOT NULL, FOREIGN KEY | ゲームバージョンID |
| host_user_id | UUID | NOT NULL, FOREIGN KEY | ホストユーザーID |
| max_players | INTEGER | NOT NULL, DEFAULT 4, CHECK (1-4) | 最大人数 |
| password_hash | VARCHAR(255) | | パスワードハッシュ |
| quest_type | VARCHAR(50) | NOT NULL | クエストタイプ |
| target_monster | VARCHAR(100) | | ターゲットモンスター |
| rank_requirement | VARCHAR(20) | | ランク条件 |
| is_active | BOOLEAN | NOT NULL, DEFAULT true | アクティブフラグ |
| is_closed | BOOLEAN | NOT NULL, DEFAULT false | クローズフラグ |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### room_members（ルームメンバー）
ルーム参加者の管理テーブル。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| room_id | UUID | NOT NULL, FOREIGN KEY | ルームID |
| user_id | UUID | NOT NULL, FOREIGN KEY | ユーザーID |
| player_number | INTEGER | NOT NULL, CHECK (1-4) | プレイヤー番号 |
| is_host | BOOLEAN | NOT NULL, DEFAULT false | ホストフラグ |
| joined_at | TIMESTAMP | NOT NULL | 参加日時 |
| left_at | TIMESTAMP | | 退出日時 |

### room_messages（ルームメッセージ）
ルーム内チャットメッセージ。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| room_id | UUID | NOT NULL, FOREIGN KEY | ルームID |
| user_id | UUID | NOT NULL, FOREIGN KEY | ユーザーID |
| message_type | VARCHAR(20) | NOT NULL, DEFAULT 'chat' | メッセージタイプ |
| content | TEXT | NOT NULL | メッセージ内容 |
| is_deleted | BOOLEAN | NOT NULL, DEFAULT false | 削除フラグ |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### room_logs（ルームログ）
ルームアクションの監査ログ。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| room_id | UUID | NOT NULL, FOREIGN KEY | ルームID |
| user_id | UUID | FOREIGN KEY | ユーザーID（システムアクションの場合NULL） |
| action_type | VARCHAR(50) | NOT NULL | アクションタイプ |
| details | JSONB | | 詳細情報 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### user_blocks（ユーザーブロック）
ユーザー間のブロック関係。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| blocker_user_id | UUID | NOT NULL, FOREIGN KEY | ブロックしたユーザーID |
| blocked_user_id | UUID | NOT NULL, FOREIGN KEY | ブロックされたユーザーID |
| reason | VARCHAR(255) | | ブロック理由 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |

### player_names（プレイヤー名）
ゲームバージョンごとのプレイヤー名管理。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| user_id | UUID | NOT NULL, FOREIGN KEY | ユーザーID |
| game_version_id | UUID | NOT NULL, FOREIGN KEY | ゲームバージョンID |
| name | VARCHAR(50) | NOT NULL | プレイヤー名 |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

### password_resets（パスワードリセット）
パスワードリセット用トークン管理。

| カラム名 | 型 | 制約 | 説明 |
|---------|-----|------|------|
| id | UUID | PRIMARY KEY | 主キー |
| user_id | UUID | NOT NULL, FOREIGN KEY | ユーザーID |
| token | VARCHAR(255) | UNIQUE, NOT NULL | リセットトークン |
| expires_at | TIMESTAMP | NOT NULL | 有効期限 |
| used | BOOLEAN | NOT NULL, DEFAULT false | 使用済みフラグ |
| created_at | TIMESTAMP | NOT NULL | 作成日時 |
| updated_at | TIMESTAMP | NOT NULL | 更新日時 |

## 初期データ

### game_versions
```sql
INSERT INTO game_versions (code, name, short_name, display_order) VALUES
('MHP', 'モンスターハンターポータブル', 'MHP', 1),
('MHP2', 'モンスターハンターポータブル 2nd', 'MHP2', 2),
('MHP2G', 'モンスターハンターポータブル 2nd G', 'MHP2G', 3),
('MHP3', 'モンスターハンターポータブル 3rd', 'MHP3', 4);
```

## 制約とトリガー

### ユニーク制約
- `users`: email, username, supabase_user_id
- `game_versions`: code
- `rooms`: room_code
- `room_members`: (room_id, user_id) の組み合わせ
- `user_blocks`: (blocker_user_id, blocked_user_id) の組み合わせ
- `player_names`: (user_id, game_version_id) の組み合わせ
- `password_resets`: token

### 外部キー制約
すべての外部キーは `ON DELETE CASCADE` で設定され、親レコードの削除時に関連レコードも削除されます。

### チェック制約
- `rooms.max_players`: 1以上4以下
- `room_members.player_number`: 1以上4以下

## マイグレーション

GORMのAutoMigrate機能を使用して、モデル定義からテーブルを自動生成します。

```go
// cmd/migrate/main.go で実行
db.AutoMigrate(
    &models.User{},
    &models.GameVersion{},
    &models.Room{},
    &models.RoomMember{},
    &models.RoomMessage{},
    &models.RoomLog{},
    &models.UserBlock{},
    &models.PlayerName{},
    &models.PasswordReset{},
)
```

## 今後の拡張予定

### マルチプラットフォーム対応
- `game_versions`テーブルに`platform`カラムを追加
- 3DS、Wii U版のモンスターハンターシリーズに対応
- プラットフォーム別の設定を管理する`platform_configs`テーブルの追加