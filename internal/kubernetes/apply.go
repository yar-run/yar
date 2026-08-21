package kubernetes

import (
	"bytes"
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"unicode"

	yarerrors "github.com/yar-run/yar/internal/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
)

const yamlDecoderBufferSize = 4096

// ApplyOptions configures a server-side apply operation.
type ApplyOptions struct {
	FieldManager string
	Force        *bool
	DryRun       bool
}

// DeleteOptions configures a delete operation.
type DeleteOptions struct {
	PropagationPolicy string
}

// ListOptions configures a list operation.
type ListOptions struct {
	LabelSelector string
	Limit         int64
	Continue      string
}

// Apply applies a multi-document manifest stream through server-side apply.
func (c *client) Apply(ctx context.Context, manifests []byte, opts ApplyOptions) error {
	if err := c.owner.validate(); err != nil {
		return err
	}
	objects, err := decodeManifests(manifests)
	if err != nil {
		return err
	}
	fieldManager, err := c.applyFieldManager(opts.FieldManager)
	if err != nil {
		return err
	}

	ctx, cancel := c.operationContext(ctx)
	defer cancel()
	for _, object := range objects {
		if err := ctx.Err(); err != nil {
			return err
		}
		resource, mapping, err := c.resourceFor(object.GroupVersionKind(), object.GetNamespace())
		if err != nil {
			return err
		}
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace && object.GetNamespace() == "" {
			object.SetNamespace(c.namespace)
		}
		namespace := resourceNamespace(mapping, object.GetNamespace())
		live, err := resource.Get(ctx, object.GetName(), metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			stampOwnership(object, c.owner)
			if _, err := resource.Create(ctx, object, createOptions(fieldManager, opts.DryRun)); err == nil {
				continue
			} else if !apierrors.IsAlreadyExists(err) {
				if isContextError(err) {
					return err
				}
				return resourceOperationError("apply", mapping.Resource.Resource, object.GetName(), namespace, err)
			}

			// A concurrent creator won the race. Re-read and prove ownership before
			// considering the object eligible for server-side apply.
			live, err = resource.Get(ctx, object.GetName(), metav1.GetOptions{})
		}
		if err != nil {
			if isContextError(err) {
				return err
			}
			return resourceOperationError("apply", mapping.Resource.Resource, object.GetName(), namespace, err)
		}
		if !c.owner.matches(live.GetLabels()) {
			return resourceOperationError("apply", mapping.Resource.Resource, object.GetName(), namespace, fmt.Errorf("resource is not managed by this Yar project and environment"))
		}
		stampOwnership(object, c.owner)
		// Pin the update to the object inspected above. This keeps a replacement
		// resource from being applied after ownership was verified.
		object.SetUID(live.GetUID())

		applyOptions := metav1.ApplyOptions{FieldManager: fieldManager, Force: true}
		if opts.Force != nil {
			applyOptions.Force = *opts.Force
		}
		if opts.DryRun {
			applyOptions.DryRun = []string{metav1.DryRunAll}
		}
		if _, err := resource.Apply(ctx, object.GetName(), object, applyOptions); err != nil {
			return resourceOperationError("apply", mapping.Resource.Resource, object.GetName(), resourceNamespace(mapping, object.GetNamespace()), err)
		}
	}
	return nil
}

// Delete removes Yar-owned resources declared by a multi-document manifest stream.
func (c *client) Delete(ctx context.Context, manifests []byte, opts DeleteOptions) error {
	if err := c.owner.validate(); err != nil {
		return err
	}
	objects, err := decodeManifests(manifests)
	if err != nil {
		return err
	}
	propagationPolicy, err := deletePropagationPolicy(opts)
	if err != nil {
		return err
	}

	ctx, cancel := c.operationContext(ctx)
	defer cancel()
	for _, object := range objects {
		if err := ctx.Err(); err != nil {
			return err
		}
		resource, mapping, err := c.resourceFor(object.GroupVersionKind(), object.GetNamespace())
		if err != nil {
			return err
		}
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace && object.GetNamespace() == "" {
			object.SetNamespace(c.namespace)
		}
		namespace := resourceNamespace(mapping, object.GetNamespace())
		live, err := resource.Get(ctx, object.GetName(), metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			if isContextError(err) {
				return err
			}
			return resourceOperationError("delete", mapping.Resource.Resource, object.GetName(), namespace, err)
		}
		if !c.owner.matches(live.GetLabels()) {
			return resourceOperationError("delete", mapping.Resource.Resource, object.GetName(), namespace, fmt.Errorf("resource is not managed by this Yar project and environment"))
		}

		deleteOptions := metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{},
		}
		if uid := live.GetUID(); uid != "" {
			deleteOptions.Preconditions.UID = &uid
		}
		if resourceVersion := live.GetResourceVersion(); resourceVersion != "" {
			deleteOptions.Preconditions.ResourceVersion = &resourceVersion
		}
		if deleteOptions.Preconditions.UID == nil && deleteOptions.Preconditions.ResourceVersion == nil {
			deleteOptions.Preconditions = nil
		}
		if propagationPolicy != nil {
			deleteOptions.PropagationPolicy = propagationPolicy
		}
		if err := resource.Delete(ctx, object.GetName(), deleteOptions); err != nil {
			if isContextError(err) {
				return err
			}
			return resourceOperationError("delete", mapping.Resource.Resource, object.GetName(), namespace, err)
		}
	}
	return nil
}

// Get retrieves one resource by GVK, namespace, and name.
func (c *client) Get(ctx context.Context, gvk schema.GroupVersionKind, namespace, name string) (*unstructured.Unstructured, error) {
	if name == "" {
		return nil, validationError("name", name, "is required")
	}
	ctx, cancel := c.operationContext(ctx)
	defer cancel()
	resource, mapping, err := c.resourceFor(gvk, namespace)
	if err != nil {
		return nil, err
	}
	object, err := resource.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if isContextError(err) {
			return nil, err
		}
		return nil, resourceOperationError("get", mapping.Resource.Resource, name, c.operationNamespace(mapping, namespace), err)
	}
	return object, nil
}

// List retrieves resources matching the supplied GVK, namespace, and options.
func (c *client) List(ctx context.Context, gvk schema.GroupVersionKind, namespace string, opts ListOptions) (*unstructured.UnstructuredList, error) {
	if err := validateListOptions(opts); err != nil {
		return nil, err
	}
	ctx, cancel := c.operationContext(ctx)
	defer cancel()
	resource, mapping, err := c.resourceForList(gvk, namespace)
	if err != nil {
		return nil, err
	}
	objects, err := resource.List(ctx, metav1.ListOptions{LabelSelector: opts.LabelSelector, Limit: opts.Limit, Continue: opts.Continue})
	if err != nil {
		if isContextError(err) {
			return nil, err
		}
		return nil, resourceOperationError("list", mapping.Resource.Resource, "", resourceNamespace(mapping, namespace), err)
	}
	return objects, nil
}

func (c *client) applyFieldManager(fieldManager string) (string, error) {
	if fieldManager == "" {
		fieldManager = c.fieldManager
	}
	if fieldManager == "" {
		return "", validationError("fieldManager", fieldManager, "must not be empty")
	}
	if len(fieldManager) > 128 {
		return "", validationError("fieldManager", fieldManager, "must be at most 128 characters")
	}
	for _, character := range fieldManager {
		if !unicode.IsPrint(character) {
			return "", validationError("fieldManager", fieldManager, "must contain only printable characters")
		}
	}
	return fieldManager, nil
}

func (c *client) resourceFor(gvk schema.GroupVersionKind, namespace string) (dynamic.ResourceInterface, *meta.RESTMapping, error) {
	dynamicClient, mapper, err := c.resourceClient()
	if err != nil {
		return nil, nil, resourceOperationError("mapping", gvk.String(), "", "", err)
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, resourceOperationError("mapping", gvk.String(), "", "", err)
	}
	if mapping.Scope.Name() != meta.RESTScopeNameNamespace {
		return dynamicClient.Resource(mapping.Resource), mapping, nil
	}
	if namespace == "" {
		namespace = c.namespace
	}
	return dynamicClient.Resource(mapping.Resource).Namespace(namespace), mapping, nil
}

func (c *client) resourceForList(gvk schema.GroupVersionKind, namespace string) (dynamic.ResourceInterface, *meta.RESTMapping, error) {
	dynamicClient, mapper, err := c.resourceClient()
	if err != nil {
		return nil, nil, resourceOperationError("mapping", gvk.String(), "", "", err)
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, resourceOperationError("mapping", gvk.String(), "", "", err)
	}
	resource := dynamicClient.Resource(mapping.Resource)
	if mapping.Scope.Name() != meta.RESTScopeNameNamespace || namespace == "" {
		return resource, mapping, nil
	}
	return resource.Namespace(namespace), mapping, nil
}

func decodeManifests(manifests []byte) ([]*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifests), yamlDecoderBufferSize)
	objects := make([]*unstructured.Unstructured, 0)
	for {
		object := &unstructured.Unstructured{}
		err := decoder.Decode(object)
		if stderrors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, validationError("manifest", "", "is invalid")
		}
		if len(object.Object) == 0 {
			continue
		}
		if err := validateManifest(object); err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	if len(objects) == 0 {
		return nil, validationError("manifest", "", "must contain at least one Kubernetes object")
	}
	return objects, nil
}

func validateManifest(object *unstructured.Unstructured) error {
	if object.GetAPIVersion() == "" {
		return validationError("manifest.apiVersion", "", "is required")
	}
	if object.GetKind() == "" {
		return validationError("manifest.kind", "", "is required")
	}
	if object.GetName() == "" {
		return validationError("manifest.metadata.name", "", "is required")
	}
	if _, found, err := unstructured.NestedFieldNoCopy(object.Object, "metadata", "managedFields"); err != nil {
		return validationError("manifest.metadata.managedFields", "", "must be a valid field")
	} else if found {
		return validationError("manifest.metadata.managedFields", "", "must not be set")
	}
	for _, field := range []string{"uid", "resourceVersion"} {
		if _, found, err := unstructured.NestedFieldNoCopy(object.Object, "metadata", field); err != nil {
			return validationError("manifest.metadata."+field, "", "must be a valid field")
		} else if found {
			return validationError("manifest.metadata."+field, "", "must not be set")
		}
	}
	return nil
}

func stampOwnership(object *unstructured.Unstructured, owner ownerIdentity) {
	labels := object.GetLabels()
	if labels == nil {
		labels = make(map[string]string, len(owner.labels()))
	}
	for key, value := range owner.labels() {
		labels[key] = value
	}
	object.SetLabels(labels)
}

func (o ownerIdentity) matches(labels map[string]string) bool {
	return labels[labelManagedBy] == managedByYar && labels[labelProject] == o.project && labels[labelEnvironment] == o.environment
}

func resourceNamespace(mapping *meta.RESTMapping, namespace string) string {
	if mapping.Scope.Name() != meta.RESTScopeNameNamespace {
		return ""
	}
	return namespace
}

func (c *client) operationNamespace(mapping *meta.RESTMapping, namespace string) string {
	if mapping.Scope.Name() != meta.RESTScopeNameNamespace {
		return ""
	}
	if namespace == "" {
		return c.namespace
	}
	return namespace
}

func createOptions(fieldManager string, dryRun bool) metav1.CreateOptions {
	opts := metav1.CreateOptions{FieldManager: fieldManager}
	if dryRun {
		opts.DryRun = []string{metav1.DryRunAll}
	}
	return opts
}

func validateListOptions(opts ListOptions) error {
	if opts.Limit < 0 {
		return validationError("list.limit", opts.Limit, "must not be negative")
	}
	if opts.LabelSelector != "" {
		if _, err := labels.Parse(opts.LabelSelector); err != nil {
			return validationError("list.labelSelector", opts.LabelSelector, "must be a valid Kubernetes label selector")
		}
	}
	return nil
}

func deletePropagationPolicy(opts DeleteOptions) (*metav1.DeletionPropagation, error) {
	if opts.PropagationPolicy == "" {
		return nil, nil
	}
	var policy metav1.DeletionPropagation
	switch opts.PropagationPolicy {
	case "foreground":
		policy = metav1.DeletePropagationForeground
	case "background":
		policy = metav1.DeletePropagationBackground
	case "orphan":
		policy = metav1.DeletePropagationOrphan
	default:
		return nil, validationError("delete.propagationPolicy", opts.PropagationPolicy, "must be foreground, background, or orphan")
	}
	return &policy, nil
}

func resourceOperationError(op, resource, name, namespace string, err error) error {
	if isContextError(err) {
		return err
	}
	return &yarerrors.KubernetesError{Op: op, Resource: resource, Name: name, Namespace: namespace, Err: err}
}

func validationError(field string, value any, message string) *yarerrors.ValidationError {
	return &yarerrors.ValidationError{Field: field, Value: value, Message: message}
}
