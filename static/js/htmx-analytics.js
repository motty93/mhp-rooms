(() => {
  const attachListeners = () => {
    const target = document.body
    if (!target || !window.Analytics || !window.Analytics.isEnabled()) {
      return
    }

    target.addEventListener('htmx:afterSettle', () => {
      const path = window.location.pathname + window.location.search
      window.Analytics.trackPageView(path)
    })

    target.addEventListener('htmx:responseError', (event) => {
      const status = event.detail?.xhr?.status
      window.Analytics.track('exception', {
        description: status ? `htmx error: ${status}` : 'htmx error',
        fatal: false,
      })
    })
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', attachListeners)
  } else {
    attachListeners()
  }
})()
