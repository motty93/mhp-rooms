// お知らせ（ヘッダーのベル / モバイルメニュー）の状態管理
// 認証が確定したら未読数を取得し、パネルを開いたときに一覧を再取得して既読にする
document.addEventListener('alpine:init', () => {
  Alpine.store('notifications', {
    open: false,
    loading: false,
    loaded: false,
    unreadCount: 0,
    items: [],
    error: null,

    init() {
      // 認証状態の変化を監視: ログインで初回取得、ログアウトでリセット
      Alpine.effect(() => {
        const auth = Alpine.store('auth')
        if (!auth || !auth.initialized) return

        if (auth.isAuthenticated && auth.session?.access_token) {
          if (!this.loaded && !this.loading) {
            this.fetch()
          }
        } else if (!auth.isAuthenticated) {
          this.reset()
        }
      })
    },

    reset() {
      this.open = false
      this.loading = false
      this.loaded = false
      this.unreadCount = 0
      this.items = []
      this.error = null
    },

    authHeaders() {
      const token = Alpine.store('auth')?.session?.access_token
      return token ? { Authorization: `Bearer ${token}` } : {}
    },

    async fetch() {
      if (!Alpine.store('auth')?.session?.access_token) return

      this.loading = true
      try {
        const response = await fetch('/api/notifications', {
          headers: { ...this.authHeaders(), Accept: 'application/json' },
        })
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`)
        }
        const data = await response.json()
        this.items = Array.isArray(data.items) ? data.items : []
        this.unreadCount = data.unread_count || 0
        this.loaded = true
        this.error = null
      } catch (error) {
        console.warn('お知らせの取得に失敗:', error)
        this.error = 'お知らせを読み込めませんでした'
      } finally {
        this.loading = false
      }
    },

    async toggle() {
      if (this.open) {
        this.close()
        return
      }
      await this.openPanel()
    },

    async openPanel() {
      this.open = true
      await this.fetch()
      if (this.unreadCount > 0) {
        await this.markAllRead()
      }
    },

    close() {
      this.open = false
    },

    // すべて既読にする（一覧上の未読表示は次回取得まで残し、バッジだけ消す）
    async markAllRead() {
      try {
        const response = await fetch('/api/notifications/read', {
          method: 'POST',
          headers: this.authHeaders(),
        })
        if (response.ok) {
          this.unreadCount = 0
        }
      } catch (error) {
        console.warn('お知らせの既読更新に失敗:', error)
      }
    },

    get badgeText() {
      return this.unreadCount > 99 ? '99+' : String(this.unreadCount)
    },
  })
})
