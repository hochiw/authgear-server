package loginid

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func ProvideLoginIDChecker(
	config *config.TenantConfiguration,
	reservedNameChecker *ReservedNameChecker,
) LoginIDChecker {
	return NewDefaultLoginIDChecker(
		config.AppConfig.Auth.LoginIDKeys,
		config.AppConfig.Auth.LoginIDTypes,
		reservedNameChecker,
	)
}

var DependencySet = wire.NewSet(ProvideLoginIDChecker)
