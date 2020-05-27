package adapter

import (
	"testing"

	"adonmo.com/goroom/room"
	"github.com/go-test/deep"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	DB      *gorm.DB
	Adapter room.ORM
}

type DummyTable struct {
	ID    int `gorm:"primary_key"`
	Value string
}

type AnotherDummyTable struct {
	Num  int
	Text string
}

func (suite *IntegrationTestSuite) SetupTest() {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	suite.DB = db
	suite.Adapter = NewGORM(db)
}

func (suite *IntegrationTestSuite) TearDownTest() {
	err := suite.DB.Close()
	if err != nil {
		panic(err)
	}
	suite.DB = nil
}

func (suite *IntegrationTestSuite) TestNewGORM() {
	expected := &GORMAdapter{
		db: suite.DB,
	}
	got := NewGORM(suite.DB)
	diff := deep.Equal(expected, got)
	if diff != nil {
		suite.T().Errorf("Wrong result from New method for GORM adapter. Diff: %v", diff)
	}
}

func (suite *IntegrationTestSuite) TestHasTable() {
	dbExec := suite.DB.CreateTable(DummyTable{})
	if dbExec.Error != nil {
		panic(dbExec.Error)
	}

	if !suite.Adapter.HasTable(DummyTable{}) {
		suite.T().Errorf("Table Detection not wokring as expected.")
	}
}

func (suite *IntegrationTestSuite) TestCreateTable() {
	result := suite.Adapter.CreateTable(DummyTable{}, AnotherDummyTable{})

	print(result.Error)

	if result.Error != nil || !suite.Adapter.HasTable(DummyTable{}) || !suite.Adapter.HasTable(AnotherDummyTable{}) {
		suite.T().Errorf("Table Creation not working as expected. Error: %v", result.Error)
	}
}

func (suite *IntegrationTestSuite) TestCreate() {
	suite.Adapter.CreateTable(DummyTable{})

	dummyEntry := DummyTable{
		ID:    2,
		Value: "Two",
	}
	suite.Adapter.Create(&dummyEntry)
	var queryResult DummyTable
	suite.DB.Where("id = ?", dummyEntry.ID).First(&queryResult) //Where()

	diff := deep.Equal(dummyEntry, queryResult)
	if diff != nil {
		suite.T().Errorf("Create is not working as expected. Diff: %v", diff)
	}

}

func (suite *IntegrationTestSuite) TestTruncateTable() {
	suite.Adapter.CreateTable(DummyTable{})

	dummyEntry := DummyTable{
		ID:    2,
		Value: "Two",
	}
	anotherDummyEntry := DummyTable{
		ID:    3,
		Value: "Three",
	}
	suite.Adapter.Create(&dummyEntry)
	suite.Adapter.Create(&anotherDummyEntry)

	suite.Adapter.TruncateTable(DummyTable{})

	var queryResult DummyTable
	suite.DB.First(&queryResult)

	diff := deep.Equal(DummyTable{}, queryResult)
	if diff != nil {
		suite.T().Errorf("Truncate is not working as expected. Diff: %v", diff)
	}

}

func (suite *IntegrationTestSuite) TestDropTable() {
	suite.Adapter.CreateTable(DummyTable{})
	suite.Adapter.CreateTable(AnotherDummyTable{})
	suite.Adapter.DropTable(DummyTable{}, AnotherDummyTable{})

	if suite.DB.HasTable(DummyTable{}) || suite.DB.HasTable(AnotherDummyTable{}) {
		suite.T().Errorf("Drop Table not working as expected")
	}

}

func (suite *IntegrationTestSuite) TestGetModelDefinition() {
	expectedModel := suite.DB.NewScope(DummyTable{}).GetModelStruct()
	expectedOutput := room.ModelDefinition{
		EntityModel: expectedModel,
		TableName:   expectedModel.TableName(suite.DB),
	}

	got := suite.Adapter.GetModelDefinition(DummyTable{})
	diff := deep.Equal(expectedOutput, got)

	if diff != nil {
		suite.T().Errorf("GetModelDefinition not wokring as expected. %v", diff)
	}
}

func (suite *IntegrationTestSuite) TestGetUnderlyingORM() {
	diff := deep.Equal(suite.DB, suite.Adapter.GetUnderlyingORM())
	if diff != nil {
		suite.T().Errorf("Underlying ORM found to be different than expected. Diff: %v", diff)
	}
}

func (suite *IntegrationTestSuite) TestGetLatestSchemaIdentityHashAndVersion() {
	suite.Adapter.CreateTable(room.GoRoomSchemaMaster{})
	dummyEntry := room.GoRoomSchemaMaster{
		IdentityHash: "adaghsghas",
		Version:      room.VersionNumber(23),
	}
	anotherDummyEntry := room.GoRoomSchemaMaster{
		IdentityHash: "eyryhyeue",
		Version:      room.VersionNumber(24),
	}
	suite.Adapter.Create(&dummyEntry)
	suite.Adapter.Create(&anotherDummyEntry)

	identity, version, err := suite.Adapter.GetLatestSchemaIdentityHashAndVersion()
	queryResult := room.GoRoomSchemaMaster{
		IdentityHash: identity,
		Version:      room.VersionNumber(version),
	}

	if err != nil {
		suite.T().Errorf("No error expected when querying schema master for latest record. Got: %v", err)
	}

	diff := deep.Equal(anotherDummyEntry, queryResult)
	if diff != nil {
		suite.T().Errorf("Query Latest not working as expected. Diff: %v", diff)
	}

}

func TestMain(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
