## MHP, MHP2, MHP2G, MHP3
### 概要
モンスターハンターポータブルシリーズのアドホックパーティのルームを管理するwebアプリケーションです。

## 技術スタック
- **バックエンド**: Go (Golang)
- **フロントエンド**: HTMX
- **スタイリング**: Tailwind CSS


## プロジェクト構造

```
.
├── cmd/server/          # メインアプリケーション
├── internal/
│   ├── handlers/        # HTTPハンドラー
│   ├── models/          # データモデル
│   └── services/        # ビジネスロジック
├── templates/           # HTMLテンプレート
├── static/              # 静的ファイル
├── docs/                # ドキュメントファイル
├── Makefile            # ビルドタスク
└── README.md
```


## 開発

### 利用可能なコマンド

```bash
make build         # アプリケーションをビルド
make run           # アプリケーションを実行
make dev           # 開発サーバーを起動（ホットリロード付き）
make test          # テストを実行
make lint          # リンターを実行
make fmt           # コードをフォーマット
make clean         # ビルド成果物をクリーンアップ
```

## ライセンス

MIT
