package utils

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"

	"github.com/hashicorp/go-getter/v2"
)

func Initialize(obj interface{}) interface{} {
	v := reflect.ValueOf(obj)
	initializeNils(v)

	return obj
}

func initializeNils(v reflect.Value) {
	// Dereference pointer(s).
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	if v.Kind() == reflect.Slice {
		// Initialize a nil slice.
		if v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
			return
		}

		// Recursively iterate over slice items.
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			initializeNils(item)
		}
	}

	// Recursively iterate over struct fields.
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			initializeNils(field)
		}
	}
}

// NilSliceToEmptySlice recursively sets nil slices to empty slices
func NilSliceToEmptySlice(inter interface{}) interface{} {
	// original input that can't be modified
	val := reflect.ValueOf(inter)

	switch val.Kind() {
	case reflect.Slice:
		newSlice := reflect.MakeSlice(val.Type(), 0, val.Len())
		if !val.IsZero() {
			// iterate over each element in slice
			for j := 0; j < val.Len(); j++ {
				item := val.Index(j)

				var newItem reflect.Value
				switch item.Kind() {
				case reflect.Struct:
					// recursively handle nested struct
					newItem = reflect.Indirect(reflect.ValueOf(NilSliceToEmptySlice(item.Interface())))
				default:
					newItem = item
				}

				newSlice = reflect.Append(newSlice, newItem)
			}

		}
		return newSlice.Interface()
	case reflect.Struct:
		// new struct that will be returned
		newStruct := reflect.New(reflect.TypeOf(inter))
		newVal := newStruct.Elem()
		// iterate over input's fields
		for i := 0; i < val.NumField(); i++ {
			newValField := newVal.Field(i)
			valField := val.Field(i)
			switch valField.Kind() {
			case reflect.Slice:
				// recursively handle nested slice
				newValField.Set(reflect.Indirect(reflect.ValueOf(NilSliceToEmptySlice(valField.Interface()))))
			case reflect.Struct:
				// recursively handle nested struct
				newValField.Set(reflect.Indirect(reflect.ValueOf(NilSliceToEmptySlice(valField.Interface()))))
			default:
				newValField.Set(valField)
			}
		}

		return newStruct.Interface()
	case reflect.Map:
		// new map to be returned
		newMap := reflect.MakeMap(reflect.TypeOf(inter))
		// iterate over every key value pair in input map
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			// recursively handle nested value
			newV := reflect.Indirect(reflect.ValueOf(NilSliceToEmptySlice(v.Interface())))
			newMap.SetMapIndex(k, newV)
		}
		return newMap.Interface()
	default:
		return inter
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func IDGenerator(length int) string {
	return StringWithCharset(length, charset)
}

func DownloadFile(src string, dst string) {
	client := &getter.Client{}
	request := &getter.Request{
		Src: src,
		Dst: dst,
	}
	output, err := client.Get(context.TODO(), request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting path: %v", err)
	}
	fmt.Printf("Downloading: %s\n", output.Dst)
}
