package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/model"
	rep "main/internal/repository"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type Service struct {
	repository *rep.Service
}

func NewService(repository *rep.Service) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateMarble(ctx context.Context, marble *model.Marble) error {
	err := s.repository.Repository.CreateMarble(ctx, marble)
	if err != nil {
		return fmt.Errorf("failed to create marble: %v", err)
	}

	return nil
}

func (s *Service) ReadMarble(ctx context.Context, name string) ([]byte, error) {
	return s.repository.Repository.ReadMarble(ctx, name)
}

func (s *Service) DeleteMarble(ctx context.Context, marbleJSON model.Marble) error {
	return s.repository.Repository.DeleteMarble(ctx, marbleJSON)
}

func (s *Service) TransferMarble(ctx context.Context, marbleName string, newOwner string) error {
	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return fmt.Errorf("no 'stub' in context")
	}
	stub := ptrStub.(shim.ChaincodeStubInterface)
	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return fmt.Errorf("%s", "Failed to get marble:"+err.Error())
	} else if marbleAsBytes == nil {
		return fmt.Errorf("%s", "Marble does not exist")
	}

	marbleToTransfer := model.Marble{}
	err = json.Unmarshal(marbleAsBytes, &marbleToTransfer)
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	marbleToTransfer.Owner = newOwner
	return s.repository.Repository.TransferMarble(ctx, marbleToTransfer)
}

func (s *Service) TransferMarblesBasedOnColor(ctx context.Context, color string, newOwner string) error {
	ptrStub := ctx.Value("stub")
	if ptrStub == nil {
		return fmt.Errorf("no 'stub' in context")
	}
	stub := ptrStub.(shim.ChaincodeStubInterface)
	coloredMarbleResultsIterator, err := stub.GetStateByPartialCompositeKey("color~name", []string{color})
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	defer coloredMarbleResultsIterator.Close()

	var i int
	for i = 0; coloredMarbleResultsIterator.HasNext(); i++ {
		responseRange, err := coloredMarbleResultsIterator.Next()
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}
		returnedColor := compositeKeyParts[0]
		returnedMarbleName := compositeKeyParts[1]
		fmt.Printf("- found a marble from index:%s color:%s name:%s\n", objectType, returnedColor, returnedMarbleName)
		err = s.TransferMarble(ctx, returnedMarbleName, newOwner)
		if err != nil {
			return fmt.Errorf("%s", "Transfer failed: "+err.Error())
		}
	}
	return nil
}

func (s *Service) GetMarbleHistory(ctx context.Context, name string) (map[string]model.Marble, map[string]time.Time, error) {
	marbles, timestamsp, err := s.repository.Repository.GetMarbleHistory(ctx, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to history for asset: %v", err)
	}

	return marbles, timestamsp, nil

}
