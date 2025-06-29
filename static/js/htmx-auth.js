// htmxのAPI呼び出し時にJWTトークンを自動付与
document.addEventListener('DOMContentLoaded', () => {
    // htmxのbeforeRequestイベントで認証ヘッダーを追加
    document.body.addEventListener('htmx:beforeRequest', async (evt) => {
        // APIエンドポイントへのリクエストの場合のみ処理
        if (evt.detail.xhr && evt.detail.path && evt.detail.path.startsWith('/api/')) {
            try {
                // Supabaseからアクセストークンを取得
                if (window.supabaseAuth) {
                    const token = await window.supabaseAuth.getAccessToken();
                    
                    if (token) {
                        // Authorizationヘッダーを設定
                        evt.detail.xhr.setRequestHeader('Authorization', `Bearer ${token}`);
                    }
                }
            } catch (error) {
                console.error('トークン取得エラー:', error);
            }
        }
    });
    
    // htmxのresponseErrorイベントで401エラーをハンドリング
    document.body.addEventListener('htmx:responseError', async (evt) => {
        if (evt.detail.xhr.status === 401) {
            // 認証エラーの場合
            console.log('認証エラーが発生しました');
            
            // Alpine.jsの認証ストアを更新
            if (window.Alpine && window.Alpine.store('auth')) {
                // セッションを再チェック
                await window.Alpine.store('auth').checkAuth();
                
                // 認証されていない場合はログインページへ
                if (!window.Alpine.store('auth').isAuthenticated) {
                    if (confirm('セッションの有効期限が切れました。ログインページに移動しますか？')) {
                        window.location.href = '/auth/login';
                    }
                }
            }
        }
    });
    
    // カスタムイベント: 強制的にトークンをリフレッシュ
    document.addEventListener('refresh-auth-token', async () => {
        if (window.supabase) {
            try {
                const { data, error } = await window.supabase.auth.refreshSession();
                if (error) throw error;
                
                console.log('トークンがリフレッシュされました');
            } catch (error) {
                console.error('トークンリフレッシュエラー:', error);
            }
        }
    });
});

// htmxのヘッダー設定ヘルパー関数
window.htmxAuthHeaders = async () => {
    const headers = {};
    
    try {
        if (window.supabaseAuth) {
            const token = await window.supabaseAuth.getAccessToken();
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }
        }
    } catch (error) {
        console.error('ヘッダー設定エラー:', error);
    }
    
    return headers;
};