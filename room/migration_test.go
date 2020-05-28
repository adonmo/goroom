package room

import (
	"adonmo.com/goroom/orm"
	"adonmo.com/goroom/orm/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MigrationSetupTestSuite struct {
	suite.Suite
	MockCtrl            *gomock.Controller
	UpgradeMigrations   []orm.Migration
	DowngradeMigrations []orm.Migration
}

func (suite *MigrationSetupTestSuite) SetupTest() {
	suite.MockCtrl = gomock.NewController(suite.T())

	upgradeMigrationVersions := [][2]orm.VersionNumber{
		{2, 3}, {3, 4}, {4, 5}, {2, 5}, {5, 6},
	}

	downgradeMigrationVersions := [][2]orm.VersionNumber{
		{5, 4}, {4, 3}, {5, 3}, {3, 2},
	}

	for _, versionPair := range upgradeMigrationVersions {
		m := mocks.NewMockMigration(suite.MockCtrl)
		m.EXPECT().GetBaseVersion().Return(versionPair[0]).AnyTimes()
		m.EXPECT().GetTargetVersion().Return(versionPair[1]).AnyTimes()
		m.EXPECT().Apply(gomock.Any).Return(nil).AnyTimes()
		suite.UpgradeMigrations = append(suite.UpgradeMigrations, m)
	}

	for _, versionPair := range downgradeMigrationVersions {
		m := mocks.NewMockMigration(suite.MockCtrl)
		m.EXPECT().GetBaseVersion().Return(versionPair[0]).AnyTimes()
		m.EXPECT().GetTargetVersion().Return(versionPair[1]).AnyTimes()
		m.EXPECT().Apply(gomock.Any).Return(nil).AnyTimes()
		suite.DowngradeMigrations = append(suite.DowngradeMigrations, m)
	}
}

func (suite *MigrationSetupTestSuite) TestGetApplicableMigrationsForUpgrade() {

	migration24, err := GetApplicableMigrations(suite.UpgradeMigrations, 2, 4)
	isValidUpgradePath := len(migration24) == 2 && migration24[0].GetBaseVersion() == 2 && migration24[0].GetTargetVersion() == 3 &&
		migration24[1].GetBaseVersion() == 3 && migration24[1].GetTargetVersion() == 4 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 2 to 4 using %v. Err: %v", migration24, suite.UpgradeMigrations, err)

	migration25, err := GetApplicableMigrations(suite.UpgradeMigrations, 2, 5)
	isValidUpgradePath = len(migration25) == 1 && migration25[0].GetBaseVersion() == 2 && migration25[0].GetTargetVersion() == 5 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 2 to 5 using %v. Err: %v", migration25, suite.UpgradeMigrations, err)

	migration56, err := GetApplicableMigrations(suite.UpgradeMigrations, 5, 6)
	isValidUpgradePath = len(migration25) == 1 && migration56[0].GetBaseVersion() == 5 && migration56[0].GetTargetVersion() == 6 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 5 to 6 using %v. Err: %v", migration56, suite.UpgradeMigrations, err)

	migration34, err := GetApplicableMigrations(suite.UpgradeMigrations, 3, 4)
	isValidUpgradePath = len(migration25) == 1 && migration34[0].GetBaseVersion() == 3 && migration34[0].GetTargetVersion() == 4 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 3 to 4 using %v. Err: %v", migration34, suite.UpgradeMigrations, err)

}
