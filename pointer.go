// Copyright 2015 xeipuuv ( https://github.com/xeipuuv )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author  			xeipuuv
// author-github 	https://github.com/xeipuuv
// author-mail		xeipuuv@gmail.com
//
// repository-name	gojsonpointer
// repository-desc	An implementation of JSON Pointer - Go language
//
// description		Main and unique file.
//
// created      	25-02-2013

package gojsonpointer

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	const_empty_pointer     = ``
	const_pointer_separator = `/`

	const_invalid_start = `JSON pointer must be empty or start with a "` + const_pointer_separator + `"`
)

type implStruct struct {
	mode string // "SET" or "GET"

	inDocument interface{}

	setInValue interface{}

	getOutNode interface{}
	getOutKind reflect.Kind
	outError   error
}

type JsonPointer struct {
	referenceTokens []string
}

// NewJsonPointer parses the given string JSON pointer and returns an object
func NewJsonPointer(jsonPointerString string) (p JsonPointer, err error) {

	// Pointer to the root of the document
	if len(jsonPointerString) == 0 {
		// Keep referenceTokens nil
		return
	}
	if jsonPointerString[0] != '/' {
		return p, errors.New(const_invalid_start)
	}

	p.referenceTokens = strings.Split(jsonPointerString[1:], const_pointer_separator)
	return
}

// Uses the pointer to retrieve a value from a JSON document
func (p *JsonPointer) Get(document interface{}) (interface{}, reflect.Kind, error) {

	is := &implStruct{mode: "GET", inDocument: document}
	p.implementation(is)
	return is.getOutNode, is.getOutKind, is.outError

}

// Uses the pointer to update a value from a JSON document
func (p *JsonPointer) Set(document interface{}, value interface{}) (interface{}, error) {

	is := &implStruct{mode: "SET", inDocument: document, setInValue: value}
	p.implementation(is)
	return document, is.outError

}

// Uses the pointer to delete a value from a JSON document
func (p *JsonPointer) Delete(document interface{}) (interface{}, error) {
	is := &implStruct{mode: "DEL", inDocument: document}
	p.implementation(is)
	return document, is.outError
}

// Both Get and Set functions use the same implementation to avoid code duplication
func (p *JsonPointer) implementation(i *implStruct) {

	kind := reflect.Invalid

	// Full document when empty
	if len(p.referenceTokens) == 0 {
		i.getOutNode = i.inDocument
		i.outError = nil
		i.getOutKind = kind
		i.outError = nil
		return
	}

	node := i.inDocument

	previousNodes := make([]interface{}, len(p.referenceTokens))
	previousTokens := make([]string, len(p.referenceTokens))

	for ti, token := range p.referenceTokens {

		isLastToken := ti == len(p.referenceTokens)-1
		previousNodes[ti] = node
		previousTokens[ti] = token

		switch v := node.(type) {

		case map[string]interface{}:
			decodedToken := decodeReferenceToken(token)
			if _, ok := v[decodedToken]; ok {
				node = v[decodedToken]
				if isLastToken && i.mode == "SET" {
					v[decodedToken] = i.setInValue
				} else if isLastToken && i.mode == "DEL" {
					delete(v, decodedToken)
				}
			} else if isLastToken && i.mode == "SET" {
				v[decodedToken] = i.setInValue
			} else {
				i.outError = fmt.Errorf("Object has no key '%s'", decodedToken)
				i.getOutKind = reflect.Map
				i.getOutNode = nil
				return
			}

		case []interface{}:
			tokenIndex, err := strconv.Atoi(token)
			if err != nil {
				i.outError = fmt.Errorf("Invalid array index '%s'", token)
				i.getOutKind = reflect.Slice
				i.getOutNode = nil
				return
			}
			if tokenIndex < 0 || tokenIndex >= len(v) {
				i.outError = fmt.Errorf("Out of bound array[0,%d] index '%d'", len(v), tokenIndex)
				i.getOutKind = reflect.Slice
				i.getOutNode = nil
				return
			}

			node = v[tokenIndex]
			if isLastToken && i.mode == "SET" {
				v[tokenIndex] = i.setInValue
			} else if isLastToken && i.mode == "DEL" {
				v[tokenIndex] = v[len(v)-1]
				v[len(v)-1] = nil
				v = v[:len(v)-1]
				previousNodes[ti-1].(map[string]interface{})[previousTokens[ti-1]] = v
			}

		default:
			if node == nil {
				i.outError = fmt.Errorf("Invalid token reference '%s'", token)
				i.getOutKind = reflect.Invalid
				i.getOutNode = nil
				return
			}

			rValue := reflect.ValueOf(node)
			for rValue.IsValid() && (rValue.Kind() == reflect.Ptr || rValue.Kind() == reflect.Interface) {
				if rValue.IsNil() {
					i.outError = fmt.Errorf("Invalid token reference '%s'", token)
					i.getOutKind = rValue.Kind()
					i.getOutNode = nil
					return
				}
				rValue = rValue.Elem()
			}

			if !rValue.IsValid() {
				i.outError = fmt.Errorf("Invalid token reference '%s'", token)
				i.getOutKind = reflect.Invalid
				i.getOutNode = nil
				return
			}

			switch rValue.Kind() {

			case reflect.Struct:
				decodedToken := decodeReferenceToken(token)
				fieldVal, ok := findStructField(rValue, decodedToken)
				if !ok {
					i.outError = fmt.Errorf("Object has no key '%s'", decodedToken)
					i.getOutKind = reflect.Struct
					i.getOutNode = nil
					return
				}

				if isLastToken && i.mode == "SET" {
					if !fieldVal.CanSet() {
						i.outError = fmt.Errorf("Field '%s' cannot be set", decodedToken)
						return
					}
					valToSet := reflect.ValueOf(i.setInValue)
					if !valToSet.IsValid() {
						fieldVal.Set(reflect.Zero(fieldVal.Type()))
					} else if valToSet.Type().AssignableTo(fieldVal.Type()) {
						fieldVal.Set(valToSet)
					} else if valToSet.Type().ConvertibleTo(fieldVal.Type()) {
						fieldVal.Set(valToSet.Convert(fieldVal.Type()))
					} else if fieldVal.Kind() == reflect.Ptr && valToSet.Type().AssignableTo(fieldVal.Type().Elem()) {
						if fieldVal.IsNil() {
							fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
						}
						fieldVal.Elem().Set(valToSet)
					} else if fieldVal.Kind() == reflect.Ptr && valToSet.Type().ConvertibleTo(fieldVal.Type().Elem()) {
						if fieldVal.IsNil() {
							fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
						}
						fieldVal.Elem().Set(valToSet.Convert(fieldVal.Type().Elem()))
					} else {
						i.outError = fmt.Errorf("Cannot assign %s to field %s of type %s", valToSet.Type(), decodedToken, fieldVal.Type())
						return
					}
					node = fieldVal.Interface()
				} else if isLastToken && i.mode == "DEL" {
					i.outError = fmt.Errorf("Struct field '%s' cannot be deleted", decodedToken)
					i.getOutKind = reflect.Struct
					i.getOutNode = nil
					return
				} else {
					if i.mode == "SET" && fieldVal.CanAddr() {
						node = fieldVal.Addr().Interface()
					} else {
						node = fieldVal.Interface()
					}
				}

			case reflect.Map:
				decodedToken := decodeReferenceToken(token)
				mapType := rValue.Type()
				keyType := mapType.Key()
				var matchedKey reflect.Value

				tokenVal := reflect.ValueOf(decodedToken)
				if tokenVal.Type().AssignableTo(keyType) {
					matchedKey = tokenVal
				} else if tokenVal.Type().ConvertibleTo(keyType) {
					matchedKey = tokenVal.Convert(keyType)
				} else {
					for _, k := range rValue.MapKeys() {
						if tm, ok := k.Interface().(encoding.TextMarshaler); ok {
							if text, err := tm.MarshalText(); err == nil && string(text) == decodedToken {
								matchedKey = k
								break
							}
						} else if stringer, ok := k.Interface().(fmt.Stringer); ok {
							if stringer.String() == decodedToken {
								matchedKey = k
								break
							}
						} else if fmt.Sprint(k.Interface()) == decodedToken {
							matchedKey = k
							break
						}
					}
				}

				var mapVal reflect.Value
				if matchedKey.IsValid() {
					mapVal = rValue.MapIndex(matchedKey)
				}

				if mapVal.IsValid() {
					node = mapVal.Interface()
					if isLastToken && i.mode == "SET" {
						valToSet := reflect.ValueOf(i.setInValue)
						if !valToSet.IsValid() {
							rValue.SetMapIndex(matchedKey, reflect.Zero(mapType.Elem()))
						} else if valToSet.Type().AssignableTo(mapType.Elem()) {
							rValue.SetMapIndex(matchedKey, valToSet)
						} else if valToSet.Type().ConvertibleTo(mapType.Elem()) {
							rValue.SetMapIndex(matchedKey, valToSet.Convert(mapType.Elem()))
						} else {
							i.outError = fmt.Errorf("Cannot assign %s to map value of type %s", valToSet.Type(), mapType.Elem())
							return
						}
					} else if isLastToken && i.mode == "DEL" {
						rValue.SetMapIndex(matchedKey, reflect.Value{})
					}
				} else if isLastToken && i.mode == "SET" {
					if matchedKey.IsValid() {
						valToSet := reflect.ValueOf(i.setInValue)
						if !valToSet.IsValid() {
							rValue.SetMapIndex(matchedKey, reflect.Zero(mapType.Elem()))
						} else if valToSet.Type().AssignableTo(mapType.Elem()) {
							rValue.SetMapIndex(matchedKey, valToSet)
						} else if valToSet.Type().ConvertibleTo(mapType.Elem()) {
							rValue.SetMapIndex(matchedKey, valToSet.Convert(mapType.Elem()))
						} else {
							i.outError = fmt.Errorf("Cannot assign %s to map value of type %s", valToSet.Type(), mapType.Elem())
							return
						}
					} else if tokenVal.Type().AssignableTo(keyType) {
						valToSet := reflect.ValueOf(i.setInValue)
						if !valToSet.IsValid() {
							rValue.SetMapIndex(tokenVal, reflect.Zero(mapType.Elem()))
						} else if valToSet.Type().AssignableTo(mapType.Elem()) {
							rValue.SetMapIndex(tokenVal, valToSet)
						} else if valToSet.Type().ConvertibleTo(mapType.Elem()) {
							rValue.SetMapIndex(tokenVal, valToSet.Convert(mapType.Elem()))
						} else {
							i.outError = fmt.Errorf("Cannot assign %s to map value of type %s", valToSet.Type(), mapType.Elem())
							return
						}
					} else {
						i.outError = fmt.Errorf("Object has no key '%s'", decodedToken)
						i.getOutKind = reflect.Map
						i.getOutNode = nil
						return
					}
				} else {
					i.outError = fmt.Errorf("Object has no key '%s'", decodedToken)
					i.getOutKind = reflect.Map
					i.getOutNode = nil
					return
				}

			case reflect.Slice, reflect.Array:
				tokenIndex, err := strconv.Atoi(token)
				if err != nil {
					i.outError = fmt.Errorf("Invalid array index '%s'", token)
					i.getOutKind = rValue.Kind()
					i.getOutNode = nil
					return
				}
				if tokenIndex < 0 || tokenIndex >= rValue.Len() {
					i.outError = fmt.Errorf("Out of bound array[0,%d] index '%d'", rValue.Len(), tokenIndex)
					i.getOutKind = rValue.Kind()
					i.getOutNode = nil
					return
				}

				elemVal := rValue.Index(tokenIndex)
				if isLastToken && i.mode == "SET" {
					if !elemVal.CanSet() {
						i.outError = fmt.Errorf("Array index '%d' cannot be set", tokenIndex)
						return
					}
					valToSet := reflect.ValueOf(i.setInValue)
					if !valToSet.IsValid() {
						elemVal.Set(reflect.Zero(elemVal.Type()))
					} else if valToSet.Type().AssignableTo(elemVal.Type()) {
						elemVal.Set(valToSet)
					} else if valToSet.Type().ConvertibleTo(elemVal.Type()) {
						elemVal.Set(valToSet.Convert(elemVal.Type()))
					} else {
						i.outError = fmt.Errorf("Cannot assign %s to array element of type %s", valToSet.Type(), elemVal.Type())
						return
					}
					node = elemVal.Interface()
				} else if isLastToken && i.mode == "DEL" {
					i.outError = fmt.Errorf("Delete not supported on %s", rValue.Kind())
					i.getOutKind = rValue.Kind()
					i.getOutNode = nil
					return
				} else {
					if i.mode == "SET" && elemVal.CanAddr() {
						node = elemVal.Addr().Interface()
					} else {
						node = elemVal.Interface()
					}
				}

			default:
				i.outError = fmt.Errorf("Invalid token reference '%s'", token)
				i.getOutKind = rValue.Kind()
				i.getOutNode = nil
				return
			}
		}

	}

	i.getOutNode = node
	i.getOutKind = reflect.ValueOf(node).Kind()
	i.outError = nil
}

// Pointer to string representation function
func (p *JsonPointer) String() string {

	if len(p.referenceTokens) == 0 {
		return const_empty_pointer
	}

	pointerString := const_pointer_separator + strings.Join(p.referenceTokens, const_pointer_separator)

	return pointerString
}

// Specific JSON pointer encoding here
// ~0 => ~
// ~1 => /
// ... and vice versa

func decodeReferenceToken(token string) string {
	step1 := strings.Replace(token, `~1`, `/`, -1)
	step2 := strings.Replace(step1, `~0`, `~`, -1)
	return step2
}

func encodeReferenceToken(token string) string {
	step1 := strings.Replace(token, `~`, `~0`, -1)
	step2 := strings.Replace(step1, `/`, `~1`, -1)
	return step2
}

func findStructField(structVal reflect.Value, token string) (reflect.Value, bool) {
	structType := structVal.Type()
	var caseInsensitiveMatch reflect.Value
	hasCaseInsensitiveMatch := false

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		jsonName := ""
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			jsonName = parts[0]
		}

		if (jsonName != "" && jsonName == token) || fieldName == token {
			return structVal.Field(i), true
		}

		if !hasCaseInsensitiveMatch {
			if (jsonName != "" && strings.EqualFold(jsonName, token)) || strings.EqualFold(fieldName, token) {
				caseInsensitiveMatch = structVal.Field(i)
				hasCaseInsensitiveMatch = true
			}
		}
	}

	if hasCaseInsensitiveMatch {
		return caseInsensitiveMatch, true
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Anonymous {
			fieldVal := structVal.Field(i)
			for fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
				fieldVal = fieldVal.Elem()
			}
			if fieldVal.Kind() == reflect.Struct {
				if f, ok := findStructField(fieldVal, token); ok {
					return f, true
				}
			}
		}
	}

	return reflect.Value{}, false
}
