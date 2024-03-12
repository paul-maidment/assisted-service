package controllers

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/go-multierror"
	configv1 "github.com/openshift/api/config/v1"
	hiveext "github.com/openshift/assisted-service/api/hiveextension/v1beta1"
	aiv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/openshift/hive/apis/hive/v1/agent"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	agentServiceConfigLocalClusterImportFinalizerName = "agentserviceconfig." + aiv1beta1.Group + "/local-cluster-import-deprovision"
)

type LocalClusterImportReconciler struct {
	Client                 client.Client
	LocalClusterName       string
	Log                    *logrus.Logger
	AgentServiceConfigName string
}

func NewLocalClusterImportReconciler(client client.Client, localClusterName string, agentServiceConfigName string, log *logrus.Logger) *LocalClusterImportReconciler {
	return &LocalClusterImportReconciler{
		Client:                 client,
		LocalClusterName:       localClusterName,
		Log:                    log,
		AgentServiceConfigName: agentServiceConfigName,
	}
}

func (r *LocalClusterImportReconciler) setReconciliationStatus(agentServiceConfig *aiv1beta1.AgentServiceConfig, completed bool, reason string) error {
	status := v1.ConditionFalse
	if completed {
		status = v1.ConditionTrue
	}
	conditionsv1.SetStatusConditionNoHeartbeat(&agentServiceConfig.Status.Conditions, conditionsv1.Condition{
		Type:    aiv1beta1.ConditionReconcileCompleted,
		Status:  status,
		Reason:  aiv1beta1.ReasonReconcileSucceeded,
		Message: reason, 
	})
	err := r.Client.Update(context.Background(), agentServiceConfig)
	if err != nil {
		r.Log.Errorf("Unable to update status of ASC while attempting to set condition %s", aiv1beta1.ConditionReconcileCompleted)
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups=config.openshift.io,resources=dnses,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies,verbs=get;list;watch

func (r *LocalClusterImportReconciler) Reconcile(origCtx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var asc ASC
	ctx := addRequestIdIfNeeded(origCtx)
	defer func() {
		r.Log.Info("AgentServiceConfig (LocalClusterImport) Reconcile ended")
	}()

	r.Log.Info("AgentServiceConfig (LocalClusterImport) Reconcile started")

	instance := &aiv1beta1.AgentServiceConfig{}
	asc.Client = r.Client
	asc.Object = instance
	asc.spec = &instance.Spec
	asc.conditions = &instance.Status.Conditions

	// NOTE: ignoring the Namespace that seems to get set on request when syncing on namespaced objects
	// when our AgentServiceConfig is ClusterScoped.
	if err := r.Client.Get(ctx, types.NamespacedName{Name: req.NamespacedName.Name}, instance); err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		r.Log.WithError(err).Error("Failed to get resource", req.NamespacedName)
		return ctrl.Result{}, err
	}

	err := r.ensureFinalizers(ctx, instance, agentServiceConfigLocalClusterImportFinalizerName)
	if err != nil {
		r.Log.Errorf("Could not apply finalizers on local cluster import due to error %s", err.Error())
		_ = r.setReconciliationStatus(instance, false, fmt.Sprintf("An error occurred: %s", err.Error()))
		return ctrl.Result{Requeue: true}, err
	}
	_ = r.setReconciliationStatus(instance, true, "LocalClusterImport reconcile completed without error.")
	return ctrl.Result{}, err
}

func (r *LocalClusterImportReconciler) handleManagedCluster() error {
	hasManagedCluster, err := r.HasLocalManagedCluster()
	if err != nil {
		r.Log.WithError(err).Error("error while attempting to determine presence of managed cluster")
		return err
	}
	if hasManagedCluster {
		if err = r.createLocalClusterCRs(); err != nil {
			r.Log.WithError(err).Error("failed to create managed cluster CRs")
			return err
		}
		return nil
	}
	err = r.deleteLocalClusterCRs()
	if err != nil {
		r.Log.WithError(err).Error("failed to clean up local cluster CRs")
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) ensureFinalizers(ctx context.Context, agentServiceConfig *aiv1beta1.AgentServiceConfig, finalizerName string) error {
	if agentServiceConfig.GetDeletionTimestamp().IsZero() {
		r.Log.Infof("Zero deletion timestamp")
		// Check to see if the ASC has our finalizer, if so, we must add it
		if !controllerutil.ContainsFinalizer(agentServiceConfig, finalizerName) {
			controllerutil.AddFinalizer(agentServiceConfig, finalizerName)
			if err := r.Client.Update(ctx, agentServiceConfig); err != nil {
				r.Log.WithError(err).Error("failed to add finalizer to AgentServiceConfig")
				return err
			}
		}
		if err := r.handleManagedCluster(); err != nil {
			r.Log.WithError(err).Error("error while handling managed cluster")
			return err
		}
	} else {
		r.Log.Infof("Non zero deletion timestamp")
		err := r.deleteLocalClusterCRs()
		if err != nil {
			r.Log.WithError(err).Error("failed to clean up local cluster CRs")
			return err
		}
		controllerutil.RemoveFinalizer(agentServiceConfig, finalizerName)
		if err := r.Client.Update(ctx, agentServiceConfig); err != nil {
			r.Log.WithError(err).Error("failed to remove finalizer from AgentServiceConfig")
			return err
		}
	}
	return nil
}

func (r *LocalClusterImportReconciler) createLocalClusterCRs() error {
	err := r.ImportLocalCluster()
	if err != nil {
		// Failure to import the local cluster is not fatal but we should warn in the log.
		r.Log.Warnf("Could not import local cluster due to error %s", err.Error())
	}
	return nil
}

func (r *LocalClusterImportReconciler) deleteLocalClusterCRs() error {
	err := r.DeleteLocalClusterImportCRs()
	if err != nil {
		// Failed to clean up the local cluster import CR's
		r.Log.Warnf("Could not clean up CR's for local cluster import due to error %s", err.Error())
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocalClusterImportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ingressCMPredicates := builder.WithPredicates(predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return checkIngressCMName(e.Object) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return checkIngressCMName(e.ObjectNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return checkIngressCMName(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return checkIngressCMName(e.Object) },
	})
	ingressCMHandler := handler.EnqueueRequestsFromMapFunc(
		func(_ context.Context, _ client.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: AgentServiceConfigName}}}
		},
	)
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiv1beta1.AgentServiceConfig{}).
		Watches(&clusterv1.ManagedCluster{}, ingressCMHandler, ingressCMPredicates).
		Complete(r)
}

func (r *LocalClusterImportReconciler) GetManagedCluster(name string) (*clusterv1.ManagedCluster, error) {
	managedCluster := &clusterv1.ManagedCluster{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      name,
	}
	err := r.Client.Get(context.Background(), namespacedName, managedCluster)
	if err != nil {
		return nil, err
	}
	return managedCluster, nil
}

func (r *LocalClusterImportReconciler) GetAgentServiceConfig() (*aiv1beta1.AgentServiceConfig, error) {
	agentServiceConfig := &aiv1beta1.AgentServiceConfig{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      r.AgentServiceConfigName,
	}
	err := r.Client.Get(context.Background(), namespacedName, agentServiceConfig)
	if err != nil {
		return nil, err
	}
	return agentServiceConfig, nil
}

func (r *LocalClusterImportReconciler) GetClusterDeployment(namespace string, name string) (*hivev1.ClusterDeployment, error) {
	clusterDeployment := &hivev1.ClusterDeployment{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := r.Client.Get(context.Background(), namespacedName, clusterDeployment)
	if err != nil {
		return nil, err
	}
	return clusterDeployment, nil
}

func (r *LocalClusterImportReconciler) GetAgentClusterInstall(namespace string, name string) (*hiveext.AgentClusterInstall, error) {
	agentClusterInstall := &hiveext.AgentClusterInstall{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := r.Client.Get(context.Background(), namespacedName, agentClusterInstall)
	if err != nil {
		return nil, err
	}
	return agentClusterInstall, nil
}

func (r *LocalClusterImportReconciler) GetNamespace(name string) (*v1.Namespace, error) {
	ns := &v1.Namespace{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      name,
	}
	err := r.Client.Get(context.Background(), namespacedName, ns)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func (r *LocalClusterImportReconciler) CreateNamespace(name string) error {
	ns := &v1.Namespace{}
	ns.Name = name
	err := r.Client.Create(context.Background(), ns)
	if err != nil {
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) GetSecret(namespace string, name string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := r.Client.Get(context.Background(), namespacedName, secret)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to fetch secret %s from namespace %s", name, namespace)
	}
	return secret, nil
}

func (r *LocalClusterImportReconciler) GetClusterVersion(name string) (*configv1.ClusterVersion, error) {
	clusterVersion := &configv1.ClusterVersion{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      name,
	}
	err := r.Client.Get(context.Background(), namespacedName, clusterVersion)
	if err != nil {
		return nil, err
	}
	return clusterVersion, nil
}

func (r *LocalClusterImportReconciler) GetClusterImageSet(name string) (*hivev1.ClusterImageSet, error) {
	clusterImageSet := &hivev1.ClusterImageSet{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      name,
	}
	err := r.Client.Get(context.Background(), namespacedName, clusterImageSet)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to fetch cluster image set %s", name)
	}
	return clusterImageSet, nil
}

func (r *LocalClusterImportReconciler) CreateAgentClusterInstall(agentClusterInstall *hiveext.AgentClusterInstall) error {
	err := r.Client.Create(context.Background(), agentClusterInstall)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) CreateSecret(namespace string, secret *v1.Secret) error {
	err := r.Client.Create(context.Background(), secret)
	if err != nil {
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) CreateClusterImageSet(clusterImageSet *hivev1.ClusterImageSet) error {
	err := r.Client.Create(context.Background(), clusterImageSet)
	if err != nil {
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) CreateClusterDeployment(clusterDeployment *hivev1.ClusterDeployment) error {
	err := r.Client.Create(context.Background(), clusterDeployment)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) GetNodes() (*v1.NodeList, error) {
	nodeList := &v1.NodeList{}
	err := r.Client.List(context.Background(), nodeList)
	if err != nil {
		return nil, err
	}
	return nodeList, nil
}

func (r *LocalClusterImportReconciler) GetNumberOfControlPlaneNodes() (int, error) {
	// Determine the number of control plane agents we have
	// Control plane nodes have a specific label that we can look for.,
	numberOfControlPlaneNodes := 0
	nodeList, err := r.GetNodes()
	if err != nil {
		return 0, err
	}
	for _, node := range nodeList.Items {
		for nodeLabelKey := range node.Labels {
			if nodeLabelKey == "node-role.kubernetes.io/control-plane" {
				numberOfControlPlaneNodes++
			}
		}
	}
	return numberOfControlPlaneNodes, nil
}

func (r *LocalClusterImportReconciler) GetClusterDNS() (*configv1.DNS, error) {
	dns := &configv1.DNS{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      "cluster",
	}
	err := r.Client.Get(context.Background(), namespacedName, dns)
	if err != nil {
		return nil, err
	}
	return dns, nil
}

func (r *LocalClusterImportReconciler) GetClusterProxy() (*configv1.Proxy, error) {
	proxy := &configv1.Proxy{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      "cluster",
	}
	err := r.Client.Get(context.Background(), namespacedName, proxy)
	if err != nil {
		return nil, err
	}
	return proxy, nil
}

func (r *LocalClusterImportReconciler) DeleteClusterDeployment(namespace string, name string) error {
	clusterDeployment, err := r.GetClusterDeployment(namespace, name)
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "error fetching ClusterDeployment %s in namespace %s for delete", name, namespace)
	}
	err = r.Client.Delete(context.Background(), clusterDeployment)
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "error deleting ClusterDeployment %s in namespace %s", name, namespace)
	}
	return nil
}

func (r *LocalClusterImportReconciler) DeleteAgentClusterInstall(namespace string, name string) error {
	agentClusterInstall, err := r.GetAgentClusterInstall(namespace, name)
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "error fetching AgentClusterInstall %s in namespace %s for delete", name, namespace)
	}
	err = r.Client.Delete(context.Background(), agentClusterInstall)
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "error deleting AgentClusterInstall %s in namespace %s", name, namespace)
	}
	return nil
}

func (r *LocalClusterImportReconciler) HasLocalManagedCluster() (bool, error) {
	mc, err := r.GetManagedCluster(r.LocalClusterName)
	if err != nil && !k8serrors.IsNotFound(err) {
		r.Log.Errorf("unable to fetch ManagedCluster for local-cluster due to error: %s", err.Error())
		return false, err
	}
	return mc != nil, nil
}

func (r *LocalClusterImportReconciler) createClusterImageSet(release_image string) error {
	var err error
	clusterImageSet := hivev1.ClusterImageSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "local-cluster-image-set",
		},
		Spec: hivev1.ClusterImageSetSpec{
			ReleaseImage: release_image,
		},
	}
	err = r.CreateClusterImageSet(&clusterImageSet)
	if err != nil {
		r.Log.Errorf("unable to create ClusterImageSet due to error: %s", err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) createAdminKubeConfig(kubeConfigSecret *v1.Secret) error {
	var err error
	// Store the kubeconfig data in the local cluster namespace
	localClusterSecret := v1.Secret{}
	localClusterSecret.Name = fmt.Sprintf("%s-admin-kubeconfig", r.LocalClusterName)
	localClusterSecret.Namespace = r.LocalClusterName
	localClusterSecret.Data = make(map[string][]byte)
	localClusterSecret.Data["kubeconfig"] = kubeConfigSecret.Data["lb-ext.kubeconfig"]
	err = r.CreateSecret(r.LocalClusterName, &localClusterSecret)
	if err != nil {
		r.Log.Errorf("to store secret due to error %s", err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) createLocalClusterPullSecret(sourceSecret *v1.Secret) error {
	var err error
	// Store the pull secret in the local cluster namespace
	hubPullSecret := v1.Secret{}
	hubPullSecret.Name = sourceSecret.Name
	hubPullSecret.Namespace = r.LocalClusterName
	hubPullSecret.Data = make(map[string][]byte)
	// .dockerconfigjson is double base64 encoded for some reason.
	// simply obtaining the secret above will perform one layer of decoding.
	// we need to manually perform another to ensure that the data is correctly copied to the new secret.
	hubPullSecret.Data[".dockerconfigjson"], err = base64.StdEncoding.DecodeString(string(sourceSecret.Data[".dockerconfigjson"]))
	if err != nil {
		r.Log.Errorf("unable to decode base64 pull secret data due to error %s", err.Error())
		return err
	}
	hubPullSecret.OwnerReferences = []metav1.OwnerReference{}
	err = r.CreateSecret(r.LocalClusterName, &hubPullSecret)
	if err != nil {
		r.Log.Errorf("unable to store hub pull secret due to error %s", err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) createAgentClusterInstall(numberOfControlPlaneNodes int, proxy *configv1.Proxy) error {
	//Create an AgentClusterInstall in the local cluster namespace
	userManagedNetworkingActive := true
	agentClusterInstall := &hiveext.AgentClusterInstall{
		Spec: hiveext.AgentClusterInstallSpec{
			Networking: hiveext.Networking{
				UserManagedNetworking: &userManagedNetworkingActive,
			},
			ClusterDeploymentRef: v1.LocalObjectReference{
				Name: r.LocalClusterName,
			},
			ImageSetRef: &hivev1.ClusterImageSetReference{
				Name: "local-cluster-image-set",
			},
			ProvisionRequirements: hiveext.ProvisionRequirements{
				ControlPlaneAgents: numberOfControlPlaneNodes,
			},
		},
	}
	agentClusterInstall.Namespace = r.LocalClusterName
	agentClusterInstall.Name = r.LocalClusterName
	if proxy != nil {
		agentClusterInstall.Spec.Proxy = &hiveext.Proxy{
			HTTPProxy:  proxy.Spec.HTTPProxy,
			HTTPSProxy: proxy.Spec.HTTPSProxy,
			NoProxy:    proxy.Spec.NoProxy,
		}
	}

	err := r.CreateAgentClusterInstall(agentClusterInstall)
	if err != nil {
		r.Log.Errorf("could not create AgentClusterInstall due to error %s", err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) createClusterDeployment(pullSecret *v1.Secret, dns *configv1.DNS, kubeConfigSecret *v1.Secret, agentServiceConfig *aiv1beta1.AgentServiceConfig) error {
	if pullSecret == nil {
		r.Log.Errorf("Skipping creation of cluster deployment due to nil pullsecret")
		return nil
	}
	if dns == nil {
		r.Log.Errorf("Skipping creation of cluster deployment due to nil dns")
		return nil
	}
	if kubeConfigSecret == nil {
		r.Log.Errorf("Skipping creation of cluster deployment due to nil kubeconfigsecret")
		return nil
	}
	if agentServiceConfig == nil {
		r.Log.Errorf("Skipping creation of cluster deployment due to nil agentServiceConfig")
		return nil
	}
	// Create a cluster deployment in the local cluster namespace
	clusterDeployment := &hivev1.ClusterDeployment{
		Spec: hivev1.ClusterDeploymentSpec{
			Installed: true, // Let assisted know to import this cluster
			ClusterMetadata: &hivev1.ClusterMetadata{
				ClusterID:                "",
				InfraID:                  "",
				AdminKubeconfigSecretRef: v1.LocalObjectReference{Name: fmt.Sprintf("%s-admin-kubeconfig", r.LocalClusterName)},
			},
			ClusterInstallRef: &hivev1.ClusterInstallLocalReference{
				Name:    r.LocalClusterName,
				Group:   "extensions.hive.openshift.io",
				Kind:    "AgentClusterInstall",
				Version: "v1beta1",
			},
			Platform: hivev1.Platform{
				AgentBareMetal: &agent.BareMetalPlatform{
					AgentSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"infraenv": "local-cluster"},
					},
				},
			},
			PullSecretRef: &v1.LocalObjectReference{
				Name: pullSecret.Name,
			},
		},
	}
	clusterDeployment.Name = r.LocalClusterName
	clusterDeployment.Namespace = r.LocalClusterName
	clusterDeployment.Spec.ClusterName = r.LocalClusterName
	clusterDeployment.Spec.BaseDomain = dns.Spec.BaseDomain
	agentClusterInstallOwnerRef := metav1.OwnerReference{
		Kind:       "AgentServiceConfig",
		APIVersion: "agent-install.openshift.io/v1beta1",
		Name:       agentServiceConfig.Name,
		UID:        agentServiceConfig.UID,
	}
	clusterDeployment.OwnerReferences = []metav1.OwnerReference{agentClusterInstallOwnerRef}
	err := r.CreateClusterDeployment(clusterDeployment)
	if err != nil {
		r.Log.Errorf("could not create ClusterDeployment due to error %s", err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) createNamespace(name string) error {
	err := r.CreateNamespace(name)
	if err != nil {
		r.Log.Errorf("could not create Namespace due to error %s", err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) DeleteLocalClusterImportCRs() error {
	err := r.DeleteClusterDeployment(r.LocalClusterName, r.LocalClusterName)
	if err != nil {
		r.Log.Errorf("could not delete local ClusterDeployment due to error %s", err.Error())
		return err
	}
	err = r.DeleteAgentClusterInstall(r.LocalClusterName, r.LocalClusterName)
	if err != nil {
		r.Log.Errorf("could not delete local AgentClusterInstall due to error %s", err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) ImportLocalCluster() error {

	var errorList error

	clusterVersion, err := r.GetClusterVersion("version")
	if err != nil {
		r.Log.Errorf("unable to find cluster version info due to error: %s", err.Error())
		errorList = multierror.Append(nil, err)
		return errorList
	}

	release_image := ""
	if clusterVersion != nil && clusterVersion.Status.History[0].State != configv1.CompletedUpdate {
		message := "the release image in the cluster version is not ready yet, please wait for installation to complete and then restart the assisted-service"
		r.Log.Error(message)
		errorList = multierror.Append(errorList, errors.New(message))
	}

	kubeConfigSecret, err := r.GetSecret("openshift-kube-apiserver", "node-kubeconfigs")
	if err != nil {
		r.Log.Errorf("unable to fetch local cluster kubeconfigs due to error %s", err.Error())
		errorList = multierror.Append(errorList, err)
	}

	pullSecret, err := r.GetSecret("openshift-machine-api", "pull-secret")
	if err != nil {
		r.Log.Errorf("unable to fetch pull secret due to error %s", err.Error())
		errorList = multierror.Append(errorList, err)
	}

	dns, err := r.GetClusterDNS()
	if err != nil {
		r.Log.Errorf("could not fetch DNS due to error %s", err.Error())
		errorList = multierror.Append(errorList, err)
	}

	proxy, err := r.GetClusterProxy()
	if err != nil {
		r.Log.Errorf("could not fetch proxy due to error %s", err.Error())
		errorList = multierror.Append(errorList, err)
	}

	numberOfControlPlaneNodes, err := r.GetNumberOfControlPlaneNodes()
	if err != nil {
		r.Log.Errorf("unable to determine the number of control plane nodes due to error %s", err.Error())
		errorList = multierror.Append(errorList, err)
	}

	if clusterVersion.Status.History[0].Image == "" {
		r.Log.Error("unable to determine release image")
		errorList = multierror.Append(errorList, err)
	}

	// If we already have errors before we start writing things, then let's stop
	// better to let the user fix the problems than have to delete things later.
	if errorList != nil {
		return errorList
	}

	release_image = clusterVersion.Status.History[0].Image

	err = r.createClusterImageSet(release_image)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		errorList = multierror.Append(errorList, err)
	}

	err = r.createNamespace(r.LocalClusterName)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		errorList = multierror.Append(errorList, err)
	}

	if kubeConfigSecret != nil {
		err = r.createAdminKubeConfig(kubeConfigSecret)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			errorList = multierror.Append(errorList, err)
		}
	}

	if pullSecret != nil {
		err = r.createLocalClusterPullSecret(pullSecret)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			errorList = multierror.Append(errorList, err)
		}
	}

	if numberOfControlPlaneNodes > 0 {
		err := r.createAgentClusterInstall(numberOfControlPlaneNodes, proxy)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			errorList = multierror.Append(errorList, err)
		}

		asc, err := r.GetAgentServiceConfig()
		if err != nil {
			r.Log.Errorf("failed to fetch AgentServiceConfig due to error %s", err.Error())
			errorList = multierror.Append(errorList, err)
		}

		err = r.createClusterDeployment(pullSecret, dns, kubeConfigSecret, asc)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			errorList = multierror.Append(errorList, err)
		}
	}

	if errorList != nil {
		return errorList
	}

	r.Log.Info("completed Day2 import of hub cluster")
	return nil
}
