package localclusterimport

import (
	"context"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	hiveext "github.com/openshift/assisted-service/api/hiveextension/v1beta1"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -source=local_cluster_import_operations.go -package=localclusterimport -destination=local_cluster_import_operations_mocks.go
type ClusterImportOperations interface {
	GetAgentClusterInstall(namespace string, name string) (*hiveext.AgentClusterInstall, error)
	CreateAgentClusterInstall(agentClusterInstall *hiveext.AgentClusterInstall) error
	GetNamespace(name string) (*v1.Namespace, error)
	GetSecret(namespace string, name string) (*v1.Secret, error)
	CreateSecret(namespace string, secret *v1.Secret) (*v1.Secret, error)
	GetClusterVersion(name string) (*configv1.ClusterVersion, error)
	GetClusterImageSet(name string) (*hivev1.ClusterImageSet, error)
	CreateClusterImageSet(clusterImageSet *hivev1.ClusterImageSet) (*hivev1.ClusterImageSet, error)
	CreateClusterDeployment(clusterDeployment *hivev1.ClusterDeployment) (*hivev1.ClusterDeployment, error)
	GetNodes() (*v1.NodeList, error)
	GetNumberOfControlPlaneNodes() (int, error)
	GetDNS(name string) (*configv1.DNS, error)

	DeleteSecret(namespace string, name string) error
	DeleteClusterImageSet(name string) error
	DeleteClusterDeployment(namespace string, name string) error
	DeleteAgentClusterInstall(namespace string, name string) error
}

type LocalClusterImportOperations struct {
	context         context.Context
	apiReader       client.Reader
	cachedApiClient client.Writer
	log             *logrus.Logger
}

func NewLocalClusterImportOperations(apiReader client.Reader, cachedApiClient client.Writer, log *logrus.Logger) LocalClusterImportOperations {
	return LocalClusterImportOperations{context: context.TODO(), apiReader: apiReader, cachedApiClient: cachedApiClient, log: log}
}

func (o *LocalClusterImportOperations) DeleteSecret(namespace string, name string) error {
	secret, err := o.GetSecret(namespace, name)
	if err != nil {
		o.log.Errorf("Could not find secret %s in namespace %s to delete it due to error %s", name, namespace, err)
		return err
	}
	err = o.cachedApiClient.Delete(o.context, secret)
	if err != nil {
		o.log.Errorf("Could not delete secret %s in namespace %s due to error %s", name, namespace, err)
		return err
	}
	return nil
}

func (o *LocalClusterImportOperations) DeleteClusterImageSet(name string) error {
	clusterImageSet, err := o.GetClusterImageSet(name)
	if err != nil {
		o.log.Infof("Could not find ClusterImageSet %s to delete it due to error %s", name, err)
		return err
	}
	err = o.cachedApiClient.Delete(o.context, clusterImageSet)
	if err != nil {
		o.log.Infof("Could not delete ClusterImageSet %s due to error %s", name, err)
		return err
	}
	return nil
}

func (o *LocalClusterImportOperations) DeleteClusterDeployment(namespace string, name string) error {
	clusterDeployment, err := o.GetClusterDeployment(namespace, name)
	if err != nil {
		o.log.Errorf("Could not find ClusterDeployment %s in namespace %s to delete it due to error %s", name, namespace, err)
		return err
	}
	err = o.cachedApiClient.Delete(o.context, clusterDeployment)
	if err != nil {
		o.log.Errorf("Could not delete ClusterDeployment %s in namespace %s due to error %s", name, namespace, err)
		return err
	}
	return nil
}

func (o *LocalClusterImportOperations) DeleteAgentClusterInstall(namespace string, name string) error {
	agentClusterInstall, err := o.GetAgentClusterInstall(namespace, name)
	if err != nil {
		o.log.Errorf("Could not find AgentClusterInstall %s in namespace %s to delete it due to error %s", name, namespace, err)
		return err
	}
	err = o.cachedApiClient.Delete(o.context, agentClusterInstall)
	if err != nil {
		o.log.Errorf("Could not delete AgentClusterInstall %s in namespace %s due to error %s", name, namespace, err)
		return err
	}
	return nil
}

func (o *LocalClusterImportOperations) GetClusterDeployment(namespace string, name string) (*hivev1.ClusterDeployment, error) {
	clusterDeployment := &hivev1.ClusterDeployment{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := o.apiReader.Get(o.context, namespacedName, clusterDeployment)
	if err != nil {
		return nil, err
	}
	return clusterDeployment, nil
}

func (o *LocalClusterImportOperations) GetAgentClusterInstall(namespace string, name string) (*hiveext.AgentClusterInstall, error) {
	agentClusterInstall := &hiveext.AgentClusterInstall{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := o.apiReader.Get(o.context, namespacedName, agentClusterInstall)
	if err != nil {
		return nil, err
	}
	return agentClusterInstall, nil
}

func (o *LocalClusterImportOperations) CreateAgentClusterInstall(agentClusterInstall *hiveext.AgentClusterInstall) error {
	return o.cachedApiClient.Create(o.context, agentClusterInstall)
}

func (o *LocalClusterImportOperations) GetNamespace(name string) (*v1.Namespace, error) {
	nsList := &v1.NamespaceList{}
	err := o.apiReader.List(o.context, nsList)
	if err != nil {
		return nil, err
	}
	for _, ns := range nsList.Items {
		if ns.Name == name {
			return &ns, nil
		}
	}
	return nil, fmt.Errorf("unable to find namespace '%s'", name)
}

func (o *LocalClusterImportOperations) GetSecret(namespace string, name string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	err := o.apiReader.Get(o.context, namespacedName, secret)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to fetch secret %s from namespace %s", name, namespace)
	}
	return secret, nil
}

func (o *LocalClusterImportOperations) CreateSecret(namespace string, secret *v1.Secret) (*v1.Secret, error) {
	err := o.cachedApiClient.Create(o.context, secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (o *LocalClusterImportOperations) GetClusterVersion(name string) (*configv1.ClusterVersion, error) {
	clusterVersionList := &configv1.ClusterVersionList{}
	err := o.apiReader.List(o.context, clusterVersionList)
	if err != nil {
		return nil, err
	}
	for _, cv := range clusterVersionList.Items {
		if cv.Name == name {
			return &cv, nil
		}
	}
	return nil, fmt.Errorf("unable to find cluster version '%s'", name)
}

func (o *LocalClusterImportOperations) GetClusterImageSet(name string) (*hivev1.ClusterImageSet, error) {
	clusterImageSet := &hivev1.ClusterImageSet{}
	namespacedName := types.NamespacedName{
		Namespace: "",
		Name:      name,
	}
	err := o.apiReader.Get(o.context, namespacedName, clusterImageSet)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to fetch cluster image set %s", name)
	}
	return clusterImageSet, nil
}

func (o *LocalClusterImportOperations) CreateClusterImageSet(clusterImageSet *hivev1.ClusterImageSet) (*hivev1.ClusterImageSet, error) {
	err := o.cachedApiClient.Create(o.context, clusterImageSet)
	if err != nil {
		return nil, err
	}
	return clusterImageSet, nil
}

func (o *LocalClusterImportOperations) CreateClusterDeployment(clusterDeployment *hivev1.ClusterDeployment) (*hivev1.ClusterDeployment, error) {
	err := o.cachedApiClient.Create(o.context, clusterDeployment)
	if err != nil {
		return nil, err
	}
	return clusterDeployment, nil
}

func (o *LocalClusterImportOperations) GetNodes() (*v1.NodeList, error) {
	nodeList := &v1.NodeList{}
	err := o.apiReader.List(o.context, nodeList)
	if err != nil {
		return nil, err
	}
	return nodeList, nil
}

func (o *LocalClusterImportOperations) GetNumberOfControlPlaneNodes() (int, error) {
	// Determine the number of control plane agents we have
	// Control plane nodes have a specific label that we can look for.,
	numberOfControlPlaneNodes := 0
	nodeList, err := o.GetNodes()
	o.log.Infof("Number of nodes in nodeList %d", len(nodeList.Items))
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

func (o *LocalClusterImportOperations) GetDNS(name string) (*configv1.DNS, error) {
	dnsList := &configv1.DNSList{}
	err := o.apiReader.List(o.context, dnsList)
	if err != nil {
		return nil, err
	}
	for _, cv := range dnsList.Items {
		if cv.Name == name {
			return &cv, nil
		}
	}
	return nil, fmt.Errorf("unable to find dns record '%s'", name)
}
