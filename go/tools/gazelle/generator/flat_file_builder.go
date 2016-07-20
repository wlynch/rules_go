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

type flatFileBuilder struct {
	f bzl.File
}

func (b *flatFileBuilder) addRules(rel string, rules []*bzl.Rule) {
	for _, r := range rules {
		b.f.Stmt = append(b.f.Stmt, r.Call)
	}
}

func (b *flatFileBuilder) isEmpty() bool {
	return len(b.f.Stmt) == 0
}

func (b *flatFileBuilder) files() []*bzl.File {
	b.f.Path = "BUILD"
	if load := generateLoad(&b.f); load != nil {
		b.f.Stmt = append([]bzl.Expr{load}, b.f.Stmt...)
	}
	return []*bzl.File{&b.f}
}
