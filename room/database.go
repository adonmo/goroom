package room

import (
	"fmt"

	"adonmo.com/goroom/logger"
)

func (appDB *Room) runFirstTimeDBCreation() error {
	identityHash, err := appDB.calculateIdentityHash()
	if err != nil {
		return err
	}
	appDB.createSchemaMaster()
	appDB.createEntities()

	metadata := GoRoomSchemaMaster{
		Version:      appDB.version,
		IdentityHash: identityHash,
	}

	dbExec := appDB.dba.Create(&metadata)
	if dbExec.Error != nil {
		logger.Errorf("Error while adding entity hash to Room Schema Master. %v", dbExec.Error)
		return dbExec.Error
	}

	return nil
}

func (appDB *Room) wipeOutExistingDB() {

	if appDB.isSchemaMasterPresent() {
		appDB.dba.DropTable(GoRoomSchemaMaster{})
	}

	for _, entity := range appDB.entities {
		if appDB.dba.HasTable(entity) {
			appDB.dba.DropTable(entity)
		}
	}

	appDB.dba = nil
}

func (appDB *Room) peformDatabaseSanityChecks(currentIdentityHash string, roomMetadata *GoRoomSchemaMaster) error {
	if currentIdentityHash != roomMetadata.IdentityHash {
		logger.Error("Database Hash does not match. Looks like you changed entity definitions but forgot to upgrade version.")
		return fmt.Errorf("Database signature mismatch. Version %v", appDB.version)
	}

	return nil
}
