package room

import (
	"fmt"

	"adonmo.com/goroom/logger"
	"adonmo.com/goroom/orm"
)

//Room Tracks the database objects, properties and configuration
type Room struct {
	entities                       []interface{}
	version                        orm.VersionNumber
	migrations                     []orm.Migration
	fallbackToDestructiveMigration bool
	dba                            orm.ORM
	identityCalculator             orm.IdentityHashCalculator
}

//New Returns a new room struct that can be used to initialize and get a DB managed by room
func New(entities []interface{}, dba orm.ORM, version orm.VersionNumber,
	migrations []orm.Migration, fallbackToDestructiveMigration bool, identityCalculator orm.IdentityHashCalculator) (room *Room, errors []error) {

	if len(entities) < 1 {
		errors = append(errors, fmt.Errorf("No entities provided for the database"))
	}
	if dba == nil {
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
			dba:                            dba,
			identityCalculator:             identityCalculator,
		}
	}

	return
}

/* Initialization Scenarios In Brief:
Scenario 1:
	Trigger: 	No Schema Master Present.
	Action:		Room creates Schema Master and any entity tables that are not there already.
	Gotcha:		Pre Existing Tables are assumed to have schema same as current version

Scenario 2:
	Trigger: 	Schema Master Present and Version is same.
	Action:		Room verfies integrity by comparing current and saved hash. Triggers Error if not equal.
	Gotcha: 	Schema Master is assumed to have latest(that is last known) version record stored.

Scenario 3:
	Trigger:	Schema Master Present and Version is different
	Action: 	Room triggers migration. Triggers error if migration fails
	Gotcha: 	An Empty migration must be specified even if no database action(like altering tables etc) is required for version change.

If the initialization fails for any reason in any of the three scenarios then we check for destructive migration option.
If enabled whole DB(Schema Master and known entities) is wiped out and init is retried
*/

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
			appDB.dba = nil
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
		err = appDB.performMigrations(currentIdentityHash, applicableMigrations)
	}

	return err
}
