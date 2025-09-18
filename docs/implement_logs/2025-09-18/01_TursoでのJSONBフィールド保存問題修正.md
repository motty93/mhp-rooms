# TursoでのJSONBフィールド保存問題修正

**実装日**: 2025年9月18日
**実装時間**: 約2時間
**対象機能**: プロフィール編集のお気に入りゲーム・プレイ時間帯保存機能

## 問題の概要

プロフィール編集でお気に入りゲームとプレイ時間帯を選択・保存しても、Tursoデータベースで空のオブジェクト `{}` として保存される問題が発生していた。

## 根本原因

### 1. データベーススキーマの問題
- モデル定義で `gorm:"type:jsonb"` を指定していたが、**SQLite（Turso）にはJSONB型が存在しない**
- 結果として認識されない型がBLOB型として扱われていた

### 2. JSONB.Value()メソッドの問題
- `json.Marshal()` の結果を `[]byte` 形式で返していた
- SQLiteでは `[]byte` を返すとBLOB型として保存され、`string` を返すとTEXT型として保存される

## 修正内容

### 1. モデル定義の修正
```go
// 修正前
FavoriteGames JSONB `gorm:"type:jsonb;default:'[]'" json:"favorite_games"`
PlayTimes     JSONB `gorm:"type:jsonb;default:'{}'" json:"play_times"`

// 修正後
FavoriteGames JSONB `gorm:"type:text;default:'[]'" json:"favorite_games"`
PlayTimes     JSONB `gorm:"type:text;default:'{}'" json:"play_times"`
```

**対象ファイル**:
- `internal/models/user.go`
- `internal/models/user_activity.go`
- `internal/models/room_log.go`

### 2. JSONB.Value()メソッドの修正
```go
// 修正前
func (j JSONB) Value() (driver.Value, error) {
    marshaled, err := json.Marshal(j.Data)
    return marshaled, err  // []byte形式（BLOB型になる）
}

// 修正後
func (j JSONB) Value() (driver.Value, error) {
    marshaled, err := json.Marshal(j.Data)
    if err != nil {
        return nil, err
    }
    return string(marshaled), nil  // string形式（TEXT型になる）
}
```

**対象ファイル**: `internal/models/room_log.go`

### 3. UpdateUser関数の修正
Turso対応として明示的なトランザクション処理を追加：

```go
// 修正前
func (r *userRepository) UpdateUser(user *models.User) error {
    return r.db.GetConn().Save(user).Error
}

// 修正後
func (r *userRepository) UpdateUser(user *models.User) error {
    return r.db.GetConn().Transaction(func(tx *gorm.DB) error {
        if err := tx.Save(user).Error; err != nil {
            return err
        }
        return nil
    })
}
```

**対象ファイル**: `internal/repository/user_repository.go`

### 4. データベーススキーマの手動修正
既存のBLOB型列をTEXT型に変更：

```sql
-- favorite_games列をtext型に変更
ALTER TABLE users ADD COLUMN favorite_games_new TEXT DEFAULT '[]';
UPDATE users SET favorite_games_new = CASE
  WHEN favorite_games = '{}' OR favorite_games IS NULL THEN '[]'
  ELSE favorite_games
END;
ALTER TABLE users DROP COLUMN favorite_games;
ALTER TABLE users RENAME COLUMN favorite_games_new TO favorite_games;

-- play_times列をtext型に変更
ALTER TABLE users ADD COLUMN play_times_new TEXT DEFAULT '{}';
UPDATE users SET play_times_new = CASE
  WHEN play_times = '{}' OR play_times IS NULL THEN '{}'
  ELSE play_times
END;
ALTER TABLE users DROP COLUMN play_times;
ALTER TABLE users RENAME COLUMN play_times_new TO play_times;

-- BLOBデータをTEXT形式に変換
UPDATE users SET favorite_games = CAST(favorite_games AS TEXT) WHERE typeof(favorite_games) = 'blob';
UPDATE users SET play_times = CAST(play_times AS TEXT) WHERE typeof(play_times) = 'blob';
```

## 調査で追加したデバッグログ

問題の特定のため、以下のデバッグログを追加：

```go
// SetFavoriteGames関数
log.Printf("[DEBUG] SetFavoriteGames: 入力値 = %+v", games)
log.Printf("[DEBUG] SetFavoriteGames: 設定後のData = %+v (type: %T)", u.FavoriteGames.Data, u.FavoriteGames.Data)

// JSONB.Value()関数
log.Printf("[DEBUG] JSONB.Value(): Data = %+v (type: %T)", j.Data, j.Data)
log.Printf("[DEBUG] JSONB.Value(): returning string = %s", result)

// JSONB.Scan()関数
log.Printf("[DEBUG] JSONB.Scan(): input value = %+v (type: %T)", value, value)
log.Printf("[DEBUG] JSONB.Scan(): final Data = %+v (type: %T)", j.Data, j.Data)
```

## 学んだ重要なポイント

### 1. データベース型の重要性
- **PostgreSQL**: `jsonb` 型が存在する
- **SQLite（Turso）**: `jsonb` 型は存在せず、`text` 型でJSONを保存する必要がある
- データベース移行時は型の互換性を必ず確認すること

### 2. driver.Valuerインターフェースの返り値
- SQLiteでは返り値の型によってデータベース内での保存形式が決まる
- `[]byte` → BLOB型
- `string` → TEXT型

### 3. TursoでのJSON扱い
- Tursoでは `jsonb` 型は使用できない
- JSON データは `TEXT` 型で保存し、アプリケーション側でシリアライゼーション/デシリアライゼーションを行う

### 4. デバッグの重要性
- データベースへの保存と読み取りの両方にログを追加することで、問題箇所を特定できる
- `typeof()` 関数でデータベース内の実際の型を確認することが重要

## 今後の注意事項

### 1. 新しいJSONBフィールド追加時
- 必ず `gorm:"type:text"` を指定すること
- PostgreSQL環境では `gorm:"type:jsonb"` を使用可能だが、Turso環境では `text` を使用すること

### 2. データベース移行時のチェック項目
- [ ] スキーマ定義の型が対象データベースで使用可能か確認
- [ ] `PRAGMA table_info(table_name)` でスキーマを確認
- [ ] `typeof(column_name)` で実際のデータ型を確認

### 3. 環境別設定の検討
将来的にはデータベース種別によって型を切り替える仕組みを検討：

```go
// 環境別型定義の例
var jsonFieldType string
if dbType == "postgres" {
    jsonFieldType = "jsonb"
} else {
    jsonFieldType = "text"
}
```

## テスト結果

修正後のテスト結果：
- ✅ プロフィール編集でゲーム選択・保存が正常動作
- ✅ Tursoデータベースで正しいJSON配列として保存
- ✅ データベース型が `text` として正しく設定
- ✅ 画面での表示・編集が正常動作

## 関連Issue・改善点

今回の修正で以下の副次的な問題も解決：
- プロフィール表示での LocalStorage キャッシュが正しいデータを反映
- プロフィール編集での初期値設定が正常動作

## 実装完了の確認事項

- [x] お気に入りゲームの保存・表示
- [x] プレイ時間帯の保存・表示
- [x] データベーススキーマの一貫性
- [x] Turso環境での動作確認
- [x] デバッグログによる問題特定手法の確立