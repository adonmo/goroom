package room

import (
	"fmt"

	"adonmo.com/goroom/orm"
	"adonmo.com/goroom/orm/mocks"
	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SchemaMasterTestSuite struct {
	suite.Suite
	MockCtrl *gomock.Controller
}

func (s *SchemaMasterTestSuite) SetupTest() {
	s.MockCtrl = gomock.NewController(s.T())
}

func (s *SchemaMasterTestSuite) TestIsSchemaMasterPresent() {
	dba := mocks.NewMockORM(s.MockCtrl)
	appDB := &Room{
		dba: dba,
	}

	dba.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(true)
	assert.True(s.T(), appDB.isSchemaMasterPresent())

	dba.EXPECT().HasTable(GoRoomSchemaMaster{}).Return(false)
	assert.True(s.T(), !appDB.isSchemaMasterPresent())
}

func (s *SchemaMasterTestSuite) TestGetRoomMetadataFromDB() {
	dba := mocks.NewMockORM(s.MockCtrl)
	appDB := &Room{
		dba: dba,
	}

	identityHash := "asasasasa"
	version := 4
	expected := &GoRoomSchemaMaster{
		IdentityHash: identityHash,
		Version:      orm.VersionNumber(version),
	}
	dba.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return(identityHash, version, nil)
	got, err := appDB.getRoomMetadataFromDB()
	diff := deep.Equal(expected, got)

	if diff != nil || err != nil {
		s.Errorf(err, "Room Metadata Fetching not working as expected. Diff: %v", diff)
	}
}

func (s *SchemaMasterTestSuite) TestGetRoomMetadataFromDBWithError() {
	dba := mocks.NewMockORM(s.MockCtrl)
	appDB := &Room{
		dba: dba,
	}

	expectedErr := fmt.Errorf("DB Error while fetching")
	dba.EXPECT().GetLatestSchemaIdentityHashAndVersion().Return("", 0, expectedErr)
	_, gotErr := appDB.getRoomMetadataFromDB()

	diff := deep.Equal(expectedErr, gotErr)

	if diff != nil {
		s.T().Errorf("Room Metadata Fetching not working as expected in case of error from DBA. Diff: %v", diff)
	}
}
