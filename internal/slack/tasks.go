package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TasksClient struct {
	client *cloudtasks.Client
	config *Config
}

func NewTasksClient(ctx context.Context, config *Config) (*TasksClient, error) {
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Tasks client: %w", err)
	}

	return &TasksClient{
		client: client,
		config: config,
	}, nil
}

func (tc *TasksClient) Close() error {
	return tc.client.Close()
}

func (tc *TasksClient) EnqueueAnalyzeTask(ctx context.Context, req *AnalyzeRequest) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		tc.config.GoogleCloudProject, tc.config.TasksLocation, tc.config.TasksQueueName)

	task := &cloudtaskspb.Task{
		MessageType: &cloudtaskspb.Task_HttpRequest{
			HttpRequest: &cloudtaskspb.HttpRequest{
				HttpMethod: cloudtaskspb.HttpMethod_POST,
				Url:        tc.config.WorkerEndpoint,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: payload,
			},
		},
		ScheduleTime: timestamppb.Now(),
	}

	createTaskReq := &cloudtaskspb.CreateTaskRequest{
		Parent: queuePath,
		Task:   task,
	}

	_, err = tc.client.CreateTask(ctx, createTaskReq)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	log.Printf("Task enqueued for trace_id: %s", req.TraceID)
	return nil
}
