# player_namesテーブルのリファクタリング

実装時間: 20分

## 概要
player_namesテーブルをゲームバージョンごとにプレイヤー名を管理できる適切な構造にリファクタリングしました。

## 実装内容

### 1. モデルの修正
`internal/models/player_name.go`を修正：
- `GameVersion string`から`GameVersionID uuid.UUID`に変更
- game_versionsテーブルとの外部キー関係を追加
- カラム名を`PlayerName`から`Name`に変更
- 複合ユニークインデックスを追加（user_id + game_version_id）

### 2. PlayerNameRepositoryの作成
`internal/repository/player_name_repository.go`を新規作成：
- `CreatePlayerName`: 新規作成
- `UpdatePlayerName`: 更新
- `FindPlayerNameByUserAndGame`: 特定のゲームのプレイヤー名取得
- `FindAllPlayerNamesByUser`: ユーザーの全プレイヤー名取得
- `DeletePlayerName`: 削除
- `UpsertPlayerName`: 作成または更新

### 3. マイグレーションスクリプト
`scripts/migrate_player_names.sql`を作成：
- 既存データの移行（game_version文字列→game_version_id UUID）
- 外部キー制約の追加
- ユニーク制約の追加
- インデックスの作成

### 4. 新規登録フローの更新
`internal/handlers/auth.go`の`RegisterHandler`を修正：
- ユーザー作成後、デフォルトでMHP3のプレイヤー名を保存
- MHP3がない場合は最初のゲームバージョンを使用

### 5. マイグレーション設定の更新
`internal/infrastructure/persistence/postgres/db.go`を修正：
- player_namesテーブルの外部キー制約を追加
- インデックスを新しい構造に対応

## 特に注意した点
1. **データの整合性**: game_versionsテーブルとの外部キー関係により、不正なゲームバージョンの登録を防止
2. **既存データの移行**: SQLスクリプトで既存データを新しい構造に移行可能
3. **パフォーマンス**: 複合インデックスによりユーザーとゲームバージョンの組み合わせでの検索を高速化

## 今後の展開
- プロフィール画面でゲームごとのプレイヤー名を編集できる機能
- ルーム作成時に適切なゲームバージョンのプレイヤー名を自動選択
- ゲームバージョン追加時の対応