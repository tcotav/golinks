package routes

import (
	"log"
	"testing"
)

func TestCreate(t *testing.T) {
	_, err := NewRoute("d", "http://www.google.com", 1, 1)
	if err != nil {
		t.Error(err)
	}

	// bad url format
	_, err = NewRoute("d", "www.google.com", 1, 1)
	if err == nil {
		log.Fatal("Expected failure for bad url")
	}
}