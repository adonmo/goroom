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
		modelA := room.db.NewScope(sortedEntities[i]).GetModelStruct()
		modelB := room.db.NewScope(sortedEntities[j]).GetModelStruct()

		return modelA.ModelType.Name() < modelB.ModelType.Name()
	})

	for _, entity := range sortedEntities {
		model := room.db.NewScope(entity).GetModelStruct()
		sum, err := deephash.ConstructHash(model)
		if err != nil {
			return "", fmt.Errorf("Error while calculating identity hash for Table %v", model.ModelType.Name())
		}
		entityHashArr = append(entityHashArr, sum)
	}

	identity, err := deephash.ConstructHash(entityHashArr)
	if err != nil {
		return "", fmt.Errorf("Error while calculating schema identity %v", entityHashArr)
	}

	return identity, nil
}
