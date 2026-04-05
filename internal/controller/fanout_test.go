package controller

import (
	"context"
	"sort"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := appsv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

func rolloutAnn() map[string]string {
	return map[string]string{RolloutAnnotation: "true"}
}

func cmVolume(name string) []corev1.Volume {
	return []corev1.Volume{{
		Name: "v",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: name},
			},
		},
	}}
}

func secretVolume(name string) []corev1.Volume {
	return []corev1.Volume{{
		Name: "v",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{SecretName: name},
		},
	}}
}

func TestMapConfigMapToWorkloads_wrongType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := &RolloutReconciler{Client: fake.NewClientBuilder().WithScheme(testScheme(t)).Build()}
	if got := r.mapConfigMapToWorkloads(ctx, &corev1.Secret{}); got != nil {
		t.Fatalf("wrong type: got %v, want nil", got)
	}
}

func TestMapConfigMapToWorkloads_enqueuesAndDedupes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "fanout"
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "cfg"}}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "app", Annotations: rolloutAnn()},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "b"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"a": "b"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "i"}}, Volumes: cmVolume("cfg")},
			},
		},
	}
	ignored := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "norollout", Annotations: map[string]string{}},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "c"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"a": "c"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "i"}}, Volumes: cmVolume("cfg")},
			},
		},
	}
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "app", Annotations: rolloutAnn()},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "d"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"a": "d"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "i"}}, Volumes: cmVolume("cfg")},
			},
		},
	}
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "agent", Annotations: rolloutAnn()},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "e"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"a": "e"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "i"}}, Volumes: cmVolume("cfg")},
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(cm, dep, ignored, sts, ds).Build()
	r := &RolloutReconciler{Client: cl}
	got := r.mapConfigMapToWorkloads(ctx, cm)
	want := []types.NamespacedName{
		{Namespace: ns, Name: "app"},
		{Namespace: ns, Name: "agent"},
	}
	var names []types.NamespacedName
	for _, req := range got {
		names = append(names, req.NamespacedName)
	}
	sort.Slice(names, func(i, j int) bool {
		if names[i].Name != names[j].Name {
			return names[i].Name < names[j].Name
		}
		return names[i].Namespace < names[j].Namespace
	})
	sort.Slice(want, func(i, j int) bool {
		if want[i].Name != want[j].Name {
			return want[i].Name < want[j].Name
		}
		return want[i].Namespace < want[j].Namespace
	})
	if len(names) != len(want) {
		t.Fatalf("len %d, want %d: got %#v", len(names), len(want), names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("got %#v, want %#v", names, want)
		}
	}
}

func TestMapSecretToWorkloads_enqueues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "secns"
	sec := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "s1"}}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "d1", Annotations: rolloutAnn()},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"a": "b"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"a": "b"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "c", Image: "i"}}, Volumes: secretVolume("s1")},
			},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(sec, dep).Build()
	r := &RolloutReconciler{Client: cl}
	got := r.mapSecretToWorkloads(ctx, sec)
	want := []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: ns, Name: "d1"}}}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
