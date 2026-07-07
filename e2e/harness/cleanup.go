package harness

// CleanupTracker is a placeholder until Phase 2 domain e2e tests are added.
type CleanupTracker struct{}

func NewCleanupTracker() *CleanupTracker {
	return &CleanupTracker{}
}

func (c *CleanupTracker) CleanupAll(h *Harness) []error {
	return nil
}
