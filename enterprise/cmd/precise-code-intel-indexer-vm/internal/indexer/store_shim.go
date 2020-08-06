package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type apiStore struct {
	queueClient client.Client
}

var _ workerutil.Store = &apiStore{}

func (s *apiStore) Dequeue(ctx context.Context, extraArguments interface{}) (workerutil.Record, workerutil.Store, bool, error) {
	index, dequeued, err := s.queueClient.Dequeue(ctx)
	return index, s, dequeued, err
}

func (s *apiStore) MarkComplete(ctx context.Context, id int) (bool, error) {
	return true, s.queueClient.Complete(ctx, id, nil)
}

func (s *apiStore) MarkErrored(ctx context.Context, id int, failureMessage string) (bool, error) {
	return true, s.queueClient.Complete(ctx, id, errors.New(failureMessage))
}

func (s *apiStore) Done(err error) error {
	return nil
}
