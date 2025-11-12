# 部屋一覧フィルタリングSQL発行の経緯調査

## 調査日時
2025年11月12日

## 調査の背景

部屋一覧ページのゲーム絞り込みselect boxを操作すると、以下のSQLが発行されることが判明した：

```sql
SELECT * FROM `game_versions` WHERE code = "MHP" ORDER BY `game_versions`.`id` LIMIT 1

SELECT rooms.id, rooms.room_code, ...
FROM rooms
LEFT JOIN game_versions gv ON rooms.game_version_id = gv.id
...
WHERE rooms.is_active = true AND rooms.game_version_id = "19f4d072-3627-424a-9aaa-4da1a519a578"
ORDER BY ... LIMIT 20 OFFSET 0

SELECT count(*) FROM `rooms`
WHERE is_active = true AND game_version_id = "19f4d072-3627-424a-9aaa-4da1a519a578"
```

初期実装時はクライアント側フィルタリングでSQLが発生しない想定だったため、経緯を調査した。

## 実装の変遷

### 1. 初期実装（2025年6月26日）

**実装ログ:** `docs/implement_logs/2025-06-26/05_部屋一覧ハイブリッドフィルタリング.md`

**実装内容:**
- クライアント側でフィルタリング処理
- サーバーから全部屋データを取得し、JavaScriptでフィルタリング
- ゲームバージョン選択時にSQLは発生しない

```javascript
// クライアント側フィルタリング（初期実装）
filterRooms(gameVersion) {
  this.activeFilter = gameVersion;
  if (gameVersion === '') {
    this.filteredRooms = this.allRooms;
  } else {
    this.filteredRooms = this.allRooms.filter(room => room.gameVersion.code === gameVersion);
  }
  // URLを更新（履歴を追加せずに）
  window.history.replaceState({}, '', url);
}
```

**メリット:**
- レスポンスの高速化（ネットワーク遅延なし）
- サーバー負荷の軽減
- SQLクエリの削減

### 2. ページネーション実装（2025年10月14日）

**コミット:** `d135ad1e0b1e32f3712b9548fd819842f4337d02`
**実装ログ:** `docs/implement_logs/2025-10-14/01_部屋一覧ページネーション実装.md`

**変更内容:**
- サーバーサイドページネーション（offset/limit方式）に変更
- フィルタリング処理もサーバー側に移行
- select box選択時にページ全体をリロード

```javascript
// サーバー側フィルタリング（ページネーション実装後）
filterRooms(gameVersion) {
  this.activeFilter = gameVersion;
  // フィルタ変更時はサーバー側でフィルタリングを行うため、ページをリロード
  const url = new URL(window.location);
  if (gameVersion) {
    url.searchParams.set('game_version', gameVersion);
  } else {
    url.searchParams.delete('game_version');
  }
  // ページ番号をリセット
  url.searchParams.delete('page');
  window.location.href = url.toString();  // ← ページ全体をリロード
}
```

**変更理由:**
- ページネーション機能を追加するにあたり、サーバー側でoffset/limitを管理する必要があった
- クライアント側で全データを保持しない設計に変更

**影響:**
- フィルタリング時にSQLが発生するようになった
- ページリロードが発生し、ユーザー体験が若干低下

## 現在の実装仕様

### サーバー側（`internal/handlers/rooms.go`）

```go
func (h *RoomHandler) Rooms(w http.ResponseWriter, r *http.Request) {
    filter := r.URL.Query().Get("game_version")
    page := 1
    perPage := 20

    // ページ番号の計算
    offset := (page - 1) * perPage

    // ゲームバージョンフィルタの処理
    var gameVersionID *uuid.UUID
    if filter != "" {
        gv, err := h.repo.GameVersion.FindGameVersionByCode(filter)  // SQL発行
        if err == nil {
            gameVersionID = &gv.ID
        }
    }

    // 部屋データ取得（ページング + フィルタリング）
    roomsWithJoinStatus, err := h.repo.Room.GetActiveRoomsWithJoinStatus(
        &dbUser.ID, gameVersionID, perPage, offset)  // SQL発行

    // 総件数取得
    total, err := h.repo.Room.CountActiveRooms(gameVersionID)  // SQL発行
}
```

### クライアント側（`templates/pages/rooms.tmpl`）

```javascript
init() {
    // サーバーサイドレンダリングのデータを使用
    const roomsData = {{.PageData.Rooms | json}};  // ← 20件のみ受け取る
    this.allRooms = (roomsData || []).map(room => ({ ... }));

    // サーバー側でフィルタリング済みのため、そのまま表示
    this.filteredRooms = this.allRooms;
}
```

**データの流れ:**
1. サーバーから20件のみ取得（全26件中）
2. 総件数（26件）は別途取得
3. クライアント側には表示中の20件のデータのみ存在
4. フィルタ変更時はサーバーに再リクエスト

## なぜクライアント側フィルタリングに戻さないのか

### 検討した案

**案1: 全データをクライアントに渡す**
- サーバー側で全件取得し、クライアントでフィルタリング・ページネーション
- SQLは初回のみ
- フィルタリング・ページ切替が高速

**案2: サーバーサイドページネーション維持（現状）**
- フィルタ・ページ切替時にSQLが発生
- データ量が増えても対応可能

### 現状維持の理由

1. **スケーラビリティ:**
   - 将来的に部屋数が数百〜数千件に増加する可能性がある
   - 全データをクライアントに渡すとパフォーマンス問題が発生

2. **メモリ効率:**
   - クライアント側のメモリ使用量を抑制
   - 低スペック端末でも安定動作

3. **実装の安定性:**
   - 既にページネーションが実装済みで動作している
   - リスクを取って変更する必要性が低い

## 今後の改善案

### 短期的な改善
- フィルタ変更時の`window.location.href`によるページリロードを`fetch`による非同期リクエストに変更
- ページ遷移なしでデータを更新し、ユーザー体験を改善

### 中期的な改善
- 閾値ベースのハイブリッド実装
  - データ量が少ない場合（例: 100件以下）: クライアント側フィルタリング
  - データ量が多い場合: サーバー側フィルタリング
  - サーバーから総件数を取得して動的に切り替え

### 長期的な改善
- 無限スクロール/仮想スクロールの導入
- サーバー側での高速なフィルタリング・キャッシュ機構
- Redis等を使用したキャッシュ層の追加

## まとめ

部屋一覧のフィルタリングでSQLが発行されるのは、**2025年10月14日のページネーション実装時にサーバーサイドページネーションを採用したため**である。

初期実装（2025年6月26日）ではクライアント側フィルタリングでSQLが発生しない設計だったが、ページネーション機能追加に伴い、スケーラビリティを重視してサーバー側処理に変更された。

現在26件程度のデータ量であればクライアント側フィルタリングでも問題ないが、将来的な拡張性を考慮し、現状のサーバーサイドページネーション方式を維持する方針とした。
