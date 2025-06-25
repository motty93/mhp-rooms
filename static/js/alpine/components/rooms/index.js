export function rooms() {
  return {
    showModal: false,
    showLoginModal: false,
    currentRoom: {},
    password: '',
    isJoining: false,

    // 参加ボタンの有効/無効判定
    get isJoinDisabled() {
      return this.isJoining || (this.currentRoom.hasPassword && !this.password.trim())
    },

    // 部屋参加モーダルを開く
    openModal(room) {
      this.currentRoom = room
      this.password = ''
      this.isJoining = false
      this.showModal = true

      if (room.hasPassword) {
        this.$nextTick(() => {
          const passwordInput = this.$el.querySelector("input[type='password']")
          if (passwordInput) passwordInput.focus()
        })
      }
    },

    // 部屋参加モーダルを閉じる
    closeModal() {
      this.showModal = false
      this.currentRoom = {}
      this.password = ''
      this.isJoining = false
    },

    // ログイン案内モーダルを開く
    openLoginModal() {
      this.showLoginModal = true
    },

    // ログイン案内モーダルを閉じる
    closeLoginModal() {
      this.showLoginModal = false
    },

    // 部屋に参加する
    async joinRoom() {
      if (this.isJoinDisabled) return

      this.isJoining = true
      const requestData = {}

      if (this.currentRoom.hasPassword) {
        requestData.password = this.password
      }

      try {
        const response = await fetch(`/rooms/${this.currentRoom.id}/join`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(requestData),
        })

        if (!response.ok) {
          throw new Error('参加に失敗しました')
        }

        alert('部屋に参加しました！')
        this.closeModal()
        window.location.reload()
      } catch (error) {
        alert(error.message || '参加に失敗しました。もう一度お試しください。')
        this.isJoining = false

        if (this.currentRoom.hasPassword) {
          this.password = ''
          this.$nextTick(() => {
            const passwordInput = this.$el.querySelector("input[type='password']")
            if (passwordInput) passwordInput.focus()
          })
        }
      }
    },

    // Escキーでモーダルを閉じる
    handleEsc(event) {
      if (event.key === 'Escape' && this.showModal) {
        this.closeModal()
      }
    },

    // 初期化処理
    init() {
      document.addEventListener('keydown', this.handleEsc.bind(this))
    },

    // クリーンアップ処理
    destroy() {
      document.removeEventListener('keydown', this.handleEsc.bind(this))
    },
  }
}
