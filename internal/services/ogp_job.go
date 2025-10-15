package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

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
		if os.Getenv("PROJECT_ID") != "" {
			mode = "cloud"
		} else {
			mode = "local"
		}
	}

	return &OGPJobService{
		projectID: os.Getenv("PROJECT_ID"),
		location:  os.Getenv("LOCATION"),
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
