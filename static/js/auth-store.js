// Alpine.js認証ストア
document.addEventListener('alpine:init', () => {
    Alpine.store('auth', {
        // 認証状態
        user: null,
        session: null,
        loading: true,
        error: null,
        
        // 初期化
        init() {
            // Supabaseが初期化されたら認証状態を取得
            if (window.supabase) {
                this.checkAuth();
            } else {
                // Supabaseが初期化されるまで待機
                document.addEventListener('supabase-initialized', () => {
                    this.checkAuth();
                });
            }
        },
        
        // 認証状態チェック
        async checkAuth() {
            this.loading = true;
            try {
                if (window.supabaseAuth) {
                    const session = await window.supabaseAuth.getSession();
                    this.updateSession(session);
                }
            } catch (error) {
                console.error('認証状態チェックエラー:', error);
                this.error = error.message;
            } finally {
                this.loading = false;
            }
        },
        
        // セッション更新
        updateSession(session) {
            this.session = session;
            this.user = session?.user || null;
            this.error = null;
        },
        
        // ログイン状態判定
        get isAuthenticated() {
            return !!this.user;
        },
        
        // ユーザー名取得
        get username() {
            return this.user?.email?.split('@')[0] || this.user?.user_metadata?.name || 'ゲスト';
        },
        
        // ログイン
        async signIn(email, password) {
            this.loading = true;
            this.error = null;
            
            try {
                const data = await window.supabaseAuth.signIn(email, password);
                // 認証成功後、onAuthStateChangeで自動的にセッションが更新される
                return data;
            } catch (error) {
                this.error = error.message;
                throw error;
            } finally {
                this.loading = false;
            }
        },
        
        // サインアップ
        async signUp(email, password, metadata = {}) {
            this.loading = true;
            this.error = null;
            
            try {
                const data = await window.supabaseAuth.signUp(email, password, metadata);
                return data;
            } catch (error) {
                this.error = error.message;
                throw error;
            } finally {
                this.loading = false;
            }
        },
        
        // ログアウト
        async signOut() {
            this.loading = true;
            this.error = null;
            
            try {
                await window.supabaseAuth.signOut();
                // 認証解除後、onAuthStateChangeで自動的にセッションがクリアされる
                
                // トップページへリダイレクト
                window.location.href = '/';
            } catch (error) {
                this.error = error.message;
                throw error;
            } finally {
                this.loading = false;
            }
        },
        
        // パスワードリセット
        async resetPassword(email) {
            this.loading = true;
            this.error = null;
            
            try {
                const data = await window.supabaseAuth.resetPassword(email);
                return data;
            } catch (error) {
                this.error = error.message;
                throw error;
            } finally {
                this.loading = false;
            }
        }
    });
});