package routes

import (
	"log"
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
		log.Fatal("Expected failure for bad url")
	}

	_, err = NewRoute("d", "www.google.com", "t", "team@t.com")
	if err == nil {
		log.Fatal("Expected failure for bad creator")
	}

	_, err = NewRoute("d", "www.google.com", "t@t.com", "team")
	if err == nil {
		log.Fatal("Expected failure for bad team email")
	}
}
