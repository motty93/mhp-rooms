export async function fetchActiveGameVersions() {
  const response = await fetch('/api/game-versions/active', { credentials: 'same-origin' })
  if (!response.ok) throw new Error('ゲーム情報の取得に失敗')

  return response.json()
}

export async function fetchRooms(token) {
  const res = await fetch('/api/roosm', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  })
  if (!res.ok) throw new Error('ルーム情報の取得に失敗')

  return res.json()
}

export async function getUserRoomStatus(token) {
  const res = await fetch('/api/user/current/room-status', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  })
  if (!res.ok) throw new Error('部屋状態の取得に失敗')

  return res.json()
}

export async function joinRoom(roomId, body, tokenOpt) {
  const headers = { 'Content-Type': 'application/json' }
  if (tokenOpt) {
    headers.Authorization = `Bearer ${tokenOpt}`
  }

  const res = await fetch(`/rooms/${roomId}/join`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body || {}),
  })

  return res
}

export async function createRoom(payload, token) {
  const res = await fetch('/rooms', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(payload),
  })

  return res
}
