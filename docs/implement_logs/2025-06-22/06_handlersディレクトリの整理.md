# handlersディレクトリの整理

**実装時間**: 約5分

## 実装概要

混乱しやすかった`handler.go`と`handlers.go`の2つのファイルを統合し、より分かりやすいディレクトリ構造に整理しました。

## 実装内容

### 1. ファイル統合

#### 統合前の構造
```
internal/handlers/
├── handler.go       # Handler構造体とNewHandler関数（15行）
├── handlers.go      # 共通機能（TemplateData、renderTemplate等）
├── home.go         # ホームページハンドラー
└── rooms.go        # ルーム関連ハンドラー
```

#### 統合後の構造
```
internal/handlers/
├── handlers.go     # Handler構造体 + 共通機能
├── home.go         # ホームページハンドラー
└── rooms.go        # ルーム関連ハンドラー
```

### 2. 実装詳細

#### handlers.goの変更
- `handler.go`の内容を`handlers.go`に統合
- Handler構造体とNewHandler関数を追加
- 既存の共通機能（TemplateData、renderTemplate、HelloHandler）を保持

#### handler.goの削除
- 不要になった`handler.go`を削除

## 技術的な改善点

### 1. 可読性の向上
- 似た名前のファイル（`handler.go`と`handlers.go`）による混乱を解消
- ファイル名から役割が明確になった

### 2. 保守性の向上
- 関連するコード（Handler構造体と共通機能）が一箇所に集約
- 小さなファイルを統合してプロジェクト構造を簡潔化

### 3. 論理的な構造
- `handlers.go`: ハンドラーの基盤となる構造体と共通機能
- `home.go`: ホームページ専用のハンドラー
- `rooms.go`: ルーム機能専用のハンドラー

## 実装の特徴

### 1. 非破壊的変更
- 既存のインポート文は変更不要
- パッケージ名とエクスポートされる関数・構造体は同じ
- 外部からの利用方法は変更なし

### 2. 機能的な統合
- Handler構造体の定義と共通ユーティリティを一箇所に配置
- テンプレート処理とHTTPハンドラーの基盤機能を統合

## ファイル構成の最終形

### handlers.go
```go
// Handler構造体とコンストラクタ
type Handler struct { ... }
func NewHandler(...) *Handler { ... }

// 共通データ構造
type TemplateData struct { ... }

// 共通ユーティリティ
func renderTemplate(...) { ... }

// 汎用ハンドラー
func HelloHandler(...) { ... }
```

### home.go
```go
// ホームページ専用ハンドラー
func (h *Handler) HomeHandler(...) { ... }
```

### rooms.go
```go
// ルーム関連のハンドラー
func (h *Handler) RoomsHandler(...) { ... }
func (h *Handler) CreateRoomHandler(...) { ... }
func (h *Handler) JoinRoomHandler(...) { ... }
func (h *Handler) LeaveRoomHandler(...) { ... }
```

## 動作確認

- ビルドエラーなし
- 既存の機能に影響なし
- ファイル数が1つ減少（4ファイル → 3ファイル）

## 今後の改善案

1. **更なる分離検討**
   - 認証機能実装時は`auth.go`を追加
   - 管理機能実装時は`admin.go`を追加

2. **テスト構造の整理**
   - ハンドラーテストファイルも同様に整理

3. **ミドルウェアの分離**
   - 認証やロギングなどのミドルウェアは別ディレクトリで管理

## 利点

- **開発者体験の向上**: ファイル名の混乱がなくなった
- **メンテナンス性**: 関連コードが集約され、変更時の影響範囲が明確
- **プロジェクト構造**: より論理的で理解しやすい構造