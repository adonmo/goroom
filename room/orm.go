package room

import "github.com/jinzhu/gorm"

//ORM The orm component used by Room
type ORM interface {
	HasTable(value interface{}) bool
	CreateTable(models ...interface{}) ORM
	Delete(value interface{}, where ...interface{}) ORM
	Create(value interface{}) ORM
	DropTable(values ...interface{}) ORM
	GetModelDefinition(entity interface{}) ModelDefinition
}

//ModelDefinition Interface to access Definition of ORM Entity Model
type ModelDefinition struct {
	TableName   string
	EntityModel interface{}
}

//Result Result from DB operations
type Result struct {
	Error error
}

//GORMAdapter Adpater for GORM as used by Room
type GORMAdapter struct {
	db *gorm.DB
}

//HasTable Check Table exists
func (adapter *GORMAdapter) HasTable(value interface{}) bool {
	return adapter.db.HasTable(value)
}

//CreateTable Create a Table
func (adapter *GORMAdapter) CreateTable(value interface{}) Result {
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
