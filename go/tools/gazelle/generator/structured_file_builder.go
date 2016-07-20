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
	"path/filepath"

	bzl "github.com/bazelbuild/buildifier/core"
)

type structuredFileBuilder struct {
	fs []*bzl.File
}

func (b *structuredFileBuilder) addRules(rel string, rules []*bzl.Rule) {
	f := &bzl.File{Path: filepath.Join(rel, "BUILD")}
	for _, r := range rules {
		f.Stmt = append(f.Stmt, r.Call)
	}
	if load := generateLoad(f); load != nil {
		f.Stmt = append([]bzl.Expr{load}, f.Stmt...)
	}

	b.fs = append(b.fs, f)
}

func (b *flatFileBuilder) isEmpty() bool {
	return len(b.fs) == 0
}

func (b *structuredFileBuilder) files() []*bzl.File {
	return b.fs
}
