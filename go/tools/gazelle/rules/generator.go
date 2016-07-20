/* Copyright 2016 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rules

import (
	"fmt"
	"go/build"
	"path"
	"strings"

	bzl "github.com/bazelbuild/buildifier/core"
)

const (
	// defaultLibName is the name of the default go_library rule in a Go
	// package directory. It must be consistent to _DEFAULT_LIB in go/def.bzl.
	defaultLibName = "go_default_library"
	// defaultTestName is a name of an internal test corresponding to
	// defaultLibName. It does not need to be consistent to something but it
	// just needs to be unique in the Bazel package
	defaultTestName = "go_default_test"
	// defaultXTestName is a name of an external test corresponding to
	// defaultLibName.
	defaultXTestName = "go_default_xtest"
)

// Generator generates Bazel build rules for Go build targets
type Generator interface {
	// Generate generates build rules for build targets in a Go package in a
	// repository.
	//
	// "rel" is a relative slash-separated path from the repostiry root
	// directory to the Go package directory. It is empty if the package
	// directory is the repository root itself.
	// "pkg" is a description about the package.
	Generate(rel string, pkg *build.Package) ([]*bzl.Rule, error)
}

// A Style describes a strategy of organizing BUILD files
type Style int

const (
	// StructuredStyle means that individual Go packages will have their own Bazel packages.
	StructuredStyle = Style(iota) + 1
	// FlatStyle means that the caller concatenates Go packages under a specific go_prefix
	// into a single BUILD file.
	FlatStyle
)

// NewGenerator returns an implementation of Generator.
//
// "goPrefix" is the go_prefix corresponding to the repository root.
// See also https://github.com/bazelbuild/rules_go#go_prefix.
func NewGenerator(goPrefix string, s Style) Generator {
	var (
		r      labelResolver
		refSrc func(rel string, srcs []string) []string

		e externalResolver
	)

	switch s {
	case StructuredStyle:
		r = structuredResolver{goPrefix: goPrefix}
		refSrc = func(rel string, srcs []string) []string { return srcs }
	case FlatStyle:
		r = flatResolver{goPrefix: goPrefix}
		refSrc = func(rel string, srcs []string) []string {
			var ret []string
			for _, s := range srcs {
				ret = append(ret, path.Join(rel, s))
			}
			return ret
		}
	default:
		panic(fmt.Sprintf("unrecognized style: %d", s))
	}

	return &generator{
		goPrefix: goPrefix,
		r: resolverFunc(func(importpath, dir string) (label, error) {
			if importpath != goPrefix && !strings.HasPrefix(importpath, goPrefix+"/") && !strings.HasPrefix(importpath, "./") {
				return e.resolve(importpath, dir)
			}
			return r.resolve(importpath, dir)
		}),
		refSrc: refSrc,
	}
}

type generator struct {
	goPrefix string
	r        labelResolver
	refSrc   func(rel string, srcs []string) []string
}

func (g *generator) Generate(rel string, pkg *build.Package) ([]*bzl.Rule, error) {
	var rules []*bzl.Rule
	if rel == "" {
		p, err := newRule("go_prefix", []interface{}{g.goPrefix}, nil)
		if err != nil {
			return nil, err
		}
		rules = append(rules, p)
	}

	r, err := g.generate(rel, pkg)
	if err != nil {
		return nil, err
	}
	rules = append(rules, r)

	if len(pkg.TestGoFiles) > 0 {
		t, err := g.generateTest(rel, pkg, r.AttrString("name"))
		if err != nil {
			return nil, err
		}
		rules = append(rules, t)
	}

	if len(pkg.XTestGoFiles) > 0 {
		t, err := g.generateXTest(rel, pkg, r.AttrString("name"))
		if err != nil {
			return nil, err
		}
		rules = append(rules, t)
	}
	return rules, nil
}

func (g *generator) generate(rel string, pkg *build.Package) (*bzl.Rule, error) {
	l, err := g.r.resolve(path.Join(g.goPrefix, rel), "")
	if err != nil {
		return nil, err
	}
	name := l.name
	kind := "go_library"
	if pkg.IsCommand() {
		kind = "go_binary"
		name = path.Base(pkg.Dir)
	}

	visibility := "//visibility:public"
	if i := strings.LastIndex(rel, "/internal/"); i >= 0 {
		visibility = fmt.Sprintf("//%s:__subpackages__", rel[:i])
	} else if strings.HasPrefix(rel, "internal/") {
		visibility = "//:__subpackages__"
	}

	attrs := []keyvalue{
		{key: "name", value: name},
		{key: "srcs", value: g.refSrc(rel, pkg.GoFiles)},
		{key: "visibility", value: []string{visibility}},
	}

	deps, err := g.dependencies(pkg.Imports, rel)
	if err != nil {
		return nil, err
	}
	if len(deps) > 0 {
		attrs = append(attrs, keyvalue{key: "deps", value: deps})
	}

	return newRule(kind, nil, attrs)
}

func (g *generator) generateTest(rel string, pkg *build.Package, library string) (*bzl.Rule, error) {
	name := library + "_test"
	if library == defaultLibName {
		name = defaultTestName
	}
	attrs := []keyvalue{
		{key: "name", value: name},
		{key: "srcs", value: g.refSrc(rel, pkg.TestGoFiles)},
		{key: "library", value: ":" + library},
	}

	deps, err := g.dependencies(pkg.TestImports, rel)
	if err != nil {
		return nil, err
	}
	if len(deps) > 0 {
		attrs = append(attrs, keyvalue{key: "deps", value: deps})
	}
	return newRule("go_test", nil, attrs)
}

func (g *generator) generateXTest(rel string, pkg *build.Package, library string) (*bzl.Rule, error) {
	name := library + "_xtest"
	if library == defaultLibName {
		name = defaultXTestName
	}
	attrs := []keyvalue{
		{key: "name", value: name},
		{key: "srcs", value: g.refSrc(rel, pkg.XTestGoFiles)},
	}

	deps, err := g.dependencies(pkg.XTestImports, rel)
	if err != nil {
		return nil, err
	}
	attrs = append(attrs, keyvalue{key: "deps", value: deps})
	return newRule("go_test", nil, attrs)
}

func (g *generator) dependencies(imports []string, dir string) ([]string, error) {
	var deps []string
	for _, p := range imports {
		if isStandard(p) {
			continue
		}
		l, err := g.r.resolve(p, dir)
		if err != nil {
			return nil, err
		}
		deps = append(deps, l.String())
	}
	return deps, nil
}

// isStandard determines if importpath points a Go standard package.
func isStandard(importpath string) bool {
	seg := strings.SplitN(importpath, "/", 2)[0]
	return !strings.Contains(seg, ".")
}
