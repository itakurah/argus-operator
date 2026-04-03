package refs

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestCollectConfigMapNames_uniqueSorted(t *testing.T) {
	spec := &corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "v1",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "zebra"},
					},
				},
			},
			{
				Name: "v2",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{ConfigMap: &corev1.ConfigMapProjection{LocalObjectReference: corev1.LocalObjectReference{Name: "alpha"}}},
						},
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name: "c",
				EnvFrom: []corev1.EnvFromSource{
					{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "alpha"}}},
				},
				Env: []corev1.EnvVar{
					{
						Name: "x",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "beta"}, Key: "k"},
						},
					},
				},
			},
		},
	}
	got := CollectConfigMapNames(spec)
	want := []string{"alpha", "beta", "zebra"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestPodReferencesSecret(t *testing.T) {
	spec := &corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: "s",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{SecretName: "mysecret"},
				},
			},
		},
	}
	if !PodReferencesSecret(spec, "mysecret") {
		t.Fatal("expected reference")
	}
	if PodReferencesSecret(spec, "other") {
		t.Fatal("unexpected reference")
	}
}
