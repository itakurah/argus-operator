package controller

import (
	"context"
	"errors"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	rolloutPollInterval = 2 * time.Second
	rolloutPollTimeout  = 15 * time.Minute

	rolloutKindDeployment  = "Deployment"
	rolloutKindStatefulSet = "StatefulSet"
	rolloutKindDaemonSet   = "DaemonSet"

	deploymentProgressingReasonNewRS = "NewReplicaSetAvailable"
)

// errWorkloadDeleted stops the poll when the workload object no longer exists
var errWorkloadDeleted = errors.New("workload was deleted")

// waitAndLogRollout polls workload status in the background
func waitAndLogRollout(ctx context.Context, c client.Client, nn types.NamespacedName, generation int64, kind string) {
	waitCtx, cancel := context.WithTimeout(ctx, rolloutPollTimeout)
	defer cancel()

	log := ctrl.Log.WithName("rollout")
	err := wait.PollUntilContextTimeout(waitCtx, rolloutPollInterval, rolloutPollTimeout, true, func(ctx context.Context) (bool, error) {
		switch kind {
		case rolloutKindDeployment:
			dep := &appsv1.Deployment{}
			if err := c.Get(ctx, nn, dep); err != nil {
				if apierrors.IsNotFound(err) {
					return false, errWorkloadDeleted
				}
				return false, err
			}
			return deploymentRolloutComplete(dep, generation), nil
		case rolloutKindStatefulSet:
			sts := &appsv1.StatefulSet{}
			if err := c.Get(ctx, nn, sts); err != nil {
				if apierrors.IsNotFound(err) {
					return false, errWorkloadDeleted
				}
				return false, err
			}
			return statefulSetRolloutComplete(sts, generation), nil
		case rolloutKindDaemonSet:
			ds := &appsv1.DaemonSet{}
			if err := c.Get(ctx, nn, ds); err != nil {
				if apierrors.IsNotFound(err) {
					return false, errWorkloadDeleted
				}
				return false, err
			}
			return daemonSetRolloutComplete(ds, generation), nil
		default:
			return true, nil
		}
	})

	if errors.Is(err, errWorkloadDeleted) {
		return
	}
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || waitCtx.Err() == context.DeadlineExceeded {
			log.Info("timed out waiting for workload rollout to finish",
				"kind", kind, "namespace", nn.Namespace, "name", nn.Name,
				"generation", generation, "timeout", rolloutPollTimeout.String())
			return
		}
		log.Error(err, "waiting for workload rollout", "kind", kind, "namespace", nn.Namespace, "name", nn.Name)
		return
	}

	log.Info("workload rollout finished successfully",
		"kind", kind, "namespace", nn.Namespace, "name", nn.Name, "generation", generation)
}

func deploymentRolloutComplete(d *appsv1.Deployment, patchedGeneration int64) bool {
	if d.Generation < patchedGeneration {
		return false
	}
	if d.Spec.Replicas != nil && *d.Spec.Replicas == 0 {
		return d.Status.ObservedGeneration >= d.Generation
	}
	for i := range d.Status.Conditions {
		c := &d.Status.Conditions[i]
		if c.Type == appsv1.DeploymentProgressing && c.Status == corev1.ConditionTrue && c.Reason == deploymentProgressingReasonNewRS {
			return true
		}
	}
	return false
}

func statefulSetRolloutComplete(s *appsv1.StatefulSet, patchedGeneration int64) bool {
	if s.Status.ObservedGeneration < patchedGeneration {
		return false
	}
	spec := s.Spec.Replicas
	if spec == nil {
		return false
	}
	if *spec == 0 {
		return s.Status.UpdatedReplicas == 0
	}
	return s.Status.CurrentRevision == s.Status.UpdateRevision &&
		s.Status.ReadyReplicas == *spec &&
		s.Status.UpdatedReplicas == *spec
}

func daemonSetRolloutComplete(d *appsv1.DaemonSet, patchedGeneration int64) bool {
	if d.Status.ObservedGeneration < patchedGeneration {
		return false
	}
	des := d.Status.DesiredNumberScheduled
	if des == 0 {
		return true
	}
	return d.Status.UpdatedNumberScheduled == des &&
		d.Status.NumberUnavailable == 0 &&
		d.Status.NumberReady == des
}
