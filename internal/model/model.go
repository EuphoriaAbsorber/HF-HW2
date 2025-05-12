package model

import "time"

type Marble struct {
	ObjectType string `json:"docType"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

type MarbleHistoryResponseItem struct {
	TxId      string    `json:"tx_id,omitempty"`
	Value     Marble    `json:"value,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	IsDelete  bool      `json:"is_delete"`
}

type MarbleHistoryResponse map[string]MarbleHistoryResponseItem

func (item MarbleHistoryResponse) FromModel(marbles map[string]Marble, timestamps map[string]time.Time) {
	for k, a := range marbles {
		item[k] = MarbleHistoryResponseItem{
			TxId:      k,
			Value:     a,
			Timestamp: timestamps[k],
			IsDelete:  a.Name == "",
		}
	}
}
