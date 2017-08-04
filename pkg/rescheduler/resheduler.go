/*
Copyright 2017 The Kubernetes Authors.

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

package rescheduler

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/aveshagarwal/rescheduler/cmd/rescheduler/app/options"
	"github.com/aveshagarwal/rescheduler/pkg/api/v1alpha1"
	"github.com/aveshagarwal/rescheduler/pkg/rescheduler/client"
	eutils "github.com/aveshagarwal/rescheduler/pkg/rescheduler/evictions/utils"
	"github.com/aveshagarwal/rescheduler/pkg/rescheduler/strategies"
)

func Run(rs *options.ReschedulerServer) error {
	rsclient, err := client.CreateClient(rs.KubeconfigFile)
	if err != nil {
		return err
	}
	rs.Client = rsclient

	reschedulerPolicy := v1alpha1.ReschedulerPolicy{}
	if len(rs.PolicyConfigFile) > 0 {
		filename, err := filepath.Abs(rs.PolicyConfigFile)
		if err != nil {
			return err
		}
		fd, err := os.Open(filename)
		if err != nil {
			return err
		}

		if err := yaml.NewYAMLOrJSONDecoder(fd, 4096).Decode(&reschedulerPolicy); err != nil {
			return err
		}

	}
	fmt.Printf("\nreschedulerPolicy: %#v\n", reschedulerPolicy)

	policyGroupVersion, err := eutils.SupportEviction(rs.Client)
	if err != nil || len(policyGroupVersion) == 0 {
		return err
	}

	stopChannel := make(chan struct{})
	nodes, err := nodeutil.ReadyNodes(client, stopChannel)
	if err != nil {
		return err
	}

	strategies.LowNodeUtilization(rs.Client, policyGroupVersion, nodes)
	strategies.RemoveDuplicatePods(rs.Client, policyGroupVersion, nodes)
	return nil
}
