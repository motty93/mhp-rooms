import { authStore } from './stores/auth.js'
import { mobileMenuStore } from './stores/mobileMenu.js'

window.Alpine = window.Alpine || {}

document.addEventListener('alpine:init', () => {
  Alpine.store('mobileMenu', mobileMenuStore)
  Alpine.store('auth', authStore)

  // グローバル設定
  Alpine.store('config', {
    apiBaseUrl: '',
    version: '1.0.0',
  })

  // 認証ストアの初期化
  Alpine.store('auth').init()

  // デバッグヘルパー（開発環境のみ）
  if (window.location.hostname === 'localhost') {
    window.debug = {
      login: () => {
        Alpine.store('auth').login(`debug_token_ ${Date.now()}`)
        location.reload()
      },
      logout: () => {
        Alpine.store('auth').logout()
        location.reload()
      },
      checkStatus: () => Alpine.store('auth').checkStatus(),
    }
  }
})

// // すべてのコンポーネントが読み込まれた後にAlpineを手動で開始
// window.addEventListener('load', () => {
//   console.log('Starting Alpine.js manually')
//   window.Alpine.start()
// })
