package orm

//ORM The orm component used by Room
type ORM interface {
	HasTable(value interface{}) bool
	CreateTable(models ...interface{}) Result
	Delete(value interface{}, where ...interface{}) Result
	Create(value interface{}) Result
	DropTable(values ...interface{}) Result
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
