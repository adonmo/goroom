package room

import (
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
	MockControl *gomock.Controller
}

func (suite *RoomConstructorTestSuite) BeforeTest() {
	mockCtrl := gomock.NewController(suite.T())
	suite.MockControl = mockCtrl
}

func (suite *RoomConstructorTestSuite) TearDownTest() {
	suite.MockControl.Finish()
}

func (suite *RoomConstructorTestSuite) TestNew() {
	// entities = []interface{}{DummyTable{}, AnotherDummyTable{}}
	// orm = mocks.NewMockORM(suite.MockControl)

}
