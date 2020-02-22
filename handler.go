package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	cloudevents "github.com/cloudevents/sdk-go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Handler interface {
	OnAdd(obj interface{})
	OnUpdate(old, new interface{})
	OnDelete(obj interface{})
}

type GameserverHandler struct {
	mu        *sync.Mutex
	consumers map[string]map[types.UID]*Consumer
}

func (gh *GameserverHandler) OnAdd(obj interface{}) {

}

func (gh *GameserverHandler) OnUpdate(old, new interface{}) {

}

func (gh *GameserverHandler) OnDelete(obj interface{}) {

}

type PodHandler struct {
	mu             *sync.Mutex
	consumerPerPod map[types.UID]Consumer
	consumers      map[string]map[types.UID]*Consumer
	labelsPerPod   map[types.UID][]string
	client         cloudevents.Client
}

func (ph *PodHandler) OnAdd(obj interface{}) {
	pod := obj.(v1.Object)
	annotations := pod.GetAnnotations()

	if ph.invalidPod(pod) {
		log.Println(fmt.Sprintf("Pod %s does not meet the requirements."))
		return
	}

	// TODO: establish dynamic service upstream

	list := strings.Split(annotations["freggy.dev/gameserver-events"], ",")

	ph.mu.Lock()
	defer ph.mu.Unlock()

	consumer := Consumer{
		list,
		ph.client,
	}

	ph.labelsPerPod[pod.GetUID()] = list
	ph.consumerPerPod[pod.GetUID()] = consumer

	for _, label := range list {
		ph.consumers[label][pod.GetUID()] = &consumer
	}
}

func (ph *PodHandler) OnUpdate(old, new interface{}) {
	oldPod := old.(v1.Object)
	newPod := new.(v1.Object)

	if ph.invalidPod(oldPod) {
		return
	}

	if ph.invalidPod(newPod) {
		log.Println(fmt.Sprintf("Updated pod %s does not meet the requirements anymore."))
		ph.free(newPod)
	}

	list := strings.Split(newPod.GetAnnotations()["freggy.dev/gameserver-events"], ",")
	defer func() {
		// update new labels
		ph.labelsPerPod[oldPod.GetUID()] = list
	}()

	for _, label := range list {
		// if the label is not contained in the new list BUT there is currently a consumer
		// associated for that label we know it has been removed so deassociate the consumer from the specifc label
		if !containsString(ph.labelsPerPod[oldPod.GetUID()], label) && ph.consumers[label][oldPod.GetUID()] != nil {
			delete(ph.consumers[label], oldPod.GetUID())
			return
		}

		// The label is NOT contained in the current label pool and NO consumer is associated with it
		// this means that the label has been newly added -> associate consumer
		if !containsString(ph.labelsPerPod[oldPod.GetUID()], label) && ph.consumers[label][oldPod.GetUID()] == nil {
			consumer := ph.consumerPerPod[oldPod.GetUID()]
			ph.consumers[label][oldPod.GetUID()] = &consumer
			return
		}
	}
}

func (ph *PodHandler) OnDelete(obj interface{}) {
	pod := obj.(v1.Object)
	if ph.invalidPod(pod) {
		return
	}
	ph.free(pod)
}

func (ph PodHandler) invalidPod(obj v1.Object) bool {
	return obj.GetAnnotations()["consul.hashicorp.com/connect-service"] == "" ||
		obj.GetAnnotations()["consul.hashicorp.com/connect-service-port"] == "" ||
		obj.GetAnnotations()["freggy.dev/gameserver-events"] == ""
}

func (ph *PodHandler) free(pod v1.Object) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	log.Println(fmt.Sprintf("Deassociating all consumers for pod %s."))
	delete(ph.consumerPerPod, pod.GetUID())
	for _, label := range ph.labelsPerPod[pod.GetUID()] {
		log.Println(fmt.Sprintf("Deassociating label %s from pod %s.", label, pod.GetName()))
		delete(ph.consumers[label], pod.GetUID())
	}
	delete(ph.labelsPerPod, pod.GetUID())
}

func containsString(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}
