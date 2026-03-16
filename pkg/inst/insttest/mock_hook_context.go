// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package insttest provides test helpers for instrumentation hook testing.
package insttest

// MockHookContext is a shared test implementation of inst.HookContext.
type MockHookContext struct {
	Params      []interface{}
	ReturnVals  []interface{}
	Data        map[string]interface{}
	SkipCall    bool
	FuncName    string
	PackageName string
}

// NewMockHookContext creates a new MockHookContext with optional initial params.
func NewMockHookContext(params ...interface{}) *MockHookContext {
	return &MockHookContext{
		Params:      params,
		Data:        make(map[string]interface{}),
		FuncName:    "mockFunc",
		PackageName: "mock",
	}
}

func (m *MockHookContext) SetSkipCall(skip bool) { m.SkipCall = skip }
func (m *MockHookContext) IsSkipCall() bool      { return m.SkipCall }

func (m *MockHookContext) SetData(data interface{})          { m.Data["_default"] = data }
func (m *MockHookContext) GetData() interface{}              { return m.Data["_default"] }
func (m *MockHookContext) SetKeyData(key string, val interface{}) { m.Data[key] = val }

// GetKeyData returns the value for key. It first checks the direct Data map, then
// falls back to inspecting Data["_default"] when it holds a map[string]interface{},
// which matches the pattern used by nethttp hooks that call SetData(aMap) and then
// retrieve individual fields via GetKeyData.
func (m *MockHookContext) GetKeyData(key string) interface{} {
	if val, ok := m.Data[key]; ok {
		return val
	}
	if defaultVal, ok := m.Data["_default"]; ok {
		if dataMap, ok := defaultVal.(map[string]interface{}); ok {
			return dataMap[key]
		}
	}
	return nil
}

// HasKeyData returns true when key exists in the direct Data map or inside
// Data["_default"] when it holds a map[string]interface{}.
func (m *MockHookContext) HasKeyData(key string) bool {
	if _, ok := m.Data[key]; ok {
		return true
	}
	if defaultVal, ok := m.Data["_default"]; ok {
		if dataMap, ok := defaultVal.(map[string]interface{}); ok {
			_, ok := dataMap[key]
			return ok
		}
	}
	return false
}

func (m *MockHookContext) GetParamCount() int { return len(m.Params) }
func (m *MockHookContext) GetParam(idx int) interface{} {
	if idx < 0 || idx >= len(m.Params) {
		return nil
	}
	return m.Params[idx]
}
func (m *MockHookContext) SetParam(idx int, val interface{}) {
	for len(m.Params) <= idx {
		m.Params = append(m.Params, nil)
	}
	m.Params[idx] = val
}

func (m *MockHookContext) GetReturnValCount() int { return len(m.ReturnVals) }
func (m *MockHookContext) GetReturnVal(idx int) interface{} {
	if idx < 0 || idx >= len(m.ReturnVals) {
		return nil
	}
	return m.ReturnVals[idx]
}
func (m *MockHookContext) SetReturnVal(idx int, val interface{}) {
	for len(m.ReturnVals) <= idx {
		m.ReturnVals = append(m.ReturnVals, nil)
	}
	m.ReturnVals[idx] = val
}

func (m *MockHookContext) GetFuncName() string    { return m.FuncName }
func (m *MockHookContext) GetPackageName() string { return m.PackageName }
