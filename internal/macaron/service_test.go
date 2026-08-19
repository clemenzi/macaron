package macaron

import "testing"

func TestParseInstall(t *testing.T) {
	opts, help, err := parseInstall([]string{"--branch", "next", "--skip-build", "--yes", "repo.git"})
	if err != nil || help {
		t.Fatalf("parseInstall error = %v, help = %v", err, help)
	}
	if opts.source != "repo.git" || opts.branch != "next" || !opts.skipBuild || !opts.yes {
		t.Fatalf("unexpected options: %#v", opts)
	}
	if _, _, err := parseInstall([]string{"one", "two"}); err == nil {
		t.Fatal("multiple sources accepted")
	}
	if _, _, err := parseInstall([]string{"--name"}); err == nil {
		t.Fatal("missing option value accepted")
	}
}

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
