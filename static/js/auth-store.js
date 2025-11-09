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
    _lastSyncTime: 0,
    currentRoom: null,
    _currentRoomLoading: false,
    _currentRoomFetched: false,
    dbUser: null,

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
      // 前のユーザーIDを保存
      const previousUserId = this.user?.id

      this.session = session
      this.user = session?.user || null
      this.error = null

      if (this.user && session?.access_token) {
        // ローカルストレージからDBユーザー情報を読み込み
        this.loadDbUserFromStorage()

        if (!this._syncInProgress) {
          this.syncUser(session.access_token)
        }

        // 認証成功時にcurrentRoomを取得（初回のみ）
        if (!this._currentRoomFetched && !this._currentRoomLoading) {
          this.fetchCurrentRoom()
        }
      } else {
        // 認証がない場合はcurrentRoomとdbUserをクリア
        this.currentRoom = null
        this.dbUser = null
        this._currentRoomFetched = false
        // 前のユーザーIDを使ってローカルストレージをクリア
        if (previousUserId) {
          this.clearDbUserFromStorage(previousUserId)
        } else {
          this.clearDbUserFromStorage()
        }
      }
    },

    clearSession() {
      // セッション情報をクリア
      this.session = null
      this.user = null
      this.dbUser = null
      this.error = null
      this.currentRoom = null
      this._currentRoomFetched = false
      this._currentRoomLoading = false

      // ローカルストレージからDBユーザー情報を削除
      this.clearDbUserFromStorage()

      // クッキーを削除
      document.cookie = 'sb-access-token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT'
    },

    async syncUser(accessToken) {
      // 同期処理の重複実行を防ぐ（5秒以内の重複実行を防ぐ）
      const now = Date.now()
      if (this._syncInProgress || now - this._lastSyncTime < 5000) {
        return
      }
      this._syncInProgress = true
      this._lastSyncTime = now

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
        } else {
          // 同期成功時にDBユーザー情報を取得（キャッシュされていない場合のみ）
          if (!this.dbUser) {
            await this.fetchDbUser(accessToken)
          }
        }
      } catch (error) {
        console.error('ユーザー同期エラー:', error)
      } finally {
        this._syncInProgress = false
      }
    },

    async fetchDbUser(accessToken) {
      try {
        const response = await fetch('/api/user/me', {
          headers: {
            Authorization: `Bearer ${accessToken}`,
          },
        })

        if (response.ok) {
          const userData = await response.json()
          this.dbUser = userData.user
          // ローカルストレージに保存
          this.saveDbUserToStorage(userData.user)
        }
      } catch (error) {
        console.error('DBユーザー情報取得エラー:', error)
      }
    },

    saveDbUserToStorage(dbUser) {
      try {
        const storageKey = `mhp-rooms-dbuser-${this.user?.id}`
        const storageData = {
          user: dbUser,
          timestamp: Date.now(),
          expires: Date.now() + 24 * 60 * 60 * 1000, // 24時間後に期限切れ
        }
        localStorage.setItem(storageKey, JSON.stringify(storageData))
      } catch (error) {
        console.error('ローカルストレージへの保存エラー:', error)
      }
    },

    loadDbUserFromStorage() {
      try {
        if (!this.user?.id) return null

        const storageKey = `mhp-rooms-dbuser-${this.user.id}`
        const storedData = localStorage.getItem(storageKey)

        if (!storedData) return null

        const parsedData = JSON.parse(storedData)

        // 期限切れチェック
        if (Date.now() > parsedData.expires) {
          localStorage.removeItem(storageKey)
          return null
        }

        this.dbUser = parsedData.user
        return parsedData.user
      } catch (error) {
        console.error('ローカルストレージからの読み込みエラー:', error)
        return null
      }
    },

    clearDbUserFromStorage(userId = null) {
      try {
        // userIdが渡されない場合は、現在のユーザーIDを使用
        const targetUserId = userId || this.user?.id
        if (!targetUserId) {
          // すべてのdbユーザーキャッシュをクリア
          const keys = Object.keys(localStorage)
          keys.forEach((key) => {
            if (key.startsWith('mhp-rooms-dbuser-')) {
              localStorage.removeItem(key)
            }
          })
          return
        }

        const storageKey = `mhp-rooms-dbuser-${targetUserId}`
        localStorage.removeItem(storageKey)
      } catch (error) {
        console.error('ローカルストレージのクリアエラー:', error)
      }
    },

    // DBユーザー情報を強制的に更新する
    async refreshDbUser() {
      if (!this.session?.access_token) return

      this.clearDbUserFromStorage()
      await this.fetchDbUser(this.session.access_token)
    },

    get isAuthenticated() {
      return !!this.user
    },

    get displayName() {
      return this.dbUser?.display_name || ''
    },

    get username() {
      return (
        this.dbUser?.username || this.user?.email?.split('@')[0] || this.user?.user_metadata?.name || 'ゲスト'
      )
    },

    get avatarUrl() {
      return this.dbUser?.avatar_url || '/static/images/default-avatar.webp'
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
          credentials: 'same-origin',
        })

        if (response.ok) {
          const data = await response.json()
          this.currentRoom = data.current_room
          this._currentRoomFetched = true
        } else {
          this.currentRoom = null
        }
      } catch (error) {
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
        const previousRoomId = this.currentRoom?.id
        const response = await fetch('/api/leave-current-room', {
          method: 'POST',
          credentials: 'same-origin',
        })

        if (response.ok) {
          if (previousRoomId && window.Analytics && window.Analytics.isEnabled()) {
            window.Analytics.trackRoomLeave(previousRoomId)
          }
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
