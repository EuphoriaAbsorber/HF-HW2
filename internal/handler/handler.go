package handler

import (
	"context"
	"encoding/json"
	"fmt"
	logic "main/internal/logic"
	"main/internal/model"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Service struct {
	contractapi.Contract

	bl *logic.Service
}

func NewService(bl *logic.Service) *Service {
	return &Service{
		bl: bl,
	}
}

func (s *Service) InitMarble(hlfContext contractapi.TransactionContextInterface, name string, color string, size int, owner string) error {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)

	_, err := hlfContext.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return fmt.Errorf("failed to get x509 certificate: %v", err)
	}

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
	objectType := "marble"
	marble := &model.Marble{ObjectType: objectType, Name: marbleName, Color: color, Size: size, Owner: owner}
	err = s.bl.CreateMarble(ctx, marble)
	if err != nil {
		return fmt.Errorf("could not create marble")
	}
	fmt.Println("- end init marble")
	return nil
}

func (s *Service) ReadMarble(hlfContext contractapi.TransactionContextInterface, name string) (*model.Marble, error) {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)

	_, err := hlfContext.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return nil, fmt.Errorf("failed to get x509 certificate: %v", err)
	}

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

	_, err := hlfContext.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return fmt.Errorf("failed to get x509 certificate: %v", err)
	}

	valAsbytes, err := stub.GetState(marbleName)
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
	_, err := hlfContext.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return fmt.Errorf("failed to get x509 certificate: %v", err)
	}
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
	_, err := hlfContext.GetClientIdentity().GetX509Certificate()
	if err != nil {
		return fmt.Errorf("failed to get x509 certificate: %v", err)
	}
	newOwner = strings.ToLower(newOwner)
	fmt.Println("- start transferMarblesBasedOnColor ", color, newOwner)
	s.bl.TransferMarblesBasedOnColor(ctx, color, newOwner)

	responsePayload := fmt.Sprintf("Transferred  %s marbles to %s", color, newOwner)
	fmt.Println("- end transferMarblesBasedOnColor: " + responsePayload)
	return nil
}

func (s *Service) GetHistoryForMarble(hlfContext contractapi.TransactionContextInterface, marbleName string) (model.MarbleHistoryResponse, error) {
	stub := hlfContext.GetStub()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "stub", stub)

	fmt.Printf("- start getHistoryForMarble: %s\n", marbleName)

	marbles, timestamps, err := s.bl.GetMarbleHistory(ctx, marbleName)
	if err != nil {
		return nil, fmt.Errorf("%s", "Failed to get marble history: "+err.Error())
	}

	result := make(model.MarbleHistoryResponse)

	result.FromModel(marbles, timestamps)

	return result, nil

}
