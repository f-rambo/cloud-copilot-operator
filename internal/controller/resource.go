package controller

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func createResource(ctx context.Context, client dynamic.Interface, resource *unstructured.Unstructured) error {
	gvk := resource.GroupVersionKind()
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
	_, err := client.Resource(gvr).Namespace(resource.GetNamespace()).Create(ctx, resource, metav1.CreateOptions{})
	return err
}

func deleteResource(ctx context.Context, client dynamic.Interface, resource *unstructured.Unstructured) error {
	gvk := resource.GroupVersionKind()
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
	err := client.Resource(gvr).Namespace(resource.GetNamespace()).Delete(ctx, resource.GetName(), metav1.DeleteOptions{})
	return err
}

func getResource(ctx context.Context, client dynamic.Interface, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvk := resource.GroupVersionKind()
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
	return client.Resource(gvr).Namespace(resource.GetNamespace()).Get(ctx, resource.GetName(), metav1.GetOptions{})
}

func updateResource(ctx context.Context, client dynamic.Interface, resource *unstructured.Unstructured) error {
	gvk := resource.GroupVersionKind()
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s",
	}
	_, err := client.Resource(gvr).Namespace(resource.GetNamespace()).Update(ctx, resource, metav1.UpdateOptions{})
	return err
}

func removeString(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
