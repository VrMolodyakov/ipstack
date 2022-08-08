package entity

type User struct {
	Id       int    `json:"id"`
	Nickname string `json:"nickname"`
}

type UserIdDto struct {
	Id int `json:"id"`
}
