package worker

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

func NewWorker(
	s store.Store,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int64,
	metrics metrics.WorkerMetrics,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
		metrics:             metrics,
	}

	handler := &handler{
		store:           s,
		processor:       processor,
		enableBudget:    budgetMax > 0,
		budgetRemaining: budgetMax,
	}

	workerMetrics := workerutil.WorkerMetrics{
		HandleOperation: metrics.ProcessOperation,
	}

	options := dbworker.WorkerOptions{
		Handler:     handler,
		NumHandlers: numProcessorRoutines,
		Interval:    pollInterval,
		Metrics:     workerMetrics,
	}

	return dbworker.NewWorker(rootContext, store.WorkerutilUploadStore(s), options)
}
