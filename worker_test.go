package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	worker "github.com/laouji/async-worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type WorkerSuite struct {
	suite.Suite
}

func TestWorkerSuite(t *testing.T) {
	suite.Run(t, new(WorkerSuite))
}

func (s *WorkerSuite) TestTriggerTask() {
	expectedErr := errors.New("expected")
	task := func(ctx context.Context) error {
		return expectedErr
	}
	errorHandler := func(err error) {
		require.Error(s.T(), err)
		assert.Equal(s.T(), err, expectedErr)
	}

	w := worker.New(task, errorHandler)

	ctx, cancel := context.WithCancel(context.Background())
	done := w.Start(ctx)
	err := w.TriggerTask()
	require.NoError(s.T(), err)

	time.Sleep(30 * time.Millisecond)
	cancel()
	<-done
}

func (s *WorkerSuite) TestCannotTriggerSimultaneousTasks() {
	task := func(ctx context.Context) error {
		time.Sleep(30 * time.Millisecond)
		return nil
	}
	w := worker.New(task, func(err error) {})

	ctx, cancel := context.WithCancel(context.Background())
	done := w.Start(ctx)
	err := w.TriggerTask()
	require.NoError(s.T(), err)
	err = w.TriggerTask()
	require.Error(s.T(), err)
	assert.Equal(s.T(), err, worker.ErrTooManyTasks)

	cancel()
	<-done
}
