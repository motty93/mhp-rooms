// モバイルメニューストア
export const mobileMenuStore = {
  open: false,

  toggle() {
    this.open = !this.open
  },

  close() {
    this.open = false
  },
}
