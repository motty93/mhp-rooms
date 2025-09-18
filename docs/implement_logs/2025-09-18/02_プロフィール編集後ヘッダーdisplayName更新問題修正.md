# プロフィール編集後のヘッダーdisplayName更新問題修正

**実装日**: 2025年9月18日
**実装時間**: 約30分
**対象機能**: プロフィール編集機能、ヘッダー表示機能

## 問題の概要

プロフィール編集でニックネーム（表示名）を変更・保存しても、ページヘッダーのユーザー名表示が更新されない問題が発生していた。

## 根本原因

### 1. ヘッダーの表示名データソース
ヘッダーコンポーネント（`templates/components/header.tmpl`）では、Alpine.jsの認証ストアから表示名を取得：

```html
<span class="font-medium hidden sm:block" x-text="$store.auth.displayName || $store.auth.username"></span>
```

### 2. auth storeのdisplayNameの実装
認証ストア（`static/js/auth-store.js`）では、`dbUser`からdisplayNameを取得：

```javascript
get displayName() {
  return this.dbUser?.display_name || ''
},
```

### 3. キャッシュ更新処理の不備
- **アバター更新時**: `authStore.refreshDbUser()`が呼ばれてキャッシュが更新される
- **プロフィール更新時**: `authStore.refreshDbUser()`が呼ばれずキャッシュが古いまま

## 修正内容

### プロフィール保存処理の修正

**対象ファイル**: `static/js/profile.js`

```javascript
// 修正前
if (response.ok) {
  const result = await response.json()
  // 通知を表示
  showNotification('プロフィールを更新しました', 'success')
  // プロフィール表示に戻る
  returnToProfileView()
}

// 修正後
if (response.ok) {
  const result = await response.json()
  // 通知を表示
  showNotification('プロフィールを更新しました', 'success')

  // Alpine.jsのauth storeを更新（DBから最新情報を取得）
  if (window.Alpine && Alpine.store('auth')) {
    const authStore = Alpine.store('auth')
    // DBから最新のユーザー情報を取得してstoreを更新
    await authStore.refreshDbUser()
  }

  // プロフィール表示に戻る
  returnToProfileView()
}
```

### 修正の詳細

1. **authStore.refreshDbUser()の追加**
   - プロフィール保存成功後に明示的にDBからユーザー情報を再取得
   - ローカルストレージのキャッシュもクリアして最新データに更新

2. **非同期処理の適切な処理**
   - `await authStore.refreshDbUser()`で確実にデータ更新を待機
   - その後にプロフィール表示画面に戻る

## データフローの確認

### 修正前の問題フロー
1. ユーザーがプロフィール編集でdisplayNameを変更
2. APIでデータベースは正しく更新される
3. しかし、`authStore.dbUser`のキャッシュは古いまま
4. ヘッダーの`$store.auth.displayName`は古い値を表示し続ける

### 修正後の正常フロー
1. ユーザーがプロフィール編集でdisplayNameを変更
2. APIでデータベースが正しく更新される
3. `authStore.refreshDbUser()`でキャッシュを最新データに更新
4. ヘッダーの`$store.auth.displayName`が即座に新しい値を表示

## 調査で確認した技術要素

### 1. ヘッダーコンポーネントの実装
- Alpine.jsのリアクティブシステムを使用
- `$store.auth.displayName`でauth storeの値を動的表示
- デスクトップとモバイルで表示制御が分かれている

### 2. 認証ストアの構造
- `dbUser`: データベースのユーザー情報をキャッシュ
- `displayName`: `dbUser.display_name`のgetter
- `refreshDbUser()`: DBから最新データを取得してキャッシュ更新

### 3. 既存の類似処理
- アバター更新処理では既に`refreshDbUser()`が実装済み
- 同じパターンをプロフィール更新にも適用

## テスト結果

修正後のテスト結果：
- ✅ プロフィール編集でdisplayName変更・保存
- ✅ 保存後すぐにヘッダーの表示名が更新される
- ✅ ページリロード不要で即座に反映
- ✅ 他のプロフィール情報も正常に保存・表示

## 学んだ重要なポイント

### 1. キャッシュ一貫性の重要性
- データベース更新とキャッシュ更新は必ずセットで行う
- 片方だけ更新されると画面表示と実際のデータが不整合になる

### 2. Alpine.jsのリアクティブシステム
- auth storeの値が更新されると、参照している全ての要素が自動更新される
- `refreshDbUser()`を呼ぶだけで全画面のユーザー情報表示が更新される

### 3. 既存パターンの活用
- アバター更新で既に実装されていたパターンを流用
- 同じような処理は統一されたアプローチを取ることが重要

## 今後の改善提案

### 1. 共通化の検討
プロフィール関連の更新処理を共通関数にまとめることを検討：

```javascript
// 共通のプロフィール更新後処理
async function handleProfileUpdateSuccess(message = 'プロフィールを更新しました') {
  showNotification(message, 'success')

  // auth storeを更新
  if (window.Alpine && Alpine.store('auth')) {
    const authStore = Alpine.store('auth')
    await authStore.refreshDbUser()
  }

  returnToProfileView()
}
```

### 2. エラーハンドリングの追加
`refreshDbUser()`の失敗時の処理も検討：

```javascript
try {
  await authStore.refreshDbUser()
} catch (error) {
  console.warn('ユーザー情報の更新に失敗しましたが、データは正常に保存されました', error)
}
```

## 関連Issue・副次的な効果

今回の修正により、以下の副次的な効果も期待される：
- プロフィール表示画面の情報も即座に更新される
- ユーザー体験の向上（ページリロード不要）
- データ整合性の向上

## 実装完了の確認事項

- [x] プロフィール編集でdisplayName変更が可能
- [x] 保存後にヘッダーの表示名が即座に更新
- [x] 他のプロフィール情報も正常に動作
- [x] アバター更新機能に影響がないことを確認
- [x] エラー時の動作も正常であることを確認