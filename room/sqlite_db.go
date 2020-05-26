package room

import (
	"fmt"

	"adonmo.com/goroom/logger"
	"github.com/jinzhu/gorm"
)

func (room *Room) getSqliteDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", room.dbFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to open Database at the given file path %v", room.dbFilePath)
	}

	return db, nil
}

func (room *Room) runFirstTimeDBCreation() error {
	identityHash, err := room.calculateIdentityHash()
	if err != nil {
		return err
	}
	room.createSchemaMaster()
	room.createEntities()

	metadata := GoRoomSchemaMaster{
		Version:      room.version,
		IdentityHash: identityHash,
	}

	dbExec := room.db.Create(&metadata)
	if dbExec.Error != nil {
		logger.Errorf("Error while adding entity hash to Room Schema Master. %v", dbExec.Error)
		return dbExec.Error
	}

	return nil
}

func (room *Room) wipeOutExistingDB() {

	if room.isSchemaMasterPresent() {
		room.db.DropTable(GoRoomSchemaMaster{})
	}

	for _, entity := range room.entities {
		if room.db.HasTable(entity) {
			room.db.DropTable(entity)
		}
	}

	room.db = nil
}

func (room *Room) peformDatabaseSanityChecks(currentIdentityHash string, roomMetadata *GoRoomSchemaMaster) error {
	if currentIdentityHash != roomMetadata.IdentityHash {
		logger.Error("Database Hash does not match. Looks like you changed entity definitions but forgot to upgrade version.")
		return fmt.Errorf("Database signature mismatch. Version %v", room.version)
	}

	return nil
}
