// Copyright 2022 Weaveworks or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MPL-2.0

package microvm

import (
	"context"
	"encoding/base64"
	"fmt"

	flclient "github.com/weaveworks-liquidmetal/controller-pkg/client"
	microvm "github.com/weaveworks-liquidmetal/controller-pkg/types/microvm"
	flintlockv1 "github.com/weaveworks-liquidmetal/flintlock/api/services/microvm/v1alpha1"
	flintlocktypes "github.com/weaveworks-liquidmetal/flintlock/api/types"
	"github.com/weaveworks-liquidmetal/flintlock/client/cloudinit/instance"
	"github.com/weaveworks-liquidmetal/flintlock/client/cloudinit/userdata"
	"github.com/yitsushi/macpot"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v2"
	"k8s.io/utils/pointer"
)

const (
	cloudInitHeader = "#cloud-config\n"
)

// Scope contains functions for operators to provide microvm request  configuration
// to the service.
type Scope interface {
	// MicroVMSpec returns the full spec as configured by the calling operator.
	MicroVMSpec() microvm.VMSpec
	// Name returns the kubernetes name of the object creating the microvm.
	Name() string
	// Namespace returns the kubernetes namespace of the object creating the microvm.
	Namespace() string
	// InstanceID returns the UUID of the microvm.
	InstanceID() string
	// UserData returns customised commands/cloud init to be run at microvm boot.
	UserData() string
	// SSHKeys returns the public keys to be added to the microvm.
	SSHKeys() []microvm.SSHPublicKey
}

type Service struct {
	scope Scope

	client flclient.Client
	hostID string
}

func New(scope Scope, client flclient.Client, hostID string) *Service {
	return &Service{
		scope:  scope,
		client: client,
		hostID: hostID,
	}
}

func (s *Service) Create(ctx context.Context) (*flintlocktypes.MicroVM, error) {
	// s.scope.V(defaults.LogLevelDebug).Info("Creating microvm",
	// 	"machine-name", s.scope.Name(),
	// 	"cluster-name", s.scope.ClusterName())

	apiMicroVM := convertToFlintlockAPI(s.scope)

	if err := s.addMetadata(apiMicroVM); err != nil {
		return nil, fmt.Errorf("adding metadata: %w", err)
	}

	for i := range apiMicroVM.Interfaces {
		iface := apiMicroVM.Interfaces[i]

		if iface.GuestMac == nil || *iface.GuestMac == "" {
			mac, err := macpot.New(macpot.AsLocal(), macpot.AsUnicast())
			if err != nil {
				return nil, fmt.Errorf("creating mac address client: %w", err)
			}

			iface.GuestMac = pointer.String(mac.ToString())
		}
	}

	input := &flintlockv1.CreateMicroVMRequest{
		Microvm: apiMicroVM,
	}

	resp, err := s.client.CreateMicroVM(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("creating microvm: %w", err)
	}

	// s.scope.
	// 	V(2).
	// 	Info("Successfully created microvm", "machine-name", s.scope.Name())

	return resp.Microvm, nil
}

func (s *Service) Get(ctx context.Context) (*flintlocktypes.MicroVM, error) {
	// s.scope.
	// 	V(2).
	// 	Info("Getting microvm", "name", s.scope.Name())

	input := &flintlockv1.GetMicroVMRequest{
		Uid: s.scope.InstanceID(),
	}

	resp, err := s.client.GetMicroVM(ctx, input)
	if err != nil {
		return nil, err
	}

	return resp.Microvm, nil
}

func (s *Service) Delete(ctx context.Context) (*emptypb.Empty, error) {
	// s.scope.
	// 	V(2).
	// 	Info("Deleting microvm", "name", s.scope.Name())

	input := &flintlockv1.DeleteMicroVMRequest{
		Uid: s.scope.InstanceID(),
	}

	return s.client.DeleteMicroVM(ctx, input)
}

func (s *Service) Close() {
	s.client.Close()
}

func (s *Service) addMetadata(apiMicroVM *flintlocktypes.MicroVMSpec) error {
	apiMicroVM.Metadata["user-data"] = s.scope.UserData()

	vendorData, err := s.createVendorData()
	if err != nil {
		return fmt.Errorf("creating vendor data for microvm: %w", err)
	}

	apiMicroVM.Metadata["vendor-data"] = vendorData

	instanceData, err := s.createInstanceData()
	if err != nil {
		return fmt.Errorf("creating instance metadata: %w", err)
	}

	apiMicroVM.Metadata["meta-data"] = instanceData

	return nil
}

func (s *Service) createVendorData() (string, error) {
	// TODO: remove the boot command temporary fix after image-builder change #89
	vendorUserdata := &userdata.UserData{
		HostName:     s.scope.Name(),
		FinalMessage: "The Liquid Metal booted system is good to go after $UPTIME seconds",
		BootCommands: []string{
			"ln -sf /run/systemd/resolve/stub-resolv.conf /etc/resolv.conf",
		},
	}

	for _, key := range s.scope.SSHKeys() {
		user := userdata.User{
			Name:              key.User,
			SSHAuthorizedKeys: key.AuthorizedKeys,
		}

		vendorUserdata.Users = append(vendorUserdata.Users, user)
	}

	data, err := yaml.Marshal(vendorUserdata)
	if err != nil {
		return "", fmt.Errorf("marshalling bootstrap data: %w", err)
	}

	dataWithHeader := append([]byte(cloudInitHeader), data...)

	return base64.StdEncoding.EncodeToString(dataWithHeader), nil
}

func (s *Service) createInstanceData() (string, error) {
	userMetadata := instance.New(
		instance.WithLocalHostname(s.scope.Name()),
		instance.WithPlatform(platformLiquidMetal),
		instance.WithKeyValue("vm_host", s.hostID),
	)

	userMeta, err := yaml.Marshal(userMetadata)
	if err != nil {
		return "", fmt.Errorf("unable to marshal metadata: %w", err)
	}

	return base64.StdEncoding.EncodeToString(userMeta), nil
}
