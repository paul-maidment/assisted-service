package localclusterimport

import (
	"encoding/base64"
	"errors"
	"fmt"

	hiveext "github.com/openshift/assisted-service/api/hiveextension/v1beta1"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/openshift/hive/apis/hive/v1/agent"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	configv1 "github.com/openshift/api/config/v1"
)

const LocalClusterNamespace = "local-cluster"

type LocalClusterImport struct {
	clusterImportOperations ClusterImportOperations
	log                          *logrus.Logger
}

func NewLocalClusterImport(localClusterImportOperations ClusterImportOperations, log *logrus.Logger) LocalClusterImport {
	return LocalClusterImport{clusterImportOperations: localClusterImportOperations, log: log}
}

func (i *LocalClusterImport) DeleteAllImportCRsForCluster() error {
	err := i.clusterImportOperations.DeleteClusterDeployment(LocalClusterNamespace, LocalClusterNamespace + "-cluster-deployment")
	if err != nil {
		return err
	}
	err = i.clusterImportOperations.DeleteAgentClusterInstall(LocalClusterNamespace, LocalClusterNamespace + "-cluster-install")
	if err != nil {
		return err
	}
	err = i.clusterImportOperations.DeleteClusterImageSet("local-cluster-image-set")
	if err != nil {
		return err
	}
	err = i.clusterImportOperations.DeleteSecret(LocalClusterNamespace, fmt.Sprintf("%s-admin-kubeconfig", LocalClusterNamespace))
	if err != nil {
		return err
	}
	err = i.clusterImportOperations.DeleteSecret(LocalClusterNamespace, "pull-secret")
	if err != nil {
		return err
	}
	return nil
}

func (i *LocalClusterImport) shouldImportLocalCluster(clusterVersion *configv1.ClusterVersion) (bool, error) {

	// Determine whether or not a change has been detected in cluster version, if so an update is required.
	localClusterImageSet, err := i.clusterImportOperations.GetClusterImageSet("local-cluster-image-set")
	if err != nil && !k8serrors.IsNotFound(err) {
		i.log.Errorf("hubClusterDay2Importer: There was an error fetching the ClusterImageSet while checking for need to register local cluster.")
	}
	updateRequired := false
	if clusterVersion.Status.History[0].State != configv1.CompletedUpdate {
		i.log.Info("hubClusterDay2Importer: Waiting for current openshift upgrade to complete, please restart assisted operator on completion:")
		return false, nil
	}
	if localClusterImageSet != nil && clusterVersion.Status.History[0].Image != localClusterImageSet.Spec.ReleaseImage {
		i.log.Info("hubClusterDay2Importer: an update to the deployed cluster version has been detected, re-importing:")
		updateRequired = true
		// It is simpler to delete all of the CRs associated with import and start over.
		err = i.DeleteAllImportCRsForCluster()
		{
			if err != nil {
				i.log.Errorf("hubClusterDay2Importer: During update, failed to delete all import CR's for cluster")
				return updateRequired, err
			}
		}
		return true, nil
	}

	// The presence of agentClusterInstall for the local cluster will determine whether or not we need to perform an import
	// If an update has been flagged then we will skip these checks (as we deleted the CR's we are checking in the previous step)
	agentClusterInstall, err := i.clusterImportOperations.GetAgentClusterInstall(LocalClusterNamespace, LocalClusterNamespace+"-cluster-install")
	if err != nil && !k8serrors.IsNotFound(err) {
		i.log.Errorf("hubClusterDay2Importer: There was an error fetching the AgentClusterInstall while checking for need to register local cluster.")
		return false, err
	}
	if agentClusterInstall != nil {
		i.log.Infof("hubClusterDay2Importer: Found AgentClusterInstall for hub cluster, assuming that hub has been correctly registered for ZTP day 2 operations: %s", agentClusterInstall.Name)
		return false, nil
	}
	return true, nil
}

func (i *LocalClusterImport) ImportLocalCluster() error {
	clusterVersion, err := i.clusterImportOperations.GetClusterVersion("version")
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Unable to find cluster version info due to error: %s", err.Error())
		return err
	}

	shouldImportLocalCluster, err := i.shouldImportLocalCluster(clusterVersion)
	if err != nil {
		return err
	}
	if !shouldImportLocalCluster {
		message := "hubClusterDay2Importer: no need to import local cluster as registration already detected"
		i.log.Info(message)
		return errors.New(message)
	}
	namespace, err := i.clusterImportOperations.GetNamespace(LocalClusterNamespace)
	if err != nil {
		message := fmt.Sprintf("hubClusterDay2Importer: Unable to find the %s namesapce due to error: %s", LocalClusterNamespace, err.Error())
		i.log.Error(message)
		return err
	}
	if namespace != nil {
		i.log.Infof("hubClusterDay2Importer: Found namespace %s", namespace.Name)
	}
	openshift_version := ""
	release_image := ""
	if clusterVersion != nil {
		openshift_version = clusterVersion.Status.History[0].Version
		release_image = clusterVersion.Status.History[0].Image
		i.log.Infof("hubClusterDay2Importer: Found openshift_version %s", openshift_version)
	} else {
		message := "hubClusterDay2Importer: Cluster version info is empty, cannot proceed"
		i.log.Error(message)
		return errors.New(message)
	}

	kubeConfigSecret, err := i.clusterImportOperations.GetSecret("openshift-kube-apiserver", "node-kubeconfigs")
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Unable to fetch local cluster kubeconfigs due to error %s", err.Error())
		return err
	}
	if kubeConfigSecret != nil {
		i.log.Infof("hubClusterDay2Importer: Found secret %s", kubeConfigSecret.Name)
	} else {
		message := "hubClusterDay2Importer: Could not find cluster kubeconfigs for local cluster"
		i.log.Error(message)
		return errors.New(message)
	}

	pullSecret, err := i.clusterImportOperations.GetSecret("openshift-machine-api", "pull-secret")
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Unable to fetch pull secret due to error %s", err.Error())
		return err
	}
	if pullSecret != nil {
		i.log.Infof("hubClusterDay2Importer: Found secret %s", pullSecret.Name)
	} else {
		message := "hubClusterDay2Importer: Could not find pull secret for local cluster"
		i.log.Error(message)
		return errors.New(message)
	}

	dns, err := i.clusterImportOperations.GetDNS("cluster")
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Could not fetch DNS due to error %s", err.Error())
		return err
	}

	numberOfControlPlaneNodes, err := i.clusterImportOperations.GetNumberOfControlPlaneNodes()
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Unable to determine the number of control plane nodes due to error %s", err.Error())
		return err
	}
	i.log.Infof("hubClusterDay2Importer: Number of control plane nodes is %d", numberOfControlPlaneNodes)

	clusterImageSet := hivev1.ClusterImageSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: "local-cluster-image-set",
		},
		Spec: hivev1.ClusterImageSetSpec{
			ReleaseImage: release_image,
		},
	}
	_, err = i.clusterImportOperations.CreateClusterImageSet(&clusterImageSet)
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: unable to create ClusterImageSet due to error: %s", err.Error())
		return err
	}

	// Store the kubeconfig data in the local cluster namespace
	localClusterSecret := v1.Secret{}
	localClusterSecret.Name = fmt.Sprintf("%s-admin-kubeconfig", LocalClusterNamespace)
	localClusterSecret.Namespace = LocalClusterNamespace
	localClusterSecret.Data = make(map[string][]byte)
	localClusterSecret.Data["kubeconfig"] = kubeConfigSecret.Data["lb-ext.kubeconfig"]
	kubeConfigSecret, err = i.clusterImportOperations.CreateSecret(LocalClusterNamespace, &localClusterSecret)
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: to store secret due to error %s", err.Error())
		return err
	}
	i.log.Infof("hubClusterDay2Importer: Created secret %s", localClusterSecret.Name)

	// Store the pull secret in the local cluster namespace
	hubPullSecret := v1.Secret{}
	hubPullSecret.Name = pullSecret.Name
	hubPullSecret.Namespace = LocalClusterNamespace
	hubPullSecret.Data = make(map[string][]byte)
	// .dockerconfigjson is double base64 encoded for some reason.
	// simply obtaining the secret above will perform one layer of decoding.
	// we need to manually perform another to ensure that the data is correctly copied to the new secret.
	hubPullSecret.Data[".dockerconfigjson"], err = base64.StdEncoding.DecodeString(string(pullSecret.Data[".dockerconfigjson"]))
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Unable to decode base64 pull secret data due to error %s", err.Error())
		return err
	}
	hubPullSecret.OwnerReferences = []metav1.OwnerReference{}
	_, err = i.clusterImportOperations.CreateSecret(LocalClusterNamespace, &hubPullSecret)
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Unable to store hub pull secret due to error %s", err.Error())
		return err
	}

	//Create an AgentClusterInstall in the local cluster namespace
	userManagedNetworkingActive := true
	agentClusterInstall := &hiveext.AgentClusterInstall{
		Spec: hiveext.AgentClusterInstallSpec{
			Networking: hiveext.Networking{
				UserManagedNetworking: &userManagedNetworkingActive,
			},
			ClusterDeploymentRef: v1.LocalObjectReference{
				Name: LocalClusterNamespace + "-cluster-deployment",
			},
			ImageSetRef: &hivev1.ClusterImageSetReference{
				Name: "local-cluster-image-set",
			},
			ProvisionRequirements: hiveext.ProvisionRequirements{
				ControlPlaneAgents: numberOfControlPlaneNodes,
			},
		},
	}
	agentClusterInstall.Namespace = LocalClusterNamespace
	agentClusterInstall.Name = LocalClusterNamespace + "-cluster-install"
	err = i.clusterImportOperations.CreateAgentClusterInstall(agentClusterInstall)
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Could not create AgentClusterInstall due to error %s", err.Error())
		return err
	}

	// Fetch the recently stored AgentClusterInstall so that we can obtain the UID
	aci, err := i.clusterImportOperations.GetAgentClusterInstall(LocalClusterNamespace,  LocalClusterNamespace + "-cluster-install")
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Failed to fetch created AgentClusterInstall due to error %s", err.Error())
		return err
	}

	// Create a cluster deployment in the local cluster namespace
	clusterDeployment := &hivev1.ClusterDeployment{
		Spec: hivev1.ClusterDeploymentSpec{
			Installed: true, // Let ZTP know to import this cluster
			ClusterMetadata: &hivev1.ClusterMetadata{
				ClusterID:                "",
				InfraID:                  "",
				AdminKubeconfigSecretRef: v1.LocalObjectReference{Name: kubeConfigSecret.Name},
			},
			ClusterInstallRef: &hivev1.ClusterInstallLocalReference{
				Name:    LocalClusterNamespace + "-cluster-install",
				Group:   "extensions.hive.openshift.io",
				Kind:    "AgentClusterInstall",
				Version: "v1beta1",
			},
			Platform: hivev1.Platform{
				AgentBareMetal: &agent.BareMetalPlatform{
					AgentSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"infraenv": fmt.Sprintf("%s-day2-infraenv", LocalClusterNamespace)},
					},
				},
			},
			PullSecretRef: &v1.LocalObjectReference{
				Name: pullSecret.Name,
			},
		},
	}
	clusterDeployment.Name = LocalClusterNamespace + "-cluster-deployment"
	clusterDeployment.Namespace = LocalClusterNamespace
	clusterDeployment.Spec.ClusterName = LocalClusterNamespace + "-cluster-deployment"
	clusterDeployment.Spec.BaseDomain = dns.Spec.BaseDomain
	//
	// Adding this ownership reference to ensure we can submit clusterDeployment without ManagedClusterSet/join permission
	//
	agentClusterInstallOwnerRef := metav1.OwnerReference{
		Kind:       "AgentClusterInstall",
		APIVersion: "extensions.hive.openshift.io/v1beta1",
		Name:       LocalClusterNamespace + "-cluster-install",
		UID:        aci.UID,
	}
	clusterDeployment.OwnerReferences = []metav1.OwnerReference{agentClusterInstallOwnerRef}
	_, err = i.clusterImportOperations.CreateClusterDeployment(clusterDeployment)
	if err != nil {
		i.log.Errorf("hubClusterDay2Importer: Could not create ClusterDeployment due to error %s", err.Error())
		return err
	}
	i.log.Info("hubClusterDay2Importer: Completed Day2 import of hub cluster")
	return nil
}
