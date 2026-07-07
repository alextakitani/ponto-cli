package harness

import "fmt"

// SharedFixture is intentionally empty during Phase 1. Domain fixtures return in Phase 2.
type SharedFixture struct {
	cfg *Config
}

func NewSharedFixture(cfg *Config) (*SharedFixture, error) {
	return &SharedFixture{cfg: cfg}, nil
}

func (f *SharedFixture) Teardown() error {
	if f == nil {
		return fmt.Errorf("nil fixture")
	}
	return nil
}
