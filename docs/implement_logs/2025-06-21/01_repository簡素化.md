# Repository構造体の簡素化

## 実装内容

複雑化していたRepository構造体を簡素化し、メンテナンス性を向上させた。

### 変更前
```go
type Repository struct {
    User        UserRepositoryInterface
    GameVersion GameVersionRepositoryInterface  
    Room        RoomRepositoryInterface
}

// 使用例
h.repo.User.FindByID()
h.repo.GameVersion.GetActiveVersions()
h.repo.Room.GetActiveRooms()
```

### 変更後
```go
type Repository struct {
    db *database.DB
}

// 使用例
h.repo.FindUserByID()
h.repo.GetActiveGameVersions()
h.repo.GetActiveRooms()
```

## 主な変更

1. **ネストした構造体の削除**
   - UserRepository、GameVersionRepository、RoomRepositoryを削除
   - Repository構造体に直接メソッドを実装

2. **メソッドの統合**
   - `internal/repository/repository.go`に全メソッドを統合
   - 関連ファイル（user.go、game_version.go、room.go）を削除

3. **ハンドラーでの呼び出し修正**
   - `h.repo.GameVersion.GetActiveVersions()` → `h.repo.GetActiveGameVersions()`
   - `h.repo.GameVersion.FindByCode()` → `h.repo.FindGameVersionByCode()`
   - `h.repo.Room.GetActiveRooms()` → `h.repo.GetActiveRooms()`

## メリット

- 構造がシンプルになり理解しやすい
- ファイル数の削減
- メソッド呼び出しが簡潔
- 保守性の向上

## 影響範囲

- `internal/repository/` ディレクトリ
- `internal/handlers/rooms.go`
- Repository を使用する全ハンドラー