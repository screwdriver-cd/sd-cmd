package util

import "testing"

func TestSplitCmd(t *testing.T) {
	// success
	fullCommands := []struct {
		fullCommand  string
		namespaceAns string
		nameAns      string
		versionAns   string
	}{
		{"foo/bar@1.2.*", "foo", "bar", "1.2.*"},
		{"foo/bar@1.2", "foo", "bar", "1.2"},
		{"foo/bar@1.x", "foo", "bar", "1.x"},
		{"foo/bar@1", "foo", "bar", "1"},
		{"foo/bar@~1.2.3", "foo", "bar", "~1.2.3"},
		{"foo/bar@~1.2", "foo", "bar", "~1.2"},
		{"foo/bar@~1", "foo", "bar", "~1"},
		{"foo/bar@^1.2.3", "foo", "bar", "^1.2.3"},
		{"foo/bar@^0.2.5", "foo", "bar", "^0.2.5"},
		{"foo/bar@^0.0.4", "foo", "bar", "^0.0.4"},
		{"foo/bar@1.2.3", "foo", "bar", "1.2.3"},
		{"foo/bar@1.5.3", "foo", "bar", "1.5.3"},
		{"foo/bar@latest", "foo", "bar", "latest"},
		{"foo/bar@stable", "foo", "bar", "stable"},
		{"foo/bar@feature-abc", "foo", "bar", "feature-abc"},
		{"Foo/Bar@feature-abc", "Foo", "Bar", "feature-abc"},
		{"foo/bar@v1.0.0", "foo", "bar", "v1.0.0"},
	}

	for _, c := range fullCommands {
		smallSpec, err := SplitCmd(c.fullCommand)
		if err != nil {
			t.Errorf("%q err=%q, want nil", c.fullCommand, err)
		}
		if smallSpec.Namespace != c.namespaceAns || smallSpec.Name != c.nameAns || smallSpec.Version != c.versionAns {
			t.Errorf("namespace=%q, name=%q, version=%q, want %q, %q, %q", smallSpec.Namespace, smallSpec.Name, smallSpec.Version, c.namespaceAns, c.nameAns, c.versionAns)
		}
	}

	// failure
	fullCommandNames := []string{
		"foo/bar/1.0",
		"foo@bar/1.0",
		"foo@bar@1.0",
		"foobar@1.0",
		"foo/bar1.0",
		"forbar1.0",
		"foo-bar@1.0",
		"foo/bar-1.0",
		"",
	}
	for _, cmdName := range fullCommandNames {
		_, err := SplitCmd(cmdName)
		if err == nil {
			t.Errorf("%q err=nil, want error", cmdName)
		}
	}

	// failure by invalid version
	fullCommandNames = []string{
		"foo/bar@1.0.",
		"foo/bar@*.1.0",
		"foo/bar@1.0.",
		"foo/bar@1.0.1.0",
		"foo/bar@aaa_bbb",
		"foo/bar@-tag",
		"foo/bar@Tag",
	}
	for _, cmdName := range fullCommandNames {
		_, err := SplitCmd(cmdName)
		if err == nil {
			t.Errorf("%q err=nil, want error", cmdName)
		}
	}
}

func TestSplitCmdWithSearch(t *testing.T) {
	// success
	smallSpec, pos, err := SplitCmdWithSearch([]string{"exec", "foo/bar@1.0", "sample"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if smallSpec.Namespace != "foo" || smallSpec.Name != "bar" || smallSpec.Version != "1.0" || pos != 1 {
		t.Errorf("namespace=%q, name=%q, version=%q, want %q, %q, %q", smallSpec.Namespace, smallSpec.Name, smallSpec.Version, "foo", "bar", "1.0")
	}

	// failure
	_, _, err = SplitCmdWithSearch([]string{"exec", "foo-bar-1.0", "sample"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
