import { rooms } from './state.js'

function register() {
  if (!window.Alpine) return false

  // 1) コンポーネント登録（x-data="rooms" 用）
  window.Alpine.data('rooms', rooms)

  // 2) 旧テンプレ互換（x-data="rooms()" にも対応）
  window.rooms = rooms

  // 3) すでに初期化済みだった場合に備えて再スキャン
  //    （alpine:init を取り逃しても動くように保険をかける）
  queueMicrotask(() => {
    try {
      window.Alpine.initTree(document.body)
      console.log('[rooms] registered & re-initialized')
    } catch (e) {
      console.warn('[rooms] initTree failed (maybe not started yet):', e)
    }
  })

  return true
}

console.log('[rooms] index loaded')

// A) すでに Alpine が居れば即登録
if (!register()) {
  // B) まだなら、初期化イベントで登録
  document.addEventListener('alpine:init', () => {
    console.log('[rooms] alpine:init caught')
    register()
  })

  // C) DOM 構築後にもう一度試す（フォールバック）
  document.addEventListener('DOMContentLoaded', () => {
    setTimeout(() => {
      if (!register()) console.warn('[rooms] DOMContentLoaded: Alpine not ready')
    }, 0)
  })
}

// 方式B（deferLoadingAlpine）を使う場合のみ：ここで起動を解放
if (window._startAlpine) window._startAlpine()
