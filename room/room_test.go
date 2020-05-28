package room

import (
	"testing"
)

type DummyTable struct {
	ID    int `gorm:"primary_key"`
	Value string
}

type AnotherDummyTable struct {
	Num  int
	Text string
}

func TestNewRoom(t *testing.T) {

}
