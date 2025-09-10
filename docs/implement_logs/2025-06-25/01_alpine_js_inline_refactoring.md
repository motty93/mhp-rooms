# Alpine.js インライン化リファクタリング

実装時間: 2025年6月25日 15:30 - 15:45 (約15分)

## 実装概要

Alpine.jsのコンポーネントをimport形式から各ページへの直書き形式に変更しました。

## 変更内容

### 1. JavaScriptコンポーネントの直書き化

以下のページでAlpine.jsコンポーネントを直書き形式に変更:

- `/templates/pages/rooms.html` - 部屋一覧ページ
  - rooms()コンポーネントを`<script>`タグ内に直接記述
  - 部屋参加モーダル、ログイン案内モーダルの制御ロジックを含む

- `/templates/pages/contact.html` - お問い合わせページ
  - contactForm()コンポーネントを`<script>`タグ内に直接記述
  - フォームバリデーション、送信処理を含む

- `/templates/pages/complete_profile.html` - プロフィール補完ページ
  - completeProfile()コンポーネントを`<script>`タグ内に直接記述
  - PSN IDバリデーション、プロフィール更新処理を含む

### 2. グローバルストアの統合

`/templates/layouts/base.html`にグローバルストアを直接記述:

- **mobileMenu**ストア - モバイルメニューの開閉制御
- **auth**ストア - 認証状態の管理
  - localStorage連携
  - 認証チェック機能
  - ログイン/ログアウト処理

### 3. 不要なファイルの削除

以下のディレクトリ・ファイルを削除:

- `/static/js/alpine/components/` - コンポーネントディレクトリ全体
- `/static/js/alpine/stores/` - ストアディレクトリ全体
- `/static/js/alpine/app.js` - アプリケーション初期化ファイル

## 注意点

- 各ページのAlpine.jsコンポーネントは`alpine:init`イベントで登録
- import文を使用せず、すべて直接記述形式に統一
- APIクライアントやバリデーターなどの共通関数も各ページに必要に応じて直接記述

## 動作確認

基本的な動作確認項目:
- 部屋一覧ページのモーダル動作
- お問い合わせフォームのバリデーション
- プロフィール補完ページの入力チェック
- モバイルメニューの開閉
- 認証状態の管理