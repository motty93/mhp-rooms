document.addEventListener('alpine:init', () => {
    Alpine.store('auth', {
        user: null,
        session: null,
        loading: true,
        error: null,
        configError: null,
        initialized: false,
        
        init() {
            if (window.supabaseClient) {
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
                this.initialized = true;
            }
        },
        
        updateSession(session) {
            this.session = session;
            this.user = session?.user || null;
            this.error = null;
            
            if (this.user && session?.access_token) {
                this.syncUser(session.access_token);
            }
        },
        
        async syncUser(accessToken) {
            try {
                const response = await fetch('/api/auth/sync', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${accessToken}`
                    },
                    body: JSON.stringify({
                        psn_id: this.user?.user_metadata?.psn_id || ''
                    })
                });
                
                if (!response.ok) {
                    console.error('ユーザー同期に失敗しました:', response.status);
                }
            } catch (error) {
                console.error('ユーザー同期エラー:', error);
            }
        },
        
        get isAuthenticated() {
            return !!this.user;
        },
        
        get username() {
            return this.user?.email?.split('@')[0] || this.user?.user_metadata?.name || 'ゲスト';
        },

        get needsPSNId() {
            return this.isAuthenticated && (!this.user?.user_metadata?.psn_id);
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
        },
        
        async updatePSNId(psnId) {
            if (!this.session?.access_token) {
                throw new Error('認証が必要です');
            }
            
            this.loading = true;
            this.error = null;
            
            try {
                const response = await fetch('/api/auth/psn-id', {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${this.session.access_token}`
                    },
                    body: JSON.stringify({
                        psn_id: psnId
                    })
                });
                
                if (!response.ok) {
                    throw new Error('PSN IDの更新に失敗しました');
                }
                
                if (window.supabaseAuth && typeof window.supabaseAuth.updateUserMetadata === 'function') {
                    await window.supabaseAuth.updateUserMetadata({ psn_id: psnId });
                }
                
                // ユーザー情報を再取得
                await this.checkAuth();
                
                return response.json();
            } catch (error) {
                this.error = error.message;
                throw error;
            } finally {
                this.loading = false;
            }
        }
    });
});