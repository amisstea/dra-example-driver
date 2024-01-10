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
	"os"
	"strings"

	cdiapi "github.com/container-orchestrated-devices/container-device-interface/pkg/cdi"
	cdispec "github.com/container-orchestrated-devices/container-device-interface/specs-go"
	"k8s.io/klog/v2"
)

const (
	cdiVendor = "k8s." + DriverName
	cdiClass  = "space"
	cdiKind   = cdiVendor + "/" + cdiClass

	cdiCommonDeviceName = "common"
)

type CDIHandler struct {
	registry cdiapi.Registry
}

func NewCDIHandler(config *Config) (*CDIHandler, error) {
	registry := cdiapi.GetRegistry(
		cdiapi.WithSpecDirs(config.flags.cdiRoot),
	)

	err := registry.Refresh()
	if err != nil {
		return nil, fmt.Errorf("unable to refresh the CDI registry: %v", err)
	}

	handler := &CDIHandler{
		registry: registry,
	}

	return handler, nil
}

func (cdi *CDIHandler) CreateCommonSpecFile() error {
	spec := &cdispec.Spec{
		Kind: cdiKind,
		Devices: []cdispec.Device{
			{
				Name: cdiCommonDeviceName,
				ContainerEdits: cdispec.ContainerEdits{
					Env: []string{
						fmt.Sprintf("DRA_RESOURCE_DRIVER_NAME=%s", DriverName),
					},
				},
			},
		},
	}

	minVersion, err := cdiapi.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("failed to get minimum required CDI spec version: %v", err)
	}
	spec.Version = minVersion

	specName, err := cdiapi.GenerateNameForTransientSpec(spec, cdiCommonDeviceName)
	if err != nil {
		return fmt.Errorf("failed to generate Spec name: %w", err)
	}

	return cdi.registry.SpecDB().WriteSpec(spec, specName)
}

func (cdi *CDIHandler) CreateClaimSpecFile(claimUid string, claimName string, space string) error {
	logger := klog.FromContext(context.TODO())
	specName := cdiapi.GenerateTransientSpecName(cdiVendor, cdiClass, claimUid)

	// TODO look for an alternative naming convention which doesn't use the claim name.
	// Claim names can be dynamic when using ResourceClaimTemplates.
	envBase := strings.ReplaceAll(strings.ToUpper(claimName), "-", "_")

	// TODO remove hard coded host path prefix
	hostPath := fmt.Sprintf("/var/run/claim-artifacts/%s", claimUid)
	kubeConfigPath := fmt.Sprintf("%s/kubeconfig", hostPath)
	containerPath := fmt.Sprintf("/etc/%s", claimName)

	logger.Info("creating claim artifacts", "claimUid", claimUid, "hostPath", hostPath)
	err := os.MkdirAll(hostPath, os.ModeDir)
	if err != nil {
		return err
	}

	kubeconfig, err := os.Create(kubeConfigPath)
	if err != nil {
		return err
	}
	defer kubeconfig.Close()

	_, err = kubeconfig.WriteString("TODO")
	if err != nil {
		return err
	}

	cdiDevice := cdispec.Device{
		Name: space,
		ContainerEdits: cdispec.ContainerEdits{
			Env: []string{
				fmt.Sprintf("%s_CLUSTER=%s", envBase, "https://cluster.todo"),
				fmt.Sprintf("%s_NAMESPACE=%s", envBase, space),
				fmt.Sprintf("%s_KUBECONFIG=%s", envBase, kubeConfigPath),
			},
			Mounts: []*cdispec.Mount{
				{
					HostPath:      hostPath,
					ContainerPath: containerPath,
					Options:       []string{"bind"}, // TODO is this necessary?
				},
			},
		},
	}

	spec := &cdispec.Spec{
		Kind:    cdiKind,
		Devices: []cdispec.Device{cdiDevice},
	}

	minVersion, err := cdiapi.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("failed to get minimum required CDI spec version: %v", err)
	}
	spec.Version = minVersion

	logger.Info("creating CDI spec", "claimUid", claimUid, "specName", specName)
	return cdi.registry.SpecDB().WriteSpec(spec, specName)
}

func (cdi *CDIHandler) DeleteClaimSpecFile(claimUid string) error {
	logger := klog.FromContext(context.TODO())

	hostPath := fmt.Sprintf("/tmp/%s", claimUid)

	logger.Info("deleting claim artifacts", "claimUid", claimUid, "hostPath", hostPath)
	err := os.RemoveAll(hostPath)
	if err != nil {
		return err
	}

	logger.Info("deleting CDI spec", "claimUid", claimUid)
	specName := cdiapi.GenerateTransientSpecName(cdiVendor, cdiClass, claimUid)
	return cdi.registry.SpecDB().RemoveSpec(specName)
}

func (cdi *CDIHandler) GetClaimDevices(claimUid string, space string) []string {
	return []string{
		cdiapi.QualifiedName(cdiVendor, cdiClass, cdiCommonDeviceName),
		cdiapi.QualifiedName(cdiVendor, cdiClass, space),
	}
}
