package nonblocking

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func checkIdentityEventSupportLoginIDType(loginIDType string) bool {
	return loginIDType == string(config.LoginIDKeyTypeEmail) ||
		loginIDType == string(config.LoginIDKeyTypePhone) ||
		loginIDType == string(config.LoginIDKeyTypeUsername)
}

const (
	IdentityLoginIDAddedFormat string = "identity.%s.added"
)

type IdentityLoginIDAddedEvent struct {
	User        model.User     `json:"user"`
	Identity    model.Identity `json:"identity"`
	LoginIDType string         `json:"-"`
}

func NewIdentityLoginIDAddedEvent(
	user model.User,
	identity model.Identity,
	loginIDType string,
) *IdentityLoginIDAddedEvent {
	if checkIdentityEventSupportLoginIDType(loginIDType) {
		return &IdentityLoginIDAddedEvent{
			User:        user,
			Identity:    identity,
			LoginIDType: loginIDType,
		}
	}
	return nil
}

func (e *IdentityLoginIDAddedEvent) NonBlockingEventType() event.Type {
	return event.Type(fmt.Sprintf(IdentityLoginIDAddedFormat, e.LoginIDType))
}

func (e *IdentityLoginIDAddedEvent) UserID() string {
	return e.User.ID
}

var _ event.NonBlockingPayload = &IdentityLoginIDAddedEvent{}
