package room

import (
	"fmt"

	"adonmo.com/goroom/orm"
	"adonmo.com/goroom/orm/mocks"
	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type MockEntityModel struct {
	Fields []string
}

type EntityTestSuite struct {
	suite.Suite
	AppDB                        *Room
	DBA                          *mocks.MockORM
	IdentityCalc                 *mocks.MockIdentityHashCalculator
	MockCtrl                     *gomock.Controller
	DummyTableEntityModel        MockEntityModel
	AnotherDummyTableEntityModel MockEntityModel
}

func (s *EntityTestSuite) SetupTest() {
	s.MockCtrl = gomock.NewController(s.T())
	s.DBA = mocks.NewMockORM(s.MockCtrl)
	s.IdentityCalc = mocks.NewMockIdentityHashCalculator(s.MockCtrl)
	s.AppDB = &Room{
		dba:                s.DBA,
		identityCalculator: s.IdentityCalc,
	}

	s.DummyTableEntityModel = MockEntityModel{
		Fields: []string{"id", "value"},
	}

	s.DBA.EXPECT().GetModelDefinition(DummyTable{}).Return(
		orm.ModelDefinition{
			TableName:   "dummy_table",
			EntityModel: s.DummyTableEntityModel,
		},
	).AnyTimes()

	s.AnotherDummyTableEntityModel = MockEntityModel{
		Fields: []string{"num", "text"},
	}

	s.DBA.EXPECT().GetModelDefinition(AnotherDummyTable{}).Return(
		orm.ModelDefinition{
			TableName:   "another_dummy_table",
			EntityModel: s.AnotherDummyTableEntityModel,
		},
	).AnyTimes()
}

func (s *EntityTestSuite) TestCalculateIdentityHash() {

	dummyTableModelHash := "asasasadefe"
	anotherDummyTableModelHash := "fefefefefe"

	entityHashArr := []string{anotherDummyTableModelHash, dummyTableModelHash}
	expectedIdentityHash := "erecasadfergf"

	entitiesOrder1 := []interface{}{DummyTable{}, AnotherDummyTable{}}
	entitiesOrder2 := []interface{}{AnotherDummyTable{}, DummyTable{}}

	s.AppDB.entities = entitiesOrder1
	gomock.InOrder(
		s.IdentityCalc.EXPECT().ConstructHash(s.AnotherDummyTableEntityModel).Return(anotherDummyTableModelHash, nil),
		s.IdentityCalc.EXPECT().ConstructHash(s.DummyTableEntityModel).Return(dummyTableModelHash, nil),
		s.IdentityCalc.EXPECT().ConstructHash(entityHashArr).Return(expectedIdentityHash, nil),
	)
	identityHash1, err1 := s.AppDB.calculateIdentityHash()

	s.AppDB.entities = entitiesOrder2
	gomock.InOrder(
		s.IdentityCalc.EXPECT().ConstructHash(s.AnotherDummyTableEntityModel).Return(anotherDummyTableModelHash, nil),
		s.IdentityCalc.EXPECT().ConstructHash(s.DummyTableEntityModel).Return(dummyTableModelHash, nil),
		s.IdentityCalc.EXPECT().ConstructHash(entityHashArr).Return(expectedIdentityHash, nil),
	)
	identityHash2, err2 := s.AppDB.calculateIdentityHash()

	diff := deep.Equal(identityHash1, identityHash2)
	diffFromExpected := deep.Equal(identityHash1, expectedIdentityHash)
	if diff != nil || err1 != nil || err2 != nil || diffFromExpected != nil {
		s.T().Errorf("Identity Hash Calculation not deterministic for a given set of entities. DiffForRuns: %v DiffFromExpected: %v. Errors: %v %v",
			diff, diffFromExpected, err1, err2)
	}
}

func (s *EntityTestSuite) TestCalculateIdentityHashWithErrorInModelHashConstruction() {

	expectedError := fmt.Errorf("Error while calculating identity hash for Table %v", "another_dummy_table")

	entitiesOrder := []interface{}{DummyTable{}, AnotherDummyTable{}}

	s.AppDB.entities = entitiesOrder
	gomock.InOrder(
		s.IdentityCalc.EXPECT().ConstructHash(s.AnotherDummyTableEntityModel).Return("", expectedError),
	)
	_, err := s.AppDB.calculateIdentityHash()
	diff := deep.Equal(expectedError, err)

	if diff != nil {
		s.T().Errorf("Identity Hash Calculation not working per expectation in error scenario. Diff:%v", diff)
	}
}

func (s *EntityTestSuite) TestCalculateIdentityHashWithErrorInOverallHashConstruction() {

	dummyTableModelHash := "asasasadefe"
	anotherDummyTableModelHash := "fefefefefe"

	entityHashArr := []string{anotherDummyTableModelHash, dummyTableModelHash}
	entitiesOrder := []interface{}{DummyTable{}, AnotherDummyTable{}}

	s.AppDB.entities = entitiesOrder
	expectedError := fmt.Errorf("Error while calculating schema identity %v", entityHashArr)
	gomock.InOrder(
		s.IdentityCalc.EXPECT().ConstructHash(s.AnotherDummyTableEntityModel).Return(anotherDummyTableModelHash, nil),
		s.IdentityCalc.EXPECT().ConstructHash(s.DummyTableEntityModel).Return(dummyTableModelHash, nil),
		s.IdentityCalc.EXPECT().ConstructHash(entityHashArr).Return("", fmt.Errorf("Some error in Hashing")),
	)
	_, err := s.AppDB.calculateIdentityHash()

	diff := deep.Equal(expectedError, err)
	if diff != nil {
		s.T().Errorf("Identity Hash Calculation not working per expectation in error scenario. Diff:%v", diff)
	}
}
