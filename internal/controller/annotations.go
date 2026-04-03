package controller

// RolloutAnnotation must be "true" on the workload (Deployment/StatefulSet/DaemonSet)
// metadata.annotations to enable rollout-on-config updates.
const RolloutAnnotation = "argus.io/rollout-on-update"

func rolloutEnabled(ann map[string]string) bool {
	if ann == nil {
		return false
	}
	return ann[RolloutAnnotation] == "true"
}
