package main

import "testing"

func TestRegA(t *testing.T){
	var r registers
	var data uint8
	data = 200
	if r.setA(data).getA() != data {
		t.Error("Error storing and loading A")
	}
}
func TestRegPC(t *testing.T){
	var r registers
	var data uint16
	data = 0xc600
	if r.setPC(data).getPC() != data {
		t.Error("Error storing and loading PC")
	}
}