# MHP Rooms API仕様書

## 概要

モンスターハンターポータブルシリーズ（MHP、MHP2、MHP2G、MHP3）のアドホックパーティルーム管理システムのAPI仕様書です。

## 基本情報

- **ベースURL**: `https://mhp-rooms.fly.dev`（本番環境）
- **認証方式**: JWT（Bearer Token）
- **データ形式**: JSON
- **文字エンコーディング**: UTF-8

## 認証

### 認証方式

本システムではSupabaseを使用したJWT認証を採用しています。保護されたエンドポイントへのアクセスには、`Authorization`ヘッダーにBearerトークンを含める必要があります。

```
Authorization: Bearer <jwt_token>
```

### 認証が必要なエンドポイント

- `/api/user/*`
- `/api/auth/sync`
- `/api/auth/psn-id`

## エンドポイント一覧

### 1. ページルート（HTML）

これらのエンドポイントはHTMLページを返します。APIとしての利用はできません。

| エンドポイント | メソッド | 説明 |
|--------------|---------|------|
| `/` | GET | ホームページ |
| `/terms` | GET | 利用規約 |
| `/privacy` | GET | プライバシーポリシー |
| `/contact` | GET/POST | お問い合わせ |
| `/faq` | GET | よくある質問 |
| `/guide` | GET | 利用ガイド |
| `/sitemap.xml` | GET | サイトマップ（XML形式） |

### 2. 認証関連

#### 2.1 ログインページ表示
```
GET /auth/login
```
ログインページのHTMLを返します。

#### 2.2 登録ページ表示
```
GET /auth/register
```
新規登録ページのHTMLを返します。

#### 2.3 ユーザー同期
```
POST /api/auth/sync
```
Supabase認証後、アプリケーションDBにユーザー情報を同期します。

**認証**: 必須

**リクエストボディ**:
```json
{
  "psn_id": "string (optional)"
}
```

**レスポンス**:
```json
{
  "id": "uuid",
  "supabase_user_id": "uuid",
  "email": "string",
  "username": "string",
  "display_name": "string",
  "avatar_url": "string",
  "bio": "string",
  "psn_online_id": "string",
  "twitter_id": "string",
  "is_active": true,
  "role": "user",
  "created_at": "2025-06-06T12:00:00Z",
  "updated_at": "2025-06-06T12:00:00Z"
}
```

#### 2.4 PSN ID更新
```
PUT /api/auth/psn-id
```
ユーザーのPSN IDを更新します。

**認証**: 必須

**リクエストボディ**:
```json
{
  "psn_id": "string"
}
```

**レスポンス**:
```json
{
  "message": "PSN IDが更新されました"
}
```

### 3. ルーム管理

#### 3.1 ルーム一覧取得（HTML）
```
GET /rooms?game_version={version_code}
```
ルーム一覧ページのHTMLを返します。

**クエリパラメータ**:
- `game_version`: ゲームバージョンでフィルタリング（オプション）

#### 3.2 ルーム一覧取得（API）
```
GET /api/rooms
```
すべてのアクティブなルーム情報をJSON形式で返します。

**認証**: オプション（認証済みの場合、ユーザー固有の情報を含む）

**レスポンス**:
```json
{
  "rooms": [
    {
      "id": "uuid",
      "room_code": "ROOM001",
      "name": "初心者歓迎！",
      "description": "初心者の方も気軽に参加してください",
      "game_version_id": "uuid",
      "host_user_id": "uuid",
      "max_players": 4,
      "current_players": 2,
      "quest_type": "キークエスト",
      "target_monster": "リオレウス",
      "rank_requirement": "HR4以上",
      "is_active": true,
      "is_closed": false,
      "created_at": "2025-06-06T12:00:00Z",
      "updated_at": "2025-06-06T12:00:00Z",
      "game_version": {
        "id": "uuid",
        "code": "MHP3",
        "name": "モンスターハンターポータブル 3rd",
        "display_order": 4,
        "platform_id": "uuid",
        "is_active": true
      },
      "host": {
        "id": "uuid",
        "display_name": "ハンター太郎",
        "psn_online_id": "hunter_taro"
      }
    }
  ],
  "total": 15
}
```

#### 3.3 ルーム作成
```
POST /rooms
```
新しいルームを作成します。

**認証**: 必須

**リクエストボディ**:
```json
{
  "name": "初心者歓迎！",
  "description": "初心者の方も気軽に参加してください",
  "game_version_id": "uuid",
  "max_players": 4,
  "password": "string (optional)",
  "quest_type": "キークエスト",
  "target_monster": "リオレウス",
  "rank_requirement": "HR4以上"
}
```

**レスポンス**:
```json
{
  "id": "uuid",
  "room_code": "ROOM001",
  "name": "初心者歓迎！",
  "description": "初心者の方も気軽に参加してください",
  "game_version_id": "uuid",
  "host_user_id": "uuid",
  "max_players": 4,
  "current_players": 1,
  "quest_type": "キークエスト",
  "target_monster": "リオレウス",
  "rank_requirement": "HR4以上",
  "is_active": true,
  "is_closed": false,
  "created_at": "2025-06-06T12:00:00Z",
  "updated_at": "2025-06-06T12:00:00Z"
}
```

#### 3.4 ルーム参加
```
POST /rooms/{id}/join
```
指定されたルームに参加します。

**認証**: 必須

**パスパラメータ**:
- `id`: ルームID（UUID）

**リクエストボディ**:
```json
{
  "password": "string (パスワード保護されている場合は必須)"
}
```

**レスポンス**:
```json
{
  "message": "ルームに参加しました",
  "room": {
    "id": "uuid",
    "room_code": "ROOM001",
    "name": "初心者歓迎！",
    "current_players": 2
  }
}
```

#### 3.5 ルーム退出
```
POST /rooms/{id}/leave
```
指定されたルームから退出します。

**認証**: 必須

**パスパラメータ**:
- `id`: ルームID（UUID）

**レスポンス**:
```json
{
  "message": "ルームから退出しました"
}
```

#### 3.6 ルーム開閉切り替え
```
PUT /rooms/{id}/toggle-closed
```
ルームの募集状態を切り替えます（ホストのみ実行可能）。

**認証**: 必須

**パスパラメータ**:
- `id`: ルームID（UUID）

**レスポンス**:
```json
{
  "message": "ルームの募集状態を更新しました",
  "is_closed": true
}
```

### 4. ユーザー情報

#### 4.1 現在のユーザー情報取得
```
GET /api/user/current
```
現在ログイン中のユーザー情報を取得します。

**認証**: 必須

**レスポンス**:
```json
{
  "id": "uuid",
  "supabase_user_id": "uuid",
  "email": "user@example.com",
  "username": "hunter_taro",
  "display_name": "ハンター太郎",
  "avatar_url": "https://example.com/avatar.png",
  "bio": "MHP3をメインでプレイしています",
  "psn_online_id": "hunter_taro_psn",
  "twitter_id": "hunter_taro",
  "is_active": true,
  "role": "user",
  "created_at": "2025-06-06T12:00:00Z",
  "updated_at": "2025-06-06T12:00:00Z"
}
```

### 5. 設定情報

#### 5.1 Supabase設定取得
```
GET /api/config/supabase
```
フロントエンドで使用するSupabaseの設定情報を取得します。

**認証**: 不要

**レスポンス**:
```json
{
  "url": "https://xxxxx.supabase.co",
  "anon_key": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

### 6. ヘルスチェック

#### 6.1 サービス稼働確認
```
GET /api/health
```
サービスの稼働状態を確認します。

**認証**: 不要

**レスポンス**:
```json
{
  "status": "ok",
  "timestamp": "2025-06-06T12:00:00Z"
}
```

## エラーレスポンス

すべてのAPIエンドポイントは、エラー発生時に以下の形式でレスポンスを返します：

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーの詳細メッセージ"
  }
}
```

### HTTPステータスコード

- `200 OK`: リクエスト成功
- `201 Created`: リソース作成成功
- `400 Bad Request`: リクエスト形式エラー
- `401 Unauthorized`: 認証エラー
- `403 Forbidden`: アクセス権限なし
- `404 Not Found`: リソースが見つからない
- `409 Conflict`: リソースの競合（例：すでに参加済みのルームへの参加）
- `500 Internal Server Error`: サーバーエラー
- `501 Not Implemented`: 未実装のエンドポイント

## データモデル

### User（ユーザー）
```typescript
interface User {
  id: string;                    // UUID
  supabase_user_id: string;      // UUID
  email: string;
  username?: string;
  display_name: string;
  avatar_url?: string;
  bio?: string;
  psn_online_id?: string;
  twitter_id?: string;
  is_active: boolean;
  role: string;                  // "user" | "admin"
  created_at: string;            // ISO 8601
  updated_at: string;            // ISO 8601
}
```

### Room（ルーム）
```typescript
interface Room {
  id: string;                    // UUID
  room_code: string;             // 一意のルームコード
  name: string;
  description?: string;
  game_version_id: string;       // UUID
  host_user_id: string;          // UUID
  max_players: number;
  current_players: number;
  password_hash?: string;        // パスワードハッシュ（レスポンスには含まれない）
  quest_type?: string;
  target_monster?: string;
  rank_requirement?: string;
  is_active: boolean;
  is_closed: boolean;
  created_at: string;            // ISO 8601
  updated_at: string;            // ISO 8601
  closed_at?: string;            // ISO 8601
  
  // リレーション
  game_version?: GameVersion;
  host?: User;
  members?: RoomMember[];
}
```

### GameVersion（ゲームバージョン）
```typescript
interface GameVersion {
  id: string;                    // UUID
  code: string;                  // "MHP" | "MHP2" | "MHP2G" | "MHP3"
  name: string;
  display_order: number;
  platform_id: string;           // UUID
  is_active: boolean;
  created_at: string;            // ISO 8601
  updated_at: string;            // ISO 8601
  
  // リレーション
  platform?: Platform;
}
```

### Platform（プラットフォーム）
```typescript
interface Platform {
  id: string;                    // UUID
  code: string;                  // "PSP" | "VITA"
  name: string;
  display_order: number;
  is_active: boolean;
  created_at: string;            // ISO 8601
  updated_at: string;            // ISO 8601
}
```

## 制限事項

- APIレート制限: 1分間に60リクエストまで（認証済みユーザー）
- ルーム作成制限: 1ユーザーあたり同時に5ルームまで
- ルーム最大人数: 4人
- ルーム非アクティブ期限: 最終更新から24時間

## 今後の実装予定

- WebSocket対応（リアルタイムルーム更新）
- ルーム内チャット機能
- ユーザーブロック機能
- ルーム検索・フィルタリング機能の拡充
- プレイヤー評価システム