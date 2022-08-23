# Agent-based installer

The assisted-service repo contains a client used by the agent-based installer. 
The client code is located at [../cmd/agentbasedinstaller](../cmd/agentbasedinstaller/).

The client's purpose is to register a cluster and infra-env for cluster0. 
The ISO generated by the agent-based installer contains a service which
starts this client and feeds it ZTP manifests which are used to specify 
the cluster that will be installed.

The ISO also contains all components needed to run assisted-service
and the agent that performs host registration and validation. The image 
is intend to be used by users to who are not currently supported by the 
IPI installer. It is also targeted for disconnected on-premises use.

The client performs three main tasks:
1. Reads in ZTP manifests containing information about cluster0. The manifests
   are either provided by users or generated by the ISO generation tool using
   source files like install-config.yaml. The manifests are bind mounted
   to the container running the client.
2. Registers a cluster in assisted service using the REST-API.
3. Registers a infraenv using the REST-API. The infraenv contains a
   reference to the created cluster and contains any nmstate configs.

A kube-api is not available in the agent-based installer environment. This
is why the ZTP manifests are not applied directly to kubernetes. Instead,
the manifests are translated into create cluster and infraenv params, and
then posted to the REST-API to create those resources.

This client is needed to register the cluster and infraenv to minimize
user interaction and to enable automation use cases.

A service running on the image, waits for the infraenv to be registered
to assisted-service. Once registered, an infraenv id is available and
the agent is started. The agent registers the host to assisted service
allowing the installation to continue.

A separate service waits for host validation to complete and initiates
the cluster installation when the cluster status becomes ready.

These two services are embedded in the image generated by the agent-based
installer tooling and are not part of this client.