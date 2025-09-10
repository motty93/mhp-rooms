# Googleサイトリンク対策

**実装時間**: 約30分  
**実装日**: 2025年6月24日

## 概要

Google検索結果で複数のリンク（サイトリンク）を表示させるためのSEO対策を実装しました。これにより、リピーターが部屋一覧やログインページに直接アクセスできるようになります。

## 実装した変更

### 1. 構造化データの拡張（SiteNavigationElement）

**ファイル**: `/templates/layouts/base.html`

#### 追加した構造化データ
```json
{
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "WebSite",
      "@id": "https://adpahub.com/#website",
      "url": "https://adpahub.com/",
      "name": "アドパHub",
      "description": "PSPゲームのアドホックパーティを簡単に作成・参加できるサービス",
      "inLanguage": "ja"
    },
    {
      "@type": "SiteNavigationElement",
      "@id": "https://adpahub.com/#navigation",
      "name": ["ホーム", "部屋一覧", "ログイン", "新規登録", "お問い合わせ", "利用規約", "プライバシーポリシー"],
      "url": ["https://adpahub.com/", "https://adpahub.com/rooms", "https://adpahub.com/auth/login", "https://adpahub.com/auth/register", "https://adpahub.com/contact", "https://adpahub.com/terms", "https://adpahub.com/privacy"]
    }
  ]
}
```

### 2. XMLサイトマップの実装

**新規ファイル**: `/internal/handlers/sitemap.go`

#### 機能
- 動的なXMLサイトマップ生成
- 優先度設定による重要ページの明示
- 更新頻度の設定

#### 優先度設定
- ホーム: 1.0（最高）
- 部屋一覧: 0.9
- ログイン/新規登録: 0.8
- お問い合わせ: 0.7
- 利用規約/プライバシーポリシー: 0.5

### 3. robots.txtの作成

**新規ファイル**: `/static/robots.txt`

#### 内容
- 全クローラーを許可
- 管理画面やAPIエンドポイントを除外
- サイトマップの場所を明記
- クロール遅延を1秒に設定

### 4. ルーティングの追加

**ファイル**: `/cmd/server/main.go`
- `/sitemap.xml`エンドポイントを追加

## 技術的な実装詳細

### 構造化データの利点
1. **@graph形式**: 複数の構造化データを関連付け
2. **@id属性**: 各要素を一意に識別
3. **SiteNavigationElement**: ナビゲーション構造を明示

### XMLサイトマップの構造
```xml
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://adpahub.com/</loc>
    <lastmod>2025-06-24</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <!-- 他のURL -->
</urlset>
```

## 期待される効果

### Google検索での表示改善
1. **サイトリンク表示**: 最大6つのサブリンクが表示可能
2. **直接アクセス**: リピーターが目的のページに直接移動
3. **検索結果の占有率向上**: より多くの画面領域を占有

### SEO効果
1. **クロール効率向上**: サイトマップによる効率的なインデックス
2. **構造理解の促進**: 構造化データによるサイト構造の明確化
3. **ユーザビリティ向上**: 目的のページへの到達時間短縮

## 今後の対応

### Google Search Consoleでの設定
1. サイトマップの送信
2. インデックス状況の監視
3. サイトリンクの表示確認

### 追加の最適化
1. パンくずリスト（BreadcrumbList）の実装
2. ページ速度の最適化
3. 内部リンク構造の強化

## 学んだ点・工夫した点

1. **構造化データの@graph形式**: 複数の構造化データを効率的に関連付け

2. **動的サイトマップ生成**: ハードコードではなく動的生成で保守性向上

3. **優先度の戦略的設定**: ユーザー行動を考慮した優先度付け

4. **robots.txtの適切な設定**: クロール効率とサーバー負荷のバランス

この実装により、Google検索でのサイトリンク表示が期待でき、リピーターユーザーの利便性が大幅に向上します。