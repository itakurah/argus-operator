package hash

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestComputeForPodTemplate_orderStable(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	ns := "default"
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: ns},
		Data:       map[string]string{"z": "2", "a": "1"},
	}
	cm2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: ns},
		Data:       map[string]string{"k": "v"},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm1, cm2).Build()

	spec := &corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "v",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "b"},
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name: "c",
				EnvFrom: []corev1.EnvFromSource{
					{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "a"}}},
				},
			},
		},
	}

	h1, err := ComputeForPodTemplate(ctx, cl, ns, spec)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := ComputeForPodTemplate(ctx, cl, ns, spec)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("hash not stable: %s vs %s", h1, h2)
	}
}

func TestComputeForPodTemplate_emptyRefs(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	h, err := ComputeForPodTemplate(ctx, cl, "default", &corev1.PodSpec{})
	if err != nil {
		t.Fatal(err)
	}
	empty := sha256.Sum256(nil)
	want := hex.EncodeToString(empty[:])
	if h != want {
		t.Fatalf("expected empty digest %s, got %s", want, h)
	}
}
