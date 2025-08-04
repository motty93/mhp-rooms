let supabase

async function initializeSupabase() {
  try {
    const response = await fetch('/api/config/supabase')
    const data = await response.json()

    if (data.error || !data.config || !data.config.url || !data.config.anonKey) {
      // 認証機能を無効化

      window.supabaseAuth = createDummyAuth()

      if (window.Alpine && window.Alpine.store('auth')) {
        window.Alpine.store('auth').updateSession(null)
        window.Alpine.store('auth').configError = data.message || 'Supabase設定が未設定です'
      }

      return null
    }

    window.supabaseClient = window.supabase.createClient(data.config.url, data.config.anonKey, {
      auth: {
        autoRefreshToken: true,
        persistSession: true,
        detectSessionInUrl: true,
      },
    })

    supabase = window.supabaseClient

    supabase.auth.onAuthStateChange((event, session) => {
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

    const {
      data: { session },
    } = await supabase.auth.getSession()
    if (window.Alpine && window.Alpine.store('auth')) {
      window.Alpine.store('auth').updateSession(session)
    }

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
    signIn: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    signUp: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    signOut: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    getUser: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    getSession: async () => {
      return null
    },
    resetPassword: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
    },
    updatePassword: async () => {
      throw new Error('認証機能が無効です。Supabase設定を確認してください。')
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
  async signIn(email, password) {
    const { data, error } = await supabase.auth.signInWithPassword({
      email,
      password,
    })

    if (error) throw error
    return data
  },

  async signUp(email, password, metadata = {}) {
    const { data, error } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: metadata,
      },
    })

    if (error) throw error
    return data
  },

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

  async resetPassword(email) {
    const { data, error } = await supabase.auth.resetPasswordForEmail(email, {
      redirectTo: `${window.location.origin}/auth/reset-password`,
    })

    if (error) throw error
    return data
  },

  async updatePassword(newPassword) {
    const { data, error } = await supabase.auth.updateUser({
      password: newPassword,
    })

    if (error) throw error
    return data
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
