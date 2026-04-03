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

// DaemonSetRolloutReconciler patches DaemonSet pod templates when referenced config changes.
type DaemonSetRolloutReconciler struct {
	client.Client
}

func (r *DaemonSetRolloutReconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	ds := &appsv1.DaemonSet{}
	if err := r.Get(ctx, req.NamespacedName, ds); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !rolloutEnabled(ds.GetAnnotations()) {
		return ctrl.Result{}, nil
	}

	orig := ds.DeepCopy()
	newHash, err := hash.ComputeForPodTemplate(ctx, r.Client, ds.Namespace, &ds.Spec.Template.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}

	cur := ""
	if ds.Spec.Template.Annotations != nil {
		cur = ds.Spec.Template.Annotations[hash.AnnotationKey]
	}
	if cur == newHash {
		return ctrl.Result{}, nil
	}

	if ds.Spec.Template.Annotations == nil {
		ds.Spec.Template.Annotations = make(map[string]string)
	}
	ds.Spec.Template.Annotations[hash.AnnotationKey] = newHash

	if err := r.Patch(ctx, ds, client.MergeFrom(orig)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *DaemonSetRolloutReconciler) mapConfigMapToDaemonSets(ctx context.Context, obj client.Object) []reconcile.Request {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil
	}
	list := &appsv1.DaemonSetList{}
	if err := r.List(ctx, list, client.InNamespace(cm.Namespace)); err != nil {
		ctrl.Log.Error(err, "list DaemonSets for ConfigMap watch", "namespace", cm.Namespace)
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

func (r *DaemonSetRolloutReconciler) mapSecretToDaemonSets(ctx context.Context, obj client.Object) []reconcile.Request {
	sec, ok := obj.(*corev1.Secret)
	if !ok {
		return nil
	}
	list := &appsv1.DaemonSetList{}
	if err := r.List(ctx, list, client.InNamespace(sec.Namespace)); err != nil {
		ctrl.Log.Error(err, "list DaemonSets for Secret watch", "namespace", sec.Namespace)
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
