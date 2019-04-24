package tomcat

import (
	"github.com/go-openapi/spec"
	"context"

	jwsv1alpha1 "github.com/jws-image-operator/pkg/apis/jws/v1alpha1"

	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	imagev1 "github.com/openshift/api/image/v1"
	buildv1 "github.com/openshift/api/build/v1"
	corev1 "github.com/openshift/api/vendor/k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_tomcat")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Tomcat Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTomcat{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("tomcat-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Tomcat
	err = c.Watch(&source.Kind{Type: &jwsv1alpha1.Tomcat{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Tomcat
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &jwsv1alpha1.Tomcat{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTomcat{}

// ReconcileTomcat reconciles a Tomcat object
type ReconcileTomcat struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Tomcat object and makes changes based on the state read
// and what is in the Tomcat.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTomcat) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Tomcat")

	// Fetch the Tomcat tomcat
	tomcat := &jwsv1alpha1.Tomcat{}
	err := r.client.Get(context.TODO(), request.NamespacedName, tomcat)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Tomcat resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Tomcat")
		return reconcile.Result{}, err
	}

	/*
		// Define a new Pod object
		pod := newPodForCR(tomcat)

		// Set Tomcat tomcat as the owner and controller
		if err := controllerutil.SetControllerReference(tomcat, pod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this Pod already exists
		found := &corev1.Pod{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			err = r.client.Create(context.TODO(), pod)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Pod created successfully - don't requeue
			return reconcile.Result{}, nil
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Pod already exists - don't requeue
		reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	*/
	return reconcile.Result{}, nil
}

func (r *ReconcileTomcat) serviceForTomcat(t *jwsv1alpha1.Tomcat) *corev1.Service {

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
			Labels:    map[string]string{
				"application": t.Spec.ApplicationName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "ui",
				Port:       8080,
				TargetPort: 8080,
			}},
			Selector: map[string]string{
				"deploymentConfig": t.Spec.ApplicationName,
			},
		},
	}

	return service
}

func (r *ReconcileTomcat) deploymentConfigForTomcat(t *jwsv1alpha1.Tomcat) *appsv1.DeploymentConfig {

	terminationGracePeriodSeconds := int64(60)

	deploymentConfig := &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.openshift.io/v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.Name,
			Labels: map[string]string{
				"application": t.Spec.ApplicationName,
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRecreate,
			},
			Triggers: []appsv1.DeploymentTriggerPolicy{{
				Type: appsv1.DeploymentTriggerOnImageChange,
				ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
					Automatic:      true,
					ContainerNames: []string{t.Spec.ApplicationName},
					From: corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: t.Spec.ApplicationName + ":latest",
					},
				},
			},
				{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				}},
			Replicas: 1,
			Selector: map[string]string{
				"deploymentConfig": t.Spec.ApplicationName,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: t.Name,
					Labels: map[string]string{
						"application":      t.Spec.ApplicationName,
						"deploymentConfig": t.Spec.ApplicationName,
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Containers: []corev1.Container{{
						Name:            t.Spec.ApplicationName,
						Image:           t.Spec.ApplicationName,
						ImagePullPolicy: "Always",
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								Exec: &corev1.ExecAction{
									Command: []string{
										"/bin/bash",
										"-c",
										"curl --noproxy '*' -s -u ${JWS_ADMIN_USERNAME}:${JWS_ADMIN_PASSWORD} 'http://localhost:8080/manager/jmxproxy/?get=Catalina%3Atype%3DServer&att=stateName' |grep -iq 'stateName *= *STARTED'",
									},
								},
							},
						},
						Ports: []corev1.ContainerPort{{
							Name:          "jolokia",
							ContainerPort: 8778,
							Protocol:      corev1.ProtocolTCP,
						}, {
							Name:          "http",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						}},
						Env: []corev1.EnvVar{{
							Name:  "JWS_ADMIN_USERNAME",
							Value: t.Spec.JwsAdminUsername,
						}, {
							Name:  "JWS_ADMIN_PASSWORD",
							Value: t.Spec.JwsAdminPassword,
						}},
					}},
				},
			},
		},
	}

	return deploymentConfig
}

func (r *ReconcileTomcat) routeForTomcat(t *jwsv1alpha1.Tomcat) *routev1.Route {

	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "route.openshift.io/v1",
			Kind:       "Route",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   t.Name,
			Labels: map[string]string{
				"application": t.Spec.ApplicationName,
			},
			Annotations: map[string]string {
				"description": "Route for application's http service.",
			}
		},
		Spec: routev1.RouteSpec{
			Host: t.Spec.HostnameHttp,
			To: routev1.RouteTargetReference{
				Name: t.Spec.ApplicationName,
			},
		},
	}

	return route
}

func (r *ReconcileTomcat) imageStreamForTomcat(t *jwsv1alpha1.Tomcat) *imagev1.ImageStream {
	
	imageStream := &imagev1.ImageStream{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "image.openshift.io/v1",
			Kind:       "ImageStream",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   t.Name,
			Labels: map[string]string{
				"application": t.Spec.ApplicationName,
			},
		},
	}

	return imageStream
}

func (r *ReconcileTomcat) buildConfigForTomcat(t *jwsv1alpha1.Tomcat) *buildv1.BuildConfig {
	
	buildConfig := &buildv1.BuildConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "build.openshift.io/v1",
			Kind:       "BuildConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   t.Name,
			Labels: map[string]string{
				"application": t.Spec.ApplicationName,
			},
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{
				Source: buildv1.BuildSource{
					Type: "Git",
					Git: &buildv1.GitBuildSource{
						URI: t.Spec.SourceRepositoryUrl,
						Ref: t.Spec.SourceRepositoryRef,
					},
					ContextDir: t.Spec.ContextDir,
				},
				Strategy: buildv1.BuildStrategy{
					Type: "Source",
					SourceStrategy: &buildv1.SourceBuildStrategy{
						Env: []corev1.EnvVar{{
							Name: "MAVEN_MIRROR_URL",
							Value: t.Spec.MavenMirrorUrl,
						},{
							Name: "ARTIFACT_DIR",
							Value: t.Spec.ArtifactDir,
						}},
						ForcePull: true,
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Namespace: t.Spec.Namespace,
							Name: "jboss-webserver31-tomcat8-openshift:1.2",
						},
					},
				},
				Output: buildv1.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: t.Spec.ApplicationName + ":latest",
					},
				},
			},
			Triggers: []buildv1.BuildTriggerPolicy{{
				Type: "Github",
				GitHubWebHook: &buildv1.WebHookTrigger{
					Secret: t.Spec.GithubWebhookSecret,
				},
			},{
				Type: "Generic",
				GenericWebHook: &buildv1.WebHookTrigger{
					Secret: t.Spec.GenericWebhookSecret,
				},
			},{
				Type: "ImageChange",
				ImageChange: &buildv1.ImageChangeTrigger{},
			},{
				Type: "ConfigChange",
			}},
		},
	}

	return buildConfig
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *jwsv1alpha1.Tomcat) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}