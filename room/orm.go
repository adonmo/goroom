package room

//ORM The orm component used by Room
type ORM interface {
	HasTable(value interface{}) bool
	CreateTable(models ...interface{}) ORM
	Delete(value interface{}, where ...interface{}) ORM
	Create(value interface{}) ORM
	DropTable(values ...interface{}) ORM
	GetModelStruct(entity interface{}) ModelDefinition
}

//ModelDefinition Interface to access Definition of ORM Entity Model
type ModelDefinition interface {
	TableName(ORM) string
}
