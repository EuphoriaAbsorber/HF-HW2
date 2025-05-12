package handler

import (
	"context"
	"encoding/json"
	"fmt"
	log "main/internal/logic"
	"main/internal/model"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Service struct {
	contractapi.Contract

	bl *log.Service
}

func NewService(bl *log.Service) *Service {
	return &Service{
		bl: bl,
	}
}

// func (s *Service) CreateMarble(hlfContext contractapi.TransactionContextInterface, req dto.UpsertAssetRequest) error {

// 	cert, err := hlfContext.GetClientIdentity().GetX509Certificate()
// 	if err != nil {
// 		return fmt.Errorf("failed to get x509 certificate: %v", err)
// 	}
// 	clientID := cert.Subject.CommonName

// 	ctx := context.Background()
// 	context.WithValue(ctx, "stub", hlfContext.GetStub())

// 	err = s.bl.CreateMarble(ctx, req.ID, req.AppraisedValue, req.Color, req.Size, clientID)
// 	if err != nil {
// 		// log err
// 		return fmt.Errorf("could not create asset")
// 	}

// 	return nil
// }

func (s *Service) InitMarble(hlfContext contractapi.TransactionContextInterface, name string, color string, size int, owner string) error {
	// var err error
	// cert, err := hlfContext.GetClientIdentity().GetX509Certificate()
	// if err != nil {
	// 	return fmt.Errorf("failed to get x509 certificate: %v", err)
	// }
	// clientID := cert.Subject.CommonName

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	// if len(args) != 4 {
	// 	return fmt.Errorf("incorrect number of arguments. Expecting 4")
	// }

	// ==== Input sanitation ====
	// fmt.Println("- start init marble")
	// if len(args[0]) <= 0 {
	// 	return fmt.Errorf("1st argument must be a non-empty string")
	// }
	// if len(args[1]) <= 0 {
	// 	return fmt.Errorf("2nd argument must be a non-empty string")
	// }
	// if len(args[2]) <= 0 {
	// 	return fmt.Errorf("3rd argument must be a non-empty string")
	// }
	// if len(args[3]) <= 0 {
	// 	return fmt.Errorf("4th argument must be a non-empty string")
	// }
	// marbleName := args[0]
	// color := strings.ToLower(args[1])
	// owner := strings.ToLower(args[3])
	// size, err := strconv.Atoi(args[2])
	// if err != nil {
	// 	return fmt.Errorf("3rd argument must be a numeric string")
	// }
	marbleName := name
	stub := hlfContext.GetStub()
	// ==== Check if marble already exists ====
	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return fmt.Errorf("%s", "Failed to get marble: "+err.Error())
	} else if marbleAsBytes != nil {
		fmt.Println("This marble already exists: " + marbleName)
		return fmt.Errorf("%s", "This marble already exists: "+marbleName)
	}

	ctx := context.Background()
	context.WithValue(ctx, "stub", stub)

	// ==== Create marble object and marshal to JSON ====
	objectType := "marble"
	marble := &model.Marble{ObjectType: objectType, Name: marbleName, Color: color, Size: size, Owner: owner}
	err = s.bl.CreateMarble(ctx, marble)
	if err != nil {
		return fmt.Errorf("could not create marble")
	}

	return nil

}

func (s *Service) ReadMarble(hlfContext contractapi.TransactionContextInterface, name string) ([]byte, error) {
	//var name, jsonResp string
	var jsonResp string
	var err error

	// if len(args) != 1 {
	// 	return nil, fmt.Errorf("incorrect number of arguments. Expecting name of the marble to query")
	// }

	stub := hlfContext.GetStub()
	//name = args[0]
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

func (s *Service) Delete(hlfContext contractapi.TransactionContextInterface, marbleName string) error {
	var jsonResp string
	var marbleJSON model.Marble
	// if len(args) != 1 {
	// 	return fmt.Errorf("incorrect number of arguments. Expecting 1")
	// }
	//marbleName := args[0]

	stub := hlfContext.GetStub()
	// to maintain the color~name index, we need to read the marble first and get its color
	valAsbytes, err := stub.GetState(marbleName) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + marbleName + "\"}"
		return fmt.Errorf("%s", jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + marbleName + "\"}"
		return fmt.Errorf("%s", jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &marbleJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + marbleName + "\"}"
		return fmt.Errorf("%s", jsonResp)
	}

	err = stub.DelState(marbleName) //remove the marble from chaincode state
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

func (s *Service) TransferMarble(hlfContext contractapi.TransactionContextInterface, args []string) error {
	stub := hlfContext.GetStub()
	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return fmt.Errorf("%s", "Incorrect number of arguments. Expecting 2")
	}

	marbleName := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferMarble ", marbleName, newOwner)

	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return fmt.Errorf("%s", "Failed to get marble:"+err.Error())
	} else if marbleAsBytes == nil {
		return fmt.Errorf("%s", "Marble does not exist")
	}

	marbleToTransfer := &model.Marble{}
	err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	marbleToTransfer.Owner = newOwner //change the owner

	marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
	err = stub.PutState(marbleName, marbleJSONasBytes) //rewrite the marble
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	fmt.Println("- end transferMarble (success)")
	return nil
}
