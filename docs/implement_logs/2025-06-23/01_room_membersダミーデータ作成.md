# room_membersテーブルのダミーデータ作成

**実装時間**: 約20分（14:10 - 14:30）

## 実装概要

room_membersテーブル用のダミーデータSQLファイルを作成し、既存のrooms、usersテーブルのデータと整合性を保ちながら、リアルなルームメンバー構成を実装しました。

## 実装内容

### 1. テーブル構造の調査
- `internal/models/room_member.go`でモデル定義を確認
- 以下のカラム構成を把握：
  - `id` (UUID): 主キー
  - `room_id` (UUID): ルームへの外部キー
  - `user_id` (UUID): ユーザーへの外部キー
  - `player_number` (INT): プレイヤー番号（1-4）
  - `is_host` (BOOLEAN): ホストフラグ
  - `status` (VARCHAR): アクティブ状態（active/inactive）
  - `joined_at` (TIMESTAMP): 参加時刻
  - `left_at` (TIMESTAMP): 退出時刻（NULL可）

### 2. ダミーデータの設計
- 各ルームに1〜4人のメンバーを配置
- ホストユーザーは必ず`player_number=1`、`is_host=true`
- 参加時刻を時系列順に設定してリアリティを追求
- 一部のメンバーは退出済み（inactive状態）として履歴を表現
- 「進行中の部屋」は満室（4人）に設定

### 3. 実装上の課題と解決
- **問題**: 4人目のメンバー用にgen_random_uuid()で生成したユーザーIDが外部キー制約違反
- **解決**: 固定IDのゲストユーザーを事前に作成し、そのIDを使用

### 4. データ投入結果
```sql
-- 最終的なルーム別メンバー数
進行中の部屋: 4人（満室）
上位ティガレックス討伐: 1-3人（複数ルーム）
初心者歓迎部屋: 1-3人（複数ルーム）
レア素材狙い: 1-3人（複数ルーム）
```

## 作成ファイル

- `/scripts/seed_room_members.sql`: room_membersテーブルのダミーデータSQL

## テスト結果

1. SQLファイルの実行に成功
2. 各ルームに適切な人数のメンバーが配置されたことを確認
3. ホストユーザーの設定が正しいことを確認
4. 外部キー制約が正しく機能していることを確認

## 実行コマンド

```bash
# SQLファイルの実行
cat scripts/seed_room_members.sql | docker exec -i mhp-rooms-db-1 psql -U mhp_user -d mhp_rooms_dev

# 結果確認
docker exec mhp-rooms-db-1 psql -U mhp_user -d mhp_rooms_dev -c "SELECT r.name, COUNT(rm.id) as members FROM rooms r LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active' GROUP BY r.id, r.name ORDER BY members DESC;"
```

## 今後の改善点

- ゲストユーザーの作成をSQLファイル内に含める
- より多様なステータス（観戦者など）の追加検討
- 参加・退出の履歴をより詳細に記録する仕組み