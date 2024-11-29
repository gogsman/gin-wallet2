package main

import (
	"testing"
)

func TestInitDB(t *testing.T) {
	db1 := InitDB()
	if db1 == nil {
		t.Error("Expected database connection, got nil")
	}

	db2 := InitDB()
	if db2 == nil {
		t.Error("Expected database connection, got nil")
	}

	if db1 != db2 {
		t.Error("Expected same database instance, got different instances")
	}

	err := db1.Ping()
	if err != nil {
		t.Errorf("Database connection is not valid: %v", err)
	}

	// Clean up
	defer db1.Close()
}
