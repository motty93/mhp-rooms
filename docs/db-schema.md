# データベーススキーマ設計

## 概要

モンスターハンターポータブルシリーズのアドホックパーティルーム管理システムのデータベース設計ドキュメントです。Fly.io PostgreSQLを使用し、GORMをORMとして採用しています。Supabase Authenticationと連携した認証システムを採用しています。

## テーブル一覧

### 1. users（ユーザー）

ユーザー情報を管理するマスターテーブル（Supabase Authと連携）

| カラム名         | データ型     | NULL | デフォルト        | 説明                                   |
| ---------------- | ------------ | ---- | ----------------- | -------------------------------------- |
| id               | UUID         | NO   | gen_random_uuid() | 主キー                                 |
| supabase_user_id | UUID         | NO   | -                 | Supabase AuthユーザーID（ユニーク）    |
| email            | VARCHAR(255) | NO   | -                 | メールアドレス（ユニーク）             |
| username         | VARCHAR(50)  | YES  | NULL              | ユーザー名（ユニーク、後から設定可能） |
| display_name     | VARCHAR(100) | NO   | -                 | 表示名                                 |
| avatar_url       | TEXT         | YES  | NULL              | アバター画像URL                        |
| bio              | TEXT         | YES  | NULL              | 自己紹介                               |
| is_active        | BOOLEAN      | NO   | true              | アクティブフラグ                       |
| role             | VARCHAR(20)  | NO   | 'user'            | ロール（user/admin）                   |
| created_at       | TIMESTAMP    | NO   | CURRENT_TIMESTAMP | 作成日時                               |
| updated_at       | TIMESTAMP    | NO   | CURRENT_TIMESTAMP | 更新日時                               |

**インデックス:**

- UNIQUE INDEX on supabase_user_id
- UNIQUE INDEX on email
- UNIQUE INDEX on username (NULL値を許可)
- INDEX on is_active, created_at

### 2. game_versions（ゲームバージョンマスター）

対応ゲームバージョンを管理するマスターテーブル

**特徴：**
- 固定マスターデータ（アプリケーション起動時に初期データとして投入）
- 運用中は基本的に変更されない
- ルームの分類とゲームバージョン識別に使用
- 表示順序とアクティブ/非アクティブの制御が可能

**主な用途：**
- ルーム作成時のゲームバージョン選択
- ルーム一覧でのフィルタリング
- ゲームバージョン別の表示順序制御

| カラム名      | データ型    | NULL | デフォルト        | 説明                                    |
| ------------- | ----------- | ---- | ----------------- | --------------------------------------- |
| id            | UUID        | NO   | gen_random_uuid() | 主キー                                  |
| code          | VARCHAR(10) | NO   | -                 | バージョンコード（MHP/MHP2/MHP2G/MHP3） |
| name          | VARCHAR(50) | NO   | -                 | ゲーム名                                |
| display_order | INTEGER     | NO   | -                 | 表示順                                  |
| is_active     | BOOLEAN     | NO   | true              | 有効フラグ                              |
| created_at    | TIMESTAMP   | NO   | CURRENT_TIMESTAMP | 作成日時                                |

**インデックス:**

- UNIQUE INDEX on code
- INDEX on is_active, display_order

### 3. rooms（ルーム）

アドホックパーティのルーム情報を管理

| カラム名         | データ型     | NULL | デフォルト        | 説明                                 |
| ---------------- | ------------ | ---- | ----------------- | ------------------------------------ |
| id               | UUID         | NO   | gen_random_uuid() | 主キー                               |
| room_code        | VARCHAR(20)  | NO   | -                 | ルームコード（ユニーク）             |
| name             | VARCHAR(100) | NO   | -                 | ルーム名                             |
| description      | TEXT         | YES  | NULL              | ルーム説明                           |
| game_version_id  | UUID         | NO   | -                 | ゲームバージョンID（外部キー）       |
| host_user_id     | UUID         | NO   | -                 | ホストユーザーID（外部キー）         |
| max_players      | INTEGER      | NO   | 4                 | 最大人数（固定で4人）                |
| current_players  | INTEGER      | NO   | 0                 | 現在の人数                           |
| password_hash    | VARCHAR(255) | YES  | NULL              | パスワード（NULL=公開）              |
| status           | VARCHAR(20)  | NO   | 'waiting'         | ステータス（waiting/playing/closed） |
| quest_type       | VARCHAR(50)  | YES  | NULL              | クエストタイプ                       |
| target_monster   | VARCHAR(100) | YES  | NULL              | ターゲットモンスター                 |
| rank_requirement | VARCHAR(20)  | YES  | NULL              | ランク制限                           |
| is_active        | BOOLEAN      | NO   | true              | アクティブフラグ                     |
| created_at       | TIMESTAMP    | NO   | CURRENT_TIMESTAMP | 作成日時                             |
| updated_at       | TIMESTAMP    | NO   | CURRENT_TIMESTAMP | 更新日時                             |
| closed_at        | TIMESTAMP    | YES  | NULL              | クローズ日時                         |

**インデックス:**

- UNIQUE INDEX on room_code
- INDEX on game_version_id, status, is_active
- INDEX on host_user_id
- INDEX on created_at DESC

### 4. room_members（ルームメンバー）

ルームの参加メンバーを管理

| カラム名      | データ型    | NULL | デフォルト        | 説明                             |
| ------------- | ----------- | ---- | ----------------- | -------------------------------- |
| id            | UUID        | NO   | gen_random_uuid() | 主キー                           |
| room_id       | UUID        | NO   | -                 | ルームID（外部キー）             |
| user_id       | UUID        | NO   | -                 | ユーザーID（外部キー）           |
| player_number | INTEGER     | NO   | -                 | プレイヤー番号（1-4）            |
| is_host       | BOOLEAN     | NO   | false             | ホストフラグ                     |
| status        | VARCHAR(20) | NO   | 'active'          | ステータス（active/kicked/left） |
| joined_at     | TIMESTAMP   | NO   | CURRENT_TIMESTAMP | 参加日時                         |
| left_at       | TIMESTAMP   | YES  | NULL              | 退出日時                         |

**インデックス:**

- UNIQUE INDEX on room_id, user_id, status='active'
- INDEX on user_id, status
- INDEX on room_id, player_number

### 5. room_messages（ルームメッセージ）

ルーム内のチャットメッセージを管理

| カラム名     | データ型    | NULL | デフォルト        | 説明                             |
| ------------ | ----------- | ---- | ----------------- | -------------------------------- |
| id           | UUID        | NO   | gen_random_uuid() | 主キー                           |
| room_id      | UUID        | NO   | -                 | ルームID（外部キー）             |
| user_id      | UUID        | NO   | -                 | ユーザーID（外部キー）           |
| message      | TEXT        | NO   | -                 | メッセージ内容                   |
| message_type | VARCHAR(20) | NO   | 'chat'            | タイプ（chat/system/join/leave） |
| is_deleted   | BOOLEAN     | NO   | false             | 削除フラグ                       |
| created_at   | TIMESTAMP   | NO   | CURRENT_TIMESTAMP | 作成日時                         |

**インデックス:**

- INDEX on room_id, created_at DESC
- INDEX on user_id

### 6. user_blocks（ユーザーブロック）

ユーザー間のブロック関係を管理

| カラム名        | データ型  | NULL | デフォルト        | 説明                                 |
| --------------- | --------- | ---- | ----------------- | ------------------------------------ |
| id              | UUID      | NO   | gen_random_uuid() | 主キー                               |
| blocker_user_id | UUID      | NO   | -                 | ブロックしたユーザーID（外部キー）   |
| blocked_user_id | UUID      | NO   | -                 | ブロックされたユーザーID（外部キー） |
| reason          | TEXT      | YES  | NULL              | ブロック理由                         |
| created_at      | TIMESTAMP | NO   | CURRENT_TIMESTAMP | 作成日時                             |

**インデックス:**

- UNIQUE INDEX on blocker_user_id, blocked_user_id
- INDEX on blocked_user_id

### 7. room_logs（ルームログ）

ルームの活動ログを記録

| カラム名   | データ型    | NULL | デフォルト        | 説明                                       |
| ---------- | ----------- | ---- | ----------------- | ------------------------------------------ |
| id         | UUID        | NO   | gen_random_uuid() | 主キー                                     |
| room_id    | UUID        | NO   | -                 | ルームID（外部キー）                       |
| user_id    | UUID        | YES  | NULL              | ユーザーID（外部キー）                     |
| action     | VARCHAR(50) | NO   | -                 | アクション（create/join/leave/kick/close） |
| details    | JSONB       | YES  | NULL              | 詳細情報                                   |
| created_at | TIMESTAMP   | NO   | CURRENT_TIMESTAMP | 作成日時                                   |

**インデックス:**

- INDEX on room_id, created_at DESC
- INDEX on user_id
- INDEX on action

## 制約とトリガー

### 外部キー制約

```sql
-- rooms
ALTER TABLE rooms
  ADD CONSTRAINT fk_rooms_game_version
  FOREIGN KEY (game_version_id) REFERENCES game_versions(id);

ALTER TABLE rooms
  ADD CONSTRAINT fk_rooms_host_user
  FOREIGN KEY (host_user_id) REFERENCES users(id);

-- room_members
ALTER TABLE room_members
  ADD CONSTRAINT fk_room_members_room
  FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE;

ALTER TABLE room_members
  ADD CONSTRAINT fk_room_members_user
  FOREIGN KEY (user_id) REFERENCES users(id);

-- room_messages
ALTER TABLE room_messages
  ADD CONSTRAINT fk_room_messages_room
  FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE;

ALTER TABLE room_messages
  ADD CONSTRAINT fk_room_messages_user
  FOREIGN KEY (user_id) REFERENCES users(id);

-- user_blocks
ALTER TABLE user_blocks
  ADD CONSTRAINT fk_user_blocks_blocker
  FOREIGN KEY (blocker_user_id) REFERENCES users(id);

ALTER TABLE user_blocks
  ADD CONSTRAINT fk_user_blocks_blocked
  FOREIGN KEY (blocked_user_id) REFERENCES users(id);

-- room_logs
ALTER TABLE room_logs
  ADD CONSTRAINT fk_room_logs_room
  FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE;

ALTER TABLE room_logs
  ADD CONSTRAINT fk_room_logs_user
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- users
ALTER TABLE users
  ADD CONSTRAINT chk_users_supabase_user_id
  CHECK (supabase_user_id IS NOT NULL);

-- rooms
ALTER TABLE rooms
  ADD CONSTRAINT chk_rooms_max_players
  CHECK (max_players = 4);
```

### トリガー

1. **updated_at自動更新トリガー**

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON rooms
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

2. **ルーム人数自動更新トリガー**

```sql
CREATE OR REPLACE FUNCTION update_room_player_count()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE rooms
  SET current_players = (
    SELECT COUNT(*)
    FROM room_members
    WHERE room_id = COALESCE(NEW.room_id, OLD.room_id)
    AND status = 'active'
  )
  WHERE id = COALESCE(NEW.room_id, OLD.room_id);
  RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_room_count_on_join
  AFTER INSERT OR UPDATE OR DELETE ON room_members
  FOR EACH ROW EXECUTE FUNCTION update_room_player_count();
```

## Supabase Authentication連携

### ユーザー作成フロー

1. Supabase Authでユーザー作成：
   - メールアドレス/パスワード認証
   - Google OAuth 2.0認証（Supabase Authの機能として）
2. Webhookまたは初回ログイン時にusersテーブルにレコード作成
3. supabase_user_idでSupabase Authと連携

### セッション管理

- Supabase Authの統一JWTトークンを使用（メール・Google共通）
- セッション管理は完全にSupabase Authに一任

## パフォーマンス考慮事項

### インデックス戦略

- 頻繁に検索条件となるカラムにインデックスを設定
- 複合インデックスは検索パターンに合わせて設計
- Fly.io PostgreSQLの自動VACUUM、ANALYZEを活用

### パーティショニング

- room_messagesとroom_logsは月単位でパーティショニングを検討
- 古いデータのアーカイブ戦略

### キャッシング

- アクティブなルーム情報のキャッシュ
- ユーザープロフィールのキャッシュ
- Supabase AuthのJWTトークン検証結果の短時間キャッシュ（メール・Google共通）

## 初期データ

### game_versionsテーブルの初期データ

アプリケーション起動時に以下のマスターデータを投入：

```sql
INSERT INTO game_versions (id, code, name, display_order, is_active) VALUES
  (gen_random_uuid(), 'MHP', 'モンスターハンターポータブル', 1, true),
  (gen_random_uuid(), 'MHP2', 'モンスターハンターポータブル 2nd', 2, true),
  (gen_random_uuid(), 'MHP2G', 'モンスターハンターポータブル 2nd G', 3, true),
  (gen_random_uuid(), 'MHP3', 'モンスターハンターポータブル 3rd', 4, true);
```

**注意事項：**
- codeカラムはアプリケーション内での識別子として使用
- display_orderでUI上での表示順序を制御
- is_activeで将来的なゲームバージョンの有効/無効切り替えに対応

## バックアップ戦略

- Fly.io PostgreSQLの自動バックアップ機能を活用
- 日次自動バックアップ
- ポイントインタイムリカバリのサポート
- クリティカルデータの定期エクスポート
