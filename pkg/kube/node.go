package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func RunScriptOnNode(client *kubernetes.Clientset, node v1.Node, namespacedName types.NamespacedName, image string, script string) (pod *corev1.Pod, err error) {
	priviBool := true
	tolerations := []v1.Toleration{}
	for _, taint := range node.Spec.Taints {
		tolerations = append(tolerations, v1.Toleration{
			Key:      taint.Key,
			Value:    "",
			Operator: v1.TolerationOperator(v1.TolerationOpExists),
			Effect:   taint.Effect,
		})
	}
	automountSA := false
	pod, err = client.CoreV1().Pods(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
			Spec: corev1.PodSpec{
				AutomountServiceAccountToken: &automountSA,
				NodeName:                     node.Name,
				Containers: []corev1.Container{
					{
						Name:    "script",
						Image:   image,
						Command: []string{"sh"},
						Args:    []string{"-c", "echo \"sudo " + script + "\" | nsenter -t 1 -m -u -i -n"},
						SecurityContext: &corev1.SecurityContext{
							Privileged: &priviBool,
						},
						ImagePullPolicy: corev1.PullAlways,
					},
				},
				HostIPC:       true,
				HostNetwork:   true,
				HostPID:       true,
				RestartPolicy: corev1.RestartPolicyNever,
				Tolerations:   tolerations,
			},
		},
		metav1.CreateOptions{},
	)
	return
}

func DownloadFileOnNode(client *kubernetes.Clientset, node v1.Node, namespacedName types.NamespacedName, image, remotefile, localfile string) (pod *corev1.Pod, err error) {
	tolerations := []v1.Toleration{}
	for _, taint := range node.Spec.Taints {
		tolerations = append(tolerations, v1.Toleration{
			Key:      taint.Key,
			Value:    "",
			Operator: v1.TolerationOperator(v1.TolerationOpExists),
			Effect:   taint.Effect,
		})
	}
	automountSA := false
	pod, err = client.CoreV1().Pods(namespacedName.Namespace).Create(
		context.TODO(),
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
			Spec: corev1.PodSpec{
				AutomountServiceAccountToken: &automountSA,
				NodeName:                     node.Name,
				Containers: []corev1.Container{
					{
						Name:            "file",
						Image:           image,
						Command:         []string{"sh"},
						Args:            []string{"-c", fmt.Sprintf("cp -R %s /host%s", remotefile, localfile)},
						ImagePullPolicy: corev1.PullAlways,
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "data",
								MountPath: "/host",
							},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
				Tolerations:   tolerations,
				Volumes: []v1.Volume{
					{
						Name: "data",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/",
							},
						},
					},
				},
			},
		},
		metav1.CreateOptions{},
	)
	return
}