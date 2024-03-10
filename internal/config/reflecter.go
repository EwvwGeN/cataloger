package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
)

type innerStack struct {
	stuctFields []reflect.Value
	tagsStack []string
	fullTag string
	fieldIndexStack []*int
	size int
}

func (is *innerStack) push(rt reflect.Value, tag string) {
	if tag != "" {
		is.tagsStack = append(is.tagsStack, tag)
		is.fullTag = strings.Join(is.tagsStack,"_")
	}
	is.stuctFields = append(is.stuctFields, rt)
	is.fieldIndexStack = append(is.fieldIndexStack, new(int))
	is.size++
}

func (is *innerStack) peak() (reflect.Value, string, *int) {
	return is.stuctFields[is.size-1], is.fullTag, is.fieldIndexStack[is.size-1]
}

func (is *innerStack) pop() {
	if is.size <= 0 {
		return
	}
	is.size--
	is.stuctFields = is.stuctFields[:is.size]
	is.fieldIndexStack = is.fieldIndexStack[:is.size]
	if is.size >= 1 {
		is.tagsStack = is.tagsStack[:is.size-1]
		is.fullTag = strings.Join(is.tagsStack,"_")
	}
}

func (is *innerStack) isEmpty() bool {
	return is.size == 0
}

// loadEnv parse  tags from the structure and inserts values into its fields
//
// yaml tags are used
func loadEnv(cfg *Config) error {
	cfgType := reflect.ValueOf(cfg).Elem()
	stack := &innerStack{}
	stack.push(cfgType, "")
	for {
		field, tag, curFieldIndex := stack.peak()
		if *curFieldIndex == field.NumField() {
			stack.pop()
			if stack.isEmpty() {
				break
			}
			continue
		}
		filedTag := strings.ToUpper(strings.Split(field.Type().Field(*curFieldIndex).Tag.Get("yaml"), ",")[0])
		nestedField := field.Field(*curFieldIndex)
		if nestedField.Kind() == reflect.Struct {
			stack.push(nestedField, filedTag)
			*curFieldIndex++
			continue
		}
		if tag != "" {
			tag = fmt.Sprintf("%s_%s", tag, filedTag)
		} else {
			tag = filedTag
		}
		fmt.Println(tag)
		if err := setValue(nestedField, getEnv(tag)); err != nil {
			return err
		}
		*curFieldIndex++
	}
	return nil
}

func getEnv(envKey string) string {
	if value, exists := os.LookupEnv(envKey); exists {
		return value
	}
	return ""
}

func setValue(r reflect.Value, value string) error {
	switch r.Kind() {
	case reflect.Int64:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		r.SetInt(reflect.ValueOf(dur).Int())
	default:
		r.Set(reflect.ValueOf(value))
	}
	return nil
}
