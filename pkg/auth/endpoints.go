package auth

import (
	"net/url"
	"path"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type OriginProvider interface {
	Origin() *url.URL
}

type MainOriginProvider struct {
	HTTPHost  httputil.HTTPHost
	HTTPProto httputil.HTTPProto
}

func (p *MainOriginProvider) Origin() *url.URL {
	return &url.URL{
		Host:   string(p.HTTPHost),
		Scheme: string(p.HTTPProto),
	}
}

type EndpointsProvider struct {
	OriginProvider OriginProvider
}

func (p *EndpointsProvider) BaseURL() *url.URL {
	return p.OriginProvider.Origin()
}

func (p *EndpointsProvider) urlOf(relPath string) *url.URL {
	u := p.BaseURL()
	u.Path = path.Join(u.Path, relPath)
	return u
}

func (p *EndpointsProvider) AuthorizeEndpointURL() *url.URL     { return p.urlOf("oauth2/authorize") }
func (p *EndpointsProvider) FromWebAppEndpointURL() *url.URL    { return p.urlOf("oauth2/_from_webapp") }
func (p *EndpointsProvider) TokenEndpointURL() *url.URL         { return p.urlOf("oauth2/token") }
func (p *EndpointsProvider) RevokeEndpointURL() *url.URL        { return p.urlOf("oauth2/revoke") }
func (p *EndpointsProvider) JWKSEndpointURL() *url.URL          { return p.urlOf("oauth2/jwks") }
func (p *EndpointsProvider) UserInfoEndpointURL() *url.URL      { return p.urlOf("oauth2/userinfo") }
func (p *EndpointsProvider) EndSessionEndpointURL() *url.URL    { return p.urlOf("oauth2/end_session") }
func (p *EndpointsProvider) OAuthEntrypointURL() *url.URL       { return p.urlOf("./_oauth_entrypoint") }
func (p *EndpointsProvider) LoginEndpointURL() *url.URL         { return p.urlOf("./login") }
func (p *EndpointsProvider) SignupEndpointURL() *url.URL        { return p.urlOf("./signup") }
func (p *EndpointsProvider) PromoteUserEndpointURL() *url.URL   { return p.urlOf("./promote_user") }
func (p *EndpointsProvider) LogoutEndpointURL() *url.URL        { return p.urlOf("./logout") }
func (p *EndpointsProvider) SettingsEndpointURL() *url.URL      { return p.urlOf("./settings") }
func (p *EndpointsProvider) ResetPasswordEndpointURL() *url.URL { return p.urlOf("./reset_password") }
func (p *EndpointsProvider) SSOCallbackEndpointURL() *url.URL   { return p.urlOf("sso/oauth2/callback") }

func (p *EndpointsProvider) WeChatAuthorizeEndpointURL() *url.URL { return p.urlOf("sso/wechat/auth") }
func (p *EndpointsProvider) WeChatCallbackEndpointURL() *url.URL {
	return p.urlOf("sso/wechat/callback")
}
