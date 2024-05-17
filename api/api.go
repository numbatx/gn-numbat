package api

import (
	"reflect"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/numbatx/gn-numbat/api/block"
	"gopkg.in/go-playground/validator.v8"

	"github.com/numbatx/gn-numbat/api/address"
	"github.com/numbatx/gn-numbat/api/middleware"
	"github.com/numbatx/gn-numbat/api/node"
	"github.com/numbatx/gn-numbat/api/transaction"
)

type validatorInput struct {
	Name      string
	Validator validator.Func
}

// Start will boot up the api and appropriate routes, handlers and validators
func Start(numbatFacade middleware.NumbatHandler) error {
	ws := gin.Default()
	ws.Use(cors.Default())

	err := registerValidators()
	if err != nil {
		return err
	}
	registerRoutes(ws, numbatFacade)

	return ws.Run()
}

func registerRoutes(ws *gin.Engine, numbatFacade middleware.NumbatHandler) {
	nodeRoutes := ws.Group("/node")
	nodeRoutes.Use(middleware.WithNumbatFacade(numbatFacade))
	node.Routes(nodeRoutes)

	addressRoutes := ws.Group("/address")
	addressRoutes.Use(middleware.WithNumbatFacade(numbatFacade))
	address.Routes(addressRoutes)

	txRoutes := ws.Group("/transaction")
	txRoutes.Use(middleware.WithNumbatFacade(numbatFacade))
	transaction.Routes(txRoutes)

	txsRoutes := ws.Group("/transactions")
	txsRoutes.Use(middleware.WithNumbatFacade(numbatFacade))
	transaction.RoutesForTransactionsLists(txsRoutes)

	blockRoutes := ws.Group("/block")
	blockRoutes.Use(middleware.WithNumbatFacade(numbatFacade))
	block.Routes(blockRoutes)

	blocksRoutes := ws.Group("/blocks")
	blocksRoutes.Use(middleware.WithNumbatFacade(numbatFacade))
	block.RoutesForBlocksLists(blocksRoutes)
}

func registerValidators() error {
	validators := []validatorInput{
		{Name: "skValidator", Validator: skValidator},
	}
	for _, validatorFunc := range validators {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			err := v.RegisterValidation(validatorFunc.Name, validatorFunc.Validator)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// skValidator validates a secret key from user input for correctness
func skValidator(
	_ *validator.Validate,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Value,
	_ reflect.Type,
	_ reflect.Kind,
	_ string,
) bool {
	return true
}
