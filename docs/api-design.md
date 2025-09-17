# MHP Rooms API仕様書

## 概要

モンスターハンターポータブルシリーズ（MHP、MHP2、MHP2G、MHP3）のアドホックパーティルーム管理システムのAPI仕様書です。

## 基本情報

- **ベースURL**: `https://mhp-rooms.fly.dev`（本番環境）
- **認証方式**: JWT（Bearer Token）
- **データ形式**: JSON, HTML
- **文字エンコーディング**: UTF-8

## 認証

本システムではSupabaseを使用したJWT認証を採用しています。保護されたエンドポイントへのアクセスには、`Authorization`ヘッダーにBearerトークンを含める必要があります。

```
Authorization: Bearer <jwt_token>
```

## エンドポイント一覧

### 1. ページルート（HTML）

これらのエンドポイントは主にHTMLページを返します。

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/` | GET | ホームページ | オプショナル |
| `/terms` | GET | 利用規約 | オプショナル |
| `/privacy` | GET | プライバシーポリシー | オプショナル |
| `/contact` | GET | お問い合わせ | オプショナル |
| `/faq` | GET | よくある�����問 | オプショナル |
| `/guide` | GET | 利用ガイド | オプショナル |
| `/sitemap.xml` | GET | サイトマップ（XML形式） | オプショナル |
| `/profile` | GET | マイプロフィールページ | **必須** |
| `/profile/edit` | GET | プロフィール編集ページ | **必須** |
| `/profile/view` | GET | プロフィール表示ページ | **必須** |
| `/users/{uuid}` | GET | 他ユーザーのプロフィールページ | オプショナル |
| `/rooms` | GET | ルーム一覧ページ | オプショナル |
| `/rooms/{id}` | GET | ルーム詳細ページ | オプショナル |

### 2. 認証関連 (HTML & API)

#### 2.1 ページ表示

| エンドポイント | メソッド | 説明 |
|---|---|---|
| `/auth/login` | GET | ログインページ |
| `/auth/register` | GET | 新規登録ページ |
| `/auth/password-reset` | GET | パスワードリセット申請ページ |
| `/auth/password-reset/confirm` | GET | パスワードリセット確認ページ |
| `/auth/complete-profile` | GET | プロフィール補完ページ |

#### 2.2 認証アクション

| エンドポイント | メソッド | 説明 |
|---|---|---|
| `/auth/login` | POST | ログイン処理 |
| `/auth/register` | POST | 新規登録処理 |
| `/auth/logout` | POST | ログアウト処理 |
| `/auth/password-reset` | POST | パスワードリセット申請 |
| `/auth/password-reset/confirm` | POST | パスワードリセット実行 |
| `/auth/google` | GET | Google認証開始 |
| `/auth/google/callback` | GET | Google認証コールバック |
| `/auth/callback` | GET | 汎用認証コールバック |
| `/auth/complete-profile` | POST | プロフィール補完処理 |

#### 2.3 認証API

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/api/auth/sync` | POST | Supabase認証後、ユーザー情報をDBに同期 | **必須** |
| `/api/auth/psn-id` | PUT | PSN IDを更新 | **必須** |

### 3. ルーム管理 (HTML & API)

#### 3.1 ルーム操作

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/rooms` | POST | 新規ルーム作成 | **必須** |
| `/rooms/{id}` | PUT | ルーム情報更新 | **必須** |
| `/rooms/{id}` | DELETE | ルーム解散 | **必須** |
| `/rooms/{id}/join` | POST | ルームに参加 | **必須** |
| `/rooms/{id}/leave` | POST | ルームから退出 | **必須** |
| `/rooms/{id}/toggle-closed` | PUT | ルームの募集状態を切り替え | **必須** |

#### 3.2 ルームメッセージ

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/rooms/{id}/messages` | GET | メッセージ一覧を取得 | **必須** |
| `/rooms/{id}/messages` | POST | メッセージを送信 | **必須** |
| `/rooms/{id}/messages/stream` | GET | SSEでメッセージをストリーム | **必須 (一時トークン)** |
| `/rooms/{id}/sse-token` | POST | SSE接続用の一時トークンを生成 | **必須** |

### 4. APIエンドポイント (`/api`)

#### 4.1 ユーザー・プロフィール関連

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/api/user/current` | GET | ログイン中ユーザーの基本情報を取得 | **必須** |
| `/api/user/me` | GET | ログイン中ユーザーの詳細情報を取得 | **必須** |
| `/api/user/current-room` | GET | ログイン中ユーザーが参加しているルームを取得 | **必須** |
| `/api/user/current/room-status` | GET | ログイン中ユーザーのルーム参加状態を取得 | **必須** |
| `/api/leave-current-room` | POST | 現在参加中のルームから退出 | **必須** |
| `/api/profile/update` | POST | プロフィール情報を更新 | **必須** |
| `/api/profile/upload-avatar` | POST | アバター画像をアップロード | **必須** |
| `/api/users/{uuid}` | GET | 指定ユーザーのプロフィール情報を取得 | オプショナル |
| `/api/users/{uuid}/rooms` | GET | 指定ユーザーが作成したルーム一覧を取得 | オプショナル |
| `/api/users/{uuid}/activity` | GET | 指定ユーザーのアクティビティを取得 | オプショナル |
| `/api/users/{uuid}/followers` | GET | 指定ユーザーのフォロワー一覧を取得 | オプショナル |
| `/api/users/{uuid}/following` | GET | 指定ユーザーがフォロー中のユーザー一覧を取得 | オプショナル |

#### 4.2 フォロー関連

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/api/users/{userID}/follow` | POST | ユーザーをフォローする | **必須** |
| `/api/users/{userID}/unfollow` | DELETE | ユーザーのフォローを解除する | **必須** |
| `/api/users/{userID}/follow-status` | GET | フォロー状態を取得 | **必須** |

#### 4.3 リアクション関連

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/api/messages/{messageId}/reactions` | GET | メッセージのリアクション一覧を取得 | オプショナル |
| `/api/messages/{messageId}/reactions` | POST | メッセージにリアクションを追加 | **必須** |
| `/api/messages/{messageId}/reactions/{reactionType}` | DELETE | リアクションを削除 | **必須** |
| `/api/reactions/types` | GET | 利用可能なリアクション種別一覧を取得 | オプショナル |

#### 4.4 その他API

| エンドポイント | メソッド | 説明 | 認証 |
|---|---|---|---|
| `/api/rooms` | GET | アクティブなルーム一覧をJSONで取得 | オプショナル |
| `/api/config/supabase` | GET | Supabaseのフロントエンド用設定を取得 | 不要 |
| `/api/health` | GET | サービス稼働状態を確認 | 不要 |
| `/api/game-versions/active` | GET | アクティブなゲームバージョン一覧を取得 | 不要 |

## データモデル

(データモデルのセクションは変更ありません)

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
  nintendo_network_id?: string;
  nintendo_switch_id?: string;
  pretendo_network_id?: string;
  twitter_id?: string;
  favorite_games: string[];      // JSONBフィールド
  play_times: {                  // JSONBフィールド
    weekday?: string;
    weekend?: string;
  };
  is_active: boolean;
  role: string;                  // "user" | "admin" | "dummy"
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
  target_monster?: string;
  rank_requirement?: string;
  is_active: boolean;
  is_closed: boolean;
  created_at: string;            // ISO 8601
  updated_at: string;            // ISO 8601
  closed_at?: string;            // ISO 8601
}
```

## エラーレスポンス

(エラーレスポンスのセクションは変更ありません)

すべてのAPIエンドポイントは、エラー発生時に以下の形式でレスポンスを返します：

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "エラーの詳細メッセージ"
  }
}
```
