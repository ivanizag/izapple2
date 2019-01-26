package main

import "testing"

func TestRegA(t *testing.T) {
	var r registers
	var data uint8
	data = 200
	r.setA(data)
	if r.getA() != data {
		t.Error("Error storing and loading A")
	}
}
func TestRegPC(t *testing.T) {
	var r registers
	var data uint16
	data = 0xc600
	r.setPC(data)
	if r.getPC() != data {
		t.Error("Error storing and loading PC")
	}
}
