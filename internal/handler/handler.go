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

func (s *Service) InitMarble(hlfContext contractapi.TransactionContextInterface, name string, color string, size int, owner string) error {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)

	color = strings.ToLower(color)
	owner = strings.ToLower(owner)
	marbleName := name

	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return fmt.Errorf("%s", "Failed to get marble: "+err.Error())
	} else if marbleAsBytes != nil {
		fmt.Println("This marble already exists: " + marbleName)
		return fmt.Errorf("%s", "This marble already exists: "+marbleName)
	}

	// ==== Create marble object and marshal to JSON ====
	objectType := "marble"
	marble := &model.Marble{ObjectType: objectType, Name: marbleName, Color: color, Size: size, Owner: owner}
	err = s.bl.CreateMarble(ctx, marble)
	if err != nil {
		return fmt.Errorf("could not create marble")
	}
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
	newOwner = strings.ToLower(newOwner)
	fmt.Println("- start transferMarble ", marbleName, newOwner)
	s.bl.TransferMarble(ctx, marbleName, newOwner)
	fmt.Println("- end transferMarble (success)")
	return nil
}

func (s *Service) TransferMarblesBasedOnColor(hlfContext contractapi.TransactionContextInterface, color string, newOwner string) error {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)
	newOwner = strings.ToLower(newOwner)
	fmt.Println("- start transferMarblesBasedOnColor ", color, newOwner)
	s.bl.TransferMarblesBasedOnColor(ctx, color, newOwner)

	responsePayload := fmt.Sprintf("Transferred  %s marbles to %s", color, newOwner)
	fmt.Println("- end transferMarblesBasedOnColor: " + responsePayload)
	return nil
}

func (s *Service) GetHistoryForMarble(hlfContext contractapi.TransactionContextInterface, marbleName string) ([]byte, error) {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)

	fmt.Printf("- start getHistoryForMarble: %s\n", marbleName)

	// resultsIterator, err := stub.GetHistoryForKey(marbleName)
	// if err != nil {
	// 	return nil, fmt.Errorf("%s", err.Error())
	// }
	// defer resultsIterator.Close()

	// // buffer is a JSON array containing historic values for the marble
	// var buffer bytes.Buffer
	// buffer.WriteString("[")

	// bArrayMemberAlreadyWritten := false
	// for resultsIterator.HasNext() {
	// 	response, err := resultsIterator.Next()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("%s", err.Error())
	// 	}
	// 	// Add a comma before array members, suppress it for the first array member
	// 	if bArrayMemberAlreadyWritten == true {
	// 		buffer.WriteString(",")
	// 	}
	// 	buffer.WriteString("{\"TxId\":")
	// 	buffer.WriteString("\"")
	// 	buffer.WriteString(response.TxId)
	// 	buffer.WriteString("\"")

	// 	buffer.WriteString(", \"Value\":")
	// 	// if it was a delete operation on given key, then we need to set the
	// 	//corresponding value null. Else, we will write the response.Value
	// 	//as-is (as the Value itself a JSON marble)
	// 	if response.IsDelete {
	// 		buffer.WriteString("null")
	// 	} else {
	// 		buffer.WriteString(string(response.Value))
	// 	}

	// 	buffer.WriteString(", \"Timestamp\":")
	// 	buffer.WriteString("\"")
	// 	buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
	// 	buffer.WriteString("\"")

	// 	buffer.WriteString(", \"IsDelete\":")
	// 	buffer.WriteString("\"")
	// 	buffer.WriteString(strconv.FormatBool(response.IsDelete))
	// 	buffer.WriteString("\"")

	// 	buffer.WriteString("}")
	// 	bArrayMemberAlreadyWritten = true
	// }
	// buffer.WriteString("]")

	buffer, err := s.bl.GetHistoryForMarble(ctx, marbleName)
	if err != nil {
		//fmt.Printf("- getHistoryForMarble returning:\n%s\n", buffer.String())
		fmt.Errorf("%s", "Failed to get marble history: "+err.Error())
		return nil, err
	}
	fmt.Printf("- getHistoryForMarble returning:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
