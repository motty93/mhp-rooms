# DDoS対策メモ

本プロジェクトで想定するDDoS/大量アクセス対策の考え方をまとめる。実装変更が発生した場合は随時更新すること。

## 1. 基本方針

- **静的レスポンス化**: トップページなどは Supabase 認証JSを読み込まず、テンプレートもDBアクセス不要にする。これによりアプリ側の負荷を最小化。
- **最前面での防御**: 可能な限り CDN/WAF (Cloudflare, CloudFront + AWS WAF など) を挟み、L3/L4/L7 での攻撃をインフラレイヤーで止める。
- **段階的レートリミット**: 入口（CDN/WAF）→LB→アプリ (chi ミドルウェア) の順に閾値を設定し、手前でブロックできなかった分に対する最終防衛をアプリで担当する。

## 2. インフラ側で行うこと

1. **CDN/WAF**
   - Bot Fight、User-Agent ブロック、Geo制限など提供機能を活用。
   - `/` はキャッシュヒットさせ、Cache-Control: `public, max-age=60` などのヘッダーをアプリ側で返すと効果的。
2. **ロードバランサ設定**
   - Keep-Alive / IdleTimeout を適切に設定し、ウンロード接続を早めに切断。
   - SYN Flood 対応のConnection Limiterを有効にする。
3. **監視**
   - HTTP 429/5xxの急増をアラート化、CDN側の攻撃ログを定期レビュー。

## 3. アプリ側で行うこと

1. **chiレートリミット**
   - 既存の `middleware.RateLimitMiddleware` を `/` にも適用済。閾値 (`config/limits.yml`) を環境に応じて調整。
2. **軽量レスポンス**
   - StaticPageフラグでSupabase/Alpineストアを読み込まないページは極力それを使う。
   - DB呼び出しが必要なページでもクエリ数を把握し、N+1などを排除。
3. **非同期処理/タイムアウト**
   - Goサーバの `http.Server` 設定 (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`) を30s/30s/60s程度にしておく。
   - 外部API呼び出しが増えたら context timeout を必ず設定。

## 4. 追加施策候補

- **IP Reputationリスト連携**: Spamhaus / Project Honey Pot などをWAFと連携。
- **CAPTCHA導入**: `/auth/register` や投稿フォームが攻撃される場合に reCAPTCHA / hCaptcha を追加。
- **バックプレッシャー**: アプリが高負荷になったら 503 + Retry-After を返して上流に制御を委ねる仕組みを検討。
- **自動ブラックリスト**: Rate Limit違反IPを短時間（5〜10分）だけブロックするメモリキャッシュを設ける。

## 5. 運用

- デプロイ前後で CDN キャッシュ/設定をメモしておく。
- 週次でアクセスログをサンプリングし、異常アクセスを可視化。
- AWS であれば Shield Advanced / GuardDuty のメトリクスを監視リストに入れる。

> メモ: 本ドキュメントは戦術レベルのガイド。実際の設定値やツール選定はインフラ構成に合わせて別途Runbookに記載する。
