## OpenTofu Provider Kind

Compared to original [terraform-provider-kind](https://github.com/tehcyx/terraform-provider-kind/), 
this provider terraform plugin SDK, and more rigorous testing approach.

## Features

- ✅ Modern Terraform Plugin Framework (not legacy SDKv2)
- ✅ Full support for Kind v1alpha4 cluster configuration
- ✅ Comprehensive test coverage with Ginkgo/Gomega
- ✅ Timeout handling for cluster operations
- ✅ Support for multi-node and HA clusters
- ✅ IPv6 and dual-stack networking
- ✅ Port mappings and volume mounts
- ✅ Kubeadm and containerd configuration patches

## Quick Start

```hcl
terraform {
  required_providers {
    kind = {
      source  = "sumicare/kind"
      version = "~> 1.1.0"
    }
  }
}

provider "kind" {}

resource "kind_cluster" "default" {
  name           = "test-cluster"
  wait_for_ready = true

  kind_config {
    kind = "Cluster"
    # api_version defaults to "kind.x-k8s.io/v1alpha4" if not specified

    node {
      role = "control-plane"
    }

    node {
      role = "worker"
    }
  }
}
```

## Examples

See the [example/](./example/) directory for comprehensive examples including:

- **Basic cluster** - Simple single control-plane cluster
- **Advanced configuration** - Port mappings, mounts, patches, networking
- **Multi-node HA** - High-availability cluster with multiple control planes
- **IPv6/Dual-stack** - IPv6-only and dual-stack networking examples

## Development

### Building

```bash
go build
```

### Local Development with OpenTofu

For local development and testing, use the `.tofurc` configuration to override the provider with your local build:

1. **Build the provider**:
   ```bash
   go build -o terraform-provider-kind
   ```

2. **Set up the dev override**:
   ```bash
   # Copy the example .tofurc to your home directory or use TF_CLI_CONFIG_FILE
   cp example/.tofurc ~/.tofurc
   
   # Or set the config file path
   export TF_CLI_CONFIG_FILE=/path/to/example/.tofurc
   ```

3. **Update the path in `.tofurc`** to point to your local build directory:
   ```hcl
   provider_installation {
     dev_overrides {
       "sumicare/kind" = "/path/to/opentofu-provider-kind"
     }
     direct {}
   }
   ```

4. **Run examples**:
   ```bash
   cd example
   tofu init
   tofu plan
   tofu apply
   ```

**Note**: When using `dev_overrides`, OpenTofu will skip provider version checks and use your local binary directly.

### Testing

```bash
# Unit tests
yarn test

# Integration tests (requires Docker)
yarn test:integration
```

### License

Sumicare OpenTofu Provider Kind is licensed under the terms of [Apache License 2.0](LICENSE) as the original [tehcyx/terraform-provider-kind](https://github.com/tehcyx/terraform-provider-kind)
