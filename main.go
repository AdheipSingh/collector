package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"kube-collector/pkg/collector"
	"kube-collector/pkg/k8s"
	"kube-collector/pkg/store"

	corev1 "k8s.io/api/core/v1"

	"strings"
	"time"
)

// logMessage is the CRI internal log type.
type logMessage struct {
	timestamp time.Time
	log       []string
}

func main() {
	pods, _ := k8s.K8s.ListPods("operator", "namespace=operator")

	ticker := time.NewTicker(5 * time.Second)
	for t := range ticker.C {
		for _, p := range pods.Items {
			fmt.Println(store.GetTime(p.Name))

			fmt.Println(p.GetName())
			fmt.Println("Invoked at ", t)

			collector.GetPodLogs(p)
			///fmt.Println(a)
		}
	}

}

func getPodLogs(pod corev1.Pod) (logMessage, error) {

	var newLogTime int64

	var podLogOpts corev1.PodLogOptions

	if store.GetTime(pod.GetName()) != (time.Time{}) {
		newLogTime = int64(time.Now().Sub(store.GetTime(pod.GetName())).Seconds())
		podLogOpts = corev1.PodLogOptions{
			SinceSeconds: &newLogTime,
			Timestamps:   true,
		}
	} else {
		podLogOpts = corev1.PodLogOptions{
			//	SinceSeconds: &newLogTime,
			Timestamps: true,
		}
	}

	req := k8s.K8s.GetPodLogs(pod, podLogOpts)

	podLogs, err := req.Stream(context.TODO())

	if err != nil {
		return logMessage{}, err
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return logMessage{}, err
	}
	str := buf.String()

	newStr := strings.Split(str, "\n")

	if len(newStr) > 1 {
		a := newStr[len(newStr)-2]

		words := strings.Fields(a)

		aa, _ := time.Parse(time.RFC3339, words[0])
		if err != nil {

		}

		store.PutPoNameTime(pod.GetName(), aa)

		var lm logMessage

		lm.timestamp = aa
		lm.log = newStr[1:]

		fmt.Println(lm)

		return lm, nil
	} else {
		return logMessage{}, nil
	}
}
