package worker

import (
	"context"
	"errors"
	"sync"
)

var ErrTooManyTasks = errors.New("cannot run tasks simultaneously")

type WorkerTask func(context.Context) error

type Worker struct {
	task      WorkerTask
	handleErr func(error)

	wg     *sync.WaitGroup
	cancel *context.CancelFunc
	mux    *sync.Mutex
	done   chan struct{}
}

func New(task WorkerTask, errorHandler func(error)) *Worker {
	return &Worker{
		task:      task,
		handleErr: errorHandler,
		mux:       &sync.Mutex{},
		wg:        &sync.WaitGroup{},
		done:      make(chan struct{}),
	}
}

// Start starts the worker and returns a channel that indicates that
// the worker has finished any running task
func (w *Worker) Start(ctx context.Context) chan struct{} {
	go w.manageLifecycle(ctx)
	return w.done
}

func (w *Worker) manageLifecycle(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.cancelTask()
			w.wg.Wait()
			close(w.done)
			return
		default:
			continue
		}
	}
}

func (w *Worker) cancelTask() {
	w.mux.Lock()
	defer w.mux.Unlock()
	if w.cancel != nil {
		doCancel := *w.cancel
		doCancel()
	}
	w.cancel = nil
}

// TriggerTask starts a single asynchronous task that will be
// gracefully shutdown via the context sent to Start()
func (w *Worker) TriggerTask() error {
	w.mux.Lock()
	defer w.mux.Unlock()
	if w.cancel != nil {
		return ErrTooManyTasks
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = &cancel

	go func() {
		w.wg.Add(1)
		err := w.task(ctx)
		if err != nil {
			w.handleErr(err)
		}
		w.wg.Done()
	}()
	return nil
}
