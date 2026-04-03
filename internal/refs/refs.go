package refs

import (
	"slices"

	corev1 "k8s.io/api/core/v1"
)

// CollectConfigMapNames returns unique ConfigMap names referenced by the pod spec.
func CollectConfigMapNames(spec *corev1.PodSpec) []string {
	var out []string
	add := func(name string) {
		if name == "" {
			return
		}
		if !slices.Contains(out, name) {
			out = append(out, name)
		}
	}

	for _, v := range spec.Volumes {
		if v.ConfigMap != nil {
			add(v.ConfigMap.Name)
		}
		if v.Projected != nil {
			for _, s := range v.Projected.Sources {
				if s.ConfigMap != nil {
					add(s.ConfigMap.Name)
				}
			}
		}
	}

	visit := func(containers []corev1.Container) {
		for i := range containers {
			c := &containers[i]
			for _, e := range c.EnvFrom {
				if e.ConfigMapRef != nil {
					add(e.ConfigMapRef.Name)
				}
			}
			for _, e := range c.Env {
				if e.ValueFrom != nil && e.ValueFrom.ConfigMapKeyRef != nil {
					add(e.ValueFrom.ConfigMapKeyRef.Name)
				}
			}
		}
	}
	visit(spec.Containers)
	visit(spec.InitContainers)
	slices.Sort(out)
	return out
}

// CollectSecretNames returns unique Secret names referenced by the pod spec.
func CollectSecretNames(spec *corev1.PodSpec) []string {
	var out []string
	add := func(name string) {
		if name == "" {
			return
		}
		if !slices.Contains(out, name) {
			out = append(out, name)
		}
	}

	for _, v := range spec.Volumes {
		if v.Secret != nil {
			add(v.Secret.SecretName)
		}
		if v.Projected != nil {
			for _, s := range v.Projected.Sources {
				if s.Secret != nil {
					add(s.Secret.Name)
				}
			}
		}
	}

	visit := func(containers []corev1.Container) {
		for i := range containers {
			c := &containers[i]
			for _, e := range c.EnvFrom {
				if e.SecretRef != nil {
					add(e.SecretRef.Name)
				}
			}
			for _, e := range c.Env {
				if e.ValueFrom != nil && e.ValueFrom.SecretKeyRef != nil {
					add(e.ValueFrom.SecretKeyRef.Name)
				}
			}
		}
	}
	visit(spec.Containers)
	visit(spec.InitContainers)
	slices.Sort(out)
	return out
}

// PodReferencesConfigMap reports whether the pod spec references the given ConfigMap name.
func PodReferencesConfigMap(spec *corev1.PodSpec, name string) bool {
	if name == "" {
		return false
	}
	for _, n := range CollectConfigMapNames(spec) {
		if n == name {
			return true
		}
	}
	return false
}

// PodReferencesSecret reports whether the pod spec references the given Secret name.
func PodReferencesSecret(spec *corev1.PodSpec, name string) bool {
	if name == "" {
		return false
	}
	for _, n := range CollectSecretNames(spec) {
		if n == name {
			return true
		}
	}
	return false
}
