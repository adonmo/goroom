package room

import (
	"adonmo.com/goroom/logger"
	"github.com/jinzhu/gorm"
)

//VersionNumber Type for specifying version number across Room
type VersionNumber int

//Room Tracks the database objects, properties and configuration
type Room struct {
	entities                       []interface{}
	dbFilePath                     string
	version                        VersionNumber
	migrations                     []Migration
	fallbackToDestructiveMigration bool
	db                             *gorm.DB
}

//GetDB Returns Database object to be used by the application
func (room *Room) GetDB() (*gorm.DB, error) {
	sqliteDB, err := room.getSqliteDB()
	if err != nil {
		return nil, err
	}

	err = room.initRoomDB(sqliteDB)
	if err != nil && room.fallbackToDestructiveMigration {
		room.wipeOutExistingDB()
		err = room.initRoomDB(sqliteDB)
	}

	return room.db, err
}

//Init Initialize Room Database
func (room *Room) initRoomDB(db *gorm.DB) (err error) {
	defer func() {
		if err != nil {
			room.db = nil
		}
	}()

	room.db = db
	if !room.isSchemaMasterPresent() {
		logger.Info("No Room Schema Master Detected in existing SQL DB. Creating now..")
		err = room.runFirstTimeDBCreation()
		if err != nil {
			logger.Errorf("Unable to Initialize Room. Unexpected Error. %v", err)
			return err
		}
		return nil
	}

	roomMetadata, err := room.getRoomMetadataFromDB()
	if err != nil {
		logger.Error("Unable to fetch metadata although room master exists. This could be a sign of database corruption.")
		return err
	}
	currentIdentityHash, err := room.calculateIdentityHash()
	if err != nil {
		logger.Errorf("Error while calculating signature of current Entity collection. %v", err)
		return err
	}

	applicableMigrations, err := GetApplicableMigrations(room.migrations, roomMetadata.Version, room.version)
	if err != nil {
		return err
	}

	if room.version == roomMetadata.Version {
		err = room.peformDatabaseSanityChecks(currentIdentityHash, roomMetadata)
	} else {
		room.performMigrations(currentIdentityHash, applicableMigrations)
	}

	return err
}
