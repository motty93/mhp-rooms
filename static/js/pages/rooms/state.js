import * as api from './api.js'
import * as ui from './ui.js'
import { validateCreateForm } from './validator.js'

export function rooms() {
  return {
    // ---- state ----
    showModal: false,
    showLoginModal: false,
    currentRoom: {},
    password: '',
    isJoining: false,
    showError: false,
    errorMessage: '',
    allRooms: [],
    filteredRooms: [],
    activeFilter: '',
    lastFocusedElement: null,
    showConfirmDialog: false,
    confirmRoomId: null,
    showBlockWarningDialog: false,
    blockWarningMessage: '',
    showHostRestrictionModal: false,
    hostRoomInfo: null,

    // create modal
    showCreateModal: false,
    showHostWarningModal: false,
    showGuestConfirmModal: false,
    currentRoomInfo: null,
    createFormData: {
      name: '',
      gameVersionId: '',
      maxPlayers: '',
      password: '',
      targetMonster: '',
      rankRequirement: '',
      description: '',
    },
    createFormErrors: {},
    gameVersions: [],
    isSubmitting: false,

    // ---- lifecycle ----
    async checkPSNIdRequired() {
      try {
        const authStore = ui.getAuthStore()
        if (!authStore?.initialized) return
        if (authStore.needsPSNId) {
          console.log('PSN IDが未設定です。プロフィール設定画面へリダイレクトします。')
          window.location.href = '/auth/complete-profile'
        }
      } catch (e) {
        console.error('PSN IDチェックエラー:', e)
      }
    },

    init() {
      // グローバルESCの登録（解除用に参照保持）
      this._boundEsc = this.handleEsc.bind(this)
      document.addEventListener('keydown', this._boundEsc)

      // ゲームバージョン取得
      this.loadGameVersions()

      // SSR埋め込みデータから初期化
      const roomsSSR = ui.parseJsonScript('rooms-data') || []
      this.allRooms = ui.normalizeRooms(roomsSSR)

      // デバッグ: 初心者歓迎部屋の参加状態
      const shoshinsha = this.allRooms.find((r) => r.name === '初心者歓迎部屋')
      if (shoshinsha) {
        console.log('初心者歓迎部屋:', shoshinsha.name, 'isJoined:', shoshinsha.isJoined)
      }

      // URLパラメータの初期フィルタ
      const urlParams = new URLSearchParams(window.location.search)
      const gameVersionFilter = urlParams.get('game_version') || ''
      this.activeFilter = gameVersionFilter
      this.filterRooms(gameVersionFilter)

      // 認証状態監視
      this.watchAuthState()
    },

    destroy() {
      // Alpineは自動で呼ばないが、DOMから外す場合のため
      document.removeEventListener('keydown', this._boundEsc)
    },

    // ---- auth watch ----
    watchAuthState() {
      let lastAuthState = null
      const checkAuthChange = () => {
        const authStore = ui.getAuthStore()
        if (!authStore) return

        const now = {
          initialized: authStore.initialized,
          isAuthenticated: authStore.isAuthenticated,
          userId: authStore.user?.id,
        }

        if (lastAuthState === null) {
          lastAuthState = now
          return
        }
        if (
          now.initialized &&
          now.isAuthenticated &&
          (!lastAuthState.isAuthenticated || lastAuthState.userId !== now.userId)
        ) {
          console.log('認証状態変化を検出、部屋データを再取得します')
          this.refreshRoomData()
        }
        lastAuthState = now
      }

      const authWatcher = setInterval(checkAuthChange, 1000)
      setTimeout(() => clearInterval(authWatcher), 300000) // 5分で停止
    },

    // ---- data ops ----
    async loadGameVersions() {
      try {
        const data = await api.fetchActiveGameVersions()
        this.gameVersions = data?.game_versions || []
      } catch (e) {
        console.error('ゲーム情報取得エラー:', e)
      }
    },

    async refreshRoomData() {
      try {
        const token = ui.getAuthToken()
        if (!token) {
          console.log('未認証のため再取得スキップ')
          return
        }
        const data = await api.fetchRooms(token)
        this.allRooms = ui.normalizeRooms(data?.rooms || [])
        this.filterRooms(this.activeFilter)

        const shoshinsha = this.allRooms.find((r) => r.name === '初心者歓迎部屋')
        if (shoshinsha)
          console.log('再取得後 - 初心者歓迎部屋:', shoshinsha.name, 'isJoined:', shoshinsha.isJoined)
      } catch (e) {
        console.error('部屋データ再取得エラー:', e)
      }
    },

    // ---- filter ----
    filterRooms(gameVersion) {
      this.activeFilter = gameVersion
      this.filteredRooms = gameVersion
        ? this.allRooms.filter((r) => r.gameVersion.code === gameVersion)
        : this.allRooms

      const url = new URL(window.location)
      if (gameVersion) url.searchParams.set('game_version', gameVersion)
      else url.searchParams.delete('game_version')
      window.history.replaceState({}, '', url)
    },

    // ---- join modal ----
    get isJoinDisabled() {
      return this.isJoining || (this.currentRoom.hasPassword && !this.password.trim())
    },

    async openModal(room) {
      const auth = ui.getAuthStore()
      if (auth?.isAuthenticated && auth?.session?.access_token) {
        try {
          const data = await api.getUserRoomStatus(auth.session.access_token)
          if (data.status === 'HOST') {
            this.hostRoomInfo = data.room
            this.showHostRestrictionModal = true
            return
          }
        } catch (e) {
          console.error('部屋状態チェックエラー:', e)
        }
      }

      this.lastFocusedElement = document.activeElement
      this.currentRoom = {
        id: room.id,
        name: room.name,
        description: room.description,
        gameVersionCode: room.gameVersion.code,
        gameVersionName: room.gameVersion.name,
        hostName: room.host.username || room.host.displayName,
        currentPlayers: room.currentPlayers,
        maxPlayers: room.maxPlayers,
        hasPassword: room.hasPassword,
      }
      this.password = ''
      this.isJoining = false
      this.showModal = true

      this.$nextTick(() => {
        const modal = this.$el.querySelector('[role="dialog"]')
        if (!modal) return
        const focusables = modal.querySelectorAll('button, input, a')
        if (room.hasPassword) {
          const i = modal.querySelector("input[type='password']")
          if (i) i.focus()
        } else if (focusables.length) {
          focusables[focusables.length - 1].focus()
        }
      })
    },

    closeModal() {
      this.showModal = false
      setTimeout(() => {
        this.currentRoom = {}
        this.password = ''
        this.isJoining = false
        this.showError = false
        this.errorMessage = ''
      }, 300)
      if (this.lastFocusedElement) {
        this.lastFocusedElement.focus()
        this.lastFocusedElement = null
      }
    },

    openLoginModal() {
      this.showLoginModal = true
    },
    closeLoginModal() {
      this.showLoginModal = false
    },

    async joinRoom() {
      if (this.isJoinDisabled) return

      const auth = ui.getAuthStore()
      if (!auth?.isAuthenticated) {
        this.closeModal()
        this.openLoginModal()
        return
      }

      this.isJoining = true
      const req = {}
      if (this.currentRoom.hasPassword) req.password = this.password

      try {
        const r = await api.joinRoom(this.currentRoom.id, req, auth.session?.access_token)

        if (!r.ok) {
          if (r.status === 401) throw new Error('認証が必要です。ログインしてください。')

          if (r.status === 409) {
            // JSONの可能性とテキストの可能性があるため両対応
            let errorData = null
            let errorText = ''
            try {
              errorData = await r.json()
            } catch {
              errorText = await r.text()
            }

            if (errorData?.error === 'HOST_CANNOT_JOIN') {
              this.closeModal()
              this.hostRoomInfo = errorData.room
              this.showHostRestrictionModal = true
              this.isJoining = false
              return
            }
            const text = errorText || errorData?.message || ''
            if (text.includes('既に別の部屋に参加しています')) {
              this.showConfirmDialog = true
              this.confirmRoomId = this.currentRoom.id
              this.isJoining = false
              return
            }
            throw new Error(text || '参加に失敗しました')
          }

          if (r.status === 403) {
            const errorData = await r.json()
            if (errorData.error === 'BLOCKED_BY_HOST') this.showErrorMessage('このルームには参加できません')
            else if (errorData.error === 'BLOCKED_BY_MEMBER')
              this.showErrorMessage('ブロック関係により参加できません')
            else this.showErrorMessage(errorData.message || '参加に失敗しました')
            this.isJoining = false
            return
          }

          if (r.status === 400) {
            const t = await r.text()
            throw new Error(t || '参加に失敗しました')
          }

          throw new Error('参加に失敗しました')
        }

        const result = await r.json()

        if (result.warning === 'USER_BLOCKING_HOST' && result.requiresConfirmation) {
          this.showBlockWarningDialog = true
          this.blockWarningMessage = result.message
          this.isJoining = false
          return
        }

        ui.redirectOrReload(result)
      } catch (e) {
        this.showErrorMessage(e.message || '参加に失敗しました。もう一度お試しください。')
        this.isJoining = false
        if (this.currentRoom.hasPassword) {
          this.password = ''
          this.$nextTick(() => {
            const input = this.$el.querySelector("input[type='password']")
            if (input) input.focus()
          })
        }
      }
    },

    // Esc（共通）
    handleEsc(ev) {
      if (ev.key !== 'Escape') return
      if (this.showModal) this.closeModal()
      else if (this.showCreateModal) this.closeCreateModal()
      else if (this.showHostWarningModal) this.closeHostWarningModal()
      else if (this.showGuestConfirmModal) this.closeGuestConfirmModal()
    },

    // アイコン
    getGameIconClass(code) {
      const valid = ['mhp', 'mhp2', 'mhp2g', 'mhp3', 'mhxx']
      const lower = (code || '').toLowerCase()
      return valid.includes(lower) ? `${lower}-icon` : 'default-icon'
    },

    // エラー表示
    showErrorMessage(message) {
      this.errorMessage = message
      this.showError = true
      setTimeout(() => {
        this.showError = false
        this.errorMessage = ''
      }, 3000)
    },

    // 退出して参加
    async confirmJoinRoom() {
      if (!this.confirmRoomId) return
      this.isJoining = true

      try {
        const token = ui.getAuthToken(true)
        const req = { forceJoin: true }
        if (this.currentRoom.hasPassword) req.password = this.password

        const r = await api.joinRoom(this.confirmRoomId, req, token)
        if (!r.ok) {
          const text = await r.text()
          if (r.status === 400 && text.includes('パスワードが間違っています')) {
            throw new Error('パスワードが間違っています')
          }
          throw new Error(text || '参加に失敗しました')
        }

        const result = await r.json()
        this.closeConfirmDialog()
        this.closeModal()
        ui.redirectOrReload(result)
      } catch (e) {
        console.error('処理エラー:', e)
        const msg = e?.message || '処理に失敗しました。もう一度お試しください。'
        if (msg.includes('パスワードが間違っています')) {
          this.password = ''
          this.closeConfirmDialog()
          this.$nextTick(() => {
            const input = this.$el.querySelector("input[type='password']")
            if (input) input.focus()
          })
        }
        this.showErrorMessage(msg)
        this.isJoining = false
      }
    },

    closeConfirmDialog() {
      this.showConfirmDialog = false
      this.confirmRoomId = null
      this.isJoining = false
      this.closeModal()
    },

    // ブロック警告
    closeBlockWarningDialog() {
      this.showBlockWarningDialog = false
      this.blockWarningMessage = ''
      this.isJoining = false
    },

    async confirmJoinWithBlock() {
      this.isJoining = true
      const req = { confirmJoin: true }
      if (this.currentRoom.hasPassword) req.password = this.password

      try {
        const headersToken = ui.getAuthToken()
        const r = await api.joinRoom(this.currentRoom.id, req, headersToken)

        if (!r.ok) {
          if (r.status === 403) {
            const data = await r.json()
            this.showErrorMessage(data.message || '参加に失敗しました')
            this.closeBlockWarningDialog()
            return
          }
          throw new Error('参加に失敗しました')
        }

        const result = await r.json()
        this.closeBlockWarningDialog()
        this.closeModal()
        ui.redirectOrReload(result)
      } catch (e) {
        this.showErrorMessage(e.message || '参加に失敗しました。もう一度お試しください。')
        this.isJoining = false
      }
    },

    // 作成モーダル
    resetCreateForm() {
      this.createFormData = {
        name: '',
        gameVersionId: '',
        maxPlayers: '',
        password: '',
        targetMonster: '',
        rankRequirement: '',
        description: '',
      }
      this.createFormErrors = {}
      this.isSubmitting = false
      this.currentRoomInfo = null
    },

    async openCreateModal() {
      const auth = ui.getAuthStore()
      if (!auth?.initialized || !auth?.isAuthenticated) {
        window.location.href = '/auth/login'
        return
      }

      try {
        const data = await api.getUserRoomStatus(auth.session.access_token)
        if (data.status === 'HOST') {
          this.currentRoomInfo = data.room
          this.showHostWarningModal = true
        } else if (data.status === 'GUEST') {
          this.currentRoomInfo = data.room
          this.showGuestConfirmModal = true
        } else {
          this.resetCreateForm()
          this.showCreateModal = true
        }
      } catch (e) {
        console.error('部屋状態チェックエラー:', e)
        this.resetCreateForm()
        this.showCreateModal = true
      }
    },

    closeCreateModal() {
      this.showCreateModal = false
    },
    closeHostWarningModal() {
      this.showHostWarningModal = false
      this.currentRoomInfo = null
    },
    closeGuestConfirmModal() {
      this.showGuestConfirmModal = false
      this.currentRoomInfo = null
    },
    confirmAndOpenCreateModal() {
      this.showGuestConfirmModal = false
      this.resetCreateForm()
      this.showCreateModal = true
    },

    get isValidCreateForm() {
      return validateCreateForm(this.createFormData)
    },

    async createRoom() {
      if (!this.isValidCreateForm || this.isSubmitting) return

      this.isSubmitting = true
      this.createFormErrors = {}

      try {
        const token = ui.getAuthToken(true)
        const f = this.createFormData

        const payload = {
          name: f.name.trim(),
          game_version_id: ui.coerceId(f.gameVersionId), // ← UUID/数値どちらでもOK
          max_players: Number.parseInt(f.maxPlayers, 10),
          password: f.password.trim() || null,
          target_monster: f.targetMonster.trim() || null,
          rank_requirement: f.rankRequirement.trim() || null,
          description: f.description.trim() || null,
        }

        const r = await api.createRoom(payload, token)

        if (!r.ok) {
          if (r.status === 409) {
            const data = await r.json()
            if (data.error === 'HOST_ROOM_ACTIVE') {
              this.currentRoomInfo = data.room
              this.showHostWarningModal = true
              this.closeCreateModal()
              return
            }
          } else if (r.status === 400) {
            const data = await r.json()
            this.createFormErrors = data.errors || {}
            throw new Error(data.message || 'バリデーションエラー')
          }
          const t = await r.text()
          throw new Error(t || '部屋の作成に失敗しました')
        }

        const result = await r.json()
        this.closeCreateModal()
        ui.redirectOrReload(result)
      } catch (e) {
        console.error('部屋作成エラー:', e)
        alert(e.message || '部屋の作成に失敗しました')
      } finally {
        this.isSubmitting = false
      }
    },
  }
}
