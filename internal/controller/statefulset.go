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

// StatefulSetRolloutReconciler patches StatefulSet pod templates when referenced config changes.
type StatefulSetRolloutReconciler struct {
	client.Client
}

func (r *StatefulSetRolloutReconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, req.NamespacedName, sts); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !rolloutEnabled(sts.GetAnnotations()) {
		return ctrl.Result{}, nil
	}

	orig := sts.DeepCopy()
	newHash, err := hash.ComputeForPodTemplate(ctx, r.Client, sts.Namespace, &sts.Spec.Template.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}

	cur := ""
	if sts.Spec.Template.Annotations != nil {
		cur = sts.Spec.Template.Annotations[hash.AnnotationKey]
	}
	if cur == newHash {
		return ctrl.Result{}, nil
	}

	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = make(map[string]string)
	}
	sts.Spec.Template.Annotations[hash.AnnotationKey] = newHash

	if err := r.Patch(ctx, sts, client.MergeFrom(orig)); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	fresh := &appsv1.StatefulSet{}
	if err := r.Get(ctx, req.NamespacedName, fresh); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	logConfigHashPatched("StatefulSet", fresh.Namespace, fresh.Name, cur, newHash)
	go waitAndLogRollout(context.Background(), r.Client, req.NamespacedName, fresh.Generation, rolloutKindStatefulSet)
	return ctrl.Result{}, nil
}

func (r *StatefulSetRolloutReconciler) mapConfigMapToStatefulSets(ctx context.Context, obj client.Object) []reconcile.Request {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil
	}
	list := &appsv1.StatefulSetList{}
	if err := r.List(ctx, list, client.InNamespace(cm.Namespace)); err != nil {
		ctrl.Log.Error(err, "list StatefulSets for ConfigMap watch", "namespace", cm.Namespace)
		return nil
	}
	var out []reconcile.Request
	for i := range list.Items {
		s := &list.Items[i]
		if !rolloutEnabled(s.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesConfigMap(&s.Spec.Template.Spec, cm.Name) {
			out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: s.Namespace, Name: s.Name}})
		}
	}
	logEnqueueFromConfigTrigger("ConfigMap", cm.Namespace, cm.Name, len(out))
	return out
}

func (r *StatefulSetRolloutReconciler) mapSecretToStatefulSets(ctx context.Context, obj client.Object) []reconcile.Request {
	sec, ok := obj.(*corev1.Secret)
	if !ok {
		return nil
	}
	list := &appsv1.StatefulSetList{}
	if err := r.List(ctx, list, client.InNamespace(sec.Namespace)); err != nil {
		ctrl.Log.Error(err, "list StatefulSets for Secret watch", "namespace", sec.Namespace)
		return nil
	}
	var out []reconcile.Request
	for i := range list.Items {
		s := &list.Items[i]
		if !rolloutEnabled(s.GetAnnotations()) {
			continue
		}
		if refs.PodReferencesSecret(&s.Spec.Template.Spec, sec.Name) {
			out = append(out, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: s.Namespace, Name: s.Name}})
		}
	}
	logEnqueueFromConfigTrigger("Secret", sec.Namespace, sec.Name, len(out))
	return out
}
