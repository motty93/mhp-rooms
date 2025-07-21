package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
)

// ReactionHandler はリアクション関連のハンドラー
type ReactionHandler struct {
	BaseHandler
}

// NewReactionHandler は新しいReactionHandlerを作成する
func NewReactionHandler(repo *repository.Repository) *ReactionHandler {
	return &ReactionHandler{
		BaseHandler: BaseHandler{repo: repo},
	}
}

// AddReaction はメッセージにリアクションを追加する
func (h *ReactionHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	// ユーザー認証チェック
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// URLパラメータからメッセージIDを取得
	vars := mux.Vars(r)
	messageIDStr := vars["messageId"]
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// リクエストボディからリアクションタイプを取得
	var req struct {
		ReactionType string `json:"reaction_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// メッセージの存在確認
	if err := h.repo.Reaction.CheckMessageExists(messageID); err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	// リアクションタイプの有効性確認
	if err := h.repo.Reaction.CheckReactionTypeExists(req.ReactionType); err != nil {
		http.Error(w, "Invalid reaction type", http.StatusBadRequest)
		return
	}

	// リアクションを追加
	reaction := &models.MessageReaction{
		MessageID:    messageID,
		UserID:       userID,
		ReactionType: req.ReactionType,
	}

	if err := h.repo.Reaction.AddReaction(reaction); err != nil {
		// 既にリアクションが存在する場合
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"unique_user_message_reaction\" (SQLSTATE 23505)" {
			http.Error(w, "Reaction already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	// 成功レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"reaction": reaction,
	})
}

// RemoveReaction はメッセージからリアクションを削除する
func (h *ReactionHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	// ユーザー認証チェック
	userID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// URLパラメータから情報を取得
	vars := mux.Vars(r)
	messageIDStr := vars["messageId"]
	reactionType := vars["reactionType"]

	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// リアクションを削除
	if err := h.repo.Reaction.RemoveReaction(messageID, userID, reactionType); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "Reaction not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	// 成功レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Reaction removed successfully",
	})
}

// GetMessageReactions はメッセージのリアクション一覧を取得する
func (h *ReactionHandler) GetMessageReactions(w http.ResponseWriter, r *http.Request) {
	// URLパラメータからメッセージIDを取得
	vars := mux.Vars(r)
	messageIDStr := vars["messageId"]
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// 現在のユーザーID（ログインしていない場合もある）
	userID, isAuthenticated := r.Context().Value("userID").(uuid.UUID)

	// リアクションの集計データを取得
	var userIDPtr *uuid.UUID
	if isAuthenticated {
		userIDPtr = &userID
	}

	reactionCounts, err := h.repo.Reaction.GetMessageReactions(messageID, userIDPtr)
	if err != nil {
		http.Error(w, "Failed to fetch reactions", http.StatusInternalServerError)
		return
	}

	// 成功レスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"reactions": reactionCounts,
	})
}

// GetAvailableReactions は利用可能なリアクションタイプを取得する
func (h *ReactionHandler) GetAvailableReactions(w http.ResponseWriter, r *http.Request) {
	reactionTypes, err := h.repo.Reaction.GetReactionTypes()
	if err != nil {
		http.Error(w, "Failed to fetch reaction types", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"types":   reactionTypes,
	})
}