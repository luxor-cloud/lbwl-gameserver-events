package main

import (
	"log"
	"os"

	gamev1beta1 "go.freggy.dev/lbwl-gameserver-controllers/api/v1beta1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var consumer map[string]string

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(err.Error())
	}

	dynamicclient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}
}

func createGameserverInformer(client dynamic.Interface, namespace string, handler cache.ResourceEventHandler) cache.SharedIndexInformer {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, namespace, nil)
	generic := factory.ForResource(gamev1beta1.GroupVersion.WithResource("gameservers"))
	informer := generic.Informer()
	informer.AddEventHandler(handler)
	return informer
}

func createPodInformer(client kubernetes.Interface, handler cache.ResourceEventHandler) cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(client, 0)
	informer := factory.Core().V1().Pods().Informer()
	informer.AddEventHandler(handler)
	return informer
}
