package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("k", filepath.Join(home, ".kube", "config"), "(optional) absolute `path` to the kubeconfig file")
	} else {
		kubeconfig = flag.String("k", "", "absolute `path` to the kubeconfig file")
	}
	var context *string
	context = flag.String("c", "", "(optional) target `context`")
	var namespace *string
	namespace = flag.String("n", "", "(optional) target `namespace`")
	flag.Parse()
	if *context == "" && *namespace == "" {
		usage()
		os.Exit(1)
	}
	if *context != "" {
		err := switchContext(*context, *kubeconfig)
		panic(err)
	}
	if *namespace != "" {
		err := switchNamespace(*namespace, *kubeconfig)
		panic(err)
	}
}

func usage() {
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "usage of kctx:\nswitches kubernetes contexts and namespaces")
	flag.PrintDefaults()
}

func switchContext(context, kubeconfig string) error {
	config, err := readConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("error reading kubeconfig: %w", err)
	}
	if config.Contexts[context] == nil {
		return fmt.Errorf("context '%s' does not exist", context)
	}
	config.CurrentContext = context
	if err := writeConfig(config); err != nil {
		return fmt.Errorf("error writing kubeconfig: %w", err)
	}
	fmt.Printf("switched to context '%s'", context)
	return nil
}

func switchNamespace(namespace, kubeconfig string) error {
	config, err := readConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("error reading kubeconfig: %w", err)
	}
	context := config.Contexts[config.CurrentContext]
	if context == nil {
		return fmt.Errorf("context '%s' does not exist", config.CurrentContext)
	}
	context.Namespace = namespace
	if err := writeConfig(config); err != nil {
		return fmt.Errorf("error writing kubeconfig: %w", err)
	}
	fmt.Printf("switched to namespace '%s'", namespace)
	return nil
}

func readConfig(kubeconfig string) (api.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.RawConfig()
}

func writeConfig(config api.Config) (err error) {
	return clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), config, true)
}
