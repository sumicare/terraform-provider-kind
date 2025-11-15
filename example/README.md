# OpenTofu Provider Kind - Examples

This directory contains example configurations for the OpenTofu Kind provider.

## Examples

- **`main.tf`** - Basic cluster with default settings
- **`advanced.tf`** - Advanced configuration with port mappings, mounts, patches, and networking
- **`multi-node.tf`** - High-availability cluster with multiple control planes
- **`ipv6.tf`** - IPv6-only and dual-stack networking examples

## Using Examples for Development

### With Local Provider Build

For local development, use the `.tofurc` file to override the provider with your local build:

1. **Build the provider** (from the root directory):
   ```bash
   cd ..
   go build -o terraform-provider-kind
   cd example
   ```

2. **Configure OpenTofu to use the local build**:
   ```bash
   # Option 1: Copy .tofurc to your home directory
   cp .tofurc ~/.tofurc
   
   # Option 2: Use TF_CLI_CONFIG_FILE environment variable
   export TF_CLI_CONFIG_FILE=$(pwd)/.tofurc
   ```

3. **Update the path in `.tofurc`** if needed to match your local directory structure.

4. **Run the examples**:
   ```bash
   tofu init
   tofu plan
   tofu apply
   ```

### With Published Provider

If you want to use the published provider instead of a local build:

1. **Remove or rename `.tofurc`**:
   ```bash
   mv .tofurc .tofurc.dev
   unset TF_CLI_CONFIG_FILE
   ```

2. **Run the examples**:
   ```bash
   tofu init
   tofu plan
   tofu apply
   ```

## Cleanup

To destroy the created clusters:

```bash
tofu destroy
```

Or manually delete Kind clusters:

```bash
kind delete cluster --name test-cluster
kind delete cluster --name advanced-cluster
kind delete cluster --name ha-cluster
kind delete cluster --name ipv6-cluster
```
