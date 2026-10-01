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
	"reflect"
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

type testBar struct {
	Number float64
}

type testDocument struct {
	Foo testBar
}

type testTaggedBar struct {
	Number float64 `json:"number"`
}

type testTaggedDocument struct {
	Foo    testTaggedBar            `json:"foo"`
	Items  []testTaggedBar          `json:"items"`
	Map    map[string]testTaggedBar `json:"map"`
	PtrBar *testTaggedBar           `json:"ptr_bar"`
	secret string
}

type testStringerKey string

func (k testStringerKey) String() string {
	return string(k)
}

type testEmbeddedBase struct {
	ID string `json:"id"`
}

type testEmbeddedDoc struct {
	testEmbeddedBase
	Name string `json:"name"`
}

func TestStructDocument(t *testing.T) {
	// Issue #8 example
	doc := testDocument{
		Foo: testBar{
			Number: 42.0,
		},
	}

	p, err := NewJsonPointer("/Foo/Number")
	if err != nil {
		t.Fatalf("NewJsonPointer failed: %v", err)
	}

	val, kind, err := p.Get(doc)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if kind != reflect.Float64 || val != 42.0 {
		t.Errorf("Expected 42.0 (float64), got %v (%v)", val, kind)
	}

	// Pointer to struct root
	valPtr, kindPtr, err := p.Get(&doc)
	if err != nil {
		t.Fatalf("Get on pointer to struct failed: %v", err)
	}
	if kindPtr != reflect.Float64 || valPtr != 42.0 {
		t.Errorf("Expected 42.0 (float64), got %v (%v)", valPtr, kindPtr)
	}

	// Get struct field directly
	pFoo, _ := NewJsonPointer("/Foo")
	valFoo, kindFoo, err := pFoo.Get(doc)
	if err != nil {
		t.Fatalf("Get /Foo failed: %v", err)
	}
	if kindFoo != reflect.Struct || valFoo.(testBar).Number != 42.0 {
		t.Errorf("Expected testBar{Number: 42.0}, got %v", valFoo)
	}
}

func TestStructWithJsonTags(t *testing.T) {
	doc := testTaggedDocument{
		Foo:    testTaggedBar{Number: 100.5},
		secret: "hidden",
	}

	// Match by JSON tag
	pTag, _ := NewJsonPointer("/foo/number")
	val, _, err := pTag.Get(doc)
	if err != nil {
		t.Fatalf("Get(/foo/number) failed: %v", err)
	}
	if val != 100.5 {
		t.Errorf("Expected 100.5, got %v", val)
	}

	// Match by struct field name
	pField, _ := NewJsonPointer("/Foo/Number")
	valField, _, err := pField.Get(doc)
	if err != nil {
		t.Fatalf("Get(/Foo/Number) failed: %v", err)
	}
	if valField != 100.5 {
		t.Errorf("Expected 100.5, got %v", valField)
	}

	// Unexported field should not be accessible
	pSecret, _ := NewJsonPointer("/secret")
	_, _, err = pSecret.Get(doc)
	if err == nil {
		t.Errorf("Expected error accessing unexported field, got nil")
	}

	// Non-existent field
	pNonExist, _ := NewJsonPointer("/non_existent")
	_, _, err = pNonExist.Get(doc)
	if err == nil {
		t.Errorf("Expected error for non-existent field, got nil")
	}
}

func TestStructSet(t *testing.T) {
	doc := testDocument{
		Foo: testBar{
			Number: 10.0,
		},
	}

	p, _ := NewJsonPointer("/Foo/Number")

	// Set on pointer to struct
	_, err := p.Set(&doc, 99.0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if doc.Foo.Number != 99.0 {
		t.Errorf("Expected 99.0, got %v", doc.Foo.Number)
	}

	// Set on non-pointer struct should fail
	_, err = p.Set(doc, 123.0)
	if err == nil {
		t.Errorf("Expected error setting on non-pointer struct, got nil")
	}
}

func TestStructComplexTypes(t *testing.T) {
	doc := testTaggedDocument{
		Foo: testTaggedBar{Number: 1.0},
		Items: []testTaggedBar{
			{Number: 10.0},
			{Number: 20.0},
		},
		Map: map[string]testTaggedBar{
			"first": {Number: 30.0},
		},
		PtrBar: &testTaggedBar{Number: 40.0},
	}

	// Struct -> Slice -> Struct
	pSlice, _ := NewJsonPointer("/items/1/number")
	valSlice, _, err := pSlice.Get(doc)
	if err != nil {
		t.Fatalf("Get(/items/1/number) failed: %v", err)
	}
	if valSlice != 20.0 {
		t.Errorf("Expected 20.0, got %v", valSlice)
	}

	// Struct -> Map -> Struct
	pMap, _ := NewJsonPointer("/map/first/number")
	valMap, _, err := pMap.Get(doc)
	if err != nil {
		t.Fatalf("Get(/map/first/number) failed: %v", err)
	}
	if valMap != 30.0 {
		t.Errorf("Expected 30.0, got %v", valMap)
	}

	// Struct -> Pointer to Struct
	pPtr, _ := NewJsonPointer("/ptr_bar/number")
	valPtr, _, err := pPtr.Get(doc)
	if err != nil {
		t.Fatalf("Get(/ptr_bar/number) failed: %v", err)
	}
	if valPtr != 40.0 {
		t.Errorf("Expected 40.0, got %v", valPtr)
	}

	// Mutate through pointer
	_, err = pPtr.Set(&doc, 88.0)
	if err != nil {
		t.Fatalf("Set(/ptr_bar/number) failed: %v", err)
	}
	if doc.PtrBar.Number != 88.0 {
		t.Errorf("Expected 88.0, got %v", doc.PtrBar.Number)
	}
}

func TestMapStringerKey(t *testing.T) {
	m := map[testStringerKey]int{
		testStringerKey("apple"):  1,
		testStringerKey("banana"): 2,
	}

	p, _ := NewJsonPointer("/banana")
	val, _, err := p.Get(m)
	if err != nil {
		t.Fatalf("Get(/banana) failed: %v", err)
	}
	if val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}

	_, err = p.Set(m, 42)
	if err != nil {
		t.Fatalf("Set(/banana) failed: %v", err)
	}
	if m[testStringerKey("banana")] != 42 {
		t.Errorf("Expected 42, got %v", m[testStringerKey("banana")])
	}
}

func TestEmbeddedStruct(t *testing.T) {
	doc := testEmbeddedDoc{
		testEmbeddedBase: testEmbeddedBase{ID: "base-123"},
		Name:             "test-user",
	}

	pID, _ := NewJsonPointer("/id")
	valID, _, err := pID.Get(doc)
	if err != nil {
		t.Fatalf("Get(/id) failed: %v", err)
	}
	if valID != "base-123" {
		t.Errorf("Expected 'base-123', got %v", valID)
	}

	pName, _ := NewJsonPointer("/name")
	valName, _, err := pName.Get(doc)
	if err != nil {
		t.Fatalf("Get(/name) failed: %v", err)
	}
	if valName != "test-user" {
		t.Errorf("Expected 'test-user', got %v", valName)
	}
}
