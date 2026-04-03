package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// SetupWithManager registers rollout controllers for Deployments, StatefulSets, and DaemonSets.
func SetupWithManager(mgr manager.Manager) error {
	dr := &DeploymentRolloutReconciler{Client: mgr.GetClient()}
	if err := ctrl.NewControllerManagedBy(mgr).
		Named("deployment-rollout").
		For(&appsv1.Deployment{}).
		Watches(
			&corev1.ConfigMap{},
			mapHandler("ConfigMap", dr.mapConfigMapToDeployments),
		).
		Watches(
			&corev1.Secret{},
			mapHandler("Secret", dr.mapSecretToDeployments),
		).
		Complete(dr); err != nil {
		return err
	}

	sr := &StatefulSetRolloutReconciler{Client: mgr.GetClient()}
	if err := ctrl.NewControllerManagedBy(mgr).
		Named("statefulset-rollout").
		For(&appsv1.StatefulSet{}).
		Watches(
			&corev1.ConfigMap{},
			mapHandler("ConfigMap", sr.mapConfigMapToStatefulSets),
		).
		Watches(
			&corev1.Secret{},
			mapHandler("Secret", sr.mapSecretToStatefulSets),
		).
		Complete(sr); err != nil {
		return err
	}

	dsr := &DaemonSetRolloutReconciler{Client: mgr.GetClient()}
	return ctrl.NewControllerManagedBy(mgr).
		Named("daemonset-rollout").
		For(&appsv1.DaemonSet{}).
		Watches(
			&corev1.ConfigMap{},
			mapHandler("ConfigMap", dsr.mapConfigMapToDaemonSets),
		).
		Watches(
			&corev1.Secret{},
			mapHandler("Secret", dsr.mapSecretToDaemonSets),
		).
		Complete(dsr)
}
