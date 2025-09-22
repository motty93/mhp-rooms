package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"mhp-rooms/internal/middleware"
	"mhp-rooms/internal/models"
	"mhp-rooms/internal/repository"
	"mhp-rooms/internal/storage"
)

// ReportHandler 通報関連のハンドラー
type ReportHandler struct {
	reportRepo  repository.ReportRepositoryInterface
	userRepo    repository.UserRepository
	gcsUploader *storage.GCSUploader
}

// NewReportHandler ハンドラーのコンストラクタ
func NewReportHandler(reportRepo repository.ReportRepositoryInterface, userRepo repository.UserRepository, gcsUploader *storage.GCSUploader) *ReportHandler {
	return &ReportHandler{
		reportRepo:  reportRepo,
		userRepo:    userRepo,
		gcsUploader: gcsUploader,
	}
}

// ReportRequest 通報リクエストの構造体
type ReportRequest struct {
	Reason      models.ReportReason `json:"reason"`
	Description string              `json:"description"`
}

// CreateReport ユーザーを通報する
func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	// URLパラメータからユーザーIDを取得
	reportedUserID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "無効なユーザーIDです"})
		return
	}

	// セッションから通報者のユーザーIDを取得
	reporterUserID := getUserIDFromSession(r)
	fmt.Printf("通報者のユーザーID: %s\n", reporterUserID)
	if reporterUserID == uuid.Nil {
		renderJSON(w, http.StatusUnauthorized, map[string]string{"error": "ログインが必要です"})
		return
	}

	// デバッグ: 存在しないユーザーIDの場合、既存のユーザーIDを使用
	// TODO: 本番環境では削除すること
	if os.Getenv("ENV") == "development" {
		// ユーザーが存在するか確認
		if _, err := h.userRepo.FindUserByID(reporterUserID); err != nil {
			fmt.Printf("通報者ユーザーが存在しません。デフォルトユーザーを使用: %s -> d4d3e6ec-128a-44e6-be58-d59f0a8d993d\n", reporterUserID)
			reporterUserID = uuid.MustParse("d4d3e6ec-128a-44e6-be58-d59f0a8d993d") // プロハンター
		}
	}

	// 自分自身を通報できないようにする
	if reporterUserID == reportedUserID {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "自分自身を通報することはできません"})
		return
	}

	// リクエストボディをパース
	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "リクエストの形式が正しくありません"})
		return
	}

	// デバッグ用ログ
	fmt.Printf("受信したリクエスト: %+v\n", req)
	fmt.Printf("受信した理由: %+v\n", req.Reason)

	// バリデーション
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

	// 通報を作成
	fmt.Printf("データベース用理由: %+v\n", req.Reason)

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

// UploadAttachment 通報に画像を添付する
func (h *ReportHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	// URLパラメータから通報IDを取得
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

	// デバッグ: 存在しないユーザーIDの場合、既存のユーザーIDを使用
	// TODO: 本番環境では削除すること
	if os.Getenv("ENV") == "development" {
		// ユーザーが存在するか確認
		if _, err := h.userRepo.FindUserByID(userID); err != nil {
			fmt.Printf("アップロードユーザーが存在しません。デフォルトユーザーを使用: %s -> d4d3e6ec-128a-44e6-be58-d59f0a8d993d\n", userID)
			userID = uuid.MustParse("d4d3e6ec-128a-44e6-be58-d59f0a8d993d") // プロハンター
		}
	}

	// 通報が存在し、通報者が本人か確認
	report, err := h.reportRepo.GetByID(reportID)
	if err != nil || report == nil {
		renderJSON(w, http.StatusNotFound, map[string]string{"error": "通報が見つかりません"})
		return
	}

	// デバッグログ
	fmt.Printf("アップロード権限チェック: report.ReporterUserID=%s, userID=%s\n", report.ReporterUserID, userID)

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

	// ファイルを取得
	file, header, err := r.FormFile("image")
	if err != nil {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "画像ファイルが見つかりません"})
		return
	}
	defer file.Close()

	// ファイルタイプを確認
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		renderJSON(w, http.StatusBadRequest, map[string]string{"error": "対応していない画像形式です（jpg, jpeg, png, gif のみ）"})
		return
	}

	// ファイル名を生成
	ext := getFileExtension(header.Filename)
	filename := fmt.Sprintf("%s_%d_%s%s",
		reportID.String(),
		time.Now().Unix(),
		generateRandomString(8),
		ext,
	)

	// 保存先ディレクトリを作成
	uploadDir := filepath.Join("static", "uploads", "reports")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "ファイルの保存に失敗しました"})
		return
	}

	// ファイルを保存
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "ファイルの保存に失敗しました"})
		return
	}
	defer dst.Close()

	fileSize, err := io.Copy(dst, file)
	if err != nil {
		renderJSON(w, http.StatusInternalServerError, map[string]string{"error": "ファイルの保存に失敗しました"})
		return
	}

	// データベースに保存
	attachment := &models.ReportAttachment{
		ReportID:     reportID,
		FilePath:     "/" + strings.ReplaceAll(filePath, "\\", "/"), // URLパスに変換
		FileType:     contentType,
		FileSize:     fileSize,
		OriginalName: header.Filename,
	}

	if err := h.reportRepo.AddAttachment(attachment); err != nil {
		// ファイルを削除
		os.Remove(filePath)
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

// GetReportReasons 通報理由の一覧を取得
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

// 補助関数

// getUserIDFromSession セッションからユーザーIDを取得（実際の実装に合わせて調整必要）
func getUserIDFromSession(r *http.Request) uuid.UUID {
	// middleware.UserContextKeyを使用してユーザー情報を取得
	if user, ok := middleware.GetUserFromContext(r.Context()); ok && user != nil {
		if id, err := uuid.Parse(user.ID); err == nil {
			return id
		}
	}

	// middleware.DBUserContextKeyも試す
	if dbUser, ok := middleware.GetDBUserFromContext(r.Context()); ok && dbUser != nil {
		return dbUser.ID
	}

	return uuid.Nil
}

// isValidImageType 有効な画像タイプか確認
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

// getFileExtension ファイル拡張子を取得
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

// renderJSON JSONレスポンスを返す
func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
