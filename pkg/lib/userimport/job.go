package userimport

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const recordsPerTask = 100
const jobIDPrefix = "userimport_"
const taskIDPrefix = "userimporttask_"

func newJobID() string {
	return jobIDPrefix + rand.StringWithAlphabet(32, base32.Alphabet, rand.SecureRand)
}

type TaskProducer interface {
	NewTask(appID string, input json.RawMessage, taskIDPrefix string) *redisqueue.Task
	EnqueueTask(ctx context.Context, task *redisqueue.Task) error
	GetTask(ctx context.Context, item *redisqueue.QueueItem) (*redisqueue.Task, error)
}

const (
	usageLimitUserImport usage.LimitName = "UserImport"
)

type UsageLimiter interface {
	ReserveN(ctx context.Context, name usage.LimitName, n int, config *config.UsageLimitConfig) (*usage.Reservation, error)
}

type JobManager struct {
	AppID config.AppID
	Clock clock.Clock

	AdminAPIFeatureConfig *config.AdminAPIFeatureConfig
	TaskProducer          TaskProducer
	UsageLimiter          UsageLimiter
	Store                 Store
}

func (m *JobManager) EnqueueJob(ctx context.Context, request *Request) (*Response, error) {
	_, err := m.UsageLimiter.ReserveN(
		ctx,
		usageLimitUserImport,
		len(request.Records),
		m.AdminAPIFeatureConfig.UserImportUsage,
	)
	if err != nil {
		return nil, err
	}

	var taskIDs []string
	records := request.Records
	for len(records) > 0 {
		taskRecords := records[:min(recordsPerTask, len(records))]
		records = records[len(taskRecords):]

		taskRequest := *request
		taskRequest.Records = taskRecords

		rawRequest, err := json.Marshal(&taskRequest)
		if err != nil {
			return nil, err
		}

		task := m.TaskProducer.NewTask(string(m.AppID), rawRequest, taskIDPrefix)
		err = m.TaskProducer.EnqueueTask(ctx, task)
		if err != nil {
			return nil, err
		}

		taskIDs = append(taskIDs, task.ID)
	}

	job := &Job{
		ID:        newJobID(),
		CreatedAt: m.Clock.NowUTC(),
		TaskIDs:   taskIDs,
	}
	err = m.Store.CreateJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return m.GetJob(ctx, job.ID)
}

func (m *JobManager) GetJob(ctx context.Context, jobID string) (*Response, error) {
	if !strings.HasPrefix(jobID, jobIDPrefix) {
		// Backward compatibility.
		queueItem := &redisqueue.QueueItem{
			AppID:  string(m.AppID),
			TaskID: jobID,
		}

		task, err := m.TaskProducer.GetTask(ctx, queueItem)
		if err != nil {
			return nil, err
		}

		return NewResponseFromTask(task)
	}

	job, err := m.Store.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	resp := NewResponseFromJob(job)
	for idx, taskID := range job.TaskIDs {
		queueItem := &redisqueue.QueueItem{
			AppID:  string(m.AppID),
			TaskID: taskID,
		}

		task, err := m.TaskProducer.GetTask(ctx, queueItem)
		if err != nil {
			return nil, err
		}
		if err = resp.AggregateTaskResult(idx*recordsPerTask, task); err != nil {
			return nil, err
		}
	}
	return resp, nil
}
