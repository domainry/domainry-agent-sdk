package modulehost

import (
	"context"
	agentsdk "github.com/domainry/domainry-agent-sdk"
)

type Host interface{ RuntimeID() string }
type Factory interface {
	OpenModule(context.Context, agentsdk.ApplicationRef, Host) (agentsdk.Binding, error)
}
