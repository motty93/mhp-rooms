package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReportHandler struct {
	reportRepo  repository.ReportRepository
	userRepo    repository.UserRepository
	gcsUploader *storage.GCSUploader
}

func NewReportHandler(reportRepo repository.ReportRepository, userRepo repository.UserRepository, gcsUploader *storage.GCSUploader) *ReportHandler {
	return &ReportHandler{
		reportRepo:  reportRepo,
		userRepo:    userRepo,
		gcsUploader: gcsUploader,
	}
}

type ReportRequest struct {
	Reason      models.ReportReason `json:"reason"`
	Description string              `json:"description"`
}

func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	reportedUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "無効なユーザーIDです"})
		return
	}

	reporterUserID := getUserIDFromSession(r)
	fmt.Printf("通報者のユーザーID: %s\n", reporterUserID)
	if reporterUserID == uuid.Nil {
		renderJSON(w, http.StatusUnauthorized, map[string]string{"error": "ログインが必要です"})
		return
	}

	// 自分自身を通報できないようにする
	if reporterUserID == reportedUserID {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "自分自身を通報することはできません"})
		return
	}

	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "リクエストの形式が正しくありません"})
		return
	}

	if req.Reason == "" {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "通報理由を選択してください"})
		return
	}

	if len(req.Description) == 0 {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "詳細な説明を入力してください"})
		return
	}

	if len(req.Description) > 500 {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "説明は500文字以内で入力してください"})
		return
	}

	// 24時間以内の重複通報をチェック
	isDuplicate, err := h.reportRepo.CheckDuplicateReport(reporterUserID, reportedUserID)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "通報処理中にエラーが発生しました"})
		return
	}

	if isDuplicate {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "24時間以内に同じユーザーを既に通報しています"})
		return
	}

	// 通報対象ユーザーが存在するか確認
	reportedUser, err := h.userRepo.FindUserByID(reportedUserID)
	if err != nil || reportedUser == nil {

		renderJSON(w, http.StatusNotFound, map[string]string{"error": "通報対象のユーザーが見つかりません"})
		return
	}

	// 通報者ユーザーが存在するか確認
	reporterUser, err := h.userRepo.FindUserByID(reporterUserID)
	if err != nil || reporterUser == nil {

		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "通報者のユーザー情報が見つかりません"})
		return
	}

	report := &models.UserReport{
		ReporterUserID: reporterUserID,
		ReportedUserID: reportedUserID,
		Reason:         req.Reason,
		Description:    req.Description,
		Status:         models.ReportStatusPending,
	}

	if err := h.reportRepo.Create(report); err != nil {
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "通報の送信に失敗しました"})
		return
	}

	renderJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "通報を受け付けました。ご報告ありがとうございます。",
		"report_id": report.ID,
	})
}

// 通報に画像を添付する
func (h *ReportHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	reportID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "無効な通報IDです"})
		return
	}

	// セッションからユーザーIDを取得
	userID := getUserIDFromSession(r)
	if userID == uuid.Nil {
		renderJSON(w, http.StatusUnauthorized, map[string]string{"error": "ログインが必要です"})
		return
	}

	// 通報が存在し、通報者が本人か確認
	report, err := h.reportRepo.GetByID(reportID)
	if err != nil || report == nil {
		renderJSON(w, http.StatusNotFound, map[string]string{"error": "通報が見つかりません"})
		return
	}

	if report.ReporterUserID != userID {
		renderJSON(w, http.StatusForbidden, map[string]string{"error": "この通報にアクセスする権限がありません"})
		return
	}

	// 既存の添付ファイル数を確認（最大3枚）
	attachments, err := h.reportRepo.GetAttachmentsByReportID(reportID)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "添付ファイルの確認に失敗しました"})
		return
	}

	if len(attachments) >= 3 {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "添付ファイルは最大3枚までです"})
		return
	}

	// マルチパートフォームをパース（最大5MB）
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "ファイルサイズが大きすぎます（最大5MB）"})
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "画像ファイルが見つかりません"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "対応していない画像形式です（jpg, jpeg, png, gif のみ）"})
		return
	}

	// GCSのプライベートバケットにアップロード
	result, err := h.gcsUploader.UploadReportAttachment(r.Context(), reportID.String(), file, header)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "画像のアップロードに失敗しました"})
		return
	}

	attachment := &models.ReportAttachment{
		ReportID:     reportID,
		FilePath:     result.URL, // プライベートバケットのgs://URL
		FileType:     result.ContentType,
		FileSize:     header.Size, // ファイルサイズ
		OriginalName: header.Filename,
	}

	if err := h.reportRepo.AddAttachment(attachment); err != nil {
		// TODO: GCSからファイルを削除する処理を追加
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "添付ファイルの登録に失敗しました"})
		return
	}

	renderJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       "画像をアップロードしました",
		"attachment_id": attachment.ID,
		"file_path":     attachment.FilePath,
	})
}

func (h *ReportHandler) GetReportReasons(w http.ResponseWriter, r *http.Request) {
	reasons := []map[string]string{}
	labels := models.GetReasonLabels()
	descriptions := models.GetReasonDescription()

	for reason, label := range labels {
		reasons = append(reasons, map[string]string{
			"value":       string(reason),
			"label":       label,
			"description": descriptions[reason],
		})
	}

	renderJSON(w, http.StatusOK, reasons)
}

// セッションからユーザーIDを取得
func getUserIDFromSession(r *http.Request) uuid.UUID {
	// 最初にDBUserContextKeyを確認（ローカルDBのIDが必要）
	if dbUser, ok := middleware.GetDBUserFromContext(r.Context()); ok && dbUser != nil {
		return dbUser.ID
	}

	// DBUserが見つからない場合はNilを返す
	return uuid.Nil
}

// 有効な画像タイプか確認
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
	}
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".jpg" // デフォルト
	}
	return ext
}

// generateRandomString ランダム文字列を生成
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
