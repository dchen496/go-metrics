package metrics

import (
	"testing"
)

type testRegistryType struct{}

func testRegistryInitialize() *Registry {
	r := NewRegistry("testRegistry")

	t := testRegistryType{}
	r.NewCounter(t, "internal_type")
	r.NewDistribution(&t, "ptr_to_internal_type")
	r.NewGauge(testTime, "external_type")
	r.NewMeter(&testTime, "ptr_to_external_type")
	return r
}

func TestRegistryListMetrics(t *testing.T) {
	r := testRegistryInitialize()
	m := r.ListMetrics()
	expectedNames := []string{
		"metrics.testRegistryType.internal_type",
		"metrics.testRegistryType.ptr_to_internal_type",
		"time.Time.external_type",
		"time.Time.ptr_to_external_type",
	}
	var fail bool
	for _, name := range expectedNames {
		if m[name] == nil {
			t.Errorf("List did not include name %s", name)
			fail = true
		}
	}
	if fail {
		t.Errorf("Got %v", m)
		t.Errorf("Expected names are %v", expectedNames)
	}
}

func TestRegistryFind(t *testing.T) {
	r := testRegistryInitialize()
	if r.Find(testTime, "external_type") == nil {
		t.Errorf("Type time.Time and name external_type should have a metric")
	}
	if r.Find(testTime, "ptr_to_external_type") == nil {
		t.Errorf("Type *time.Time and name ptr_to_external_type " +
			"should have a metric, accessible with time.Time")
	}
	if r.Find(testRegistryInitialize, "internal_type") != nil {
		t.Errorf("Type func() *Registry should not have a metric")
	}
}

func TestRegistryFindS(t *testing.T) {
	r := testRegistryInitialize()
	if r.FindS("metrics.testRegistryType.internal_type") == nil {
		t.Errorf("Type metrics.testRegistryType and name internal_type" +
			"should have a metric")
	}
}
