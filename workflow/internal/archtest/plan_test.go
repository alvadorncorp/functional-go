package archtest

import (
	"reflect"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/plan"
)

func TestPlanIsOpaqueAndNonExecutable(t *testing.T) {
	planType := reflect.TypeFor[plan.Plan]()
	for i := range planType.NumField() {
		field := planType.Field(i)
		if field.IsExported() {
			t.Errorf("Plan exports field %s", field.Name)
		}
	}

	var methods []string
	for i := range planType.NumMethod() {
		methods = append(methods, planType.Method(i).Name)
	}
	slices.Sort(methods)
	want := []string{"Children", "Kind", "Name"}
	if !slices.Equal(methods, want) {
		t.Fatalf("Plan methods = %v, want %v", methods, want)
	}
	for _, forbidden := range []string{"Execute", "Run"} {
		if _, found := planType.MethodByName(forbidden); found {
			t.Errorf("Plan exposes execution method %s", forbidden)
		}
	}
}
