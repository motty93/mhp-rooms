document.addEventListener('alpine:init', () => {
    Alpine.store('auth', {
        user: null,
        session: null,
        loading: true,
        error: null,
        configError: null,
        
        init() {
            if (window.supabase) {
                this.checkAuth();
            } else {
                document.addEventListener('supabase-initialized', () => {
                    this.checkAuth();
                });
            }
        },
        
        async checkAuth() {
            this.loading = true;
            this.error = null;
            
            try {
                if (window.supabaseAuth) {
                    const session = await window.supabaseAuth.getSession();
                    this.updateSession(session);
                } else {
                    this.updateSession(null);
                }
            } catch (error) {
                console.error('認証状態チェックエラー:', error);
                this.error = error.message;
                this.updateSession(null);
            } finally {
                this.loading = false;
            }
        },
        
        updateSession(session) {
            this.session = session;
            this.user = session?.user || null;
            this.error = null;
        },
        
        get isAuthenticated() {
            return !!this.user;
        },
        
        get username() {
            return this.user?.email?.split('@')[0] || this.user?.user_metadata?.name || 'ゲスト';
        },
        
        async signIn(email, password) {
            if (!window.supabaseAuth) {
                throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。');
            }
            
            this.loading = true;
            this.error = null;
            
            try {
                const data = await window.supabaseAuth.signIn(email, password);
                return data;
            } catch (error) {
                this.error = error.message;
                throw error;
            } finally {
                this.loading = false;
            }
        },
        
        async signUp(email, password, metadata = {}) {
            if (!window.supabaseAuth) {
                throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。');
            }
            
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
        
        async signOut() {
            if (!window.supabaseAuth) {
                window.location.href = '/';
                return;
            }
            
            this.loading = true;
            this.error = null;
            
            try {
                await window.supabaseAuth.signOut();
                window.location.href = '/';
            } catch (error) {
                this.error = error.message;
                window.location.href = '/';
            } finally {
                this.loading = false;
            }
        },
        
        async resetPassword(email) {
            if (!window.supabaseAuth) {
                throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。');
            }
            
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
        },
        
        async signInWithGoogle() {
            if (!window.supabaseAuth) {
                throw new Error('認証システムが初期化されていません。Supabase設定を確認してください。');
            }
            
            this.loading = true;
            this.error = null;
            
            try {
                const data = await window.supabaseAuth.signInWithGoogle();
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