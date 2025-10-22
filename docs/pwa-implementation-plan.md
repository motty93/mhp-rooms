# PWA実装計画書

## 目次
1. [概要](#概要)
2. [実装フェーズ](#実装フェーズ)
3. [技術仕様](#技術仕様)
4. [実装詳細](#実装詳細)
5. [テスト計画](#テスト計画)
6. [デプロイ計画](#デプロイ計画)
7. [注意事項とリスク](#注意事項とリスク)

---

## 概要

### 目的
MonHubをProgressive Web App (PWA)化し、以下の機能を提供する：
- ホーム画面へのインストール可能化
- オフライン時の基本的なフォールバック表示
- アプリライクなユーザー体験の提供
- より高速なページ読み込み

### スコープ
- **対象**: 既存のWebアプリケーション全体
- **優先度**: リリース後の段階的実装
- **対象ブラウザ**: Chrome, Safari, Firefox, Edge（最新2バージョン）

### 除外スコープ
- 完全なオフライン機能（認証やリアルタイム機能はオンライン必須）
- プッシュ通知（将来の拡張として検討）
- バックグラウンド同期

---

## 実装フェーズ

### Phase 1: 基盤準備
- [ ] Webアプリマニフェストファイル作成
- [ ] アイコンセット準備（192x192, 512x512, maskable）
- [ ] スプラッシュスクリーン設定
- [ ] マニフェスト配信設定（Go側）

### Phase 2: Service Worker実装
- [ ] 基本的なService Worker作成
- [ ] キャッシュ戦略の実装
- [ ] オフラインフォールバックページ作成
- [ ] Service Worker登録処理

### Phase 3: 最適化と統合
- [ ] htmxとの統合テスト
- [ ] Alpine.js動作確認
- [ ] Supabase認証フロー確認
- [ ] パフォーマンス最適化

### Phase 4: テストとデプロイ
- [ ] 各種ブラウザでの動作確認
- [ ] インストール体験のテスト
- [ ] Lighthouseスコア確認
- [ ] 本番環境デプロイ

---

## 技術仕様

### 必要なファイル構成

```
/
├── static/
│   ├── manifest.json          # Webアプリマニフェスト
│   ├── sw.js                   # Service Worker
│   ├── images/
│   │   ├── icons/
│   │   │   ├── icon-192.png
│   │   │   ├── icon-512.png
│   │   │   ├── icon-maskable-192.png
│   │   │   └── icon-maskable-512.png
│   │   └── offline.svg         # オフライン画面用画像
│   └── offline.html            # オフラインフォールバックページ
└── templates/layouts/
    └── base.tmpl               # マニフェストリンク追加
```

---

## 実装詳細

### 1. Webアプリマニフェスト

**ファイル**: `static/manifest.json`

```json
{
  "name": "MonHub - モンハンポータブルオンラインルーム",
  "short_name": "MonHub",
  "description": "PSPゲームのアドホックパーティを簡単に作成・参加できるサービス",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#1f2937",
  "orientation": "portrait-primary",
  "scope": "/",
  "icons": [
    {
      "src": "/static/images/icons/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/images/icons/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/images/icons/icon-maskable-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "maskable"
    },
    {
      "src": "/static/images/icons/icon-maskable-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ],
  "categories": ["games", "social"],
  "screenshots": [
    {
      "src": "/static/images/screenshots/home.png",
      "sizes": "540x720",
      "type": "image/png",
      "form_factor": "narrow"
    },
    {
      "src": "/static/images/screenshots/rooms.png",
      "sizes": "1280x720",
      "type": "image/png",
      "form_factor": "wide"
    }
  ]
}
```

### 2. Service Worker

**ファイル**: `static/sw.js`

```javascript
const CACHE_NAME = 'monhub-v1';
const OFFLINE_URL = '/static/offline.html';

// キャッシュする静的リソース
const STATIC_CACHE_URLS = [
  '/',
  '/static/offline.html',
  '/static/css/style.css',
  '/static/js/vendor/htmx.min.js',
  '/static/js/vendor/alpine.min.js',
  '/static/js/supabase.js',
  '/static/js/auth-store.js',
  '/static/js/room-create-store.js',
  '/static/js/htmx-auth.js',
  '/static/images/icon.webp',
  '/static/images/hero.webp'
];

// キャッシュから除外するURL（認証、API、外部リソース）
const EXCLUDED_URLS = [
  'googlesyndication.com',
  'googleadservices.com',
  'doubleclick.net',
  'google-analytics.com',
  'googletagmanager.com',
  'supabase.co',
  'googleapis.com',
  '/api/',
  '/auth/'
];

// インストール時: 静的リソースをキャッシュ
self.addEventListener('install', (event) => {
  console.log('[SW] Install event');
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('[SW] Caching static resources');
        return cache.addAll(STATIC_CACHE_URLS);
      })
      .then(() => self.skipWaiting()) // 即座にアクティブ化
  );
});

// アクティベーション時: 古いキャッシュを削除
self.addEventListener('activate', (event) => {
  console.log('[SW] Activate event');
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            console.log('[SW] Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => self.clients.claim()) // 即座に制御開始
  );
});

// フェッチ時: キャッシュ戦略を適用
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // 除外URLはキャッシュしない（ネットワークのみ）
  if (EXCLUDED_URLS.some(excluded => url.href.includes(excluded))) {
    return;
  }

  // GETリクエスト以外はキャッシュしない
  if (request.method !== 'GET') {
    return;
  }

  event.respondWith(
    caches.match(request)
      .then((cachedResponse) => {
        // キャッシュ戦略: Stale-While-Revalidate
        // キャッシュがあればそれを返し、バックグラウンドで更新
        const fetchPromise = fetch(request)
          .then((networkResponse) => {
            // 成功したレスポンスをキャッシュに保存
            if (networkResponse && networkResponse.status === 200) {
              const responseToCache = networkResponse.clone();
              caches.open(CACHE_NAME).then((cache) => {
                cache.put(request, responseToCache);
              });
            }
            return networkResponse;
          })
          .catch(() => {
            // ネットワークエラー時: オフラインページを返す
            if (request.destination === 'document') {
              return caches.match(OFFLINE_URL);
            }
          });

        // キャッシュがあればすぐに返す、なければネットワークを待つ
        return cachedResponse || fetchPromise;
      })
  );
});

// メッセージハンドラー（将来の拡張用）
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});
```

### 3. Service Worker登録

**ファイル**: `static/js/sw-register.js`（新規作成）

```javascript
// Service Worker登録処理
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/static/sw.js')
      .then((registration) => {
        console.log('Service Worker registered:', registration.scope);

        // 更新があるか定期的にチェック
        registration.addEventListener('updatefound', () => {
          const newWorker = registration.installing;
          console.log('Service Worker update found');

          newWorker.addEventListener('statechange', () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
              // 新しいバージョンが利用可能
              console.log('New Service Worker available');
              
              // ユーザーに更新を通知（オプション）
              if (confirm('新しいバージョンが利用可能です。更新しますか？')) {
                newWorker.postMessage({ type: 'SKIP_WAITING' });
                window.location.reload();
              }
            }
          });
        });
      })
      .catch((error) => {
        console.error('Service Worker registration failed:', error);
      });

    // Service Workerが更新されたらページをリロード
    let refreshing = false;
    navigator.serviceWorker.addEventListener('controllerchange', () => {
      if (!refreshing) {
        refreshing = true;
        window.location.reload();
      }
    });
  });
}
```

### 4. オフラインページ

**ファイル**: `static/offline.html`

```html
<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>オフライン - MonHub</title>
  <style>
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    body {
      font-family: 'Noto Sans JP', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
      padding: 20px;
    }
    .container {
      text-align: center;
      max-width: 500px;
    }
    .icon {
      width: 120px;
      height: 120px;
      margin: 0 auto 30px;
      opacity: 0.9;
    }
    h1 {
      font-size: 28px;
      margin-bottom: 15px;
      font-weight: 700;
    }
    p {
      font-size: 16px;
      line-height: 1.6;
      margin-bottom: 30px;
      opacity: 0.9;
    }
    .retry-button {
      background: rgba(255, 255, 255, 0.2);
      border: 2px solid rgba(255, 255, 255, 0.8);
      color: #fff;
      padding: 12px 30px;
      border-radius: 8px;
      font-size: 16px;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.3s ease;
      text-decoration: none;
      display: inline-block;
    }
    .retry-button:hover {
      background: rgba(255, 255, 255, 0.3);
      transform: translateY(-2px);
    }
  </style>
</head>
<body>
  <div class="container">
    <svg class="icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
            d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414" />
    </svg>
    <h1>オフラインです</h1>
    <p>
      インターネット接続が利用できません。<br>
      接続を確認してから、もう一度お試しください。
    </p>
    <button class="retry-button" onclick="location.reload()">
      再試行
    </button>
  </div>
</body>
</html>
```

### 5. ベーステンプレート修正

**ファイル**: `templates/layouts/base.tmpl`

以下を `<head>` 内に追加：

```html
<!-- PWA Manifest -->
<link rel="manifest" href="/static/manifest.json">
<meta name="theme-color" content="#1f2937">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="MonHub">

<!-- iOS用アイコン（既存のapple-touch-iconを確認） -->
<link rel="apple-touch-icon" sizes="192x192" href="/static/images/icons/icon-192.png">
```

以下を `</body>` 直前に追加：

```html
<!-- Service Worker登録 -->
<script src="/static/js/sw-register.js"></script>
```

### 6. Go側のルーティング設定

**ファイル**: `cmd/server/routes.go`

`setupStaticRoutes` メソッドに以下を追加：

```go
func (app *Application) setupStaticRoutes(r chi.Router) {
	// 既存の静的ファイル配信
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// PWA用の追加設定
	// manifest.json
	r.Get("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		http.ServeFile(w, r, "./static/manifest.json")
	})

	// Service Worker
	r.Get("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Service-Worker-Allowed", "/")
		// キャッシュさせない
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		http.ServeFile(w, r, "./static/sw.js")
	})

	// オフラインページ
	r.Get("/offline", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/offline.html")
	})
}
```

### 7. アイコン生成

既存の `static/images/icon.webp` から以下のサイズを生成：

- `icon-192.png` (192x192)
- `icon-512.png` (512x512)
- `icon-maskable-192.png` (192x192, セーフゾーン考慮)
- `icon-maskable-512.png` (512x512, セーフゾーン考慮)

**ツール**: 
- ImageMagick
- オンラインツール: https://www.pwabuilder.com/imageGenerator

**コマンド例**:
```bash
# WebPからPNGに変換し、リサイズ
convert static/images/icon.webp -resize 192x192 static/images/icons/icon-192.png
convert static/images/icon.webp -resize 512x512 static/images/icons/icon-512.png

# Maskable版（20%のセーフゾーンを確保）
convert static/images/icon.webp -resize 154x154 -gravity center -extent 192x192 -background white static/images/icons/icon-maskable-192.png
convert static/images/icon.webp -resize 410x410 -gravity center -extent 512x512 -background white static/images/icons/icon-maskable-512.png
```

---

## テスト計画

### 1. ローカルテスト

#### 事前準備
```bash
# HTTPSが必要なため、ローカルでHTTPSサーバーを起動
# 方法1: mkcertを使用
mkcert -install
mkcert localhost

# 方法2: Caddyを使用
caddy reverse-proxy --from localhost:443 --to localhost:8080
```

#### テスト項目
- [ ] Service Workerが正常に登録される
- [ ] マニフェストが読み込まれる
- [ ] オフライン時にフォールバックページが表示される
- [ ] インストールプロンプトが表示される（Chrome, Edge）
- [ ] htmxの動的コンテンツ読み込みが正常動作
- [ ] Alpine.jsの状態管理が正常動作
- [ ] Supabase認証フローが正常動作

### 2. ブラウザ別テスト

| ブラウザ | インストール | オフライン | キャッシュ | 備考 |
|---------|------------|-----------|----------|------|
| Chrome (Desktop) | ✓ | ✓ | ✓ | 完全サポート |
| Chrome (Android) | ✓ | ✓ | ✓ | 完全サポート |
| Safari (iOS) | 部分的 | ✓ | ✓ | プロンプトなし |
| Safari (macOS) | 部分的 | ✓ | ✓ | プロンプトなし |
| Edge | ✓ | ✓ | ✓ | 完全サポート |
| Firefox | ✓ | ✓ | ✓ | 完全サポート |

### 3. Lighthouseスコア目標

| カテゴリ | 目標スコア | 現状予測 |
|---------|----------|---------|
| Performance | 90+ | 70-80 |
| Accessibility | 95+ | 85-90 |
| Best Practices | 95+ | 90+ |
| SEO | 100 | 95+ |
| PWA | 100 | 0 → 100 |

**重点項目**:
- Install prompt available
- Service worker registered
- Responds with 200 when offline
- Has a themed omnibar
- Content is sized correctly for the viewport

### 4. パフォーマンステスト

```bash
# Lighthouseでテスト
npx lighthouse https://localhost:8080 --view

# 特定のカテゴリのみテスト
npx lighthouse https://localhost:8080 --only-categories=pwa --view
```

---

## デプロイ計画

### 事前確認
```bash
# ビルド確認
make build

# 静的ファイルの存在確認
ls -la static/manifest.json
ls -la static/sw.js
ls -la static/images/icons/
```

### デプロイ前チェックリスト
- [ ] 全テストが合格
- [ ] Lighthouseスコアが目標を達成
- [ ] 全ブラウザで動作確認完了

### デプロイ後確認
- [ ] 本番URLでインストール確認
- [ ] Service Workerの登録確認（DevTools）
- [ ] オフライン動作確認
- [ ] 既存機能の動作確認（認証、部屋作成等）
- [ ] ログでエラー監視

---

## 注意事項とリスク

### 技術的リスク

#### 1. htmxとの互換性
**リスク**: htmxで動的に読み込むコンテンツがキャッシュされて古い情報が表示される

**対策**:
- htmxリクエストはStale-While-Revalidate戦略を使用
- 重要な動的コンテンツ（部屋一覧等）は短いキャッシュTTL
- 必要に応じてキャッシュバスティング（クエリパラメータ）

#### 2. Supabase認証
**リスク**: 認証トークンの有効期限とキャッシュの不整合

**対策**:
- 認証関連のリクエスト（`/auth/*`, `supabase.co`）は完全にキャッシュから除外
- Service Worker内で認証エラー時は強制リロード

#### 3. SSE (Server-Sent Events)
**リスク**: Service WorkerがSSE接続を妨げる可能性

**対策**:
- SSEエンドポイントをキャッシュ除外リストに追加
- イベントストリームは常にネットワーク経由

#### 4. ブラウザ互換性
**リスク**: Safariでの制限（インストールプロンプトなし、Service Worker制約）

**対策**:
- Safari用の代替インストール手順をUIで提示
- 「ホーム画面に追加」の手動手順を案内

### 運用上の注意点

#### 1. Service Workerの更新
- **問題**: ユーザーのブラウザに古いService Workerがキャッシュされる
- **対策**: `sw.js`のバージョン管理と強制更新メカニズム

#### 2. デバッグの難しさ
- **問題**: キャッシュが原因で変更が反映されない
- **対策**: DevToolsの「Bypass for network」を開発時に使用

#### 3. キャッシュサイズの管理
- **問題**: 過度なキャッシュでストレージを圧迫
- **対策**: キャッシュサイズ制限と定期的なクリーンアップ

### セキュリティ考慮事項

1. **HTTPS必須**: 本番環境では必ずHTTPSを使用（Cloud Runは自動対応）
2. **CSP設定**: Content Security Policyの適切な設定
3. **Service Workerのスコープ**: ルート (`/`) に限定

### パフォーマンス考慮事項

1. **初回読み込み**: Service Workerのインストールで若干遅くなる可能性
2. **キャッシュストレージ**: ブラウザのストレージ制限を考慮
3. **更新頻度**: Service Workerの更新頻度とユーザー体験のバランス

---

## 成功指標 (KPI)

### 技術指標
- [ ] Lighthouse PWAスコア: 100
- [ ] Service Worker登録率: 95%以上
- [ ] オフライン時のエラー率: 0%
- [ ] インストール率: 5%以上（3ヶ月以内）

### ユーザー体験指標
- [ ] ページ読み込み速度: 2秒以内（リピート訪問時）
- [ ] インストールプロンプト表示率: 50%以上
- [ ] インストール後のリテンション率: 通常の1.5倍

### ビジネス指標
- [ ] エンゲージメント率の向上: 20%以上
- [ ] セッション時間の増加: 10%以上
- [ ] 直帰率の減少: 15%以上

---

## 参考リソース

### 公式ドキュメント
- [MDN: Progressive Web Apps](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
- [Google: PWA Checklist](https://web.dev/pwa-checklist/)
- [Web.dev: Service Workers](https://web.dev/service-workers/)

### ツール
- [PWA Builder](https://www.pwabuilder.com/)
- [Lighthouse](https://developers.google.com/web/tools/lighthouse)
- [Workbox](https://developers.google.com/web/tools/workbox)（将来的な導入検討）

### テスト
- [PWA Testing](https://web.dev/pwa-testing/)
- [Can I Use: Service Workers](https://caniuse.com/serviceworkers)

---

## 更新履歴

| 日付 | バージョン | 変更内容 | 担当者 |
|------|----------|---------|--------|
| 2025-10-06 | 1.0 | 初版作成 | - |

---

## 添付資料

### A. チェックリスト

```markdown
## PWA実装チェックリスト

### Phase 1: 基盤準備
- [ ] manifest.json 作成
- [ ] アイコン生成（192x192, 512x512, maskable）
- [ ] base.tmpl にマニフェストリンク追加
- [ ] Go側のルーティング設定

### Phase 2: Service Worker
- [ ] sw.js 作成
- [ ] sw-register.js 作成
- [ ] offline.html 作成
- [ ] キャッシュ戦略実装

### Phase 3: テスト
- [ ] ローカルHTTPS環境構築
- [ ] Lighthouseテスト実行
- [ ] Chrome/Edge インストールテスト
- [ ] Safari 動作確認
- [ ] オフライン動作確認

### Phase 4: デプロイ
- [ ] ステージング環境デプロイ
- [ ] 本番環境デプロイ
- [ ] 本番動作確認
- [ ] モニタリング設定
```

### B. トラブルシューティング

```markdown
## よくある問題と解決方法

### 1. Service Workerが登録されない
- HTTPSで接続しているか確認
- ブラウザのDevToolsでエラーを確認
- sw.jsのパスが正しいか確認

### 2. インストールプロンプトが表示されない
- manifest.jsonが正しく読み込まれているか確認
- アイコンが存在するか確認
- Lighthouseで「installable」をチェック

### 3. オフライン時にページが表示されない
- offline.htmlがキャッシュされているか確認
- Service Workerのfetchイベントを確認

### 4. キャッシュが更新されない
- Cache-Control ヘッダーを確認
- Service Workerのバージョンを変更
- DevToolsで「Update on reload」を有効化

### 5. htmxの動的コンテンツが古い
- htmxリクエストのCache-Controlを確認
- Stale-While-Revalidate戦略を使用
```
