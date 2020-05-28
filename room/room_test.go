package room

import (
	"fmt"
	"testing"

	"adonmo.com/goroom/orm"
	"adonmo.com/goroom/orm/mocks"
	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
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

func TestMain(t *testing.T) {
	suite.Run(t, new(RoomConstructorTestSuite))
	suite.Run(t, new(MigrationSetupTestSuite))
}
