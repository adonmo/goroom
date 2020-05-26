package adapter

import (
	"adonmo.com/goroom/util/deephash"
)

//EntityHashConstructor Constructs entity Hash for a given ORM Model description
type EntityHashConstructor struct{}

//ConstructHash Constructs Hash for given input
func (c *EntityHashConstructor) ConstructHash(input interface{}) (ans string, err error) {
	return deephash.ConstructHash(input)
}
