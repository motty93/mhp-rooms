document.addEventListener('alpine:init', () => {
  Alpine.store('roomCreate', {
    // モーダル表示状態
    showModal: false,
    showHostWarningModal: false,
    showGuestConfirmModal: false,

    // 現在参加中の部屋情報
    currentRoomInfo: null,

    // フォームデータ
    formData: {
      name: '',
      gameVersionId: '',
      maxPlayers: '',
      password: '',
      targetMonster: '',
      rankRequirement: '',
      description: '',
    },

    // フォームエラー
    formErrors: {},

    // ゲームバージョンリスト
    gameVersions: [],

    // 送信中フラグ
    isSubmitting: false,

    // 初期化
    async init() {
      // ページ読み込み時には何もしない（モーダル開く時にゲームバージョンを取得）
    },

    // ゲームバージョンを取得
    async loadGameVersions() {
      console.log('loadGameVersions called, current length:', this.gameVersions.length)
      console.trace('Stack trace for loadGameVersions call:')

      // 既に取得済みの場合はスキップ
      if (this.gameVersions.length > 0) {
        console.log('Skipping loadGameVersions - already loaded')
        return
      }

      console.log('Fetching game versions from API...')
      try {
        const response = await fetch('/api/game-versions/active', {
          credentials: 'same-origin',
        })

        if (response.ok) {
          const data = await response.json()
          this.gameVersions = data.game_versions || []
          console.log('Loaded game versions:', this.gameVersions.length)
        } else {
          console.error('ゲーム情報の取得に失敗しました')
        }
      } catch (error) {
        console.error('ゲーム情報取得エラー:', error)
      }
    },

    // モーダルを開く
    async open() {
      console.log('roomCreate.open() called - SHOULD ONLY BE CALLED WHEN BUTTON CLICKED')
      console.trace('Stack trace for open() call:')
      console.log('roomCreate.open() called')

      // 認証チェック
      const authStore = Alpine.store('auth')
      if (!authStore.initialized || !authStore.isAuthenticated) {
        // 未認証の場合はログインページへ
        window.location.href = '/auth/login'
        return
      }

      // ゲームバージョンを取得（まだ取得していない場合のみ）
      await this.loadGameVersions()

      try {
        // 部屋状態をチェック
        const response = await fetch('/api/user/current/room-status', {
          headers: {
            Authorization: `Bearer ${authStore.session.access_token}`,
          },
        })

        if (!response.ok) {
          throw new Error('部屋状態の取得に失敗しました')
        }

        const data = await response.json()

        if (data.status === 'HOST') {
          // ホスト中の場合は警告モーダルを表示
          this.currentRoomInfo = data.room
          this.showHostWarningModal = true
        } else if (data.status === 'GUEST') {
          // 参加中の場合は確認モーダルを表示
          this.currentRoomInfo = data.room
          this.showGuestConfirmModal = true
        } else {
          // 未参加の場合は通常どおり作成モーダルを開く
          this.resetForm()
          this.showModal = true
        }
      } catch (error) {
        console.error('部屋状態チェックエラー:', error)
        // エラーの場合も作成モーダルを開く（サーバー側でチェックされる）
        this.resetForm()
        this.showModal = true
      }
    },

    // モーダルを閉じる
    close() {
      this.showModal = false
      this.resetForm()
    },

    // フォームリセット
    resetForm() {
      this.formData = {
        name: '',
        gameVersionId: '',
        maxPlayers: '',
        password: '',
        targetMonster: '',
        rankRequirement: '',
        description: '',
      }
      this.formErrors = {}
      this.isSubmitting = false
      this.currentRoomInfo = null
    },

    // ホスト警告モーダルを閉じる
    closeHostWarningModal() {
      this.showHostWarningModal = false
      this.currentRoomInfo = null
    },

    // ゲスト確認モーダルを閉じる
    closeGuestConfirmModal() {
      this.showGuestConfirmModal = false
      this.currentRoomInfo = null
    },

    // 確認後に部屋作成モーダルを開く
    confirmAndOpenModal() {
      this.showGuestConfirmModal = false
      this.resetForm()
      this.showModal = true
    },

    // フォームバリデーション
    get isValidForm() {
      return this.formData.name.trim() && this.formData.gameVersionId && this.formData.maxPlayers
    },

    // 部屋作成処理
    async createRoom() {
      if (!this.isValidForm || this.isSubmitting) {
        return
      }

      this.isSubmitting = true
      this.formErrors = {}

      try {
        const authStore = Alpine.store('auth')
        if (!authStore.isAuthenticated || !authStore.session?.access_token) {
          throw new Error('認証が必要です')
        }

        const requestData = {
          name: this.formData.name.trim(),
          game_version_id: this.formData.gameVersionId,
          max_players: Number.parseInt(this.formData.maxPlayers),
          password: this.formData.password.trim() || null,
          target_monster: this.formData.targetMonster.trim() || null,
          rank_requirement: this.formData.rankRequirement.trim() || null,
          description: this.formData.description.trim() || null,
        }

        const response = await fetch('/rooms', {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${authStore.session.access_token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestData),
        })

        if (!response.ok) {
          if (response.status === 409) {
            const errorData = await response.json()
            if (errorData.error === 'HOST_ROOM_ACTIVE') {
              // ホスト制限エラー
              this.currentRoomInfo = errorData.room
              this.showHostWarningModal = true
              this.close()
              return
            }
          } else if (response.status === 400) {
            const errorData = await response.json()
            this.formErrors = errorData.errors || {}
            throw new Error(errorData.message || 'バリデーションエラー')
          }

          const errorText = await response.text()
          throw new Error(errorText || '部屋の作成に失敗しました')
        }

        const result = await response.json()

        // 成功時は部屋詳細画面に遷移
        this.close()

        if (result.redirect) {
          window.location.href = result.redirect
        } else {
          window.location.reload()
        }
      } catch (error) {
        console.error('部屋作成エラー:', error)
        // エラー表示
        alert(error.message || '部屋の作成に失敗しました')
      } finally {
        this.isSubmitting = false
      }
    },

    // Escキー処理
    handleEscape() {
      if (this.showModal) {
        this.close()
      } else if (this.showHostWarningModal) {
        this.closeHostWarningModal()
      } else if (this.showGuestConfirmModal) {
        this.closeGuestConfirmModal()
      }
    },
  })
})
