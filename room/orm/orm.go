package orm

//ORM The orm component used by Room
type ORM interface {
	HasTable(entity interface{}) bool
	CreateTable(models ...interface{}) Result
	TruncateTable(entity interface{}) Result
	Create(entity interface{}) Result
	DropTable(entities ...interface{}) Result
	GetModelDefinition(entity interface{}) ModelDefinition
	GetUnderlyingORM() interface{}
	QueryLatest(entity interface{}, orderByColumnName string, orderByType string) (result interface{}, err error)
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
