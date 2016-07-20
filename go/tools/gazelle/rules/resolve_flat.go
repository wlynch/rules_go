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
	"path"
	"strings"
)

// flatResolver resolves go_library labels within the same repository as
// the one of goPrefix, assuming all rules are defined in a single BUILD file.
type flatResolver struct {
	goPrefix string
}

func (r flatResolver) resolve(importpath, dir string) (label, error) {
	if strings.HasPrefix(importpath, "./") {
		importpath = path.Join(r.goPrefix, dir, importpath[2:])
	}

	if importpath == r.goPrefix {
		return label{name: "go_default_library", relative: true}, nil
	}

	if prefix := r.goPrefix + "/"; strings.HasPrefix(importpath, prefix) {
		return label{
			name:     strings.TrimPrefix(importpath, prefix),
			relative: true,
		}, nil
	}

	return label{}, fmt.Errorf("importpath %q does not start with goPrefix %q", importpath, r.goPrefix)
}
