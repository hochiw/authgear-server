//+build wireinject

package server

import (
	"github.com/google/wire"

	configsource "github.com/authgear/authgear-server/pkg/lib/config/source"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/queue"
)

func newConfigSourceController(p *deps.RootProvider) *configsource.Controller {
	panic(wire.Build(deps.RootDependencySet))
}

func newInProcessQueue(p *deps.AppProvider, e *executor.InProcessExecutor) *queue.InProcessQueue {
	panic(wire.Build(
		deps.RootDependencySet,
		wire.FieldsOf(new(*deps.AppProvider),
			"Config",
			"Database",
		),
		queue.DependencySet,
		wire.Bind(new(queue.Executor), new(*executor.InProcessExecutor)),
	))
}
