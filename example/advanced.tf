# Advanced cluster configuration with networking, port mappings, and extra mounts
resource "kind_cluster" "advanced" {
  name           = "advanced-cluster"
  wait_for_ready = true
  node_image     = "kindest/node:v1.34.0"

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    # Control plane node with port mappings and extra mounts
    node {
      role = "control-plane"

      # Port mappings for ingress
      extra_port_mappings {
        container_port = 80
        host_port      = 80
        protocol       = "TCP"
      }

      extra_port_mappings {
        container_port = 443
        host_port      = 443
        protocol       = "TCP"
      }

      # Mount local directory into the node
      extra_mounts {
        host_path      = "/tmp/kind-data"
        container_path = "/data"
        read_only      = false
      }

      # Kubeadm config patches
      kubeadm_config_patches = [
        <<-EOT
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
        EOT
      ]
    }

    # Worker nodes
    node {
      role = "worker"

      extra_mounts {
        host_path      = "/tmp/kind-data"
        container_path = "/data"
      }
    }

    node {
      role = "worker"

      extra_mounts {
        host_path      = "/tmp/kind-data"
        container_path = "/data"
      }
    }

    # Networking configuration
    networking {
      api_server_address = "127.0.0.1"
      api_server_port    = 6443
      pod_subnet         = "10.244.0.0/16"
      service_subnet     = "10.96.0.0/12"
      disable_default_cni = false
      kube_proxy_mode    = "iptables"
    }

    # Containerd configuration patches
    containerd_config_patches = [
      <<-EOT
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
        endpoint = ["http://kind-registry:5000"]
      EOT
    ]

    # Feature gates
    feature_gates = {
      "EphemeralContainers" = "true"
      "CSINodeExpandSecret" = "true"
    }

    # Runtime config
    runtime_config = {
      "api_all" = "true"
    }
  }
}

output "advanced_cluster_endpoint" {
  value       = kind_cluster.advanced.endpoint
  description = "Advanced cluster API endpoint"
}

output "advanced_kubeconfig" {
  value       = kind_cluster.advanced.kubeconfig
  sensitive   = true
  description = "Kubeconfig for advanced cluster"
}
