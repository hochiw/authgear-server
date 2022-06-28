// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package usage

import (
	"context"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/lib/meter"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/cobrasentry"
	"github.com/getsentry/sentry-go"
)

// Injectors from wire.go:

func NewCountCollector(ctx context.Context, pool *db.Pool, databaseCredentials *config.DatabaseCredentials, auditDatabaseCredentials *config.AuditDatabaseCredentials, redisPool *redis.Pool, credentials *config.AnalyticRedisCredentials, hub *sentry.Hub) *usage.CountCollector {
	globalDatabaseCredentialsEnvironmentConfig := NewGlobalDatabaseCredentials(databaseCredentials)
	databaseEnvironmentConfig := config.NewDefaultDatabaseEnvironmentConfig()
	factory := cobrasentry.NewLoggerFactory(hub)
	handle := globaldb.NewHandle(ctx, pool, globalDatabaseCredentialsEnvironmentConfig, databaseEnvironmentConfig, factory)
	sqlBuilder := globaldb.NewSQLBuilder(globalDatabaseCredentialsEnvironmentConfig)
	sqlExecutor := globaldb.NewSQLExecutor(ctx, handle)
	globalDBStore := &usage.GlobalDBStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	redisEnvironmentConfig := config.NewDefaultRedisEnvironmentConfig()
	analyticredisHandle := analyticredis.NewHandle(redisPool, redisEnvironmentConfig, credentials, factory)
	readStoreRedis := &meter.ReadStoreRedis{
		Context: ctx,
		Redis:   analyticredisHandle,
	}
	readHandle := auditdb.NewReadHandle(ctx, pool, databaseEnvironmentConfig, auditDatabaseCredentials, factory)
	auditdbSQLBuilder := auditdb.NewSQLBuilder(auditDatabaseCredentials)
	readSQLExecutor := auditdb.NewReadSQLExecutor(ctx, readHandle)
	auditDBReadStore := &meter.AuditDBReadStore{
		SQLBuilder:  auditdbSQLBuilder,
		SQLExecutor: readSQLExecutor,
	}
	countCollector := &usage.CountCollector{
		GlobalHandle:  handle,
		GlobalDBStore: globalDBStore,
		ReadCounter:   readStoreRedis,
		AuditHandle:   readHandle,
		Meters:        auditDBReadStore,
	}
	return countCollector
}
