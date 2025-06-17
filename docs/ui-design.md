# UI設計仕様書

## 概要

モンスターハンターポータブルシリーズのアドホックパーティルーム管理システムのUI設計仕様です。htmx + Alpine.js + Tailwind CSSを使用したサーバーサイドレンダリングベースのWebアプリケーションです。

## 基本レイアウト

### ページ構成
```
┌─────────────────────────────────────┐
│              Header                 │
├─────────────────────────────────────┤
│                                     │
│            Main Content             │
│                                     │
├─────────────────────────────────────┤
│              Footer                 │
└─────────────────────────────────────┘
```

## Header

### 基本構成
```
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│ MHP Rooms    ナビゲーションメニュー        ログイン 新規登録 │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

#### 未認証時
- 左側: アプリケーションロゴ/タイトル「MHP Rooms」
- 中央: 空白（ナビゲーションは認証後のみ）
- 右側: **ログインボタン** と **新規登録ボタン**

#### 認証済み時
- 左側: アプリケーションロゴ/タイトル「MHP Rooms」
- 中央: ナビゲーションメニュー（ルーム一覧、ルーム作成等）
- 右側: **ユーザーメニュー**（アバター、ユーザー名、ドロップダウン）

### Headerボタン仕様

#### 未認証時の右上ボタン
```html
<div class="flex items-center gap-3">
  <!-- ログインボタン -->
  <button 
    hx-get="/auth/login" 
    hx-target="#main-content"
    class="text-blue-600 hover:text-blue-800 font-medium transition-colors">
    ログイン
  </button>
  
  <!-- 新規登録ボタン -->
  <button 
    hx-get="/auth/register" 
    hx-target="#main-content"
    class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors">
    新規登録
  </button>
</div>
```

#### 認証済み時の右上メニュー
```html
<div class="relative" x-data="{ open: false }">
  <button 
    @click="open = !open" 
    class="flex items-center gap-2 text-gray-700 hover:text-gray-900 transition-colors">
    <img 
      src="${user.avatar_url || '/default-avatar.png'}" 
      class="w-8 h-8 rounded-full object-cover" 
      alt="${user.display_name}のアバター">
    <span class="font-medium">${user.display_name}</span>
    <svg class="w-4 h-4 transition-transform" :class="open ? 'rotate-180' : ''" fill="currentColor" viewBox="0 0 20 20">
      <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd"/>
    </svg>
  </button>
  
  <div 
    x-show="open" 
    @click.away="open = false"
    x-transition:enter="transition ease-out duration-100"
    x-transition:enter-start="transform opacity-0 scale-95"
    x-transition:enter-end="transform opacity-100 scale-100"
    x-transition:leave="transition ease-in duration-75"
    x-transition:leave-start="transform opacity-100 scale-100"
    x-transition:leave-end="transform opacity-0 scale-95"
    class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-50 border border-gray-200">
    
    <a href="/profile" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors">
      プロフィール
    </a>
    <a href="/rooms/my" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors">
      マイルーム
    </a>
    <hr class="my-1 border-gray-200">
    <button 
      hx-post="/auth/logout"
      hx-confirm="ログアウトしますか？"
      class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100 transition-colors">
      ログアウト
    </button>
  </div>
</div>
```

### レスポンシブ対応
- **モバイル**: ハンバーガーメニューによる折りたたみ
  - ログイン/新規登録ボタンもメニュー内に配置
- **タブレット以上**: 水平展開

## Main Content

### トップページ（ルーム一覧）
- ゲームバージョン別タブ（MHP、MHP2、MHP2G、MHP3）
- ルームカード形式での一覧表示
- 各ルームカードには以下の情報を表示：
  - ルーム名
  - 現在の参加人数/最大人数（4人）
  - ホスト名
  - ゲーム内容（クエストタイプ、ターゲットモンスター等）
  - パスワード有無のアイコン
  - 参加ボタン（認証済みユーザーのみ）

### ルーム詳細ページ
- ルーム情報の詳細表示
- メンバー一覧
- チャット機能（認証済みユーザーのみ）
- ルーム設定変更（ホストのみ）

## Footer

### 基本構成
```
┌─────────────────────────────────────┐
│ ┌─────────────┐         ┌─────────┐ │
│ │  サイト情報  │         │ ログイン │ │
│ │             │         │ ボタン   │ │
│ └─────────────┘         └─────────┘ │
└─────────────────────────────────────┘
```

### 左側：サイト情報
- コピーライト表示
- プライバシーポリシーリンク
- 利用規約リンク
- お問い合わせリンク

### 右上：ログインボタン
#### 未認証時
- **ログインボタン**: メールアドレス認証用
- **Googleでログイン**: Google OAuth認証用
- ボタンスタイル：
  ```html
  <!-- メールログイン -->
  <button class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors">
    ログイン
  </button>
  
  <!-- Googleログイン -->
  <button class="bg-white hover:bg-gray-50 text-gray-900 font-medium py-2 px-4 rounded-md border border-gray-300 transition-colors flex items-center gap-2">
    <svg class="w-5 h-5" viewBox="0 0 24 24"><!-- Google icon --></svg>
    Googleでログイン
  </button>
  ```

#### 認証済み時
- **ユーザー名表示**
- **ログアウトボタン**
- ドロップダウンメニュー形式：
  ```html
  <div class="relative" x-data="{ open: false }">
    <button @click="open = !open" class="flex items-center gap-2 text-gray-700 hover:text-gray-900">
      <img src="avatar_url" class="w-8 h-8 rounded-full" alt="アバター">
      <span>ユーザー名</span>
      <svg class="w-4 h-4" fill="currentColor"><!-- dropdown arrow --></svg>
    </button>
    
    <div x-show="open" @click.away="open = false" class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg">
      <a href="/profile" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">プロフィール</a>
      <a href="/rooms/my" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100">マイルーム</a>
      <hr class="my-1">
      <button hx-post="/auth/logout" class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100">
        ログアウト
      </button>
    </div>
  </div>
  ```

### レスポンシブ対応
- **デスクトップ**: 左右に配置
- **タブレット**: 左右に配置（ボタンサイズ調整）
- **モバイル**: 縦積み（左側情報を上、ログイン関連を下）

## カラーパレット

### プライマリカラー
- **メイン**: `#2563eb` (blue-600) - ログインボタン、アクション要素
- **ホバー**: `#1d4ed8` (blue-700)

### セカンダリカラー
- **Google**: `#ffffff` (white) - Googleログインボタン背景
- **ボーダー**: `#d1d5db` (gray-300)

### テキストカラー
- **プライマリ**: `#111827` (gray-900)
- **セカンダリ**: `#374151` (gray-700)
- **ミュート**: `#6b7280` (gray-500)

### ステータスカラー
- **成功**: `#10b981` (emerald-500)
- **警告**: `#f59e0b` (amber-500)
- **エラー**: `#ef4444` (red-500)

## コンポーネント仕様

### ログイン関連コンポーネント

#### HeaderAuthButtons（Headerの未認証時）
```html
<div class="flex items-center gap-3">
  <!-- ログインボタン -->
  <button 
    hx-get="/auth/login" 
    hx-target="#main-content"
    hx-push-url="true"
    class="text-blue-600 hover:text-blue-800 font-medium transition-colors px-3 py-2 rounded-md hover:bg-blue-50">
    ログイン
  </button>
  
  <!-- 新規登録ボタン -->
  <button 
    hx-get="/auth/register" 
    hx-target="#main-content"
    hx-push-url="true"
    class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors shadow-sm">
    新規登録
  </button>
</div>
```

#### FooterLoginButtons（Footerの未認証時）
```html
<div class="flex flex-col sm:flex-row gap-2">
  <!-- メールログイン -->
  <button 
    hx-get="/auth/login" 
    hx-target="#main-content"
    class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md transition-colors">
    ログイン
  </button>
  
  <!-- Googleログイン -->
  <a 
    href="/auth/google"
    class="bg-white hover:bg-gray-50 text-gray-900 font-medium py-2 px-4 rounded-md border border-gray-300 transition-colors flex items-center justify-center gap-2">
    <svg class="w-5 h-5" viewBox="0 0 24 24">
      <!-- Google Gアイコン -->
      <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
      <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
      <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
      <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
    </svg>
    Googleでログイン
  </a>
</div>
```

#### HeaderUserMenu（Headerの認証済み時）
```html
<div class="relative" x-data="{ open: false }">
  <button 
    @click="open = !open" 
    class="flex items-center gap-2 text-gray-700 hover:text-gray-900 transition-colors p-2 rounded-md hover:bg-gray-100">
    <img 
      src="${user.avatar_url || '/default-avatar.png'}" 
      class="w-8 h-8 rounded-full object-cover" 
      alt="${user.display_name}のアバター">
    <span class="font-medium hidden sm:block">${user.display_name}</span>
    <svg class="w-4 h-4 transition-transform" :class="open ? 'rotate-180' : ''" fill="currentColor" viewBox="0 0 20 20">
      <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd"/>
    </svg>
  </button>
  
  <div 
    x-show="open" 
    @click.away="open = false"
    x-transition:enter="transition ease-out duration-100"
    x-transition:enter-start="transform opacity-0 scale-95"
    x-transition:enter-end="transform opacity-100 scale-100"
    x-transition:leave="transition ease-in duration-75"
    x-transition:leave-start="transform opacity-100 scale-100"
    x-transition:leave-end="transform opacity-0 scale-95"
    class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-50 border border-gray-200">
    
    <a href="/profile" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors">
      プロフィール
    </a>
    <a href="/rooms/my" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors">
      マイルーム
    </a>
    <a href="/settings" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors">
      設定
    </a>
    <hr class="my-1 border-gray-200">
    <button 
      hx-post="/auth/logout"
      hx-confirm="ログアウトしますか？"
      class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100 transition-colors">
      ログアウト
    </button>
  </div>
</div>
```

#### FooterUserMenu（Footerの認証済み時）
```html
<div class="relative" x-data="{ open: false }">
  <button 
    @click="open = !open" 
    class="flex items-center gap-2 text-gray-700 hover:text-gray-900 transition-colors">
    <img 
      src="${user.avatar_url || '/default-avatar.png'}" 
      class="w-8 h-8 rounded-full object-cover" 
      alt="${user.display_name}のアバター">
    <span class="font-medium">${user.display_name}</span>
    <svg class="w-4 h-4 transition-transform" :class="open ? 'rotate-180' : ''" fill="currentColor" viewBox="0 0 20 20">
      <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd"/>
    </svg>
  </button>
  
  <div 
    x-show="open" 
    @click.away="open = false"
    x-transition:enter="transition ease-out duration-100"
    x-transition:enter-start="transform opacity-0 scale-95"
    x-transition:enter-end="transform opacity-100 scale-100"
    x-transition:leave="transition ease-in duration-75"
    x-transition:leave-start="transform opacity-100 scale-100"
    x-transition:leave-end="transform opacity-0 scale-95"
    class="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-50 border border-gray-200">
    
    <a href="/profile" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors">
      プロフィール
    </a>
    <a href="/rooms/my" class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors">
      マイルーム
    </a>
    <hr class="my-1 border-gray-200">
    <button 
      hx-post="/auth/logout"
      hx-confirm="ログアウトしますか？"
      class="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100 transition-colors">
      ログアウト
    </button>
  </div>
</div>
```

## アクセシビリティ

### キーボードナビゲーション
- Tabキーによる順次移動
- Enterキー/Spaceキーでのボタン操作
- Escapeキーでのメニュー閉じる

### スクリーンリーダー対応
- 適切なaria-label属性
- role属性の設定
- フォーカス管理

### カラーコントラスト
- WCAG 2.1 AA基準準拠
- テキストと背景のコントラスト比 4.5:1以上

## インタラクション仕様

### ログインフロー
1. **未認証時**: Headerの右上にログイン・新規登録ボタンを表示
2. **ログインボタンクリック**: ログインページへ遷移（htmx部分更新）
3. **新規登録ボタンクリック**: 新規登録ページへ遷移（htmx部分更新）
4. **認証成功**: ページリロードまたはhtmxによる部分更新
5. **認証後**: ログイン/新規登録ボタンがユーザーメニューに変更
6. **ナビゲーション表示**: 認証後にHeader中央にナビゲーションメニューが表示される

### htmx統合
- ログイン状態の動的更新
- 認証が必要な操作での適切なリダイレクト
- エラーメッセージの表示

## 今後の拡張

### Phase 2機能
- ダークモード対応
- 通知システム
- PWA対応

### Phase 3機能
- モバイルアプリライクなUI
- プッシュ通知
- オフライン対応