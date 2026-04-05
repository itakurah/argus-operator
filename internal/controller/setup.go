package controller

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// SetupWithManager registers a single rollout controller
func SetupWithManager(mgr manager.Manager) error {
	r := &RolloutReconciler{Client: mgr.GetClient()}
	return ctrl.NewControllerManagedBy(mgr).
		Named("rollout").
		For(&appsv1.Deployment{}).
		Watches(&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(enqueueRequestForWorkload)).
		Watches(&appsv1.DaemonSet{}, handler.EnqueueRequestsFromMapFunc(enqueueRequestForWorkload)).
		Watches(&corev1.ConfigMap{}, mapHandler("ConfigMap", r.mapConfigMapToWorkloads)).
		Watches(&corev1.Secret{}, mapHandler("Secret", r.mapSecretToWorkloads)).
		Complete(r)
}
