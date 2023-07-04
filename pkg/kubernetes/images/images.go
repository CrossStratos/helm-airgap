package images

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	kind = "Kind"
)

type Image struct {
	name     string
	registry string
	tag      string
}

func (i Image) String() string {
	return fmt.Sprintf("%s/%s:%s", i.registry, i.name, i.tag)
}

// ParseImagesFromYaml takes a kubernetes yaml resource
// and extracts images if possible.
func ParseImagesFromYaml(b []byte) ([]Image, error) {
	// Create a runtime.Decoder from the Codecs field within
	// k8s.io/client-go that's pre-loaded with the schemas for all
	// the standard Kubernetes resource types.
	decoder := scheme.Codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode(b, nil, nil)
	if err != nil {
		return nil, err
	}

	images := make([]Image, 0)

	switch resource := obj.(type) {
	case *appsv1.Deployment:
		images = append(images, imagesFromPodSpec(resource.Spec.Template.Spec)...)
	case *corev1.Pod:
		images = append(images, imagesFromPodSpec(resource.Spec)...)
	case *batchv1.Job:
		images = append(images, imagesFromPodSpec(resource.Spec.Template.Spec)...)
	}

	return images, nil
}

// imagesFromPodSpec parses images from a pod spec.
func imagesFromPodSpec(podSpec corev1.PodSpec) []Image {
	imageNames := make([]string, 0)
	for _, container := range podSpec.Containers {
		imageNames = append(imageNames, container.Image)
	}

	for _, container := range podSpec.InitContainers {
		imageNames = append(imageNames, container.Image)
	}

	for _, container := range podSpec.EphemeralContainers {
		imageNames = append(imageNames, container.Image)
	}

	return ParseImageNames(imageNames)
}

// ParseImageNames takes image names and splits them
// into an image struct.
func ParseImageNames(imageNames []string) []Image {
	images := make([]Image, 0, len(imageNames))

	uniqueImages := make(map[string]bool, 0)
	for i := range imageNames {
		exists, ok := uniqueImages[imageNames[i]]
		if ok || exists {
			continue
		}

		// Update image map if entry doesn't exist.
		uniqueImages[imageNames[i]] = true

		image := Image{registry: "docker.io", name: imageNames[i]}
		// Check if image has a tag (if it doesn't it's an invalid image)
		s := strings.Split(image.name, ":")
		if len(s) < 1 {
			panic("invalid image, missing tag")
		}

		image.name = s[0]
		image.tag = s[1]

		// Source: https://github.com/google/go-containerregistry/blob/main/pkg/name/repository.go#L81
		parts := strings.SplitN(image.name, "/", 2)
		if len(parts) == 2 && (strings.ContainsRune(parts[0], '.') || strings.ContainsRune(parts[0], ':')) {
			// The first part of the repository is treated as the registry domain
			// iff it contains a '.' or ':' character, otherwise it is all repository
			// and the domain defaults to Docker Hub.
			image.registry = parts[0]
			image.name = parts[1]
		}

		images = append(images, image)
	}

	return images
}
