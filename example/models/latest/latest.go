package latest

import "github.com/jinzhu/gorm"

//User User Entity
type User struct {
	gorm.Model
	Name    string
	Credits int
}

//Profile `Profile` belongs to `User`, `UserID` is the foreign key
type Profile struct {
	gorm.Model
	UserID int
	User   User `gorm:"foreignkey:UserRefer"`
	Name   string
}
