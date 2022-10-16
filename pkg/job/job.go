package job

import (
	"fmt"
	"sync"
)

type JobResult[Out any] struct {
	Result Out
	Err    error
}

func Dispatch[Job, Out any](jobs []Job, f func(Job) (Out, error)) <-chan JobResult[Out] {
	var wg sync.WaitGroup
	wg.Add(len(jobs))

	// Dispatch everything
	results := make(chan JobResult[Out])
	for _, job := range jobs {
		go func(job Job) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					results <- JobResult[Out]{
						Err: fmt.Errorf("panic in job: %v", r),
					}
				}
			}()

			res, err := f(job)
			results <- JobResult[Out]{
				Result: res,
				Err:    err,
			}
		}(job)
	}

	// Close when all jobs are done
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
