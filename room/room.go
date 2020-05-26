package room

import (
	"fmt"

	"adonmo.com/goroom/logger"
	"adonmo.com/goroom/room/orm"
	"github.com/jinzhu/gorm"
)

//VersionNumber Type for specifying version number across Room
type VersionNumber uint

//Room Tracks the database objects, properties and configuration
type Room struct {
	entities                       []interface{}
	dbFilePath                     string
	version                        VersionNumber
	migrations                     []Migration
	fallbackToDestructiveMigration bool
	orm                            orm.ORM
}

//New Returns a new room struct that can be used to initialize and get a DB managed by room
func New(entities []interface{}, dbFilePath string, version VersionNumber, migrations []Migration, fallbackToDestructiveMigration bool) (room *Room, errors []error) {
	if len(entities) < 1 {
		errors = append(errors, fmt.Errorf("No entities provided for the database"))
	}
	if dbFilePath == "" {
		errors = append(errors, fmt.Errorf("File path for DB missing"))
	}
	if version < 1 {
		errors = append(errors, fmt.Errorf("Only non zero versions allowed"))
	}

	if len(errors) < 1 {
		room = &Room{
			entities:                       entities,
			dbFilePath:                     dbFilePath,
			version:                        version,
			migrations:                     migrations,
			fallbackToDestructiveMigration: fallbackToDestructiveMigration,
		}
	}

	return
}

//GetDB Returns Database object to be used by the application
func (room *Room) GetDB() (*gorm.DB, error) {
	sqliteDB, err := room.getSqliteDB()
	if err != nil {
		return nil, err
	}

	ormAdapter := orm.NewGORM(sqliteDB)
	err = room.initRoomDB(ormAdapter)
	if err != nil && room.fallbackToDestructiveMigration {
		room.wipeOutExistingDB()
		err = room.initRoomDB(ormAdapter)
	}

	return room.orm.GetUnderlyingORM().(*gorm.DB), err
}

//Init Initialize Room Database
func (room *Room) initRoomDB(orm orm.ORM) (err error) {
	defer func() {
		if err != nil {
			room.orm = nil
		}
	}()

	room.orm = orm
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
