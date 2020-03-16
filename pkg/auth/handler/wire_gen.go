// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package handler

import (
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authn"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	pq3 "github.com/skygeario/skygear-server/pkg/auth/dependency/mfa/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	auth2 "github.com/skygeario/skygear-server/pkg/core/auth"
	pq2 "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/mail"
	"github.com/skygeario/skygear-server/pkg/core/sms"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	"net/http"
)

// Injectors from wire.go:

func newLoginHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	contextGetter := auth2.ProvideAuthContextGetter(context)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	requireAuthz := handler.NewRequireAuthzFactory(contextGetter, factory)
	validator := auth.ProvideValidator(m)
	provider := time.NewProvider()
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := pq.ProvidePasswordHistoryStore(provider, sqlBuilder, sqlExecutor)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	passwordProvider := password.ProvidePasswordProvider(sqlBuilder, sqlExecutor, provider, store, factory, tenantConfiguration, reservedNameChecker)
	oauthProvider := oauth.ProvideOAuthProvider(sqlBuilder, sqlExecutor)
	v := auth.ProvidePrincipalProviders(oauthProvider, passwordProvider)
	identityProvider := principal.ProvideIdentityProvider(sqlBuilder, sqlExecutor, v)
	authenticateProcess := authn.ProvideAuthenticateProcess(factory, provider, passwordProvider, oauthProvider, identityProvider)
	passwordChecker := audit.ProvidePasswordChecker(tenantConfiguration, store)
	loginIDChecker := loginid.ProvideLoginIDChecker(tenantConfiguration, reservedNameChecker)
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	urlprefixProvider := urlprefix.NewProvider(r)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	hookProvider := hook.ProvideHookProvider(sqlBuilder, sqlExecutor, requestID, tenantConfiguration, urlprefixProvider, contextGetter, txContext, provider, authinfoStore, userprofileStore, passwordProvider, factory)
	executor := auth.ProvideTaskExecutor(m)
	queue := async.ProvideTaskQueue(context, txContext, requestID, tenantConfiguration, executor)
	signupProcess := authn.ProvideSignupProcess(passwordChecker, loginIDChecker, identityProvider, passwordProvider, oauthProvider, provider, authinfoStore, userprofileStore, hookProvider, tenantConfiguration, urlprefixProvider, queue)
	oAuthCoordinator := &authn.OAuthCoordinator{
		Authn:  authenticateProcess,
		Signup: signupProcess,
	}
	mfaStore := pq3.ProvideStore(tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	client := sms.ProvideSMSClient(tenantConfiguration)
	sender := mail.ProvideMailSender(tenantConfiguration)
	engine := auth.ProvideTemplateEngine(tenantConfiguration, m)
	mfaSender := mfa.ProvideMFASender(tenantConfiguration, client, sender, engine)
	mfaProvider := mfa.ProvideMFAProvider(mfaStore, tenantConfiguration, provider, mfaSender)
	sessionStore := redis.ProvideStore(context, tenantConfiguration, provider, factory)
	eventStore := redis.ProvideEventStore(context, tenantConfiguration)
	sessionProvider := session.ProvideSessionProvider(r, sessionStore, eventStore, tenantConfiguration)
	authnSessionProvider := authn.ProvideSessionProvider(mfaProvider, sessionProvider, tenantConfiguration, provider, authinfoStore, userprofileStore, identityProvider, hookProvider)
	authnProvider := &authn.Provider{
		OAuth:   oAuthCoordinator,
		Authn:   authenticateProcess,
		Signup:  signupProcess,
		Session: authnSessionProvider,
	}
	httpHandler := provideLoginHandler(requireAuthz, validator, authnProvider, txContext)
	return httpHandler
}

// wire.go:

func provideLoginHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	ap LoginAuthnProvider,
	tx db.TxContext,
) http.Handler {
	h := &LoginHandler{
		Validator:     v,
		AuthnProvider: ap,
		TxContext:     tx,
	}
	return requireAuthz(h, h)
}
