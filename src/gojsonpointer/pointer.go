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
	"strings"
)

const (
	EMPTY_POINTER     = ``
	POINTER_SEPARATOR = `/`
	FRAGMENT          = `#`
)

const (
	INVALID_START = `JSON pointer must be empty, start with a "` + POINTER_SEPARATOR + `" or a "` + FRAGMENT + `"`
)

func NewJsonPointer(jsonPointerString string) (JsonPointer, error) {

	var p JsonPointer
	err := p.parse(jsonPointerString)
	return p, err

}

type JsonPointer struct {
	hasFragment     bool
	referenceTokens []string
}

func (p *JsonPointer) parse(jsonPointerString string) error {

	var err error

	if jsonPointerString != EMPTY_POINTER {
		p.hasFragment = strings.HasPrefix(jsonPointerString, FRAGMENT)
		if !strings.HasPrefix(jsonPointerString, POINTER_SEPARATOR) && !p.hasFragment {
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

func (p *JsonPointer) Get(document interface{}) (interface{}, error) {

	if len(p.referenceTokens) == 0 {
		return document, nil
	}

	node := document

	for _, token := range p.referenceTokens {

		rValue := reflect.ValueOf(node)
		kind := rValue.Kind()

		switch kind {

		case reflect.Map:
			m := node.(map[string]interface{})
			if _, ok := m[token]; ok {
				node = m[token]
			} else {
				return nil, errors.New("No value found using this pointer")
			}

		default:
			return nil, errors.New(fmt.Sprintf("Unhandled kind %s in JsonPointer.Get", kind))
		}

	}

	return node, nil
}

func (p *JsonPointer) String() string {

	tokens := p.referenceTokens
	for i := range tokens {
		tokens[i] = encodeReferenceToken(tokens[i])
	}

	pointerString := strings.Join(tokens, POINTER_SEPARATOR)

	if p.hasFragment {
		return FRAGMENT + POINTER_SEPARATOR + pointerString
	}

	return pointerString
}

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
