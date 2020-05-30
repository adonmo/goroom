package migrations

import (
	"fmt"

	"adonmo.com/goroom/example/models/latest"
	"adonmo.com/goroom/example/models/old"
	"adonmo.com/goroom/orm"
	"github.com/jinzhu/gorm"
)

//UserDBMigration Represents migration objects used for the example DB
type UserDBMigration struct {
	BaseVersion   orm.VersionNumber
	TargetVersion orm.VersionNumber
	MigrationFunc func(db interface{}) error
}

//GetBaseVersion ...
func (m *UserDBMigration) GetBaseVersion() orm.VersionNumber {
	return m.BaseVersion
}

//GetTargetVersion ...
func (m *UserDBMigration) GetTargetVersion() orm.VersionNumber {
	return m.TargetVersion
}

//Apply ....
func (m *UserDBMigration) Apply(db interface{}) error {
	return m.MigrationFunc(db)
}

//GetMigrations Returns migrations applicable to the given DB over various version transistions
func GetMigrations() (migrations []orm.Migration) {

	migration12 := &UserDBMigration{
		BaseVersion:   1,
		TargetVersion: 2,
		MigrationFunc: func(db interface{}) error {
			gormDB, ok := db.(*gorm.DB)
			if !ok {
				return fmt.Errorf("Unable to get the desired DB object")
			}

			return gormDB.CreateTable(old.Profile{}).Error
		},
	}

	migration23 := &UserDBMigration{
		BaseVersion:   2,
		TargetVersion: 3,
		MigrationFunc: func(db interface{}) error {
			gormDB, ok := db.(*gorm.DB)
			if !ok {
				return fmt.Errorf("Unable to get the desired DB object")
			}

			return gormDB.AutoMigrate(latest.User{}).Error
		},
	}

	var migration34 = &UserDBMigration{
		BaseVersion:   3,
		TargetVersion: 4,
		MigrationFunc: func(db interface{}) error {
			gormDB, ok := db.(*gorm.DB)
			if !ok {
				return fmt.Errorf("Unable to get the desired DB object")
			}

			return gormDB.AutoMigrate(latest.Profile{}).Error
		},
	}

	migrations = append(migrations, migration12, migration23, migration34)
	return
}
