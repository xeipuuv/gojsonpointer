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
// description		Automated tests on package.
//
// created      	03-03-2013

package gojsonpointer

import (
	"encoding/json"
	"testing"
)

const (
	TEST_DOCUMENT_NB_ELEMENTS = 13
	TEST_NODE_OBJ_NB_ELEMENTS = 4
	TEST_DOCUMENT_STRING      = `{
"foo": ["bar", "baz"],
"obj": { "a":1, "b":2, "c":[3,4], "d":[ {"e":9}, {"f":[50,51]} ] },
"": 0,
"a/b": 1,
"c%d": 2,
"e^f": 3,
"g|h": 4,
"i\\j": 5,
"k\"l": 6,
" ": 7,
"m~n": 8,
"o~/p": 9,
"q/~r": 10
}`
)

var testDocumentJson interface{}

func init() {
	json.Unmarshal([]byte(TEST_DOCUMENT_STRING), &testDocumentJson)
}

func TestEscaping(t *testing.T) {

	ins := []string{`/`, `/`, `/a~1b`, `/a~1b`, `/c%d`, `/e^f`, `/g|h`, `/i\j`, `/k"l`, `/ `, `/m~0n`, `/o~0~1p`, `/q~1~0r`}
	outs := []float64{0, 0, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	for i := range ins {

		p, err := NewJsonPointer(ins[i])
		if err != nil {
			t.Errorf("NewJsonPointer(%v) error %v", ins[i], err.Error())
		}
		s := p.String()
		if s != ins[i] {
			t.Errorf("\"%v\" -> \"%v\"", ins[i], s)
		}

		result, _, err := p.Get(testDocumentJson)
		if err != nil {
			t.Errorf("Get(%v) error %v", ins[i], err.Error())
		}

		if result != outs[i] {
			t.Errorf("Get(%v) = %v, expect %v", ins[i], result, outs[i])
		}
	}

}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewJsonPointer(`/definitions/simple/0/next`)
	}
}

func BenchmarkParseWithEscape(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewJsonPointer(`/definiti~0ons/simple/0/next`)
	}
}

func BenchmarkString(b *testing.B) {
	p, _ := NewJsonPointer(`/definitions/simple/0/next`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.String()
	}
}

func BenchmarkStringWithEscape(b *testing.B) {
	p, _ := NewJsonPointer(`/definiti~0ons/simple/0/next`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.String()
	}
}

func TestFullDocument(t *testing.T) {

	in := ``

	p, err := NewJsonPointer(in)
	if err != nil {
		t.Errorf("NewJsonPointer(%v) error %v", in, err.Error())
	}

	result, _, err := p.Get(testDocumentJson)
	if err != nil {
		t.Errorf("Get(%v) error %v", in, err.Error())
	}

	if len(result.(map[string]interface{})) != TEST_DOCUMENT_NB_ELEMENTS {
		t.Errorf("Get(%v) = %v, expect full document", in, result)
	}
}

func TestGetNode(t *testing.T) {

	in := `/obj`

	p, err := NewJsonPointer(in)
	if err != nil {
		t.Errorf("NewJsonPointer(%v) error %v", in, err.Error())
	}

	result, _, err := p.Get(testDocumentJson)
	if err != nil {
		t.Errorf("Get(%v) error %v", in, err.Error())
	}

	if len(result.(map[string]interface{})) != TEST_NODE_OBJ_NB_ELEMENTS {
		t.Errorf("Get(%v) = %v, expect full document", in, result)
	}
}

func BenchmarkGet(b *testing.B) {
	p, _ := NewJsonPointer(`/obj/d/1/f`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Get(testDocumentJson)
	}
}

func TestArray(t *testing.T) {

	ins := []string{`/foo/0`, `/foo/0`, `/foo/1`}
	outs := []string{"bar", "bar", "baz"}

	for i := range ins {

		p, err := NewJsonPointer(ins[i])
		if err != nil {
			t.Errorf("NewJsonPointer(%v) error %v", ins[i], err.Error())
		}

		result, _, err := p.Get(testDocumentJson)
		if err != nil {
			t.Errorf("Get(%v) error %v", ins[i], err.Error())
		}

		if result != outs[i] {
			t.Errorf("Get(%v) = %v, expect %v", ins[i], result, outs[i])
		}
	}

}

func TestObject(t *testing.T) {

	ins := []string{`/obj/a`, `/obj/b`, `/obj/c/0`, `/obj/c/1`, `/obj/c/1`, `/obj/d/1/f/0`}
	outs := []float64{1, 2, 3, 4, 4, 50}

	for i := range ins {

		p, err := NewJsonPointer(ins[i])
		if err != nil {
			t.Errorf("NewJsonPointer(%v) error %v", ins[i], err.Error())
		}

		result, _, err := p.Get(testDocumentJson)
		if err != nil {
			t.Errorf("Get(%v) error %v", ins[i], err.Error())
		}

		if result != outs[i] {
			t.Errorf("Get(%v) = %v, expect %v", ins[i], result, outs[i])
		}
	}

}

func TestSetNode(t *testing.T) {

	jsonText := `{"a":[{"b": 1, "c": 2}], "d": 3}`

	var jsonDocument interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	in := "/a/0/c"

	p, err := NewJsonPointer(in)
	if err != nil {
		t.Errorf("NewJsonPointer(%v) error %v", in, err.Error())
	}

	_, err = p.Set(jsonDocument, 999)
	if err != nil {
		t.Errorf("Set(%v) error %v", in, err.Error())
	}

	firstNode := jsonDocument.(map[string]interface{})
	if len(firstNode) != 2 {
		t.Errorf("Set(%s) failed", in)
	}

	sliceNode := firstNode["a"].([]interface{})
	if len(sliceNode) != 1 {
		t.Errorf("Set(%s) failed", in)
	}

	changedNode := sliceNode[0].(map[string]interface{})
	changedNodeValue := changedNode["c"].(int)

	if changedNodeValue != 999 {
		if len(sliceNode) != 1 {
			t.Errorf("Set(%s) failed", in)
		}
	}

}

func TestSetEmptyNode(t *testing.T) {

	jsonText := `{}`

	var jsonDocument interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	in := "/a"

	p, err := NewJsonPointer(in)
	if err != nil {
		t.Errorf("NewJsonPointer(%v) error %v", in, err.Error())
	}

	_, err = p.Set(jsonDocument, 999)
	if err != nil {
		t.Errorf("Set(%v) error %v", in, err.Error())
	}

	firstNode := jsonDocument.(map[string]interface{})
	target := firstNode["a"].(int)
	if target != 999 {
		t.Errorf("Set(%s) failed", in)
	}
}

func TestDelObject(t *testing.T) {
	jsonText := `{
		"a":["apple sauce", "ketchup", "soy sauce"],
		"d": {
			"z" : {
				"v" : {
					"name" : "donald mcbobble",
					"occupation" : "corporate overlord"
				}
			}
		}
	}`

	var jsonDocument map[string]interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	//Deleting an object key
	in := "/d/z/v/occupation"
	p, err := NewJsonPointer(in)
	if err != nil {
		t.Errorf("NewJsonPointer(%v) error %v", in, err.Error())
	}

	_, err = p.Delete(jsonDocument)
	if err != nil {
		t.Errorf("Delete(%v) error %v", in, err.Error())
	}

	var d map[string]interface{} = jsonDocument["d"].(map[string]interface{})
	var z map[string]interface{} = d["z"].(map[string]interface{})
	var v map[string]interface{} = z["v"].(map[string]interface{})

	if _, present := v["occupation"]; present {
		t.Errorf("Delete (%s) failed: key is still present in the map", in)
	}
}

func TestDelArray(t *testing.T) {
	jsonText := `{
		"a":["applesauce", "ketchup", "soysauce", "oliveoil"],
		"d": {
			"z" : {
				"v" : {
					"name" : "donald mcbobble",
					"occupation" : "corporate overlord",
					"responsibilities" : ["managing", "hiring"]
				}
			}
		}
	}`

	var jsonDocument map[string]interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	//Deleting an array member
	in := "/a/2"
	p, err := NewJsonPointer(in)
	if err != nil {
		t.Errorf("NewJsonPointer(%v) error %v", in, err.Error())
	}

	_, err = p.Delete(jsonDocument)
	if err != nil {
		t.Errorf("Delete(%v) error %v", in, err.Error())
	}

	a := jsonDocument["a"].([]interface{})
	if len(a) != 3 || a[2] == "soysauce" {
		t.Errorf("Delete(%v) error (%s)", in, a)
	}

}

func TestAppendArray(t *testing.T) {
	jsonText := `{
		"name":["younus","fami","zayan"],
		"occupation": {
			"title" : "King"
		}
	}`

	var jsonDocument map[string]interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	in := "/name/3"
	p, err := NewJsonPointer(in)
	if err != nil {
		t.Fatalf("NewJsonPointer(%v) error %v", in, err)
	}

	_, err = p.Set(jsonDocument, "Aryan")
	if err != nil {
		t.Fatalf("Set(%v) error %v", in, err)
	}

	nameList, ok := jsonDocument["name"].([]interface{})
	if !ok {
		t.Fatalf("jsonDocument[\"name\"] is not []interface{}")
	}

	if len(nameList) != 4 {
		t.Fatalf("Expected 4 elements, got %d", len(nameList))
	}

	if nameList[3] != "Aryan" {
		t.Errorf("Expected Aryan, got %v", nameList[3])
	}
}

func TestAppendArrayHyphen(t *testing.T) {
	jsonText := `{"name":["younus","fami"]}`

	var jsonDocument map[string]interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	in := "/name/-"
	p, err := NewJsonPointer(in)
	if err != nil {
		t.Fatalf("NewJsonPointer(%v) error %v", in, err)
	}

	_, err = p.Set(jsonDocument, "zayan")
	if err != nil {
		t.Fatalf("Set(%v) error %v", in, err)
	}

	nameList := jsonDocument["name"].([]interface{})
	if len(nameList) != 3 || nameList[2] != "zayan" {
		t.Errorf("Expected zayan at index 2, got %v", nameList)
	}

	// Hyphen in Get should return error (RFC 6901 section 4)
	_, _, err = p.Get(jsonDocument)
	if err == nil {
		t.Errorf("Expected error getting hyphen index, got nil")
	}
}

func TestAppendArrayRoot(t *testing.T) {
	jsonText := `["a", "b"]`

	var jsonDocument []interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	// Append via index
	p1, err := NewJsonPointer("/2")
	if err != nil {
		t.Fatalf("NewJsonPointer error: %v", err)
	}

	doc, err := p1.Set(jsonDocument, "c")
	if err != nil {
		t.Fatalf("Set error: %v", err)
	}

	arr := doc.([]interface{})
	if len(arr) != 3 || arr[2] != "c" {
		t.Fatalf("Expected [a b c], got %v", arr)
	}

	// Append via hyphen
	p2, err := NewJsonPointer("/-")
	if err != nil {
		t.Fatalf("NewJsonPointer error: %v", err)
	}

	doc, err = p2.Set(arr, "d")
	if err != nil {
		t.Fatalf("Set error: %v", err)
	}

	arr = doc.([]interface{})
	if len(arr) != 4 || arr[3] != "d" {
		t.Fatalf("Expected [a b c d], got %v", arr)
	}

	// Out of bounds set should error
	p3, err := NewJsonPointer("/10")
	if err != nil {
		t.Fatalf("NewJsonPointer error: %v", err)
	}
	_, err = p3.Set(arr, "e")
	if err == nil {
		t.Errorf("Expected out of bound error for /10, got nil")
	}
}

func TestDelRootArray(t *testing.T) {
	jsonText := `["first", "second", "third"]`

	var jsonDocument []interface{}
	json.Unmarshal([]byte(jsonText), &jsonDocument)

	p, err := NewJsonPointer("/1")
	if err != nil {
		t.Fatalf("NewJsonPointer error: %v", err)
	}

	doc, err := p.Delete(jsonDocument)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	arr := doc.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("Expected 2 elements after deletion, got %d", len(arr))
	}
}
