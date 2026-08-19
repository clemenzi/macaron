package macaron

import "testing"

func TestValidServiceName(t *testing.T) {
	for _, valid := range []string{"web", "my-service", "service.txt"} {
		if !validServiceName(valid) {
			t.Errorf("%q should be valid", valid)
		}
	}
	for _, invalid := range []string{"", ".", "..", "../web", "a/b", `a\\b`} {
		if validServiceName(invalid) {
			t.Errorf("%q should be invalid", invalid)
		}
	}
}
