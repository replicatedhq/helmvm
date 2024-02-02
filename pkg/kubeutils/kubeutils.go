package kubeutils

import (
	"context"
	"fmt"
	"time"

	"github.com/k0sproject/k0s/pkg/apis/helm/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BackOffToDuration returns the maximum duration of the provided backoff.
func BackOffToDuration(backoff wait.Backoff) time.Duration {
	var total time.Duration
	duration := backoff.Duration
	for i := 0; i < backoff.Steps; i++ {
		total += duration
		duration = time.Duration(float64(duration) * backoff.Factor)
	}
	return total
}

// WaitForChart waits for the provided helm chart to be applied on the cluster. Applied does not
// mean fully processed, it means that k0s has applied the chart and it is now being processed by
// the cluster.
func WaitForChart(ctx context.Context, cli client.Client, ns, name string) error {
	backoff := wait.Backoff{Steps: 60, Duration: 5 * time.Second, Factor: 1.0, Jitter: 0.1}
	nsn := types.NamespacedName{Namespace: ns, Name: name}
	if err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		var chart v1beta1.Chart
		if err := cli.Get(ctx, nsn, &chart); err != nil {
			return false, err
		}
		return chart.Status.ReleaseName != "", nil
	}); err != nil {
		return fmt.Errorf("timed out waiting for chart %s: %v", name, err)
	}
	return nil
}

// WaitForDeployment waits for the provided deployment to be ready.
func WaitForDeployment(ctx context.Context, cli client.Client, ns, name string) error {
	backoff := wait.Backoff{Steps: 60, Duration: 5 * time.Second, Factor: 1.0, Jitter: 0.1}
	var lasterr error
	if err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		ready, err := IsDeploymentReady(ctx, cli, ns, name)
		if err != nil {
			lasterr = fmt.Errorf("unable to get deploy %s status: %v", name, err)
			return false, nil
		}
		return ready, nil
	}); err != nil {
		return fmt.Errorf("timed out waiting for deploy %s: %v", name, lasterr)
	}
	return nil
}

// IsDeploymentReady returns true if the deployment is ready.
func IsDeploymentReady(ctx context.Context, cli client.Client, ns, name string) (bool, error) {
	var deploy appsv1.Deployment
	nsn := types.NamespacedName{Namespace: ns, Name: name}
	if err := cli.Get(ctx, nsn, &deploy); err != nil {
		return false, err
	}
	if deploy.Spec.Replicas == nil {
		return false, nil
	}
	return deploy.Status.ReadyReplicas == *deploy.Spec.Replicas, nil
}

// IsStatefulSetReady returns true if the statefulset is ready.
func IsStatefulSetReady(ctx context.Context, cli client.Client, ns, name string) (bool, error) {
	var statefulset appsv1.StatefulSet
	nsn := types.NamespacedName{Namespace: ns, Name: name}
	if err := cli.Get(ctx, nsn, &statefulset); err != nil {
		return false, err
	}
	if statefulset.Spec.Replicas == nil {
		return false, nil
	}
	return statefulset.Status.ReadyReplicas == *statefulset.Spec.Replicas, nil
}
