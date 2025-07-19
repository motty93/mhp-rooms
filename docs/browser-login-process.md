# ブラウザでのログイン状態確認プロセス

## 実施内容

### 1. 開発環境の準備
- ローカル開発サーバーが既に起動していることを確認（http://localhost:8080）
- `local-storage.json`ファイルに保存されている認証情報を確認

### 2. Playwrightでブラウザアクセス
- Playwright MCPを使用してhttp://localhost:8080にアクセス
- 初期状態では未認証（ヘッダーに「ログイン」「新規登録」ボタンが表示）
- ブラウザサイズを1280x800に変更してデスクトップビューで確認

### 3. 認証情報の設定
local storageに認証データを設定する必要がある。

**方法1: 環境変数から読み込む（推奨）**

```javascript
// 環境変数TEST_AUTH_DATAにJSON文字列として認証情報を設定
// 例: TEST_AUTH_DATA='{"provider_token":"xxx","access_token":"xxx",...}'
const authData = JSON.parse(process.env.TEST_AUTH_DATA || '{}');

// local storageに設定
localStorage.setItem('sb-pfkqrwtgfpbxduxtecby-auth-token', JSON.stringify(authData));
console.log('認証トークンをlocal storageに設定しました');

// ページをリロード
location.reload();
```

**方法2: local-storage.jsonファイルから読み込む**

```javascript
// local-storage.jsonファイルの内容をコピーしてauthDataに設定
// ※ファイルの内容を直接貼り付ける場合
const authData = /* local-storage.jsonの内容をここに貼り付け */;

// local storageに設定
localStorage.setItem('sb-pfkqrwtgfpbxduxtecby-auth-token', JSON.stringify(authData));
console.log('認証トークンをlocal storageに設定しました');

// ページをリロード
location.reload();
```

**開発環境での使用方法**:
1. `.env`ファイルに`TEST_AUTH_DATA`として認証情報のJSONを設定
2. または、`local-storage.json`ファイルに実際の認証情報を保存（このファイルは.gitignoreに追加済み）
3. ブラウザのコンソールで上記のコードを実行

### 4. 手動での認証情報設定手順
1. ブラウザで http://localhost:8080 を開く
2. F12キーまたは右クリック→「検証」でDevToolsを開く
3. Consoleタブに切り替える
4. 上記のJavaScriptコードを貼り付けて実行
5. ページが自動的にリロードされる

### 5. 期待される結果
- ヘッダーの右上にユーザーアイコンとメールアドレスが表示される
- 「ログイン」「新規登録」ボタンが非表示になる
- ログイン状態でのみ利用可能な機能にアクセスできるようになる

### 6. 注意事項
- Playwrightの制限により、直接local storageを操作することができないため、手動でのJavaScript実行が必要
- 認証トークンには有効期限があるため、期限切れの場合は新しいトークンが必要
- CLAUDE.mdのUI/UX設計ルールにより、モバイルビュー（768px未満）では認証ボタンはヘッダーに表示されない
