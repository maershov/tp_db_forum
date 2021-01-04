package models

//easyjson:json
type Vote struct {
	Nickname string `json:"nickname"`
	Thread   string `json:"-"`
	Voice    int    `json:"voice"`
}
