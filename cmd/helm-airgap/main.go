package main

import (
	"fmt"
	"strings"

	"github.com/crossstratos/helm-airgap/pkg/kubernetes/images"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
)

func main() {
	client := action.NewInstall(&action.Configuration{})

	client.DryRun = true
	client.Atomic = true
	client.ClientOnly = true
	client.GenerateName = true
	client.ReleaseName = "test"

	c, err := loader.Load("./testchart")
	if err != nil {
		panic(err)
	}

	r, err := client.Run(c, map[string]interface{}{})
	if err != nil {
		panic(err)
	}

	s := strings.Split(r.Manifest, "---")
	for i := range s {
		if len(s[i]) == 0 {
			continue
		}

		images, err := images.ParseImagesFromYaml([]byte(s[i]))
		if err != nil {
			panic(err)
		}

		for i := range images {
			fmt.Println(images[i].String())
		}
	}
}
