package room

import "adonmo.com/goroom/logger"

//GoRoomSchemaMaster Tracks the schema of entities against current version of DB
type GoRoomSchemaMaster struct {
	Version      VersionNumber `gorm:"primary_key"`
	IdentityHash string
}

func (appDB *Room) isSchemaMasterPresent() bool {
	return appDB.orm.HasTable(&GoRoomSchemaMaster{})
}

func (appDB *Room) createSchemaMaster() {
	appDB.orm.CreateTable(&GoRoomSchemaMaster{})
}

func (appDB *Room) getRoomMetadataFromDB() (*GoRoomSchemaMaster, error) {
	identityHash, version, err := appDB.orm.GetLatestSchemaIdentityHashAndVersion()
	if err != nil {
		logger.Errorf("Error while fetching room metadata from the DB. %v", err)
		return nil, err
	}
	return &GoRoomSchemaMaster{
		IdentityHash: identityHash,
		Version:      VersionNumber(version),
	}, err
}
