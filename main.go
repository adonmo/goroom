package main

import (
	"fmt"

	deephash "adonmo.com/goroom/util/deephash"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

//Product Buy
type Product struct {
	Code  string
	Price uint
}

//TestStruct ts
type TestStruct struct {
	a int
	B bool
}

func main() {
	fmt.Println("Hello Bitches")
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Product{})

	// Create
	db.Create(&Product{Code: "L1212", Price: 1000})

	// Read
	var product Product
	db.First(&product, 1)                   // find product with id 1
	db.First(&product, "code = ?", "L1212") // find product with code l1212

	// Update - update product's price to 2000
	db.Model(&product).Update("Price", 2000)

	// Delete - delete product
	db.Delete(&product)

	opScope := db.NewScope(&Product{})
	for _, field := range opScope.GetStructFields() {
		fmt.Println(field.Name)
		fmt.Println(deephash.ConstructHash(TestStruct{
			a: 1,
			B: false,
		}))
		fmt.Println(deephash.ConstructHash(opScope.GetModelStruct()))
	}

}
