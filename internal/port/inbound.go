package port

import (
	"context"
)

type Inbound interface {
	HealthService
}

type HealthService interface {
	Check(ctx context.Context) error
}
