package adapter

import (
	"reflect"

	"go/ast"

	"github.com/adonmo/goroom/orm"
	"github.com/adonmo/goroom/room"
	"github.com/jinzhu/gorm"
)

//GORMField Representation
type GORMField struct {
	Name string
	Tag  reflect.StructTag
}

//GORMEntityModel Entity Model for GORM for Room
type GORMEntityModel struct {
	Fields []*GORMField
}

//GORMAdapter Adpater for GORM as used by Room
type GORMAdapter struct {
	db *gorm.DB
}

//NewGORM Returns a new GORMAdapter
func NewGORM(db *gorm.DB) orm.ORM {
	return &GORMAdapter{
		db: db,
	}
}

//HasTable Check Table exists
func (adapter *GORMAdapter) HasTable(entity interface{}) bool {
	return adapter.db.HasTable(entity)
}

//CreateTable Create a Table
func (adapter *GORMAdapter) CreateTable(entities ...interface{}) orm.Result {
	return orm.Result{
		Error: adapter.db.CreateTable(entities...).Error,
	}
}

//TruncateTable Delete All Values from table
func (adapter *GORMAdapter) TruncateTable(entity interface{}) orm.Result {
	return orm.Result{
		Error: adapter.db.Delete(entity).Error,
	}
}

//Create Create a row
func (adapter *GORMAdapter) Create(entity interface{}) orm.Result {
	return orm.Result{
		Error: adapter.db.Create(entity).Error,
	}
}

//DropTable Drop a table
func (adapter *GORMAdapter) DropTable(entities ...interface{}) orm.Result {
	return orm.Result{
		Error: adapter.db.DropTable(entities...).Error,
	}
}

//GetModelDefinition Get representation of a database table(entity) as done by ORM
func (adapter *GORMAdapter) GetModelDefinition(entity interface{}) (modelDefinition orm.ModelDefinition) {
	if entity == nil {
		return
	}

	reflectType := reflect.ValueOf(entity).Type()
	for reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	if reflectType.Kind() != reflect.Struct {
		return
	}

	fields := make([]*GORMField, 0, reflectType.NumField())
	for i := 0; i < reflectType.NumField(); i++ {
		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {
			fields = append(fields, &GORMField{
				Name: fieldStruct.Name + ":" + fieldStruct.Type.Name(),
				Tag:  fieldStruct.Tag,
			})
		}
	}

	model := adapter.db.NewScope(entity).GetModelStruct()
	return orm.ModelDefinition{
		EntityModel: &GORMEntityModel{
			Fields: fields,
		},
		TableName: model.TableName(adapter.db),
	}
}

//GetUnderlyingORM Get the underlying ORM for advanced usage
func (adapter *GORMAdapter) GetUnderlyingORM() interface{} {
	return adapter.db
}

//GetLatestSchemaIdentityHashAndVersion Query the latest schema master entry
func (adapter *GORMAdapter) GetLatestSchemaIdentityHashAndVersion() (identityHash string, version int, err error) {
	var latest room.GoRoomSchemaMaster
	dbExec := adapter.db.Order("version DESC").First(&latest)
	return latest.IdentityHash, int(latest.Version), dbExec.Error
}

//DoInTransaction Perform operations specified in the input function in a transaction
func (adapter *GORMAdapter) DoInTransaction(fc func(tx orm.ORM) error) (err error) {
	gormTxFunc := func(tx *gorm.DB) error {
		return fc(NewGORM(tx))
	}

	return adapter.db.Transaction(gormTxFunc)
}
