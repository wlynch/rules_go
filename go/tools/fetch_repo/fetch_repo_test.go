package main

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/vcs"
)

var (
	root = &vcs.RepoRoot{
		VCS:  vcs.ByCmd("git"),
		Repo: "https://github.com/bazeltest/rules_go",
		Root: "github.com/bazeltest/rules_go",
	}
)

func TestGetRepoRoot(t *testing.T) {
	for _, tc := range []struct {
		label      string
		remote     string
		cmd        string
		importpath string
		r          *vcs.RepoRoot
	}{
		{
			label:      "all",
			remote:     "https://github.com/bazeltest/rules_go",
			cmd:        "git",
			importpath: "github.com/bazeltest/rules_go",
			r:          root,
		},
		{
			label:      "different remote",
			remote:     "https://example.com/rules_go",
			cmd:        "git",
			importpath: "github.com/bazeltest/rules_go",
			r: &vcs.RepoRoot{
				VCS:  vcs.ByCmd("git"),
				Repo: "https://example.com/rules_go",
				Root: "github.com/bazeltest/rules_go",
			},
		},
		{
			label:      "only importpath",
			importpath: "github.com/bazeltest/rules_go",
			r:          root,
		},
		{
			label:      "missing vcs",
			remote:     "https://github.com/bazeltest/rules_go",
			importpath: "github.com/bazeltest/rules_go",
			r:          root,
		},
		{
			label:      "missing remote",
			cmd:        "git",
			importpath: "github.com/bazeltest/rules_go",
			r:          root,
		},
		{
			label:  "old args",
			remote: "github.com/bazeltest/rules_go",
			r:      root,
		},
	} {
		r, err := getRepoRoot(tc.remote, tc.cmd, tc.importpath)
		if err != nil {
			t.Errorf("[%s] %v", tc.label, err)
		}
		if !reflect.DeepEqual(r, tc.r) {
			t.Errorf("[%s] Expected %+v, got %+v", tc.label, tc.r, r)
		}
	}
}
