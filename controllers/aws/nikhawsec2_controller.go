/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awsv1 "github.com/nikhilshinde5/nikhil-vmstate-operator/apis/aws/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// NikhAWSEC2Reconciler reconciles a NikhAWSEC2 object
type NikhAWSEC2Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type ConfigMap struct {
	InstanceType string `json:"instance_type"`
	ImageId      string `json:"image_id"`
}

//+kubebuilder:rbac:groups=aws.nikhilshinde.com,resources=nikhawsec2s,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aws.nikhilshinde.com,resources=nikhawsec2s/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aws.nikhilshinde.com,resources=nikhawsec2s/finalizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/finalizers;jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NikhAWSEC2 object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

const NikhAWSEC2Finalizer = "aws.nikhilshinde.com/finalizer"

func (r *NikhAWSEC2Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//_ = log.FromContext(ctx)
	log := ctrllog.FromContext(ctx)
	//log := r.Log.WithValues("NikhAWSEC2", req.NamespacedName)
	log.Info("Reconciling NikhAWSEC2s CRs")

	// Fetch the NikhAWSEC2 CR
	//NikhAWSEC2, err := services.FetchNikhAWSEC2CR(req.Name, req.Namespace)

	// Fetch the NikhAWSEC2 instance
	NikhAWSEC2 := &awsv1.NikhAWSEC2{}
	//ctrl.SetControllerReference(NikhAWSEC2, NikhAWSEC2, r.Scheme)
	log.Info(req.NamespacedName.Name)

	err := r.Client.Get(ctx, req.NamespacedName, NikhAWSEC2)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("NikhAWSEC2 resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get NikhAWSEC2.")
		return ctrl.Result{}, err
	}

	// Add const values for mandatory specs ( if left blank)
	// log.Info("Adding NikhAWSEC2 mandatory specs")
	// utils.AddBackupMandatorySpecs(NikhAWSEC2)
	// Check if the jobJob already exists, if not create a new one

	found := &batchv1.Job{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: NikhAWSEC2.Name + "create", Namespace: NikhAWSEC2.Namespace}, found)
	//log.Info(*found.)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Job
		job := r.JobForNikhAWSEC2(NikhAWSEC2, "create")
		log.Info("Creating a new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
		err = r.Client.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
			return ctrl.Result{}, err
		}
		// job created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get job")
		return ctrl.Result{}, err
	}

	// Check for any updates for redeployment
	/*applyChange := false

	// Ensure image name is correct, update image if required
	newInstanceIds := NikhAWSEC2.Spec.InstanceIds
	log.Info(newInstanceIds)

	newStartSchedule := NikhAWSEC2.Spec.StartSchedule
	log.Info(newStartSchedule)

	newImage := NikhAWSEC2.Spec.Image
	log.Info(newImage)

	var currentImage string = ""
	var currentStartSchedule string = ""
	var currentInstanceIds string = ""

	// Check existing schedule
	if found.Spec.Schedule != "" {
		currentStartSchedule = found.Spec.Schedule
	}

	if newStartSchedule != currentStartSchedule {
		found.Spec.Schedule = newStartSchedule
		applyChange = true
	}

	// Check existing image
	if found.Spec.JobTemplate.Spec.Template.Spec.Containers != nil {
		currentImage = found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
	}

	if newImage != currentImage {
		found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image = newImage
		applyChange = true
	}

	// Check instanceIds
	if found.Spec.JobTemplate.Spec.Template.Spec.Containers != nil {
		currentInstanceIds = found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env[0].Value
		log.Info(currentInstanceIds)
	}

	if newInstanceIds != currentInstanceIds {
		found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env[0].Value = newInstanceIds
		applyChange = true
	}

	log.Info(currentInstanceIds)
	log.Info(currentImage)
	log.Info(currentStartSchedule)

	log.Info(strconv.FormatBool(applyChange))

	if applyChange {
		log.Info(strconv.FormatBool(applyChange))
		err = r.Client.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update jobJob", "jobJob.Namespace", found.Namespace, "jobJob.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}*/

	// Update the NikhAWSEC2 status
	// TODO: Define what needs to be added in status. Currently adding just instanceIds
	/*if !reflect.DeepEqual(currentInstanceIds, NikhAWSEC2.Status.VMStartStatus) ||
		!reflect.DeepEqual(currentInstanceIds, NikhAWSEC2.Status.VMStopStatus) {
		NikhAWSEC2.Status.VMStartStatus = currentInstanceIds
		NikhAWSEC2.Status.VMStopStatus = currentInstanceIds
		err := r.Client.Status().Update(ctx, NikhAWSEC2)
		if err != nil {
			log.Error(err, "Failed to update NikhAWSEC2 status")
			return ctrl.Result{}, err
		}
	}*/
	// Check if the NikhAWSEC2 instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isNikhAWSEC2MarkedToBeDeleted := NikhAWSEC2.GetDeletionTimestamp() != nil
	if isNikhAWSEC2MarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(NikhAWSEC2, NikhAWSEC2Finalizer) {
			// Run finalization logic for NikhAWSEC2Finalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			log.Info(NikhAWSEC2.Name)
			log.Info("CR is marked for deletion")
			if err := r.finalizeNikhAWSEC2(ctx, NikhAWSEC2); err != nil {
				return ctrl.Result{}, err
			}

			// Remove NikhAWSEC2Finalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(NikhAWSEC2, NikhAWSEC2Finalizer)
			err := r.Client.Update(ctx, NikhAWSEC2)
			if err != nil {
				return ctrl.Result{}, err
			}
			log.Info("Finalizer removed")
			log.Info(NikhAWSEC2.Name)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(NikhAWSEC2, NikhAWSEC2Finalizer) {
		log.Info("Finalizer added again")
		log.Info(NikhAWSEC2.Name)
		controllerutil.AddFinalizer(NikhAWSEC2, NikhAWSEC2Finalizer)
		err = r.Client.Update(ctx, NikhAWSEC2)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
func (r *NikhAWSEC2Reconciler) finalizeNikhAWSEC2(ctx context.Context, NikhAWSEC2 *awsv1.NikhAWSEC2) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	log := ctrllog.FromContext(ctx)
	log.Info("Successfully finalized NikhAWSEC2")
	found := &batchv1.Job{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: NikhAWSEC2.Name + "delete", Namespace: NikhAWSEC2.Namespace}, found)
	//log.Info(*found.)
	if err != nil && errors.IsNotFound(err) {
		// Define a new job
		job := r.JobForNikhAWSEC2(NikhAWSEC2, "delete")
		log.Info("Creating a new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
		err = r.Client.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
			return err
		}
		// job created successfully - return and requeue
		return nil
	} else if err != nil {
		log.Error(err, "Failed to get job")
		return err
	}
	return nil
}

// Job Spec
func (r *NikhAWSEC2Reconciler) JobForNikhAWSEC2(NikhAWSEC2 *awsv1.NikhAWSEC2, command string) *batchv1.Job {

	// configMapData := make(map[string]string, 0)
	// configMapData["config.json"] = "ami-0d0ca2066b861631c"

	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	panic(err)
	// }
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err)
	// }

	// cm, err := clientset.CoreV1().ConfigMaps(NikhAWSEC2.Namespace).Get(NikhAWSEC2.Spec.ConfigMapName, metav1.GetOptions{})
	// if err != nil {
	// 	panic(err)
	// }

	// // Get the config.json data from the ConfigMap.
	// configData := cm.Data["config.json"]

	// // Unmarshal the JSON data into a Config object.
	// var configMap ConfigMap
	// err = json.Unmarshal([]byte(configData), &config)
	// if err != nil {
	// 	panic(err)
	// }

	jobName := NikhAWSEC2.Name + command
	job := &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: NikhAWSEC2.Namespace,
			Labels:    NikhAWSEC2Labels(NikhAWSEC2, "NikhAWSEC2"),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: NikhAWSEC2.Spec.ConfigMapName,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{{
						Name:  NikhAWSEC2.Name,
						Image: NikhAWSEC2.Spec.Image,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config",
								MountPath: "/opt/data",
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "ec2_command",
								Value: NikhAWSEC2.Spec.Command,
							},
							{
								Name:  "ec2_tag_key",
								Value: NikhAWSEC2.Spec.TagKey,
							},
							{
								Name:  "ec2_tag_value",
								Value: NikhAWSEC2.Spec.TagValue,
							},
							{
								Name:  "ec2_image_id",
								Value: "ami-0d0ca2066b861631c",
							},
							{
								Name:  "ec2_instance_type",
								Value: "t2.micro",
							},
							{
								Name: "AWS_ACCESS_KEY_ID",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-access-key-id",
									},
								},
							},
							{
								Name: "AWS_SECRET_ACCESS_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-secret-access-key",
									},
								},
							},
							{
								Name: "AWS_DEFAULT_REGION",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-default-region",
									},
								},
							}},
						ImagePullPolicy: NikhAWSEC2.Spec.ImagePullPolicy,
					}},

					RestartPolicy: "OnFailure",
				},
			},
		},
	}

	// Set NikhAWSEC2 instance as the owner and controller
	ctrl.SetControllerReference(NikhAWSEC2, job, r.Scheme)

	return job
}

// func setFromConfigMap(tag string) string {
// 	file, err := os.Open("/opt/data/config.json")
// 	if err != nil {
// 		fmt.Println("Error opening config file:", err)
// 		os.Exit(1)
// 	}
// 	defer file.Close()

// 	var config ConfigMap
// 	if err := json.NewDecoder(file).Decode(&config); err != nil {
// 		fmt.Println("Error decoding config:", err)
// 		os.Exit(1)
// 	}
// 	if tag == "ec2_image_id" {
// 		return config.ImageID
// 	} else if tag == "ec2_instance_type" {
// 		return config.InstanceType
// 	}
// 	return ""
// }

func NikhAWSEC2Labels(v *awsv1.NikhAWSEC2, tier string) map[string]string {
	return map[string]string{
		"app":           "NikhAWSEC2",
		"NikhAWSEC2_cr": v.Name,
		"tier":          tier,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NikhAWSEC2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&awsv1.NikhAWSEC2{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
