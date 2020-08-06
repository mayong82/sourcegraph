package worker

import (
	"context"
	"sync/atomic"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// TODO - can we collapse the handler and processor?
type handler struct {
	store           store.Store
	processor       *processor
	enableBudget    bool
	budgetRemaining int64
}

var _ dbworker.Handler = &handler{}
var _ workerutil.WithPreDequeue = &handler{}
var _ workerutil.WithHooks = &handler{}

func (h *handler) Handle(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
	_, err := h.processor.Process(ctx, h.store.With(tx), record.(store.Upload))
	return err
}

func (h *handler) PreDequeue(ctx context.Context) (bool, interface{}, error) {
	if !h.enableBudget {
		return true, nil, nil
	}

	budgetRemaining := atomic.LoadInt64(&h.budgetRemaining)
	if budgetRemaining <= 0 {
		return false, nil, nil
	}

	return true, []*sqlf.Query{sqlf.Sprintf("(upload_size IS NULL OR upload_size <= %s)", budgetRemaining)}, nil
}

func (h *handler) PreHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, -h.getSize(record))
}

func (h *handler) PostHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, +h.getSize(record))
}

func (h *handler) getSize(record workerutil.Record) int64 {
	if size := record.(store.Upload).UploadSize; size != nil {
		return *size
	}

	return 0
}
