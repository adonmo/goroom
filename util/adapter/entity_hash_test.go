package adapter

import (
	"testing"

	"github.com/adonmo/goroom/util/deephash"
	"github.com/go-test/deep"
)

type AnotherStruct struct {
	MapVar map[string]interface{}
}

type TestStruct struct {
	StringVar      string
	IntVar         int
	StructVar      AnotherStruct
	PtrToStructVar *AnotherStruct
	MapVar         map[int]string
	MapItoIVar     map[interface{}]interface{}
	IgnoreVar      int `hash:"ignore"`
	Ivar           interface{}
}

type YetAnother struct {
	StringVar string
	MapVar    map[int]string
	hiddenVar string //This value will not be accounted for when calculating the digest
}

func TestConstructHash(t *testing.T) {
	ptr := &AnotherStruct{MapVar: map[string]interface{}{"one": "two", "three": []string{"four", "five"}}}

	data := TestStruct{
		StringVar:      "test string",
		IntVar:         123,
		StructVar:      AnotherStruct{MapVar: map[string]interface{}{"one": "two", "three": []string{"four", "five"}}},
		PtrToStructVar: ptr,
		MapVar:         map[int]string{1: "one", 2: "two"},
		IgnoreVar:      1,
		MapItoIVar: map[interface{}]interface{}{
			"test": YetAnother{
				StringVar: "strvartest",
				MapVar:    map[int]string{44: "forty-four"},
			},
			"another_test": YetAnother{
				StringVar: "strvartesttwo",
				MapVar:    map[int]string{55: "fifty-five"},
			},
		},
		Ivar: []YetAnother{{
			StringVar: "aaaaa",
			MapVar:    map[int]string{4333: "asdf"},
		}, {
			StringVar: "bbbbbbb",
			MapVar:    map[int]string{555: "ddd"},
		},
		},
	}

	expected, errExpected := deephash.ConstructHash(data)
	got, errGot := new(EntityHashConstructor).ConstructHash(data)

	diff := deep.Equal(expected, got)
	if diff != nil || deep.Equal(errExpected, errGot) != nil || errExpected != nil || errGot != nil {
		t.Errorf("Hash Construction has a problem. %v", diff)
	}
}
