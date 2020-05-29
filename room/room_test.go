package room

import (
	"fmt"
	"testing"

	"adonmo.com/goroom/orm"
	"adonmo.com/goroom/orm/mocks"
	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DummyTable struct {
	ID    int `gorm:"primary_key"`
	Value string
}

type AnotherDummyTable struct {
	Num  int
	Text string
}

type RoomConstructorTestSuite struct {
	suite.Suite
	MockControl                    *gomock.Controller
	Entities                       []interface{}
	Version                        orm.VersionNumber
	Migrations                     []orm.Migration
	FallbackToDestructiveMigration bool
	Dba                            orm.ORM
	IdentityCalculator             orm.IdentityHashCalculator
}

func (suite *RoomConstructorTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	suite.MockControl = mockCtrl
	suite.Entities = []interface{}{DummyTable{}, AnotherDummyTable{}}
	suite.Dba = mocks.NewMockORM(suite.MockControl)
	suite.Version = orm.VersionNumber(3)
	suite.FallbackToDestructiveMigration = false
	suite.IdentityCalculator = mocks.NewMockIdentityHashCalculator(suite.MockControl)
	suite.Migrations = []orm.Migration{}
}

func (suite *RoomConstructorTestSuite) TestNewWithValidParams() {
	expected := &Room{
		entities:                       suite.Entities,
		dba:                            suite.Dba,
		version:                        suite.Version,
		migrations:                     suite.Migrations,
		fallbackToDestructiveMigration: suite.FallbackToDestructiveMigration,
		identityCalculator:             suite.IdentityCalculator,
	}

	got, errors := New(suite.Entities, suite.Dba, suite.Version, suite.Migrations, suite.FallbackToDestructiveMigration, suite.IdentityCalculator)
	diff := deep.Equal(expected, got)

	if diff != nil || len(errors) > 0 {
		suite.T().Errorf("Creation of Room not working as expected. Diff: %v, Errors: %v", diff, errors)
	}

}

func (suite *RoomConstructorTestSuite) TestNewWithEmptyEntities() {

	var expected *Room
	got, errors := New([]interface{}{}, suite.Dba, suite.Version, suite.Migrations, suite.FallbackToDestructiveMigration, suite.IdentityCalculator)
	diff := deep.Equal(expected, got)

	expectedError := fmt.Errorf("No entities provided for the database")

	if diff != nil || len(errors) != 1 || deep.Equal(expectedError, errors[0]) != nil {
		suite.T().Errorf("Creation of Room not working as expected for empty entity list. Diff: %v, Errors: %v", diff, errors)
	}
}

func (suite *RoomConstructorTestSuite) TestNewWithMissingDBA() {

	var expected *Room
	got, errors := New(suite.Entities, nil, suite.Version, suite.Migrations, suite.FallbackToDestructiveMigration, suite.IdentityCalculator)
	diff := deep.Equal(expected, got)

	expectedError := fmt.Errorf("Need an ORM to work with")

	if diff != nil || len(errors) != 1 || deep.Equal(expectedError, errors[0]) != nil {
		suite.T().Errorf("Creation of Room not working as expected for missing DBA. Diff: %v, Errors: %v", diff, errors)
	}
}

func (suite *RoomConstructorTestSuite) TestNewWithBadVersion() {

	var expected *Room
	got, errors := New(suite.Entities, suite.Dba, 0, suite.Migrations, suite.FallbackToDestructiveMigration, suite.IdentityCalculator)
	diff := deep.Equal(expected, got)

	expectedError := fmt.Errorf("Only non zero versions allowed")

	if diff != nil || len(errors) != 1 || deep.Equal(expectedError, errors[0]) != nil {
		suite.T().Errorf("Creation of Room not working as expected for bad version. Diff: %v, Errors: %v", diff, errors)
	}
}

func (suite *RoomConstructorTestSuite) TestNewWithMissingIdentityCalculator() {

	var expected *Room
	got, errors := New(suite.Entities, suite.Dba, suite.Version, suite.Migrations, suite.FallbackToDestructiveMigration, nil)
	diff := deep.Equal(expected, got)

	fmt.Printf("%v", suite)

	expectedError := fmt.Errorf("Need an identity calculator")

	if diff != nil || len(errors) != 1 || deep.Equal(expectedError, errors[0]) != nil {
		suite.T().Errorf("Creation of Room not working as expected for missing identity calculator. Diff: %v, Errors: %v", diff, errors)
	}
}

type RoomInitTestSuite struct {
	suite.Suite
	MockControl                    *gomock.Controller
	Entities                       []interface{}
	Version                        orm.VersionNumber
	Migrations                     []orm.Migration
	FallbackToDestructiveMigration bool
	Dba                            orm.ORM
	IdentityCalculator             orm.IdentityHashCalculator

	MockORM          *mocks.MockORM
	MockIdentityCalc *mocks.MockIdentityHashCalculator
	AppDB            *Room
}

func (s *RoomInitTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(s.T())
	s.MockControl = mockCtrl
	s.MockORM = mocks.NewMockORM(s.MockControl)
	s.MockIdentityCalc = mocks.NewMockIdentityHashCalculator(s.MockControl)
	s.Entities = []interface{}{DummyTable{}, AnotherDummyTable{}}
	s.Dba = s.MockORM
	s.Version = orm.VersionNumber(3)
	s.FallbackToDestructiveMigration = false
	s.IdentityCalculator = s.MockIdentityCalc
	s.Migrations = []orm.Migration{}

	s.AppDB = &Room{
		entities:                       s.Entities,
		dba:                            s.Dba,
		version:                        s.Version,
		migrations:                     s.Migrations,
		fallbackToDestructiveMigration: s.FallbackToDestructiveMigration,
		identityCalculator:             s.IdentityCalculator,
	}
}

func (s *RoomInitTestSuite) TestInitializeAppDBForScenario1() {

	identityHash := "asasaasa"

	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(false)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return(identityHash, nil).AnyTimes()

	dbCreationFunc := getFirstTimeDBCreationFunction(identityHash, s.AppDB.version, s.AppDB.entities)
	//TODO Tighter check on function arguments
	s.MockORM.EXPECT().DoInTransaction(gomock.AssignableToTypeOf(dbCreationFunc)).Return(nil)

	assert.Nil(s.T(), s.AppDB.InitializeAppDB(), "No error expected here for Scenario 1")
}

func (s *RoomInitTestSuite) TestInitializeAppDBForScenario1WithErrorInIdentityHashCalculation() {

	someError := fmt.Errorf("Hash calculation failed")

	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(false)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return("", someError).AnyTimes()

	assert.NotNil(s.T(), s.AppDB.InitializeAppDB(), "Expected an Error for Scenario 1 Hash Problem")
}

func (s *RoomInitTestSuite) TestInitializeAppDBForScenario1WithErrorInDBCreation() {

	identityHash := "asasaasa"

	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(false)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return(identityHash, nil).AnyTimes()

	dbCreationFunc := getFirstTimeDBCreationFunction(identityHash, s.AppDB.version, s.AppDB.entities)
	//TODO Tighter check on function arguments
	someError := fmt.Errorf("Creation Transaction Failed")
	s.MockORM.EXPECT().DoInTransaction(gomock.AssignableToTypeOf(dbCreationFunc)).Return(someError)

	assert.NotNil(s.T(), s.AppDB.InitializeAppDB(), "Expected an error for Scenario 1 Creation Problem")
}

func TestMain(t *testing.T) {
	suite.Run(t, new(RoomConstructorTestSuite))
	suite.Run(t, new(MigrationSetupTestSuite))
	suite.Run(t, new(MigrationExecutionTestSuite))
	suite.Run(t, new(SchemaMasterTestSuite))
	suite.Run(t, new(EntityTestSuite))
	suite.Run(t, new(DatabaseOperationsTestSuite))
	suite.Run(t, new(RoomInitTestSuite))
}
