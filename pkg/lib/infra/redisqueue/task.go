package redisqueue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
)

type Task struct {
	ID          string              `json:"id,omitempty"`
	AppID       string              `json:"app_id,omitempty"`
	CreatedAt   *time.Time          `json:"created_at,omitempty"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	Status      TaskStatus          `json:"status,omitempty"`
	Input       json.RawMessage     `json:"input,omitempty"`
	Output      json.RawMessage     `json:"output,omitempty"`
	Error       *apierrors.APIError `json:"error,omitempty"`
}

func (t *Task) RedisKey() string {
	return RedisKeyForTask(t.AppID, t.ID)
}

func (t *Task) ToQueueItem() *QueueItem {
	return &QueueItem{
		AppID:  t.AppID,
		TaskID: t.ID,
	}
}

type QueueItem struct {
	AppID  string `json:"app_id,omitempty"`
	TaskID string `json:"task_id,omitempty"`
}

func (i *QueueItem) RedisKey() string {
	return RedisKeyForTask(i.AppID, i.TaskID)
}

func RedisKeyForTask(appID string, taskID string) string {
	return fmt.Sprintf("app:%v:redis-queue-task:%v", appID, taskID)
}

func RedisKeyForQueue(queueName string) string {
	return fmt.Sprintf("redis-queue:%v", queueName)
}
