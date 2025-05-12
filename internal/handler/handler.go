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
	// marbleName := name
	// stub := hlfContext.GetStub()
	// // ==== Check if marble already exists ====
	// marbleAsBytes, err := stub.GetState(marbleName)
	// if err != nil {
	// 	return fmt.Errorf("%s", "Failed to get marble: "+err.Error())
	// } else if marbleAsBytes != nil {
	// 	fmt.Println("This marble already exists: " + marbleName)
	// 	return fmt.Errorf("%s", "This marble already exists: "+marbleName)
	// }

	// ctx := context.Background()
	// context.WithValue(ctx, "stub", stub)

	// // ==== Create marble object and marshal to JSON ====
	// objectType := "marble"
	// marble := &model.Marble{ObjectType: objectType, Name: marbleName, Color: color, Size: size, Owner: owner}
	// err = s.bl.CreateMarble(ctx, marble)
	// if err != nil {
	// 	return fmt.Errorf("could not create marble")
	// }

	// return nil
	color = strings.ToLower(color)
	owner = strings.ToLower(owner)
	marbleName := name
	stub := hlfContext.GetStub()
	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return fmt.Errorf("Failed to get marble: " + err.Error())
	} else if marbleAsBytes != nil {
		fmt.Println("This marble already exists: " + marbleName)
		return fmt.Errorf("This marble already exists: " + marbleName)
	}

	// ==== Create marble object and marshal to JSON ====
	objectType := "marble"
	marble := &model.Marble{ObjectType: objectType, Name: marbleName, Color: color, Size: size, Owner: owner}
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)

	// === Save marble to state ===
	err = stub.PutState(marbleName, marbleJSONasBytes)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init marble")
	return nil
}

func (s *Service) ReadMarble(hlfContext contractapi.TransactionContextInterface, name string) (*model.Marble, error) {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)

	bytes, err := s.bl.ReadMarble(ctx, name)
	if err != nil {
		return nil, err
	}
	marble := model.Marble{}
	err = json.Unmarshal([]byte(bytes), &marble)
	if err != nil {
		return nil, err
	}

	return &marble, nil
}

func (s *Service) Delete(hlfContext contractapi.TransactionContextInterface, marbleName string) error {
	var jsonResp string
	var marbleJSON model.Marble

	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)
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
	return s.bl.DeleteMarble(ctx, marbleJSON)
}

func (s *Service) TransferMarble(hlfContext contractapi.TransactionContextInterface, marbleName string, newOwner string) error {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)
	//   0       1
	// "name", "bob"
	// if len(args) < 2 {
	// 	return fmt.Errorf("%s", "Incorrect number of arguments. Expecting 2")
	// }

	//marbleName := args[0]
	newOwner = strings.ToLower(newOwner)
	fmt.Println("- start transferMarble ", marbleName, newOwner)
	s.bl.TransferMarble(ctx, marbleName, newOwner)
	// marbleAsBytes, err := stub.GetState(marbleName)
	// if err != nil {
	// 	return fmt.Errorf("%s", "Failed to get marble:"+err.Error())
	// } else if marbleAsBytes == nil {
	// 	return fmt.Errorf("%s", "Marble does not exist")
	// }

	// marbleToTransfer := &model.Marble{}
	// err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
	// if err != nil {
	// 	return fmt.Errorf("%s", err.Error())
	// }
	// marbleToTransfer.Owner = newOwner //change the owner
	//
	//
	// s.bl.TransferMarble(ctx, *marbleToTransfer)
	//s.bl.TransferMarble()
	// marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
	// err = stub.PutState(marbleName, marbleJSONasBytes) //rewrite the marble
	// if err != nil {
	// 	return fmt.Errorf("%s", err.Error())
	// }

	fmt.Println("- end transferMarble (success)")
	return nil
}

func (s *Service) TransferMarblesBasedOnColor(hlfContext contractapi.TransactionContextInterface, color string, newOwner string) error {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)
	//   0       1
	// "color", "bob"
	// if len(args) < 2 {
	// 	return shim.Error("Incorrect number of arguments. Expecting 2")
	// }

	//color := args[0]
	newOwner = strings.ToLower(newOwner)
	fmt.Println("- start transferMarblesBasedOnColor ", color, newOwner)
	s.bl.TransferMarblesBasedOnColor(ctx, color, newOwner)
	// Query the color~name index by color
	// This will execute a key range query on all keys starting with 'color'
	// coloredMarbleResultsIterator, err := stub.GetStateByPartialCompositeKey("color~name", []string{color})
	// if err != nil {
	// 	return fmt.Errorf("%s", err.Error())
	// }
	// defer coloredMarbleResultsIterator.Close()

	// // Iterate through result set and for each marble found, transfer to newOwner
	// var i int
	// for i = 0; coloredMarbleResultsIterator.HasNext(); i++ {
	// 	// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
	// 	responseRange, err := coloredMarbleResultsIterator.Next()
	// 	if err != nil {
	// 		return fmt.Errorf("%s", err.Error())
	// 	}

	// 	// get the color and name from color~name composite key
	// 	objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
	// 	if err != nil {
	// 		return fmt.Errorf("%s", err.Error())
	// 	}
	// 	returnedColor := compositeKeyParts[0]
	// 	returnedMarbleName := compositeKeyParts[1]
	// 	fmt.Printf("- found a marble from index:%s color:%s name:%s\n", objectType, returnedColor, returnedMarbleName)

	// 	// Now call the transfer function for the found marble.
	// 	// Re-use the same function that is used to transfer individual marbles
	// 	err = s.TransferMarble(stub, []string{returnedMarbleName, newOwner})
	// 	// if the transfer failed break out of loop and return error
	// 	if err != nil {
	// 		return fmt.Errorf("%s", "Transfer failed: "+err.Error())
	// 	}
	// 	// if response.Status != shim.OK {
	// 	// 	return shim.Error("Transfer failed: " + response.Message)
	// 	// }
	// }

	responsePayload := fmt.Sprintf("Transferred %d %s marbles to %s", i, color, newOwner)
	fmt.Println("- end transferMarblesBasedOnColor: " + responsePayload)
	//return shim.Success([]byte(responsePayload))
	return nil
}
