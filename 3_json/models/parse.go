// models/models.go
package models

//easyjson:json
type User struct {
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Browser []string `json:"browsers"`
}

//easyjson:json
type UserList []User
