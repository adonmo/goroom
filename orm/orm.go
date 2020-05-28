package orm

//VersionNumber Type for specifying version number across Room
type VersionNumber uint

//ORM The orm component used by Room
type ORM interface {
	HasTable(entity interface{}) bool
	CreateTable(models ...interface{}) Result
	TruncateTable(entity interface{}) Result
	Create(entity interface{}) Result
	DropTable(entities ...interface{}) Result
	GetModelDefinition(entity interface{}) ModelDefinition
	GetUnderlyingORM() interface{}
	GetLatestSchemaIdentityHashAndVersion() (identityHash string, version int, err error)
	DoInTransaction(fc func(tx ORM) error) (err error) //In the event of error returned by fc rollback should happen, nil return value should lead to commit
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

//IdentityHashCalculator Calculates Identity based on the entity model definition returned by ORM
type IdentityHashCalculator interface {
	ConstructHash(entityModel interface{}) (ans string, err error)
}

//Migration Interface against users can define their migrations on the DB
type Migration interface {
	GetBaseVersion() VersionNumber
	GetTargetVersion() VersionNumber
	Apply(db interface{}) error
}
