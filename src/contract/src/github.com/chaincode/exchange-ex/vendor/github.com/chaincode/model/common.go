package model


type Model interface {
	GetObjectType() string
}

type Entity struct {
	ObjectType string `json:"docType"`
}

func (e *Entity) GetObjectType() string {
	return e.ObjectType
}