package internal

import (
	"context"
	"fmt"

	h "main/internal/handler"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"go.uber.org/fx"
)

type App struct {
	chaincode *contractapi.ContractChaincode
}

func NewApp(lc fx.Lifecycle, handler *h.Service) *App {
	srv := &App{}
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			var err error
			srv.chaincode, err = contractapi.NewChaincode(handler)
			if err != nil {
				return fmt.Errorf("error starting marble02 chaincode: %v", err)
			}

			if err = srv.chaincode.Start(); err != nil {
				return fmt.Errorf("error starting marble02 chaincode: %v", err)
			}

			return nil
		},
	})

	return srv
}
