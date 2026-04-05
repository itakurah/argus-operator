package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/itakurah/argus-operator/internal/refs"
)

func (r *RolloutReconciler) mapConfigMapToWorkloads(ctx context.Context, obj client.Object) []reconcile.Request {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil
	}
	ns := cm.Namespace
	seen := make(map[types.NamespacedName]struct{})
	var out []reconcile.Request

	deployList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployList, client.InNamespace(ns)); err != nil {
		ctrl.Log.Error(err, "list Deployments for ConfigMap watch", "namespace", ns)
		return nil
	}
	for i := range deployList.Items {
		d := &deployList.Items[i]
		if !rolloutEnabled(d.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesConfigMap(&d.Spec.Template.Spec, cm.Name) {
			nn := types.NamespacedName{Namespace: d.Namespace, Name: d.Name}
			if _, dup := seen[nn]; !dup {
				seen[nn] = struct{}{}
				out = append(out, reconcile.Request{NamespacedName: nn})
			}
		}
	}

	stsList := &appsv1.StatefulSetList{}
	if err := r.List(ctx, stsList, client.InNamespace(ns)); err != nil {
		ctrl.Log.Error(err, "list StatefulSets for ConfigMap watch", "namespace", ns)
		return nil
	}
	for i := range stsList.Items {
		s := &stsList.Items[i]
		if !rolloutEnabled(s.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesConfigMap(&s.Spec.Template.Spec, cm.Name) {
			nn := types.NamespacedName{Namespace: s.Namespace, Name: s.Name}
			if _, dup := seen[nn]; !dup {
				seen[nn] = struct{}{}
				out = append(out, reconcile.Request{NamespacedName: nn})
			}
		}
	}

	dsList := &appsv1.DaemonSetList{}
	if err := r.List(ctx, dsList, client.InNamespace(ns)); err != nil {
		ctrl.Log.Error(err, "list DaemonSets for ConfigMap watch", "namespace", ns)
		return nil
	}
	for i := range dsList.Items {
		d := &dsList.Items[i]
		if !rolloutEnabled(d.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesConfigMap(&d.Spec.Template.Spec, cm.Name) {
			nn := types.NamespacedName{Namespace: d.Namespace, Name: d.Name}
			if _, dup := seen[nn]; !dup {
				seen[nn] = struct{}{}
				out = append(out, reconcile.Request{NamespacedName: nn})
			}
		}
	}

	logEnqueueFromConfigTrigger("ConfigMap", cm.Namespace, cm.Name, len(out))
	return out
}

func (r *RolloutReconciler) mapSecretToWorkloads(ctx context.Context, obj client.Object) []reconcile.Request {
	sec, ok := obj.(*corev1.Secret)
	if !ok {
		return nil
	}
	ns := sec.Namespace
	seen := make(map[types.NamespacedName]struct{})
	var out []reconcile.Request

	deployList := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployList, client.InNamespace(ns)); err != nil {
		ctrl.Log.Error(err, "list Deployments for Secret watch", "namespace", ns)
		return nil
	}
	for i := range deployList.Items {
		d := &deployList.Items[i]
		if !rolloutEnabled(d.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesSecret(&d.Spec.Template.Spec, sec.Name) {
			nn := types.NamespacedName{Namespace: d.Namespace, Name: d.Name}
			if _, dup := seen[nn]; !dup {
				seen[nn] = struct{}{}
				out = append(out, reconcile.Request{NamespacedName: nn})
			}
		}
	}

	stsList := &appsv1.StatefulSetList{}
	if err := r.List(ctx, stsList, client.InNamespace(ns)); err != nil {
		ctrl.Log.Error(err, "list StatefulSets for Secret watch", "namespace", ns)
		return nil
	}
	for i := range stsList.Items {
		s := &stsList.Items[i]
		if !rolloutEnabled(s.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesSecret(&s.Spec.Template.Spec, sec.Name) {
			nn := types.NamespacedName{Namespace: s.Namespace, Name: s.Name}
			if _, dup := seen[nn]; !dup {
				seen[nn] = struct{}{}
				out = append(out, reconcile.Request{NamespacedName: nn})
			}
		}
	}

	dsList := &appsv1.DaemonSetList{}
	if err := r.List(ctx, dsList, client.InNamespace(ns)); err != nil {
		ctrl.Log.Error(err, "list DaemonSets for Secret watch", "namespace", ns)
		return nil
	}
	for i := range dsList.Items {
		d := &dsList.Items[i]
		if !rolloutEnabled(d.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesSecret(&d.Spec.Template.Spec, sec.Name) {
			nn := types.NamespacedName{Namespace: d.Namespace, Name: d.Name}
			if _, dup := seen[nn]; !dup {
				seen[nn] = struct{}{}
				out = append(out, reconcile.Request{NamespacedName: nn})
			}
		}
	}

	logEnqueueFromConfigTrigger("Secret", sec.Namespace, sec.Name, len(out))
	return out
}
