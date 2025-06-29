document.addEventListener('DOMContentLoaded', () => {
    document.body.addEventListener('htmx:beforeRequest', async (evt) => {
        if (evt.detail.xhr && evt.detail.path && evt.detail.path.startsWith('/api/')) {
            try {
                if (window.supabaseAuth && typeof window.supabaseAuth.getAccessToken === 'function') {
                    const token = await window.supabaseAuth.getAccessToken();
                    
                    if (token) {
                        evt.detail.xhr.setRequestHeader('Authorization', `Bearer ${token}`);
                    }
                }
            } catch (error) {
                // 認証が無効な場合は通常の動作なので、ログレベルを下げる
                if (error.message.includes('認証機能が無効')) {
                    // 無効化されている場合はログ出力を控える
                } else {
                    console.warn('トークン取得エラー:', error.message);
                }
            }
        }
    });
    
    document.body.addEventListener('htmx:responseError', async (evt) => {
        if (evt.detail.xhr.status === 401) {
            if (window.Alpine && window.Alpine.store('auth')) {
                await window.Alpine.store('auth').checkAuth();
                
                if (!window.Alpine.store('auth').isAuthenticated) {
                    if (confirm('セッションの有効期限が切れました。ログインページに移動しますか？')) {
                        window.location.href = '/auth/login';
                    }
                }
            }
        }
    });
    
    document.addEventListener('refresh-auth-token', async () => {
        if (window.supabase) {
            try {
                const { data, error } = await window.supabase.auth.refreshSession();
                if (error) throw error;
            } catch (error) {
                console.error('トークンリフレッシュエラー:', error);
            }
        }
    });
});

window.htmxAuthHeaders = async () => {
    const headers = {};
    
    try {
        if (window.supabaseAuth && typeof window.supabaseAuth.getAccessToken === 'function') {
            const token = await window.supabaseAuth.getAccessToken();
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }
        }
    } catch (error) {
        if (error.message.includes('認証機能が無効')) {
            // 無効化されている場合はログ出力を控える
        } else {
            console.warn('ヘッダー設定エラー:', error.message);
        }
    }
    
    return headers;
};