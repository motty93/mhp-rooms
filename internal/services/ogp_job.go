package services

import (
	"context"
	"fmt"
	"log"
	"os"

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
}

func NewOGPJobService() *OGPJobService {
	return &OGPJobService{
		projectID: os.Getenv("PROJECT_ID"),
		location:  os.Getenv("LOCATION"),
		jobName:   os.Getenv("OGP_JOB_NAME"),
		ogBucket:  os.Getenv("OG_BUCKET"),
		ogPrefix:  os.Getenv("OG_PREFIX"),
	}
}

func (s *OGPJobService) TriggerOGPGeneration(ctx context.Context, roomID uuid.UUID) error {
	if s.projectID == "" || s.location == "" || s.jobName == "" {
		log.Printf("OGP生成ジョブ実行をスキップ: 環境変数未設定 (PROJECT_ID=%s, LOCATION=%s, JOB_NAME=%s)",
			s.projectID, s.location, s.jobName)
		return nil
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
