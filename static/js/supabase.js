let supabase

async function initializeSupabase() {
  try {
    // テンプレートから埋め込まれた設定を使用
    const config = window.SUPABASE_CONFIG || {}

    if (!config.url || !config.anonKey) {
      // 認証機能を無効化
      window.supabaseAuth = createDummyAuth()

      if (window.Alpine && window.Alpine.store('auth')) {
        window.Alpine.store('auth').updateSession(null)
        window.Alpine.store('auth').configError = 'Supabase設定が未設定です'
      }

      return null
    }

    window.supabaseClient = window.supabase.createClient(config.url, config.anonKey, {
      auth: {
        autoRefreshToken: true,
        persistSession: true,
        detectSessionInUrl: true,
      },
    })

    supabase = window.supabaseClient

    // 初期セッションを取得
    const {
      data: { session },
    } = await supabase.auth.getSession()

    // 初期セッションを設定（セッションがある場合はクッキーも設定）
    if (session && session.access_token) {
      // アクセストークンをクッキーに保存（SSR用）
      document.cookie = `sb-access-token=${session.access_token}; path=/; max-age=3600; SameSite=Lax`
    }

    if (window.Alpine && window.Alpine.store('auth')) {
      window.Alpine.store('auth').updateSession(session)
    }

    // 認証状態の変更を監視（初期化時は呼ばれない）
    supabase.auth.onAuthStateChange((event, session) => {
      // 初期化イベント（INITIAL_SESSION）は無視する
      if (event === 'INITIAL_SESSION') {
        return
      }

      // セッション変更時にクッキーを設定/削除
      if (session && session.access_token) {
        // アクセストークンをクッキーに保存（SSR用）
        document.cookie = `sb-access-token=${session.access_token}; path=/; max-age=3600; SameSite=Lax`
      } else {
        // ログアウト時はクッキーを削除
        document.cookie = 'sb-access-token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT'
      }

      if (window.Alpine && window.Alpine.store('auth')) {
        window.Alpine.store('auth').updateSession(session)
      }

      document.body.dispatchEvent(
        new CustomEvent('auth-state-changed', {
          detail: { event, session },
        }),
      )
    })

    window.supabaseAuth = auth
    return supabase
  } catch (error) {
    console.error('Supabase初期化エラー:', error)

    window.supabaseAuth = createDummyAuth()

    if (window.Alpine && window.Alpine.store('auth')) {
      window.Alpine.store('auth').updateSession(null)
      window.Alpine.store('auth').configError = 'Supabase初期化に失敗しました'
    }

    return null
  }
}

function createDummyAuth() {
  return {
    signOut: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    getUser: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    getSession: async () => {
      return null
    },
    getAccessToken: async () => {
      return null
    },
    signInWithGoogle: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    updateUserMetadata: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
  }
}

const auth = {
  async signOut() {
    const { error } = await supabase.auth.signOut()
    if (error) throw error
  },

  async getUser() {
    const {
      data: { user },
      error,
    } = await supabase.auth.getUser()
    if (error) throw error
    return user
  },

  async getSession() {
    const {
      data: { session },
      error,
    } = await supabase.auth.getSession()
    if (error) throw error
    return session
  },

  async getAccessToken() {
    const session = await this.getSession()
    return session?.access_token || null
  },

  async signInWithGoogle() {
    const { data, error } = await supabase.auth.signInWithOAuth({
      provider: 'google',
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
        skipBrowserRedirect: false,
      },
    })

    if (error) {
      console.error('Google認証エラー:', error)
      throw error
    }

    return data
  },

  async updateUserMetadata(metadata) {
    const { data, error } = await supabase.auth.updateUser({
      data: metadata,
    })

    if (error) throw error
    return data
  },
}

window.initializeSupabase = initializeSupabase
