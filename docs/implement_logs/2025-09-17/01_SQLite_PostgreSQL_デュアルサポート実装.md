# SQLite/PostgreSQL デュアルサポート実装

## 実装開始時間
2025-09-17 23:30

## 実装完了時間
2025-09-18 00:05

## 所要時間
約35分

## 実装概要

月間100万PVまではSQLiteで運用し、それを超えたらPostgreSQLに移行する段階的な戦略を実装しました。

## 実装内容

### 1. データベースアダプター層の作成

#### インターフェース定義 (`internal/infrastructure/persistence/adapter.go`)
- `DBAdapter`インターフェースで共通のデータベース操作を定義
- `GetConn()`, `Close()`, `Migrate()`, `GetType()`メソッド
- 共通マイグレーション関数`CommonMigrate()`を提供

#### ファクトリーパターン (`internal/infrastructure/persistence/factory.go`)
- `NewDBAdapter()`関数で設定に応じて適切なアダプターを作成
- SQLite/PostgreSQLの選択を環境変数で制御

### 2. SQLiteアダプターの実装 (`internal/infrastructure/persistence/sqlite/db.go`)

#### 主要機能
- WALモードとジャーナルモードでSQLiteの並行性を向上
- 外部キー制約の有効化
- SQLite固有のインデックスとユニーク制約の作成
- 初期データの挿入（Platform, GameVersion, ReactionType）

#### SQLite固有の最適化
```go
dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_foreign_keys=ON", dbPath)
sqlDB.SetMaxOpenConns(1) // SQLiteはシングル書き込みなので1に制限
```

### 3. PostgreSQLアダプターの調整

#### 既存実装の統合
- 既存のPostgreSQL実装を新しいアダプターインターフェースに準拠
- 共通マイグレーション関数の使用
- PostgreSQL固有の制約とインデックス処理を維持

### 4. モデル層のUUID対応

#### BeforeCreateフックの実装
- `internal/models/hooks.go`でUUID自動生成フックを集約
- PostgreSQLの`gen_random_uuid()`からGoの`uuid.New()`へ統一
- 全モデルに`BeforeCreate`フックを追加

#### 主要な変更
```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    return nil
}
```

### 5. 設定ファイルの拡張

#### 新しい設定項目
```go
type DatabaseConfig struct {
    Type       string // "sqlite" or "postgres"
    // 既存フィールド...
    SQLitePath string // SQLite用のファイルパス
}
```

#### 環境変数
- `DB_TYPE`: "sqlite" or "postgres" (デフォルト: sqlite)
- `SQLITE_DB_PATH`: SQLiteファイルパス (デフォルト: ./data/mhp_rooms.db)

### 6. マイグレーション処理の共通化

#### 待機処理の統合 (`internal/infrastructure/persistence/wait.go`)
- SQLiteは待機不要
- PostgreSQLのみ接続確認を実行

#### cmd/migrate/main.go の更新
- 新しいファクトリー関数を使用
- データベース種別の表示

## テスト結果

### SQLiteテスト
```bash
DB_TYPE=sqlite go run cmd/migrate/main.go
# ✅ 成功: マイグレーションが正常に完了
```

### PostgreSQLテスト
```bash
DB_TYPE=postgres go run cmd/migrate/main.go
# ✅ 成功: マイグレーションが正常に完了
```

## 注意した点・工夫した点

### 1. インポートサイクルの回避
- パッケージ構造を適切に分離
- 共通ロジックを独立したファイルに配置

### 2. UUID生成の統一
- PostgreSQL固有の`gen_random_uuid()`を排除
- Goのuuid.Newを使用してクロスプラットフォーム対応

### 3. SQLite最適化
- WALモードで読み書き性能向上
- 接続数を1に制限してSQLiteの特性に合わせる
- 外部キー制約を有効化してデータ整合性を確保

### 4. 初期データの整合性
- `Code`フィールドなど必須項目の適切な設定
- ユニークインデックス制約への対応

## 今後の改善点・実装予定

### 1. Turso（クラウドSQLite）対応【重要】

現在の実装は標準SQLiteドライバーを使用していますが、TursoにはlibSQLドライバーが必要です。

#### 必要な追加実装
- `github.com/tursodatabase/libsql-client-go/libsql`ドライバーの統合
- 認証トークン対応
- 接続文字列の処理（`libsql://`プロトコル）

```go
// 実装予定のTurso対応コード
if strings.HasPrefix(cfg.Database.URL, "libsql://") {
    connector, err := libsql.NewConnector(cfg.Database.URL, 
        libsql.WithAuthToken(cfg.Database.AuthToken))
    conn, err := gorm.Open(sqlite.Dialector{Conn: sql.OpenDB(connector)}, &gorm.Config{})
}
```

### 2. データベース移行戦略

#### SQLite → PostgreSQL移行パス
1. **段階的移行**：月間100万PV到達時
2. **移行手順**：
   - 新PostgreSQL環境の準備
   - データエクスポート・インポートツールの実行
   - DNS切り替えによる最小ダウンタイム移行
   - 旧環境のバックアップ保持

#### 移行ツールの実装予定
```bash
# 予定している移行コマンド
go run cmd/db_migrate/main.go --from=sqlite --to=postgres
```

### 3. 運用監視機能

#### パフォーマンス監視
- 月間PV数の自動カウント機能
- データベースサイズ監視
- レスポンス時間測定
- 移行タイミング通知システム

#### 閾値設定
- 月間PV: 1,000,000件
- DB容量: 1GB
- 同時接続数: 100セッション

### 4. クラウド環境対応

#### Turso（開発・小規模運用）
- 地理的分散レプリケーション
- 自動バックアップ
- 低レイテンシアクセス

#### PostgreSQL（本格運用）
- Neon Serverless PostgreSQL
- 接続プーリング
- 読み取り専用レプリカ

### 5. 設定管理の強化
- 環境別設定ファイルの分離
- Kubernetes/Docker Swarm対応
- シークレット管理の強化

## データベース移行について

### 移行判断基準
以下の条件のいずれかを満たした場合にPostgreSQLへの移行を検討：

1. **トラフィック**：月間100万PV超過
2. **データ量**：SQLiteファイルサイズ1GB超過
3. **同時接続**：ピーク時100接続超過
4. **複雑性**：複雑なクエリでのパフォーマンス劣化

### 移行プロセス

#### 事前準備
1. PostgreSQL環境の構築（Neon等）
2. 移行ツールの準備・テスト
3. ダウンタイム最小化戦略の策定

#### 移行実行
1. **メンテナンスモード**：一時的なサービス停止
2. **データエクスポート**：SQLiteからのフルダンプ
3. **データインポート**：PostgreSQLへの変換・投入
4. **動作確認**：全機能のテスト実行
5. **サービス再開**：新DB環境での運用開始

#### 移行後
1. **パフォーマンス監視**：移行効果の測定
2. **バックアップ戦略**：PostgreSQL用バックアップ
3. **旧環境保持**：ロールバック用に一定期間保持

### 移行リスク管理

#### リスク要素
- データ損失の可能性
- ダウンタイムの延長
- パフォーマンス劣化
- アプリケーション互換性問題

#### 軽減策
- 段階的移行（読み取り専用モードでの事前検証）
- ロールバック計画の策定
- 十分なテスト環境での検証
- 移行手順の自動化

## メリット

- **初期コスト削減**: SQLiteで簡単に開始可能
- **開発効率**: ローカル開発環境の簡素化
- **スムーズな移行**: 成長に応じてPostgreSQLへ段階的移行
- **一貫性**: 両DB環境で同じアプリケーションロジック

## 運用戦略

### フェーズ1：開発・初期運用（SQLite）
- **期間**：サービス開始〜月間10万PV
- **DB**：ローカルSQLite
- **特徴**：シンプル、高速な開発サイクル

### フェーズ2：成長期（Turso）
- **期間**：月間10万PV〜100万PV
- **DB**：TursoクラウドSQLite
- **特徴**：地理的分散、自動バックアップ、スケーラビリティ

### フェーズ3：本格運用（PostgreSQL）
- **期間**：月間100万PV以上
- **DB**：Neon PostgreSQL等
- **特徴**：高性能、複雑クエリ対応、エンタープライズ機能

## 設定ファイル追加

### compose.sqlite.yml
SQLite開発環境用のDocker Compose設定を追加：

```yaml
services:
  app:
    environment:
      - DB_TYPE=sqlite
      - SQLITE_DB_PATH=/var/sqlite/mhp_rooms_dev.sqlite
    volumes:
      - sqlite_data:/var/sqlite
```

### .env.sqlite.sample
SQLite環境用のサンプル設定ファイルを追加：

```bash
DB_TYPE=sqlite
SQLITE_DB_PATH=/var/sqlite/mhp_rooms_dev.sqlite
ENV=development
```

## 将来的な拡張性

この実装により、以下の成長パスが可能になりました：

1. **小規模開始**：ローカルSQLiteで素早くプロトタイプ
2. **クラウド移行**：Tursoで地理的分散とバックアップ
3. **企業レベル**：PostgreSQLで高性能・高可用性

各フェーズで蓄積されたデータとアプリケーションロジックは維持され、技術的負債を最小限に抑えながらスケールアップできる設計となっています。