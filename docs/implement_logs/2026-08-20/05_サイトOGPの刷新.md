# サイト用デフォルト OGP 画像の刷新

**実装時間**: 約30分（2026-08-20）

## 概要

トップページなど独自 OGP を持たないページの共有カード（`{prefix}/ogp/og_image.png`）を刷新した。従来は白背景に小さなロゴだけの手動アップロード画像（リポジトリ管理外）だったが、`ogp-renderer` に**サイトモード**を追加してコードで生成するようにした。

## デザイン

部屋 OGP と同じデザイン言語（グラデーション枠 + 白カード）で、役割に合わせて別レイアウトにした（設計相談で合意済みの推奨案）。

- 枠: サイト共通のグレー（`view.GetPalette("")` の既定パレット）
- 中央: PSP アイコン（120px）+「HuntersHub」（84px）
- タグライン: 「モンハンシリーズのパーティ募集掲示板」（グレー）
- 対応タイトルのバッジ列: MHP / MHP2 / MHP2G / MHP3 / MHXX を各ゲームのテーマ色（`GameVersionPalettes` の BottomColor）のピルで表示

## 実装のポイント

- `OGP_TARGET=site` でサイトモード起動。**DB 接続・ROOM_ID 不要**（部屋モードのフローに入る前に分岐して return）
- `saveToLocal` / `uploadToGCS` の引数を `roomID` から相対オブジェクトパスに汎用化し、両モードで共用
- **キャッシュ制御を用途別に**: 部屋 OGP は従来どおり `immutable`（部屋ごとに URL が違う）、`og_image.png` は同一 URL を上書きするため `public, max-age=3600`
- `make generate-site-ogp` でローカル出力（`tmp/images/dev/ogp/og_image.png`）、環境反映は `gcloud run jobs execute ogp-renderer[-stg] --update-env-vars OGP_TARGET=site --wait`（docs/deploy.md に手順を追記）

## テスト・確認

- `cmd/ogp-renderer/main_test.go`: アセット込みで描画でき 1200x630 で出力されることのスモークテスト
- ローカル生成した画像を目視確認（枠・ロゴ・タグライン・5 色バッジの配置）
- `go build ./...` / `go test ./...` パス

## リリース後の作業

デプロイだけでは画像は変わらない。stg / 本番それぞれで Job を 1 回実行して GCS の `og_image.png` を差し替える（手順は docs/deploy.md）。SNS 側のキャッシュは Card Validator 等で再取得させる。

## 更新情報

「サービス改善アップデート - 2026年8月」に「🖼️ SNS 共有時のサイトカードを刷新」を追記し、summary を更新した。
