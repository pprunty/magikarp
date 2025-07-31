package speech

import (
	"context"

	"fmt"

	"github.com/pprunty/magikarp/internal/config"
)

// Listen is a stub implementation used when the "speech" build tag is not set.
// It immediately returns an error indicating speech support is disabled.
func Listen(_ context.Context, _ chan<- string, _ config.Speech) error {
	return fmt.Errorf("speech support is not enabled; build with '-tags speech' to enable")
}
