package controller

import (
	"context"
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/itakurah/argus-operator/internal/hash"
)

// RolloutReconciler patches Deployment, StatefulSet, and DaemonSet pod templates when referenced config changes.
type RolloutReconciler struct {
	client.Client
}

func (r *RolloutReconciler) Reconcile(ctx context.Context, req reconcile.Request) (ctrl.Result, error) {
	var errs []error

	dep := &appsv1.Deployment{}
	switch err := r.Get(ctx, req.NamespacedName, dep); {
	case err == nil:
		if e := r.reconcileDeployment(ctx, req, dep); e != nil {
			errs = append(errs, e)
		}
	case apierrors.IsNotFound(err):
	default:
		return ctrl.Result{}, err
	}

	sts := &appsv1.StatefulSet{}
	switch err := r.Get(ctx, req.NamespacedName, sts); {
	case err == nil:
		if e := r.reconcileStatefulSet(ctx, req, sts); e != nil {
			errs = append(errs, e)
		}
	case apierrors.IsNotFound(err):
	default:
		return ctrl.Result{}, err
	}

	ds := &appsv1.DaemonSet{}
	switch err := r.Get(ctx, req.NamespacedName, ds); {
	case err == nil:
		if e := r.reconcileDaemonSet(ctx, req, ds); e != nil {
			errs = append(errs, e)
		}
	case apierrors.IsNotFound(err):
	default:
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, errors.Join(errs...)
}

func (r *RolloutReconciler) reconcileDeployment(ctx context.Context, req reconcile.Request, dep *appsv1.Deployment) error {
	if !rolloutEnabled(dep.GetAnnotations()) {
		return nil
	}

	orig := dep.DeepCopy()
	newHash, err := hash.ComputeForPodTemplate(ctx, r.Client, dep.Namespace, &dep.Spec.Template.Spec)
	if err != nil {
		return err
	}

	cur := ""
	if dep.Spec.Template.Annotations != nil {
		cur = dep.Spec.Template.Annotations[hash.AnnotationKey]
	}
	if cur == newHash {
		return nil
	}

	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = make(map[string]string)
	}
	dep.Spec.Template.Annotations[hash.AnnotationKey] = newHash

	if err := r.Patch(ctx, dep, client.MergeFrom(orig)); err != nil {
		return client.IgnoreNotFound(err)
	}
	fresh := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, fresh); err != nil {
		return client.IgnoreNotFound(err)
	}
	logConfigHashPatched("Deployment", fresh.Namespace, fresh.Name, cur, newHash)
	go waitAndLogRollout(context.Background(), r.Client, req.NamespacedName, fresh.Generation, rolloutKindDeployment)
	return nil
}

func (r *RolloutReconciler) reconcileStatefulSet(ctx context.Context, req reconcile.Request, sts *appsv1.StatefulSet) error {
	if !rolloutEnabled(sts.GetAnnotations()) {
		return nil
	}

	orig := sts.DeepCopy()
	newHash, err := hash.ComputeForPodTemplate(ctx, r.Client, sts.Namespace, &sts.Spec.Template.Spec)
	if err != nil {
		return err
	}

	cur := ""
	if sts.Spec.Template.Annotations != nil {
		cur = sts.Spec.Template.Annotations[hash.AnnotationKey]
	}
	if cur == newHash {
		return nil
	}

	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = make(map[string]string)
	}
	sts.Spec.Template.Annotations[hash.AnnotationKey] = newHash

	if err := r.Patch(ctx, sts, client.MergeFrom(orig)); err != nil {
		return client.IgnoreNotFound(err)
	}
	fresh := &appsv1.StatefulSet{}
	if err := r.Get(ctx, req.NamespacedName, fresh); err != nil {
		return client.IgnoreNotFound(err)
	}
	logConfigHashPatched("StatefulSet", fresh.Namespace, fresh.Name, cur, newHash)
	go waitAndLogRollout(context.Background(), r.Client, req.NamespacedName, fresh.Generation, rolloutKindStatefulSet)
	return nil
}

func (r *RolloutReconciler) reconcileDaemonSet(ctx context.Context, req reconcile.Request, ds *appsv1.DaemonSet) error {
	if !rolloutEnabled(ds.GetAnnotations()) {
		return nil
	}

	orig := ds.DeepCopy()
	newHash, err := hash.ComputeForPodTemplate(ctx, r.Client, ds.Namespace, &ds.Spec.Template.Spec)
	if err != nil {
		return err
	}

	cur := ""
	if ds.Spec.Template.Annotations != nil {
		cur = ds.Spec.Template.Annotations[hash.AnnotationKey]
	}
	if cur == newHash {
		return nil
	}

	if ds.Spec.Template.Annotations == nil {
		ds.Spec.Template.Annotations = make(map[string]string)
	}
	ds.Spec.Template.Annotations[hash.AnnotationKey] = newHash

	if err := r.Patch(ctx, ds, client.MergeFrom(orig)); err != nil {
		return client.IgnoreNotFound(err)
	}
	fresh := &appsv1.DaemonSet{}
	if err := r.Get(ctx, req.NamespacedName, fresh); err != nil {
		return client.IgnoreNotFound(err)
	}
	logConfigHashPatched("DaemonSet", fresh.Namespace, fresh.Name, cur, newHash)
	go waitAndLogRollout(context.Background(), r.Client, req.NamespacedName, fresh.Generation, rolloutKindDaemonSet)
	return nil
}
