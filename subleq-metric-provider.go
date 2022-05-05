package main

import (
	"context"
	"fmt"
	"time"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider"
	"sigs.k8s.io/custom-metrics-apiserver/pkg/provider/helpers"
)

// Adapted from the example in this repo: https://github.com/kubernetes-sigs/custom-metrics-apiserver

// The name of the metric we are creating.
const subleqMetricName = "subleq-metric"
const targetMetricValue = 1.0

type subleqMetricsProvider struct {
	client          dynamic.Interface
	mapper          apimeta.RESTMapper
	programs        map[string]*SubleqProgram
	numPodsForApp         map[string]int
}

// Backwards calculates the target metric value from the metricGoal in order to produce the
// expected number of pods as the program output.
// https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#algorithm-details
func getTargetMetricValue(metricGoal float64, currentPods int, targetPods int) float64 {
	// Solve for currentMetricValue:
  // desiredReplicas = ceil[currentReplicas * ( currentMetricValue / desiredMetricValue )]
	return float64(targetPods) * metricGoal / float64(currentPods)
}

// NewSubleqMetricProvider returns an instance of the subleq metrics provider.
func NewSubleqMetricProvider(k8sClient dynamic.Interface, mapper apimeta.RESTMapper) provider.CustomMetricsProvider {
	provider := &subleqMetricsProvider{
		client:          k8sClient,
		mapper:          mapper,
		programs:        make(map[string]*SubleqProgram),
		numPodsForApp:   make(map[string]int),
	}
	return provider
}

func (p *subleqMetricsProvider) metricFor(value float64, name types.NamespacedName, info provider.CustomMetricInfo) (*custom_metrics.MetricValue, error) {
	// construct a reference referring to the described object
	objRef, err := helpers.ReferenceFor(p.mapper, name, info)
	if err != nil {
		return nil, err
	}

	return &custom_metrics.MetricValue{
		DescribedObject: objRef,
		Metric: custom_metrics.MetricIdentifier{
			Name: info.Metric,
		},
		Timestamp: metav1.Time{time.Now()},
		Value:     *resource.NewMilliQuantity(int64(value*1000), resource.DecimalSI),
	}, nil
}

// Always add 2 to the number of output pods.
// 1 = -1 (terminated)
// 2 = 0 (empty output)
// 3+ = non empty output
func getNumPodsFromOutput(output int) int {
	return output + 2
}

// GetMetricByName returns the metric for a given name.
func (p *subleqMetricsProvider) GetMetricByName(ctx context.Context, name types.NamespacedName, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValue, error) {
	if info.Metric != subleqMetricName || info.GroupResource.Resource != "pods" {
		return nil, provider.NewMetricNotFoundError(info.GroupResource, info.Metric)
	}
	res, err := helpers.ResourceFor(p.mapper, info)
	if err != nil {
		return nil, err
	}
	resClient := p.client.Resource(res).Namespace(name.Namespace)
	getResult, err := resClient.Get(ctx, name.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if getResult == nil || getResult.GetLabels() == nil {
		return nil, fmt.Errorf("Pod %s/%s is missing 'name' label", name.Namespace, name.Name)
	}
	appName, ok := getResult.GetLabels()["name"]
	if appName == "" || !ok {
		return nil, fmt.Errorf("Pod %s/%s is missing 'name' label", name.Namespace, name.Name)
	}
	program, ok := p.programs[appName]
	if !ok {
		return nil, fmt.Errorf("Pod %s/%s has not been associated with an autoscaled resource", name.Namespace, name.Name)
	}
	currNumPods, ok := p.numPodsForApp[appName]
	if !ok {
		return nil, fmt.Errorf("Pod %s/%s has not been associated with an autoscaled resource", name.Namespace, name.Name)
	}
	metric := getTargetMetricValue(1.0, currNumPods, getNumPodsFromOutput(program.LastValue))
	return p.metricFor(metric, name, info)
}

// GetMetricBySelector returns the metric for a given selector.
func (p *subleqMetricsProvider) GetMetricBySelector(ctx context.Context, namespace string, selector labels.Selector, info provider.CustomMetricInfo, metricSelector labels.Selector) (*custom_metrics.MetricValueList, error) {
	names, err := helpers.ListObjectNames(p.mapper, p.client, namespace, selector, info)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		 return &custom_metrics.MetricValueList{
			 Items: nil,
		 }, nil
	}

	res, err := helpers.ResourceFor(p.mapper, info)
	if err != nil {
		return nil, err
	}
	resClient := p.client.Resource(res).Namespace(namespace)
	getResult, err := resClient.Get(ctx, names[0], metav1.GetOptions{})
	// Requires the deployment to have "name" set.
	appName, ok := getResult.GetLabels()["name"]
	if appName == "" || !ok {
		return nil, fmt.Errorf("Pod %s/%s is missing 'name' label", namespace, names[0])
	}

	program, ok := p.programs[appName]
	if !ok {
		program = CreateSubleqProgram(appName)
		p.programs[appName] = program
	}
	// Current number of pods
	p.numPodsForApp[appName] = len(names)
	// Execute a single instruction. This value will be cached by the program
	// and accessed by the calls to p.GetMetricForName.
	output := program.GetNextOutputValue()
	fmt.Printf("Output value for %s is %d (step %d)\n", appName, output, program.Step)
	metrics := make([]custom_metrics.MetricValue, 0, len(names))

	for _, name := range names {
		namespacedName := types.NamespacedName{Name: name, Namespace: namespace}
		value, err := p.GetMetricByName(ctx, namespacedName, info, metricSelector)
		if err != nil || value == nil {
			if apierr.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		metrics = append(metrics, *value)
	}

	return &custom_metrics.MetricValueList{
		Items: metrics,
	}, nil
}

// ListAllMetrics returns the single metric defined by this provider.
func (p *subleqMetricsProvider) ListAllMetrics() []provider.CustomMetricInfo {
	return []provider.CustomMetricInfo{
		{
			GroupResource: schema.GroupResource{Group: "", Resource: "pods"},
			Metric:        subleqMetricName,
			Namespaced:    true,
		},
	}
}
