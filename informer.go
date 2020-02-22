package main

import (
	"k8s.io/client-go/tools/cache"
)

type Informer struct {
	informer cache.SharedIndexInformer
	handler  Handler
	stopChan <-chan struct{}
}

func NewInformer(
	informer cache.SharedIndexInformer,
	handler Handler,
	stopChan <-chan struct{},
) *Informer {
	return &Informer{
		informer,
		handler,
		stopChan,
	}
}

func (g *Informer) Start() {
	g.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			g.handler.OnAdd(obj)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			g.handler.OnUpdate(old, new)
		},
		DeleteFunc: func(obj interface{}) {
			g.handler.OnDelete(obj)
		},
	})
	g.informer.Run(g.stopChan)
}
