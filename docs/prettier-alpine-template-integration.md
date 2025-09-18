# Prettier と Alpine.js のテンプレート統合問題の解決

## 問題の概要

GoテンプレートでAlpine.jsの`x-init`属性に複雑なテンプレート構文を含める際、Prettierによる自動整形で以下の問題が発生：

1. **構文エラー**: Prettierが`x-init`内のGoテンプレート構文を改行し、JavaScriptとして無効な構文になる
2. **空白問題**: `{{ if }}` ブロックが改行されることで、属性値に余分な空白が含まれる
3. **nil値の表示**: nil値が`<nil>`として表示される

### 発生した具体的なエラー例
```javascript
// Alpine.jsのエラー
Uncaught SyntaxError: Invalid or unexpected token
    at new AsyncFunction (<anonymous>)

// Prettierで整形後のコード（問題あり）
x-init="playTimes = { weekday: '{{ if .PlayTimes }}
      {{ jsEscape .PlayTimes.Weekday }}
    {{ end }}' }"
```

## 試した解決策

### 1. ❌ .prettierignoreでテンプレートファイルを除外
```
*.tmpl
```
**問題**: テンプレート全体が整形されなくなり、根本解決ではない

### 2. ❌ prettier-ignoreコメント
```html
<!-- prettier-ignore -->
<div x-init="...">
```
**問題**: `prettier-plugin-go-template`では効果がない

### 3. ❌ x-init内で条件分岐
```html
data-init-bio="{{ if .User.Bio }}{{ .User.Bio }}{{ end }}"
```
**問題**: Prettierで改行され、フォームに空白が入る

## 最終的な解決策 ✅

### サーバー側でnil値を空文字列に変換

#### 1. ハンドラー側の実装（`internal/handlers/profile.go`）

```go
func (ph *ProfileHandler) EditForm(w http.ResponseWriter, r *http.Request) {
    // ... ユーザー情報取得 ...

    // nil値を空文字列に変換する関数
    safeString := func(s *string) string {
        if s != nil {
            return *s
        }
        return ""
    }

    // テンプレート用データ（nil安全な値を提供）
    data := struct {
        User              *models.User
        FavoriteGames     []string
        PlayTimes         *models.PlayTimes
        // テンプレート用にnil安全な値を提供
        Bio               string
        PSNOnlineID       string
        NintendoNetworkID string
        NintendoSwitchID  string
        TwitterID         string
    }{
        User:              user,
        FavoriteGames:     favoriteGames,
        PlayTimes:         playTimes,
        Bio:               safeString(user.Bio),
        PSNOnlineID:       safeString(user.PSNOnlineID),
        NintendoNetworkID: safeString(user.NintendoNetworkID),
        NintendoSwitchID:  safeString(user.NintendoSwitchID),
        TwitterID:         safeString(user.TwitterID),
    }

    // ... テンプレートレンダリング ...
}
```

#### 2. テンプレート側の実装（`templates/components/profile_edit_form.tmpl`）

```html
<div
  x-data="profileEditForm()"
  data-init-display-name="{{ .User.DisplayName }}"
  data-init-bio="{{ .Bio }}"
  data-init-psn-online-id="{{ .PSNOnlineID }}"
  data-init-nintendo-network-id="{{ .NintendoNetworkID }}"
  data-init-nintendo-switch-id="{{ .NintendoSwitchID }}"
  data-init-twitter-id="{{ .TwitterID }}"
  data-init-favorite-games="{{ json .FavoriteGames }}"
  data-init-play-times="{{ json .PlayTimes }}"
  x-init="$data.initFromDataAttributes($el)"
  class="flex flex-col items-center"
>
```

#### 3. JavaScript側の実装（`static/js/profile.js`）

```javascript
window.profileEditForm = function() {
    return {
        displayName: '',
        bio: '',
        psnOnlineId: '',
        // ... その他のプロパティ ...

        // data属性から初期値を読み込む
        initFromDataAttributes(el) {
            // data属性から値を取得
            this.displayName = el.dataset.initDisplayName || '';
            this.bio = el.dataset.initBio || '';
            this.psnOnlineId = el.dataset.initPsnOnlineId || '';
            this.nintendoNetworkId = el.dataset.initNintendoNetworkId || '';
            this.nintendoSwitchId = el.dataset.initNintendoSwitchId || '';
            this.twitterId = el.dataset.initTwitterId || '';
            
            // JSONデータのパース
            try {
                this.favoriteGames = el.dataset.initFavoriteGames 
                    ? JSON.parse(el.dataset.initFavoriteGames) 
                    : [];
            } catch (e) {
                this.favoriteGames = [];
            }
            
            try {
                this.playTimes = el.dataset.initPlayTimes 
                    ? JSON.parse(el.dataset.initPlayTimes) 
                    : { weekday: '', weekend: '' };
            } catch (e) {
                this.playTimes = { weekday: '', weekend: '' };
            }
            
            this.init();
        },

        init() {
            // 初期化処理
        }
    };
};
```

## メリット

1. ✅ **Prettier互換性**: 整形しても問題なし
2. ✅ **空白問題解決**: サーバー側で空文字列を保証
3. ✅ **nil値処理**: `<nil>`が表示されない
4. ✅ **コード可読性**: テンプレートがシンプル
5. ✅ **デバッグ容易**: data属性として値が可視化
6. ✅ **Alpine.jsエラーなし**: 構文エラーが発生しない

## 重要なポイント

### 1. data属性のケバブケース変換
HTML属性名はケバブケース（`data-init-psn-online-id`）だが、JavaScriptではキャメルケース（`dataset.initPsnOnlineId`）でアクセス

### 2. x-initでのスコープ指定
`x-init="$data.initFromDataAttributes($el)"` のように `$data.` を付けてAlpine.jsのスコープ内で関数を参照

### 3. JSONデータの適切なパース
エラーハンドリングを含めてJSONデータをパース

## 適用可能な他のケース

この解決策は以下のような場合にも適用可能：

- 複雑な初期値を持つフォーム
- 動的にロードされるコンポーネント
- htmxで部分的に更新されるUI要素
- Prettierとテンプレートエンジンの競合が発生する場合

## まとめ

テンプレートエンジンとPrettier、Alpine.jsの3つを併用する際は、サーバー側でデータを適切に前処理することが重要。nil値の処理や複雑なテンプレート構文は、可能な限りサーバー側で解決し、クライアント側にはシンプルな形で渡すことで、開発体験とコードの保守性が大幅に向上する。