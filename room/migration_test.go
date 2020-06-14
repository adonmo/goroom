package room

import (
	"fmt"

	"github.com/adonmo/goroom/orm"
	"github.com/adonmo/goroom/orm/mocks"
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
		{2, 3}, {3, 4}, {4, 5}, {2, 5}, {5, 6}, {6, 8},
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

	migrations := append(suite.UpgradeMigrations, suite.DowngradeMigrations...)

	migration24, err := GetApplicableMigrations(migrations, 2, 4)
	isValidUpgradePath := len(migration24) == 2 && migration24[0].GetBaseVersion() == 2 && migration24[0].GetTargetVersion() == 3 &&
		migration24[1].GetBaseVersion() == 3 && migration24[1].GetTargetVersion() == 4 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 2 to 4 using %v. Err: %v", migration24, migrations, err)

	migration25, err := GetApplicableMigrations(migrations, 2, 5)
	isValidUpgradePath = len(migration25) == 1 && migration25[0].GetBaseVersion() == 2 && migration25[0].GetTargetVersion() == 5 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 2 to 5 using %v. Err: %v", migration25, migrations, err)

	migration56, err := GetApplicableMigrations(migrations, 5, 6)
	isValidUpgradePath = len(migration25) == 1 && migration56[0].GetBaseVersion() == 5 && migration56[0].GetTargetVersion() == 6 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 5 to 6 using %v. Err: %v", migration56, migrations, err)

	migration34, err := GetApplicableMigrations(migrations, 3, 4)
	isValidUpgradePath = len(migration25) == 1 && migration34[0].GetBaseVersion() == 3 && migration34[0].GetTargetVersion() == 4 && err == nil
	assert.Truef(suite.T(), isValidUpgradePath, "Wrong migration plan %v for 3 to 4 using %v. Err: %v", migration34, migrations, err)

}

func (suite *MigrationSetupTestSuite) TestGetApplicableMigrationsForDowngrade() {

	migrations := append(suite.UpgradeMigrations, suite.DowngradeMigrations...)

	migration52, err := GetApplicableMigrations(migrations, 5, 2)
	isValidDowngradePath := len(migration52) == 2 && migration52[0].GetBaseVersion() == 5 && migration52[0].GetTargetVersion() == 3 &&
		migration52[1].GetBaseVersion() == 3 && migration52[1].GetTargetVersion() == 2 && err == nil
	assert.Truef(suite.T(), isValidDowngradePath, "Wrong migration plan %v for 5 to 2 using %v. Err: %v", migration52, migrations, err)

	migration32, err := GetApplicableMigrations(migrations, 3, 2)
	isValidDowngradePath = len(migration32) == 1 && migration32[0].GetBaseVersion() == 3 && migration32[0].GetTargetVersion() == 2 && err == nil
	assert.Truef(suite.T(), isValidDowngradePath, "Wrong migration plan %v for 3 to 2 using %v. Err: %v", migration32, migrations, err)

	migration53, err := GetApplicableMigrations(migrations, 5, 3)
	isValidDowngradePath = len(migration32) == 1 && migration53[0].GetBaseVersion() == 5 && migration53[0].GetTargetVersion() == 3 && err == nil
	assert.Truef(suite.T(), isValidDowngradePath, "Wrong migration plan %v for 5 to 3 using %v. Err: %v", migration53, migrations, err)

	migration43, err := GetApplicableMigrations(migrations, 4, 3)
	isValidDowngradePath = len(migration32) == 1 && migration43[0].GetBaseVersion() == 4 && migration43[0].GetTargetVersion() == 3 && err == nil
	assert.Truef(suite.T(), isValidDowngradePath, "Wrong migration plan %v for 4 to 3 using %v. Err: %v", migration43, migrations, err)

}

func (suite *MigrationSetupTestSuite) TestGetApplicableMigrationsForNonExistentSourceVersion() {

	migrations := append(suite.UpgradeMigrations, suite.DowngradeMigrations...)
	src := orm.VersionNumber(1)
	dest := orm.VersionNumber(2)
	expectedError := fmt.Errorf("Unable to generate path for migration from %v to %v", src, dest)

	_, err := GetApplicableMigrations(migrations, src, dest)
	assert.Equal(suite.T(), expectedError, err, "Incorrect Error when fetching migrations for non existent source version")

}

func (suite *MigrationSetupTestSuite) TestGetApplicableMigrationsForNonExistentDestinationVersion() {

	migrations := suite.UpgradeMigrations
	src := orm.VersionNumber(6)
	dest := orm.VersionNumber(7)
	expectedError := fmt.Errorf("Unable to generate path for migration from %v to %v", src, dest)

	_, err := GetApplicableMigrations(migrations, src, dest)
	assert.Equal(suite.T(), expectedError, err, "Incorrect Error when fetching migrations for non existent destination version")

}

type MigrationExecutionTestSuite struct {
	suite.Suite
	MockCtrl          *gomock.Controller
	ValidMigrations   []orm.Migration
	InvalidMigrations []orm.Migration
	AppDB             *Room
	MockDBA           *mocks.MockORM
}

func (suite *MigrationExecutionTestSuite) SetupTest() {

	suite.MockCtrl = gomock.NewController(suite.T())

	for range []int{1, 2, 3} {
		m := mocks.NewMockMigration(suite.MockCtrl)
		m.EXPECT().Apply(gomock.Any()).Return(nil).AnyTimes()
		suite.ValidMigrations = append(suite.ValidMigrations, m)

		m = mocks.NewMockMigration(suite.MockCtrl)
		m.EXPECT().Apply(gomock.Any()).Return(fmt.Errorf("Some DB Error")).AnyTimes()
		suite.InvalidMigrations = append(suite.ValidMigrations, m)
	}

	suite.MockDBA = mocks.NewMockORM(suite.MockCtrl)
	suite.AppDB = &Room{
		version: orm.VersionNumber(3),
		dba:     suite.MockDBA,
	}
}

func (suite *MigrationExecutionTestSuite) TestGetMigrationTransactionFunctionWithInvalidMigrations() {

	var dummyORM interface{}
	suite.MockDBA.EXPECT().GetUnderlyingORM().Return(dummyORM).AnyTimes()

	migrationFunc := getMigrationTransactionFunction(suite.AppDB.version, "asasasa", append(suite.ValidMigrations, suite.InvalidMigrations...))

	err := migrationFunc(suite.AppDB.dba)
	assert.NotNil(suite.T(), err, "Should have received an error for invalid migrations")
}

func (suite *MigrationExecutionTestSuite) TestGetMigrationTransactionFunctionWithFailedTruncationOfSchemaMaster() {

	var dummyORM interface{}
	expectedError := fmt.Errorf("Some DB mess happened")
	suite.MockDBA.EXPECT().GetUnderlyingORM().Return(dummyORM).AnyTimes()
	suite.MockDBA.EXPECT().TruncateTable(GoRoomSchemaMaster{}).Return(orm.Result{
		Error: expectedError,
	})

	migrationFunc := getMigrationTransactionFunction(suite.AppDB.version, "asasasa", append(suite.ValidMigrations))

	err := migrationFunc(suite.AppDB.dba)
	assert.Equal(suite.T(), expectedError, err, "Should have received the expected error for failed truncation of schema master")
}

func (suite *MigrationExecutionTestSuite) TestGetMigrationTransactionFunctionWithFailedCreationOfMetadata() {

	var dummyORM interface{}
	expectedError := fmt.Errorf("Creation Failed")
	identityHash := "asasasa"
	suite.MockDBA.EXPECT().GetUnderlyingORM().Return(dummyORM).AnyTimes()
	suite.MockDBA.EXPECT().TruncateTable(GoRoomSchemaMaster{}).Return(orm.Result{
		Error: nil,
	})
	suite.MockDBA.EXPECT().Create(&GoRoomSchemaMaster{
		Version:      suite.AppDB.version,
		IdentityHash: identityHash,
	}).Return(orm.Result{
		Error: expectedError,
	})

	migrationFunc := getMigrationTransactionFunction(suite.AppDB.version, identityHash, append(suite.ValidMigrations))

	err := migrationFunc(suite.AppDB.dba)
	assert.Equal(suite.T(), expectedError, err, "Should have received the expected error for failed creation of schema master record")
}

func (suite *MigrationExecutionTestSuite) TestGetMigrationTransactionFunction() {

	var dummyORM interface{}
	identityHash := "asasasa"
	suite.MockDBA.EXPECT().GetUnderlyingORM().Return(dummyORM).AnyTimes()
	suite.MockDBA.EXPECT().TruncateTable(GoRoomSchemaMaster{}).Return(orm.Result{
		Error: nil,
	})
	suite.MockDBA.EXPECT().Create(&GoRoomSchemaMaster{
		Version:      suite.AppDB.version,
		IdentityHash: identityHash,
	}).Return(orm.Result{
		Error: nil,
	})

	migrationFunc := getMigrationTransactionFunction(suite.AppDB.version, identityHash, append(suite.ValidMigrations))

	err := migrationFunc(suite.AppDB.dba)
	assert.Equal(suite.T(), nil, err, "No error expected. This is supposed to be the ideal scenario.")
}
