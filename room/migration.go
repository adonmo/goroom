package room

import (
	"fmt"
	"sort"

	"adonmo.com/goroom/logger"
	"adonmo.com/goroom/orm"
)

//GetApplicableMigrations Fetches applicable migrations based on src and destination version numbers
func GetApplicableMigrations(migrations []orm.Migration, src orm.VersionNumber, dest orm.VersionNumber) (applicableMigrations []orm.Migration, err error) {
	migrationMap := getMigrationMap(migrations)
	isUpgrade := src < dest

	for isUpgrade && src < dest || !isUpgrade && dest < src {
		applicableTargets := migrationMap[src]
		if len(applicableTargets) < 1 {
			return []orm.Migration{}, fmt.Errorf("Unable to generate path for migration from %v to %v", src, dest)
		}

		first := len(applicableTargets) - 1
		last := -1
		searchStepIncrement := -1

		if !isUpgrade {
			first = 0
			last = len(applicableTargets)
			searchStepIncrement = 1
		}

		pathFound := false
		for i := first; i != last; i += searchStepIncrement {
			targetVersion := applicableTargets[i].GetTargetVersion()

			if isUpgrade && targetVersion <= dest || !isUpgrade && targetVersion >= dest {
				pathFound = true
				src = targetVersion
				applicableMigrations = append(applicableMigrations, applicableTargets[i])
				break
			}
		}

		if !pathFound {
			return []orm.Migration{}, fmt.Errorf("Unable to generate path for migration from %v to %v", src, dest)
		}
	}

	return
}

func getMigrationMap(migrations []orm.Migration) map[orm.VersionNumber][]orm.Migration {

	migrationMap := make(map[orm.VersionNumber][]orm.Migration)
	for _, migration := range migrations {
		start := migration.GetBaseVersion()

		applicableTargetsForStart := migrationMap[start]
		migrationMap[start] = append(applicableTargetsForStart, migration)
	}

	for _, candidates := range migrationMap {
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].GetTargetVersion() < candidates[j].GetTargetVersion()
		})
	}

	return migrationMap
}

func (appDB *Room) performMigrations(currentIdentityHash string, applicableMigrations []orm.Migration) error {
	migrationFunc := func(dba orm.ORM) error {
		for _, migration := range applicableMigrations {
			err := migration.Apply(dba.GetUnderlyingORM())
			if err != nil {
				logger.Errorf("Failed while applying migration. %v", migration)
				return err
			}
		}

		dbExec := dba.TruncateTable(GoRoomSchemaMaster{})
		if dbExec.Error != nil {
			logger.Errorf("Error while purging Room Schema Master. %v", dbExec.Error)
			return dbExec.Error
		}

		metadata := GoRoomSchemaMaster{
			Version:      appDB.version,
			IdentityHash: currentIdentityHash,
		}

		dbExec = dba.Create(&metadata)
		if dbExec.Error != nil {
			logger.Errorf("Error while adding entity hash to Room Schema Master. %v", dbExec.Error)
			return dbExec.Error
		}

		return nil
	}

	return appDB.dba.DoInTransaction(migrationFunc)
}
