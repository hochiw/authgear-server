package nodes

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/successpage"
)

func init() {
	interaction.RegisterNode(&NodeResetPasswordBegin{})
	interaction.RegisterNode(&NodeResetPasswordEnd{})
}

type NodeResetPasswordBegin struct{}

func (n *NodeResetPasswordBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeResetPasswordBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{&EdgeResetPassword{}}, nil
}

type InputResetPassword interface {
	GetResetPasswordUserID() string
	GetNewPassword() string
}

type InputResetPasswordByCode interface {
	GetCode() string
	GetNewPassword() string
}

type EdgeResetPassword struct{}

func (e *EdgeResetPassword) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var resetInput InputResetPassword
	var codeInput InputResetPasswordByCode
	successPageCookie := ctx.CookieManager.ValueCookie(successpage.PathCookieDef, "/reset_password/success")
	if interaction.Input(rawInput, &resetInput) {
		userID := resetInput.GetResetPasswordUserID()
		newPassword := resetInput.GetNewPassword()

		oldInfo, newInfo, err := ctx.ResetPassword.ResetPassword(userID, newPassword)
		if err != nil {
			return nil, err
		}

		return &NodeResetPasswordEnd{
			OldAuthenticator:  oldInfo,
			NewAuthenticator:  newInfo,
			SuccessPageCookie: successPageCookie,
		}, nil

	} else if interaction.Input(rawInput, &codeInput) {
		code := codeInput.GetCode()
		newPassword := codeInput.GetNewPassword()

		codeHash := ctx.ResetPassword.HashCode(code)
		oldInfo, newInfo, err := ctx.ResetPassword.ResetPasswordByCode(code, newPassword)
		if err != nil {
			return nil, err
		}

		err = ctx.ResetPassword.AfterResetPasswordByCode(codeHash)
		if err != nil {
			return nil, err
		}

		return &NodeResetPasswordEnd{
			OldAuthenticator:  oldInfo,
			NewAuthenticator:  newInfo,
			SuccessPageCookie: successPageCookie,
		}, nil
	} else {
		return nil, interaction.ErrIncompatibleInput
	}
}

type NodeResetPasswordEnd struct {
	OldAuthenticator  *authenticator.Info `json:"old_authenticator"`
	NewAuthenticator  *authenticator.Info `json:"new_authenticator,omitempty"`
	SuccessPageCookie *http.Cookie        `json:"success_page_cookie,omitempty"`
}

// GetCookies implements CookiesGetter
func (n *NodeResetPasswordEnd) GetCookies() []*http.Cookie {
	if n.SuccessPageCookie == nil {
		return nil
	}
	return []*http.Cookie{n.SuccessPageCookie}
}

func (n *NodeResetPasswordEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeResetPasswordEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeResetPasswordEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	// Password authenticator is always primary for now
	if n.NewAuthenticator != nil {
		return []interaction.Edge{
			&EdgeDoUpdateAuthenticator{
				Stage:                     authn.AuthenticationStagePrimary,
				AuthenticatorBeforeUpdate: n.OldAuthenticator,
				AuthenticatorAfterUpdate:  n.NewAuthenticator,
			},
		}, nil
	}

	// Password is not changed, ends the interaction
	return nil, nil
}
