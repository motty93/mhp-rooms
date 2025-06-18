// 認証状態管理用のJavaScript

// 認証状態をチェックする関数（仮実装）
function checkAuthStatus() {
    // TODO: 実際のSupabase認証実装時に置き換える
    // 現在はlocalStorageまたはcookieで認証状態を確認する仮実装
    
    // セッションストレージまたはlocalStorageで認証状態を確認
    const authToken = localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token');
    const isAuthenticated = authToken !== null && authToken !== '';
    
    return isAuthenticated;
}

// UIを認証状態に応じて更新する関数
function updateAuthUI() {
    const isAuthenticated = checkAuthStatus();
    
    // ヘッダーの認証ボタンとユーザーメニューの表示切り替え
    const authButtons = document.getElementById('auth-buttons');
    const userMenu = document.getElementById('user-menu');
    const navMenu = document.getElementById('nav-menu');
    const mobileAuthButtons = document.getElementById('mobile-auth-buttons');
    const mobileUserMenu = document.getElementById('mobile-user-menu');
    
    // 部屋作成ボタンの表示切り替え
    const createRoomAuth = document.getElementById('create-room-auth');
    const createRoomUnauth = document.getElementById('create-room-unauth');
    const createRoomEmptyAuth = document.getElementById('create-room-empty-auth');
    const createRoomEmptyUnauth = document.getElementById('create-room-empty-unauth');
    
    if (isAuthenticated) {
        // 認証済みの場合
        if (authButtons) authButtons.style.display = 'none';
        if (userMenu) userMenu.classList.remove('hidden');
        if (navMenu) navMenu.style.display = 'flex';
        if (mobileAuthButtons) mobileAuthButtons.style.display = 'none';
        if (mobileUserMenu) mobileUserMenu.classList.remove('hidden');
        
        // 部屋作成ボタンを有効化
        if (createRoomAuth) createRoomAuth.classList.remove('hidden');
        if (createRoomUnauth) createRoomUnauth.style.display = 'none';
        if (createRoomEmptyAuth) createRoomEmptyAuth.classList.remove('hidden');
        if (createRoomEmptyUnauth) createRoomEmptyUnauth.style.display = 'none';
    } else {
        // 未認証の場合
        if (authButtons) authButtons.style.display = 'flex';
        if (userMenu) userMenu.classList.add('hidden');
        if (navMenu) navMenu.style.display = 'none';
        if (mobileAuthButtons) mobileAuthButtons.style.display = 'block';
        if (mobileUserMenu) mobileUserMenu.classList.add('hidden');
        
        // 部屋作成ボタンを無効化
        if (createRoomAuth) createRoomAuth.classList.add('hidden');
        if (createRoomUnauth) createRoomUnauth.style.display = 'inline-block';
        if (createRoomEmptyAuth) createRoomEmptyAuth.classList.add('hidden');
        if (createRoomEmptyUnauth) createRoomEmptyUnauth.style.display = 'inline-block';
    }
}

// ログイン処理（仮実装）
function login(token) {
    localStorage.setItem('auth_token', token);
    updateAuthUI();
}

// ログアウト処理（仮実装）
function logout() {
    localStorage.removeItem('auth_token');
    sessionStorage.removeItem('auth_token');
    updateAuthUI();
}

// 未認証ユーザーが部屋作成ボタンをクリックした時の処理
function handleUnauthenticatedCreateRoom() {
    alert('部屋を作成するにはログインが必要です。');
    // ログインページへリダイレクト（必要に応じて）
    // window.location.href = '/auth/login';
}

// ページ読み込み時に認証状態をチェック
document.addEventListener('DOMContentLoaded', function() {
    updateAuthUI();
    
    // 未認証用の部屋作成ボタンにクリックイベントを追加
    const createRoomUnauth = document.getElementById('create-room-unauth');
    const createRoomEmptyUnauth = document.getElementById('create-room-empty-unauth');
    
    if (createRoomUnauth) {
        createRoomUnauth.addEventListener('click', handleUnauthenticatedCreateRoom);
    }
    if (createRoomEmptyUnauth) {
        createRoomEmptyUnauth.addEventListener('click', handleUnauthenticatedCreateRoom);
    }
});

// 開発用：認証状態をテストするための関数
function debugLogin() {
    login('debug_token_' + Date.now());
    console.log('Debug: ログイン状態にしました');
}

function debugLogout() {
    logout();
    console.log('Debug: ログアウト状態にしました');
}

// コンソールからテスト可能にする（開発用）
window.debugAuth = {
    login: debugLogin,
    logout: debugLogout,
    checkStatus: checkAuthStatus
};