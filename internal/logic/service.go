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
	err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	marbleToTransfer.Owner = newOwner //change the owner
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

	// Iterate through result set and for each marble found, transfer to newOwner
	var i int
	for i = 0; coloredMarbleResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
		responseRange, err := coloredMarbleResultsIterator.Next()
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}

		// get the color and name from color~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return fmt.Errorf("%s", err.Error())
		}
		returnedColor := compositeKeyParts[0]
		returnedMarbleName := compositeKeyParts[1]
		fmt.Printf("- found a marble from index:%s color:%s name:%s\n", objectType, returnedColor, returnedMarbleName)

		// Now call the transfer function for the found marble.
		// Re-use the same function that is used to transfer individual marbles
		//err = s.TransferMarble(stub, []string{returnedMarbleName, newOwner})
		err = s.TransferMarble(ctx, returnedMarbleName, newOwner)
		// if the transfer failed break out of loop and return error
		if err != nil {
			return fmt.Errorf("%s", "Transfer failed: "+err.Error())
		}
	}
	return nil
}

// func (s *Service) GetHistoryForMarble(ctx context.Context, marbleName string) (*bytes.Buffer, error) {
// ptrStub := ctx.Value("stub")
// if ptrStub == nil {
// 	return nil, fmt.Errorf("no 'stub' in context")
// }
// stub := ptrStub.(shim.ChaincodeStubInterface)
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
// return &buffer, nil
//}

func (s *Service) GetMarbleHistory(ctx context.Context, name string) (map[string]model.Marble, map[string]time.Time, error) {
	marbles, timestamsp, err := s.repository.Repository.GetMarbleHistory(ctx, name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to history for asset: %v", err)
	}

	return marbles, timestamsp, nil

}
