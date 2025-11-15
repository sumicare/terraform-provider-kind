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
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/apis/config/defaults"
	"sigs.k8s.io/kind/pkg/cluster"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testResourceName is the Terraform resource name used in acceptance tests.
const testResourceName = "kind_cluster.test"

var _ = Describe("Kind Cluster Resource", func() {
	var (
		resourceName string
		clusterName  string
	)

	// BeforeEach is used to set up test variables before each test.
	BeforeEach(func() {
		resourceName = testResourceName
		clusterName = acctest.RandomWithPrefix("tf-acc-cluster-test")
	})

	It("creates cluster with all basic configurations", func() {
		resource.Test(GinkgoT(), resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckKindClusterResourceDestroy(clusterName),
			Steps: []resource.TestStep{
				{
					Config: renderClusterConfig(ClusterConfig{
						Name:           clusterName,
						NodeImage:      defaults.Image,
						WaitForReady:   true,
						KubeconfigPath: "/tmp/kind-provider-test/new_file",
					}),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClusterCreate(resourceName),
						checkResourceAttr(resourceName, "name", clusterName),
						checkResourceAttr(resourceName, "node_image", defaults.Image),
						checkResourceAttr(resourceName, "wait_for_ready", "true"),
						checkResourceAttr(resourceName, "kubeconfig_path", "/tmp/kind-provider-test/new_file"),
					),
				},
			},
		})
	})
})

var _ = Describe("Cluster Config Base Tests", func() {
	var (
		resourceName string
		clusterName  string
	)

	// BeforeEach is used to set up test variables before each test.
	BeforeEach(func() {
		resourceName = testResourceName
		clusterName = acctest.RandomWithPrefix("tf-acc-config-base-test")
	})

	It("creates cluster with kind_config and all options", func() {
		resource.Test(GinkgoT(), resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckKindClusterResourceDestroy(clusterName),
			Steps: []resource.TestStep{
				{
					Config: renderClusterConfig(ClusterConfig{
						Name:         clusterName,
						NodeImage:    defaults.Image,
						WaitForReady: true,
						KindConfig: &KindConfig{
							Networking: &Networking{
								APIServerAddress: "127.0.0.1",
								APIServerPort:    6443,
								KubeProxyMode:    "none",
							},
							RuntimeConfig: map[string]string{"api_alpha": "false"},
						},
					}),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClusterCreate(resourceName),
						checkResourceAttr(resourceName, "kind_config.#", "1"),
						checkResourceAttr(resourceName, "kind_config.0.kind", "Cluster"),
						checkResourceAttr(resourceName, "kind_config.0.api_version", "kind.x-k8s.io/v1alpha4"),
						checkResourceAttr(resourceName, "wait_for_ready", "true"),
						checkResourceAttr(resourceName, "node_image", defaults.Image),
						checkResourceAttr(resourceName, "kind_config.0.networking.api_server_address", "127.0.0.1"),
						checkResourceAttr(resourceName, "kind_config.0.networking.api_server_port", "6443"),
						checkResourceAttr(resourceName, "kind_config.0.networking.kube_proxy_mode", "none"),
						checkResourceAttr(resourceName, "kind_config.0.runtime_config.%", "1"),
						checkResourceAttr(resourceName, "kind_config.0.runtime_config.api_alpha", "false"),
					),
				},
			},
		})
	})
})

var _ = Describe("Cluster Config Nodes Tests", func() {
	var (
		resourceName string
		clusterName  string
	)

	BeforeEach(func() {
		resourceName = testResourceName
		clusterName = acctest.RandomWithPrefix("tf-acc-config-nodes-test")
	})

	It("creates cluster with multi-node configuration", func() {
		resource.Test(GinkgoT(), resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckKindClusterResourceDestroy(clusterName),
			Steps: []resource.TestStep{
				{
					Config: renderClusterConfig(ClusterConfig{
						Name:         clusterName,
						NodeImage:    defaults.Image,
						WaitForReady: true,
						KindConfig: &KindConfig{
							Nodes: []Node{
								{Role: "control-plane", Labels: map[string]string{"name": "node0"}},
								{Role: "worker", Image: defaultNodeImage},
								{Role: "worker"},
							},
						},
					}),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClusterCreate(resourceName),
						checkResourceAttr(resourceName, "kind_config.0.node.#", "3"),
						checkResourceAttr(resourceName, "kind_config.0.node.0.role", "control-plane"),
						checkResourceAttr(resourceName, "kind_config.0.node.0.labels.name", "node0"),
						checkResourceAttr(resourceName, "kind_config.0.node.1.role", "worker"),
						checkResourceAttr(resourceName, "kind_config.0.node.1.image", defaultNodeImage),
						checkResourceAttr(resourceName, "kind_config.0.node.2.role", "worker"),
						checkResourceAttr(resourceName, "wait_for_ready", "true"),
						checkResourceAttr(resourceName, "node_image", defaults.Image),
					),
				},
			},
		})
	})
})

var _ = Describe("Cluster Containerd Patches Tests", func() {
	var (
		resourceName string
		clusterName  string
	)

	BeforeEach(func() {
		resourceName = testResourceName
		clusterName = acctest.RandomWithPrefix("tf-acc-containerd-test")
	})

	It("creates cluster with containerd config patches", func() {
		patch := `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
  endpoint = ["http://kind-registry:5000"]`

		resource.Test(GinkgoT(), resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckKindClusterResourceDestroy(clusterName),
			Steps: []resource.TestStep{
				{
					Config: renderClusterConfig(ClusterConfig{
						Name:         clusterName,
						WaitForReady: true,
						KindConfig: &KindConfig{
							ContainerdConfigPatches: []string{patch},
						},
					}),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckClusterCreate(resourceName),
						checkResourceAttr(resourceName, "kind_config.0.containerd_config_patches.#", "1"),
					),
				},
			},
		})
	})
})

// testAccCheckKindClusterResourceDestroy verifies the kind cluster
// has been destroyed.
func testAccCheckKindClusterResourceDestroy(clusterName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		prov := cluster.NewProvider()

		list, err := prov.List()
		if err != nil {
			return fmt.Errorf("failed to list clusters: %w", err)
		}

		Expect(slices.Contains(list, clusterName)).To(BeFalse(), "cluster %s should have been removed", clusterName)

		// Verify kubeconfig context has been removed
		contextName := "kind-" + clusterName
		configAccess := clientcmd.NewDefaultPathOptions()

		config, err := configAccess.GetStartingConfig()
		if err == nil {
			_, contextExists := config.Contexts[contextName]
			Expect(contextExists).To(BeFalse(), "kubeconfig context %s should have been removed", contextName)

			_, authInfoExists := config.AuthInfos[contextName]
			Expect(authInfoExists).To(BeFalse(), "kubeconfig user %s should have been removed", contextName)

			_, clusterExists := config.Clusters[contextName]
			Expect(clusterExists).To(BeFalse(), "kubeconfig cluster %s should have been removed", contextName)
		}

		return nil
	}
}

// testAccCheckClusterCreate verifies that a cluster resource exists in the state.
func testAccCheckClusterCreate(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		Expect(ok).To(BeTrue(), "root module should have resource %s", name)

		return nil
	}
}

// checkResourceAttr verifies that a resource attribute has the expected value.
func checkResourceAttr(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		Expect(ok).To(BeTrue(), "resource %s should exist", name)
		Expect(rs.Primary.Attributes).To(HaveKeyWithValue(key, value), "attribute %s should equal %s", key, value)

		return nil
	}
}

var _ = Describe("Cluster Resource Unit Tests", func() {
	Describe("NewClusterResource", func() {
		It("creates a new cluster resource", func() {
			resource := NewClusterResource()
			Expect(resource).NotTo(BeNil(), "NewClusterResource should return a non-nil resource")
		})
	})

	Describe("ClusterResource Metadata", func() {
		It("has correct type name", func() {
			resource := &ClusterResource{}
			Expect(resource).NotTo(BeNil(), "ClusterResource should be instantiable")
		})
	})
})
