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

func (s *RoomInitTestSuite) TestInitRoomDBForScenario1() {

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

	shouldRetry, err := s.AppDB.initRoomDB(identityHash)
	assert.True(s.T(), !shouldRetry && err == nil, "No error expected here for Scenario 1")
}

func (s *RoomInitTestSuite) TestInitRoomDBForScenario1WithErrorInIdentityHashCalculation() {

	someError := fmt.Errorf("Hash calculation failed")

	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(false)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return("", someError).AnyTimes()

	assert.NotNil(s.T(), s.AppDB.InitializeAppDB(), "Expected an Error for Scenario 1 Hash Problem")
}

func (s *RoomInitTestSuite) TestInitRoomDBForScenario1WithErrorInDBCreation() {

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

	shouldRetry, err := s.AppDB.initRoomDB(identityHash)
	assert.True(s.T(), shouldRetry && err != nil, "Expected an error for Scenario 1 Creation Problem")
}

func (s *RoomInitTestSuite) TestInitRoomDBForScenario2() {

	identityHash := "asasaasa"

	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return(identityHash, nil).AnyTimes()
	s.MockORM.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return(identityHash, int(s.AppDB.version), nil)

	shouldRetry, err := s.AppDB.initRoomDB(identityHash)
	assert.True(s.T(), !shouldRetry && err == nil, "No error expected here for Scenario 2")
}

func (s *RoomInitTestSuite) TestInitRoomDBForScenario2WithMetadataNotFetched() {

	identityHash := "asasaasa"

	someError := fmt.Errorf("Unable to fetch metadata from DB")
	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return(identityHash, nil).AnyTimes()
	s.MockORM.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return("", 0, someError)

	shouldRetry, err := s.AppDB.initRoomDB(identityHash)
	assert.True(s.T(), shouldRetry && someError == err, "Error does not seem to be what is expected here for Scenario 2")
}

func (s *RoomInitTestSuite) TestInitRoomDBForScenario2WithIdentityMismatch() {

	identityHash := "asasaasa"
	storedIdentityHash := "etererere"

	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockORM.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return(storedIdentityHash, int(s.AppDB.version), nil)

	expectedError := fmt.Errorf("Database signature mismatch. Version %v", s.AppDB.version)
	shouldRetry, err := s.AppDB.initRoomDB(identityHash)
	diff := deep.Equal(expectedError, err)
	assert.True(s.T(), shouldRetry && diff == nil, "Return value does not seem to be what is expected here for Scenario 2 as signature won't match. %v %v", shouldRetry, err)
}

func (s *RoomInitTestSuite) TestInitRoomDBForScenario3() {

	identityHash := "asasaasa"

	storedVersion := s.AppDB.version - 1
	storedHash := "asaswrwdwe"

	mockMigration := mocks.NewMockMigration(s.MockControl)
	mockMigration.EXPECT().GetBaseVersion().Return(storedVersion).AnyTimes()
	mockMigration.EXPECT().GetTargetVersion().Return(s.AppDB.version).AnyTimes()
	migrations := []orm.Migration{mockMigration}

	s.AppDB.migrations = migrations
	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return(identityHash, nil).AnyTimes()
	s.MockORM.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return(storedHash, int(storedVersion), nil)
	migrationFunc := getMigrationTransactionFunction(s.AppDB.version, identityHash, migrations)
	s.MockORM.EXPECT().DoInTransaction(gomock.AssignableToTypeOf(migrationFunc)).Return(nil)

	shouldRetry, err := s.AppDB.initRoomDB(identityHash)
	assert.True(s.T(), !shouldRetry && err == nil, "No error expected here for Scenario 3")

	//Missing migration strategy
	s.AppDB.migrations = []orm.Migration{}
	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return(identityHash, nil).AnyTimes()
	s.MockORM.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return(storedHash, int(storedVersion), nil)

	shouldRetry, err = s.AppDB.initRoomDB(identityHash)
	assert.True(s.T(), shouldRetry && err != nil, "Error expected here for Scenario 3 due to missing migration")

	//Failed Migration Execution
	s.AppDB.migrations = migrations
	someError := fmt.Errorf("DB Mess when doing migration")
	s.MockORM.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true)
	s.MockORM.EXPECT().GetModelDefinition(gomock.Any()).Return(orm.ModelDefinition{
		EntityModel: MockEntityModel{},
		TableName:   "asasa",
	}).AnyTimes()
	s.MockIdentityCalc.EXPECT().ConstructHash(gomock.Any()).Return(identityHash, nil).AnyTimes()
	s.MockORM.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return(storedHash, int(storedVersion), nil)
	s.MockORM.EXPECT().DoInTransaction(gomock.AssignableToTypeOf(migrationFunc)).Return(someError)

	shouldRetry, err = s.AppDB.initRoomDB(identityHash)
	assert.True(s.T(), shouldRetry && err != nil, "Error expected here for Scenario 3 due to failed migration")
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
