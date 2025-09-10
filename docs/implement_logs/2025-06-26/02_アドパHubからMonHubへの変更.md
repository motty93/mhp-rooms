# アドパHubからMonHubへの変更

## 実装期間
2025年6月26日 19:30 - 19:50（約20分）

## 実装内容

現在のコードベース内の「アドパHub」を「MonHub（モンハブ）」に変更しました。

### 変更ファイル一覧

1. **HTMLテンプレート**
   - `templates/layouts/base.html`
     - タイトルタグ
     - メタタグ（keywords, author, og:title, og:site_name）
     - 構造化データ（JSON-LD）
     - モバイルメニューのロゴと名称
   
   - `templates/pages/home.html`
     - メタタグ（description, keywords, og:title）
     - 構造化データ（JSON-LD）
     - 本文中のサービス名（特徴、FAQ、CTA）
   
   - `templates/components/header.html`
     - ヘッダーロゴのalt属性
     - サイト名表示
   
   - `templates/components/footer.html`
     - フッターのサービス名
     - コピーライト表記

2. **Goファイル**
   - `internal/handlers/sitemap.go`
     - サイトマップ生成のベースURL

### 変更内容の詳細

#### サービス名
- 「アドパHub」→「MonHub」
- 「アドパハブ」→「モンハブ」

#### URL（構造化データ内）
- `https://adpahub.com` → `https://monhub.com`

### 確認事項

- JavaScriptファイル内に「アドパHub」の参照なし
- CSSファイル内に「アドパHub」の参照なし
- READMEファイルは変更不要（プロジェクト名として残す）

## 工夫した点

1. **SEO対応**
   - 構造化データ内のURL参照も含めて全て変更
   - メタタグの内容を適切に更新

2. **一貫性の確保**
   - 日本語表記「モンハブ」も合わせて変更
   - altタグやaria-labelなども漏れなく変更

3. **段階的な確認**
   - HTMLテンプレート→Goファイル→その他の順で確認
   - grepコマンドで最終確認を実施

## 今後の作業

1. ロゴ画像の更新（必要に応じて）
2. favicon等のブランディング素材の更新
3. 実際のドメイン取得後の本番環境設定

## まとめ

計画書作成段階での決定に基づき、実際のコードベースを「MonHub」に変更しました。これにより、将来的な3DS/WiiU展開に向けた準備が整いました。