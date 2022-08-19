package main

import (
	"testing"
)

func TestHash160(t *testing.T) {
	var data = []byte("you")
	var result = Hash160(data)
	println("rr:", result)
}
