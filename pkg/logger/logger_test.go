package logger

import (
	"testing"
)

func TestInit_Development(_ *testing.T) {
	Init("development")
	// Should not panic, uses text handler
}

func TestInit_Production(_ *testing.T) {
	Init("production")
	// Should not panic, uses JSON handler
}

func TestInit_Empty(_ *testing.T) {
	Init("")
	// Should default to text handler
}
