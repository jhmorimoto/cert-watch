package util

import (
	"fmt"
	certwatchv1 "github.com/jhmorimoto/cert-watch/apis/certwatch/v1"
	v1 "k8s.io/api/batch/v1"
	apicorev1 "k8s.io/api/core/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProcessJob(certwatcher *certwatchv1.CertWatcher) (*v1.Job, error) {
	var job v1.Job
	var jobname string

	// Default job name
	if s, err := RandoHash(12); err == nil {
		jobname = certwatcher.Name + "-" + certwatcher.Spec.Actions.Job.Name + "-" + s
	} else {
		return nil, fmt.Errorf("unable to determine new job name: %s", err.Error())
	}

	// Default volume name and mountPath
	if certwatcher.Spec.Actions.Job.VolumeName == "" {
		certwatcher.Spec.Actions.Job.VolumeName = "certs"
	}
	if certwatcher.Spec.Actions.Job.MountPath == "" {
		certwatcher.Spec.Actions.Job.MountPath = "/workspace"
	}

	job = v1.Job{
		ObjectMeta: apimachineryv1.ObjectMeta{
			Namespace: certwatcher.Spec.Secret.Namespace,
			Name:      jobname,
		},
		Spec: certwatcher.Spec.Actions.Job.Spec,
	}

	// Create an additional volume in the pod spec
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, apicorev1.Volume{
		Name: certwatcher.Spec.Actions.Job.VolumeName,
		VolumeSource: apicorev1.VolumeSource{
			Secret: &apicorev1.SecretVolumeSource{
				SecretName: certwatcher.Spec.Secret.Name,
			},
		},
	})

	// Create a volumeMount in each container in the pod spec
	for i := range job.Spec.Template.Spec.Containers {
		job.Spec.Template.Spec.Containers[i].VolumeMounts = append(job.Spec.Template.Spec.Containers[i].VolumeMounts,
			apicorev1.VolumeMount{
				Name:      certwatcher.Spec.Actions.Job.VolumeName,
				MountPath: certwatcher.Spec.Actions.Job.MountPath,
			},
		)
	}

	return &job, nil
}
