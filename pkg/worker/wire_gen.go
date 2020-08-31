// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package worker

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/password"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/task/executor"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/worker/tasks"
)

// Injectors from wire.go:

func newInProcessExecutor(p *deps.RootProvider) *executor.InProcessExecutor {
	factory := p.LoggerFactory
	inProcessExecutorLogger := executor.NewInProcessExecutorLogger(factory)
	restoreTaskContext := deps.ProvideRestoreTaskContext(p)
	inProcessExecutor := &executor.InProcessExecutor{
		Logger:         inProcessExecutorLogger,
		RestoreContext: restoreTaskContext,
	}
	return inProcessExecutor
}

func newPwHousekeeperTask(p *deps.TaskProvider) task.Task {
	appProvider := p.AppProvider
	handle := appProvider.Database
	factory := appProvider.LoggerFactory
	pwHousekeeperLogger := tasks.NewPwHousekeeperLogger(factory)
	clock := _wireSystemClockValue
	config := appProvider.Config
	secretConfig := config.SecretConfig
	databaseCredentials := deps.ProvideDatabaseCredentials(secretConfig)
	appConfig := config.AppConfig
	appID := appConfig.ID
	sqlBuilder := db.ProvideSQLBuilder(databaseCredentials, appID)
	context := p.Context
	sqlExecutor := db.SQLExecutor{
		Context:  context,
		Database: handle,
	}
	historyStore := &password.HistoryStore{
		Clock:       clock,
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	housekeeperLogger := password.NewHousekeeperLogger(factory)
	authenticatorConfig := appConfig.Authenticator
	authenticatorPasswordConfig := authenticatorConfig.Password
	housekeeper := &password.Housekeeper{
		Store:  historyStore,
		Logger: housekeeperLogger,
		Config: authenticatorPasswordConfig,
	}
	pwHousekeeperTask := &tasks.PwHousekeeperTask{
		Database:      handle,
		Logger:        pwHousekeeperLogger,
		PwHousekeeper: housekeeper,
	}
	return pwHousekeeperTask
}

var (
	_wireSystemClockValue = clock.NewSystemClock()
)

func newSendMessagesTask(p *deps.TaskProvider) task.Task {
	appProvider := p.AppProvider
	factory := appProvider.LoggerFactory
	logger := mail.NewLogger(factory)
	rootProvider := appProvider.RootProvider
	environmentConfig := rootProvider.EnvironmentConfig
	devMode := environmentConfig.DevMode
	config := appProvider.Config
	secretConfig := config.SecretConfig
	smtpServerCredentials := deps.ProvideSMTPServerCredentials(secretConfig)
	dialer := mail.NewGomailDialer(smtpServerCredentials)
	sender := &mail.Sender{
		Logger:       logger,
		DevMode:      devMode,
		GomailDialer: dialer,
	}
	smsLogger := sms.NewLogger(factory)
	appConfig := config.AppConfig
	messagingConfig := appConfig.Messaging
	twilioCredentials := deps.ProvideTwilioCredentials(secretConfig)
	twilioClient := sms.NewTwilioClient(twilioCredentials)
	nexmoCredentials := deps.ProvideNexmoCredentials(secretConfig)
	nexmoClient := sms.NewNexmoClient(nexmoCredentials)
	client := &sms.Client{
		Logger:          smsLogger,
		DevMode:         devMode,
		MessagingConfig: messagingConfig,
		TwilioClient:    twilioClient,
		NexmoClient:     nexmoClient,
	}
	sendMessagesLogger := tasks.NewSendMessagesLogger(factory)
	sendMessagesTask := &tasks.SendMessagesTask{
		EmailSender: sender,
		SMSClient:   client,
		Logger:      sendMessagesLogger,
	}
	return sendMessagesTask
}
