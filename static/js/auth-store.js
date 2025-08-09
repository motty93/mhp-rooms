document.addEventListener('alpine:init', () => {
  Alpine.store('auth', {
    user: null,
    session: null,
    loading: true,
    error: null,
    configError: null,
    initialized: false,
    _initStarted: false,
    _syncInProgress: false,
    currentRoom: null,
    _currentRoomLoading: false,
    _currentRoomFetched: false,

    init() {
      // 初期化の重複実行を防ぐ
      if (this._initStarted) {
        return
      }
      this._initStarted = true

      if (window.supabaseClient) {
        this.checkAuth()
      } else {
        document.addEventListener('supabase-initialized', () => {
          this.checkAuth()
        })
      }
    },

    async checkAuth() {
      this.loading = true
      this.error = null

      try {
        if (window.supabaseAuth) {
          const session = await window.supabaseAuth.getSession()
          this.updateSession(session)
        } else {
          this.updateSession(null)
        }
      } catch (error) {
        console.error('認証状態チェックエラー:', error)
        this.error = error.message
        this.updateSession(null)
      } finally {
        this.loading = false
        this.initialized = true
      }
    },

    updateSession(session) {
      this.session = session
      this.user = session?.user || null
      this.error = null

      if (this.user && session?.access_token && !this._syncInProgress) {
        this.syncUser(session.access_token)
        // 認証成功時にcurrentRoomを取得
        this.fetchCurrentRoom()
      } else {
        // 認証がない場合はcurrentRoomをクリア
        this.currentRoom = null
        this._currentRoomFetched = false
      }
    },

    async syncUser(accessToken) {
      // 同期処理の重複実行を防ぐ
      if (this._syncInProgress) {
        return
      }
      this._syncInProgress = true

      try {
        const response = await fetch('/api/auth/sync', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${accessToken}`,
          },
          body: JSON.stringify({
            psn_id: this.user?.user_metadata?.psn_id || '',
          }),
        })

        if (!response.ok) {
          console.error('ユーザー同期に失敗しました:', response.status)
        }
      } catch (error) {
        console.error('ユーザー同期エラー:', error)
      } finally {
        this._syncInProgress = false
      }
    },

    get isAuthenticated() {
      return !!this.user
    },

    get username() {
      return this.user?.email?.split('@')[0] || this.user?.user_metadata?.name || 'ゲスト'
    },

    get needsPSNId() {
      return this.isAuthenticated && !this.user?.user_metadata?.psn_id
    },

    async signIn(email, password) {
      if (!window.supabaseAuth) {
        throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。')
      }

      this.loading = true
      this.error = null

      try {
        const data = await window.supabaseAuth.signIn(email, password)
        return data
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    async signUp(email, password, metadata = {}) {
      if (!window.supabaseAuth) {
        throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。')
      }

      this.loading = true
      this.error = null

      try {
        const data = await window.supabaseAuth.signUp(email, password, metadata)
        return data
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    async signOut() {
      if (!window.supabaseAuth) {
        window.location.href = '/'
        return
      }

      this.loading = true
      this.error = null

      try {
        await window.supabaseAuth.signOut()
        window.location.href = '/'
      } catch (error) {
        this.error = error.message
        window.location.href = '/'
      } finally {
        this.loading = false
      }
    },

    async resetPassword(email) {
      if (!window.supabaseAuth) {
        throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。')
      }

      this.loading = true
      this.error = null

      try {
        const data = await window.supabaseAuth.resetPassword(email)
        return data
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    async signInWithGoogle() {
      if (!window.supabaseAuth) {
        throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。')
      }

      this.loading = true
      this.error = null

      try {
        const data = await window.supabaseAuth.signInWithGoogle()
        return data
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    async updatePSNId(psnId) {
      if (!this.session?.access_token) {
        throw new Error('認証が必要です')
      }

      this.loading = true
      this.error = null

      try {
        const response = await fetch('/api/auth/psn-id', {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${this.session.access_token}`,
          },
          body: JSON.stringify({
            psn_id: psnId,
          }),
        })

        if (!response.ok) {
          throw new Error('PSN IDの更新に失敗しました')
        }

        if (window.supabaseAuth && typeof window.supabaseAuth.updateUserMetadata === 'function') {
          await window.supabaseAuth.updateUserMetadata({ psn_id: psnId })
        }

        // PSN ID更新後はcheckAuth()を呼ばずに、ユーザーメタデータのみ更新
        // 重複SQLを避けるため、フル認証チェックは行わない
        if (this.user && this.user.user_metadata) {
          this.user.user_metadata.psn_id = psnId
        }

        return response.json()
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    async fetchCurrentRoom() {
      // 重複実行を防ぐ
      if (this._currentRoomLoading || !this.isAuthenticated) {
        return
      }

      this._currentRoomLoading = true

      try {
        const response = await fetch('/api/user/current-room', {
          credentials: 'same-origin'
        })

        if (response.ok) {
          const data = await response.json()
          this.currentRoom = data.current_room
          this._currentRoomFetched = true
        } else {
          console.log('参加中の部屋の取得に失敗しました')
          this.currentRoom = null
        }
      } catch (error) {
        console.log('参加中の部屋の取得に失敗:', error)
        this.currentRoom = null
      } finally {
        this._currentRoomLoading = false
      }
    },

    async leaveCurrentRoom() {
      if (!this.currentRoom || !this.isAuthenticated) {
        return
      }

      try {
        const response = await fetch('/api/leave-current-room', {
          method: 'POST',
          credentials: 'same-origin'
        })

        if (response.ok) {
          this.currentRoom = null
          return true
        }
      } catch (error) {
        console.error('部屋からの退出に失敗しました:', error)
      }
      return false
    },
  })
})
