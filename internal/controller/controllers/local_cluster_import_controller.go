package controllers

import (
	"context"
	"fmt"
	"strings"

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
	localClusterImageSetName                          = "local-cluster-image-set"
	hubKubeConfigNamespace                            = "openshift-kube-apiserver"
	hubKubeConfigName                                 = "node-kubeconfigs"
	hubPullSecretNamespace                            = "openshift-config" // #nosec G101
	hubPullSecretName                                 = "pull-secret"      // #nosec G101
	importLocalClusterEnabledAnnotation               = "agent-install.openshift.io/enable-local-cluster-import"
	localClusterLabel                                 = "local-cluster"
	clusterIDLabel                                    = "clusterID"
)

type LocalClusterImportReconciler struct {
	client                 client.Client
	localClusterName       string
	log                    *logrus.Logger
	agentServiceConfigName string
}

func NewLocalClusterImportReconciler(client client.Client, agentServiceConfigName string, log *logrus.Logger) *LocalClusterImportReconciler {
	return &LocalClusterImportReconciler{
		client:                 client,
		log:                    log,
		agentServiceConfigName: agentServiceConfigName,
	}
}

func (r *LocalClusterImportReconciler) setReconciliationStatus(ctx context.Context, completed bool, reason string, message string) error {
	status := v1.ConditionFalse
	if completed {
		status = v1.ConditionTrue
	}
	agentServiceConfig := &aiv1beta1.AgentServiceConfig{}
	err := r.client.Get(ctx, types.NamespacedName{Name: "agent"}, agentServiceConfig)
	if err != nil {
		return errors.Wrapf(err, "Unable to fetch AgentServiceConfig while setting reconciliation status.")
	}
	conditionsv1.SetStatusConditionNoHeartbeat(&agentServiceConfig.Status.Conditions, conditionsv1.Condition{
		Type:    aiv1beta1.ConditionLocalClusterManaged,
		Status:  status,
		Reason:  reason,
		Message: message,
	})
	err = r.client.Status().Update(ctx, agentServiceConfig)
	if err != nil {
		r.log.Errorf("Unable to update status of AgentServiceConfig while attempting to set condition %s", aiv1beta1.ConditionReconcileCompleted)
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups=config.openshift.io,resources=dnses,verbs=get;list;watch
// +kubebuilder:rbac:groups=cluster.open-cluster-management.io,resources=managedclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=config.openshift.io,resources=proxies,verbs=get;list;watch

func (r *LocalClusterImportReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	defer func() {
		r.log.Info("AgentServiceConfig (LocalClusterImport) Reconcile ended")
	}()

	r.log.Info("AgentServiceConfig (LocalClusterImport) Reconcile started")

	instance := &aiv1beta1.AgentServiceConfig{}

	// NOTE: ignoring the Namespace that seems to get set on request when syncing on namespaced objects
	// when our AgentServiceConfig is ClusterScoped.
	if err := r.client.Get(ctx, types.NamespacedName{Name: req.NamespacedName.Name}, instance); err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		r.log.WithError(err).Error("Failed to get resource", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if instance.GetDeletionTimestamp().IsZero() {
		// Check to see if the ASC has our finalizer, if so, we must add it
		if !controllerutil.ContainsFinalizer(instance, agentServiceConfigLocalClusterImportFinalizerName) {
			controllerutil.AddFinalizer(instance, agentServiceConfigLocalClusterImportFinalizerName)
			if err := r.client.Update(ctx, instance); err != nil {
				return reconcile.Result{}, errors.Wrap(err, "failed to add finalizer to AgentServiceConfig")
			}
		}
	} else {
		err := r.ensureLocalClusterCRsDeleted(ctx)
		if err != nil {
			r.log.WithError(err).Error("failed to clean up local cluster CRs")
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(instance, agentServiceConfigLocalClusterImportFinalizerName)
		if err := r.client.Update(ctx, instance); err != nil {
			r.log.WithError(err).Error("failed to remove finalizer from AgentServiceConfig")
			return ctrl.Result{}, err
		}
		// Stop reconciliation as the item is being deleted
		r.log.Infof("Finalizer removed by local cluster import controller (local cluster CRs are cleared)")
		return ctrl.Result{}, nil
	}

	// Enable local cluster import if annotation allows this.
	_, importLocalClusterEnabled := instance.GetAnnotations()[importLocalClusterEnabledAnnotation]
	if !importLocalClusterEnabled {
		err := r.setReconciliationStatus(ctx, false, aiv1beta1.ReasonLocalClusterImportNotEnabled, "Local cluster import is not enabled")
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Unable to set reconciliation status of LocalClusterImport on AgentServiceConfig")
		}
		return ctrl.Result{}, nil
	}

	clusterVersion := &configv1.ClusterVersion{}
	namespacedName := types.NamespacedName{
		Name: "version",
	}
	err := r.client.Get(ctx, namespacedName, clusterVersion)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to fetch local cluster version")
	}

	err = r.getLocalManagedClusterName(ctx, clusterVersion)
	if err != nil {
		err = r.setReconciliationStatus(ctx, false, aiv1beta1.ReasonUnableToDetermineLocalClusterManagedStatus, err.Error())
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Unable to set reconciliation status of LocalClusterImport on AgentServiceConfig")
		}
		return ctrl.Result{}, errors.Wrap(err, "error while attempting to determine local cluster name")
	}
	hasLocalManagedCluster, err := r.hasLocalManagedCluster(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	if hasLocalManagedCluster {
		if err = r.importLocalCluster(ctx, instance, clusterVersion); err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to create managed cluster CRs")
		}
		err = r.setReconciliationStatus(ctx, true, aiv1beta1.ReasonLocalClusterManaged, "")
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Unable to set reconciliation status of LocalClusterImport on AgentServiceConfig")
		}
	} else {
		err = r.ensureLocalClusterCRsDeleted(ctx)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to clean up local cluster CRs")
		}
		err = r.setReconciliationStatus(ctx, false, aiv1beta1.ReasonLocalClusterNotManaged, "")
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "Unable to set reconciliation status of LocalClusterImport on AgentServiceConfig")
		}
		return ctrl.Result{}, nil
	}

	// Reconciliation complete
	return ctrl.Result{}, nil
}

func checkIsLocalManagedCluster(obj metav1.Object) bool {
	value, ok := obj.GetLabels()[localClusterLabel]
	return strings.ToLower(value) == "true" && ok
}

func CheckSecretName(obj metav1.Object) bool {
	return obj.GetNamespace() == hubKubeConfigNamespace && obj.GetName() == hubKubeConfigName || obj.GetNamespace() == hubPullSecretNamespace && obj.GetName() == hubPullSecretName
}

// SetupWithManager sets up the controller with the Manager.
func (r *LocalClusterImportReconciler) SetupWithManager(mgr ctrl.Manager) error {
	managedClusterPredicates := builder.WithPredicates(predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return checkIsLocalManagedCluster(e.Object) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return checkIsLocalManagedCluster(e.ObjectNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return checkIsLocalManagedCluster(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return checkIsLocalManagedCluster(e.Object) },
	})
	secretPredicates := builder.WithPredicates(predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return CheckSecretName(e.Object) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return CheckSecretName(e.ObjectNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return CheckSecretName(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return CheckSecretName(e.Object) },
	})
	enqueRequestForAgentServiceConfig := handler.EnqueueRequestsFromMapFunc(
		func(_ context.Context, _ client.Object) []reconcile.Request {
			return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: AgentServiceConfigName}}}
		},
	)
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiv1beta1.AgentServiceConfig{}).
		Owns(&hivev1.ClusterImageSet{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.Secret{}).
		Owns(&hiveext.AgentClusterInstall{}).
		Owns(&hivev1.ClusterDeployment{}).
		Watches(&configv1.Proxy{}, enqueRequestForAgentServiceConfig).
		Watches(&configv1.DNS{}, enqueRequestForAgentServiceConfig).
		Watches(&configv1.ClusterVersion{}, enqueRequestForAgentServiceConfig).
		Watches(&v1.Secret{}, enqueRequestForAgentServiceConfig, secretPredicates).
		Watches(&clusterv1.ManagedCluster{}, enqueRequestForAgentServiceConfig, managedClusterPredicates).
		Complete(r)
}

func (r *LocalClusterImportReconciler) getSecret(ctx context.Context, namespace string, name string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := r.client.Get(ctx, namespacedName, secret)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to fetch secret %s from namespace %s", name, namespace)
	}
	return secret, nil
}

func (r *LocalClusterImportReconciler) deleteClusterDeployment(ctx context.Context, namespace string, name string) error {
	clusterDeployment := &hivev1.ClusterDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := r.client.Delete(ctx, clusterDeployment); err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete ClusterDeployment %s in namespace %s", name, namespace)
	}
	return nil
}

func (r *LocalClusterImportReconciler) deleteAgentClusterInstall(ctx context.Context, namespace string, name string) error {
	agentClusterInstall := &hiveext.AgentClusterInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := r.client.Delete(ctx, agentClusterInstall); err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete AgentClusterInstall %s in namespace %s", name, namespace)
	}
	return nil
}

func (r *LocalClusterImportReconciler) deleteInfraEnv(ctx context.Context, namespace string, name string) error {
	infraEnv := &aiv1beta1.InfraEnv{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := r.client.Delete(ctx, infraEnv); err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete InfraEnv %s in namespace %s", name, namespace)
	}
	return nil
}

func (r *LocalClusterImportReconciler) deleteSecret(ctx context.Context, namespace string, name string) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := r.client.Delete(ctx, secret); err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete Secret %s in namespace %s", name, namespace)
	}
	return nil
}

func (r *LocalClusterImportReconciler) getLocalManagedClusterName(ctx context.Context, clusterVersion *configv1.ClusterVersion) error {
	if r.localClusterName != "" {
		return nil
	}
	managedClusters := &clusterv1.ManagedClusterList{}
	// This should result in a single ManagedCluster that matches our clusterID
	err := r.client.List(ctx, managedClusters, client.MatchingLabels{localClusterLabel: "true", clusterIDLabel: string(clusterVersion.Spec.ClusterID)})
	if err != nil {
		return err
	}
	count := len(managedClusters.Items)
	if count == 0 {
		r.localClusterName = ""
	}
	if count > 1 {
		return errors.Errorf("expectd a single matching local managed cluster and received %d instead", count)
	}
	if count == 1 {
		r.localClusterName = managedClusters.Items[0].Name
	}
	return nil
}

func (r *LocalClusterImportReconciler) hasLocalManagedCluster(ctx context.Context) (bool, error) {
	managedCluster := &clusterv1.ManagedCluster{}
	err := r.client.Get(ctx, types.NamespacedName{Name: r.localClusterName}, managedCluster)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// This will ensure that the ClusterDeployment and AgentClusterInstall are not present
// If they are present, they will be deleted
// No error will be returned if these are not found as this is the desired state.
func (r *LocalClusterImportReconciler) ensureLocalClusterCRsDeleted(ctx context.Context) error {
	if r.localClusterName == "" {
		r.log.Infof("skipping local cluster import cleanup as feature has not previously been enabled")
		return nil
	}
	err := r.deleteClusterDeployment(ctx, r.localClusterName, r.localClusterName)
	if err != nil && !k8serrors.IsNotFound(err) {
		r.log.Errorf("could not delete local cluster ClusterDeployment due to error %s", err.Error())
		return err
	}
	err = r.deleteAgentClusterInstall(ctx, r.localClusterName, r.localClusterName)
	if err != nil && !k8serrors.IsNotFound(err) {
		r.log.Errorf("could not delete local cluster AgentClusterInstall due to error %s", err.Error())
		return err
	}
	err = r.deleteInfraEnv(ctx, r.localClusterName, r.localClusterName)
	if err != nil && !k8serrors.IsNotFound(err) {
		r.log.Errorf("could not delete local cluster InfraEnv due to error %s", err.Error())
		return err
	}
	adminKubeConfigSecretName := fmt.Sprintf(adminKubeConfigStringTemplate, r.localClusterName)
	err = r.deleteSecret(ctx, r.localClusterName, adminKubeConfigSecretName)
	if err != nil && !k8serrors.IsNotFound(err) {
		r.log.Errorf("could not delete local cluster Secret %s due to error %s", adminKubeConfigSecretName, err.Error())
		return err
	}
	err = r.deleteSecret(ctx, r.localClusterName, hubPullSecretName)
	if err != nil && !k8serrors.IsNotFound(err) {
		r.log.Errorf("could not delete local cluster Secret %s due to error %s", hubPullSecretName, err.Error())
		return err
	}
	return nil
}

func (r *LocalClusterImportReconciler) createOrUpdateInfraEnv(ctx context.Context, pullSecretRefName string) error {
	infraEnv := aiv1beta1.InfraEnv{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.localClusterName,
			Namespace: r.localClusterName,
		},
		Spec: aiv1beta1.InfraEnvSpec{
			ClusterRef: &aiv1beta1.ClusterReference{
				Name:      r.localClusterName,
				Namespace: r.localClusterName,
			},
			PullSecretRef: &v1.LocalObjectReference{
				Name: pullSecretRefName,
			},
		},
	}
	mutateFn := func() error {
		return nil
	}
	createOrUpdateResult, err := controllerutil.CreateOrUpdate(ctx, r.client, &infraEnv, mutateFn)
	if err != nil {
		return errors.Wrap(err, "could not create InfraEnv")
	}
	if createOrUpdateResult != controllerutil.OperationResultNone {
		r.log.Infof("InfraEnv %s has been %s", infraEnv.Name, createOrUpdateResult)
	}
	return nil
}

func (r *LocalClusterImportReconciler) createOrUpdateClusterImageSet(ctx context.Context, release_image string, instance *aiv1beta1.AgentServiceConfig) error {
	clusterImageSet := hivev1.ClusterImageSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: localClusterImageSetName,
		},
	}
	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, &clusterImageSet, r.client.Scheme()); err != nil {
			return err
		}
		clusterImageSet.Spec.ReleaseImage = release_image
		return nil
	}
	createOrUpdateResult, err := controllerutil.CreateOrUpdate(ctx, r.client, &clusterImageSet, mutateFn)
	if err != nil {
		return errors.Wrap(err, "could not create cluster image set")
	}
	if createOrUpdateResult != controllerutil.OperationResultNone {
		r.log.Infof("ClusterImageSet %s has been %s", clusterImageSet.Name, createOrUpdateResult)
	}
	return nil
}

func (r *LocalClusterImportReconciler) createOrUpdateSecret(ctx context.Context, namespace string, name string, data map[string][]byte, instance *aiv1beta1.AgentServiceConfig) error {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, &secret, r.client.Scheme()); err != nil {
			return err
		}
		secret.Data = data
		return nil
	}
	createOrUpdateResult, err := controllerutil.CreateOrUpdate(ctx, r.client, &secret, mutateFn)
	if err != nil {
		return errors.Wrapf(err, "could not create or secret %s in namespace %s", name, namespace)
	}
	if createOrUpdateResult != controllerutil.OperationResultNone {
		r.log.Infof("Secret %s/%s has been %s", secret.Namespace, secret.Name, createOrUpdateResult)
	}
	return nil
}

func (r *LocalClusterImportReconciler) createOrUpdateAgentClusterInstall(ctx context.Context, numberOfControlPlaneNodes int, proxy *configv1.Proxy, instance *aiv1beta1.AgentServiceConfig) error {
	agentClusterInstall := hiveext.AgentClusterInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.localClusterName,
			Namespace: r.localClusterName,
		},
	}
	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, &agentClusterInstall, r.client.Scheme()); err != nil {
			return err
		}
		userManagedNetworkingActive := true
		agentClusterInstall.Spec.Networking.UserManagedNetworking = &userManagedNetworkingActive
		agentClusterInstall.Spec.ClusterDeploymentRef = v1.LocalObjectReference{
			Name: r.localClusterName,
		}
		agentClusterInstall.Spec.ImageSetRef = &hivev1.ClusterImageSetReference{
			Name: localClusterImageSetName,
		}
		agentClusterInstall.Spec.ProvisionRequirements = hiveext.ProvisionRequirements{
			ControlPlaneAgents: numberOfControlPlaneNodes,
		}
		if proxy != nil {
			agentClusterInstall.Spec.Proxy = &hiveext.Proxy{
				HTTPProxy:  proxy.Spec.HTTPProxy,
				HTTPSProxy: proxy.Spec.HTTPSProxy,
				NoProxy:    proxy.Spec.NoProxy,
			}
		}
		return nil
	}
	createOrUpdateResult, err := controllerutil.CreateOrUpdate(ctx, r.client, &agentClusterInstall, mutateFn)
	if err != nil {
		return errors.Wrap(err, "could not create or AgentClusterInstall")
	}
	if createOrUpdateResult != controllerutil.OperationResultNone {
		r.log.Infof("AgentClusterInstall %s/%s has been %s", agentClusterInstall.Namespace, agentClusterInstall.Name, createOrUpdateResult)
	}
	return nil
}

func (r *LocalClusterImportReconciler) createOrUpdateClusterDeployment(ctx context.Context, pullSecret *v1.Secret, dns *configv1.DNS, kubeConfigSecret *v1.Secret, instance *aiv1beta1.AgentServiceConfig) error {
	if pullSecret == nil {
		return errors.New("pull-secret is not defined, unable to perform local-cluster import")
	}
	if dns == nil {
		return errors.New("cluster dns is not defined, unable to perform local-cluster import")
	}
	if kubeConfigSecret == nil {
		return errors.New("kubeconfig secret is not defined, unable to perform local-cluster import")
	}
	if instance == nil {
		return errors.New("agentServiceConfig is not defined, unable to perform local-cluster import")
	}
	clusterDeployment := hivev1.ClusterDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.localClusterName,
			Namespace: r.localClusterName,
		},
	}
	mutateFn := func() error {
		if err := controllerutil.SetControllerReference(instance, &clusterDeployment, r.client.Scheme()); err != nil {
			return err
		}
		clusterDeployment.Spec.Installed = true
		clusterDeployment.Spec.ClusterMetadata = &hivev1.ClusterMetadata{
			ClusterID:                "",
			InfraID:                  "",
			AdminKubeconfigSecretRef: v1.LocalObjectReference{Name: fmt.Sprintf(adminKubeConfigStringTemplate, r.localClusterName)},
		}
		clusterDeployment.Spec.ClusterInstallRef = &hivev1.ClusterInstallLocalReference{
			Name:    r.localClusterName,
			Group:   hiveext.Group,
			Kind:    "AgentClusterInstall",
			Version: hiveext.Version,
		}
		clusterDeployment.Spec.Platform = hivev1.Platform{
			AgentBareMetal: &agent.BareMetalPlatform{
				AgentSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{"infraenv": "local-cluster"},
				},
			},
		}
		clusterDeployment.Spec.PullSecretRef = &v1.LocalObjectReference{
			Name: pullSecret.Name,
		}
		// By convention the local cluster has a domain in the form <cluster name>.<base domain>
		baseDomainParts := strings.Split(dns.Spec.BaseDomain, ".")
		if len(baseDomainParts) < 2 {
			return errors.Errorf("The base domain %s could not be parsed into cluster name + base domain, cannot proceed with import of local cluster", dns.Spec.BaseDomain)
		}
		clusterDeployment.Spec.ClusterName = baseDomainParts[0]
		clusterDeployment.Spec.BaseDomain = strings.Join(baseDomainParts[1:], ".")
		return nil
	}
	createOrUpdateResult, err := controllerutil.CreateOrUpdate(ctx, r.client, &clusterDeployment, mutateFn)
	if err != nil {
		return errors.Wrap(err, "could not create or ClusterDeployment")
	}
	if createOrUpdateResult != controllerutil.OperationResultNone {
		r.log.Infof("ClusterDeployment %s/%s has been %s", clusterDeployment.Namespace, clusterDeployment.Name, createOrUpdateResult)
	}
	return nil
}

func (r *LocalClusterImportReconciler) importLocalCluster(ctx context.Context, instance *aiv1beta1.AgentServiceConfig, clusterVersion *configv1.ClusterVersion) error {
	kubeConfigSecret, err := r.getSecret(ctx, hubKubeConfigNamespace, hubKubeConfigName)
	if err != nil {
		return errors.Wrap(err, "unable to fetch local cluster kubeconfigs")
	}

	pullSecret, err := r.getSecret(ctx, hubPullSecretNamespace, hubPullSecretName)
	if err != nil {
		return errors.Wrap(err, "unable to fetch pull secret")
	}

	dns := &configv1.DNS{}
	namespacedName := types.NamespacedName{
		Name: "cluster",
	}
	err = r.client.Get(ctx, namespacedName, dns)
	if err != nil {
		return errors.Wrap(err, "could not fetch DNS")
	}

	proxy := &configv1.Proxy{}
	namespacedName = types.NamespacedName{
		Name: "cluster",
	}
	err = r.client.Get(ctx, namespacedName, proxy)
	if err != nil {
		return errors.Wrap(err, "could not fetch proxy")
	}

	numberOfControlPlaneNodes := 0
	nodeList := &v1.NodeList{}
	err = r.client.List(ctx, nodeList)
	if err != nil {
		return errors.Wrap(err, "error while fetching nodes")
	}
	for _, node := range nodeList.Items {
		for nodeLabelKey := range node.Labels {
			if nodeLabelKey == "node-role.kubernetes.io/control-plane" {
				numberOfControlPlaneNodes++
			}
		}
	}

	if clusterVersion.Status.Desired.Image == "" {
		return errors.Wrap(err, "unable to determine desired release image")
	}

	err = r.createOrUpdateClusterImageSet(ctx, clusterVersion.Status.Desired.Image, instance)
	if err != nil {
		return err
	}

	// Store the kubeconfig data in the local cluster namespace
	err = r.createOrUpdateSecret(
		ctx,
		r.localClusterName,
		fmt.Sprintf(adminKubeConfigStringTemplate, r.localClusterName),
		map[string][]byte{"kubeconfig": kubeConfigSecret.Data["lb-ext.kubeconfig"]},
		instance)
	if err != nil {
		return err
	}

	// Store the pull secret in the local cluster namespace
	err = r.createOrUpdateSecret(
		ctx,
		r.localClusterName,
		pullSecret.Name,
		pullSecret.Data,
		instance)
	if err != nil {
		return err
	}

	err = r.createOrUpdateAgentClusterInstall(ctx, numberOfControlPlaneNodes, proxy, instance)
	if err != nil {
		return err
	}

	err = r.createOrUpdateClusterDeployment(ctx, pullSecret, dns, kubeConfigSecret, instance)
	if err != nil {
		return err
	}

	err = r.createOrUpdateInfraEnv(ctx, pullSecret.Name)
	if err != nil {
		return err
	}

	r.log.Info("completed import of hub cluster")
	return nil
}
