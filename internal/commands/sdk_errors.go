package commands

import (
	"errors"
	"net"

	"github.com/basecamp/cli/output"
)

// convertSDKError is kept as a compatibility shim for shared error paths.
// Phase 1 removes the generated SDK, so only generic network errors are mapped.
func convertSDKError(err error) error {
	if err == nil {
		return nil
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return &output.Error{
			Code:      output.CodeNetwork,
			Message:   netErr.Error(),
			Hint:      "Check your internet connection",
			Retryable: true,
		}
	}

	return err
}
