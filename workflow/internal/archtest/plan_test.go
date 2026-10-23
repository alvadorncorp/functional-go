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
	want := []string{"Children", "Kind", "Name", "String"}
	if !slices.Equal(methods, want) {
		t.Fatalf("Plan methods = %v, want %v", methods, want)
	}
	for _, forbidden := range []string{"Execute", "Run"} {
		if _, found := planType.MethodByName(forbidden); found {
			t.Errorf("Plan exposes execution method %s", forbidden)
		}
	}
}

func TestKindHasOnlyStringMethod(t *testing.T) {
	kindType := reflect.TypeFor[plan.Kind]()
	var methods []string
	for i := range kindType.NumMethod() {
		methods = append(methods, kindType.Method(i).Name)
	}
	slices.Sort(methods)
	want := []string{"String"}
	if !slices.Equal(methods, want) {
		t.Fatalf("Kind methods = %v, want %v", methods, want)
	}
}
