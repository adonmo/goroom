package example

import (
	"fmt"
	"os"
	"testing"

	groom "adonmo.com/goroom"
	"adonmo.com/goroom/example/migrations"
	"adonmo.com/goroom/example/models/latest"
	"adonmo.com/goroom/example/models/old"
	"adonmo.com/goroom/logger"
	"adonmo.com/goroom/orm"
	"adonmo.com/goroom/room"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"adonmo.com/goroom/util/adapter"
)

/* A sample to demonstrate the usage of go room

The primary function of Room is to ease version management and migration of embedded Data Stores.
Embedded Data Stores are databases that are created by apps on the edge devices and are tightly coupled with the
version of app that creates and manages them.

Using Room a developer can ensure that as they deliver updates to App and underlying associated Data Store they can
have a smooth transition of Data Store on the edge device without risk of data loss.

A typical way to handle upgrades would be to create the data store from scratch again. This is not desirable if the
data stored on the edge device is things like valuable insights/events recorded by the device and pending sync/upload to the
server. Since data collection is a major use case of edge devices I think a version manager like Room is necessary.

Room is inspired by its namesake in Android World which does the same thing but at a deeper level by even providing the ORM.
The Room presented here is agnostic to data stores and provides flexibility to the developer on how they signal(Check IdentityHashCalculator Interface) and handle schema changes(Check Migration Interface).

It is purely a utility that serves the minimal purpose of carrying out migrations and verifying that DB is upto the version expected by the app currently.
A lot of power is still in the developers hands as they have the freedom to execute any operations on the DB themselves.
Although doing stuff like deleting/updating Room's metadata tables is a big No-No :). Plz...
*/
func TestIntegrationWithGORM(t *testing.T) {

	//A Data Store is represented by the tables(entities) it houses. Below we will define a snapshot each of a DB in various versions.

	entitiesForVersionsArr := [][]interface{}{}
	entitiesForVersionsArr = append(entitiesForVersionsArr, []interface{}{old.User{}})                      //First Data Store Version with just User Table
	entitiesForVersionsArr = append(entitiesForVersionsArr, []interface{}{old.User{}, old.Profile{}})       //Profile Table Added
	entitiesForVersionsArr = append(entitiesForVersionsArr, []interface{}{latest.User{}, old.Profile{}})    //User Table upgraded to have a new column for Credits
	entitiesForVersionsArr = append(entitiesForVersionsArr, []interface{}{latest.User{}, latest.Profile{}}) //Profile Table upgraded to have foreign key relationship with User

	if !verifyThatEntityHashesForAllVersionsAreDifferent(entitiesForVersionsArr) {
		t.Errorf("Hash Uniqueness check failed")
	}

	if !verifyThatMigrationWorksForEachCombinationOfSourceAndTargetVersion(entitiesForVersionsArr) {
		t.Errorf("Migration testing has failed")
	}
}

func verifyThatMigrationWorksForEachCombinationOfSourceAndTargetVersion(entitiesForVersionsArr [][]interface{}) bool {

	applicableMigrations := migrations.GetMigrations()

	for srcIdx, oldEntities := range entitiesForVersionsArr {

		dbFilePath := "test_goroom.db"
		var err = os.Remove(dbFilePath)
		if err != nil && !os.IsNotExist(err) {
			fmt.Print(err)
			panic(err)
		}

		srcVersionNumber := orm.VersionNumber(srcIdx + 1)
		db, gormAdapter := getDBAndGORMAdapter(dbFilePath)
		identityCalculator := new(adapter.EntityHashConstructor)

		appDB, errList := room.New(oldEntities, gormAdapter, srcVersionNumber, applicableMigrations, identityCalculator)
		if len(errList) > 0 {
			panic(errList)
		}

		err = groom.InitializeRoom(appDB, false)
		if err != nil {
			panic(fmt.Errorf("Error while init for Version %v", srcVersionNumber))
		}
		db.Close()

		for i := srcIdx + 1; i < len(entitiesForVersionsArr); i++ {
			currentVersionNumber := orm.VersionNumber(i + 1)
			db, gormAdapter = getDBAndGORMAdapter(dbFilePath)
			appDB, errList = room.New(entitiesForVersionsArr[i], gormAdapter, currentVersionNumber, applicableMigrations, identityCalculator)

			if len(errList) > 0 {
				panic(errList)
			}

			err = groom.InitializeRoom(appDB, false)
			if err != nil {
				panic(fmt.Errorf("Error while init for Version %v", currentVersionNumber))
			}
			db.Close()
		}

	}

	return true
}

func getDBAndGORMAdapter(dbFilePath string) (*gorm.DB, orm.ORM) {
	db, err := gorm.Open("sqlite3", dbFilePath)
	if err != nil {
		panic(err)
	}

	gormAdapter := adapter.NewGORM(db)
	return db, gormAdapter
}

func verifyThatEntityHashesForAllVersionsAreDifferent(entitiesForVersionsArr [][]interface{}) bool {

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	defer func() {
		db.Close()
	}()

	identityToVersion := make(map[string]orm.VersionNumber)
	gormAdapter := adapter.NewGORM(db)
	identityCalculator := new(adapter.EntityHashConstructor)

	for idx, entities := range entitiesForVersionsArr {

		currentVersionNumber := orm.VersionNumber(idx + 1)
		//At this point we are just constructing the Room object and not really initializing the DB hence we can reuse the same DB connection and adapter safely
		appDB, errList := room.New(entities, gormAdapter, currentVersionNumber, []orm.Migration{}, identityCalculator)
		if len(errList) > 0 {
			panic(errList)
		}
		identity, err := appDB.CalculateIdentityHash()
		fmt.Printf("Version %v. Hash %v\n", currentVersionNumber, identity)
		if err != nil {
			panic(err)
		}

		if v, ok := identityToVersion[identity]; ok {
			logger.Errorf("An Older Version %v has same signature as this one %v. In current arrangment this is not expected. Remove this check if the example needs so now", v, currentVersionNumber)
			return false
		}

		identityToVersion[identity] = currentVersionNumber
	}

	return true
}
