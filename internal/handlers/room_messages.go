package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mhp-rooms/internal/infrastructure/sse"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoomMessageHandler struct {
	BaseHandler
	hub *sse.Hub
}

func NewRoomMessageHandler(repo *repository.Repository, hub *sse.Hub) *RoomMessageHandler {
	return &RoomMessageHandler{
		BaseHandler: BaseHandler{
			repo: repo,
		},
		hub: hub,
	}
}

// SendMessage はメッセージを送信
func (h *RoomMessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// URLパラメータから部屋IDを取得
	roomIDStr := chi.URLParam(r, "id")

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "無効な部屋IDです", http.StatusBadRequest)
		return
	}

	// 認証チェック（DBユーザー情報を取得）
	user, ok := middleware.GetDBUserFromContext(r.Context())
	if !ok || user == nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// 部屋のメンバーチェック
	if !h.repo.Room.IsUserJoinedRoom(roomID, user.ID) {
		http.Error(w, "部屋のメンバーではありません", http.StatusForbidden)
		return
	}

	// フォームデータの取得
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	messageText := strings.TrimSpace(r.FormValue("message"))
	if messageText == "" {
		http.Error(w, "メッセージが空です", http.StatusBadRequest)
		return
	}

	// メッセージ長制限（1000文字）
	if len(messageText) > 1000 {
		http.Error(w, "メッセージは1000文字以内で入力してください", http.StatusBadRequest)
		return
	}

	// メッセージを作成
	message := &models.RoomMessage{
		RoomID:      roomID,
		UserID:      user.ID,
		Message:     messageText,
		MessageType: "chat",
	}

	// DBに保存
	err = h.repo.RoomMessage.CreateMessage(message)
	if err != nil {
		http.Error(w, "メッセージの送信に失敗しました", http.StatusInternalServerError)
		return
	}

	// ユーザー情報を設定
	message.User = *user

	// SSEでブロードキャスト
	event := sse.Event{
		ID:   message.ID.String(),
		Type: "message",
		Data: message,
	}
	h.hub.BroadcastToRoom(roomID, event)

	// htmx用のHTMLレスポンス（自分の画面には即座に反映）
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("")) // htmxはフォームリセットのみ行う
}

// StreamMessages はSSEでメッセージをストリーミング
func (h *RoomMessageHandler) StreamMessages(w http.ResponseWriter, r *http.Request) {
	// URLパラメータから部屋IDを取得
	roomIDStr := chi.URLParam(r, "id")

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "無効な部屋IDです", http.StatusBadRequest)
		return
	}

	// SSE一時トークンによる認証
	sseToken := r.URL.Query().Get("token")
	if sseToken == "" {
		http.Error(w, "SSEトークンが必要です", http.StatusUnauthorized)
		return
	}

	// 一時トークンを検証・消費
	tokenData, valid := globalSSETokenManager.ConsumeToken(sseToken)
	if !valid {
		http.Error(w, "無効または期限切れのSSEトークンです", http.StatusUnauthorized)
		return
	}

	// トークンの部屋IDと一致することを確認
	if tokenData.RoomID != roomID {
		http.Error(w, "トークンの部屋IDが一致しません", http.StatusForbidden)
		return
	}

	// DBからユーザー情報を取得
	user, err := h.repo.User.FindUserByID(tokenData.UserID)
	if err != nil || user == nil {
		http.Error(w, "ユーザーが見つかりません", http.StatusUnauthorized)
		return
	}

	// SSEヘッダーの設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Nginxのバッファリングを無効化

	// chunked encoding関連の最適化
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// クライアントの作成
	client := &sse.Client{
		ID:     uuid.New(),
		UserID: user.ID,
		RoomID: roomID,
		Send:   make(chan sse.Event, 10),
	}

	// Hubに登録
	h.hub.Register(client)
	defer func() {
		h.hub.Unregister(client)
	}()

	// 接続確認用のping（短い間隔で接続を維持）
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// フラッシャーの取得
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "ストリーミングがサポートされていません", http.StatusInternalServerError)
		return
	}

	// クライアントの切断を検出
	notify := r.Context().Done()

	// 初期接続確認メッセージを送信
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	for {
		select {
		case event := <-client.Send:
			// イベントを送信
			data, err := sse.SerializeEvent(event)
			if err != nil {
				continue
			}
			fmt.Fprint(w, data)
			flusher.Flush()

		case <-ticker.C:
			// キープアライブ
			fmt.Fprintf(w, ":ping\n\n")
			flusher.Flush()

		case <-notify:
			// クライアントが切断
			return
		}
	}
}

// GetMessages はメッセージ履歴を取得
func (h *RoomMessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// URLパラメータから部屋IDを取得
	roomIDStr := chi.URLParam(r, "id")

	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		http.Error(w, "無効な部屋IDです", http.StatusBadRequest)
		return
	}

	// 認証チェック（DBユーザー情報を取得）
	user, ok := middleware.GetDBUserFromContext(r.Context())
	if !ok || user == nil {
		http.Error(w, "認証が必要です", http.StatusUnauthorized)
		return
	}

	// 部屋のメンバーチェック
	if !h.repo.Room.IsUserJoinedRoom(roomID, user.ID) {
		http.Error(w, "部屋のメンバーではありません", http.StatusForbidden)
		return
	}

	// クエリパラメータの取得
	beforeIDStr := r.URL.Query().Get("before")
	limitStr := r.URL.Query().Get("limit")

	var beforeID *uuid.UUID
	if beforeIDStr != "" {
		id, err := uuid.Parse(beforeIDStr)
		if err == nil {
			beforeID = &id
		}
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// メッセージを取得
	messages, err := h.repo.RoomMessage.GetMessages(roomID, limit, beforeID)
	if err != nil {
		http.Error(w, "メッセージの取得に失敗しました", http.StatusInternalServerError)
		return
	}

	// JSON形式で返却
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
