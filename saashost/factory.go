package saashost

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
	"github.com/domainry/domainry-agent-sdk/modulehost"
)

type Host interface{ RuntimeID() string }
type PersistenceHost interface {
	Host
	Database() modulehost.Database
	Dialect() modulehost.Dialect
	Migrations() modulehost.MigrationRegistrar
}
type Factory interface {
	OpenSaaS(context.Context, agentsdk.ApplicationRef, Host) (agentsdk.Binding, error)
}
