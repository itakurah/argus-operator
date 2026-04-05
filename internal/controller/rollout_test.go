package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/itakurah/argus-operator/internal/hash"
)

func TestReconcile_noWorkloads(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	r := &RolloutReconciler{Client: cl}
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: "ns", Name: "missing"}}
	if _, err := r.Reconcile(ctx, req); err != nil {
		t.Fatal(err)
	}
}

func TestReconcile_deploymentRolloutDisabled_noPatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "ns"
	name := "dep"
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c", Image: "img"}},
				},
			},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dep).Build()
	r := &RolloutReconciler{Client: cl}
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}}
	if _, err := r.Reconcile(ctx, req); err != nil {
		t.Fatal(err)
	}
	fresh := &appsv1.Deployment{}
	if err := cl.Get(ctx, req.NamespacedName, fresh); err != nil {
		t.Fatal(err)
	}
	if fresh.Spec.Template.Annotations != nil && fresh.Spec.Template.Annotations[hash.AnnotationKey] != "" {
		t.Fatalf("unexpected config hash annotation: %v", fresh.Spec.Template.Annotations)
	}
}

func TestReconcile_deploymentHashUpToDate_noPatchNoWait(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "ns"
	name := "dep"
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "cfg"},
		Data:       map[string]string{"k": "v"},
	}
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{{Name: "c", Image: "img"}},
		Volumes: []corev1.Volume{{
			Name: "v",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "cfg"},
				},
			},
		}},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
	wantHash, err := hash.ComputeForPodTemplate(ctx, cl, ns, &podSpec)
	if err != nil {
		t.Fatal(err)
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ns,
			Name:        name,
			Annotations: map[string]string{RolloutAnnotation: "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": name},
					Annotations: map[string]string{hash.AnnotationKey: wantHash},
				},
				Spec: podSpec,
			},
		},
	}
	cl = fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm, dep).Build()
	r := &RolloutReconciler{Client: cl}
	req := reconcile.Request{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}}
	if _, err := r.Reconcile(ctx, req); err != nil {
		t.Fatal(err)
	}
	fresh := &appsv1.Deployment{}
	if err := cl.Get(ctx, req.NamespacedName, fresh); err != nil {
		t.Fatal(err)
	}
	got := fresh.Spec.Template.Annotations[hash.AnnotationKey]
	if got != wantHash {
		t.Fatalf("annotation hash %q, want %q", got, wantHash)
	}
}
