package digest

import (
	"testing"

	"github.com/jasonrogena/log-analyse/types"
)

func TestCleanUri(t *testing.T) {
	if CleanUri("a/b") != "a/b" {
		t.Error("Expected a/b")
	}

	if CleanUri("a//b") != "a/b" {
		t.Error("Expected a/b")
	}

	if CleanUri("a///b") != "a/b" {
		t.Error("Expected a/b")
	}

	if CleanUri("a/") != "a/" {
		t.Error("Expected a/")
	}

	if CleanUri("//a") != "/a" {
		t.Error("Expected /a")
	}

	if CleanUri("//a//") != "/a/" {
		t.Error("Expected /a/")
	}
}

func TestGetDigestPayload(t *testing.T) {
	_, ok := GetDigestPayload("a string")
	if ok == true {
		t.Error("Expected function to return false")
	}

	testData := DigestPayload{Field: nil, Cnf: nil}
	_, ok = GetDigestPayload(testData)
	if ok == true {
		t.Error("Expected function to return false since field is nil")
	}

	field := types.Field{
		UUID:        "fddsfdsfdsfdsf",
		FieldType:   types.FieldType{Name: "blah", Typ: "blah"},
		ValueType:   "string",
		ValueString: "someValue"}
	testData = DigestPayload{Field: &field, Cnf: nil}
	_, ok = GetDigestPayload(testData)
	if ok == true {
		t.Error("Expected function to return false since field type is not 'request'")
	}

	field = types.Field{
		UUID:        "fddsfdsfdsfdsf",
		FieldType:   types.FieldType{Name: "request", Typ: "blah"},
		ValueType:   "string",
		ValueString: "someValue"}
	testData = DigestPayload{Field: &field, Cnf: nil}
	_, ok = GetDigestPayload(testData)
	if ok == false {
		t.Error("Expected function to return true")
	}
}

func TestGetUriParts(t *testing.T) {
	_, uriError1 := GetUriParts("", "*")
	if uriError1 == nil {
		t.Error("Expected GetUriParts to return error since provided request was a blank string")
	}

	_, uriError2 := GetUriParts("GET /api/v1/forms/197928.json?a=dfds", "*")
	if uriError2 == nil {
		t.Error("Expected GetUriParts to return error since provided request didn't have 3 parts")
	}

	_, uriError3 := GetUriParts("GET /blah HTTP/1.1", "^/api/*")
	if uriError3 == nil {
		t.Error("Expected GetUriParts to return error since uri in request didn't match regex")
	}

	uriParts1, uriError4 := GetUriParts("GET /api/v1/forms/197928.json?a=dfds HTTP/1.1", "^/api/*")
	if uriError4 != nil {
		t.Error("Expected GetUriParts not to return an error")
	}
	if len(uriParts1) != 5 {
		t.Error("Expected GetUriParts to return an array with 4 items")
	}
	expectedUriParts1 := [5]string{"", "api", "v1", "forms", "197928.json"}
	for idx := range expectedUriParts1 {
		if uriParts1[idx] != expectedUriParts1[idx] {
			t.Error("Expected item to be " + expectedUriParts1[idx] + " but is " + uriParts1[idx])
		}
	}

	uriParts2, uriError5 := GetUriParts("GET api/v1/forms/197928.json?a=dfds HTTP/1.1", "^api/*")
	if uriError5 != nil {
		t.Error("Expected GetUriParts not to return an error")
	}
	if len(uriParts2) != 4 {
		t.Error("Expected GetUriParts to return an array with 4 items")
	}
	expectedUriParts2 := [4]string{"api", "v1", "forms", "197928.json"}
	for idx := range expectedUriParts2 {
		if uriParts2[idx] != expectedUriParts2[idx] {
			t.Error("Expected item to be " + expectedUriParts2[idx] + " but is " + uriParts2[idx])
		}
	}
}
