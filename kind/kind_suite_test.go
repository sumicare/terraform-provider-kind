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
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// cleanupTestClusters removes all test clusters using kind binary.
func cleanupTestClusters() {
	// Use kind CLI to get list of clusters
	cmd := exec.CommandContext(context.Background(), "kind", "get", "clusters")

	output, err := cmd.Output()
	if err != nil {
		return
	}

	testPrefixes := []string{
		"tf-acc-cluster-test-",
		"tf-acc-config-base-test-",
		"tf-acc-config-nodes-test-",
		"tf-acc-containerd-test-",
	}

	clusters := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for clusterName := range clusters {
		if clusterName == "" {
			continue
		}

		for _, prefix := range testPrefixes {
			if !strings.HasPrefix(clusterName, prefix) {
				continue
			}

			deleteCmd := exec.CommandContext(context.Background(), "kind", "delete", "cluster", "--name", clusterName)

			err := deleteCmd.Run()
			if err != nil {
				Fail(fmt.Sprintf("Failed to delete cluster %s: %v", clusterName, err))
			}

			break
		}
	}
}

// SynchronizedBeforeSuite runs once before all parallel processes start.
var _ = SynchronizedBeforeSuite(func() []byte {
	if os.Getenv("TF_ACC") == "1" {
		cleanupTestClusters()
	}

	return nil
}, func(_ []byte) {})

// SynchronizedAfterSuite runs once after all parallel processes complete.
var _ = SynchronizedAfterSuite(func() {}, func() {
	if os.Getenv("TF_ACC") != "1" {
		return
	}

	cleanupTestClusters()

	// Clean up temporary config files
	configFiles, err := os.ReadDir(".")
	if err != nil {
		return
	}

	for _, entry := range configFiles {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, "tf-acc-") && strings.HasSuffix(name, "-config") {
			_ = os.Remove(name) // Ignore errors during cleanup
		}
	}
})

// TestKind runs the Ginkgo test suite for the Kind provider.
func TestKind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kind Provider Suite")
}
