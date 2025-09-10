# config.Init重複修正作業ログ

## 実施日時
2025-01-20

## 作業概要
config.Init()の重複呼び出しを修正し、main.goで一度だけ初期化するよう改善

## 発生していた問題
- `database.WaitForDB()`でconfig.Init()を呼び出し
- `database.InitDB()`でもconfig.Init()を呼び出し
- main.goの実行時に設定が2回初期化される無駄が発生

## 実施内容

### 1. main.goの修正（cmd/server/main.go）

#### configパッケージのインポート追加
```go
import (
    // ...
    "mhp-rooms/internal/config"  // 追加
    // ...
)
```

#### main関数での設定初期化
```go
func main() {
    // 設定を初期化
    log.Println("設定を初期化中...")
    config.Init()

    log.Println("データベース接続を待機中...")
    // ...
}
```

#### サーバーアドレス取得の改善
```go
// Before
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
fmt.Printf("サーバーを起動しています... :%s\n", port)
log.Fatal(http.ListenAndServe(":"+port, r))

// After
addr := config.AppConfig.GetServerAddr()
fmt.Printf("サーバーを起動しています... %s\n", addr)
log.Fatal(http.ListenAndServe(addr, r))
```

### 2. database.goの修正

#### InitDB関数からconfig.Init()を削除
```go
// Before
func InitDB() error {
    config.Init()
    dsn := config.AppConfig.GetDSN()
    // ...
}

// After
func InitDB() error {
    dsn := config.AppConfig.GetDSN()
    // ...
}
```

### 3. wait.goの修正

#### WaitForDB関数からconfig.Init()を削除
```go
// Before
func WaitForDB(maxRetries int, retryInterval time.Duration) error {
    // 設定を初期化
    config.Init()
    dsn := config.AppConfig.GetDSN()
    // ...
}

// After
func WaitForDB(maxRetries int, retryInterval time.Duration) error {
    dsn := config.AppConfig.GetDSN()
    // ...
}
```

## 修正の利点

### 1. 効率性の向上
- 設定の初期化が1回のみに削減
- 無駄な処理の排除

### 2. 明確な責任分離
- main.goが設定初期化の責任を持つ
- database関連の関数は設定を使用するのみ

### 3. 保守性の向上
- 設定の初期化箇所が明確
- デバッグ時の追跡が容易

### 4. 統一性の確保
- アプリケーション全体で同じ設定インスタンスを使用
- サーバー設定も統一された方法で取得

## 実行フロー
1. main.goでconfig.Init()実行
2. WaitForDB()で設定済みの設定を使用
3. InitDB()で設定済みの設定を使用
4. サーバー起動時も設定済みの設定を使用

## 今後の考慮事項
- 他のcmdディレクトリのmain.go（migrate、seed）でも同様の修正が必要かチェック
- テスト環境での設定初期化方法の検討