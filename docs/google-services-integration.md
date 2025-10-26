# Google AdSense & Analytics 導入ガイド

## 目次
1. [概要](#概要)
2. [Google AdSense導入](#google-adsense導入)
3. [Google Analytics 4導入](#google-analytics-4導入)
4. [実装詳細](#実装詳細)
5. [最適化とベストプラクティス](#最適化とベストプラクティス)
6. [モニタリングとレポート](#モニタリングとレポート)

---

## 概要

### 目的
HuntersHubにGoogle AdSenseとGoogle Analytics 4を導入し、以下を実現する：
- **AdSense**: 広告収益化
- **Analytics**: ユーザー行動分析、コンバージョン追跡

### 前提条件
- [ ] Google AdSenseアカウントの取得
- [ ] Google Analytics 4プロパティの作成
- [ ] サイトの所有権確認
- [ ] AdSenseのサイト審査通過

---

## Google AdSense導入

### 1. アカウント準備

#### 1.1 AdSenseアカウント作成
1. [Google AdSense](https://www.google.com/adsense/)にアクセス
2. Googleアカウントでログイン
3. サイト情報を入力
   - URL: `https://huntershub.com`（本番URL）
   - サイトの言語: 日本語
4. 利用規約に同意

#### 1.2 サイト審査準備
審査に合格するための要件：
- [ ] 独自ドメイン設定済み
- [ ] 十分なコンテンツ量（部屋一覧、プロフィール、利用規約等）
- [ ] プライバシーポリシーの設置
- [ ] お問い合わせページの設置
- [ ] ナビゲーションが明確
- [ ] モバイル対応
- [ ] 最低3ヶ月程度の運用実績（推奨）

#### 1.3 ads.txt設置
**重要**: ads.txtファイルを設置して収益性を最大化

**ファイル**: `static/ads.txt`

```
google.com, pub-XXXXXXXXXXXXXX, DIRECT, f08c47fec0942fa0
```

**Go側のルーティング**:

```go
// cmd/server/routes.go の setupStaticRoutes に追加
r.Get("/ads.txt", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	http.ServeFile(w, r, "./static/ads.txt")
})
```

### 2. AdSenseコード実装

#### 2.1 基本設定スクリプト

**ファイル**: `templates/layouts/base.tmpl`

`<head>` タグ内に追加：

```html
<!-- Google AdSense -->
{{ if .IsProduction }}
<script async src="https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=ca-pub-XXXXXXXXXXXXXX"
     crossorigin="anonymous"></script>
{{ end }}
```

**重要**: 本番環境でのみ読み込むように条件分岐

#### 2.2 自動広告（オプション）

自動広告を有効にする場合は上記のスクリプトのみでOK。
Google AdSenseが自動的に最適な位置に広告を配置します。

**メリット**:
- 手動設置不要
- Googleが最適化

**デメリット**:
- レイアウトが崩れる可能性
- 広告位置のコントロールが難しい

#### 2.3 手動広告配置（推奨）

特定の位置に広告を配置する場合。

##### 2.3.1 ディスプレイ広告コンポーネント

**ファイル**: `templates/components/ad-unit.tmpl`（新規作成）

```html
{{ define "ad-unit" }}
<!-- Google AdSense 広告ユニット -->
<div class="ad-container my-6">
  {{ if .IsProduction }}
  <ins class="adsbygoogle"
       style="display:block"
       data-ad-client="ca-pub-XXXXXXXXXXXXXX"
       data-ad-slot="{{ .AdSlot }}"
       data-ad-format="{{ .AdFormat }}"
       {{ if .FullWidth }}data-full-width-responsive="true"{{ end }}></ins>
  <script>
       (adsbygoogle = window.adsbygoogle || []).push({});
  </script>
  {{ else }}
  <!-- 開発環境: プレースホルダー -->
  <div class="bg-gray-200 p-4 text-center text-gray-500 rounded">
    <p>広告スペース ({{ .AdFormat }})</p>
    <p class="text-sm">本番環境でのみ表示されます</p>
  </div>
  {{ end }}
</div>
{{ end }}
```

##### 2.3.2 広告配置例

**部屋一覧ページ**: `templates/pages/rooms.tmpl`

```html
{{ define "page" }}
<div class="container mx-auto px-4 py-8">
  <h1 class="text-3xl font-bold mb-6">部屋一覧</h1>
  
  <!-- 広告1: ページ上部 - 横長バナー -->
  {{ template "ad-unit" (dict 
    "IsProduction" .IsProduction 
    "AdSlot" "1111111111" 
    "AdFormat" "horizontal"
    "FullWidth" true
  ) }}
  
  <!-- ゲームバージョン選択 -->
  <div class="mb-6">
    <!-- フィルター UI -->
  </div>
  
  <!-- 部屋リスト -->
  <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
    {{ range $index, $room := .Rooms }}
      <!-- 部屋カード -->
      {{ template "room-card" $room }}
      
      <!-- 広告2: 6件ごとにインフィード広告 -->
      {{ if and (eq (mod (add $index 1) 6) 0) $.IsProduction }}
        {{ template "ad-unit" (dict 
          "IsProduction" $.IsProduction 
          "AdSlot" "2222222222" 
          "AdFormat" "fluid"
          "FullWidth" true
        ) }}
      {{ end }}
    {{ end }}
  </div>
  
  <!-- 広告3: ページ下部 - レクタングル -->
  {{ template "ad-unit" (dict 
    "IsProduction" .IsProduction 
    "AdSlot" "3333333333" 
    "AdFormat" "rectangle"
  ) }}
</div>
{{ end }}
```

**部屋詳細ページ**: `templates/pages/room_detail.tmpl`

```html
{{ define "page" }}
<div class="container mx-auto px-4 py-8">
  <div class="grid md:grid-cols-3 gap-6">
    <!-- メインコンテンツ (2/3) -->
    <div class="md:col-span-2">
      <!-- 部屋情報 -->
      <div class="bg-white rounded-lg shadow p-6">
        <!-- 部屋の詳細 -->
      </div>
      
      <!-- チャットエリア -->
      <div class="mt-6">
        <!-- チャット UI -->
      </div>
    </div>
    
    <!-- サイドバー (1/3) - デスクトップのみ -->
    <div class="hidden md:block">
      <!-- 広告: サイドバー上部 -->
      {{ template "ad-unit" (dict 
        "IsProduction" .IsProduction 
        "AdSlot" "4444444444" 
        "AdFormat" "vertical"
      ) }}
      
      <!-- 参加メンバー情報 -->
      <div class="mt-6">
        <!-- メンバーリスト -->
      </div>
    </div>
  </div>
  
  <!-- 広告: モバイル用 - コンテンツ下部 -->
  <div class="md:hidden mt-6">
    {{ template "ad-unit" (dict 
      "IsProduction" .IsProduction 
      "AdSlot" "5555555555" 
      "AdFormat" "horizontal"
      "FullWidth" true
    ) }}
  </div>
</div>
{{ end }}
```

### 3. htmxとの統合

htmxで動的に読み込むコンテンツに広告を配置する場合、再初期化が必要。

**ファイル**: `static/js/htmx-ads.js`（新規作成）

```javascript
// htmxイベント後に広告を再読み込み
document.body.addEventListener('htmx:afterSwap', (event) => {
  // 新しく追加された広告ユニットを検索
  const newAds = event.detail.target.querySelectorAll('.adsbygoogle:not([data-adsbygoogle-status])');
  
  // 各広告ユニットを初期化
  newAds.forEach((ad) => {
    try {
      (adsbygoogle = window.adsbygoogle || []).push({});
    } catch (e) {
      console.error('AdSense initialization error:', e);
    }
  });
});

// htmxでコンテンツを置き換える前に古い広告をクリーンアップ（オプション）
document.body.addEventListener('htmx:beforeSwap', (event) => {
  const oldAds = event.detail.target.querySelectorAll('.adsbygoogle');
  oldAds.forEach((ad) => {
    // 広告の状態をリセット
    ad.removeAttribute('data-adsbygoogle-status');
  });
});
```

**base.tmpl に追加**:

```html
<!-- htmx広告統合 -->
{{ if .IsProduction }}
<script src="/static/js/htmx-ads.js"></script>
{{ end }}
```

### 4. Go側のテンプレート設定

**ファイル**: `internal/handlers/base_handler.go`

```go
// BaseData にIsProductionフィールド追加
type BaseData struct {
	Title         string
	IsProduction  bool
	IsAuthenticated bool
	User          *models.User
	// ... その他のフィールド
}

// NewBaseData 関数を修正
func NewBaseData(title string) BaseData {
	return BaseData{
		Title:        title,
		IsProduction: isProductionEnv(), // 既存の関数を使用
	}
}
```

---

## Google Analytics 4導入

### 1. アカウント準備

#### 1.1 GA4プロパティ作成
1. [Google Analytics](https://analytics.google.com/)にアクセス
2. 「管理」→「プロパティを作成」
3. プロパティ情報を入力
   - プロパティ名: HuntersHub
   - タイムゾーン: 日本
   - 通貨: 日本円 (JPY)
4. データストリーム作成
   - プラットフォーム: ウェブ
   - ウェブサイトURL: `https://huntershub.com`
   - ストリーム名: HuntersHub Web
5. 測定IDをコピー（`G-XXXXXXXXXX`）

### 2. GA4コード実装

#### 2.1 基本トラッキングコード

**ファイル**: `templates/layouts/base.tmpl`

`<head>` タグ内、AdSenseの後に追加：

```html
<!-- Google Analytics 4 -->
{{ if .IsProduction }}
<!-- Google tag (gtag.js) -->
<script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());

  gtag('config', 'G-XXXXXXXXXX', {
    'cookie_flags': 'SameSite=None;Secure',
    'anonymize_ip': true, // IP匿名化（GDPR対応）
  });
</script>
{{ end }}
```

#### 2.2 カスタムイベント追跡

**ファイル**: `static/js/analytics.js`（新規作成）

```javascript
// Google Analytics カスタムイベント
const Analytics = {
  // 部屋作成イベント
  trackRoomCreate: (roomId, gameVersion, maxPlayers) => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'room_create', {
        'room_id': roomId,
        'game_version': gameVersion,
        'max_players': maxPlayers
      });
    }
  },

  // 部屋参加イベント
  trackRoomJoin: (roomId, gameVersion) => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'room_join', {
        'room_id': roomId,
        'game_version': gameVersion
      });
    }
  },

  // 部屋退出イベント
  trackRoomLeave: (roomId, duration) => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'room_leave', {
        'room_id': roomId,
        'session_duration': duration
      });
    }
  },

  // ユーザー登録イベント
  trackSignup: (method) => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'sign_up', {
        'method': method // 'email', 'google', etc.
      });
    }
  },

  // ログインイベント
  trackLogin: (method) => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'login', {
        'method': method
      });
    }
  },

  // プロフィール編集イベント
  trackProfileEdit: () => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'profile_edit');
    }
  },

  // ゲームバージョンフィルターイベント
  trackGameFilter: (gameVersion) => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'filter_game_version', {
        'game_version': gameVersion
      });
    }
  },

  // 検索イベント
  trackSearch: (searchTerm) => {
    if (typeof gtag !== 'undefined') {
      gtag('event', 'search', {
        'search_term': searchTerm
      });
    }
  }
};

// グローバルに公開
window.Analytics = Analytics;
```

**base.tmpl に追加**:

```html
<!-- Analytics カスタムイベント -->
{{ if .IsProduction }}
<script src="/static/js/analytics.js"></script>
{{ end }}
```

#### 2.3 イベント統合例

**部屋作成モーダル**: `static/js/room-create-store.js`

```javascript
// 既存のcreateRoom関数に追加
async createRoom() {
  // ... 既存の処理 ...

  try {
    const response = await fetch('/api/rooms', {
      method: 'POST',
      // ... 既存のリクエスト ...
    });

    if (response.ok) {
      const data = await response.json();
      
      // Analytics イベント送信
      if (window.Analytics) {
        Analytics.trackRoomCreate(
          data.id,
          this.formData.gameVersionId,
          this.formData.maxPlayers
        );
      }

      // ... 既存の処理 ...
    }
  } catch (error) {
    // ... エラー処理 ...
  }
}
```

**認証処理**: `static/js/auth-store.js`

```javascript
// 既存のsignUp関数に追加
async signUp(email, password, username) {
  // ... 既存の処理 ...
  
  if (data.user) {
    // Analytics イベント送信
    if (window.Analytics) {
      Analytics.trackSignup('email');
    }
    
    // ... 既存の処理 ...
  }
}

// 既存のsignIn関数に追加
async signIn(email, password) {
  // ... 既存の処理 ...
  
  if (data.user) {
    // Analytics イベント送信
    if (window.Analytics) {
      Analytics.trackLogin('email');
    }
    
    // ... 既存の処理 ...
  }
}
```

### 3. htmxとの統合

htmxのページ遷移時にページビューを送信。

**ファイル**: `static/js/htmx-analytics.js`（新規作成）

```javascript
// htmx ページビュー追跡
document.body.addEventListener('htmx:afterSettle', (event) => {
  // URLが変更された場合のみページビューを送信
  const newUrl = window.location.pathname + window.location.search;
  
  if (typeof gtag !== 'undefined') {
    gtag('config', 'G-XXXXXXXXXX', {
      'page_path': newUrl
    });
  }
});

// htmx リクエストエラー追跡
document.body.addEventListener('htmx:responseError', (event) => {
  if (typeof gtag !== 'undefined') {
    gtag('event', 'exception', {
      'description': `htmx error: ${event.detail.xhr.status}`,
      'fatal': false
    });
  }
});
```

**base.tmpl に追加**:

```html
<!-- htmx Analytics統合 -->
{{ if .IsProduction }}
<script src="/static/js/htmx-analytics.js"></script>
{{ end }}
```

### 4. プライバシー対応

#### 4.1 Cookie同意バナー（オプション）

GDPR/CCPA対応が必要な場合。

**ファイル**: `templates/components/cookie-consent.tmpl`（新規作成）

```html
{{ define "cookie-consent" }}
<div id="cookie-consent" 
     class="fixed bottom-0 left-0 right-0 bg-gray-900 text-white p-4 shadow-lg z-50"
     style="display: none;">
  <div class="container mx-auto flex flex-col md:flex-row items-center justify-between gap-4">
    <div class="flex-1">
      <p class="text-sm">
        当サイトでは、サービスの改善とパーソナライズされた広告配信のためにCookieを使用しています。
        <a href="/privacy" class="underline hover:text-gray-300">プライバシーポリシー</a>
      </p>
    </div>
    <div class="flex gap-2">
      <button onclick="CookieConsent.accept()" 
              class="bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded text-sm font-medium">
        同意する
      </button>
      <button onclick="CookieConsent.reject()" 
              class="bg-gray-700 hover:bg-gray-600 px-6 py-2 rounded text-sm font-medium">
        拒否する
      </button>
    </div>
  </div>
</div>

<script>
const CookieConsent = {
  accept: () => {
    localStorage.setItem('cookie-consent', 'accepted');
    document.getElementById('cookie-consent').style.display = 'none';
    // Analytics有効化
    window['ga-disable-G-XXXXXXXXXX'] = false;
  },
  
  reject: () => {
    localStorage.setItem('cookie-consent', 'rejected');
    document.getElementById('cookie-consent').style.display = 'none';
    // Analytics無効化
    window['ga-disable-G-XXXXXXXXXX'] = true;
  },
  
  check: () => {
    const consent = localStorage.getItem('cookie-consent');
    if (!consent) {
      document.getElementById('cookie-consent').style.display = 'block';
    } else if (consent === 'rejected') {
      window['ga-disable-G-XXXXXXXXXX'] = true;
    }
  }
};

// ページ読み込み時にチェック
document.addEventListener('DOMContentLoaded', CookieConsent.check);
</script>
{{ end }}
```

**base.tmpl の `<body>` 内に追加**:

```html
<!-- Cookie同意バナー -->
{{ if .IsProduction }}
  {{ template "cookie-consent" . }}
{{ end }}
```

#### 4.2 プライバシーポリシー更新

`templates/pages/privacy.tmpl` に以下のセクションを追加：

```markdown
## Cookie とトラッキング技術

当サービスでは、以下の目的でCookieおよびトラッキング技術を使用しています：

### Google Analytics
- サイトの利用状況の分析
- ユーザー行動の理解とサービス改善
- IPアドレスは匿名化されます

### Google AdSense
- パーソナライズされた広告の配信
- 広告効果の測定

詳細は[Googleのプライバシーポリシー](https://policies.google.com/privacy)をご確認ください。

Cookieの使用を拒否する場合は、ブラウザの設定から無効化できます。
```

---

## 最適化とベストプラクティス

### AdSense最適化

#### 1. 広告配置のベストプラクティス

```
推奨配置:
✅ ファーストビュー内に1つ
✅ コンテンツ途中（自然な流れ）
✅ サイドバー（デスクトップ）
✅ コンテンツ下部

避けるべき配置:
❌ ボタンやリンクの直近
❌ ページ上部に複数配置
❌ コンテンツを隠すような配置
❌ 誤クリックを誘発する配置
```

#### 2. レスポンシブ広告

```html
<!-- 全てのデバイスに最適化 -->
<ins class="adsbygoogle"
     style="display:block"
     data-ad-format="auto"
     data-full-width-responsive="true"></ins>
```

#### 3. 広告密度

```
理想的な広告密度:
- 1スクリーン: 最大1広告
- 1ページ: 最大3-5広告
- 広告/コンテンツ比: 1:3程度
```

#### 4. CSS最適化

**ファイル**: `static/css/style.css`

```css
/* 広告コンテナ */
.ad-container {
  margin: 2rem 0;
  text-align: center;
  min-height: 250px; /* レイアウトシフト防止 */
}

/* モバイル最適化 */
@media (max-width: 768px) {
  .ad-container {
    margin: 1.5rem 0;
    min-height: 200px;
  }
}

/* デスクトップサイドバー広告 */
.ad-sidebar {
  position: sticky;
  top: 100px; /* ヘッダー高さ + 余白 */
  max-width: 300px;
}

/* 広告ラベル（AdSenseポリシー遵守） */
.ad-label {
  font-size: 0.75rem;
  color: #9ca3af;
  text-align: center;
  margin-bottom: 0.25rem;
}
```

### Analytics最適化

#### 1. カスタムディメンション

GA4でカスタムディメンションを設定：

```javascript
// ユーザープロパティ
gtag('set', 'user_properties', {
  'user_level': 'free', // または 'premium'
  'favorite_game': 'MHP2G'
});

// カスタムディメンション
gtag('event', 'page_view', {
  'user_type': isAuthenticated ? 'member' : 'guest',
  'game_filter': currentGameFilter
});
```

#### 2. eコマース追跡（将来的）

将来的に課金機能を追加する場合：

```javascript
// 購入イベント
gtag('event', 'purchase', {
  'transaction_id': 'T12345',
  'value': 500,
  'currency': 'JPY',
  'items': [{
    'item_id': 'premium_plan',
    'item_name': 'プレミアムプラン',
    'price': 500
  }]
});
```

#### 3. スクロール深度追跡

```javascript
// スクロール深度を追跡（25%, 50%, 75%, 100%）
let scrollDepths = [25, 50, 75, 100];
let trackedDepths = [];

window.addEventListener('scroll', () => {
  const scrollPercent = (window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100;
  
  scrollDepths.forEach(depth => {
    if (scrollPercent >= depth && !trackedDepths.includes(depth)) {
      trackedDepths.push(depth);
      
      if (typeof gtag !== 'undefined') {
        gtag('event', 'scroll_depth', {
          'depth': depth
        });
      }
    }
  });
});
```

---

## モニタリングとレポート

### AdSense KPI

#### 日次モニタリング
- **ページRPM**: 1,000ページビューあたりの収益
- **クリック率（CTR）**: クリック数 / 表示回数
- **クリック単価（CPC）**: 収益 / クリック数
- **表示回数**: 広告が表示された回数

#### 目標値
```
初期（1-3ヶ月）:
- ページRPM: ¥200-500
- CTR: 0.5-1.5%
- CPC: ¥20-50

成熟期（6ヶ月以降）:
- ページRPM: ¥500-1,000
- CTR: 1-2%
- CPC: ¥30-70
```

### Analytics KPI

#### ユーザー行動
- **セッション数**: 訪問回数
- **ページビュー数**: 閲覧ページ数
- **直帰率**: 1ページのみ閲覧して離脱した割合
- **平均セッション時間**: 1訪問あたりの滞在時間

#### エンゲージメント
- **部屋作成率**: 登録ユーザーのうち部屋を作成した割合
- **部屋参加率**: 訪問ユーザーのうち部屋に参加した割合
- **リテンション率**: 再訪問率

#### コンバージョン
- **新規登録数**: ユーザー登録完了数
- **登録コンバージョン率**: 訪問者のうち登録した割合

### レポート設定

#### GA4 カスタムレポート

1. **ユーザー獲得レポート**
   - 流入元（検索、SNS、直接等）
   - デバイス別（モバイル、デスクトップ）
   - 地域別

2. **エンゲージメントレポート**
   - 部屋作成・参加イベント
   - ゲームバージョン別人気度
   - セッション時間分布

3. **コンバージョンレポート**
   - 新規登録数
   - 部屋作成数
   - アクティブユーザー数

#### 定期レポート設定

```bash
# 週次レポート:
- 訪問者数、PV数
- AdSense収益
- コンバージョン数

# 月次レポート:
- 成長率（MoM）
- 収益推移
- ユーザー行動分析
```

---

## トラブルシューティング

### AdSense

#### 問題: 広告が表示されない

```
チェックリスト:
1. ads.txt が正しく設置されているか確認
2. AdSenseアカウントが有効か確認
3. ブラウザのAdBlockが無効か確認
4. JavaScriptコンソールでエラーを確認
5. 広告スロットIDが正しいか確認
```

#### 問題: 収益が低い

```
改善策:
1. 広告配置を最適化（ヒートマップ分析）
2. コンテンツの質を向上
3. トラフィックを増やす（SEO、SNS）
4. 広告サイズを最適化
5. ページ読み込み速度を改善
```

### Analytics

#### 問題: トラッキングされない

```
チェックリスト:
1. 測定IDが正しいか確認
2. JavaScriptが正常に読み込まれているか確認
3. AdBlockが無効か確認
4. リアルタイムレポートで確認
```

#### 問題: イベントが記録されない

```
チェックリスト:
1. gtag関数が定義されているか確認
2. イベント名が正しいか確認（予約語に注意）
3. DebugViewで確認（GA4管理画面）
```

---

## セキュリティとプライバシー

### 法的要件

#### 必要なページ
- [ ] プライバシーポリシー
- [ ] 利用規約
- [ ] Cookie使用に関する通知

#### GDPR対応（EU向け）
- [ ] Cookie同意バナー
- [ ] データ削除リクエスト対応
- [ ] データポータビリティ

#### CCPA対応（カリフォルニア州向け）
- [ ] 個人情報の販売オプトアウト
- [ ] プライバシーポリシーでの開示

### データ保護

```javascript
// Analytics IP匿名化（既に設定済み）
gtag('config', 'G-XXXXXXXXXX', {
  'anonymize_ip': true
});

// ユーザーIDの安全な設定
gtag('set', {
  'user_id': hashUserId(userId) // ハッシュ化されたID
});
```

---

## チェックリスト

### AdSense導入チェックリスト

- [ ] AdSenseアカウント作成
- [ ] サイト審査申請
- [ ] ads.txt 設置
- [ ] 基本スクリプト追加（base.tmpl）
- [ ] 広告ユニット作成（AdSenseダッシュボード）
- [ ] 広告コンポーネント実装（ad-unit.tmpl）
- [ ] 各ページに広告配置
- [ ] htmx統合（htmx-ads.js）
- [ ] 開発/本番環境分岐実装
- [ ] モバイル表示確認
- [ ] AdSenseポリシー確認
- [ ] プライバシーポリシー更新

### Analytics導入チェックリスト

- [ ] GA4プロパティ作成
- [ ] 測定ID取得
- [ ] 基本トラッキングコード追加
- [ ] カスタムイベント実装（analytics.js）
- [ ] 各機能にイベント追跡追加
- [ ] htmx統合（htmx-analytics.js）
- [ ] カスタムディメンション設定
- [ ] コンバージョン目標設定
- [ ] レポート設定
- [ ] デバッグビューで動作確認

---

## 参考リソース

### AdSense
- [AdSense ヘルプセンター](https://support.google.com/adsense)
- [AdSense プログラムポリシー](https://support.google.com/adsense/answer/48182)
- [ads.txt ガイド](https://support.google.com/adsense/answer/7532444)

### Analytics
- [GA4 ヘルプ](https://support.google.com/analytics)
- [GA4 イベントリファレンス](https://developers.google.com/analytics/devguides/collection/ga4/reference/events)
- [GA4 DebugView](https://support.google.com/analytics/answer/7201382)

### プライバシー
- [GDPR 公式サイト](https://gdpr.eu/)
- [Googleプライバシーポリシー](https://policies.google.com/privacy)

---

## 更新履歴

| 日付 | バージョン | 変更内容 | 担当者 |
|------|----------|---------|--------|
| 2025-10-06 | 1.0 | 初版作成 | - |
