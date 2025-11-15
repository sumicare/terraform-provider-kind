/*
   Copyright 2025 Sumicare

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

package kind

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ErrPortOutOfRange is returned when a port value is outside the valid int32 range.
//
//nolint:grouper // false positive
var ErrPortOutOfRange = errors.New("port value out of valid range")

// flattenKindConfig converts a map representation of kind configuration to v1alpha4.Cluster.
// This function processes the configuration data and returns a structured cluster configuration.
func flattenKindConfig(kindConfig map[string]any) (*v1alpha4.Cluster, error) {
	obj := &v1alpha4.Cluster{}

	// Extract basic cluster metadata.
	obj.Kind = getString(kindConfig, "kind")
	obj.APIVersion = getString(kindConfig, "api_version")

	// Process each node configuration and append to the cluster.
	for _, nodeMap := range getMapSlice(kindConfig, "node") {
		node, err := flattenKindConfigNodes(nodeMap)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten node configuration: %w", err)
		}

		obj.Nodes = append(obj.Nodes, node)
	}

	// Process networking configuration if present.
	if networkingSlice := getMapSlice(kindConfig, "networking"); len(networkingSlice) > 0 {
		networking, err := flattenKindConfigNetworking(networkingSlice[0])
		if err != nil {
			return nil, fmt.Errorf("failed to flatten networking configuration: %w", err)
		}

		obj.Networking = networking
	}

	// Extract containerd configuration patches.
	obj.ContainerdConfigPatches = getStringSlice(kindConfig, "containerd_config_patches")

	// Process runtime configuration and normalize keys.
	if runtimeConfig := getStringMap(kindConfig, "runtime_config"); runtimeConfig != nil {
		obj.RuntimeConfig = make(map[string]string, len(runtimeConfig))
		for k, v := range runtimeConfig {
			// Replace underscore with slash (e.g., api_alpha -> api/alpha)
			obj.RuntimeConfig[strings.ReplaceAll(k, "_", "/")] = v
		}
	}

	// Process feature gates and convert string values to boolean.
	if featureGates := getStringMap(kindConfig, "feature_gates"); featureGates != nil {
		obj.FeatureGates = make(map[string]bool, len(featureGates))
		for k, v := range featureGates {
			obj.FeatureGates[k] = strings.EqualFold(v, "true")
		}
	}

	return obj, nil
}

// flattenKindConfigNodes converts a map representation of node configuration to v1alpha4.Node.
func flattenKindConfigNodes(nodeConfig map[string]any) (v1alpha4.Node, error) {
	obj := v1alpha4.Node{}

	// Determine and set the node role (control-plane or worker).
	if role := getString(nodeConfig, "role"); role != "" {
		switch role {
		case string(v1alpha4.ControlPlaneRole):
			obj.Role = v1alpha4.ControlPlaneRole
		case string(v1alpha4.WorkerRole):
			obj.Role = v1alpha4.WorkerRole
		}
	}

	// Set custom node image if specified.
	if image := getString(nodeConfig, "image"); image != "" {
		obj.Image = image
	}

	// Process extra mounts for the node.
	for _, mountMap := range getMapSlice(nodeConfig, "extra_mounts") {
		obj.ExtraMounts = append(obj.ExtraMounts, flattenKindConfigExtraMounts(mountMap))
	}

	// Apply custom labels to the node.
	if labels := getStringMap(nodeConfig, "labels"); labels != nil {
		obj.Labels = labels
	}

	// Process extra port mappings for the node.
	for _, portMap := range getMapSlice(nodeConfig, "extra_port_mappings") {
		portMapping, err := flattenKindConfigExtraPortMappings(portMap)
		if err != nil {
			return obj, fmt.Errorf("failed to flatten port mapping configuration: %w", err)
		}

		obj.ExtraPortMappings = append(obj.ExtraPortMappings, portMapping)
	}

	// Extract kubeadm configuration patches.
	obj.KubeadmConfigPatches = getStringSlice(nodeConfig, "kubeadm_config_patches")

	return obj, nil
}

// flattenKindConfigNetworking converts a map representation of networking configuration to v1alpha4.Networking.
func flattenKindConfigNetworking(networkingConfig map[string]any) (v1alpha4.Networking, error) {
	// Initialize networking configuration with basic settings.
	obj := v1alpha4.Networking{
		APIServerAddress:  getString(networkingConfig, "api_server_address"),
		DisableDefaultCNI: getBool(networkingConfig, "disable_default_cni"),
	}

	// Validate and set API server port within int32 range.
	if port := getInt(networkingConfig, "api_server_port"); port != 0 {
		if port < math.MinInt32 || port > math.MaxInt32 {
			return obj, fmt.Errorf("api_server_port value %d (must be between %d and %d): %w", port, math.MinInt32, math.MaxInt32, ErrPortOutOfRange)
		}

		obj.APIServerPort = int32(port) // #nosec G115 -- validated range check
	}

	// Configure IP family (IPv4, IPv6, or dual-stack).
	if ipFamily := getString(networkingConfig, "ip_family"); ipFamily != "" {
		switch ipFamily {
		case string(v1alpha4.IPv4Family):
			obj.IPFamily = v1alpha4.IPv4Family
		case string(v1alpha4.IPv6Family):
			obj.IPFamily = v1alpha4.IPv6Family
		case string(v1alpha4.DualStackFamily):
			obj.IPFamily = v1alpha4.DualStackFamily
		}
	}

	// Configure kube-proxy mode (iptables, ipvs, or none).
	if kubeProxyMode := getString(networkingConfig, "kube_proxy_mode"); kubeProxyMode != "" {
		switch kubeProxyMode {
		case string(v1alpha4.IPTablesProxyMode):
			obj.KubeProxyMode = v1alpha4.IPTablesProxyMode
		case string(v1alpha4.IPVSProxyMode):
			obj.KubeProxyMode = v1alpha4.IPVSProxyMode
		case kubeProxyModeNone:
			obj.KubeProxyMode = kubeProxyModeNone
		}
	}

	// Set pod and service subnet CIDR ranges.
	obj.PodSubnet = getString(networkingConfig, "pod_subnet")
	obj.ServiceSubnet = getString(networkingConfig, "service_subnet")

	// Configure DNS search domains if specified.
	if dnsSearch := getStringSlice(networkingConfig, "dns_search"); dnsSearch != nil {
		obj.DNSSearch = &dnsSearch
	}

	return obj, nil
}

// flattenKindConfigExtraMounts converts a map representation of mount configuration to v1alpha4.Mount.
func flattenKindConfigExtraMounts(mountConfig map[string]any) v1alpha4.Mount {
	// Initialize mount configuration with basic settings.
	obj := v1alpha4.Mount{
		ContainerPath:  getString(mountConfig, "container_path"),
		HostPath:       getString(mountConfig, "host_path"),
		Readonly:       getBool(mountConfig, "read_only"),
		SelinuxRelabel: getBool(mountConfig, "selinux_relabel"),
	}

	// Configure mount propagation mode if specified.
	if propagation := getString(mountConfig, "propagation"); propagation != "" {
		switch propagation {
		case string(v1alpha4.MountPropagationBidirectional):
			obj.Propagation = v1alpha4.MountPropagationBidirectional
		case string(v1alpha4.MountPropagationHostToContainer):
			obj.Propagation = v1alpha4.MountPropagationHostToContainer
		case string(v1alpha4.MountPropagationNone):
			obj.Propagation = v1alpha4.MountPropagationNone
		}
	}

	return obj
}

// flattenKindConfigExtraPortMappings converts a map representation of port mapping configuration to v1alpha4.PortMapping.
func flattenKindConfigExtraPortMappings(portMappingConfig map[string]any) (v1alpha4.PortMapping, error) {
	// Initialize port mapping configuration.
	obj := v1alpha4.PortMapping{
		ListenAddress: getString(portMappingConfig, "listen_address"),
	}

	// Validate and set container port within int32 range.
	if containerPort := getInt(portMappingConfig, "container_port"); containerPort != 0 {
		if containerPort < math.MinInt32 || containerPort > math.MaxInt32 {
			return obj, fmt.Errorf("container_port value %d (must be between %d and %d): %w", containerPort, math.MinInt32, math.MaxInt32, ErrPortOutOfRange)
		}

		obj.ContainerPort = int32(containerPort) // #nosec G115 -- validated range check
	}

	// Validate and set host port within int32 range.
	if hostPort := getInt(portMappingConfig, "host_port"); hostPort != 0 {
		if hostPort < math.MinInt32 || hostPort > math.MaxInt32 {
			return obj, fmt.Errorf("host_port value %d (must be between %d and %d): %w", hostPort, math.MinInt32, math.MaxInt32, ErrPortOutOfRange)
		}

		obj.HostPort = int32(hostPort) // #nosec G115 -- validated range check
	}

	// Configure port protocol (TCP, UDP, or SCTP).
	if protocol := getString(portMappingConfig, "protocol"); protocol != "" {
		switch protocol {
		case string(v1alpha4.PortMappingProtocolSCTP):
			obj.Protocol = v1alpha4.PortMappingProtocolSCTP
		case string(v1alpha4.PortMappingProtocolTCP):
			obj.Protocol = v1alpha4.PortMappingProtocolTCP
		case string(v1alpha4.PortMappingProtocolUDP):
			obj.Protocol = v1alpha4.PortMappingProtocolUDP
		}
	}

	return obj, nil
}
