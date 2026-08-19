# お知らせ機能 MVP（Issue #68）

**実装時間**: 約 45 分（調査・設計 15 分 / 実装 25 分 / 検証・ログ 5 分、16:25〜17:10 頃）

## 概要

ヘッダーのベルアイコン（スマホはメニュー内の「お知らせ」）から、自分宛のお知らせと更新情報をまとめて確認できる機能の MVP を実装しました。未読件数をバッジ表示し、パネルを開くと既読になります。お知らせは「作成した部屋が自動削除された」「部屋から退出させられた（キック）」「フォローされた」と「更新情報の公開」で届きます。リアルタイム反映は次フェーズです。

## 設計判断

### スコープ（ユーザーと合意）
推奨 MVP（ベル + 未読バッジ + パネル、更新情報の通知、個人宛 3 種、既読処理）で進め、SSE によるリアルタイム更新・部屋参加通知・通知設定は次フェーズに回しました。実ユーザーがいる今、「何が変わったか / 自分に何が起きたか」を伝える土台を早く出すことを優先しています。

### データの持ち方
- 個人宛は `notifications` テーブル（受信者・種類・タイトル・本文・リンク・操作者・既読日時）
- 更新情報は配信時に全ユーザーへ行を作らず、既存の `content/info`（`articles.json`）を読み、**ユーザーごとの「最後に既読にした日時」より新しいものを未読**として扱う。既読日時は `user_notification_states`（user_id 主キー）に保存
- `users` にカラムを足さなかった理由: リポジトリに `Save(user)` する箇所があり、カラムを限定して SELECT したユーザーを保存すると新カラムが NULL で上書きされる恐れがあるため
- 初回（既読日時なし）はユーザーの登録日時を基準にし、登録前の更新情報は未読に数えない。パネルに載せる更新情報は最新 5 件まで

### 取得と既読のタイミング
- ページ表示時（認証確定後）に `GET /api/notifications` で未読数と一覧を取得（Alpine ストア `notifications` が `Alpine.effect` で認証状態を監視）
- ベル/メニューを開いたときに再取得し、未読があれば `POST /api/notifications/read` で全件既読。バッジはすぐ消し、一覧上の未読ハイライトは次回取得まで残す（「いま何が新しかったか」が分かるように）

### 採用案が破綻しうる変更
- 通知量が増えて「個別既読」「ページング」が必要になったら、`/api/notifications/{id}/read` とカーソル取得を追加する
- リアルタイム化する場合、SSE の既存 hub は部屋単位のため、ユーザー単位の購読チャネルが必要
- 更新情報の未読判定は `date` / `updated` に依存するため、過去エントリを大きく改稿した際は `updated` を更新する運用が前提

## 実装内容

### 変更ファイル

| ファイル | 変更 |
|---|---|
| `internal/models/notification.go` (新規) | `Notification` / `UserNotificationState`、種類定数 |
| `internal/repository/notification_repository.go` (新規) / `interfaces.go` / `repository.go` | `Create` / `ListByUser` / `CountUnread` / `MarkAllRead` / `GetState` / `UpsertInfoReadAt`（`ON CONFLICT DO UPDATE`） |
| `internal/infrastructure/persistence/adapter.go` / `turso/db.go` / `postgres/db.go` | AutoMigrate 対象に追加、`notifications(user_id, created_at DESC)` インデックス |
| `internal/services/notification_service.go` (新規) | `NotifyRoomAutoDismissed` / `NotifyRoomKicked` / `NotifyFollowed` |
| `internal/services/room_cleanup_service.go` | 自動削除時にホストへお知らせ |
| `internal/handlers/rooms.go`（KickMember） / `follow.go` | キック・フォロー時にお知らせ（失敗しても本処理は続行） |
| `internal/handlers/notification.go` (新規) | `GET /api/notifications`（個人宛 + 更新情報を新しい順にマージ、未読数）、`POST /api/notifications/read`。マージ処理は純粋関数 `buildNotificationOverview` |
| `cmd/server/routes.go` / `application.go` | ルーティングとハンドラー生成 |
| `static/js/notification-store.js` (新規) | Alpine ストア（取得・開閉・既読・バッジ表示） |
| `templates/components/notification_bell.tmpl` / `notification_panel.tmpl` (新規) | ベルとパネル（背景クリック / Esc で閉じる。デスクトップは右上のドロップダウン、モバイルは全幅） |
| `templates/components/header.tmpl` | 認証済みブロックを「ベル + ユーザーメニュー」の flex に変更 |
| `templates/layouts/base.tmpl` | ストアの読み込み、パネルの配置、モバイルメニューに「お知らせ」項目（バッジ付き） |
| `internal/view/render.go` | ページ描画の parse 対象に 2 コンポーネントを追加 |
| `content/info/2026-08-18-august-update.md` | 8 月の更新情報に「お知らせ機能を追加」を追記 |
| `internal/handlers/notification_test.go` / `profile_tabs_render_test.go` | マージ・未読判定・上限のテスト、レイアウトにベル/パネル/ストア/モバイル項目が描画されるテスト |

### ブランチについて
キック機能（PR #181、未マージ）のお知らせを組み込むため、このブランチは `feature/11-room-kick` の上に作成しています。#181 が main にマージされれば、PR の差分はお知らせ分だけになります（両機能は同じリリースで出す前提）。

## 特に注意した点・工夫した点

- **ヘッダーの UI ルール遵守**: モバイル（768px 未満）のヘッダーにはベルを出さず、ハンバーガーメニュー内の項目にした（CLAUDE.md の仕様）
- **部屋詳細ページ**: 専用レイアウト（ヘッダーなし）のためベルは表示されない。必要なら次フェーズで部屋詳細のヘッダーに追加する
- **ID 使用ルール**: 通知のリンク・操作者は主キー `User.ID`
- **環境別分岐なし**
- **検証しやすさ**: マージ/未読判定を DB に依存しない純粋関数に切り出してテスト。`ON CONFLICT` などの生成 SQL は GORM の DryRun で libSQL 向けに正しいことを確認

## テスト・動作確認

- `go build ./...` / `go vet ./...` / `gofmt -l`: 問題なし
- `go test ./...`: `buildNotificationOverview`（順序・未読判定・下書き/カテゴリ除外・件数上限・未読数の数え方）、レイアウト描画（ベル・パネル・ストア・モバイル項目）ほか ok
- `node --check` / `prettier --check`: `notification-store.js`、変更テンプレート OK
- GORM DryRun で `UPSERT` / `MARK_READ` / `LIST` の SQL を確認
- DB への実書き込み（マイグレーション含む）はローカルからは未実施。デプロイ時に `RUN_MIGRATION=true` でテーブルが作成される
- ブラウザでの手動確認は未実施。ステージング反映後に確認したい点:
  1. ログイン後、ヘッダーのベルに未読数が出る（初回は登録日以降の更新情報が未読になる）
  2. ベルを押すとパネルが開き、更新情報が並ぶ。閉じて再表示するとバッジが消えている
  3. 別ユーザーにフォローされる / 部屋からキックされると、その旨のお知らせが届く
  4. スマホ幅ではメニュー内の「お知らせ」から同じパネルが開く

## 今後の作業・改善点

- SSE でバッジをリアルタイム更新（ユーザー単位の購読）
- 部屋参加の通知（ホスト向け）、通知の個別既読、通知設定
- 部屋詳細ページのヘッダーにもベルを表示
- キック時の `alert` をお知らせ + トーストに置き換える
