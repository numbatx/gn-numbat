package chronology

import (
	"errors"
)

// ErrNilRounder is raised when a valid rounder is expected but nil used
var ErrNilRounder = errors.New("rounder is nil")

// ErrNilSyncTimer is raised when a valid sync timer is expected but nil used
var ErrNilSyncTimer = errors.New("sync timer is nil")
