package room

import (
	"fmt"

	"github.com/adonmo/goroom/logger"
	"github.com/adonmo/goroom/orm"
)

func getFirstTimeDBCreationFunction(identityHash string, version orm.VersionNumber, entitiesToCreate []interface{}) func(orm.ORM) error {

	return func(dba orm.ORM) error {

		//Explicit Create without existence check. This ensures failure if this is not really a first time DB Creation
		if err := dba.CreateTable(GoRoomSchemaMaster{}).Error; err != nil {
			return err
		}

		for _, entity := range entitiesToCreate {
			if !dba.HasTable(entity) {
				if err := dba.CreateTable(entity).Error; err != nil {
					return err
				}
			}
		}

		metadata := GoRoomSchemaMaster{
			Version:      version,
			IdentityHash: identityHash,
		}

		dbExec := dba.Create(&metadata)
		if dbExec.Error != nil {
			logger.Errorf("Error while adding entity hash to Room Schema Master. %v", dbExec.Error)
			return dbExec.Error
		}

		return nil
	}
}

//GetDBCleanUpFunction Gives you a function that dictates the DB clean up transaction
func GetDBCleanUpFunction(entities []interface{}) func(orm.ORM) error {

	return func(dba orm.ORM) error {
		for _, entity := range entities {
			if dba.HasTable(entity) {
				if err := dba.DropTable(entity).Error; err != nil {
					return err
				}
			}
		}

		return nil
	}
}

//PerformDBCleanUp Cleans up existing DB removing Room metadata and all known entities
func (appDB *Room) PerformDBCleanUp() error {
	dbCleanUpFunc := GetDBCleanUpFunction(append(appDB.entities, GoRoomSchemaMaster{}))
	return appDB.dba.DoInTransaction(dbCleanUpFunc)
}

func (appDB *Room) peformDatabaseSanityChecks(currentIdentityHash string, roomMetadata *GoRoomSchemaMaster) error {
	if currentIdentityHash != roomMetadata.IdentityHash {
		logger.Error("Database Hash does not match. Looks like you changed entity definitions but forgot to upgrade version.")
		return fmt.Errorf("Database signature mismatch. Version %v", appDB.version)
	}

	return nil
}
