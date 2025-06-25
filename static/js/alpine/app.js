// Alpine.jsアプリケーション初期化
import { mobileMenuStore } from './stores/mobileMenu.js';
import { authStore } from './stores/auth.js';
import { completeProfile } from './components/auth/completeProfile.js';

// Alpine.jsコンポーネントの登録
window.Alpine = window.Alpine || {};
Alpine.data('completeProfile', completeProfile);

// Alpine.jsストアの登録
document.addEventListener('alpine:init', () => {
  // ストアの登録
  Alpine.store('mobileMenu', mobileMenuStore);
  Alpine.store('auth', authStore);
  
  // グローバル設定
  Alpine.store('config', {
    apiBaseUrl: '',
    version: '1.0.0'
  });

  // 認証ストアの初期化
  Alpine.store('auth').init();

  // デバッグヘルパー（開発環境のみ）
  if (window.location.hostname === 'localhost') {
    window.debug = {
      login: () => {
        Alpine.store('auth').login('debug_token_' + Date.now());
        location.reload();
      },
      logout: () => {
        Alpine.store('auth').logout();
        location.reload();
      },
      checkStatus: () => Alpine.store('auth').checkStatus()
    };
  }
});