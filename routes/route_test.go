package routes

import (
	"encoding/json"
	"testing"
)

func TestCreate(t *testing.T) {
	_, err := NewRoute("d", "http://www.google.com", "test@testuser.com", "testteam@t.com")
	if err != nil {
		t.Error(err)
	}

	// bad url format
	_, err = NewRoute("d", "www.google.com", "t@t.com", "team@t.com")
	if err == nil {
		t.Error("Expected failure for bad url")
	}

	_, err = NewRoute("d", "www.google.com", "t", "team@t.com")
	if err == nil {
		t.Error("Expected failure for bad creator")
	}

	_, err = NewRoute("d", "www.google.com", "t@t.com", "team")
	if err == nil {
		t.Error("Expected failure for bad team email")
	}
}

func TestJSONMarshal(t *testing.T) {
	sValid := `{"shortkey": "google", "url":"http://www.google.com", "creator":"t@t.com"}`
	sBadCreator := `{"shortkey": "google", "url":"http://www.google.com", "creator":"t@tcom"}`

	var r Route
	if err := json.Unmarshal([]byte(sValid), &r); err != nil {
		t.Error(err.Error())
	}
	r = Route{}
	if err := json.Unmarshal([]byte(sBadCreator), &r); err == nil {
		t.Error("Expected error for bad email format of creator")
	}
}
