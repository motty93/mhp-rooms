# LocalStorageキャッシュシステム

## 概要

MHP Roomsでは、ユーザーのデータベース情報（DBユーザー情報）をlocalStorageにキャッシュすることで、API呼び出しを削減し、パフォーマンスを向上させています。

## キャッシュの仕組み

### 1. データ構造

localStorageには以下の形式でデータが保存されます：

```javascript
// キー: mhp-rooms-dbuser-{supabaseUserID}
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "display_name": "表示名",
    "psn_online_id": "psn_id",
    // その他のユーザー情報
  },
  "timestamp": 1757503357306,
  "expires": 1757589757306  // 24時間後
}
```

### 2. キャッシュのライフサイクル

#### 認証時の処理順序

1. **認証成功** → `updateSession()`が呼ばれる
2. **キャッシュ読み込み** → `loadDbUserFromStorage()`でlocalStorageから読み込み
3. **ユーザー同期** → `syncUser()`でSupabaseとの同期
4. **API呼び出し判定** → キャッシュがない場合のみ`fetchDbUser()`実行

#### 処理フロー

```javascript
// 1. キャッシュ読み込み
this.loadDbUserFromStorage()  // localStorageから読み込み

// 2. 同期処理
if (!this._syncInProgress) {
  this.syncUser(session.access_token)
}

// 3. syncUser内でのAPI呼び出し制御
if (!this.dbUser) {  // キャッシュがない場合のみ
  await this.fetchDbUser(accessToken)  // API呼び出し
}
```

### 3. キャッシュの有効期限

- **有効期限**: 24時間
- **期限切れチェック**: `loadDbUserFromStorage()`で自動実行
- **期限切れ時の処理**: 自動削除して`null`を返す

```javascript
// 期限切れチェック
if (Date.now() > parsedData.expires) {
  localStorage.removeItem(storageKey)
  return null
}
```

## パフォーマンスの利点

### API呼び出し削減

✅ **キャッシュヒット時（通常のページアクセス）**：
- API呼び出し: **なし**
- レスポンス時間: **即座**
- サーバー負荷: **なし**

❌ **キャッシュミス時（初回ログイン・期限切れ）**：
- API呼び出し: `/api/user/me`
- レスポンス時間: ネットワーク依存
- サーバー負荷: あり

### ユーザー体験の向上

1. **表示の高速化**: DisplayName/Username表示が即座に行われる
2. **ネットワーク使用量削減**: 不要なAPI呼び出しを削減
3. **オフライン耐性**: キャッシュがある限り基本情報は表示可能

## 実装詳細

### 主要メソッド

#### `saveDbUserToStorage(dbUser)`
```javascript
// DBユーザー情報をlocalStorageに保存
saveDbUserToStorage(dbUser) {
  const storageKey = `mhp-rooms-dbuser-${this.user?.id}`
  const storageData = {
    user: dbUser,
    timestamp: Date.now(),
    expires: Date.now() + (24 * 60 * 60 * 1000) // 24時間
  }
  localStorage.setItem(storageKey, JSON.stringify(storageData))
}
```

#### `loadDbUserFromStorage()`
```javascript
// localStorageからDBユーザー情報を読み込み
loadDbUserFromStorage() {
  const storageKey = `mhp-rooms-dbuser-${this.user.id}`
  const storedData = localStorage.getItem(storageKey)
  
  // 期限切れチェック
  if (Date.now() > parsedData.expires) {
    localStorage.removeItem(storageKey)
    return null
  }
  
  this.dbUser = parsedData.user
  return parsedData.user
}
```

#### `clearDbUserFromStorage(userId)`
```javascript
// 指定ユーザーのキャッシュを削除
clearDbUserFromStorage(userId = null) {
  const targetUserId = userId || this.user?.id
  if (!targetUserId) {
    // 全キャッシュをクリア
    const keys = Object.keys(localStorage)
    keys.forEach(key => {
      if (key.startsWith('mhp-rooms-dbuser-')) {
        localStorage.removeItem(key)
      }
    })
  } else {
    const storageKey = `mhp-rooms-dbuser-${targetUserId}`
    localStorage.removeItem(storageKey)
  }
}
```

#### `refreshDbUser()`
```javascript
// キャッシュを強制的に更新
async refreshDbUser() {
  if (!this.session?.access_token) return
  
  this.clearDbUserFromStorage()
  await this.fetchDbUser(this.session.access_token)
}
```

## セキュリティ考慮事項

### 保存データの制限

- **機密情報は保存しない**: パスワード、認証トークンは保存しない
- **公開可能な情報のみ**: 表示名、ユーザー名、プロフィール情報のみ
- **期限設定**: 24時間で自動期限切れ

### 注意事項

1. **共有端末での使用**: 他のユーザーがlocalStorageを見ることが可能
2. **ブラウザ制限**: localStorageの容量制限（通常5-10MB）
3. **JavaScript無効化**: localStorageが使用不可の場合はAPIに依存

## デバッグとメンテナンス

### 開発者向けコマンド

```javascript
// 現在のキャッシュ状況を確認
Object.keys(localStorage).filter(key => key.startsWith('mhp-rooms-dbuser'))

// 特定ユーザーのキャッシュ内容を確認
localStorage.getItem('mhp-rooms-dbuser-{userID}')

// キャッシュを手動で削除
Alpine.store('auth').clearDbUserFromStorage()

// キャッシュを強制更新
Alpine.store('auth').refreshDbUser()
```

### ログ出力

現在はデバッグ用のコンソールログは削除されていますが、必要に応じて以下を追加可能：

```javascript
console.log('Cache hit:', !!cachedUser)
console.log('API call skipped:', !!this.dbUser)
```

## 今後の改善案

1. **キャッシュサイズの最適化**: 不要なフィールドの除外
2. **期限の動的設定**: アクティブユーザーは期限延長
3. **バックグラウンド更新**: 期限切れ前に自動更新
4. **エラー処理強化**: localStorage無効時のフォールバック処理

## 関連ファイル

- `static/js/auth-store.js` - メインのキャッシュロジック
- `internal/handlers/auth.go` - `/api/user/me`エンドポイント
- `internal/middleware/auth.go` - 認証ミドルウェア

## 最終更新

- **更新日**: 2025-09-10
- **更新者**: Claude Code
- **バージョン**: v1.0