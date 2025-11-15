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
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Test data constants for reusability across tests.
var (
	testClusterKind      = "Cluster"
	testAPIVersion       = "kind.x-k8s.io/v1alpha4"
	testControlPlaneRole = "control-plane"
	testWorkerRole       = "worker"
	testNodeImage        = "kindest/node:v1.29.0"
	testAPIServerAddress = "127.0.0.1"
	testAPIServerPort    = 6443
	testHostPath         = "/host/path"
	testContainerPath    = "/container/path"
	testContainerPort    = 80
	testHostPort         = 8080
	testListenAddress    = "0.0.0.0"
	testPodSubnet        = "10.244.0.0/16"
	testServiceSubnet    = "10.96.0.0/12"
)

// assertNoError checks that no error occurred and provides a descriptive message.
func assertNoError(err error, message string) {
	Expect(err).ToNot(HaveOccurred(), message)
}

// assertValidResult checks that result is not nil and provides a descriptive message.
func assertValidResult(result any, message string) {
	Expect(result).NotTo(BeNil(), message)
}

// Test suite for Kind configuration structure flattening functions.
var _ = Describe("StructureKindConfig", func() {
	DescribeTable("getString - extracts string values from maps",
		func(input map[string]any, key, expected string) {
			result := getString(input, key)
			Expect(result).To(Equal(expected), "getString should handle all input types correctly")
		},
		Entry("extracts existing string value", map[string]any{"key": "value"}, "key", "value"),
		Entry("returns empty for missing key", map[string]any{}, "missing", ""),
		Entry("returns empty for nil value", map[string]any{"key": nil}, "key", ""),
		Entry("returns empty for wrong type", map[string]any{"key": 123}, "key", ""),
	)

	DescribeTable("getInt - extracts integer values from maps",
		func(input map[string]any, key string, expected int) {
			result := getInt(input, key)
			Expect(result).To(Equal(expected), "getInt should handle all input types correctly")
		},
		Entry("extracts existing int value", map[string]any{"port": testAPIServerPort}, "port", testAPIServerPort),
		Entry("returns zero for missing key", map[string]any{}, "missing", 0),
		Entry("returns zero for nil value", map[string]any{"port": nil}, "port", 0),
		Entry("returns zero for wrong type", map[string]any{"port": "invalid"}, "port", 0),
	)

	DescribeTable("getBool - extracts boolean values from maps",
		func(input map[string]any, key string, expected bool) {
			result := getBool(input, key)
			Expect(result).To(Equal(expected), "getBool should handle all input types correctly")
		},
		Entry("extracts true value", map[string]any{"enabled": true}, "enabled", true),
		Entry("extracts false value", map[string]any{"enabled": false}, "enabled", false),
		Entry("returns false for missing key", map[string]any{}, "missing", false),
		Entry("returns false for nil value", map[string]any{"enabled": nil}, "enabled", false),
		Entry("returns false for wrong type", map[string]any{"enabled": "true"}, "enabled", false),
	)

	DescribeTable("flattenKindConfig - converts map to v1alpha4.Cluster",
		func(input map[string]any, validator func(*v1alpha4.Cluster)) {
			result, err := flattenKindConfig(input)
			assertNoError(err, "flattenKindConfig should not return an error")
			assertValidResult(result, "flattenKindConfig should return a non-nil result")
			validator(result)
		},
		Entry("basic cluster config",
			map[string]any{
				"kind":        testClusterKind,
				"api_version": testAPIVersion,
			},
			func(result *v1alpha4.Cluster) {
				Expect(result.Kind).To(Equal(testClusterKind), "Kind field should be set correctly")
				Expect(result.APIVersion).To(Equal(testAPIVersion), "APIVersion field should be set correctly")
			}),
		Entry("cluster config with nodes",
			map[string]any{
				"kind":        testClusterKind,
				"api_version": testAPIVersion,
				"node": []any{
					map[string]any{"role": testControlPlaneRole},
					map[string]any{"role": testWorkerRole},
				},
			},
			func(result *v1alpha4.Cluster) {
				Expect(result.Nodes).To(HaveLen(2), "should have 2 nodes")
				Expect(result.Nodes[0].Role).To(Equal(v1alpha4.ControlPlaneRole), "first node should be control-plane")
				Expect(result.Nodes[1].Role).To(Equal(v1alpha4.WorkerRole), "second node should be worker")
			}),
		Entry("cluster config with networking",
			map[string]any{
				"kind":        testClusterKind,
				"api_version": testAPIVersion,
				"networking": []any{
					map[string]any{
						"api_server_address": testAPIServerAddress,
						"api_server_port":    testAPIServerPort,
					},
				},
			},
			func(result *v1alpha4.Cluster) {
				Expect(result.Networking.APIServerAddress).To(Equal(testAPIServerAddress), "API server address should be set correctly")
				Expect(result.Networking.APIServerPort).To(Equal(int32(testAPIServerPort)), "API server port should be set correctly")
			}),
		Entry("cluster config with containerd patches",
			map[string]any{
				"kind":        testClusterKind,
				"api_version": testAPIVersion,
				"containerd_config_patches": []any{
					"[plugins.cri]\n  sandbox_image = \"test\"",
					"[plugins.cri.registry]\n  config_path = \"/etc/containerd/certs.d\"",
				},
			},
			func(result *v1alpha4.Cluster) {
				Expect(result.ContainerdConfigPatches).To(HaveLen(2), "should have 2 containerd config patches")
				Expect(result.ContainerdConfigPatches[0]).To(ContainSubstring("sandbox_image"), "first patch should contain sandbox_image")
				Expect(result.ContainerdConfigPatches[1]).To(ContainSubstring("config_path"), "second patch should contain config_path")
			}),
		Entry("cluster config with runtime config",
			map[string]any{
				"kind":        testClusterKind,
				"api_version": testAPIVersion,
				"runtime_config": map[string]any{
					"api_alpha": "false",
					"api_beta":  "true",
				},
			},
			func(result *v1alpha4.Cluster) {
				Expect(result.RuntimeConfig).To(HaveLen(2), "should have 2 runtime config entries")
				Expect(result.RuntimeConfig["api/alpha"]).To(Equal("false"), "api/alpha should be false")
				Expect(result.RuntimeConfig["api/beta"]).To(Equal("true"), "api/beta should be true")
			}),
		Entry("cluster config with feature gates",
			map[string]any{
				"kind":        testClusterKind,
				"api_version": testAPIVersion,
				"feature_gates": map[string]any{
					"FeatureA": "true",
					"FeatureB": "false",
					"FeatureC": "True",
				},
			},
			func(result *v1alpha4.Cluster) {
				Expect(result.FeatureGates).To(HaveLen(3), "should have 3 feature gates")
				Expect(result.FeatureGates["FeatureA"]).To(BeTrue(), "FeatureA should be true")
				Expect(result.FeatureGates["FeatureB"]).To(BeFalse(), "FeatureB should be false")
				Expect(result.FeatureGates["FeatureC"]).To(BeTrue(), "FeatureC should be true")
			}),
	)

	DescribeTable("flattenKindConfigNodes - converts map to v1alpha4.Node",
		func(input map[string]any, validator func(v1alpha4.Node)) {
			result, err := flattenKindConfigNodes(input)
			assertNoError(err, "flattenKindConfigNodes should not return an error")
			validator(result)
		},
		Entry("control-plane node",
			map[string]any{"role": testControlPlaneRole},
			func(result v1alpha4.Node) {
				Expect(result.Role).To(Equal(v1alpha4.ControlPlaneRole), "role should be control-plane")
			}),
		Entry("worker node",
			map[string]any{"role": testWorkerRole},
			func(result v1alpha4.Node) {
				Expect(result.Role).To(Equal(v1alpha4.WorkerRole), "role should be worker")
			}),
		Entry("node with custom image",
			map[string]any{
				"role":  testControlPlaneRole,
				"image": testNodeImage,
			},
			func(result v1alpha4.Node) {
				Expect(result.Image).To(Equal(testNodeImage), "image should be set correctly")
			}),
		Entry("node with custom labels",
			map[string]any{
				"role": testWorkerRole,
				"labels": map[string]any{
					"app":  "test",
					"tier": "backend",
				},
			},
			func(result v1alpha4.Node) {
				Expect(result.Labels).To(HaveLen(2), "should have 2 labels")
				Expect(result.Labels["app"]).To(Equal("test"), "app label should be test")
				Expect(result.Labels["tier"]).To(Equal("backend"), "tier label should be backend")
			}),
		Entry("node with extra mounts",
			map[string]any{
				"role": testControlPlaneRole,
				"extra_mounts": []any{
					map[string]any{
						"host_path":      testHostPath,
						"container_path": testContainerPath,
					},
				},
			},
			func(result v1alpha4.Node) {
				Expect(result.ExtraMounts).To(HaveLen(1), "should have 1 extra mount")
				Expect(result.ExtraMounts[0].HostPath).To(Equal(testHostPath), "host path should be set correctly")
				Expect(result.ExtraMounts[0].ContainerPath).To(Equal(testContainerPath), "container path should be set correctly")
			}),
		Entry("node with extra port mappings",
			map[string]any{
				"role": testControlPlaneRole,
				"extra_port_mappings": []any{
					map[string]any{
						"container_port": testContainerPort,
						"host_port":      testHostPort,
					},
				},
			},
			func(result v1alpha4.Node) {
				Expect(result.ExtraPortMappings).To(HaveLen(1), "should have 1 extra port mapping")
				Expect(result.ExtraPortMappings[0].ContainerPort).To(Equal(int32(testContainerPort)), "container port should be set correctly")
				Expect(result.ExtraPortMappings[0].HostPort).To(Equal(int32(testHostPort)), "host port should be set correctly")
			}),
		Entry("node with kubeadm config patches",
			map[string]any{
				"role": testControlPlaneRole,
				"kubeadm_config_patches": []any{
					"patch1",
					"patch2",
				},
			},
			func(result v1alpha4.Node) {
				Expect(result.KubeadmConfigPatches).To(HaveLen(2), "should have 2 kubeadm config patches")
				Expect(result.KubeadmConfigPatches[0]).To(Equal("patch1"), "first patch should be patch1")
				Expect(result.KubeadmConfigPatches[1]).To(Equal("patch2"), "second patch should be patch2")
			}),
	)

	DescribeTable("flattenKindConfigNetworking - converts map to v1alpha4.Networking",
		func(input map[string]any, validator func(v1alpha4.Networking)) {
			result, err := flattenKindConfigNetworking(input)
			assertNoError(err, "flattenKindConfigNetworking should not return an error")
			validator(result)
		},
		Entry("networking with API server settings",
			map[string]any{
				"api_server_address": testAPIServerAddress,
				"api_server_port":    testAPIServerPort,
			},
			func(result v1alpha4.Networking) {
				Expect(result.APIServerAddress).To(Equal(testAPIServerAddress), "API server address should be set correctly")
				Expect(result.APIServerPort).To(Equal(int32(testAPIServerPort)), "API server port should be set correctly")
			}),
		Entry("networking with IPv4 family",
			map[string]any{"ip_family": "ipv4"},
			func(result v1alpha4.Networking) {
				Expect(result.IPFamily).To(Equal(v1alpha4.IPv4Family), "IP family should be ipv4")
			}),
		Entry("networking with IPv6 family",
			map[string]any{"ip_family": "ipv6"},
			func(result v1alpha4.Networking) {
				Expect(result.IPFamily).To(Equal(v1alpha4.IPv6Family), "IP family should be ipv6")
			}),
		Entry("networking with dual stack family",
			map[string]any{"ip_family": "dual"},
			func(result v1alpha4.Networking) {
				Expect(result.IPFamily).To(Equal(v1alpha4.DualStackFamily), "IP family should be dual stack")
			}),
		Entry("networking with kube proxy mode",
			map[string]any{"kube_proxy_mode": "iptables"},
			func(result v1alpha4.Networking) {
				Expect(result.KubeProxyMode).To(Equal(v1alpha4.IPTablesProxyMode), "kube proxy mode should be iptables")
			}),
		Entry("networking with kube proxy disabled",
			map[string]any{"kube_proxy_mode": "none"},
			func(result v1alpha4.Networking) {
				Expect(result.KubeProxyMode).To(Equal(v1alpha4.ProxyMode("none")), "kube proxy mode should be none")
			}),
		Entry("networking with subnets",
			map[string]any{
				"pod_subnet":     testPodSubnet,
				"service_subnet": testServiceSubnet,
			},
			func(result v1alpha4.Networking) {
				Expect(result.PodSubnet).To(Equal(testPodSubnet), "pod subnet should be set correctly")
				Expect(result.ServiceSubnet).To(Equal(testServiceSubnet), "service subnet should be set correctly")
			}),
		Entry("networking with disable default CNI",
			map[string]any{"disable_default_cni": true},
			func(result v1alpha4.Networking) {
				Expect(result.DisableDefaultCNI).To(BeTrue(), "disable default CNI should be true")
			}),
		Entry("networking with DNS search",
			map[string]any{
				"dns_search": []any{"example.com", "test.local"},
			},
			func(result v1alpha4.Networking) {
				Expect(result.DNSSearch).NotTo(BeNil(), "DNS search should not be nil")
				Expect(*result.DNSSearch).To(HaveLen(2), "DNS search should have 2 entries")
				Expect((*result.DNSSearch)[0]).To(Equal("example.com"), "first DNS search entry should be example.com")
				Expect((*result.DNSSearch)[1]).To(Equal("test.local"), "second DNS search entry should be test.local")
			}),
	)

	DescribeTable("flattenKindConfigExtraMounts - converts map to v1alpha4.Mount",
		func(input map[string]any, validator func(v1alpha4.Mount)) {
			result := flattenKindConfigExtraMounts(input)
			validator(result)
		},
		Entry("basic mount",
			map[string]any{
				"host_path":      testHostPath,
				"container_path": testContainerPath,
			},
			func(result v1alpha4.Mount) {
				Expect(result.HostPath).To(Equal(testHostPath), "host path should be set correctly")
				Expect(result.ContainerPath).To(Equal(testContainerPath), "container path should be set correctly")
			}),
		Entry("mount with bidirectional propagation",
			map[string]any{
				"host_path":      testHostPath,
				"container_path": testContainerPath,
				"propagation":    "Bidirectional",
			},
			func(result v1alpha4.Mount) {
				Expect(result.Propagation).To(Equal(v1alpha4.MountPropagationBidirectional), "propagation should be bidirectional")
			}),
		Entry("mount with HostToContainer propagation",
			map[string]any{
				"host_path":      testHostPath,
				"container_path": testContainerPath,
				"propagation":    "HostToContainer",
			},
			func(result v1alpha4.Mount) {
				Expect(result.Propagation).To(Equal(v1alpha4.MountPropagationHostToContainer), "propagation should be host to container")
			}),
		Entry("mount with None propagation",
			map[string]any{
				"host_path":      testHostPath,
				"container_path": testContainerPath,
				"propagation":    "None",
			},
			func(result v1alpha4.Mount) {
				Expect(result.Propagation).To(Equal(v1alpha4.MountPropagationNone), "propagation should be none")
			}),
		Entry("mount with read only flag",
			map[string]any{
				"host_path":      testHostPath,
				"container_path": testContainerPath,
				"read_only":      true,
			},
			func(result v1alpha4.Mount) {
				Expect(result.Readonly).To(BeTrue(), "read only flag should be true")
			}),
		Entry("mount with selinux relabel",
			map[string]any{
				"host_path":       testHostPath,
				"container_path":  testContainerPath,
				"selinux_relabel": true,
			},
			func(result v1alpha4.Mount) {
				Expect(result.SelinuxRelabel).To(BeTrue(), "selinux relabel should be true")
			}),
	)

	DescribeTable("flattenKindConfigExtraPortMappings - converts map to v1alpha4.PortMapping",
		func(input map[string]any, validator func(v1alpha4.PortMapping)) {
			result, err := flattenKindConfigExtraPortMappings(input)
			assertNoError(err, "flattenKindConfigExtraPortMappings should not return an error")
			validator(result)
		},
		Entry("basic port mapping",
			map[string]any{
				"container_port": testContainerPort,
				"host_port":      testHostPort,
			},
			func(result v1alpha4.PortMapping) {
				Expect(result.ContainerPort).To(Equal(int32(testContainerPort)), "container port should be set correctly")
				Expect(result.HostPort).To(Equal(int32(testHostPort)), "host port should be set correctly")
			}),
		Entry("port mapping with listen address",
			map[string]any{
				"container_port": testContainerPort,
				"host_port":      testHostPort,
				"listen_address": testListenAddress,
			},
			func(result v1alpha4.PortMapping) {
				Expect(result.ListenAddress).To(Equal(testListenAddress), "listen address should be set correctly")
			}),
		Entry("port mapping with TCP protocol",
			map[string]any{
				"container_port": testContainerPort,
				"host_port":      testHostPort,
				"protocol":       "TCP",
			},
			func(result v1alpha4.PortMapping) {
				Expect(result.Protocol).To(Equal(v1alpha4.PortMappingProtocolTCP), "protocol should be TCP")
			}),
		Entry("port mapping with UDP protocol",
			map[string]any{
				"container_port": 53,
				"host_port":      5353,
				"protocol":       "UDP",
			},
			func(result v1alpha4.PortMapping) {
				Expect(result.Protocol).To(Equal(v1alpha4.PortMappingProtocolUDP), "protocol should be UDP")
			}),
		Entry("port mapping with SCTP protocol",
			map[string]any{
				"container_port": 9999,
				"host_port":      9999,
				"protocol":       "SCTP",
			},
			func(result v1alpha4.PortMapping) {
				Expect(result.Protocol).To(Equal(v1alpha4.PortMappingProtocolSCTP), "protocol should be SCTP")
			}),
	)
})
