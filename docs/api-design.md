# API設計仕様書

## 概要

モンスターハンターポータブルシリーズのアドホックパーティルーム管理システムのAPI設計ドキュメントです。RESTful APIとWebSocket通信の両方の仕様を定義します。

## API基本仕様

### ベースURL

```
開発環境: http://localhost:8080
本番環境: https://mhp-rooms.fly.dev
```

### 認証方式

- Supabase Authenticationによる認証管理
- JWT (JSON Web Token) によるBearer認証
- ヘッダー形式: `Authorization: Bearer <token>`
- サポートする認証方法:
  - メールアドレス/パスワード認証（Supabase Auth）
  - Google OAuth 2.0（Supabase Auth経由）

### アクセス制限

- トップページ（ルーム一覧）: 認証不要
- ルーム詳細・参加・その他全ての機能: 認証必要

### 共通レスポンス形式

#### 成功時

```json
{
  "success": true,
  "data": {
    // レスポンスデータ
  },
  "meta": {
    "timestamp": "2024-01-20T12:00:00Z",
    "request_id": "uuid-v4"
  }
}
```

#### エラー時

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーメッセージ",
    "details": {
      // 詳細なエラー情報（オプション）
    }
  },
  "meta": {
    "timestamp": "2024-01-20T12:00:00Z",
    "request_id": "uuid-v4"
  }
}
```

### HTTPステータスコード

- `200 OK`: 成功
- `201 Created`: リソース作成成功
- `204 No Content`: 成功（レスポンスボディなし）
- `400 Bad Request`: リクエスト不正
- `401 Unauthorized`: 認証エラー
- `403 Forbidden`: 権限エラー
- `404 Not Found`: リソースが見つからない
- `409 Conflict`: 競合エラー
- `429 Too Many Requests`: レート制限
- `500 Internal Server Error`: サーバーエラー

## RESTful API エンドポイント

### 認証関連

#### POST /auth/register

メールアドレスによるユーザー登録（Supabase Auth使用）

**リクエスト:**

```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "display_name": "ハンター太郎"
}
```

**注意:**

- Supabase Authが自動的にメール確認を送信
- username はユーザー作成後に別途設定

**レスポンス:**

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "hunter123",
      "display_name": "ハンター太郎"
    },
    "tokens": {
      "access_token": "jwt-access-token",
      "refresh_token": "jwt-refresh-token",
      "expires_in": 3600
    }
  }
}
```

#### POST /auth/login

メールアドレスによるログイン（Supabase Auth使用）

**リクエスト:**

```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**レスポンス:** 登録と同じ形式

#### GET /auth/google

Google OAuth認証開始（Supabase Auth使用）

**説明:**

- Supabase AuthのGoogle OAuth機能を使用してGoogle認証ページへリダイレクト
- 認証後、Supabaseが指定したコールバックURLへ自動リダイレクト

#### GET /auth/google/callback

Google OAuth認証コールバック（Supabase Auth使用）

**説明:**

- Supabase Authからのコールバックを処理
- JWTトークンを取得してアプリケーションセッションを開始

**クエリパラメータ:**

- Supabase Authが管理するパラメータ（access_token、refresh_token等）

**レスポンス:**

- 成功時: SupabaseのJWTトークンを含むレスポンスと共にアプリケーションへリダイレクト
- 失敗時: エラーページへリダイレクト

#### POST /auth/logout

ログアウト（Supabase Auth使用）

**リクエスト:** なし（認証ヘッダーは必要）

**レスポンス:**

```json
{
  "success": true,
  "data": {
    "message": "ログアウトしました"
  }
}
```

#### POST /auth/refresh

トークンリフレッシュ（Supabase Auth使用）

**リクエスト:**

```json
{
  "refresh_token": "jwt-refresh-token"
}
```

**レスポンス:** ログインと同じ形式

### ユーザー関連

#### GET /users/me

自分のプロフィール取得

**レスポンス:**

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "hunter123",
    "display_name": "ハンター太郎",
    "avatar_url": "https://example.com/avatar.jpg",
    "bio": "MHP2Gメインでプレイしています",
    "role": "user",
    "created_at": "2024-01-20T12:00:00Z"
  }
}
```

#### PATCH /users/me

プロフィール更新

**リクエスト:**

```json
{
  "display_name": "新しい名前",
  "bio": "更新されたプロフィール",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

#### GET /users/:id

ユーザー情報取得

#### POST /users/:id/block

ユーザーをブロック

#### DELETE /users/:id/block

ブロック解除

#### GET /users/blocks

ブロックリスト取得

### ゲームバージョン関連

#### GET /game-versions

ゲームバージョン一覧取得

**レスポンス:**

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "code": "MHP",
      "name": "モンスターハンターポータブル",
      "display_order": 1
    },
    {
      "id": "uuid",
      "code": "MHP2",
      "name": "モンスターハンターポータブル 2nd",
      "display_order": 2
    }
  ]
}
```

### ルーム関連

#### GET /rooms

ルーム一覧取得（認証不要）

**説明:** トップページで表示される公開情報。ルームの入室人数やゲーム情報を含む。

**クエリパラメータ:**

- `game_version_id`: ゲームバージョンでフィルタ
- `status`: ステータスでフィルタ（waiting/playing）
- `search`: ルーム名で検索
- `page`: ページ番号（デフォルト: 1）
- `limit`: 件数（デフォルト: 20、最大: 100）

**レスポンス:**

```json
{
  "success": true,
  "data": {
    "rooms": [
      {
        "id": "uuid",
        "room_code": "ROOM-ABC123",
        "name": "上位キリン討伐",
        "description": "装備自由、腕に自信ある方",
        "game_version": {
          "id": "uuid",
          "code": "MHP2G",
          "name": "モンスターハンターポータブル 2nd G"
        },
        "host": {
          "id": "uuid",
          "username": "hunter123",
          "display_name": "ハンター太郎"
        },
        "current_players": 2,
        "max_players": 4,
        "has_password": true,
        "status": "waiting",
        "quest_type": "討伐",
        "target_monster": "キリン",
        "rank_requirement": "HR6以上",
        "created_at": "2024-01-20T12:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_count": 95,
      "per_page": 20
    }
  }
}
```

#### POST /rooms

ルーム作成（認証必要）

**リクエスト:**

```json
{
  "name": "上位キリン討伐",
  "description": "装備自由、腕に自信ある方",
  "game_version_id": "uuid",
  "max_players": 4,
  "password": "optional-password",
  "quest_type": "討伐",
  "target_monster": "キリン",
  "rank_requirement": "HR6以上"
}
```

#### GET /rooms/:id

ルーム詳細取得（認証必要）

**レスポンス:**

```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "room_code": "ROOM-ABC123",
    "name": "上位キリン討伐",
    "description": "装備自由、腕に自信ある方",
    "game_version": {
      "id": "uuid",
      "code": "MHP2G",
      "name": "モンスターハンターポータブル 2nd G"
    },
    "host": {
      "id": "uuid",
      "username": "hunter123",
      "display_name": "ハンター太郎"
    },
    "members": [
      {
        "id": "uuid",
        "user": {
          "id": "uuid",
          "username": "hunter123",
          "display_name": "ハンター太郎"
        },
        "player_number": 1,
        "is_host": true,
        "joined_at": "2024-01-20T12:00:00Z"
      },
      {
        "id": "uuid",
        "user": {
          "id": "uuid",
          "username": "hunter456",
          "display_name": "ハンター花子"
        },
        "player_number": 2,
        "is_host": false,
        "joined_at": "2024-01-20T12:05:00Z"
      }
    ],
    "current_players": 2,
    "max_players": 4,
    "has_password": true,
    "status": "waiting",
    "quest_type": "討伐",
    "target_monster": "キリン",
    "rank_requirement": "HR6以上",
    "created_at": "2024-01-20T12:00:00Z",
    "updated_at": "2024-01-20T12:05:00Z"
  }
}
```

#### PATCH /rooms/:id

ルーム更新（ホストのみ）

**リクエスト:**

```json
{
  "name": "新しいルーム名",
  "description": "新しい説明",
  "status": "playing"
}
```

#### DELETE /rooms/:id

ルーム削除（ホストのみ）

#### POST /rooms/:id/join

ルーム参加（認証必要）

**制約:** 最大4人まで

**リクエスト:**

```json
{
  "password": "room-password" // パスワード付きルームの場合
}
```

#### POST /rooms/:id/leave

ルーム退出

#### POST /rooms/:id/kick/:user_id

メンバーをキック（ホストのみ）

### ルームメッセージ関連

#### GET /rooms/:id/messages

メッセージ一覧取得

**クエリパラメータ:**

- `limit`: 件数（デフォルト: 50）
- `before`: このタイムスタンプより前のメッセージ
- `after`: このタイムスタンプより後のメッセージ

**レスポンス:**

```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": "uuid",
        "user": {
          "id": "uuid",
          "username": "hunter123",
          "display_name": "ハンター太郎"
        },
        "message": "よろしくお願いします！",
        "message_type": "chat",
        "created_at": "2024-01-20T12:10:00Z"
      },
      {
        "id": "uuid",
        "message": "hunter456 が参加しました",
        "message_type": "join",
        "created_at": "2024-01-20T12:11:00Z"
      }
    ]
  }
}
```

#### POST /rooms/:id/messages

メッセージ送信

**リクエスト:**

```json
{
  "message": "よろしくお願いします！"
}
```

## WebSocket API

### 接続エンドポイント

```
ws://localhost:8080/socket.io
wss://mhp-rooms.fly.dev/socket.io
```

### 認証

接続時にクエリパラメータでトークンを送信:

```
ws://localhost:8080/socket.io?token=jwt-access-token
```

### イベント一覧

#### クライアント → サーバー

##### join_room

ルームに参加

```json
{
  "room_id": "uuid"
}
```

##### leave_room

ルームから退出

```json
{
  "room_id": "uuid"
}
```

##### send_message

メッセージ送信

```json
{
  "room_id": "uuid",
  "message": "メッセージ内容"
}
```

##### update_room_status

ルームステータス更新（ホストのみ）

```json
{
  "room_id": "uuid",
  "status": "playing"
}
```

#### サーバー → クライアント

##### room_joined

ルーム参加完了

```json
{
  "room_id": "uuid",
  "members": [
    /* メンバー一覧 */
  ]
}
```

##### member_joined

新規メンバー参加

```json
{
  "room_id": "uuid",
  "member": {
    "user": {
      /* ユーザー情報 */
    },
    "player_number": 3
  }
}
```

##### member_left

メンバー退出

```json
{
  "room_id": "uuid",
  "user_id": "uuid"
}
```

##### new_message

新規メッセージ

```json
{
  "room_id": "uuid",
  "message": {
    /* メッセージ情報 */
  }
}
```

##### room_updated

ルーム情報更新

```json
{
  "room_id": "uuid",
  "updates": {
    "status": "playing"
  }
}
```

##### room_closed

ルームクローズ

```json
{
  "room_id": "uuid",
  "reason": "ホストが退出しました"
}
```

##### error

エラー通知

```json
{
  "code": "ERROR_CODE",
  "message": "エラーメッセージ"
}
```

## エラーコード一覧

| コード                   | 説明                         |
| ------------------------ | ---------------------------- |
| AUTH_INVALID_CREDENTIALS | 認証情報が無効               |
| AUTH_TOKEN_EXPIRED       | トークンの有効期限切れ       |
| AUTH_UNAUTHORIZED        | 認証が必要                   |
| AUTH_SUPABASE_ERROR      | Supabase認証エラー           |
| ROOM_NOT_FOUND           | ルームが見つからない         |
| ROOM_FULL                | ルームが満員                 |
| ROOM_PASSWORD_REQUIRED   | パスワードが必要             |
| ROOM_PASSWORD_INCORRECT  | パスワードが間違っている     |
| ROOM_ALREADY_MEMBER      | すでにメンバー               |
| ROOM_NOT_MEMBER          | メンバーではない             |
| ROOM_PERMISSION_DENIED   | 権限がない                   |
| USER_BLOCKED             | ユーザーがブロックされている |
| VALIDATION_ERROR         | バリデーションエラー         |
| RATE_LIMIT_EXCEEDED      | レート制限超過               |
| INTERNAL_ERROR           | 内部エラー                   |

## レート制限

### API呼び出し制限

- 認証なし: 60回/時間
- 認証あり: 600回/時間
- ルーム作成: 10回/時間

### WebSocket制限

- メッセージ送信: 30回/分
- ルーム参加: 10回/分

### レート制限ヘッダー

```
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 599
X-RateLimit-Reset: 1642680000
```

## Supabase Authentication統合

### 設定

- Supabase AuthプロジェクトのURL、Anonキーを環境変数で管理
- Google OAuth: SupabaseダッシュボードでGoogleプロバイダを有効化
- メール確認: Supabaseのメールテンプレート機能を使用

### セッション管理

- Supabase Authのセッション管理機能を利用
- JWTトークンの自動更新（メール・Google共通）
- マルチデバイス対応（メール・Google共通）

## セキュリティ考慮事項

1. **CORS設定**

   - 本番環境では適切なOriginのみ許可
   - 認証情報を含むリクエストの制限

2. **入力検証**

   - 全ての入力値のサニタイズ
   - SQLインジェクション対策（Fly.io PostgreSQL）
   - XSS対策

3. **認証トークン**

   - Supabase Authの標準設定に準拠（メール・Google共通）
   - アクセストークン: 1時間
   - リフレッシュトークン: 30日
   - セキュアなHTTPOnly Cookie使用を推奨

4. **データ保護**
   - パスワードはSupabase Authが管理（メール認証のみ、Googleはパスワード不要）
   - Google OAuthのアクセストークンはSupabaseが安全に管理
   - Fly.io PostgreSQLへの接続はSSL必須
   - 機密データの暗号化
