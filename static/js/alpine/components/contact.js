export function contactForm() {
  return {
    formData: {
      inquiryType: '',
      name: '',
      email: '',
      subject: '',
      message: '',
      privacyAgreed: false,
    },
    fieldErrors: {
      name: false,
      email: false,
      subject: false,
      message: false,
    },
    fieldErrorMessages: {
      name: '',
      email: '',
      subject: '',
      message: '',
    },
    isFormValid: false,
    isSubmitting: false,
    showSuccessMessage: false,
    showErrorMessage: false,
    errorMessage: '',

    validateField(fieldName) {
      this.fieldErrors[fieldName] = false
      this.fieldErrorMessages[fieldName] = ''

      switch (fieldName) {
        case 'name':
          if (this.formData.name.trim() === '') {
            this.fieldErrors.name = true
            this.fieldErrorMessages.name = 'お名前を入力してください'
          } else if (this.formData.name.trim().length > 100) {
            this.fieldErrors.name = true
            this.fieldErrorMessages.name = 'お名前は100文字以内で入力してください'
          }
          break

        case 'email':
          if (this.formData.email.trim() === '') {
            this.fieldErrors.email = true
            this.fieldErrorMessages.email = 'メールアドレスを入力してください'
          } else if (!this.isValidEmail(this.formData.email.trim())) {
            this.fieldErrors.email = true
            this.fieldErrorMessages.email = '正しいメールアドレスの形式で入力してください'
          }
          break

        case 'subject':
          if (this.formData.subject.trim() === '') {
            this.fieldErrors.subject = true
            this.fieldErrorMessages.subject = '件名を入力してください'
          } else if (this.formData.subject.trim().length > 200) {
            this.fieldErrors.subject = true
            this.fieldErrorMessages.subject = '件名は200文字以内で入力してください'
          }
          break

        case 'message':
          if (this.formData.message.trim() === '') {
            this.fieldErrors.message = true
            this.fieldErrorMessages.message = 'お問い合わせ内容を入力してください'
          } else if (this.formData.message.trim().length < 10) {
            this.fieldErrors.message = true
            this.fieldErrorMessages.message = 'お問い合わせ内容は10文字以上で入力してください'
          } else if (this.formData.message.length > 2000) {
            this.fieldErrors.message = true
            this.fieldErrorMessages.message = 'お問い合わせ内容は2000文字以内で入力してください'
          }
          break
      }

      this.updateFormValidation()
    },

    updateFormValidation() {
      this.isFormValid =
        this.formData.inquiryType !== '' &&
        this.formData.name.trim() !== '' &&
        this.formData.name.trim().length <= 100 &&
        this.formData.email.trim() !== '' &&
        this.isValidEmail(this.formData.email.trim()) &&
        this.formData.subject.trim() !== '' &&
        this.formData.subject.trim().length <= 200 &&
        this.formData.message.trim() !== '' &&
        this.formData.message.trim().length >= 10 &&
        this.formData.message.length <= 2000 &&
        this.formData.privacyAgreed
    },

    isValidEmail(email) {
      const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/

      if (email.length > 254) return false
      if (!emailRegex.test(email)) return false
      if ((email.match(/@/g) || []).length !== 1) return false

      const parts = email.split('@')
      if (parts.length !== 2) return false

      const localPart = parts[0]
      const domainPart = parts[1]

      if (localPart.length === 0 || localPart.length > 64) return false
      if (domainPart.length === 0 || domainPart.length > 253) return false
      if (domainPart.startsWith('.') || domainPart.endsWith('.')) return false
      if (domainPart.includes('..')) return false

      return true
    },

    async submitForm() {
      if (!this.isFormValid || this.isSubmitting) return

      this.isSubmitting = true
      this.showSuccessMessage = false
      this.showErrorMessage = false

      try {
        const response = await fetch('/contact', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(this.formData),
        })

        if (!response.ok) {
          throw new Error('送信に失敗しました')
        }

        const result = await response.json()

        this.showSuccessMessage = true
        this.resetForm()

        this.$nextTick(() => {
          const successElement = this.$el.querySelector('.bg-green-50')
          if (successElement) {
            successElement.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
          }
        })
      } catch (error) {
        this.showErrorMessage = true
        this.errorMessage = error.message || 'エラーが発生しました。しばらく後に再度お試しください。'
      } finally {
        this.isSubmitting = false
      }
    },

    resetForm() {
      this.formData = {
        inquiryType: '',
        name: '',
        email: '',
        subject: '',
        message: '',
        privacyAgreed: false,
      }
      this.fieldErrors = {
        name: false,
        email: false,
        subject: false,
        message: false,
      }
      this.fieldErrorMessages = {
        name: '',
        email: '',
        subject: '',
        message: '',
      }
      this.isFormValid = false
    },
  }
}
