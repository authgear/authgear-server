package userimport

import "context"

type Store interface {
	CreateJob(ctx context.Context, job *Job) error
	GetJob(ctx context.Context, jobID string) (*Job, error)
}
