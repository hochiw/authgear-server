package oauth

import "net/url"

type EndpointsProvider interface {
	AuthorizeEndpointURL() *url.URL
	FromWebAppEndpointURL() *url.URL
	TokenEndpointURL() *url.URL
	RevokeEndpointURL() *url.URL
}
