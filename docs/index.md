# MHP Rooms ドキュメント目次

このディレクトリにはMHP Roomsプロジェクトの全ての技術文書が含まれています。

## 📋 主要設計文書

### システム設計
- [**アーキテクチャ設計**](./architecture.md) - システム全体のアーキテクチャと技術スタック
- [**API設計書**](./api-design.md) - REST APIエンドポイントの仕様
- [**ルーティング設計**](./routing_design.md) - URLルーティングの設計方針

### データベース設計
- [**DBスキーマ設計**](./db-schema.md) - データベーススキーマの詳細設計
- [**ER図**](./er.md) - エンティティリレーションシップ図
- [**JSONB実装仕様**](./jsonb-implementation.md) - PostgreSQLのJSONB型活用方法

### 機能別設計書
- [**プロフィール実装状況**](./profile-implementation-status.md) - プロフィール機能の実装状況
- [**アクティビティシステム設計**](./activity-system-design.md) - ユーザーアクティビティ追跡システム
- [**部屋メッセージSSE設計**](./room-message-sse-design.md) - リアルタイムメッセージングのSSE実装

### 技術仕様書
- [**LocalStorageキャッシュシステム**](./localstorage-cache-system.md) - クライアントサイドキャッシュの仕組み
- [**ブラウザログインプロセス**](./browser-login-process.md) - 認証フローの詳細
- [**インフラ構成**](./infra.md) - デプロイメントとインフラストラクチャ

### 開発環境
- [**開発環境仕様**](./development-environment.md) - 開発環境固有の機能と制限事項

## 📝 実装ログ（日付順）

実装ログは日付別に整理されており、各機能の実装過程を詳細に記録しています。

### 2025年6月 - プロジェクト基盤構築

#### 2025-06-20 - 初期設定・ドキュメント整理
- [CLAUDE.md整理](./implement_logs/2025-06-20/01_CLAUDE-md整理.md)
- [フロントエンド技術スタック記載追加](./implement_logs/2025-06-20/02_フロントエンド技術スタック記載追加.md)
- [README更新](./implement_logs/2025-06-20/03_README更新.md)
- [実装ログ再整理](./implement_logs/2025-06-20/04_実装ログ再整理.md)
- [getEnv汎用化リファクタリング](./implement_logs/2025-06-20/05_getEnv汎用化リファクタリング.md)
- [config.go形式への修正](./implement_logs/2025-06-20/06_config.go形式への修正.md)
- [wait.go修正](./implement_logs/2025-06-20/07_wait.go修正.md)
- [config.Init重複修正](./implement_logs/2025-06-20/08_config.Init重複修正.md)

#### 2025-06-21 - データベース・アーキテクチャ基盤
- [godotenv追加とマイグレーション修正](./implement_logs/2025-06-21/01_godotenv追加とマイグレーション修正.md)
- [repository簡素化](./implement_logs/2025-06-21/01_repository簡素化.md)
- [Neonデータベース対応](./implement_logs/2025-06-21/02_Neonデータベース対応.md)
- [マイグレーション制御機能](./implement_logs/2025-06-21/02_マイグレーション制御機能.md)
- [modelsテーブル分割](./implement_logs/2025-06-21/03_modelsテーブル分割.md)
- [部屋パスワード・人数制限機能](./implement_logs/2025-06-21/03_部屋パスワード・人数制限機能.md)
- [models自明コメント削除](./implement_logs/2025-06-21/04_models自明コメント削除.md)
- [実装ログファイル名形式変更](./implement_logs/2025-06-21/05_実装ログファイル名形式変更.md)
- [依存性注入パターンへのリファクタリング](./implement_logs/2025-06-21/06_依存性注入パターンへのリファクタリング.md)
- [自明コメント削除](./implement_logs/2025-06-21/07_自明コメント削除.md)

#### 2025-06-22 - フロントエンド基盤・UI改善
- [フッターゲームリンク機能](./implement_logs/2025-06-22/01_フッターゲームリンク機能.md)
- [部屋参加モーダル機能](./implement_logs/2025-06-22/02_部屋参加モーダル機能.md)
- [モーダルレイアウト調整とアニメーション実装](./implement_logs/2025-06-22/03_モーダルレイアウト調整とアニメーション実装.md)
- [Alpine.jsへの移行](./implement_logs/2025-06-22/04_Alpine.jsへの移行.md)
- [リポジトリの分離](./implement_logs/2025-06-22/05_リポジトリの分離.md)
- [handlersディレクトリの整理](./implement_logs/2025-06-22/06_handlersディレクトリの整理.md)
- [テンプレート構造の整理](./implement_logs/2025-06-22/07_テンプレート構造の整理.md)

#### 2025-06-23 - 認証システム・部屋機能
- [auth.jsからAlpine.jsへの移行](./implement_logs/2025-06-23/01_auth.jsからAlpine.jsへの移行.md)
- [room_membersダミーデータ作成](./implement_logs/2025-06-23/01_room_membersダミーデータ作成.md)
- [ルーム開閉機能実装](./implement_logs/2025-06-23/02_ルーム開閉機能実装.md)
- [statusカラム削除](./implement_logs/2025-06-23/03_statusカラム削除.md)

#### 2025-06-24 - SEO対応・ブランディング
- [法的ページとお問い合わせ機能](./implement_logs/2025-06-24/01_法的ページとお問い合わせ機能.md)
- [SEO対応とサービス名変更](./implement_logs/2025-06-24/02_SEO対応とサービス名変更.md)
- [トップページLP化](./implement_logs/2025-06-24/03_トップページLP化.md)
- [Googleサイトリンク対策](./implement_logs/2025-06-24/04_Googleサイトリンク対策.md)
- [ログイン・新規登録画面作成](./implement_logs/2025-06-24/05_ログイン・新規登録画面作成.md)

#### 2025-06-25 - 機能拡張・リファクタリング
- [alpine_js_inline_refactoring](./implement_logs/2025-06-25/01_alpine_js_inline_refactoring.md)
- [ゲームバージョン別プレイヤーネーム機能](./implement_logs/2025-06-25/01_ゲームバージョン別プレイヤーネーム機能.md)

#### 2025-06-26 - サービス拡張・UX改善
- [サービス拡張戦略文書作成](./implement_logs/2025-06-26/01_サービス拡張戦略文書作成.md)
- [アドパHubからMonHubへの変更](./implement_logs/2025-06-26/02_アドパHubからMonHubへの変更.md)
- [PSP固有内容の汎用化](./implement_logs/2025-06-26/03_PSP固有内容の汎用化.md)
- [使い方ガイドページ作成](./implement_logs/2025-06-26/04_使い方ガイドページ作成.md)
- [部屋一覧ハイブリッドフィルタリング](./implement_logs/2025-06-26/05_部屋一覧ハイブリッドフィルタリング.md)
- [rooms.html改善](./implement_logs/2025-06-26/06_rooms.html改善.md)

#### 2025-06-27 - 認証システム実装
- [handler接尾辞削除](./implement_logs/2025-06-27/01_handler接尾辞削除.md)
- [supabase-go認証実装](./implement_logs/2025-06-27/01_supabase-go認証実装.md)
- [インフラストラクチャ層リファクタリング](./implement_logs/2025-06-27/02_インフラストラクチャ層リファクタリング.md)
- [auth-go調査とgotrue-go継続使用](./implement_logs/2025-06-27/03_auth-go調査とgotrue-go継続使用.md)
- [auth_testエラー修正](./implement_logs/2025-06-27/04_auth_testエラー修正.md)
- [認証機能の実装状況確認](./implement_logs/2025-06-27/05_認証機能の実装状況確認.md)
- [player_names_table_refactoring](./implement_logs/2025-06-27/06_player_names_table_refactoring.md)

#### 2025-06-28 - アーキテクチャ改善
- [ハンドラー構造改善](./implement_logs/2025-06-28/01_ハンドラー構造改善.md)

#### 2025-06-29 - 認証システム移行
- [supabase-js認証実装](./implement_logs/2025-06-29/01_supabase-js認証実装.md)
- [supabase-go完全削除](./implement_logs/2025-06-29/02_supabase-go完全削除.md)

#### 2025-06-30 - 認証機能強化
- [headerの認証処理](./implement_logs/2025-06-30/01_headerの認証処理.md)
- [ユーザー追加処理](./implement_logs/2025-06-30/02_ユーザー追加処理.md)

### 2025年7月 - 機能拡張・パフォーマンス改善

#### 2025-07-14 - パフォーマンス最適化
- [ユーザークエリ最適化](./implement_logs/2025-07-14/01_ユーザークエリ最適化.md)
- [quest_type削除](./implement_logs/2025-07-14/02_quest_type削除.md)
- [ユーザークエリ重複処理最適化](./implement_logs/2025-07-14/03_ユーザークエリ重複処理最適化.md)
- [ヘッダー認証ボタンちらつき問題修正](./implement_logs/2025-07-14/04_ヘッダー認証ボタンちらつき問題修正.md)

#### 2025-07-21 - メッセージ機能実装
- [メッセージスタンプ機能](./implement_logs/2025-07-21/01_メッセージスタンプ機能.md)
- [部屋詳細画面実装](./implement_logs/2025-07-21/01_部屋詳細画面実装.md)

#### 2025-07-27 - パフォーマンス改善
- [トップページ_パフォーマンス改善](./implement_logs/2025-07-27/01_トップページ_パフォーマンス改善.md)

#### 2025-07-28 - メッセージ機能完成
- [部屋詳細メッセージ送信機能の完全修正](./implement_logs/2025-07-28/00_部屋詳細メッセージ送信機能の完全修正.md)
- [メッセージ機能修正とログ整理](./implement_logs/2025-07-28/01_メッセージ機能修正とログ整理.md)
- [部屋詳細機能実装](./implement_logs/2025-07-28/01_部屋詳細機能実装.md)
- [メッセージ送信機能実装](./implement_logs/2025-07-28/02_メッセージ送信機能実装.md)

#### 2025-07-29 - 部屋管理機能
- [部屋退出機能の実装](./implement_logs/2025-07-29/01_部屋退出機能の実装.md)

#### 2025-07-30 - バグ修正
- [パスワード付き部屋への参加エラー修正](./implement_logs/2025-07-30/01_パスワード付き部屋への参加エラー修正.md)

### 2025年8月 - 本格機能実装・最適化

#### 2025-08-04 - UX改善・パフォーマンス
- [参加中部屋への導線追加](./implement_logs/2025-08-04/01_参加中部屋への導線追加.md)
- [部屋一覧クエリ最適化](./implement_logs/2025-08-04/04_部屋一覧クエリ最適化.md)
- [router_migration_and_query_optimization](./implement_logs/2025-08-04/05_router_migration_and_query_optimization.md)

#### 2025-08-08 - 技術的修正
- [current_room_api重複呼び出し修正](./implement_logs/2025-08-08/01_current_room_api重複呼び出し修正.md)
- [GORM_v1.30.1対応修正](./implement_logs/2025-08-08/02_GORM_v1.30.1対応修正.md)

#### 2025-08-09 - 管理機能実装
- [ユーザーブロック機能実装](./implement_logs/2025-08-09/01_ユーザーブロック機能実装.md)
- [部屋作成モーダル機能実装](./implement_logs/2025-08-09/01_部屋作成モーダル機能実装.md)

#### 2025-08-11 - セキュリティ機能
- [部屋入室制限機能](./implement_logs/2025-08-11/01_部屋入室制限機能.md)

#### 2025-08-12 - プロフィール機能
- [プロフィール画面実装](./implement_logs/2025-08-12/01_プロフィール画面実装.md)
- [グローバル部屋作成モーダル実装](./implement_logs/2025-08-12/02_グローバル部屋作成モーダル実装.md)

#### 2025-08-13 - アクティビティシステム
- [アクティビティシステム実装](./implement_logs/2025-08-13/01_アクティビティシステム実装.md)
- [プロフィール画面タブ修正](./implement_logs/2025-08-13/01_プロフィール画面タブ修正.md)
- [作成した部屋機能実装](./implement_logs/2025-08-13/01_作成した部屋機能実装.md)
- [プロフィール画面作成した部屋修正](./implement_logs/2025-08-13/02_プロフィール画面作成した部屋修正.md)
- [プロフィールAPI開発環境モックデータ修正](./implement_logs/2025-08-13/03_プロフィールAPI開発環境モックデータ修正.md)

#### 2025-08-14 - バグ修正
- [record_not_found_fix](./implement_logs/2025-08-14/01_record_not_found_fix.md)

### 2025年9月 - 認証・UX改善

#### 2025-09-09 - 認証システム改善
- [認証エラー時のリダイレクト実装](./implement_logs/2025-09-09/01_認証エラー時のリダイレクト実装.md)
- [認証切れ時のフロントエンド処理](./implement_logs/2025-09-09/02_認証切れ時のフロントエンド処理.md)

#### 2025-09-10 - UX最終調整
- [認証状態保持とDisplayName表示修正](./implement_logs/2025-09-10/01_認証状態保持とDisplayName表示修正.md)
- [部屋参加時の確認モーダル修正](./implement_logs/2025-09-10/01_部屋参加時の確認モーダル修正.md)
- [ヘッダーリンクの無効化実装](./implement_logs/2025-09-10/02_ヘッダーリンクの無効化実装.md)
- [ヘッダーアクティブ状態のUX改善](./implement_logs/2025-09-10/03_ヘッダーアクティブ状態のUX改善.md)

## 🔍 パフォーマンス解析ログ

`implement_logs/queries/`ディレクトリには、パフォーマンス問題の調査と最適化の記録が含まれています：

### アクティブクエリ解析
- [2025-07-14 部屋クエリ分析](./implement_logs/queries/2025-07-14-135730_rooms_query.md)
- [2025-07-27 トップページ分析](./implement_logs/queries/2025-07-27-230627_top_page.md)
- [2025-08-03 部屋詳細分析](./implement_logs/queries/2025-08-03-153515_room_show.md)
- [2025-08-04 部屋一覧分析](./implement_logs/queries/2025-08-04_rooms.md)
- [2025-08-08 部屋退出分析](./implement_logs/queries/2025-08-08_rooms_left.md)
- [2025-08-13 部屋エラー分析](./implement_logs/queries/2025-08-13_rooms-error.md)
- [2025-08-14 部屋クエリ分析](./implement_logs/queries/2025-08-14_rooms-queries.md)

### アーカイブ
- [archived/](./implement_logs/queries/archived/) - 過去のパフォーマンス解析記録

## 🚀 ドキュメントの使い方

### 新規開発者向け
1. [アーキテクチャ設計](./architecture.md)でシステム全体を把握
2. [開発環境仕様](./development-environment.md)で環境構築
3. [API設計書](./api-design.md)でAPIエンドポイントを確認

### 機能開発時
1. 関連する設計文書を確認
2. 実装ログで類似機能の実装方法を参考
3. 実装後は必ず`implement_logs/`に実装ログを記録

### トラブルシューティング
1. `implement_logs/queries/`でパフォーマンス問題を調査
2. 実装ログから類似問題の解決方法を検索
3. [開発環境仕様](./development-environment.md)で環境固有の問題を確認

## 📚 ドキュメント更新指針

- **設計文書**: 機能追加・変更時に必ず更新
- **実装ログ**: 各実装完了時に必ず作成（`yyyy-mm-dd/nn_機能名.md`形式）
- **クエリ解析**: パフォーマンス問題発生時に記録
- **目次**: 新しいドキュメント追加時に更新

---

**最終更新**: 2025-09-10  
**ドキュメント数**: 100+ files  
**プロジェクト開始**: 2025-06-16