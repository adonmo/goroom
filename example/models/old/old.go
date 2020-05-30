package old

import "github.com/jinzhu/gorm"

//User User Entity
type User struct {
	gorm.Model
	Name string
}

//Profile `Profile` belongs to `User`, `UserID` is the foreign key
type Profile struct {
	gorm.Model
	UserID int
	User   User
	Name   string
}
