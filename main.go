package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

	corev1 "k8s.io/api/core/v1"

	"github.com/mwlng/k8s_resources_sync/pkg/helpers"
	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
	"github.com/mwlng/k8s_resources_sync/pkg/utils"
)

const (
	defaultRegion  = "us-east-1"
	defaultEnviron = "alpha"
)

var (
	homeDir string
)

func init() {
	klog.InitFlags(nil)

	homeDir = utils.GetHomeDir()
}

func main() {
	defer func() {
		klog.Flush()
	}()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	environ := flag.String("e", defaultEnviron, "Target environment")
	tgtEksClusterName := flag.String("target_cluster_name", "", "Source k8s cluster name")
	namespace := flag.String("n", corev1.NamespaceDefault, "Target environment")
	replicas := flag.Int("replicas", 0, "Number of repicas will be changed to in the deployment(default=0)")

	flag.Set("v", "2")
	flag.Parse()

	if len(*tgtEksClusterName) == 0 {
		klog.Infoln("No specified source k8s cluster name, nothing to restore, exit !")
		Usage()
		os.Exit(0)
	}

	klog.Infoln("Loading client kubeconfig ...")

	tgtKubeConfig, err := helpers.GetKubeConfig(*tgtEksClusterName, *kubeconfig)
	if err != nil {
		klog.Errorf("Failed to create client kubeconfig, Err was %s\n", err)
		os.Exit(1)
	}

	klog.Infof("Envrionment: %s, Cluster: %s\n", *environ, tgtKubeConfig.Host)
	deployment, err := k8s_resources.NewDeployment(tgtKubeConfig, *namespace)
	if err != nil {
		klog.Errorf("Failed to create k8s deployment client, Err was %s\n", err)
		os.Exit(1)
	}

	klog.Infof("Starting to set deployment replicas to %d in namespace %s ...", *replicas, *namespace)
	deploymentList, err := deployment.ListDeployments()
	if err != nil {
		klog.Errorf("Failed to list services. Err was %s", err)
		os.Exit(1)
	}

	for _, d := range deploymentList.Items {
		result, err := deployment.GetDeployment(d.Name)
		if err != nil {
			klog.Errorf("Failed to fetch deployment: %s, skipped\n", d.Name)
		}
		r := int32(*replicas)
		result.Spec.Replicas = &r
		//deployment.UpdateDeployment(result)
	}
}

func Usage() {
	fmt.Println()
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
