// Supabase初期化
let supabase;

// Supabase設定を取得して初期化
async function initializeSupabase() {
    try {
        // サーバーから環境変数を取得
        const response = await fetch('/api/config/supabase');
        const config = await response.json();
        
        if (!config.url || !config.anonKey) {
            throw new Error('Supabase設定が不完全です');
        }
        
        // Supabaseクライアントを初期化
        supabase = window.supabase.createClient(config.url, config.anonKey, {
            auth: {
                autoRefreshToken: true,
                persistSession: true,
                detectSessionInUrl: true
            }
        });
        
        // 認証状態の変更を監視
        supabase.auth.onAuthStateChange((event, session) => {
            console.log('認証状態が変更されました:', event);
            
            // Alpine.jsのグローバルストアに認証情報を設定
            if (window.Alpine && window.Alpine.store('auth')) {
                window.Alpine.store('auth').updateSession(session);
            }
            
            // htmxイベントを発火して認証状態の変更を通知
            document.body.dispatchEvent(new CustomEvent('auth-state-changed', { 
                detail: { event, session } 
            }));
        });
        
        // 初期セッションチェック
        const { data: { session } } = await supabase.auth.getSession();
        if (window.Alpine && window.Alpine.store('auth')) {
            window.Alpine.store('auth').updateSession(session);
        }
        
        console.log('Supabaseが初期化されました');
        return supabase;
    } catch (error) {
        console.error('Supabase初期化エラー:', error);
        throw error;
    }
}

// 認証関連のヘルパー関数
const auth = {
    // ログイン
    async signIn(email, password) {
        const { data, error } = await supabase.auth.signInWithPassword({
            email,
            password
        });
        
        if (error) throw error;
        return data;
    },
    
    // サインアップ
    async signUp(email, password, metadata = {}) {
        const { data, error } = await supabase.auth.signUp({
            email,
            password,
            options: {
                data: metadata
            }
        });
        
        if (error) throw error;
        return data;
    },
    
    // ログアウト
    async signOut() {
        const { error } = await supabase.auth.signOut();
        if (error) throw error;
    },
    
    // 現在のユーザー取得
    async getUser() {
        const { data: { user }, error } = await supabase.auth.getUser();
        if (error) throw error;
        return user;
    },
    
    // セッション取得
    async getSession() {
        const { data: { session }, error } = await supabase.auth.getSession();
        if (error) throw error;
        return session;
    },
    
    // パスワードリセットメール送信
    async resetPassword(email) {
        const { data, error } = await supabase.auth.resetPasswordForEmail(email, {
            redirectTo: `${window.location.origin}/auth/reset-password`
        });
        
        if (error) throw error;
        return data;
    },
    
    // パスワード更新
    async updatePassword(newPassword) {
        const { data, error } = await supabase.auth.updateUser({
            password: newPassword
        });
        
        if (error) throw error;
        return data;
    },
    
    // アクセストークン取得（API呼び出し用）
    async getAccessToken() {
        const session = await this.getSession();
        return session?.access_token || null;
    }
};

// DOMContentLoadedで自動初期化
document.addEventListener('DOMContentLoaded', () => {
    initializeSupabase().catch(console.error);
});

// グローバルに公開
window.initializeSupabase = initializeSupabase;
window.supabaseAuth = auth;