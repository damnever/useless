package v1

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FunctionSpec `json:"spec"`
	// Status FuncStatus `json:"status"`
}

type FunctionSpec struct {
	FuncName    string `json:"funcName"`
	FuncContent string `json:"funcContent"`
	Image       string `json:"image"`
	Replicas    *int32 `json:"replicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Function `json:"items"`
}

func (f *Function) Deployment() *appsv1.Deployment {
	labels := map[string]string{
		"controller": f.Name,
		"useless":    "function",
		"function":   f.Spec.FuncName,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.Spec.FuncName,
			Namespace: f.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(f, SchemeGroupVersion.WithKind("Function")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: f.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  f.Spec.FuncName,
							Image: f.Spec.Image,
						},
					},
				},
			},
		},
	}
}

func (f *Function) HorizontalPodAutoscaler() *autoscalingv1.HorizontalPodAutoscaler {
	return &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.Spec.FuncName,
			Namespace: f.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(f, SchemeGroupVersion.WithKind("Function")),
			},
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind:       "Function",
				Name:       f.Spec.FuncName,
				APIVersion: "v1",
			},
			MinReplicas:                    int32ptr(1),
			MaxReplicas:                    10,
			TargetCPUUtilizationPercentage: int32ptr(50),
		},
	}
}

func (f *Function) Service() *corev1.Service {
	labels := map[string]string{
		"controller": f.Name,
		"useless":    "function",
		"function":   f.Spec.FuncName,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.Spec.FuncName,
			Namespace: f.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(f, SchemeGroupVersion.WithKind("Function")),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Name:     f.Spec.FuncName,
				Protocol: corev1.ProtocolTCP,
				Port:     80,
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func int32ptr(i int32) *int32 {
	return &i
}

func externalName(funcName string) string {
	const rootDomain = "useless.io.dev1"
	return fmt.Sprintf("%s.%s", funcName, rootDomain)
}
