// プロフィール編集機能用JavaScript

/**
 * 通知を表示する関数
 * @param {string} message - 表示するメッセージ
 * @param {string} type - 通知のタイプ ('success', 'error', 'warning', 'info')
 */
function showNotification(message, type = 'info') {
  // 既存の通知があれば削除
  const existingNotification = document.querySelector('.notification')
  if (existingNotification) {
    existingNotification.remove()
  }

  // 通知要素を作成
  const notification = document.createElement('div')
  notification.className = `notification fixed top-4 right-4 z-50 p-4 rounded-lg shadow-lg transform transition-all duration-300 ease-in-out translate-x-full opacity-0`

  // タイプ別のスタイル設定
  let bgColor, textColor, icon
  switch (type) {
    case 'success':
      bgColor = 'bg-green-500'
      textColor = 'text-white'
      icon = '<i class="fa-solid fa-check-circle mr-2"></i>'
      break
    case 'error':
      bgColor = 'bg-red-500'
      textColor = 'text-white'
      icon = '<i class="fa-solid fa-exclamation-circle mr-2"></i>'
      break
    case 'warning':
      bgColor = 'bg-yellow-500'
      textColor = 'text-white'
      icon = '<i class="fa-solid fa-exclamation-triangle mr-2"></i>'
      break
    case 'info':
    default:
      bgColor = 'bg-blue-500'
      textColor = 'text-white'
      icon = '<i class="fa-solid fa-info-circle mr-2"></i>'
      break
  }

  notification.classList.add(bgColor, textColor)
  notification.innerHTML = `
        <div class="flex items-center">
            ${icon}
            <span>${message}</span>
            <button onclick="this.parentElement.parentElement.remove()" class="ml-4 text-white hover:text-gray-200">
                <i class="fa-solid fa-times"></i>
            </button>
        </div>
    `

  // bodyに追加
  document.body.appendChild(notification)

  // アニメーション開始
  setTimeout(() => {
    notification.classList.remove('translate-x-full', 'opacity-0')
  }, 100)

  // 5秒後に自動で削除
  setTimeout(() => {
    if (notification.parentElement) {
      notification.classList.add('translate-x-full', 'opacity-0')
      setTimeout(() => {
        if (notification.parentElement) {
          notification.remove()
        }
      }, 300)
    }
  }, 5000)
}

/**
 * プロフィール編集フォームのバリデーション
 * @param {Object} formData - フォームデータ
 * @returns {Object} バリデーション結果 {isValid: boolean, errors: string[]}
 */
function validateProfileForm(formData) {
  const errors = []

  // 表示名のバリデーション
  if (!formData.display_name || formData.display_name.trim() === '') {
    errors.push('表示名は必須です')
  } else if (formData.display_name.length > 100) {
    errors.push('表示名は100文字以内で入力してください')
  }

  // 自己紹介のバリデーション
  if (formData.bio && formData.bio.length > 500) {
    errors.push('自己紹介は500文字以内で入力してください')
  }

  // PSN IDのバリデーション
  if (formData.psn_online_id && formData.psn_online_id.length > 16) {
    errors.push('PSN IDは16文字以内で入力してください')
  }

  // Nintendo Network IDのバリデーション
  if (formData.nintendo_network_id && formData.nintendo_network_id.length > 16) {
    errors.push('Nintendo Network IDは16文字以内で入力してください')
  }

  // Nintendo Switch IDのバリデーション
  if (formData.nintendo_switch_id && formData.nintendo_switch_id.length > 20) {
    errors.push('Nintendo Switch IDは20文字以内で入力してください')
  }

  // Twitter IDのバリデーション
  if (formData.twitter_id && formData.twitter_id.length > 15) {
    errors.push('Twitter IDは15文字以内で入力してください')
  }

  return {
    isValid: errors.length === 0,
    errors: errors,
  }
}

/**
 * LocalStorageのユーザー情報を更新する
 * @param {string} userId - ユーザーID
 * @param {string} displayName - 新しい表示名
 */
function updateLocalStorageDisplayName(userId, displayName) {
  try {
    const dbUserKey = `mhp-rooms-dbuser-${userId}`
    const savedDbUser = localStorage.getItem(dbUserKey)

    if (savedDbUser) {
      const dbUser = JSON.parse(savedDbUser)
      dbUser.display_name = displayName
      localStorage.setItem(dbUserKey, JSON.stringify(dbUser))

    }
  } catch (error) {
    console.error('LocalStorage更新エラー:', error)
  }
}

/**
 * プロフィール更新API呼び出し
 * @param {Object} formData - 送信するフォームデータ
 * @returns {Promise} API呼び出しのPromise
 */
async function updateProfile(formData) {
  const response = await fetch('/api/profile/update', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(formData),
  })

  if (!response.ok) {
    const errorData = await response.json()
    throw new Error(errorData.error || 'プロフィールの更新に失敗しました')
  }

  return await response.json()
}

/**
 * プロフィール表示画面に戻る
 */
function returnToProfileView() {
  htmx.ajax('GET', '/profile/view', {
    target: '#profile-card-content',
    swap: 'innerHTML',
  })
}

/**
 * Alpine.js用のプロフィール編集フォームデータ関数
 * htmxで動的読み込み時にも使用可能
 */
window.profileEditForm = (userData = {}) => ({
  displayName: userData.displayName || '',
  bio: userData.bio || '',
  psnOnlineId: userData.psnOnlineId || '',
  nintendoNetworkId: userData.nintendoNetworkId || '',
  nintendoSwitchId: userData.nintendoSwitchId || '',
  twitterId: userData.twitterId || '',
  favoriteGames: userData.favoriteGames || [],
  playTimes: {
    weekday: userData.playTimes?.weekday || '',
    weekend: userData.playTimes?.weekend || '',
  },
  availableGames: ['MHP', 'MHP2', 'MHP2G', 'MHP3', 'MHRise', 'MHWorld', 'MHXX', 'MHNow'],
  maxBioLength: 500,

  init() {
    // お気に入りゲームが空の場合は空配列で初期化
    if (!this.favoriteGames) {
      this.favoriteGames = []
    }
  },

  // data属性から初期値を読み込む
  initFromDataAttributes(el) {
    // data属性から値を取得
    this.displayName = el.dataset.initDisplayName || ''
    this.bio = el.dataset.initBio || ''
    this.psnOnlineId = el.dataset.initPsnOnlineId || ''
    this.nintendoNetworkId = el.dataset.initNintendoNetworkId || ''
    this.nintendoSwitchId = el.dataset.initNintendoSwitchId || ''
    this.twitterId = el.dataset.initTwitterId || ''

    // JSONデータのパース
    try {
      this.favoriteGames = el.dataset.initFavoriteGames ? JSON.parse(el.dataset.initFavoriteGames) : []
    } catch (e) {
      this.favoriteGames = []
    }

    try {
      this.playTimes = el.dataset.initPlayTimes
        ? JSON.parse(el.dataset.initPlayTimes)
        : { weekday: '', weekend: '' }
    } catch (e) {
      this.playTimes = { weekday: '', weekend: '' }
    }

    this.init()
  },

  async handleAvatarChange(event) {
    const file = event.target.files[0]
    if (!file) return

    // ファイルタイプをチェック
    if (!file.type.startsWith('image/')) {
      showNotification('画像ファイルを選択してください', 'error')
      return
    }

    // ファイルサイズをチェック（10MB制限）
    if (file.size > 10 * 1024 * 1024) {
      showNotification('ファイルサイズは10MB以下にしてください', 'error')
      return
    }

    // プレビュー画像を更新
    const reader = new FileReader()
    reader.onload = (e) => {
      const imgElement = event.target.closest('.relative').querySelector('img')
      if (imgElement) {
        imgElement.src = e.target.result
      }
    }
    reader.readAsDataURL(file)

    // 実際のアップロード処理
    await this.uploadAvatar(file)
  },

  async uploadAvatar(file) {
    try {
      const formData = new FormData()
      formData.append('avatar', file)

      const response = await fetch('/api/profile/upload-avatar', {
        method: 'POST',
        body: formData,
      })

      if (response.ok) {
        const result = await response.json()
        showNotification(result.message || 'アバター画像を更新しました', 'success')

        // Alpine.jsのauth storeを更新（DBから最新情報を取得）
        if (window.Alpine && Alpine.store('auth')) {
          const authStore = Alpine.store('auth')
          // DBから最新のユーザー情報を取得してstoreを更新
          await authStore.refreshDbUser()
        }

        // プロフィール表示に戻る
        returnToProfileView()
      } else {
        const error = await response.json()
        showNotification(error.error || 'アバターの更新に失敗しました', 'error')
      }
    } catch (error) {
      console.error('アバターアップロードエラー:', error)
      showNotification('アバターの更新に失敗しました', 'error')
    }
  },

  async saveProfile() {
    try {
      const response = await fetch('/api/profile/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          display_name: this.displayName,
          bio: this.bio,
          psn_online_id: this.psnOnlineId,
          nintendo_network_id: this.nintendoNetworkId,
          nintendo_switch_id: this.nintendoSwitchId,
          twitter_id: this.twitterId,
          favorite_games: this.favoriteGames,
          play_times: this.playTimes,
        }),
      })

      if (response.ok) {
        const result = await response.json()

        // 通知を表示
        showNotification('プロフィールを更新しました', 'success')

        // Alpine.jsのauth storeを更新（DBから最新情報を取得）
        if (window.Alpine && Alpine.store('auth')) {
          const authStore = Alpine.store('auth')
          // DBから最新のユーザー情報を取得してstoreを更新
          await authStore.refreshDbUser()
        }

        // プロフィール表示に戻る
        returnToProfileView()
      } else {
        const error = await response.json()
        showNotification(error.error || 'プロフィールの更新に失敗しました', 'error')
      }
    } catch (error) {
      console.error('プロフィール更新エラー:', error)
      showNotification('プロフィールの更新に失敗しました', 'error')
    }
  },
})

// ユーザープロフィール画面の部屋参加処理
window.userProfileRoomsHandler = {
  currentRoomId: null,
  currentRoomName: '',
  currentRoomDescription: '',
  currentRoomGameVersion: '',
  currentRoomPlayerCount: '',
  currentRoomStatus: '',
  isJoining: false,
  hostRoomInfo: null,

  // 部屋モーダルを開く
  async openRoomModal(roomId, name, description, gameVersion, playerCount, status) {
    this.currentRoomId = roomId
    this.currentRoomName = name
    this.currentRoomDescription = description
    this.currentRoomGameVersion = gameVersion
    this.currentRoomPlayerCount = playerCount
    this.currentRoomStatus = status

    // 認証状態をチェック
    const authStore = Alpine.store('auth')
    if (!authStore || !authStore.isAuthenticated) {
      window.location.href = '/auth/login'
      return
    }

    // ホスト状態をチェック
    try {
      const response = await fetch('/api/user/current/room-status', {
        headers: {
          Authorization: `Bearer ${authStore.session.access_token}`,
        },
      })

      if (response.ok) {
        const data = await response.json()
        if (data.status === 'HOST') {
          // ホスト中の場合は制限モーダルを表示
          this.hostRoomInfo = data.room
          this.showHostRestrictionModal()
          return
        }
      }
    } catch (error) {
      console.error('部屋状態チェックエラー:', error)
    }

    // モーダルに情報を設定
    document.getElementById('user-profile-modal-title').textContent = name
    document.getElementById('user-profile-modal-description').textContent = description || ''
    document.getElementById('user-profile-modal-game-version').textContent = gameVersion
    document.getElementById('user-profile-modal-player-count').textContent = playerCount
    document.getElementById('user-profile-modal-status').textContent = status

    // モーダルを表示（アニメーション付き）
    const modal = document.getElementById('user-profile-join-modal')
    modal.style.display = 'flex'
    // 次のフレームで透明度を変更してアニメーション
    requestAnimationFrame(() => {
      modal.style.opacity = '1'
      const content = modal.querySelector('.bg-white')
      if (content) {
        content.style.transform = 'scale(1)'
        content.style.opacity = '1'
      }
    })
  },

  // モーダルを閉じる（アニメーション付き）
  closeModal() {
    const modal = document.getElementById('user-profile-join-modal')
    modal.style.opacity = '0'
    const content = modal.querySelector('.bg-white')
    if (content) {
      content.style.transform = 'scale(0.95)'
      content.style.opacity = '0'
    }
    // アニメーション完了後に非表示
    setTimeout(() => {
      modal.style.display = 'none'
    }, 200)
    this.currentRoomId = null
    this.isJoining = false
  },

  // 部屋に参加する
  async joinRoom() {
    if (this.isJoining || !this.currentRoomId) return

    const authStore = Alpine.store('auth')
    if (!authStore || !authStore.isAuthenticated) {
      this.closeModal()
      window.location.href = '/auth/login'
      return
    }

    this.isJoining = true

    try {
      const headers = {
        'Content-Type': 'application/json',
      }

      if (authStore.session?.access_token) {
        headers['Authorization'] = `Bearer ${authStore.session.access_token}`
      }

      const response = await fetch(`/rooms/${this.currentRoomId}/join`, {
        method: 'POST',
        headers: headers,
        body: JSON.stringify({}),
      })

      if (!response.ok) {
        if (response.status === 409) {
          const errorData = await response.json()

          // ホスト中の制限
          if (errorData.error === 'HOST_CANNOT_JOIN') {
            this.closeModal()
            this.hostRoomInfo = errorData.room
            this.showHostRestrictionModal()
            this.isJoining = false
            return
          }

          // 他の部屋に参加している場合の確認ダイアログ
          if (errorData.error === 'OTHER_ROOM_ACTIVE') {
            this.closeModal()
            this.showConfirmDialog()
            this.isJoining = false
            return
          }

          // 既に同じ部屋に参加している場合
          if (errorData.redirect) {
            window.location.href = errorData.redirect
            this.isJoining = false
            return
          }

          throw new Error(errorData.message || '参加に失敗しました')
        }

        throw new Error('参加に失敗しました')
      }

      const result = await response.json()

      // 成功時は部屋詳細画面に遷移
      if (result.redirect) {
        window.location.href = result.redirect
      } else {
        window.location.reload()
      }
    } catch (error) {
      showNotification(error.message || '参加に失敗しました', 'error')
      this.isJoining = false
    }
  },

  // ホスト制限モーダルを表示
  showHostRestrictionModal() {
    if (this.hostRoomInfo) {
      document.getElementById('user-profile-host-room-link').href = `/rooms/${this.hostRoomInfo.id}`
    }
    const modal = document.getElementById('user-profile-host-restriction-modal')
    modal.style.display = 'flex'
    requestAnimationFrame(() => {
      modal.style.opacity = '1'
      const content = modal.querySelector('.bg-white')
      if (content) {
        content.style.transform = 'scale(1)'
        content.style.opacity = '1'
      }
    })
  },

  // ホスト制限モーダルを閉じる
  closeHostRestrictionModal() {
    const modal = document.getElementById('user-profile-host-restriction-modal')
    modal.style.opacity = '0'
    const content = modal.querySelector('.bg-white')
    if (content) {
      content.style.transform = 'scale(0.95)'
      content.style.opacity = '0'
    }
    setTimeout(() => {
      modal.style.display = 'none'
    }, 200)
    this.hostRoomInfo = null
  },

  // 確認ダイアログを表示
  showConfirmDialog() {
    const modal = document.getElementById('user-profile-confirm-dialog')
    modal.style.display = 'flex'
    requestAnimationFrame(() => {
      modal.style.opacity = '1'
      const content = modal.querySelector('.bg-white')
      if (content) {
        content.style.transform = 'scale(1)'
        content.style.opacity = '1'
      }
    })
  },

  // 確認ダイアログを閉じる
  closeConfirmDialog() {
    const modal = document.getElementById('user-profile-confirm-dialog')
    modal.style.opacity = '0'
    const content = modal.querySelector('.bg-white')
    if (content) {
      content.style.transform = 'scale(0.95)'
      content.style.opacity = '0'
    }
    setTimeout(() => {
      modal.style.display = 'none'
    }, 200)
    this.isJoining = false
  },

  // 確認後に部屋に参加する
  async confirmJoinRoom() {
    if (this.isJoining || !this.currentRoomId) return

    const authStore = Alpine.store('auth')
    if (!authStore || !authStore.isAuthenticated) {
      this.closeConfirmDialog()
      window.location.href = '/auth/login'
      return
    }

    this.isJoining = true

    try {
      const requestData = {
        forceJoin: true,
      }

      const response = await fetch(`/rooms/${this.currentRoomId}/join`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authStore.session.access_token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || '参加に失敗しました')
      }

      const result = await response.json()

      // モーダルを閉じて部屋詳細に遷移
      this.closeConfirmDialog()
      this.closeModal()

      if (result.redirect) {
        window.location.href = result.redirect
      } else {
        window.location.reload()
      }
    } catch (error) {
      console.error('処理エラー:', error)
      showNotification(error.message || '参加に失敗しました', 'error')
      this.isJoining = false
    }
  },
}

// DOMContentLoaded時の初期化
document.addEventListener('DOMContentLoaded', () => {
  // プロフィール関連のイベントリスナーを設定
})

// htmx関連のイベントリスナー
document.addEventListener('htmx:afterRequest', (event) => {
  // プロフィール更新後の処理
  if (event.detail.xhr.status === 200 && event.detail.pathInfo.requestPath.includes('/profile/')) {
  }
})

document.addEventListener('htmx:responseError', (event) => {
  // htmxエラー時の処理
  if (event.detail.pathInfo.requestPath.includes('/profile/')) {
    showNotification('プロフィールの操作でエラーが発生しました', 'error')
  }
})
