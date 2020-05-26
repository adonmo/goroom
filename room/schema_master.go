package room

import "adonmo.com/goroom/logger"

//GoRoomSchemaMaster Tracks the schema of entities against current version of DB
type GoRoomSchemaMaster struct {
	Version      VersionNumber `gorm:"primary_key"`
	IdentityHash string
}

func (room *Room) isSchemaMasterPresent() bool {
	return room.db.HasTable(&GoRoomSchemaMaster{})
}

func (room *Room) createSchemaMaster() {
	room.db.CreateTable(&GoRoomSchemaMaster{})
}

func (room *Room) getRoomMetadataFromDB() (*GoRoomSchemaMaster, error) {
	var roomMetadata GoRoomSchemaMaster
	dbExec := room.db.Order("version DESC").First(&roomMetadata)
	if dbExec.Error != nil {
		logger.Errorf("Error while fetching room metadata from the DB. %v", dbExec.Error)
		return nil, dbExec.Error
	}
	return &roomMetadata, nil
}
