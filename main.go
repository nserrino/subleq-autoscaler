/*
Copyright 2018 The Kubernetes Authors.
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

package main

import (
	"log"
	"os"

	"k8s.io/apimachinery/pkg/util/wait"
	basecmd "sigs.k8s.io/custom-metrics-apiserver/pkg/cmd"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
)

// Adapted from the example in this repo: https://github.com/kubernetes-sigs/custom-metrics-apiserver

type subleqAdapter struct {
	basecmd.AdapterBase
	Message string
}

func (a *subleqAdapter) makeProviderOrDie() provider.CustomMetricsProvider {
	client, err := a.DynamicClient()
	if err != nil {
		log.Fatalf("unable to construct dynamic client: %v", err)
	}

	mapper, err := a.RESTMapper()
	if err != nil {
		log.Fatalf("unable to construct discovery REST mapper: %v", err)
	}

	return NewSubleqMetricProvider(client, mapper)
}

func main() {
	cmd := &subleqAdapter{
		Message: "Starting subleq custom metrics adapter",
	}
	cmd.Flags().Parse(os.Args)

	testProvider := cmd.makeProviderOrDie()
	cmd.WithCustomMetrics(testProvider)

	log.Println(cmd.Message)
	if err := cmd.Run(wait.NeverStop); err != nil {
		log.Fatalf("unable to run custom metrics adapter: %v", err)
	}
}
