package service

import "testing"

func TestValidName(t *testing.T) {
	for _, valid := range []string{"web", "my-service", "service.txt"} {
		if !ValidName(valid) {
			t.Errorf("%q should be valid", valid)
		}
	}
	for _, invalid := range []string{"", ".", "..", "../web", "a/b", `a\\b`} {
		if ValidName(invalid) {
			t.Errorf("%q should be invalid", invalid)
		}
	}
}
