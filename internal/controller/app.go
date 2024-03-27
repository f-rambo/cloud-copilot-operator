package controller

import (
	"context"
	"strings"
	"time"

	operatoroceaniov1alpha1 "github.com/f-rambo/sailor/api/v1alpha1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Deployment  = "DEPLOYMENT"
	StatefulSet = "STATEFULSET"
	DaemonSet   = "DAEMONSET"
	Job         = "JOB"
	CronJob     = "CRONJOB"
)

func (r *AppReconciler) handlerApp(ctx context.Context, app *operatoroceaniov1alpha1.App) error {
	if app.Status.Status == operatoroceaniov1alpha1.StateSuccess || app.Status.Status == operatoroceaniov1alpha1.StateFailed {
		return nil
	}
	if app.Spec.Manifest == "" {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(10 * time.Second):
			err := r.checkAppPods(ctx, app)
			if err != nil {
				return err
			}
			if app.Status.Status == operatoroceaniov1alpha1.StateSuccess || app.Status.Status == operatoroceaniov1alpha1.StateFailed {
				err = r.Update(ctx, app)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
}

func (r *AppReconciler) checkAppPods(ctx context.Context, app *operatoroceaniov1alpha1.App) error {
	app.ClearLog()
	manifestBases := make([]operatoroceaniov1alpha1.ManifestBase, 0)
	err := yaml.Unmarshal([]byte(app.Spec.Manifest), &manifestBases)
	if err != nil {
		return err
	}
	podAllready := true
	for _, manifestBase := range manifestBases {
		labelSelector := ""
		switch strings.ToUpper(manifestBase.Kind) {
		case Deployment:
			deployment, err := r.ClientSet.AppsV1().Deployments(app.Spec.Namespace).Get(ctx, manifestBase.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			labelSelector = metav1.FormatLabelSelector(deployment.Spec.Selector)
			if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
				podAllready = false
			}
		case StatefulSet:
			statefulSet, err := r.ClientSet.AppsV1().StatefulSets(app.Spec.Namespace).Get(ctx, manifestBase.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if statefulSet.Status.ReadyReplicas != statefulSet.Status.Replicas {
				podAllready = false
			}
			labelSelector = metav1.FormatLabelSelector(statefulSet.Spec.Selector)
		case DaemonSet:
			daemonSet, err := r.ClientSet.AppsV1().DaemonSets(app.Spec.Namespace).Get(ctx, manifestBase.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if daemonSet.Status.NumberReady != daemonSet.Status.DesiredNumberScheduled {
				podAllready = false
			}
			labelSelector = metav1.FormatLabelSelector(daemonSet.Spec.Selector)
		case Job:
			job, err := r.ClientSet.BatchV1().Jobs(app.Spec.Namespace).Get(ctx, manifestBase.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if job.Status.Succeeded != *job.Spec.Completions {
				podAllready = false
			}
			labelSelector = metav1.FormatLabelSelector(job.Spec.Selector)
		case CronJob:
			cronJob, err := r.ClientSet.BatchV1().CronJobs(app.Spec.Namespace).Get(ctx, manifestBase.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			labelSelector = metav1.FormatLabelSelector(cronJob.Spec.JobTemplate.Spec.Selector)
		default:
		}
		podList, err := r.ClientSet.CoreV1().Pods(app.Spec.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return err
		}
		for _, pod := range podList.Items {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					podAllready = false
					break
				}
			}
			podLogOpts := corev1.PodLogOptions{}
			request := r.ClientSet.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
			podLog, err := request.Stream(ctx)
			if err != nil {
				return err
			}
			defer podLog.Close()
			buf := make([]byte, 1024)
			for {
				n, err := podLog.Read(buf)
				if err != nil {
					break
				}

				lower := strings.ToLower(string(buf[:n]))
				app.Spec.Log += lower
				if strings.Contains(lower, "error") {
					app.Spec.ErrLog += string(buf[:n])
				}
			}
		}
	}
	if app.Spec.ErrLog != "" {
		app.Status.Status = operatoroceaniov1alpha1.StateFailed
		return nil
	}
	if podAllready {
		app.Status.Status = operatoroceaniov1alpha1.StateSuccess
	}
	return nil
}
