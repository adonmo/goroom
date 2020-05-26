package room

import (
	"fmt"

	"adonmo.com/goroom/logger"
	"adonmo.com/goroom/room/orm"
)

//VersionNumber Type for specifying version number across Room
type VersionNumber uint

//IdentityHashCalculator Calculates Identity based on the entity model definition returned by ORM
type IdentityHashCalculator interface {
	ConstructHash(entityModel interface{}) (ans string, err error)
}

//Room Tracks the database objects, properties and configuration
type Room struct {
	entities                       []interface{}
	version                        VersionNumber
	migrations                     []Migration
	fallbackToDestructiveMigration bool
	orm                            orm.ORM
	identityCalculator             IdentityHashCalculator
}

//New Returns a new room struct that can be used to initialize and get a DB managed by room
func New(entities []interface{}, orm orm.ORM, version VersionNumber,
	migrations []Migration, fallbackToDestructiveMigration bool, identityCalculator IdentityHashCalculator) (room *Room, errors []error) {

	if len(entities) < 1 {
		errors = append(errors, fmt.Errorf("No entities provided for the database"))
	}
	if orm == nil {
		errors = append(errors, fmt.Errorf("Need an ORM to work with"))
	}
	if version < 1 {
		errors = append(errors, fmt.Errorf("Only non zero versions allowed"))
	}
	if identityCalculator == nil {
		errors = append(errors, fmt.Errorf("Need an identity calculator"))
	}

	if len(errors) < 1 {
		room = &Room{
			entities:                       entities,
			version:                        version,
			migrations:                     migrations,
			fallbackToDestructiveMigration: fallbackToDestructiveMigration,
			orm:                            orm,
			identityCalculator:             identityCalculator,
		}
	}

	return
}

//InitializeAppDB Returns Database object to be used by the application
func (appDB *Room) InitializeAppDB() error {
	err := appDB.initRoomDB()
	if err != nil && appDB.fallbackToDestructiveMigration {
		appDB.wipeOutExistingDB()
		err = appDB.initRoomDB()
	}

	return err
}

//Init Initialize Room Database
func (appDB *Room) initRoomDB() (err error) {
	defer func() {
		if err != nil {
			appDB.orm = nil
		}
	}()

	if !appDB.isSchemaMasterPresent() {
		logger.Info("No Room Schema Master Detected in existing SQL DB. Creating now..")
		err = appDB.runFirstTimeDBCreation()
		if err != nil {
			logger.Errorf("Unable to Initialize Room. Unexpected Error. %v", err)
			return err
		}
		return nil
	}

	roomMetadata, err := appDB.getRoomMetadataFromDB()
	if err != nil {
		logger.Error("Unable to fetch metadata although room master exists. This could be a sign of database corruption.")
		return err
	}
	currentIdentityHash, err := appDB.calculateIdentityHash()
	if err != nil {
		logger.Errorf("Error while calculating signature of current Entity collection. %v", err)
		return err
	}

	applicableMigrations, err := GetApplicableMigrations(appDB.migrations, roomMetadata.Version, appDB.version)
	if err != nil {
		return err
	}

	if appDB.version == roomMetadata.Version {
		err = appDB.peformDatabaseSanityChecks(currentIdentityHash, roomMetadata)
	} else {
		appDB.performMigrations(currentIdentityHash, applicableMigrations)
	}

	return err
}
