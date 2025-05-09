package model

type Marble struct {
	ObjectType string `json:"docType"`
	Name       string `json:"name"`
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}
