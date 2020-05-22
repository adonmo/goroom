package deephash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"reflect"
	"sort"
	"strings"
)

//ConstructHash Construct Hash for a given interface
func ConstructHash(input interface{}) (ans string, err error) {
	digester := sha256.New()
	err = IterateAndDigestHash(input, &digester)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(digester.Sum(nil)), nil
}

//IterateAndDigestHash Constructs recursive hash
func IterateAndDigestHash(input interface{}, digester *hash.Hash) (err error) {

	fieldValue := reflect.Indirect(reflect.ValueOf(input))
	fieldKind := fieldValue.Type().Kind()
	if !fieldValue.IsValid() || fieldValue.IsZero() {
		return nil
	}

	switch fieldKind {
	case reflect.Map:
		err = handleMap(fieldValue, digester)
	case reflect.Struct, reflect.Ptr:
		err = handleComplex(fieldValue, digester)
	case reflect.Slice, reflect.Array:
		err = handleList(fieldValue, digester)
	default:
		err = digestBasicTypeValue(fieldValue, digester)
	}

	return err
}

func digestBasicTypeValue(fieldValue reflect.Value, digester *hash.Hash) (err error) {
	_, err = fmt.Fprint(*digester, reflect.ValueOf(fieldValue).Interface())
	return
}

func handleMap(fieldValue reflect.Value, digester *hash.Hash) (err error) {
	keyHash := make([]string, len(fieldValue.MapKeys()))
	keyHashValue := make(map[string]reflect.Value)

	for i, key := range fieldValue.MapKeys() {
		kh, err := ConstructHash(key.Interface())
		if err != nil {
			//Inner Scope err explicitly returned
			return err
		}
		keyHash[i] = kh
		keyHashValue[kh] = fieldValue.MapIndex(key)
	}
	sort.Strings(keyHash)

	for _, kh := range keyHash {
		_, err = fmt.Fprint(*digester, kh)
		if err != nil {
			return
		}
		vh, err := ConstructHash(keyHashValue[kh].Interface())
		if err != nil {
			//Inner Scope err explicitly returned
			return err
		}
		_, err = fmt.Fprint(*digester, vh)
	}
	return
}

func handleComplex(fieldValue reflect.Value, digester *hash.Hash) (err error) {
	for i := 0; i < fieldValue.NumField(); i++ {
		structFieldName := fieldValue.Type().Field(i).Name
		structFieldNameStart := structFieldName[0:1]
		if structFieldNameStart != strings.ToUpper(structFieldNameStart) {
			continue
		}
		fieldTag := fieldValue.Type().Field(i).Tag.Get("hash")
		fv := fieldValue.Field(i)
		if fv.IsZero() || !fv.IsValid() || fieldTag == "ignore" {
			continue
		}
		valOf := reflect.Indirect(fv).Interface()
		err = IterateAndDigestHash(valOf, digester)
		if err != nil {
			return
		}
	}
	return
}

func handleList(fieldValue reflect.Value, digester *hash.Hash) (err error) {
	// sort first, just like reflect.Map above
	var hashesAr []string
	for it := 0; it < fieldValue.Len(); it++ {
		itH, err := ConstructHash(reflect.Indirect(fieldValue.Index(it)).Interface())
		if err != nil {
			return err
		}
		hashesAr = append(hashesAr, itH)
	}
	sort.Strings(hashesAr)
	for _, h := range hashesAr {
		err = IterateAndDigestHash(h, digester)
	}

	return err
}
