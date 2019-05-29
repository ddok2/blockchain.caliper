package model


const MemberObjectType = "Member"


type Member struct {
	Entity
	ID		string `json:"id"`
	Name	string `json:"name"`
}