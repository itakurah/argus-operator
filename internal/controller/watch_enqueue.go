package controller

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// mapHandler provides an EventHandler that logs Config/Secret changes before enqueuing requests.
func mapHandler(resourceKind string, mapFunc func(context.Context, client.Object) []reconcile.Request) handler.EventHandler {
	return handler.Funcs{
		CreateFunc: func(ctx context.Context, e event.CreateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			if e.Object != nil {
				logConfigResourceCreated(resourceKind, e.Object.GetNamespace(), e.Object.GetName())
			}
			enqueueFromMap(ctx, mapFunc, e.Object, q)
		},
		UpdateFunc: func(ctx context.Context, e event.UpdateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			if e.ObjectNew != nil {
				logConfigResourceUpdated(resourceKind, e.ObjectNew.GetNamespace(), e.ObjectNew.GetName())
			}
			enqueueFromMap(ctx, mapFunc, e.ObjectNew, q)
		},
		DeleteFunc: func(ctx context.Context, e event.DeleteEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			if e.Object != nil {
				logConfigResourceDeleted(resourceKind, e.Object.GetNamespace(), e.Object.GetName())
			}
			enqueueFromMap(ctx, mapFunc, e.Object, q)
		},
	}
}

func enqueueFromMap(ctx context.Context, mapFunc func(context.Context, client.Object) []reconcile.Request, obj client.Object, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	if obj == nil {
		return
	}
	for _, req := range mapFunc(ctx, obj) {
		q.Add(req)
	}
}
