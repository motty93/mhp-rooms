# JSONB型の実装について

## 概要

本プロジェクトでは、PostgreSQLのJSONB型を活用してログデータの柔軟な管理を実現しています。この文書では、JSONB型を採用した理由と実装詳細について説明します。

## 背景

### 解決したかった課題

1. **ログデータの多様性**: 部屋作成、参加、退室など、異なるアクションで必要なデータ構造が異なる
2. **拡張性**: 新しいログ項目を追加する際のスキーマ変更を最小限にしたい
3. **クエリ性能**: ログデータを効率的に検索したい
4. **型安全性**: Go言語でのデータ操作を安全に行いたい

### 検討した選択肢

| 方法 | メリット | デメリット | 採用判定 |
|------|----------|------------|----------|
| 個別カラム | シンプル、型安全 | 拡張性が低い、カラム数増加 | ❌ |
| TEXT型でJSON文字列 | 実装簡単 | パフォーマンス悪い、型安全性なし | ❌ |
| **JSONB型** | **柔軟性、パフォーマンス、型安全性** | **実装がやや複雑** | **✅ 採用** |

## PostgreSQLのJSONB型について

### JSON vs JSONB

```sql
-- JSON型: テキスト形式で保存
json_column JSON

-- JSONB型: バイナリ形式で保存（推奨）
jsonb_column JSONB
```

### JSONBの利点

1. **高速クエリ**: バイナリ形式により高速な検索が可能
2. **インデックス対応**: GINインデックスでさらなる高速化
3. **演算子豊富**: `->`, `->>`, `@>`, `?` など多彩な演算子
4. **重複キー排除**: 自動的に重複キーを排除

## 実装詳細

### カスタムJSONB型の定義

```go
// internal/models/room_log.go

// JSONB はPostgreSQLのJSONBフィールド用のカスタム型
type JSONB map[string]interface{}

// Value はdriver.Valuerインターフェースを実装
func (j JSONB) Value() (driver.Value, error) {
    if j == nil {
        return nil, nil
    }
    return json.Marshal(j)
}

// Scan はsql.Scannerインターフェースを実装
func (j *JSONB) Scan(value interface{}) error {
    if value == nil {
        *j = nil
        return nil
    }

    var bytes []byte
    switch v := value.(type) {
    case []byte:
        bytes = v
    case string:
        bytes = []byte(v)
    default:
        return fmt.Errorf("cannot scan %T into JSONB", value)
    }

    return json.Unmarshal(bytes, j)
}
```

### モデルでの使用

```go
// RoomLog はルームアクションの監査ログ
type RoomLog struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    RoomID    uuid.UUID `gorm:"type:uuid;not null"`
    UserID    *uuid.UUID `gorm:"type:uuid"`
    Action    string    `gorm:"type:varchar(50);not null"`
    Details   JSONB     `gorm:"type:jsonb"`  // ← カスタムJSONB型を使用
    CreatedAt time.Time
}
```

### データ作成例

```go
// 部屋作成ログ
log := models.RoomLog{
    RoomID: room.ID,
    UserID: &room.HostUserID,
    Action: "create",
    Details: models.JSONB{
        "room_name": room.Name,
        "max_players": room.MaxPlayers,
    },
}

// 参加ログ
log := models.RoomLog{
    RoomID: roomID,
    UserID: &userID,
    Action: "join",
    Details: models.JSONB{
        "user_name": user.DisplayName,
        "join_type": "manual",
    },
}
```

## SQLクエリ例

### 基本的な検索

```sql
-- 特定のユーザーのアクションを検索
SELECT * FROM room_logs 
WHERE details->>'user_name' = 'ハンター太郎';

-- 特定の部屋に関するログを検索
SELECT * FROM room_logs 
WHERE details->>'room_name' = '初心者歓迎部屋';
```

### 複雑な検索

```sql
-- 最大プレイヤー数が4人以上の部屋の作成ログ
SELECT * FROM room_logs 
WHERE action = 'create' 
  AND (details->>'max_players')::integer >= 4;

-- 特定の条件を含むログを検索
SELECT * FROM room_logs 
WHERE details @> '{"join_type": "invite"}';
```

### インデックスの作成

```sql
-- ユーザー名での検索を高速化
CREATE INDEX idx_room_logs_user_name 
ON room_logs USING gin ((details->>'user_name'));

-- 部屋名での検索を高速化
CREATE INDEX idx_room_logs_room_name 
ON room_logs USING gin ((details->>'room_name'));
```

## パフォーマンス比較

### 従来のアプローチ（個別カラム）

```sql
-- テーブル設計
CREATE TABLE room_logs_old (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL,
    user_id UUID,
    action VARCHAR(50) NOT NULL,
    user_name VARCHAR(255),      -- 参加時のみ使用
    room_name VARCHAR(255),      -- 作成時のみ使用  
    max_players INTEGER,         -- 作成時のみ使用
    join_type VARCHAR(50)        -- 参加時のみ使用
);

-- 問題点：
-- 1. NULLカラムが多数発生
-- 2. 新しいログ項目ごとにスキーマ変更が必要
-- 3. アクションタイプごとに専用テーブルが必要になる可能性
```

### 現在のアプローチ（JSONB）

```sql
-- テーブル設計
CREATE TABLE room_logs (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL,
    user_id UUID,
    action VARCHAR(50) NOT NULL,
    details JSONB                -- 柔軟なデータ構造
);

-- 利点：
-- 1. スキーマが固定
-- 2. 新しいログ項目は details に追加するだけ
-- 3. 高速なクエリが可能
```

## ベストプラクティス

### 1. データ構造の一貫性

```go
// ✅ 良い例：アクションごとに一貫した構造
switch action {
case "create":
    details = models.JSONB{
        "room_name": roomName,
        "max_players": maxPlayers,
        "game_version": gameVersion,
    }
case "join":
    details = models.JSONB{
        "user_name": userName,
        "join_type": joinType,
    }
}
```

### 2. バリデーション

```go
// ✅ 必要なフィールドのバリデーション
func validateLogDetails(action string, details models.JSONB) error {
    switch action {
    case "create":
        if _, ok := details["room_name"]; !ok {
            return errors.New("room_name is required for create action")
        }
    case "join":
        if _, ok := details["user_name"]; !ok {
            return errors.New("user_name is required for join action")
        }
    }
    return nil
}
```

### 3. 型安全なアクセス

```go
// ✅ 型アサーションを安全に行う
func getUserName(details models.JSONB) string {
    if name, ok := details["user_name"].(string); ok {
        return name
    }
    return "不明"
}
```

## 注意点と制限事項

### 1. 型の変換

```go
// ❌ 直接的な型アサーションは危険
userName := details["user_name"].(string) // panic の可能性

// ✅ 安全な型アサーション
userName, ok := details["user_name"].(string)
if !ok {
    userName = "不明"
}
```

### 2. JSONBのサイズ制限

- PostgreSQLのJSONBは理論的には無制限ですが、実用的には1MB以下に抑制
- 大きなデータは別テーブルに分離することを推奨

### 3. インデックス戦略

```sql
-- ✅ よく検索される項目にインデックス作成
CREATE INDEX idx_room_logs_user_name ON room_logs USING gin ((details->>'user_name'));

-- ❌ 過度なインデックス作成は避ける（書き込み性能に影響）
```

## 今後の拡張性

### 新しいログタイプの追加

```go
// 新しいアクション「kick」を追加する場合
log := models.RoomLog{
    Action: "kick",
    Details: models.JSONB{
        "kicked_user": "問題ユーザー",
        "reason": "不適切な行為",
        "kicked_by": "ホスト",
    },
}
// スキーマ変更は不要！
```

### 分析クエリの例

```sql
-- 部屋作成の傾向分析
SELECT 
    details->>'game_version' as game_version,
    COUNT(*) as room_count,
    AVG((details->>'max_players')::integer) as avg_max_players
FROM room_logs 
WHERE action = 'create'
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY details->>'game_version';
```

## まとめ

JSONB型の採用により、以下を実現しました：

1. **柔軟性**: 異なるアクションで異なるデータ構造を統一的に管理
2. **拡張性**: 新しいログ項目の追加がスキーマ変更なしで可能
3. **パフォーマンス**: バイナリ形式とインデックスによる高速クエリ
4. **型安全性**: Go言語でのCustom型による安全なデータ操作

この実装により、ログシステムは将来の要件変更に柔軟に対応できる設計となっています。