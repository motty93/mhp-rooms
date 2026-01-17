export function getAuthStore() {
  return window.Alpine?.store('auth')
}

export function getAuthToken(required = false) {
  const store = getAuthStore()
  const token = store?.session?.access_token
  if (required && (!store?.isAuthenticated || !token)) throw new Error('認証が必要です')

  return token || null
}

export function normalizeRooms(arr) {
  return (arr || []).map((room) => ({
    id: room.id,
    name: room.name,
    description: room.description || '',
    gameVersion: {
      code: room.game_version?.code || room.gameVersion?.code || '',
      name: room.game_version?.name || room.gameVersion?.name || '',
    },
    host: {
      username: room.host?.username || '',
      displayName: room.host?.display_name || room.host?.displayName || '',
    },
    currentPlayers: room.current_players ?? room.currentPlayers ?? 0,
    maxPlayers: room.max_players ?? room.maxPlayers ?? 4,
    isClosed: room.is_closed ?? room.isClosed ?? false,
    hasPassword: !!(room.has_password ?? room.password_hash ?? room.hasPassword),
    targetMonster: room.target_monster ?? room.targetMonster ?? '',
    rankRequirement: room.rank_requirement ?? room.rankRequirement ?? '',
    isJoined: room.is_joined ?? room.isJoined ?? false,
  }))
}

export function redirectOrReload(result) {
  if (result?.redirect) {
    window.location.href = result.redirect
  } else {
    window.location.reload()
  }
}

export function parseJsonScript(id) {
  const node = document.getElementById(id)
  if (!node) return null

  try {
    return JSON.parse(node.textContent || 'null')
  } catch {
    return null
  }
}

export function coerceId(v) {
  if (typeof v === 'number') return v

  const s = String(v || '').trim()
  return /^\d+$/.test(s) ? Number.parseInt(s, 10) : s // 数字だけなら数値、そうでなければ文字列（UUID）
}
