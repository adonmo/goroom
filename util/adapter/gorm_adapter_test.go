package adapter

import (
	"testing"

	"github.com/adonmo/goroom/orm"
	"github.com/adonmo/goroom/room"
	"github.com/go-test/deep"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	DB      *gorm.DB
	Adapter orm.ORM
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
	fields := []*GORMField{}
	fields = append(fields, &GORMField{
		Name: "ID:int",
		Tag:  `gorm:"primary_key"`,
	}, &GORMField{
		Name: "Value:string",
	})

	expectedOutput := orm.ModelDefinition{
		EntityModel: &GORMEntityModel{
			Fields: fields,
		},
		TableName: expectedModel.TableName(suite.DB),
	}

	got := suite.Adapter.GetModelDefinition(DummyTable{})
	diff := deep.Equal(expectedOutput, got)

	if diff != nil {
		suite.T().Errorf("GetModelDefinition not wokring as expected. %v", diff)
	}

	//Same test when model is passed in as a reference
	assert.Equal(suite.T(), suite.Adapter.GetModelDefinition(&DummyTable{}), expectedOutput)
}

func (suite *IntegrationTestSuite) TestGetModelDefinitionWithBadInput() {

	assert.True(suite.T(), suite.Adapter.GetModelDefinition(nil) == orm.ModelDefinition{})
	assert.True(suite.T(), suite.Adapter.GetModelDefinition([]string{"abc"}) == orm.ModelDefinition{})
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
		Version:      orm.VersionNumber(23),
	}
	anotherDummyEntry := room.GoRoomSchemaMaster{
		IdentityHash: "eyryhyeue",
		Version:      orm.VersionNumber(24),
	}
	suite.Adapter.Create(&dummyEntry)
	suite.Adapter.Create(&anotherDummyEntry)

	identity, version, err := suite.Adapter.GetLatestSchemaIdentityHashAndVersion()
	queryResult := room.GoRoomSchemaMaster{
		IdentityHash: identity,
		Version:      orm.VersionNumber(version),
	}

	if err != nil {
		suite.T().Errorf("No error expected when querying schema master for latest record. Got: %v", err)
	}

	diff := deep.Equal(anotherDummyEntry, queryResult)
	if diff != nil {
		suite.T().Errorf("Query Latest not working as expected. Diff: %v", diff)
	}

}

func (suite *IntegrationTestSuite) TestDoInTransaction() {
	dummyEntry := room.GoRoomSchemaMaster{
		IdentityHash: "adaghsghas",
		Version:      orm.VersionNumber(23),
	}
	transactionFunc := func(orm orm.ORM) error {
		suite.Adapter.CreateTable(room.GoRoomSchemaMaster{})
		suite.Adapter.Create(&dummyEntry)

		return nil
	}

	suite.Adapter.DoInTransaction(transactionFunc)

}

func TestMain(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
