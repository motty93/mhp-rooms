// 同期的にトークンを取得する関数
function getTokenSync() {
  try {
    // まずAlpineの認証ストアから取得を試行
    if (window.Alpine && window.Alpine.store('auth') && window.Alpine.store('auth').isAuthenticated) {
      const authStore = window.Alpine.store('auth')
      if (authStore.session && authStore.session.access_token) {
        return authStore.session.access_token
      }
    }

    // 次にSupabaseAuthから取得を試行
    if (window.supabaseAuth && window.supabaseAuth.session) {
      return window.supabaseAuth.session.access_token
    }

    // 最後に非同期メソッドで同期的に取得を試行（キャッシュされた値があれば即座に返される）
    if (window.supabaseAuth && typeof window.supabaseAuth.getAccessToken === 'function') {
      // 非同期メソッドだが、内部でキャッシュされている場合は同期的に取得可能
      const tokenPromise = window.supabaseAuth.getAccessToken()
      if (tokenPromise && typeof tokenPromise.then === 'function') {
        // Promiseの場合は諦める（非同期なので）
        return null
      } else {
        // 同期的に値が返された場合
        return tokenPromise
      }
    }

    return null
  } catch (error) {
    console.warn('同期トークン取得エラー:', error.message)
    return null
  }
}

function setupHtmxAuth() {
  // 同期的にイベントハンドラーを設定
  document.body.addEventListener('htmx:beforeRequest', (evt) => {
    // パス情報を取得（requestConfig.pathまたは古いdetail.pathを使用）
    const requestPath = evt.detail.requestConfig?.path || evt.detail.path

    if (evt.detail.xhr && requestPath) {
      // /api/または/rooms/で始まるパスに認証ヘッダーを追加
      if (requestPath.startsWith('/api/') || requestPath.startsWith('/rooms/')) {
        // 同期的にトークンを取得してヘッダーを設定
        const token = getTokenSync()
        if (token) {
          try {
            evt.detail.xhr.setRequestHeader('Authorization', `Bearer ${token}`)
          } catch (error) {
            console.warn('ヘッダー設定エラー:', error.message)
          }
        }
      }
    }
  })

  document.body.addEventListener('htmx:responseError', async (evt) => {
    if (evt.detail.xhr.status === 401) {
      if (window.Alpine && window.Alpine.store('auth')) {
        await window.Alpine.store('auth').checkAuth()

        if (!window.Alpine.store('auth').isAuthenticated) {
          if (confirm('セッションの有効期限が切れました。ログインページに移動しますか？')) {
            window.location.href = '/auth/login'
          }
        }
      }
    }
  })

  document.addEventListener('refresh-auth-token', async () => {
    if (window.supabaseClient) {
      try {
        const { data, error } = await window.supabaseClient.auth.refreshSession()
        if (error) throw error
      } catch (error) {
        console.error('トークンリフレッシュエラー:', error)
      }
    }
  })
}

// すぐに設定するか、DOMContentLoadedを待つ
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', setupHtmxAuth)
} else {
  setupHtmxAuth()
}

window.htmxAuthHeaders = async () => {
  const headers = {}

  try {
    if (window.supabaseAuth && typeof window.supabaseAuth.getAccessToken === 'function') {
      const token = await window.supabaseAuth.getAccessToken()
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }
    }
  } catch (error) {
    if (error.message.includes('認証機能が無効')) {
      // 無効化されている場合はログ出力を控える
    } else {
      console.warn('ヘッダー設定エラー:', error.message)
    }
  }

  return headers
}
