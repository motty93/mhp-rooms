// 認証セッションをクリアする関数
async function clearAuthSession() {
  // クッキーを削除
  document.cookie = 'sb-access-token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT'

  // Supabaseセッションをクリア
  if (window.supabaseClient && window.supabaseClient.auth) {
    try {
      await window.supabaseClient.auth.signOut()
    } catch (error) {
      console.warn('Supabaseサインアウトエラー:', error)
    }
  }

  // Alpineストアをリセット
  if (window.Alpine && window.Alpine.store('auth')) {
    window.Alpine.store('auth').clearSession()
  }
}

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

  // HX-Redirectヘッダーの処理を追加
  document.body.addEventListener('htmx:beforeSwap', (evt) => {
    const xhr = evt.detail.xhr
    if (xhr) {
      const redirectUrl = xhr.getResponseHeader('HX-Redirect')
      if (redirectUrl) {
        // 認証関連のリダイレクトの場合はセッションをクリア
        if (redirectUrl.includes('/auth/login')) {
          clearAuthSession()
        }
        evt.detail.shouldSwap = false
        window.location.href = redirectUrl
      }
    }
  })

  document.body.addEventListener('htmx:responseError', async (evt) => {
    if (evt.detail.xhr.status === 401) {
      // 認証エラー時はセッションをクリアしてリダイレクト
      await clearAuthSession()
      window.location.href = '/auth/login'
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
