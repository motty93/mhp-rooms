# Turso対応とルーム参加者数表示修正

## 実装日時
開始: 2025-09-19 11:00
完了: 2025-09-19 12:30
実装時間: 約1時間30分

## 実装した機能の概要

### 1. cmd/seed/main.go のTurso対応
- PostgreSQL専用だったシードコマンドをTurso（SQLite）でも実行可能に変更
- データベース抽象化レイヤーを利用したアダプター実装

### 2. ルーム参加者数表示の修正
- ルーム一覧で参加者数が0と表示される問題を解決
- SQL クエリの列名衝突問題を修正
- シードデータでのroom_members作成処理を追加

## 修正したファイル

### cmd/seed/main.go
**問題**: PostgreSQL専用の実装でTursoで実行できない

**修正内容**:
```go
// 修正前
import "mhp-rooms/internal/infrastructure/persistence/postgres"
db, err := postgres.NewDB(config.AppConfig)

// 修正後
import "mhp-rooms/internal/infrastructure/persistence"
db, err := persistence.NewDBAdapter(config.AppConfig)
```

**効果**: データベースファクトリパターンによりPostgreSQL・Turso両対応

### internal/repository/room_repository.go
**問題1**: GetActiveRoomsでcurrent_playersが正しく計算されない
**問題2**: GetActiveRoomsWithJoinStatusでSQL列名衝突

**修正内容**:
```go
// GetActiveRooms - 個別カウント方式に変更
for i := range rooms {
    var count int64
    r.db.GetConn().Model(&models.RoomMember{}).
        Where("room_id = ? AND status = ?", rooms[i].ID, "active").
        Count(&count)
    rooms[i].CurrentPlayers = int(count)
}

// GetActiveRoomsWithJoinStatus - 明示的な列指定で衝突回避
SELECT
    rooms.id, rooms.room_code, rooms.name, rooms.description,
    rooms.game_version_id, rooms.host_user_id, rooms.max_players,
    COUNT(DISTINCT rm_all.id) as current_players,
    // rooms.current_playersを除外
```

### internal/handlers/rooms.go
**問題**: ルーム作成時の初期参加者数が1に設定されている

**修正内容**:
```go
// 修正前
CurrentPlayers: 1, // ホストを含めた初期人数

// 修正後
CurrentPlayers: 0, // 初期人数（メンバー追加処理で更新される）
```

## 特に注意した点・工夫した点

### 1. データベース抽象化の活用
- 既存のファクトリパターンを利用してコード重複を避けた
- PostgreSQL・Turso両環境での動作を保証

### 2. SQL クエリの最適化
- GORM のエイリアス衝突問題を回避
- 明示的な列指定により予期しない動作を防止
- COUNT集計の正確性を確保

### 3. データ整合性の確保
- room_members テーブルとの外部キー関係を適切に維持
- ホストユーザーの room_members レコード作成処理を追加

## テスト結果・動作確認

### 1. Turso環境でのシード実行
```bash
ENV=development DB_TYPE=turso go run cmd/seed/main.go
# 成功: エラーなく完了
```

### 2. 参加者数表示の確認
```sql
-- テストルーム作成後の確認
SELECT
    rooms.name,
    COUNT(DISTINCT rm_all.id) as current_players
FROM rooms
LEFT JOIN room_members rm_all ON rooms.id = rm_all.room_id AND rm_all.status = 'active'
WHERE rooms.name = 'テストルーム2025'
GROUP BY rooms.id;

-- 結果: current_players = 1 (正常)
```

### 3. フロントエンド表示確認
- ルーム一覧画面で参加者数が正しく表示される
- ホストのみのルームで「1」と表示される

## 今後の作業・改善点

### 1. パフォーマンス最適化
- 大量ルーム対応時のクエリ最適化検討
- インデックス追加の検討

### 2. エラーハンドリング強化
- データベース切り替え時のエラー処理改善
- 不整合データの検出・修復機能

### 3. テストケース追加
- ユニットテストでの回帰防止
- 両データベース環境での統合テスト

## 学んだ教訓

1. **データベース抽象化の重要性**: 複数DB対応では抽象化レイヤーが必須
2. **SQL エイリアスの注意**: GORMでの列名衝突は予期しない結果を招く
3. **データ整合性の確認**: 外部キー関係は常に整合性を保つ必要がある
4. **段階的デバッグ**: SQLクエリ→バックエンド→フロントエンドの順で問題を特定

## 修正完了の確認事項

- [x] Turso環境でシードコマンド実行可能
- [x] PostgreSQL環境でシードコマンド実行可能
- [x] ルーム一覧で参加者数が正しく表示される
- [x] ホストのみルームで参加者数「1」表示
- [x] 新規ルーム作成時の参加者数計算が正常
- [x] 実装ログ作成完了