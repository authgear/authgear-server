package main

// Usage: go run ./hack/k8s-migrate.go -namespace=<k8s namespace> -cmd=hack/migrate-script.js <resource kind>
//  e.g.: go run ./hack/k8s-migrate.go -namespace=authgear-apps -cmd=hack/migrate-script.js secret
// After preview the changes, add -dry-run=false

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/exec"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var kubeconfig *string
var namespace *string
var command *string
var dryRun *bool

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	namespace = flag.String("namespace", "", "k8s namespace")
	command = flag.String("cmd", "", "mapper command")
	dryRun = flag.Bool("dry-run", true, "is dry run")
}

func getResource(dc discovery.DiscoveryInterface, name string) metav1.APIResource {
	_, resourceLists, err := dc.ServerGroupsAndResources()
	if err != nil {
		panic(err)
	}

	for _, rl := range resourceLists {
		for _, r := range rl.APIResources {
			if strings.Contains(r.Name, "/") {
				// Ignore sub-resources
				continue
			}
			names := append([]string{r.Name, r.Kind, r.SingularName}, r.ShortNames...)
			for _, n := range names {
				if n != "" && strings.EqualFold(n, name) {
					return r
				}
			}
		}
	}
	panic("resource not found on server: " + name)
}

func mapResource(in unstructured.Unstructured) (out unstructured.Unstructured, processed bool) {
	log.Printf("processing: %s", in.GetName())

	data, err := json.Marshal(&in)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, *command)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	go func() {
		defer stdin.Close()
		stdin.Write(data)
	}()

	data, err = cmd.Output()
	if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 2 {
		log.Printf("skipped: %s", in.GetName())
		processed = false
		return
	}
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(data, &out); err != nil {
		panic(err)
	}
	processed = true
	return
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
	}()

	flag.Parse()
	targetResources := flag.Args()

	if *command == "" {
		panic("mapper command is required")
	} else if *namespace == "" {
		panic("k8s namespace is required")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	dc := memory.NewMemCacheClient(discovery.NewDiscoveryClientForConfigOrDie(config))
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(dc)
	dyn := dynamic.NewForConfigOrDie(config)

	for _, name := range targetResources {
		res := getResource(dc, name)

		mapping, err := mapper.RESTMapping(schema.GroupKind{Group: res.Group, Kind: res.Kind}, res.Version)
		if err != nil {
			panic(err)
		}
		apiVersion, kind := mapping.GroupVersionKind.ToAPIVersionAndKind()

		var client dynamic.ResourceInterface
		if res.Namespaced {
			client = dyn.Resource(mapping.Resource).Namespace(*namespace)
		} else {
			client = dyn.Resource(mapping.Resource)
		}

		log.Printf("listing %s.%s...", apiVersion, kind)
		list, err := client.List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		log.Printf("mapping %d resources", len(list.Items))
		var updatedItems []unstructured.Unstructured
		for _, item := range list.Items {
			out, processed := mapResource(item)
			if processed {
				updatedItems = append(updatedItems, out)
			}
		}

		if *dryRun {
			log.Print("dry run: resources to update:")
			for _, item := range updatedItems {
				data, err := json.MarshalIndent(&item, "", "  ")
				if err != nil {
					panic(err)
				}
				log.Printf("%s\n", string(data))
			}
		} else {
			log.Printf("updating %d resources", len(updatedItems))
			for _, item := range updatedItems {
				_, err := client.Update(context.Background(), &item, metav1.UpdateOptions{})
				log.Printf("%s: %v", item.GetName(), err)
			}
		}
	}
}
