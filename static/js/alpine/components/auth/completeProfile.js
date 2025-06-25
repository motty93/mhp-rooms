// プロフィール補完コンポーネント
import { validators } from '../common/formValidation.js';
import { apiClient } from '../common/apiClient.js';

export function completeProfile() {
  return {
    form: {
      psnId: ''
    },
    userInfo: {
      displayName: '',
      email: '',
      avatarUrl: ''
    },
    fieldError: false,
    errorText: '',
    errorMessage: '',
    successMessage: '',
    isLoading: false,

    init() {
      this.fetchUserInfo();
    },

    async fetchUserInfo() {
      try {
        const response = await apiClient.get('/api/user/current');
        if (response.ok) {
          const data = await response.json();
          this.userInfo = {
            displayName: data.display_name || data.email.split('@')[0],
            email: data.email,
            avatarUrl: data.avatar_url || ''
          };
        }
      } catch (error) {
        console.error('Failed to fetch user info:', error);
      }
    },

    validatePsnId() {
      const validation = validators.psnId(this.form.psnId);
      this.fieldError = !validation.valid;
      this.errorText = validation.message || '';
    },

    async handleSubmit() {
      this.errorMessage = '';
      this.successMessage = '';
      
      this.validatePsnId();
      if (this.fieldError) {
        return;
      }

      this.isLoading = true;
      
      try {
        const response = await apiClient.post('/auth/complete-profile', this.form);
        const data = await response.json();

        if (response.ok) {
          this.successMessage = 'プロフィールが完成しました！';
          setTimeout(() => {
            window.location.href = data.redirectUrl || '/rooms';
          }, 1500);
        } else {
          this.errorMessage = data.message || 'プロフィールの更新に失敗しました';
        }
      } catch (error) {
        this.errorMessage = 'ネットワークエラーが発生しました';
      } finally {
        this.isLoading = false;
      }
    },

    handleSkip() {
      if (confirm('PSN IDを設定せずに続けますか？一部の機能が制限される場合があります。')) {
        window.location.href = '/rooms';
      }
    }
  };
}