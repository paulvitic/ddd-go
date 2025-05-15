package ddd

import (
	"testing"
)

// TestResource is a test struct with tagged dependencies
type TestResource struct {
	Name     string
	Handler  TestInterface `resource:"another_filed_name"`
	NoTagDep any
}

// DoSomething implements TestInterface
func (t TestResource) DoSomething() {}

// TestInterface is a test interface
type TestInterface interface {
	DoSomething()
}

func TestResourceCreation(t *testing.T) {
	tests := []struct {
		name        string
		value       any
		givenName   []string
		wantName    string
		wantType    string
		wantDeps    []Dependency
		shouldPanic bool
	}{
		{
			name:      "with provided name",
			value:     TestResource{},
			givenName: []string{"custom_name"},
			wantName:  "custom_name",
			wantType:  "domain.TestInterface",
			wantDeps: []Dependency{
				{
					FieldName:    "Handler",
					ResourceName: "another_filed_name",
				},
			},
			shouldPanic: false,
		},
		{
			name:      "without name - should use struct name",
			value:     TestResource{},
			givenName: nil,
			wantName:  "testResource",
			wantType:  "domain.TestInterface",
			wantDeps: []Dependency{
				{
					FieldName:    "Handler",
					ResourceName: "another_filed_name",
				},
			},
			shouldPanic: false,
		},
		{
			name:        "non-struct value should panic",
			value:       "string",
			givenName:   nil,
			shouldPanic: true,
		},
		{
			name:        "value not implementing interface should panic",
			value:       struct{}{},
			givenName:   nil,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.shouldPanic {
					t.Errorf("Resource() panic = %v, want panic = %v", r, tt.shouldPanic)
				}
			}()

			got := NewResource[TestInterface](tt.value, tt.givenName...)

			if !tt.shouldPanic {
				// Check name
				if got.ResourceName != tt.wantName {
					t.Errorf("ResourceName = %v, want %v", got.ResourceName, tt.wantName)
				}

				// Check type
				if got.ResourceType != tt.wantType {
					t.Errorf("ResourceType = %v, want %v", got.ResourceType, tt.wantType)
				}

				// Check value
				if got.Value != tt.value {
					t.Errorf("Value = %v, want %v", got.Value, tt.value)
				}

				// Check dependencies
				if len(got.Dependencies) != len(tt.wantDeps) {
					t.Errorf("Dependencies length = %v, want %v", len(got.Dependencies), len(tt.wantDeps))
				}

				for i, dep := range got.Dependencies {
					if dep.FieldName != tt.wantDeps[i].FieldName {
						t.Errorf("Dependency[%d].FieldName = %v, want %v", i, dep.FieldName, tt.wantDeps[i].FieldName)
					}
					if dep.ResourceName != tt.wantDeps[i].ResourceName {
						t.Errorf("Dependency[%d].ResourceName = %v, want %v", i, dep.ResourceName, tt.wantDeps[i].ResourceName)
					}
				}
			}
		})
	}
}
