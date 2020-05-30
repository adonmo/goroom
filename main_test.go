package goroom

import (
	"fmt"
	"testing"

	"adonmo.com/goroom/room/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RoomInitialzationTestSuite struct {
	suite.Suite
	MockCtrl    *gomock.Controller
	Initializer *mocks.MockInitializer
}

func (s *RoomInitialzationTestSuite) SetupTest() {

	s.MockCtrl = gomock.NewController(s.T())
	s.Initializer = mocks.NewMockInitializer(s.MockCtrl)
}

func (s *RoomInitialzationTestSuite) TestInitializeRoomWithIdentityCalculationError() {

	expectedError := fmt.Errorf("Some hashing error")
	s.Initializer.EXPECT().CalculateIdentityHash().Return("", expectedError).Times(2)

	//With Fallback Enabled
	assert.Equal(s.T(), expectedError, InitializeRoom(s.Initializer, true))
	//Without Fallback Enabled
	assert.Equal(s.T(), expectedError, InitializeRoom(s.Initializer, false))
}

func (s *RoomInitialzationTestSuite) TestInitializeRoomWithErrorOnFirstInit() {

	identityHash := "asasasawfw"
	initError := fmt.Errorf("Error during initialization")

	//With Retry not Recommended
	gomock.InOrder(
		s.Initializer.EXPECT().CalculateIdentityHash().Return(identityHash, nil),
		s.Initializer.EXPECT().Init(identityHash).Return(false, initError),
	)
	//With Fallback Enabled
	assert.Equal(s.T(), initError, InitializeRoom(s.Initializer, true))

	//With Retry not Recommended
	gomock.InOrder(
		s.Initializer.EXPECT().CalculateIdentityHash().Return(identityHash, nil),
		s.Initializer.EXPECT().Init(identityHash).Return(false, initError),
	)
	//With Fallback not Enabled
	assert.Equal(s.T(), initError, InitializeRoom(s.Initializer, true))

	//With Retry Recommended and Clean up success
	gomock.InOrder(
		s.Initializer.EXPECT().CalculateIdentityHash().Return(identityHash, nil),
		s.Initializer.EXPECT().Init(identityHash).Return(true, initError),
		s.Initializer.EXPECT().PerformDBCleanUp().Return(nil),
		s.Initializer.EXPECT().Init(identityHash).Return(true, nil),
	)

	//With Fallback Enabled
	assert.Equal(s.T(), nil, InitializeRoom(s.Initializer, true))

	//With Retry Recommended and Clean up success
	gomock.InOrder(
		s.Initializer.EXPECT().CalculateIdentityHash().Return(identityHash, nil),
		s.Initializer.EXPECT().Init(identityHash).Return(false, initError),
	)
	//With Fallback not Enabled
	assert.Equal(s.T(), initError, InitializeRoom(s.Initializer, true))

	dbCleanUpError := fmt.Errorf("Error in DB Cleanup")
	//With Retry Recommended and Clean up error
	gomock.InOrder(
		s.Initializer.EXPECT().CalculateIdentityHash().Return(identityHash, nil),
		s.Initializer.EXPECT().Init(identityHash).Return(true, initError),
		s.Initializer.EXPECT().PerformDBCleanUp().Return(dbCleanUpError),
	)

	//With Fallback Enabled
	assert.Equal(s.T(), dbCleanUpError, InitializeRoom(s.Initializer, true))

	//With Retry Recommended and Clean up error
	gomock.InOrder(
		s.Initializer.EXPECT().CalculateIdentityHash().Return(identityHash, nil),
		s.Initializer.EXPECT().Init(identityHash).Return(false, initError),
	)
	//With Fallback not Enabled
	assert.Equal(s.T(), initError, InitializeRoom(s.Initializer, true))

}

func TestMain(t *testing.T) {
	suite.Run(t, new(RoomInitialzationTestSuite))
}
