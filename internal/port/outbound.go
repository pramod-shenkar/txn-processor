package port

import (
	"context"
)

type Outbound interface {
	HealthDao
}

type HealthDao interface {
	Ping(context.Context) error
}
