# godotenv追加とマイグレーション修正

## 実装日時
2025-06-20

## 問題
`make migrate` でデータベースに接続できないエラーが発生。

## 原因分析
1. **環境変数とデフォルト値の不一致**
   - `database.go`のデフォルト値: `DB_USER=postgres`, `DB_NAME=mhp_rooms`
   - `compose.yml`の設定: `POSTGRES_USER=mhp_user`, `POSTGRES_DB=mhp_rooms_dev`

2. **マイグレーションコマンドで.envファイルが読み込まれない**
   - `cmd/migrate/main.go`で環境変数読み込み処理が未実装

## 実装内容

### 1. 依存関係の追加
```bash
go get github.com/joho/godotenv
```

### 2. .envファイルの修正
```diff
# データベース設定
-DB_NAME=mhp_rooms
-DB_USER=postgres
-DB_PASSWORD=postgres
+DB_NAME=mhp_rooms_dev
+DB_USER=mhp_user
+DB_PASSWORD=mhp_password
```

### 3. マイグレーションコマンドの修正
`cmd/migrate/main.go`に環境変数読み込み処理を追加：

```go
import (
    // 既存のimport
    "github.com/joho/godotenv"
)

func main() {
    // .envファイルを読み込み
    if err := godotenv.Load(); err != nil {
        log.Printf(".envファイルの読み込みをスキップします: %v", err)
    }
    
    // 既存の処理...
}
```

## 修正ファイル
- `go.mod` - godotenvパッケージ追加
- `go.sum` - 依存関係ハッシュ更新
- `.env` - データベース接続情報修正
- `cmd/migrate/main.go` - 環境変数読み込み処理追加

## 結果
- `make migrate` が正常に実行されるようになった
- データベーステーブルが正常に作成された
- 初期データ（ゲームバージョン）が正常に挿入された

## 本番環境への影響
- **本番環境でも問題なし**
- サーバー本体（`cmd/server/main.go`）は`godotenv`を使用していない
- マイグレーションの`godotenv.Load()`はエラーでも処理継続
- 本番環境では.envファイルが存在しなくても動作する

## 設定方法
### 開発環境
`.env`ファイルから環境変数を読み込み

### 本番環境（Fly.io）
```bash
fly secrets set DB_HOST=your-fly-postgres-host
fly secrets set DB_USER=your-username
fly secrets set DB_PASSWORD=your-secure-password
fly secrets set DB_NAME=your-database
fly secrets set DB_SSLMODE=require
```

## 設計の利点
1. **12-Factor App原則に準拠**
2. **開発環境と本番環境の設定分離**
3. **セキュリティ**: .envファイルはDockerイメージに含まれない
4. **柔軟性**: 環境に応じた設定が可能