package room

import (
	"fmt"
	"sort"

	"adonmo.com/goroom/logger"
	"adonmo.com/goroom/util/deephash"
	"github.com/jinzhu/gorm"
)

//VersionNumber Type for specifying version number across Room
type VersionNumber int

//Room Tracks the database objects, properties and configuration
type Room struct {
	Entities   []interface{}
	DBFilePath string
	Version    VersionNumber
	Migrations []Migration
	DB         *gorm.DB
}

//Migration Interface against users can define their migrations on the DB
type Migration interface {
	GetBaseVersion() VersionNumber
	GetTargetVersion() VersionNumber
	Apply() error
}

//GoRoomSchemaMaster Tracks the schema of entities against current version of DB
type GoRoomSchemaMaster struct {
	Version      VersionNumber `gorm:"primary_key"`
	IdentityHash string
}

//Init Initialize Room Database
func (room *Room) Init() error {
	sqliteDB, err := room.getSqliteDB()
	if err != nil {
		return err
	}
	if !room.isSchemaMasterPresent(sqliteDB) {
		logger.Info("No Room Schema Master Detected in existing SQL DB. Creating now..")
		err := room.runFirstTimeDBCreation(sqliteDB)
		if err != nil {
			logger.Errorf("Unable to Initialize Room. Unexpected Error. %v", err)
			return err
		}
		return nil
	}

	roomMetadata, err := room.getRoomMetadataFromDB(sqliteDB)
	if err != nil {
		logger.Error("Unable to fetch metadata although room master exists. This could be a sign of database corruption.")
		return err
	}
	currentIdentityHash, err := room.calculateIdentityHash(sqliteDB)
	if err != nil {
		logger.Errorf("Error while calculating signature of current Entity collection. %v", err)
		return err
	}

	if room.Version == roomMetadata.Version {
		err = room.peformDatabaseSanityChecks(currentIdentityHash, roomMetadata)
	} else if room.Version > roomMetadata.Version {
		err = room.performUpgrade(currentIdentityHash)
	} else {
		err = room.performDowngrade(currentIdentityHash)
	}

	return err
}

func (room *Room) getSqliteDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", room.DBFilePath)
	if err != nil {

		return nil, fmt.Errorf("Unable to open Database at the given file path %v", room.DBFilePath)
	}
	return db, nil
}

func (room *Room) isSchemaMasterPresent(db *gorm.DB) bool {
	return db.HasTable(&GoRoomSchemaMaster{})
}

func (room *Room) createSchemaMaster(db *gorm.DB) {
	db.CreateTable(&GoRoomSchemaMaster{})
}

func (room *Room) createEntities(db *gorm.DB) {
	for _, entity := range room.Entities {
		if !db.HasTable(entity) {
			db.CreateTable(entity)
		}
	}
}

func (room *Room) calculateIdentityHash(db *gorm.DB) (string, error) {
	var entityHashArr []string
	var sortedEntities []interface{}
	copy(sortedEntities, room.Entities)
	sort.Slice(room.Entities[:], func(i, j int) bool {
		modelA := db.NewScope(room.Entities[i]).GetModelStruct()
		modelB := db.NewScope(room.Entities[j]).GetModelStruct()

		return modelA.ModelType.Name() < modelB.ModelType.Name()
	})

	for _, entity := range sortedEntities {
		model := db.NewScope(entity).GetModelStruct()
		sum, err := deephash.ConstructHash(model)
		if err != nil {
			return "", fmt.Errorf("Error while calculating identity hash for Table %v", model.ModelType.Name())
		}
		entityHashArr = append(entityHashArr, sum)
	}

	identity, err := deephash.ConstructHash(entityHashArr)
	if err != nil {
		return "", fmt.Errorf("Error while calculating schema identity %v", entityHashArr)
	}

	return identity, nil
}

func (room *Room) runFirstTimeDBCreation(db *gorm.DB) error {
	identityHash, err := room.calculateIdentityHash(db)
	if err != nil {
		return err
	}
	room.createSchemaMaster(db)
	room.createEntities(db)

	metadata := GoRoomSchemaMaster{
		Version:      room.Version,
		IdentityHash: identityHash,
	}

	dbExec := db.Create(&metadata)
	if dbExec.Error != nil {
		logger.Errorf("Error while adding entity hash to Room Schema Master. %v", dbExec.Error)
		return dbExec.Error
	}

	return nil
}

func (room *Room) wipeOutExistingDB(db *gorm.DB) {
	for _, entity := range room.Entities {
		if db.HasTable(entity) {
			db.DropTable(entity)
		}
	}
}

func (room *Room) getRoomMetadataFromDB(db *gorm.DB) (*GoRoomSchemaMaster, error) {
	var roomMetadata GoRoomSchemaMaster
	dbExec := db.Order("version DESC").First(&roomMetadata)
	if dbExec.Error != nil {
		logger.Errorf("Error while fetching room metadata from the DB. %v", dbExec.Error)
		return nil, dbExec.Error
	}
	return &roomMetadata, nil
}

func (room *Room) peformDatabaseSanityChecks(currentIdentityHash string, roomMetadata *GoRoomSchemaMaster) error {
	if currentIdentityHash != roomMetadata.IdentityHash {
		logger.Error("Database Hash does not match. Looks like you changed entity definitions but forgot to upgrade version.")
		return fmt.Errorf("Database signature mismatch. Version %v", room.Version)
	}

	return nil
}

func (room *Room) performUpgrade(currentIdentityHash string) error {
	return nil
}

func (room *Room) performDowngrade(currentIdentityHash string) error {
	return nil
}
