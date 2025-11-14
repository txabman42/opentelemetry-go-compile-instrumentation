// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

// Structs for different test cases
type TMethod struct{}
type TStructOnly struct{}
type TMultipleFields struct{}
type TCombined struct{}

// Methods
func (t *TMethod) MethodFunc(p1 string, p2 int) (float32, error) {
	return 0.0, nil
}

// Functions for different test cases
func FuncAfterOnly(p1 string, p2 int) (float32, error) {
	println("FuncAfterOnly")
	return 0.0, nil
}

func FuncBeforeOnly(p1 string, p2 int) (float32, error) {
	println("FuncBeforeOnly")
	return 0.0, nil
}

func FuncRuleOnly(p1 string, p2 int) (float32, error) {
	println("FuncRuleOnly")
	return 0.0, nil
}

func FuncAndRaw(p1 string, p2 int) (float32, error) {
	println("FuncAndRaw")
	return 0.0, nil
}

func FuncRawOnly(p1 string, p2 int) (float32, error) {
	println("FuncRawOnly")
	return 0.0, nil
}

func FuncMultipleHooks(p1 string, p2 int) (float32, error) {
	println("FuncMultipleHooks")
	return 0.0, nil
}

func FuncCombined(p1 string, p2 int) (float32, error) {
	println("FuncCombined")
	return 0.0, nil
}

func main() {
	FuncAfterOnly("test", 1)
	FuncBeforeOnly("test", 2)
	FuncRuleOnly("test", 3)
	FuncAndRaw("test", 4)
	FuncRawOnly("test", 5)
	FuncMultipleHooks("test", 6)
	FuncCombined("test", 7)
	
	t := &TMethod{}
	t.MethodFunc("test", 8)
}
