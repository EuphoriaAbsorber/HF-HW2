package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/model"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type RepositoryRealization struct {
}

func newRepository() RepositoryInterface {
	return &RepositoryRealization{}
}

func (rep *RepositoryRealization) CreateMarble(ctx context.Context, marble *model.Marble) error {
	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return fmt.Errorf("no 'stub' in context")
	}
	stub := ptrStub.(shim.ChaincodeStubInterface)
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	err = stub.PutState(marble.Name, marbleJSONasBytes)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)
	return nil
}

func (rep *RepositoryRealization) ReadMarble(ctx context.Context, name string) ([]byte, error) {
	var jsonResp string
	var err error

	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return nil, fmt.Errorf("no 'stub' in context")
	}

	stub := ptrStub.(shim.ChaincodeStubInterface)
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, fmt.Errorf("%s", jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + name + "\"}"
		return nil, fmt.Errorf("%s", jsonResp)
	}
	return valAsbytes, nil
}

func (rep *RepositoryRealization) DeleteMarble(ctx context.Context, marbleJSON model.Marble) error {
	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return fmt.Errorf("no 'stub' in context")
	}
	stub := ptrStub.(shim.ChaincodeStubInterface)
	err := stub.DelState(marbleJSON.Name)
	if err != nil {
		return fmt.Errorf("%s", "failed to delete state:"+err.Error())
	}
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marbleJSON.Color, marbleJSON.Name})
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	err = stub.DelState(colorNameIndexKey)
	if err != nil {
		return fmt.Errorf("%s", "failed to delete state:"+err.Error())
	}
	return nil
}

func (rep *RepositoryRealization) TransferMarble(ctx context.Context, marbleToTransfer model.Marble) error {
	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return fmt.Errorf("no 'stub' in context")
	}
	stub := ptrStub.(shim.ChaincodeStubInterface)
	marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
	err := stub.PutState(marbleToTransfer.Name, marbleJSONasBytes)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	return nil
}

func (rep *RepositoryRealization) GetMarbleHistory(ctx context.Context, id string) (map[string]model.Marble, map[string]time.Time, error) {
	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return nil, nil, fmt.Errorf("no 'stub' in context")
	}
	stub := ptrStub.(shim.ChaincodeStubInterface)

	iter, err := stub.GetHistoryForKey(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get history for marble: %w", err)
	}

	resultMarbles := make(map[string]model.Marble, 0)
	resultTimstamps := make(map[string]time.Time, 0)

	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read marble history: %w", err)
		}
		var marble model.Marble
		err = json.Unmarshal(kv.Value, &marble)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal marble history: %w", err)
		}
		resultTimstamps[kv.TxId] = kv.Timestamp.AsTime()
		resultMarbles[kv.TxId] = marble
	}
	return resultMarbles, resultTimstamps, nil
}
