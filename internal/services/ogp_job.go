package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/google/uuid"
)

type OGPJobService struct {
	projectID string
	location  string
	jobName   string
	ogBucket  string
	ogPrefix  string
	mode      string // "cloud", "local", "skip"
}

func NewOGPJobService() *OGPJobService {
	mode := os.Getenv("OGP_GENERATION_MODE")
	if mode == "" {
		if os.Getenv("PROJECT_ID") != "" || getGCPProjectID() != "" {
			mode = "cloud"
		} else {
			mode = "local"
		}
	}

	// PROJECT_IDとLOCATIONはメタデータサーバーから取得（環境変数があれば優先）
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		projectID = getGCPProjectID()
	}

	location := os.Getenv("LOCATION")
	if location == "" {
		location = getGCPRegion()
	}

	return &OGPJobService{
		projectID: projectID,
		location:  location,
		jobName:   os.Getenv("OGP_JOB_NAME"),
		ogBucket:  os.Getenv("OG_BUCKET"),
		ogPrefix:  os.Getenv("OG_PREFIX"),
		mode:      mode,
	}
}

func (s *OGPJobService) TriggerOGPGeneration(ctx context.Context, roomID uuid.UUID) error {
	switch s.mode {
	case "cloud":
		return s.triggerCloudRunJob(ctx, roomID)
	case "local":
		return s.generateOGPLocally(ctx, roomID)
	case "skip":
		log.Printf("OGP生成をスキップ: room_id=%s", roomID)
		return nil
	default:
		log.Printf("不明なOGP_GENERATION_MODE: %s (スキップします)", s.mode)
		return nil
	}
}

// triggerCloudRunJob Cloud Run Jobsを実行
func (s *OGPJobService) triggerCloudRunJob(ctx context.Context, roomID uuid.UUID) error {
	if s.projectID == "" || s.location == "" || s.jobName == "" {
		return fmt.Errorf("Cloud Run Jobs実行に必要な環境変数が未設定: PROJECT_ID=%s, LOCATION=%s, JOB_NAME=%s",
			s.projectID, s.location, s.jobName)
	}

	client, err := run.NewJobsClient(ctx)
	if err != nil {
		return fmt.Errorf("Cloud Run Jobs クライアント作成失敗: %w", err)
	}
	defer client.Close()

	jobFullName := fmt.Sprintf("projects/%s/locations/%s/jobs/%s", s.projectID, s.location, s.jobName)

	req := &runpb.RunJobRequest{
		Name: jobFullName,
		Overrides: &runpb.RunJobRequest_Overrides{
			ContainerOverrides: []*runpb.RunJobRequest_Overrides_ContainerOverride{
				{
					Env: []*runpb.EnvVar{
						{
							Name:   "ROOM_ID",
							Values: &runpb.EnvVar_Value{Value: roomID.String()},
						},
						{
							Name:   "OG_BUCKET",
							Values: &runpb.EnvVar_Value{Value: s.ogBucket},
						},
						{
							Name:   "OG_PREFIX",
							Values: &runpb.EnvVar_Value{Value: s.ogPrefix},
						},
					},
				},
			},
		},
	}

	// ジョブ実行（完了を待たずに戻る）
	op, err := client.RunJob(ctx, req)
	if err != nil {
		return fmt.Errorf("Cloud Run Jobs 実行失敗: %w", err)
	}

	log.Printf("OGP生成ジョブ実行開始: room_id=%s, operation=%s", roomID, op.Name())

	return nil
}

// generateOGPLocally ローカルでOGP画像を生成
func (s *OGPJobService) generateOGPLocally(ctx context.Context, roomID uuid.UUID) error {
	log.Printf("ローカルモードでOGP画像を生成: room_id=%s", roomID)

	ogPrefix := s.ogPrefix
	if ogPrefix == "" {
		ogPrefix = "dev"
	}

	cmd := exec.CommandContext(ctx, "go", "run", "cmd/ogp-renderer/main.go")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("ROOM_ID=%s", roomID),
		fmt.Sprintf("OG_PREFIX=%s", ogPrefix),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("OGP生成失敗: %w, output: %s", err, output)
	}

	log.Printf("OGP生成完了: %s", output)
	return nil
}

// getGCPProjectID GCPメタデータサーバーからプロジェクトIDを取得
func getGCPProjectID() string {
	return getGCPMetadata("project/project-id")
}

// getGCPRegion GCPメタデータサーバーからリージョンを取得
func getGCPRegion() string {
	region := getGCPMetadata("instance/region")
	if region == "" {
		return ""
	}
	// "projects/123456/regions/asia-northeast1" → "asia-northeast1"
	parts := strings.Split(region, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return region
}

// getGCPMetadata GCPメタデータサーバーから情報を取得
func getGCPMetadata(path string) string {
	url := fmt.Sprintf("http://metadata.google.internal/computeMetadata/v1/%s", path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Metadata-Flavor", "Google")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(body))
}
