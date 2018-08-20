// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generator

import (
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	LabelGroupName = "group_name"
	labelMXJobName = "mx_job_name"
)

var (
	errPortNotFound = fmt.Errorf("Failed to found the port")
)

func GenGeneralName(mxJobName, rtype, index string) string {
	n := mxJobName + "-" + rtype + "-" + index
	return strings.Replace(n, "/", "-", -1)
}

func GenDNSRecord(mxJobName, rtype, index, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", GenGeneralName(mxJobName, rtype, index), namespace)
}

