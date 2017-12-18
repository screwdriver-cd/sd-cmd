package util

import "testing"

func TestSplitCmd(t *testing.T) {
	// success
	ns, cmd, ver, err := SplitCmd("foo/bar@1.0")
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if ns != "foo" || cmd != "bar" || ver != "1.0" {
		t.Errorf("namespace=%q, command=%q, version=%q, want %q, %q, %q", ns, cmd, ver, "foo", "bar", "1.0")
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
		_, _, _, err := SplitCmd(cmdName)
		if err == nil {
			t.Errorf("err=nil, want error")
		}
	}
}

func TestSplitCmdWithSearch(t *testing.T) {
	// success
	ns, cmd, ver, itr, err := SplitCmdWithSearch([]string{"exec", "foo/bar@1.0", "sample"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if ns != "foo" || cmd != "bar" || ver != "1.0" || itr != 1 {
		t.Errorf("namespace=%q, command=%q, version=%q, want %q, %q, %q", ns, cmd, ver, "foo", "bar", "1.0")
	}

	// failure
	_, _, _, _, err = SplitCmdWithSearch([]string{"exec", "foo-bar-1.0", "sample"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
