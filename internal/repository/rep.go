package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/model"

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
	// txTimestamp, err := stub.GetTxTimestamp()
	// if err != nil {
	// 	return fmt.Errorf("failed to get tx timestamp: %w", err)
	// }

	// asset.CreatedAt = txTimestamp.AsTime()
	// asset.UpdatedAt = txTimestamp.AsTime()

	// buf, err := json.Marshal(asset)
	// if err != nil {
	// 	return fmt.Errorf("failed to json marshal asset: %w", err)
	// }

	// err = stub.PutState(asset.ID, buf)
	// if err != nil {
	// 	return fmt.Errorf("failed to stub put asset: %w", err)
	// }

	//objectType := "marble"
	//marble := &marble{objectType, marbleName, color, size, owner}
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)

	// === Save marble to state ===
	err = stub.PutState(marble.Name, marbleJSONasBytes)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init marble")
	//return shim.Success(nil)

	return nil
}

func (rep *RepositoryRealization) ReadMarble(ctx context.Context, name string) ([]byte, error) {
	var jsonResp string
	var err error

	// if len(args) != 1 {
	// 	return nil, fmt.Errorf("incorrect number of arguments. Expecting name of the marble to query")
	// }

	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return nil, fmt.Errorf("no 'stub' in context")
	}

	stub := ptrStub.(shim.ChaincodeStubInterface)
	valAsbytes, err := stub.GetState(name) //get the marble from chaincode state
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
	err := stub.DelState(marbleJSON.Name) //remove the marble from chaincode state
	if err != nil {
		return fmt.Errorf("%s", "failed to delete state:"+err.Error())
	}

	// maintain the index
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marbleJSON.Color, marbleJSON.Name})
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(colorNameIndexKey)
	if err != nil {
		return fmt.Errorf("%s", "failed to delete state:"+err.Error())
	}
	return nil
}
