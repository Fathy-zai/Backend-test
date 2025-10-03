package providers

import (
	"context"
)

type Measurement struct {
	TempC    float64
	WindKph float64
}

type Provider interface {
	Name() string
	Get(ctx context.Context, city string) (Measurement, error)
}
