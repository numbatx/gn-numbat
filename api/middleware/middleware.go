package middleware

import (
	"github.com/gin-gonic/gin"
)

// NumbatHandler interface defines methods that can be used from `numbatFacade` context variable
type NumbatHandler interface {
}

// WithNumbatFacade middleware will set up an NumbatFacade object in the gin context
func WithNumbatFacade(numbatFacade NumbatHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("numbatFacade", numbatFacade)
		c.Next()
	}
}
