/*
Copyright 2017 The Kubernetes Authors.

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

package upgrade

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	fakediscovery "k8s.io/client-go/discovery/fake"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmapiv1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/validation"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	"k8s.io/kubernetes/cmd/kubeadm/app/phases/upgrade"
	"k8s.io/kubernetes/cmd/kubeadm/app/preflight"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/apiclient"
	configutil "k8s.io/kubernetes/cmd/kubeadm/app/util/config"
	dryrunutil "k8s.io/kubernetes/cmd/kubeadm/app/util/dryrun"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/output"
)

// isKubeadmConfigPresent checks if a kubeadm config type is found in the provided document map
func isKubeadmConfigPresent(docmap kubeadmapi.DocumentMap) bool {
	for gvk := range docmap {
		if gvk.Group == kubeadmapi.GroupName {
			return true
		}
	}
	return false
}

// enforceRequirements verifies that it's okay to upgrade and then returns the variables needed for the rest of the procedure
func enforceRequirements(flags *applyPlanFlags, args []string, dryRun bool, upgradeApply bool, printer output.Printer) (clientset.Interface, upgrade.VersionGetter, *kubeadmapi.InitConfiguration, error) {
	client, err := getClient(flags.kubeConfigPath, dryRun)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "couldn't create a Kubernetes client from file %q", flags.kubeConfigPath)
	}

	// Fetch the configuration from a file or ConfigMap and validate it
	printer.Printf("[upgrade/config] Making sure the configuration is correct:\n")

	var newK8sVersion string
	_, err = configutil.LoadUpgradeConfig(flags.cfgPath)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "[upgrade/upgrade config] FATAL")
	}

	cfg, err := configutil.FetchInitConfigurationFromCluster(client, printer, "upgrade/config", false, true)
	if err != nil {
		if apierrors.IsNotFound(err) {
			printer.Printf("[upgrade/config] In order to upgrade, a ConfigMap called %q in the %s namespace must exist.\n", constants.KubeadmConfigConfigMap, metav1.NamespaceSystem)
			printer.Printf("[upgrade/config] Without this information, 'kubeadm upgrade' won't know how to configure your upgraded cluster.\n")
			printer.Println()
			printer.Printf("[upgrade/config] Next steps:\n")
			printer.Printf("\t- OPTION 1: Run 'kubeadm config upload from-flags' and specify the same CLI arguments you passed to 'kubeadm init' when you created your control-plane.\n")
			printer.Printf("\t- OPTION 2: Run 'kubeadm config upload from-file' and specify the same config file you passed to 'kubeadm init' when you created your control-plane.\n")
			printer.Println()
			err = errors.Errorf("the ConfigMap %q in the %s namespace used for getting configuration information was not found", constants.KubeadmConfigConfigMap, metav1.NamespaceSystem)
		}
		return nil, nil, nil, errors.Wrap(err, "[upgrade/init config] FATAL")
	}

	// The version arg is mandatory, during upgrade apply, unless it's specified in the config file
	if upgradeApply && newK8sVersion == "" {
		if err := cmdutil.ValidateExactArgNumber(args, []string{"version"}); err != nil {
			return nil, nil, nil, err
		}
	}

	// If option was specified in both args and config file, args will overwrite the config file.
	if len(args) == 1 {
		newK8sVersion = args[0]
		if upgradeApply {
			// The `upgrade apply` version always overwrites the KubernetesVersion in the returned cfg with the target
			// version. While this is not the same for `upgrade plan` where the KubernetesVersion should be the old
			// one (because the call to getComponentConfigVersionStates requires the currently installed version).
			// This also makes the KubernetesVersion value returned for `upgrade plan` consistent as that command
			// allows to not specify a target version in which case KubernetesVersion will always hold the currently
			// installed one.
			cfg.KubernetesVersion = newK8sVersion
		}
	}

	ignorePreflightErrorsSet, err := validation.ValidateIgnorePreflightErrors(flags.ignorePreflightErrors, cfg.NodeRegistration.IgnorePreflightErrors)
	if err != nil {
		return nil, nil, nil, err
	}
	// Also set the union of pre-flight errors to InitConfiguration, to provide a consistent view of the runtime configuration:
	cfg.NodeRegistration.IgnorePreflightErrors = sets.List(ignorePreflightErrorsSet)

	// Ensure the user is root
	klog.V(1).Info("running preflight checks")
	if err := runPreflightChecks(client, ignorePreflightErrorsSet, printer); err != nil {
		return nil, nil, nil, err
	}

	// Run healthchecks against the cluster
	if err := upgrade.CheckClusterHealth(client, &cfg.ClusterConfiguration, ignorePreflightErrorsSet); err != nil {
		return nil, nil, nil, errors.Wrap(err, "[upgrade/health] FATAL")
	}

	// If features gates are passed to the command line, use it (otherwise use featureGates from configuration)
	if flags.featureGatesString != "" {
		cfg.FeatureGates, err = features.NewFeatureGate(&features.InitFeatureGates, flags.featureGatesString)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "[upgrade/config] FATAL")
		}
	}

	// Check if feature gate flags used in the cluster are consistent with the set of features currently supported by kubeadm
	if msg := features.CheckDeprecatedFlags(&features.InitFeatureGates, cfg.FeatureGates); len(msg) > 0 {
		for _, m := range msg {
			printer.Printf("[upgrade/config] %s\n", m)
		}
	}

	// If the user told us to print this information out; do it!
	if flags.printConfig {
		printConfiguration(&cfg.ClusterConfiguration, os.Stdout, printer)
	}

	// Use a real version getter interface that queries the API server, the kubeadm client and the Kubernetes CI system for latest versions
	return client, upgrade.NewOfflineVersionGetter(upgrade.NewKubeVersionGetter(client), newK8sVersion), cfg, nil
}

// printConfiguration prints the external version of the API to yaml
func printConfiguration(clustercfg *kubeadmapi.ClusterConfiguration, w io.Writer, printer output.Printer) {
	// Short-circuit if cfg is nil, so we can safely get the value of the pointer below
	if clustercfg == nil {
		return
	}

	cfgYaml, err := configutil.MarshalKubeadmConfigObject(clustercfg, kubeadmapiv1.SchemeGroupVersion)
	if err == nil {
		printer.Fprintln(w, "[upgrade/config] Configuration used:")

		scanner := bufio.NewScanner(bytes.NewReader(cfgYaml))
		for scanner.Scan() {
			printer.Fprintf(w, "\t%s\n", scanner.Text())
		}
	}
}

// runPreflightChecks runs the root preflight check
func runPreflightChecks(client clientset.Interface, ignorePreflightErrors sets.Set[string], printer output.Printer) error {
	printer.Printf("[preflight] Running pre-flight checks.\n")
	err := preflight.RunRootCheckOnly(ignorePreflightErrors)
	if err != nil {
		return err
	}
	return upgrade.RunCoreDNSMigrationCheck(client, ignorePreflightErrors)
}

// getClient gets a real or fake client depending on whether the user is dry-running or not
func getClient(file string, dryRun bool) (clientset.Interface, error) {
	if dryRun {
		dryRunGetter, err := apiclient.NewClientBackedDryRunGetterFromKubeconfig(file)
		if err != nil {
			return nil, err
		}

		// In order for fakeclient.Discovery().ServerVersion() to return the backing API Server's
		// real version; we have to do some clever API machinery tricks. First, we get the real
		// API Server's version
		realServerVersion, err := dryRunGetter.Client().Discovery().ServerVersion()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get server version")
		}

		// Get the fake clientset
		dryRunOpts := apiclient.GetDefaultDryRunClientOptions(dryRunGetter, os.Stdout)
		// Print GET and LIST requests
		dryRunOpts.PrintGETAndLIST = true
		fakeclient := apiclient.NewDryRunClientWithOpts(dryRunOpts)
		// As we know the return of Discovery() of the fake clientset is of type *fakediscovery.FakeDiscovery
		// we can convert it to that struct.
		fakeclientDiscovery, ok := fakeclient.Discovery().(*fakediscovery.FakeDiscovery)
		if !ok {
			return nil, errors.New("couldn't set fake discovery's server version")
		}
		// Lastly, set the right server version to be used
		fakeclientDiscovery.FakedServerVersion = realServerVersion
		// return the fake clientset used for dry-running
		return fakeclient, nil
	}
	return kubeconfigutil.ClientSetFromFile(file)
}

// getWaiter gets the right waiter implementation
func getWaiter(dryRun bool, client clientset.Interface, timeout time.Duration) apiclient.Waiter {
	if dryRun {
		return dryrunutil.NewWaiter()
	}
	return apiclient.NewKubeWaiter(client, timeout, os.Stdout)
}
