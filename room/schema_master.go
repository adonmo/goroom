package room

import (
	"github.com/adonmo/goroom/logger"
	"github.com/adonmo/goroom/orm"
)

//GoRoomSchemaMaster Tracks the schema of entities against current version of DB
type GoRoomSchemaMaster struct {
	Version      orm.VersionNumber `gorm:"primary_key"`
	IdentityHash string
}

func (appDB *Room) isSchemaMasterPresent() bool {
	return appDB.dba.HasTable(GoRoomSchemaMaster{})
}

func (appDB *Room) getRoomMetadataFromDB() (*GoRoomSchemaMaster, error) {
	identityHash, version, err := appDB.dba.GetLatestSchemaIdentityHashAndVersion()
	if err != nil {
		logger.Errorf("Error while fetching room metadata from the DB. %v", err)
		return nil, err
	}
	return &GoRoomSchemaMaster{
		IdentityHash: identityHash,
		Version:      orm.VersionNumber(version),
	}, err
}
