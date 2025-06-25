// 認証ストア
export const authStore = {
  isAuthenticated: false,
  user: null,

  init() {
    this.checkStatus();
  },

  checkStatus() {
    const authToken = localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token');
    this.isAuthenticated = authToken !== null && authToken !== '';
    if (this.isAuthenticated && !this.user) {
      this.user = { name: 'ユーザー名' }; // 仮実装
    }
  },

  login(token, user = null) {
    localStorage.setItem('auth_token', token);
    this.isAuthenticated = true;
    this.user = user || { name: 'ユーザー名' };
  },

  logout() {
    localStorage.removeItem('auth_token');
    sessionStorage.removeItem('auth_token');
    this.isAuthenticated = false;
    this.user = null;
  },

  handleUnauthenticatedAction() {
    alert('この機能を利用するにはログインが必要です。');
    window.location.href = '/auth/login';
  }
};