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
	result, err := room.db.QueryLatest(&roomMetadata, "version", "DESC")
	if err != nil {
		logger.Errorf("Error while fetching room metadata from the DB. %v", err)
		return nil, err
	}
	return result.(*GoRoomSchemaMaster), nil
}
