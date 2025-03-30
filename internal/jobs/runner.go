package job

import (
	"context"
	"fmt"
	"reflect"
)

type Job interface {
	Execute(ctx context.Context) error
}

type Runner struct {
	jobs map[string]Job
}

func NewRunner(jobs []Job) *Runner {
	jobMap := make(map[string]Job)
	for _, job := range jobs {
		jobName := reflect.TypeOf(job).Elem().Name()
		jobMap[jobName] = job
	}
	return &Runner{jobs: jobMap}
}

// GetJob 取得任務
func (f *Runner) GetJob(jobName string) (Job, error) {
	job, exists := f.jobs[jobName]
	if !exists {
		return nil, fmt.Errorf("unknown job '%s'", jobName)
	}
	return job, nil
}
