(() => {
  const getMeasurementId = () => window.huntershubAnalytics?.measurementId || ''
  const hasMeasurement = () => getMeasurementId() !== ''
  const hasGtag = () => typeof window.gtag === 'function'

  const eventQueue = []
  let flushScheduled = false

  const ensureBeacon = (params = {}) => {
    if (!params.transport_type) {
      params.transport_type = 'beacon'
    }
    return params
  }

  const enqueue = (...args) => {
    eventQueue.push(args)
    scheduleFlush()
  }

  const scheduleFlush = () => {
    if (flushScheduled) {
      return
    }
    flushScheduled = true

    const attemptFlush = () => {
      if (!hasGtag()) {
        setTimeout(attemptFlush, 250)
        return
      }

      while (eventQueue.length > 0) {
        const callArgs = eventQueue.shift()
        window.gtag(...callArgs)
      }

      flushScheduled = false
    }

    attemptFlush()
  }

  const trackEvent = (eventName, params = {}) => {
    if (!hasMeasurement()) {
      return
    }

    const payload = ensureBeacon(params)

    if (hasGtag()) {
      window.gtag('event', eventName, payload)
      return
    }

    enqueue('event', eventName, payload)
  }

  const trackConfig = (measurementId, params = {}) => {
    if (!hasMeasurement()) {
      return
    }

    const payload = ensureBeacon(params)

    if (hasGtag()) {
      window.gtag('config', measurementId, payload)
      return
    }

    enqueue('config', measurementId, payload)
  }

  const Analytics = {
    isEnabled() {
      return hasMeasurement()
    },

    track(eventName, params = {}) {
      trackEvent(eventName, params)
    },

    trackPageView(path) {
      trackConfig(getMeasurementId(), {
        page_path: path,
        transport_type: 'beacon',
      })
    },

    trackRoomCreate(roomId, gameVersion, maxPlayers) {
      if (!roomId) {
        return
      }

      const params = {
        room_id: roomId,
      }

      if (gameVersion) {
        params.game_version = gameVersion
      }

      if (Number.isFinite(Number(maxPlayers))) {
        params.max_players = Number(maxPlayers)
      }

      trackEvent('room_create', params)
    },

    trackRoomJoin(roomId, gameVersion) {
      if (!roomId) {
        return
      }

      const params = {
        room_id: roomId,
      }

      if (gameVersion) {
        params.game_version = gameVersion
      }

      trackEvent('room_join', params)
    },

    trackRoomLeave(roomId, durationSeconds) {
      if (!roomId) {
        return
      }

      const params = {
        room_id: roomId,
      }

      if (Number.isFinite(Number(durationSeconds))) {
        params.session_duration = Number(durationSeconds)
      }

      trackEvent('room_leave', params)
    },

    trackSignup(method) {
      const params = {}
      if (method) {
        params.method = method
      }
      trackEvent('sign_up', params)
    },

    trackLogin(method) {
      const params = {}
      if (method) {
        params.method = method
      }
      trackEvent('login', params)
    },

    trackProfileEdit() {
      trackEvent('profile_edit')
    },

    trackGameFilter(gameVersion) {
      const params = {}
      if (gameVersion) {
        params.game_version = gameVersion
      }
      trackEvent('filter_game_version', params)
    },

    trackSearch(term) {
      if (!term) {
        return
      }
      trackEvent('search', {
        search_term: term,
      })
    },
  }

  const attachAuthListener = () => {
    const target = document.body
    if (!target) {
      return
    }

    target.addEventListener('auth-state-changed', (event) => {
      const detail = event.detail
      if (!detail || !detail.event) {
        return
      }

      if (detail.event === 'SIGNED_IN') {
        const provider = detail.session?.user?.app_metadata?.provider || 'unknown'
        Analytics.trackLogin(provider)
      }

      if (detail.event === 'SIGNED_UP') {
        const provider = detail.session?.user?.app_metadata?.provider || 'unknown'
        Analytics.trackSignup(provider)
      }
    })
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', attachAuthListener)
  } else {
    attachAuthListener()
  }

  window.Analytics = Analytics
})()
