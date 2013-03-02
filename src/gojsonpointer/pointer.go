// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gojsonpointer
// repository-desc	An implementation of JSON Pointer - Go language
// 
// description		Main and unique file.
// 
// created      	25-02-2013

package gojsonpointer

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	EMPTY_POINTER     = ``
	POINTER_SEPARATOR = `/`

	INVALID_START = `JSON pointer must be empty or start with a "` + POINTER_SEPARATOR
)

func NewJsonPointer(jsonPointerString string) (JsonPointer, error) {

	var p JsonPointer
	err := p.parse(jsonPointerString)
	return p, err

}

type JsonPointer struct {
	referenceTokens []string
}

// "Constructor", parses the given string JSON pointer
func (p *JsonPointer) parse(jsonPointerString string) error {

	var err error

	if jsonPointerString != EMPTY_POINTER {
		if !strings.HasPrefix(jsonPointerString, POINTER_SEPARATOR) {
			err = errors.New(INVALID_START)
		} else {
			referenceTokens := strings.Split(jsonPointerString, POINTER_SEPARATOR)
			for _, referenceToken := range referenceTokens[1:] {
				p.referenceTokens = append(p.referenceTokens, decodeReferenceToken(referenceToken))
			}
		}
	}

	return err
}

// Uses the pointer to retrieve a value from a JSON document
func (p *JsonPointer) Get(document interface{}) (interface{}, reflect.Kind, error) {

	kind := reflect.Invalid

	// Full document when empty
	if len(p.referenceTokens) == 0 {
		return document, kind, nil
	}

	node := document

	for _, token := range p.referenceTokens {

		rValue := reflect.ValueOf(node)
		kind = rValue.Kind()

		switch kind {

		case reflect.Map:
			m := node.(map[string]interface{})
			if _, ok := m[token]; ok {
				node = m[token]
			} else {
				return nil, kind, errors.New(fmt.Sprintf("Object has no key '%s'", token))
			}

		case reflect.Slice:
			s := node.([]interface{})
			tokenIndex, err := strconv.Atoi(token)
			if err != nil {
				return nil, kind, errors.New(fmt.Sprintf("Invalid array index '%s'", token))
			}
			sLength := len(s)
			if tokenIndex < 0 || tokenIndex >= sLength {
				return nil, kind, errors.New(fmt.Sprintf("Out of bound array[0,%d] index '%d'", tokenIndex, sLength))
			}

			node = s[tokenIndex]

		default:
			return nil, kind, errors.New(fmt.Sprintf("Invalid token reference '%s'", token))
		}

	}

	rValue := reflect.ValueOf(node)
	kind = rValue.Kind()

	return node, kind, nil
}

// Pointer to string representation function
func (p *JsonPointer) String() string {

	if len(p.referenceTokens) == 0 {
		return EMPTY_POINTER
	}

	tokens := p.referenceTokens
	for i := range tokens {
		tokens[i] = encodeReferenceToken(tokens[i])
	}

	pointerString := POINTER_SEPARATOR + strings.Join(tokens, POINTER_SEPARATOR)

	return pointerString
}

// Specific JSON pointer encoding here
// ~0 => ~ 
// ~1 => / 
// ... and vice versa 

const (
	ENCODED_REFERENCE_TOKEN_0 = `~0`
	ENCODED_REFERENCE_TOKEN_1 = `~1`
	DECODED_REFERENCE_TOKEN_0 = `~`
	DECODED_REFERENCE_TOKEN_1 = `/`
)

func decodeReferenceToken(token string) string {
	step1 := strings.Replace(token, ENCODED_REFERENCE_TOKEN_1, DECODED_REFERENCE_TOKEN_1, -1)
	step2 := strings.Replace(step1, ENCODED_REFERENCE_TOKEN_0, DECODED_REFERENCE_TOKEN_0, -1)
	return step2
}

func encodeReferenceToken(token string) string {
	step1 := strings.Replace(token, DECODED_REFERENCE_TOKEN_1, ENCODED_REFERENCE_TOKEN_1, -1)
	step2 := strings.Replace(step1, DECODED_REFERENCE_TOKEN_0, ENCODED_REFERENCE_TOKEN_0, -1)
	return step2
}
