local env = std.extVar("__ksonnet/environments");
local params = std.extVar("__ksonnet/params").components["mxnet-operator"];

local k = import "k.libsonnet";
local operator = import "kubeflow/mxnet-job/mxnet-operator.libsonnet";

std.prune(k.core.v1.list.new(operator.all(params, env)))
