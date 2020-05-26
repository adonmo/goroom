package orm

import (
	"github.com/jinzhu/gorm"
)

//GORMAdapter Adpater for GORM as used by Room
type GORMAdapter struct {
	db *gorm.DB
}

//NewGORM Returns a new GORMAdapter
func NewGORM(db *gorm.DB) ORM {
	return &GORMAdapter{
		db: db,
	}
}

//HasTable Check Table exists
func (adapter *GORMAdapter) HasTable(value interface{}) bool {
	return adapter.db.HasTable(value)
}

//CreateTable Create a Table
func (adapter *GORMAdapter) CreateTable(value ...interface{}) Result {
	return Result{
		Error: adapter.db.CreateTable(value).Error,
	}
}

//Delete Delete Values
func (adapter *GORMAdapter) Delete(value interface{}, where ...interface{}) Result {
	return Result{
		Error: adapter.db.Delete(value, where).Error,
	}
}

//Create Create a row
func (adapter *GORMAdapter) Create(value interface{}) Result {
	return Result{
		Error: adapter.db.Create(value).Error,
	}
}

//DropTable Drop a table
func (adapter *GORMAdapter) DropTable(values ...interface{}) Result {
	return Result{
		Error: adapter.db.DropTable(values).Error,
	}
}

//GetModelDefinition Get representation of a database table(entity) as done by ORM
func (adapter *GORMAdapter) GetModelDefinition(entity interface{}) ModelDefinition {
	model := adapter.db.NewScope(entity).GetModelStruct()
	return ModelDefinition{
		EntityModel: model,
		TableName:   model.TableName(adapter.db),
	}
}

//GetUnderlyingORM Get the underlying ORM for advanced usage
func (adapter *GORMAdapter) GetUnderlyingORM() interface{} {
	return adapter.db
}

//QueryLatest Query the latest entry from database table
func (adapter *GORMAdapter) QueryLatest(entity interface{}, orderByColumnName string, orderByType string) (result interface{}, err error) {
	dbExec := adapter.db.Order(orderByColumnName + " " + orderByType).First(entity)
	if dbExec.Error != nil {
		err = dbExec.Error
	} else {
		result = entity
	}

	return
}
