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

package generator

import (
	bzl "github.com/bazelbuild/buildifier/core"
)

const (
	// goRulesBzl is the label of the Skylark file which provides Go rules
	goRulesBzl = "@io_bazel_rules_go//go:def.bzl"
)

type fileBuilder interface {
	addRules(rel string, rules []*bzl.Rule)
	isEmpty() bool
	files() []*bzl.File
}

func generateLoad(f *bzl.File) bzl.Expr {
	var list []string
	for _, kind := range []string{
		"go_prefix",
		"go_library",
		"go_binary",
		"go_test",
		// TODO(yugui): Support cgo_library
	} {
		if len(f.Rules(kind)) > 0 {
			list = append(list, kind)
		}
	}
	if len(list) == 0 {
		return nil
	}
	return loadExpr(goRulesBzl, list...)
}

func loadExpr(ruleFile string, rules ...string) bzl.Expr {
	var list []bzl.Expr
	for _, r := range append([]string{ruleFile}, rules...) {
		list = append(list, &bzl.StringExpr{Value: r})
	}

	return &bzl.CallExpr{
		X:            &bzl.LiteralExpr{Token: "load"},
		List:         list,
		ForceCompact: true,
	}
}
