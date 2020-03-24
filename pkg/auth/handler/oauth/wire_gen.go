// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package oauth

import (
	"github.com/skygeario/skygear-server/pkg/auth"
	auth2 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	redis2 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	pq3 "github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	oauth2 "github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	redis3 "github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	pq2 "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"net/http"
)

// Injectors from wire.go:

func newAuthorizeHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	authorizationStore := &pq.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	provider := time.NewProvider()
	grantStore := redis.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	urlprefixProvider := urlprefix.NewProvider(r)
	endpointsProvider := &auth.EndpointsProvider{
		PrefixProvider: urlprefixProvider,
	}
	scopesValidator := _wireScopesValidatorValue
	tokenGenerator := _wireTokenGeneratorValue
	authorizationHandler := handler.ProvideAuthorizationHandler(context, tenantConfiguration, factory, authorizationStore, grantStore, endpointsProvider, endpointsProvider, scopesValidator, tokenGenerator, provider)
	httpHandler := provideAuthorizeHandler(factory, txContext, authorizationHandler)
	return httpHandler
}

var (
	_wireScopesValidatorValue = handler.ScopesValidator(oidc.ValidateScopes)
	_wireTokenGeneratorValue  = handler.TokenGenerator(oauth.GenerateToken)
)

func newTokenHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	authorizationStore := &pq.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	provider := time.NewProvider()
	grantStore := redis.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	eventStore := redis2.ProvideEventStore(context, tenantConfiguration)
	accessEventProvider := auth2.AccessEventProvider{
		Store: eventStore,
	}
	store := redis3.ProvideStore(context, tenantConfiguration, provider, factory)
	authAccessEventProvider := &auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionProvider := session.ProvideSessionProvider(r, store, authAccessEventProvider, tenantConfiguration)
	urlprefixProvider := urlprefix.NewProvider(r)
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	oauthProvider := oauth2.ProvideOAuthProvider(sqlBuilder, sqlExecutor)
	passwordhistoryStore := pq3.ProvidePasswordHistoryStore(provider, sqlBuilder, sqlExecutor)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	passwordProvider := password.ProvidePasswordProvider(sqlBuilder, sqlExecutor, provider, passwordhistoryStore, factory, tenantConfiguration, reservedNameChecker)
	v := auth.ProvidePrincipalProviders(oauthProvider, passwordProvider)
	identityProvider := principal.ProvideIdentityProvider(sqlBuilder, sqlExecutor, v)
	idTokenIssuer := oidc.ProvideIDTokenIssuer(tenantConfiguration, urlprefixProvider, authinfoStore, userprofileStore, identityProvider, provider)
	tokenGenerator := _wireTokenGeneratorValue
	tokenHandler := handler.ProvideTokenHandler(context, tenantConfiguration, factory, authorizationStore, grantStore, grantStore, grantStore, accessEventProvider, sessionProvider, idTokenIssuer, tokenGenerator, provider)
	httpHandler := provideTokenHandler(factory, txContext, tokenHandler)
	return httpHandler
}

func newRevokeHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	provider := time.NewProvider()
	grantStore := redis.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	revokeHandler := &handler.RevokeHandler{
		OfflineGrants: grantStore,
		AccessGrants:  grantStore,
	}
	httpHandler := provideRevokeHandler(factory, txContext, revokeHandler)
	return httpHandler
}

func newMetadataHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	provider := urlprefix.NewProvider(r)
	endpointsProvider := &auth.EndpointsProvider{
		PrefixProvider: provider,
	}
	metadataProvider := &oauth.MetadataProvider{
		AuthorizeEndpoint:    endpointsProvider,
		TokenEndpoint:        endpointsProvider,
		RevokeEndpoint:       endpointsProvider,
		AuthenticateEndpoint: endpointsProvider,
	}
	oidcMetadataProvider := &oidc.MetadataProvider{
		URLPrefix:        provider,
		JWKSEndpoint:     endpointsProvider,
		UserInfoEndpoint: endpointsProvider,
	}
	httpHandler := provideMetadataHandler(metadataProvider, oidcMetadataProvider)
	return httpHandler
}

func newJWKSHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	tenantConfiguration := auth.ProvideTenantConfig(context)
	httpHandler := provideJWKSHandler(tenantConfiguration)
	return httpHandler
}

func newUserInfoHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	provider := urlprefix.NewProvider(r)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	timeProvider := time.NewProvider()
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	userprofileStore := userprofile.ProvideStore(timeProvider, sqlBuilder, sqlExecutor)
	oauthProvider := oauth2.ProvideOAuthProvider(sqlBuilder, sqlExecutor)
	passwordhistoryStore := pq3.ProvidePasswordHistoryStore(timeProvider, sqlBuilder, sqlExecutor)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	passwordProvider := password.ProvidePasswordProvider(sqlBuilder, sqlExecutor, timeProvider, passwordhistoryStore, factory, tenantConfiguration, reservedNameChecker)
	v := auth.ProvidePrincipalProviders(oauthProvider, passwordProvider)
	identityProvider := principal.ProvideIdentityProvider(sqlBuilder, sqlExecutor, v)
	idTokenIssuer := oidc.ProvideIDTokenIssuer(tenantConfiguration, provider, store, userprofileStore, identityProvider, timeProvider)
	httpHandler := provideUserInfoHandler(factory, txContext, idTokenIssuer)
	return httpHandler
}

// wire.go:

func provideAuthorizeHandler(lf logging.Factory, tx db.TxContext, ah oauthAuthorizeHandler) http.Handler {
	h := &AuthorizeHandler{
		logger:       lf.NewLogger("oauth-authz-handler"),
		txContext:    tx,
		authzHandler: ah,
	}
	return h
}

func provideTokenHandler(lf logging.Factory, tx db.TxContext, th oauthTokenHandler) http.Handler {
	h := &TokenHandler{
		logger:       lf.NewLogger("oauth-token-handler"),
		txContext:    tx,
		tokenHandler: th,
	}
	return h
}

func provideRevokeHandler(lf logging.Factory, tx db.TxContext, rh oauthRevokeHandler) http.Handler {
	h := &RevokeHandler{
		logger:        lf.NewLogger("oauth-revoke-handler"),
		txContext:     tx,
		revokeHandler: rh,
	}
	return h
}

func provideMetadataHandler(oauth3 *oauth.MetadataProvider, oidc2 *oidc.MetadataProvider) http.Handler {
	h := &MetadataHandler{
		metaProviders: []oauthMetadataProvider{oauth3, oidc2},
	}
	return h
}

func provideJWKSHandler(config2 *config.TenantConfiguration) http.Handler {
	h := &JWKSHandler{
		config: *config2.AppConfig.OIDC,
	}
	return h
}

func provideUserInfoHandler(lf logging.Factory, tx db.TxContext, uip oauthUserInfoProvider) http.Handler {
	h := &UserInfoHandler{
		logger:           lf.NewLogger("oauth-userinfo-handler"),
		txContext:        tx,
		userInfoProvider: uip,
	}
	return h
}
