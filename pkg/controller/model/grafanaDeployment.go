package model

import (
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MemoryRequest = "256Mi"
	CpuRequest    = "100m"
	MemoryLimit   = "1024Mi"
	CpuLimit      = "500m"
)

func getResources(cr *v1alpha1.Grafana) v13.ResourceRequirements {
	if cr.Spec.Resources != nil {
		return *cr.Spec.Resources
	}
	return v13.ResourceRequirements{
		Requests: v13.ResourceList{
			v13.ResourceMemory: resource.MustParse(MemoryRequest),
			v13.ResourceCPU:    resource.MustParse(CpuRequest),
		},
		Limits: v13.ResourceList{
			v13.ResourceMemory: resource.MustParse(MemoryLimit),
			v13.ResourceCPU:    resource.MustParse(CpuLimit),
		},
	}
}

func getReplicas(cr *v1alpha1.Grafana) *int32 {
	var replicas int32 = 1
	if cr.Spec.Deployment == nil {
		return &replicas
	}
	if cr.Spec.Deployment.Replicas <= 0 {
		return &replicas
	} else {
		return &cr.Spec.Deployment.Replicas
	}
}

func getRollingUpdateStrategy() *v1.RollingUpdateDeployment {
	var maxUnaval intstr.IntOrString = intstr.FromInt(25)
	var maxSurge intstr.IntOrString = intstr.FromInt(25)
	return &v1.RollingUpdateDeployment{
		MaxUnavailable: &maxUnaval,
		MaxSurge:       &maxSurge,
	}
}

func getPodAnnotations(cr *v1alpha1.Grafana) map[string]string {
	var annotations = map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Annotations != nil {
		annotations = cr.Spec.Deployment.Annotations
	}

	// Add fixed annotations
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = fmt.Sprintf("%v", GetGrafanaPort(cr))
	return annotations
}

func getPodLabels(cr *v1alpha1.Grafana) map[string]string {
	var labels = map[string]string{}
	if cr.Spec.Deployment != nil && cr.Spec.Deployment.Labels != nil {
		labels = cr.Spec.Deployment.Labels
	}
	labels["app"] = GrafanaPodLabel
	return labels
}

func getVolumes(cr *v1alpha1.Grafana) []v13.Volume {
	volumes := []v13.Volume{}
	var volumeOptional bool = true

	// Volume to mount the config file from a config map
	volumes = append(volumes, v13.Volume{
		Name: GrafanaConfigName,
		VolumeSource: v13.VolumeSource{
			ConfigMap: &v13.ConfigMapVolumeSource{
				LocalObjectReference: v13.LocalObjectReference{
					Name: GrafanaConfigName,
				},
			},
		},
	})

	// Volume to store the logs
	volumes = append(volumes, v13.Volume{
		Name: GrafanaLogsVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Data volume
	volumes = append(volumes, v13.Volume{
		Name: GrafanaDataVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Volume to store the plugins
	volumes = append(volumes, v13.Volume{
		Name: GrafanaPluginsVolumeName,
		VolumeSource: v13.VolumeSource{
			EmptyDir: &v13.EmptyDirVolumeSource{},
		},
	})

	// Extra volumes for secrets
	for _, secret := range cr.Spec.Secrets {
		volumeName := fmt.Sprintf("secret-%s", secret)
		volumes = append(volumes, v13.Volume{
			Name: volumeName,
			VolumeSource: v13.VolumeSource{
				Secret: &v13.SecretVolumeSource{
					SecretName: secret,
					Optional:   &volumeOptional,
				},
			},
		})
	}

	// Extra volumes for config maps
	for _, configmap := range cr.Spec.ConfigMaps {
		volumeName := fmt.Sprintf("secret-%s", configmap)
		volumes = append(volumes, v13.Volume{
			Name: volumeName,
			VolumeSource: v13.VolumeSource{
				ConfigMap: &v13.ConfigMapVolumeSource{
					LocalObjectReference: v13.LocalObjectReference{
						Name: configmap,
					},
				},
			},
		})
	}
	return volumes
}

func getVolumeMounts(cr *v1alpha1.Grafana) []v13.VolumeMount {
	mounts := []v13.VolumeMount{}

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaConfigName,
		MountPath: "/etc/grafana/",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaDataVolumeName,
		MountPath: "/var/lib/grafana",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaPluginsVolumeName,
		MountPath: "/var/lib/grafana/plugins",
	})

	mounts = append(mounts, v13.VolumeMount{
		Name:      GrafanaLogsVolumeName,
		MountPath: "/var/log/grafana",
	})

	for _, secret := range cr.Spec.Secrets {
		mountName := fmt.Sprintf("secret-%s", secret)
		mounts = append(mounts, v13.VolumeMount{
			Name:      mountName,
			MountPath: config.SecretsMountDir + secret,
		})
	}

	for _, configmap := range cr.Spec.ConfigMaps {
		mountName := fmt.Sprintf("configmap-%s", configmap)
		mounts = append(mounts, v13.VolumeMount{
			Name:      mountName,
			MountPath: config.ConfigMapsMountDir + configmap,
		})
	}

	return mounts
}

func getProbe(cr *v1alpha1.Grafana, delay, timeout, failure int32) *v13.Probe {
	return &v13.Probe{
		Handler: v13.Handler{
			HTTPGet: &v13.HTTPGetAction{
				Path: GrafanaHealthEndpoint,
				Port: intstr.FromInt(GetGrafanaPort(cr)),
			},
		},
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		FailureThreshold:    failure,
	}
}

func getContainers(cr *v1alpha1.Grafana, configHash string) []v13.Container {
	containers := []v13.Container{}

	cfg := config.GetControllerConfig()
	image := cfg.GetConfigString(config.ConfigGrafanaImage, config.GrafanaImage)
	tag := cfg.GetConfigString(config.ConfigGrafanaImageTag, config.GrafanaVersion)

	containers = append(containers, v13.Container{
		Name:       "grafana",
		Image:      fmt.Sprintf("%s:%s", image, tag),
		Args:       []string{"-config=/etc/grafana/grafana.ini"},
		WorkingDir: "",
		Ports: []v13.ContainerPort{
			{
				Name:          "grafana-http",
				ContainerPort: int32(GetGrafanaPort(cr)),
				Protocol:      "TCP",
			},
		},
		Env: []v13.EnvVar{
			{
				Name:  LastConfigEnvVar,
				Value: configHash,
			},
			{
				Name: GrafanaAdminUserEnvVar,
				ValueFrom: &v13.EnvVarSource{
					SecretKeyRef: &v13.SecretKeySelector{
						LocalObjectReference: v13.LocalObjectReference{
							Name: GrafanaAdminSecretName,
						},
						Key: GrafanaAdminUserEnvVar,
					},
				},
			},
			{
				Name: GrafanaAdminPasswordEnvVar,
				ValueFrom: &v13.EnvVarSource{
					SecretKeyRef: &v13.SecretKeySelector{
						LocalObjectReference: v13.LocalObjectReference{
							Name: GrafanaAdminSecretName,
						},
						Key: GrafanaAdminPasswordEnvVar,
					},
				},
			},
		},
		Resources:                getResources(cr),
		VolumeMounts:             getVolumeMounts(cr),
		LivenessProbe:            getProbe(cr, 60, 30, 10),
		ReadinessProbe:           getProbe(cr, 5, 3, 1),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: "File",
		ImagePullPolicy:          "IfNotPresent",
	})

	// Add extra containers
	for _, container := range cr.Spec.Containers {
		container.VolumeMounts = getVolumeMounts(cr)
		containers = append(containers, container)
	}

	return containers
}

func getInitContainers(plugins string) []v13.Container {
	cfg := config.GetControllerConfig()
	image := cfg.GetConfigString(config.ConfigPluginsInitContainerImage, config.PluginsInitContainerImage)
	tag := cfg.GetConfigString(config.ConfigPluginsInitContainerTag, config.PluginsInitContainerTag)

	return []v13.Container{
		{
			Name:  GrafanaInitContainerName,
			Image: fmt.Sprintf("%s:%s", image, tag),
			Env: []v13.EnvVar{
				{
					Name:  "GRAFANA_PLUGINS",
					Value: plugins,
				},
			},
			Resources: v13.ResourceRequirements{},
			VolumeMounts: []v13.VolumeMount{
				{
					Name:      GrafanaPluginsVolumeName,
					ReadOnly:  false,
					MountPath: "/opt/plugins",
				},
			},
			TerminationMessagePath:   "/dev/termination-log",
			TerminationMessagePolicy: "File",
			ImagePullPolicy:          "IfNotPresent",
		},
	}
}

func getDeploymentSpec(cr *v1alpha1.Grafana, configHash, plugins string) v1.DeploymentSpec {
	return v1.DeploymentSpec{
		Replicas: getReplicas(cr),
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app": GrafanaPodLabel,
			},
		},
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Name:        GrafanaDeploymentName,
				Labels:      getPodLabels(cr),
				Annotations: getPodAnnotations(cr),
			},
			Spec: v13.PodSpec{
				Volumes:            getVolumes(cr),
				InitContainers:     getInitContainers(plugins),
				Containers:         getContainers(cr, configHash),
				ServiceAccountName: GrafanaServiceAccountName,
			},
		},
		Strategy: v1.DeploymentStrategy{
			Type:          "RollingUpdate",
			RollingUpdate: getRollingUpdateStrategy(),
		},
	}
}

func GrafanaDeployment(cr *v1alpha1.Grafana, configHash string) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      GrafanaDeploymentName,
			Namespace: cr.Namespace,
		},
		Spec: getDeploymentSpec(cr, configHash, ""),
	}
}

func GrafanaDeploymentSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaDeploymentName,
	}
}

func GrafanaDeploymentReconciled(cr *v1alpha1.Grafana, currentState *v1.Deployment, configHash, plugins string) *v1.Deployment {
	reconciled := currentState.DeepCopy()
	reconciled.Spec = getDeploymentSpec(cr, configHash, plugins)
	return reconciled
}