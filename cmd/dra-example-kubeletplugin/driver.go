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
	"sync"

	"k8s.io/klog/v2"
	dra "k8s.io/kubelet/pkg/apis/dra/v1alpha3"
	spacecrd "sigs.k8s.io/dra-example-driver/api/example.com/resource/space/v1alpha1"
)

const (
	DriverAPIGroup     = spacecrd.GroupName
	ResourceClaimLabel = DriverAPIGroup + "/resourceclaim"
)

var _ dra.NodeServer = &driver{}

type driver struct {
	sync.Mutex
	cdi *CDIHandler
}

func NewDriver(ctx context.Context, config *Config) (*driver, error) {
	logger := klog.FromContext(ctx)
	logger.Info("NewDriver (kubelet plugin)")

	cdi, err := NewCDIHandler(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI handler: %v", err)
	}

	err = cdi.CreateCommonSpecFile()
	if err != nil {
		return nil, fmt.Errorf("unable to create CDI spec file for common edits: %v", err)
	}

	d := &driver{cdi: cdi}

	return d, nil
}

func (d *driver) Shutdown(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	logger.Info("Shutdown")
	return nil
}

func (d *driver) NodePrepareResources(ctx context.Context, req *dra.NodePrepareResourcesRequest) (*dra.NodePrepareResourcesResponse, error) {
	logger := klog.FromContext(ctx)
	logger.Info("NodePrepareResources", "numClaims", len(req.Claims))
	rsp := &dra.NodePrepareResourcesResponse{Claims: map[string]*dra.NodePrepareResourceResponse{}}
	for _, claim := range req.Claims {
		rsp.Claims[claim.Uid] = d.prepareResource(ctx, claim)
	}
	return rsp, nil
}

func (d *driver) prepareResource(ctx context.Context, claim *dra.Claim) *dra.NodePrepareResourceResponse {
	d.Lock()
	defer d.Unlock()

	rsp := &dra.NodePrepareResourceResponse{}
	ns := claim.GetResourceHandle()

	err := d.cdi.CreateClaimSpecFile(claim.Uid, claim.Name, ns)
	if err != nil {
		rsp.Error = fmt.Sprintf("unable to create CDI spec file for claim: %v", err)
		return rsp
	}

	cdiDevices := d.cdi.GetClaimDevices(claim.Uid, ns)
	if err != nil {
		rsp.Error = fmt.Sprintf("unable to get CDI devices names: %v", err)
		return rsp
	}

	rsp.CDIDevices = cdiDevices

	return rsp
}

func (d *driver) NodeUnprepareResources(ctx context.Context, req *dra.NodeUnprepareResourcesRequest) (*dra.NodeUnprepareResourcesResponse, error) {
	logger := klog.FromContext(ctx)
	logger.Info("NodeUnprepareResources", "numClaims", len(req.Claims))

	rsp := &dra.NodeUnprepareResourcesResponse{Claims: map[string]*dra.NodeUnprepareResourceResponse{}}

	for _, claim := range req.Claims {
		rsp.Claims[claim.Uid] = d.unprepareResource(ctx, claim)
	}

	return rsp, nil
}

func (d *driver) unprepareResource(ctx context.Context, claim *dra.Claim) *dra.NodeUnprepareResourceResponse {
	d.Lock()
	defer d.Unlock()

	rsp := &dra.NodeUnprepareResourceResponse{}

	err := d.cdi.DeleteClaimSpecFile(claim.Uid)
	if err != nil {
		rsp.Error = fmt.Sprintf("unable to delete CDI spec file for claim: %v", err)
	}

	return rsp
}
