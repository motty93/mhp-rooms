# Wappalyzer WordPress誤検出修正

**実装時間**: 約5分

## 概要

Chrome拡張のWappalyzerでサイトがWordPressと誤検出される問題を修正しました。

## 原因

`templates/layouts/base.tmpl`の構造化データ（JSON-LD）で使用していた `@id` のフラグメント識別子パターンが、WordPressのYoast SEOプラグインが出力する形式と同一でした。

```json
// 修正前（Yoast SEOと同じパターン）
"@id": "{{ $siteURL }}/#website"
"@id": "{{ $siteURL }}/#organization"
"@id": "{{ $siteURL }}/#navigation"
```

Wappalyzerは `/#website`, `/#organization` というパターンをYoast SEOの特徴として検出し、Yoast SEO → WordPress という推論でWordPressと判定していました。

## 修正内容

### 変更ファイル
- `templates/layouts/base.tmpl`

### 変更箇所

構造化データの `@id` パターンを以下のように変更:

| 修正前 | 修正後 |
|--------|--------|
| `/#website` | `/schema/website` |
| `/#organization` | `/schema/organization` |
| `/#navigation` | `/schema/navigation` |

```json
// 修正後
"@id": "{{ $siteURL }}/schema/website"
"@id": "{{ $siteURL }}/schema/organization"
"@id": "{{ $siteURL }}/schema/navigation"
```

## 特に注意した点

- SEO的には `@id` の値は任意のURIで問題ないため、`/schema/xxx` 形式でも構造化データとしては有効
- Schema.orgの仕様に準拠しつつ、WordPress/Yoast SEOの検出パターンを回避

## テスト・動作確認

- デプロイ後にWappalyzerで再確認が必要
- 構造化データのバリデーション（Google Rich Results Test等）での確認を推奨

## 今後の作業

- 本番環境デプロイ後のWappalyzer検出結果確認
- 必要に応じてGoogle Search Consoleで構造化データのエラーがないか確認
