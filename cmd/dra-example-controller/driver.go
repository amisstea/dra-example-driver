/*
 * Copyright 2023 The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1alpha2"
	"k8s.io/dynamic-resource-allocation/controller"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	spacecrd "sigs.k8s.io/dra-example-driver/api/example.com/resource/space/v1alpha1"
	"sigs.k8s.io/dra-example-driver/pkg/flags"
)

const (
	DriverAPIGroup     = spacecrd.GroupName
	ResourceClaimLabel = DriverAPIGroup + "/resourceclaim"
)

type driver struct {
	lock       *PerClaimMutex
	clientsets flags.ClientSets
}

var _ controller.Driver = &driver{}

func NewDriver(config *Config) *driver {
	return &driver{
		lock:       NewPerClaimMutex(),
		clientsets: config.clientSets,
	}
}

func (d *driver) GetClassParameters(ctx context.Context, class *resourcev1.ResourceClass) (interface{}, error) {
	logger := klog.FromContext(ctx)
	logger.Info("GetClassParameters", "class", class.Name)
	return nil, nil
}

func (d *driver) GetClaimParameters(ctx context.Context, claim *resourcev1.ResourceClaim, class *resourcev1.ResourceClass, classParameters interface{}) (interface{}, error) {
	logger := klog.FromContext(ctx)
	logger.Info("GetClaimParameters", "claim", claim.Name, "class", class.Name)
	if claim.Spec.ParametersRef == nil {
		return spacecrd.DefaultSpaceClaimParametersSpec(), nil
	}
	if claim.Spec.ParametersRef.APIGroup != DriverAPIGroup {
		return nil, fmt.Errorf("incorrect API group: %v", claim.Spec.ParametersRef.APIGroup)
	}

	switch claim.Spec.ParametersRef.Kind {
	case spacecrd.SpaceClaimParametersKind:
		params, err := d.clientsets.Example.SpaceV1alpha1().SpaceClaimParameters(claim.Namespace).Get(ctx, claim.Spec.ParametersRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("error getting SpaceClaimParameters called '%v' in namespace '%v': %v", claim.Spec.ParametersRef.Name, claim.Namespace, err)
		}
		// TODO validate the claim params
		return &params.Spec, nil
	default:
		return nil, fmt.Errorf("unknown ResourceClaim.ParametersRef.Kind: %v", claim.Spec.ParametersRef.Kind)
	}
}

func (d *driver) Allocate(ctx context.Context, cas []*controller.ClaimAllocation, selectedNode string) {
	logger := klog.FromContext(ctx)
	logger.Info("Allocate", "numClaims", len(cas))

	for _, ca := range cas {
		ca.Allocation, ca.Error = d.allocate(ctx, ca.Claim, ca.ClaimParameters, ca.Class, ca.ClassParameters, selectedNode)
	}
}

func (d *driver) allocate(ctx context.Context, claim *resourcev1.ResourceClaim, claimParameters interface{}, class *resourcev1.ResourceClass, classParameters interface{}, selectedNode string) (*resourcev1.AllocationResult, error) {
	if selectedNode == "" {
		return nil, fmt.Errorf("TODO: immediate allocations is not yet supported")
	}

	logger := klog.FromContext(ctx)

	claimUid := string(claim.GetUID())

	d.lock.Get(claimUid).Lock()
	defer d.lock.Get(claimUid).Unlock()

	result := &resourcev1.AllocationResult{Shareable: true}

	ns, err := d.getNamespace(ctx, claimUid)
	if err != nil {
		return nil, fmt.Errorf("unable to get namespace for claim: %v", err)
	}

	if ns == nil {
		claimParams, ok := claimParameters.(*spacecrd.SpaceClaimParametersSpec)
		if !ok {
			return nil, fmt.Errorf("unknown ResourceClaim.ParametersRef.Kind: %v", claim.Spec.ParametersRef.Kind)
		}

		ns, err = d.createNamespace(ctx, claimUid, claimParams.GenerateName)
		if err != nil {
			return nil, fmt.Errorf("namespace creation failed: %v", err)
		}
	} else {
		logger.Info("found an existing namespace for a claim", "claimUid", claimUid)
	}

	// Pass the namespace name to the kubelet plugin. This will be used as a "device" identifier for CDI.
	result.ResourceHandles = []resourcev1.ResourceHandle{
		{
			DriverName: spacecrd.GroupName,
			Data:       ns.GetName(),
		},
	}

	return result, nil
}

func (d *driver) Deallocate(ctx context.Context, claim *resourcev1.ResourceClaim) error {
	logger := klog.FromContext(ctx)
	logger.Info("Deallocate", "claim", claim.Name)

	claimUid := string(claim.GetUID())

	d.lock.Get(claimUid).Lock()
	defer d.lock.Get(claimUid).Unlock()

	ns, err := d.getNamespace(ctx, claimUid)
	if err != nil {
		return fmt.Errorf("unable to get namespace for claim: %v", err)
	}

	if ns != nil {
		err = d.deleteNamespace(ctx, ns)
		if err != nil {
			return fmt.Errorf("unable to delete namespace for claim: %v", err)
		}
	}

	return nil
}

func (d *driver) UnsuitableNodes(ctx context.Context, pod *corev1.Pod, cas []*controller.ClaimAllocation, potentialNodes []string) error {
	logger := klog.FromContext(ctx)
	logger.Info("UnsuitableNodes", "pod", pod.Name, "numClaims", len(cas), "potentialNodes", potentialNodes)

	// All nodes are suitable since namespaces aren't coupled to the host
	for _, ca := range cas {
		ca.UnsuitableNodes = []string{}
	}

	return nil
}

func (d *driver) getNamespace(ctx context.Context, claimUid string) (*corev1.Namespace, error) {
	api := d.clientsets.Core.CoreV1().Namespaces()
	selector := ResourceClaimLabel + "=" + string(claimUid)
	namespaces, err := api.List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("unable to list namespaces: %v", err)
	}

	if len(namespaces.Items) == 0 {
		return nil, nil
	} else if len(namespaces.Items) > 1 {
		return nil, fmt.Errorf("more than one namespace found for claimUid: %s", claimUid)
	}

	return &namespaces.Items[0], nil
}

func (d *driver) createNamespace(ctx context.Context, claimUid string, generateName string) (*corev1.Namespace, error) {
	logger := klog.FromContext(ctx)

	labels := map[string]string{ResourceClaimLabel: string(claimUid)}
	spec := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: generateName,
			Labels:       labels,
		},
	}

	api := d.clientsets.Core.CoreV1().Namespaces()
	ns, err := api.Create(ctx, spec, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to create namespace: %v", err)
	}

	logger.Info("created namespace", "claimUid", claimUid, "namespace", ns.Name)
	return ns, nil
}

func (d *driver) deleteNamespace(ctx context.Context, ns *corev1.Namespace) error {
	logger := klog.FromContext(ctx)

	namespaces := d.clientsets.Core.CoreV1().Namespaces()
	err := namespaces.Delete(ctx, ns.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	logger.Info("Deleted namespace", "namespace", ns.Name)
	return nil
}
