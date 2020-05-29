package room

import (
	"fmt"
	"sort"
)

func (appDB *Room) calculateIdentityHash() (string, error) {
	var entityHashArr []string
	var sortedEntities []interface{}
	copy(sortedEntities, appDB.entities)
	sort.Slice(sortedEntities[:], func(i, j int) bool {
		modelA := appDB.dba.GetModelDefinition(sortedEntities[i])
		modelB := appDB.dba.GetModelDefinition(sortedEntities[j])

		return modelA.TableName < modelB.TableName
	})

	for _, entity := range sortedEntities {
		model := appDB.dba.GetModelDefinition(entity)
		sum, err := appDB.identityCalculator.ConstructHash(model)
		if err != nil {
			return "", fmt.Errorf("Error while calculating identity hash for Table %v", model.TableName)
		}
		entityHashArr = append(entityHashArr, sum)
	}

	identity, err := appDB.identityCalculator.ConstructHash(entityHashArr)
	if err != nil {
		return "", fmt.Errorf("Error while calculating schema identity %v", entityHashArr)
	}

	return identity, nil
}
