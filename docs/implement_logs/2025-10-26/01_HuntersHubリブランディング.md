# HuntersHub リブランディング実装ログ

## 実装日時
- **開始**: 2025-10-26
- **完了**: 2025-10-26
- **所要時間**: 約1.5時間

## 概要
商標権リスク回避のため、サービス名「MonHub（モンハブ）」から「HuntersHub（ハンターズハブ）」へ全面的にリブランディングを実施しました。

## 変更内容

### 1. Cloud Build設定ファイル（2ファイル）
#### cloudbuild.yml
- サービス名: `monhub` → `huntershub`
- SSEサービス名: `monhub-sse` → `huntershub-sse`
- Artifact Registry: `monhub-registry` → `huntershub-registry`
- イメージ名: `monhub-app` → `huntershub-app`
- GCSバケット: `monhub-master` → `huntershub-master`
- GCSプライベートバケット: `monhub-private` → `huntershub-private`
- ベースURL: `https://storage.googleapis.com/monhub-master` → `https://storage.googleapis.com/huntershub-master`

#### cloudbuild.stg.yml
- 同様の変更をステージング環境用に実施

### 2. テンプレートファイル（12ファイル）
#### レイアウトテンプレート
- `templates/layouts/base.tmpl`
  - ページタイトル: `MonHub` → `HuntersHub`
  - メタキーワード: `MonHub,モンハブ` → `HuntersHub,ハンターズハブ`
  - 構造化データのURL: `https://monhub.com` → `https://huntershub.com`
  - JavaScriptグローバル変数: `window.monhubAnalytics` → `window.huntershubAnalytics`

- `templates/components/header.tmpl`
  - ロゴテキスト: `MonHub` → `HuntersHub`
  - alt属性: `MonHub` → `HuntersHub`

- `templates/components/footer.tmpl`
  - フッタータイトル: `MonHub` → `HuntersHub`
  - 著作権表示: `© 2025 MonHub` → `© 2025 HuntersHub`

#### ページテンプレート（9ファイル）
以下のファイルでサービス名を一括置換：
- `templates/pages/home.tmpl`
- `templates/pages/rooms.tmpl`
- `templates/pages/room_detail.tmpl`
- `templates/pages/contact.tmpl`
- `templates/pages/guide.tmpl`
- `templates/pages/faq.tmpl`
- `templates/pages/privacy.tmpl`
- `templates/pages/terms.tmpl`
- `templates/layouts/room_detail.tmpl`

**置換内容**:
- `MonHub` → `HuntersHub`
- `モンハブ` → `ハンターズハブ`
- `monhub.com` → `huntershub.com`

### 3. JavaScriptファイル（1ファイル）
#### static/js/analytics.js
- グローバル変数参照: `window.monhubAnalytics` → `window.huntershubAnalytics`

### 4. Goソースコード（3ファイル）
#### cmd/ogp-renderer/main.go
- コメント: `MonHubアイコンサイズ` → `HuntersHubアイコンサイズ`
- コメント: `MonHub` → `HuntersHub`（フォント設定）
- 関数名: `drawMonHubLogoBottomRight` → `drawHuntersHubLogoBottomRight`
- OGP画像内テキスト: `"MonHub"` → `"HuntersHub"`
- コメント: `MonHubテキストを描画` → `HuntersHubテキストを描画`

#### internal/handlers/sitemap.go
- ベースURL: `https://monhub.com` → `https://huntershub.com`

#### internal/services/activity_service.go
- アクティビティタイトル: `"MonHubに参加しました"` → `"HuntersHubに参加しました"`

### 5. ドキュメントファイル（約20ファイル）
実装ログ（docs/implement_logs/）を除く、全てのドキュメントファイルで一括置換：
- `README.md`
- `docs/`配下の各種設計ドキュメント

**置換内容**:
- `MonHub` → `HuntersHub`
- `モンハブ` → `ハンターズハブ`
- `monhub.com` → `huntershub.com`
- `monhub-` → `huntershub-`（インフラリソース名）

## 変更対象外
以下のファイルは履歴として保持するため、変更対象外としました：
- `docs/implement_logs/` 配下の全ての実装ログ

## 次のステップ（インフラ作業）

### フェーズ2: インフラリソースの準備
```bash
# 1. 新GCSバケット作成
gsutil mb -l asia-northeast1 gs://huntershub-master
gsutil mb -l asia-northeast1 gs://huntershub-private

# 2. CORS設定コピー
gsutil cors get gs://monhub-master > cors.json
gsutil cors set cors.json gs://huntershub-master

# 3. IAM設定（公開読み取り）
gsutil iam ch allUsers:objectViewer gs://huntershub-master

# 4. データ移行
gsutil -m cp -r gs://monhub-master/* gs://huntershub-master/

# 5. Artifact Registryリポジトリ作成（任意）
gcloud artifacts repositories create huntershub-registry \
  --repository-format=docker \
  --location=asia-northeast1
```

### フェーズ3: デプロイ・切り替え
```bash
# ステージング環境デプロイ
gcloud builds submit --config=cloudbuild.stg.yml

# 本番環境デプロイ
gcloud builds submit --config=cloudbuild.yml

# カスタムドメインマッピング
gcloud run domain-mappings create \
  --service=huntershub \
  --domain=huntershub.com \
  --region=asia-northeast1

# DNS設定（huntershub.comのAレコード/CNAMEレコード設定）
```

### フェーズ4: 後処理
- 旧GCSバケット（monhub-master）は6ヶ月間保持（SNSシェア済みOGP画像対応）
- 旧Cloud Runサービスは1ヶ月後に削除予定
- monhub.com → huntershub.com の301リダイレクト設定
- 検索エンジンへのサイトマップ再送信
- Google Search Console設定更新

## テスト項目
- [ ] ステージング環境で動作確認
  - [ ] トップページ表示
  - [ ] 部屋一覧ページ表示
  - [ ] OGP画像生成（HuntersHubロゴ）
  - [ ] サイトマップURL確認
  - [ ] Google Analytics動作（huntershubAnalytics変数）
- [ ] 本番環境デプロイ後の動作確認
- [ ] DNSカットオーバー後の動作確認

## 注意事項
1. **ダウンタイム**: DNS切り替え時に数分～数十分のダウンタイムが発生する可能性
2. **既存OGP画像**: SNSシェア済みの画像URLは変更不可（旧バケット保持理由）
3. **検索エンジン**: サイトマップ再送信、Google Search Console設定更新が必要
4. **段階的実行**: ステージング→本番の順で慎重に進める

## 工夫した点
- 実装ログを除外して履歴を保持
- sedコマンドで一括置換し、効率的に作業を実施
- インフラ作業手順を明確化し、後続作業をスムーズに

## 今後の課題
- GCSバケット作成とデータ移行の実施
- Cloud Runサービスの再デプロイ
- カスタムドメインのマッピング設定
- DNS設定とカットオーバー
- 旧リソースの削除スケジュール管理
