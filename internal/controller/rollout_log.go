package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/itakurah/argus-operator/internal/hash"
)

// logConfigHashPatched logs after a successful patch of the pod template hash annotation.
func logConfigHashPatched(kind, namespace, name, previousHash, newHash string) {
	prev := previousHash
	if prev == "" {
		prev = "(none)"
	}
	ctrl.Log.WithName("rollout").Info("patched pod template config hash; polling until rollout completes",
		"kind", kind,
		"namespace", namespace,
		"name", name,
		"annotation", hash.AnnotationKey,
		"previousHash", prev,
		"newHash", newHash,
	)
}

// logConfigResourceCreated logs when a watched ConfigMap or Secret is created.
func logConfigResourceCreated(kind, namespace, name string) {
	ctrl.Log.WithName("rollout").Info("created; reconciling referencing workloads",
		"kind", kind,
		"namespace", namespace,
		"name", name,
	)
}

// logConfigResourceUpdated logs when a watched ConfigMap or Secret is updated.
func logConfigResourceUpdated(kind, namespace, name string) {
	ctrl.Log.WithName("rollout").Info("updated; reconciling referencing workloads",
		"kind", kind,
		"namespace", namespace,
		"name", name,
	)
}

// logConfigResourceDeleted logs when a watched ConfigMap or Secret is deleted.
func logConfigResourceDeleted(kind, namespace, name string) {
	ctrl.Log.WithName("rollout").Info("deleted; reconciling referencing workloads",
		"kind", kind,
		"namespace", namespace,
		"name", name,
	)
}

// logEnqueueFromConfigTrigger logs when a ConfigMap or Secret event queues workload reconciles.
func logEnqueueFromConfigTrigger(trigger string, namespace, name string, workloadCount int) {
	if workloadCount == 0 {
		return
	}
	ctrl.Log.WithName("rollout").Info("enqueued workload reconciles for referencing workloads",
		"trigger", trigger,
		"namespace", namespace,
		"name", name,
		"workloads", workloadCount,
	)
}
