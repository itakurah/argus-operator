package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/itakurah/argus-operator/internal/hash"
	"github.com/itakurah/argus-operator/internal/refs"
)

// DeploymentRolloutReconciler patches Deployment pod templates when referenced config changes.
type DeploymentRolloutReconciler struct {
	client.Client
}

func (r *DeploymentRolloutReconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	dep := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, dep); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !rolloutEnabled(dep.GetAnnotations()) {
		return ctrl.Result{}, nil
	}

	orig := dep.DeepCopy()
	newHash, err := hash.ComputeForPodTemplate(ctx, r.Client, dep.Namespace, &dep.Spec.Template.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}

	cur := ""
	if dep.Spec.Template.Annotations != nil {
		cur = dep.Spec.Template.Annotations[hash.AnnotationKey]
	}
	if cur == newHash {
		return ctrl.Result{}, nil
	}

	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = make(map[string]string)
	}
	dep.Spec.Template.Annotations[hash.AnnotationKey] = newHash

	if err := r.Patch(ctx, dep, client.MergeFrom(orig)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DeploymentRolloutReconciler) mapConfigMapToDeployments(ctx context.Context, obj client.Object) []reconcile.Request {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil
	}
	list := &appsv1.DeploymentList{}
	if err := r.List(ctx, list, client.InNamespace(cm.Namespace)); err != nil {
		ctrl.Log.Error(err, "list Deployments for ConfigMap watch", "namespace", cm.Namespace)
		return nil
	}
	var out []reconcile.Request
	for i := range list.Items {
		d := &list.Items[i]
		if !rolloutEnabled(d.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesConfigMap(&d.Spec.Template.Spec, cm.Name) {
			out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: d.Namespace, Name: d.Name}})
		}
	}
	return out
}

func (r *DeploymentRolloutReconciler) mapSecretToDeployments(ctx context.Context, obj client.Object) []reconcile.Request {
	sec, ok := obj.(*corev1.Secret)
	if !ok {
		return nil
	}
	list := &appsv1.DeploymentList{}
	if err := r.List(ctx, list, client.InNamespace(sec.Namespace)); err != nil {
		ctrl.Log.Error(err, "list Deployments for Secret watch", "namespace", sec.Namespace)
		return nil
	}
	var out []reconcile.Request
	for i := range list.Items {
		d := &list.Items[i]
		if !rolloutEnabled(d.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesSecret(&d.Spec.Template.Spec, sec.Name) {
			out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: d.Namespace, Name: d.Name}})
		}
	}
	return out
}
