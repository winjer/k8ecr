package resources

import (
	"fmt"

	apps "github.com/isotoma/k8ecr/pkg/imagemanager"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var deploymentResource = &apps.ResourceManager{
	Kind: "Deployment",
	Fetcher: func(mgr *apps.ImageManager) ([]interface{}, error) {
		client := mgr.ClientSet.AppsV1beta1().Deployments(mgr.Namespace)
		response, err := client.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		empty := make([]interface{}, len(response.Items))
		for i, item := range response.Items {
			empty[i] = item
		}
		return empty, nil
	},
	Generator: func(item interface{}) []apps.Resource {
		var d appsv1beta1.Deployment
		d = item.(appsv1beta1.Deployment)
		allResources := make([]apps.Resource, 0)
		for _, r := range resources(d.Name, d.ObjectMeta, d.Spec.Template.Spec.Containers) {
			allResources = append(allResources, r)
		}
		return allResources
	},
	Upgrade: func(mgr *apps.ImageManager, image *apps.ImageMap, resource apps.Resource) error {
		client := mgr.ClientSet.AppsV1beta1().Deployments(mgr.Namespace)
		item, err := client.Get(resource.ContainerID.Resource, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for i, container := range item.Spec.Template.Spec.Containers {
			if container.Name == resource.ContainerID.Container {
				fmt.Printf("        %s/%s image -> %s\n", resource.ContainerID.Resource, resource.ContainerID.Container, image.NewImage())
				item.Spec.Template.Spec.Containers[i].Image = image.NewImage()
			}
		}
		_, err = client.Update(item)
		return err
	},
}
