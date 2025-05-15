package ddd

import (
	"testing"
)

// TestResourceWithInit is a test resource that implements ResourceType
type TestResourceWithInit struct {
	Handler     TestDependency `resource:"anotherFieldName"`
	initialized bool
}

// TestDependency is a separate interface for dependencies
type TestDependency interface {
	DoSomething()
}

// TestDependencyImpl implements TestDependency
type TestDependencyImpl struct{}

func (t TestDependencyImpl) DoSomething() {}

func (t TestResourceWithInit) DoSomething() {}

func (t *TestResourceWithInit) OnInit() (*Resource, error) {
	t.initialized = true
	return nil, nil
}

func (t TestResourceWithInit) OnStart()   {}
func (t TestResourceWithInit) OnDestroy() {}

func TestContextWithResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []*Resource
		wantPanic bool
		checkInit bool
	}{
		{
			name: "valid resources",
			resources: []*Resource{
				NewResource[TestDependency](TestDependencyImpl{}, "anotherFieldName"),
				NewResource[TestInterface](TestResourceWithInit{}),
			},
			wantPanic: false,
		},
		{
			name:      "empty resources",
			resources: []*Resource{},
			wantPanic: false,
		},
		{
			name:      "nil resources",
			resources: nil,
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("WithResources() panic = %v, want panic = %v", r, tt.wantPanic)
				}
			}()

			ctx := NewContext("test")
			result := ctx.WithResources(tt.resources)

			// Check if context is returned
			if result != ctx {
				t.Error("WithResources() did not return the same context")
			}
		})
	}
}
