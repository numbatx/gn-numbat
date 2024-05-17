package facade

import (
	"github.com/numbatx/gn-numbat/core/logger"
	"github.com/numbatx/gn-numbat/ntp"
)

// GetLogger returns the current logger
func (ef *NumbatNodeFacade) GetLogger() *logger.Logger {
	return ef.log
}

// GetSyncer returns the current syncer
func (ef *NumbatNodeFacade) GetSyncer() ntp.SyncTimer {
	return ef.syncer
}
