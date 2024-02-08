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
	var configPath *string
	if home := homedir.HomeDir(); home != "" {
		configPath = flag.String("k", filepath.Join(home, ".kube", "config"), "(optional) absolute `path` to the configPath file")
	} else {
		configPath = flag.String("k", "", "absolute `path` to the configPath file")
	}
	context := flag.String("c", "", "(optional) target `context`")
	namespace := flag.String("n", "", "(optional) target `namespace`")
	verbose := flag.Bool("v", false, "(optional) verbose output")
	output := flag.Bool("o", false, "(optional) output the current context and namespace")
	flag.Parse()
	config, configErr := readConfig(*configPath)
	if configErr != nil {
		fmt.Println(configErr)
		os.Exit(1)
	}
	if *output {
		fmt.Printf("current context '%s' namespace '%s'\n", config.CurrentContext, config.Contexts[config.CurrentContext].Namespace)
	}
	if !*output && *context == "" && *namespace == "" {
		usage()
		os.Exit(1)
	}
	if *context != "" {
		err := switchContext(*context, config, *verbose)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	if *namespace != "" {
		err := switchNamespace(*namespace, config, *verbose)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	os.Exit(0)
}

func usage() {
	_, _ = fmt.Fprintln(flag.CommandLine.Output(), "usage of kctx:\nswitches kubernetes contexts and namespaces")
	flag.PrintDefaults()
}

func switchContext(context string, config api.Config, verbose bool) error {
	if config.Contexts[context] == nil {
		return fmt.Errorf("context '%s' does not exist", context)
	}
	if config.CurrentContext == context {
		if verbose {
			fmt.Printf("context is already '%s'\n", context)
		}
		return nil
	}
	config.CurrentContext = context
	if writeErr := writeConfig(config); writeErr != nil {
		return fmt.Errorf("error writing config: %w", writeErr)
	}
	if verbose {
		fmt.Printf("switched to context '%s'\n", context)
	}
	return nil
}

func switchNamespace(namespace string, config api.Config, verbose bool) error {
	context := config.Contexts[config.CurrentContext]
	if context == nil {
		return fmt.Errorf("context '%s' does not exist", config.CurrentContext)
	}
	if context.Namespace == namespace {
		if verbose {
			fmt.Printf("namespace is already '%s'\n", namespace)
		}
		return nil
	}
	context.Namespace = namespace
	if err := writeConfig(config); err != nil {
		return fmt.Errorf("error writing kubeconfig: %w", err)
	}
	if verbose {
		fmt.Printf("switched to namespace '%s'\n", namespace)
	}
	return nil
}

func readConfig(configPath string) (api.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: configPath}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.RawConfig()
}

func writeConfig(config api.Config) error {
	return clientcmd.ModifyConfig(clientcmd.NewDefaultPathOptions(), config, true)
}
