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

func (item MarbleHistoryResponse) FromModel(assets map[string]Marble, timestamps map[string]time.Time) {
	for k, a := range assets {
		//var marbleResponse Marble
		//marbleResponse.FromModel(&a)
		item[k] = MarbleHistoryResponseItem{
			TxId:      k,
			Value:     a,
			Timestamp: timestamps[k],
			IsDelete:  a.Name == "",
		}
	}
}

// func (item *GetMarbleResponse) FromModel(in *Marble) {
// 	/*
// 		item.ID = in.ID
// 		item.AppraisedValue = in.AppraisedValue
// 		item.Color = in.Color
// 		item.Size = in.Size
// 		item.Owner = in.Owner
// 	*/

// 	*item = GetAssetResponse{
// 		ID:             in.ID,
// 		AppraisedValue: in.AppraisedValue,
// 		Color:          in.Color,
// 		Size:           0,
// 		Owner:          in.Owner,
// 	}
// }
