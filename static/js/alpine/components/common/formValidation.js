// フォームバリデーション共通関数
export const validators = {
  email(value) {
    if (!value) return { valid: false, message: 'メールアドレスを入力してください' };
    const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!regex.test(value)) return { valid: false, message: '有効なメールアドレスを入力してください' };
    return { valid: true };
  },

  psnId(value) {
    if (!value) return { valid: false, message: 'PSN IDを入力してください' };
    if (value.length < 3) return { valid: false, message: 'PSN IDは3文字以上で入力してください' };
    if (value.length > 16) return { valid: false, message: 'PSN IDは16文字以内で入力してください' };
    if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
      return { valid: false, message: 'PSN IDは英数字、ハイフン、アンダースコアのみ使用できます' };
    }
    return { valid: true };
  },

  password(value) {
    if (!value) return { valid: false, message: 'パスワードを入力してください' };
    if (value.length < 6) return { valid: false, message: 'パスワードは6文字以上で入力してください' };
    return { valid: true };
  },

  required(value, fieldName = 'この項目') {
    if (!value || value.trim() === '') {
      return { valid: false, message: `${fieldName}を入力してください` };
    }
    return { valid: true };
  }
};