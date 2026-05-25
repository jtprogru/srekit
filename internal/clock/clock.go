// Package clock exposes the package-level wall clock used by srekit commands.
// Tests can swap Now for deterministic output.
package clock

import "time"

// Now is the wall clock. Override in tests for deterministic output.
//
//nolint:gochecknoglobals // intentional package-level injection point
var Now = time.Now
