package hash

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/itakurah/argus-operator/internal/refs"
)

// AnnotationKey is written to spec.template.metadata.annotations to trigger rollouts.
const AnnotationKey = "argus.io/config-hash"

// ComputeForPodTemplate loads all referenced ConfigMaps and Secrets in the namespace
// and returns a SHA256 hex digest of their .Data fields.
func ComputeForPodTemplate(ctx context.Context, c client.Reader, namespace string, spec *corev1.PodSpec) (string, error) {
	if spec == nil {
		return digestEmpty(), nil
	}
	cmNames := refs.CollectConfigMapNames(spec)
	secNames := refs.CollectSecretNames(spec)

	var parts []string
	for _, name := range cmNames {
		cm := &corev1.ConfigMap{}
		key := types.NamespacedName{Namespace: namespace, Name: name}
		if err := c.Get(ctx, key, cm); err != nil {
			if errors.IsNotFound(err) {
				parts = append(parts, "ConfigMap/"+name+"\n"+canonicalConfigMapData(nil))
				continue
			}
			return "", fmt.Errorf("get configmap %s/%s: %w", namespace, name, err)
		}
		parts = append(parts, "ConfigMap/"+name+"\n"+canonicalConfigMapData(cm.Data))
	}
	for _, name := range secNames {
		sec := &corev1.Secret{}
		key := types.NamespacedName{Namespace: namespace, Name: name}
		if err := c.Get(ctx, key, sec); err != nil {
			if errors.IsNotFound(err) {
				parts = append(parts, "Secret/"+name+"\n"+canonicalSecretData(nil))
				continue
			}
			return "", fmt.Errorf("get secret %s/%s: %w", namespace, name, err)
		}
		parts = append(parts, "Secret/"+name+"\n"+canonicalSecretData(sec.Data))
	}
	if len(parts) == 0 {
		return digestEmpty(), nil
	}
	h := sha256.Sum256([]byte(strings.Join(parts, "\n---\n")))
	return hex.EncodeToString(h[:]), nil
}

func digestEmpty() string {
	h := sha256.Sum256(nil)
	return hex.EncodeToString(h[:])
}

func canonicalConfigMapData(data map[string]string) string {
	if len(data) == 0 {
		return ""
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte(0)
		b.WriteString(data[k])
		b.WriteByte('\n')
	}
	return b.String()
}

func canonicalSecretData(data map[string][]byte) string {
	if len(data) == 0 {
		return ""
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte(0)
		b.Write(data[k])
		b.WriteByte('\n')
	}
	return b.String()
}
