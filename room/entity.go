package room

import (
	"fmt"
	"sort"

	"adonmo.com/goroom/util/deephash"
)

func (room *Room) createEntities() {
	for _, entity := range room.entities {
		if !room.db.HasTable(entity) {
			room.db.CreateTable(entity)
		}
	}
}

func (room *Room) calculateIdentityHash() (string, error) {
	var entityHashArr []string
	var sortedEntities []interface{}
	copy(sortedEntities, room.entities)
	sort.Slice(sortedEntities[:], func(i, j int) bool {
		modelA := room.db.GetModelDefinition(sortedEntities[i])
		modelB := room.db.GetModelDefinition(sortedEntities[j])

		return modelA.TableName < modelB.TableName
	})

	for _, entity := range sortedEntities {
		model := room.db.GetModelDefinition(entity)
		sum, err := deephash.ConstructHash(model)
		if err != nil {
			return "", fmt.Errorf("Error while calculating identity hash for Table %v", model.TableName)
		}
		entityHashArr = append(entityHashArr, sum)
	}

	identity, err := deephash.ConstructHash(entityHashArr)
	if err != nil {
		return "", fmt.Errorf("Error while calculating schema identity %v", entityHashArr)
	}

	return identity, nil
}
