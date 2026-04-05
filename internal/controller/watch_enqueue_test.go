package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnqueueRequestForWorkload(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	if got := enqueueRequestForWorkload(ctx, nil); got != nil {
		t.Fatalf("nil object: got %v, want nil", got)
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns1", Name: "web"},
	}
	want := []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: "ns1", Name: "web"}}}
	if got := enqueueRequestForWorkload(ctx, dep); len(got) != 1 || got[0] != want[0] {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
