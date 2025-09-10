# auth-go調査とgotrue-go継続使用

実装時間: 15分

## 概要
ドキュメントでauth-goの使用が推奨されていたため調査しましたが、現在のsupabase-go (v0.0.4)はまだgotrue-goを内部で使用しているため、gotrue-goをそのまま使用することにしました。

## 調査内容

### 1. auth-goパッケージの確認
```bash
go get github.com/supabase-community/auth-go
```
- auth-go v1.4.0がインストールされました
- APIはgotrue-goとほぼ同じインターフェースを持っています

### 2. supabase-goの内部実装
```go
type Client struct {
    Storage *storage_go.Client
    Auth      gotrue.Client  // gotrue-goのインターフェースを使用
    Functions *functions.Client
}
```
- supabase-go v0.0.4はまだgotrue.Clientインターフェースを使用
- auth-goのtypesは互換性がありません

### 3. 型の不一致問題
```go
// auth-goのtypes
"github.com/supabase-community/auth-go/types".SignupRequest

// supabase-goが期待するtypes
"github.com/supabase-community/gotrue-go/types".SignupRequest
```
これらは別々のパッケージのため、直接的な互換性がありません。

## 決定事項

1. **gotrue-goの継続使用**
   - 現在のsupabase-go (v0.0.4)はgotrue-goを内部で使用
   - auth-goへの移行は現時点では不可能

2. **将来的な移行の可能性**
   - supabase-goのv2がリリースされた場合に再検討
   - 公式ドキュメントの更新を待つ

## 現在の実装

handlers/auth.goでの使用：
```go
import "github.com/supabase-community/gotrue-go/types"

// ログイン
resp, err := h.supabase.Auth.SignInWithEmailPassword(req.Email, req.Password)

// 新規登録
resp, err := h.supabase.Auth.Signup(types.SignupRequest{
    Email:    req.Email,
    Password: req.Password,
})
```

## 今後の対応

1. supabase-goの更新を定期的にチェック
2. v2がリリースされた場合は移行を検討
3. 現在の実装で問題なく動作することを確認済み