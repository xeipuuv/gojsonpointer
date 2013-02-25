// @author       sigu-399
// @description  An implementation of json pointer in Golang
// @created      25-02-2013

package gojsonpointer

import (
	"errors"
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
	stringRepresentation string
	referenceTokens      []string
}

func (p *JsonPointer) parse(jsonPointerString string) error {

	var err error
	p.stringRepresentation = jsonPointerString

	if p.stringRepresentation != EMPTY_POINTER {
		if !strings.HasPrefix(p.stringRepresentation, POINTER_SEPARATOR) && !strings.HasPrefix(p.stringRepresentation, FRAGMENT) {
			err = errors.New(INVALID_START)
		} else {
			referenceTokens := strings.Split(p.stringRepresentation, POINTER_SEPARATOR)
			for _, referenceToken := range referenceTokens[1:] {
				p.referenceTokens = append(p.referenceTokens, decodeReferenceToken(referenceToken))
			}
		}
	}

	return err
}

func (p *JsonPointer) String() string {
	return p.stringRepresentation
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
