package room

import (
	"fmt"

	"github.com/adonmo/goroom/orm"
	"github.com/adonmo/goroom/orm/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DatabaseOperationsTestSuite struct {
	suite.Suite
	MockCtrl *gomock.Controller
	DBA      *mocks.MockORM
}

func (s *DatabaseOperationsTestSuite) SetupTest() {
	s.MockCtrl = gomock.NewController(s.T())
	s.DBA = mocks.NewMockORM(s.MockCtrl)
}

func (s *DatabaseOperationsTestSuite) TestGetFirstTimeDBCreationFunction() {

	identityHash := "asasasasa"
	version := orm.VersionNumber(4)
	entitiesToCreate := []interface{}{DummyTable{}, AnotherDummyTable{}}

	creationFunc := getFirstTimeDBCreationFunction(identityHash, version, entitiesToCreate)

	gomock.InOrder(
		s.DBA.EXPECT().CreateTable(GoRoomSchemaMaster{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(DummyTable{}).Return(false),
		s.DBA.EXPECT().CreateTable(DummyTable{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(AnotherDummyTable{}).Return(true),
		s.DBA.EXPECT().Create(&GoRoomSchemaMaster{
			Version:      version,
			IdentityHash: identityHash,
		}).Return(orm.Result{
			Error: nil,
		}),
	)

	assert.Nil(s.T(), creationFunc(s.DBA), "No error expected during DB creation for specified conditions in this test")
}

func (s *DatabaseOperationsTestSuite) TestGetFirstTimeDBCreationFunctionWithErrorInEntityCreation() {

	identityHash := "asasasasa"
	version := orm.VersionNumber(4)
	entitiesToCreate := []interface{}{DummyTable{}, AnotherDummyTable{}}

	creationFunc := getFirstTimeDBCreationFunction(identityHash, version, entitiesToCreate)

	expectedError := fmt.Errorf("DB mess in creating table")

	gomock.InOrder(
		s.DBA.EXPECT().CreateTable(GoRoomSchemaMaster{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(DummyTable{}).Return(false),
		s.DBA.EXPECT().CreateTable(DummyTable{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(AnotherDummyTable{}).Return(false),
		s.DBA.EXPECT().CreateTable(AnotherDummyTable{}).Return(orm.Result{
			Error: expectedError,
		}),
		s.DBA.EXPECT().Create(&GoRoomSchemaMaster{
			Version:      version,
			IdentityHash: identityHash,
		}).Return(orm.Result{
			Error: nil,
		}),
	)

	assert.Equal(s.T(), expectedError, creationFunc(s.DBA), "Error returned is incorrect")
}

func (s *DatabaseOperationsTestSuite) TestGetFirstTimeDBCreationFunctionWithErrorInMetadataCreation() {

	identityHash := "asasasasa"
	version := orm.VersionNumber(4)
	entitiesToCreate := []interface{}{DummyTable{}, AnotherDummyTable{}}

	creationFunc := getFirstTimeDBCreationFunction(identityHash, version, entitiesToCreate)

	expectedError := fmt.Errorf("DB mess in creating entry")

	gomock.InOrder(
		s.DBA.EXPECT().CreateTable(GoRoomSchemaMaster{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(DummyTable{}).Return(false),
		s.DBA.EXPECT().CreateTable(DummyTable{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(AnotherDummyTable{}).Return(true),
		s.DBA.EXPECT().Create(&GoRoomSchemaMaster{
			Version:      version,
			IdentityHash: identityHash,
		}).Return(orm.Result{
			Error: expectedError,
		}),
	)

	assert.Equal(s.T(), expectedError, creationFunc(s.DBA), "Error returned is incorrect")
}

func (s *DatabaseOperationsTestSuite) TestGetFirstTimeDBCreationFunctionWithErrorInSchemaMasterCreation() {

	identityHash := "asasasasa"
	version := orm.VersionNumber(4)
	entitiesToCreate := []interface{}{DummyTable{}, AnotherDummyTable{}}

	creationFunc := getFirstTimeDBCreationFunction(identityHash, version, entitiesToCreate)

	expectedError := fmt.Errorf("DB mess in creating schema master")

	gomock.InOrder(
		s.DBA.EXPECT().CreateTable(GoRoomSchemaMaster{}).Return(orm.Result{
			Error: expectedError,
		}),
	)

	assert.Equal(s.T(), expectedError, creationFunc(s.DBA), "Error returned is incorrect")
}

func (s *DatabaseOperationsTestSuite) TestGetDBCleanUpFunction() {

	entitiesToDelete := []interface{}{GoRoomSchemaMaster{}, DummyTable{}, AnotherDummyTable{}}
	deleteFunc := GetDBCleanUpFunction(entitiesToDelete)

	gomock.InOrder(
		s.DBA.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true),
		s.DBA.EXPECT().DropTable(GoRoomSchemaMaster{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(DummyTable{}).Return(false),
		s.DBA.EXPECT().HasTable(AnotherDummyTable{}).Return(true),
		s.DBA.EXPECT().DropTable(AnotherDummyTable{}).Return(orm.Result{
			Error: nil,
		}),
	)

	assert.Nil(s.T(), deleteFunc(s.DBA), "No Error Expected")
}

func (s *DatabaseOperationsTestSuite) TestGetDBCleanUpFunctionWithErrorInDroppingATable() {

	entitiesToDelete := []interface{}{GoRoomSchemaMaster{}, DummyTable{}, AnotherDummyTable{}}
	deleteFunc := GetDBCleanUpFunction(entitiesToDelete)
	expectedError := fmt.Errorf("DB could not drop it")

	gomock.InOrder(
		s.DBA.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true),
		s.DBA.EXPECT().DropTable(GoRoomSchemaMaster{}).Return(orm.Result{
			Error: nil,
		}),
		s.DBA.EXPECT().HasTable(DummyTable{}).Return(true),
		s.DBA.EXPECT().DropTable(DummyTable{}).Return(orm.Result{
			Error: expectedError,
		}),
	)

	assert.Equal(s.T(), expectedError, deleteFunc(s.DBA), "Incorrect Error Returned")
}

func (s *DatabaseOperationsTestSuite) TestPerformDBCleanup() {

	entitiesToDelete := []interface{}{GoRoomSchemaMaster{}, DummyTable{}, AnotherDummyTable{}}
	deleteFunc := GetDBCleanUpFunction(entitiesToDelete)

	s.DBA.EXPECT().DoInTransaction(gomock.AssignableToTypeOf(deleteFunc)).Return(nil)

	room := &Room{
		dba: s.DBA,
	}
	assert.Equal(s.T(), nil, room.PerformDBCleanUp(), "No Error expected when DB clenaup goes in successfully")

}

func (s *DatabaseOperationsTestSuite) TestPerformDBCleanupWithErrorInTransaction() {

	entitiesToDelete := []interface{}{GoRoomSchemaMaster{}, DummyTable{}, AnotherDummyTable{}}
	deleteFunc := GetDBCleanUpFunction(entitiesToDelete)

	expectedError := fmt.Errorf("Transaction Error from DB")
	s.DBA.EXPECT().DoInTransaction(gomock.AssignableToTypeOf(deleteFunc)).Return(expectedError)

	room := &Room{
		dba: s.DBA,
	}
	assert.Equal(s.T(), expectedError, room.PerformDBCleanUp(), "Unexpected error output when transaction fails during DB deletion")
}
