workspace(name = "io_bazel_rules_go")

load("//go:def.bzl", "go_repositories", "new_go_repository")
load("//go/private:go_repositories.bzl", "go_internal_tools_deps")

go_repositories()

new_go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b4da8ac2ed8007337bc5eb5007998",
    # DO NOT specify this internal attribute outside of rules_go project itself.
    go_rules_bzl_only_for_internal_use = "@//go:def.bzl",
    importpath = "github.com/golang/glog",
)

go_internal_tools_deps()

# Protocol buffers

load("//proto:go_proto_library.bzl", "go_proto_repositories")

go_proto_repositories(
    # DO NOT specify this internal attribute outside of rules_go project itself.
    go_rules_bzl_only_for_internal_use = "@//go:def.bzl",
)
