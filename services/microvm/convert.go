// Copyright 2023 Weaveworks or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MPL-2.0

package microvm

import (
	types "github.com/weaveworks-liquidmetal/controller-pkg/types/microvm"
	flintlocktypes "github.com/weaveworks-liquidmetal/flintlock/api/types"
)

const platformLiquidMetal = "liquid_metal"

func convertToFlintlockAPI(mvmScope Scope) *flintlocktypes.MicroVMSpec {
	mvmSpec := mvmScope.GetMicrovmSpec()

	apiVM := newVM(
		withProvider(&mvmSpec.Provider),
		withNamespaceName(mvmScope.Name(), mvmScope.Namespace()),
		withLabels(mvmScope.GetLabels()),
		withResources(mvmSpec.VCPU, mvmSpec.MemoryMb),
		withInitRD(mvmSpec.Initrd),
		withNetworkInterfaces(mvmSpec.NetworkInterfaces),
		withAdditionalVolumes(mvmSpec.AdditionalVolumes),
		withKernel(mvmSpec.Kernel, mvmSpec.KernelCmdLine),
		withRootVolume(mvmSpec.RootVolume),
	)

	return apiVM
}

type specOption func(*flintlocktypes.MicroVMSpec)

func newVM(opts ...specOption) *flintlocktypes.MicroVMSpec {
	s := &flintlocktypes.MicroVMSpec{
		Interfaces:        []*flintlocktypes.NetworkInterface{},
		AdditionalVolumes: []*flintlocktypes.Volume{},
		Metadata:          map[string]string{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func withNamespaceName(name, namespace string) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		s.Id = name
		s.Namespace = namespace
	}
}

func withProvider(p *string) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		s.Provider = p
	}
}

func withLabels(labels map[string]string) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		s.Labels = labels
	}
}

func withResources(vcpu, mem int64) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		s.Vcpu = int32(vcpu)
		s.MemoryInMb = int32(mem)
	}
}

func withInitRD(initrd *types.ContainerFileSource) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		if initrd != nil {
			s.Initrd = &flintlocktypes.Initrd{
				Image:    initrd.Image,
				Filename: &initrd.Filename,
			}
		}
	}
}

func withNetworkInterfaces(interfaces []types.NetworkInterface) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		for i := range interfaces {
			iface := interfaces[i]

			apiIface := &flintlocktypes.NetworkInterface{
				DeviceId: iface.GuestDeviceName,
				GuestMac: &iface.GuestMAC,
			}

			if iface.Address != "" {
				apiIface.Address = &flintlocktypes.StaticAddress{
					Address: iface.Address,
				}
			}

			switch iface.Type {
			case types.IfaceTypeMacvtap:
				apiIface.Type = flintlocktypes.NetworkInterface_MACVTAP
			case types.IfaceTypeTap:
				apiIface.Type = flintlocktypes.NetworkInterface_TAP
			}

			s.Interfaces = append(s.Interfaces, apiIface)
		}
	}
}

func withAdditionalVolumes(volumes []types.Volume) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		for i := range volumes {
			volume := volumes[i]

			addVol := &flintlocktypes.Volume{
				Id:         volume.ID,
				IsReadOnly: volume.ReadOnly,
				Source: &flintlocktypes.VolumeSource{
					ContainerSource: &volume.Image,
				},
			}

			if volume.MountPoint != "" {
				addVol.MountPoint = &volume.MountPoint
			}

			s.AdditionalVolumes = append(s.AdditionalVolumes, addVol)
		}
	}
}

func withKernel(k types.ContainerFileSource, cmdLine map[string]string) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		s.Kernel = &flintlocktypes.Kernel{
			Image:            k.Image,
			Filename:         &k.Filename,
			Cmdline:          cmdLine,
			AddNetworkConfig: true,
		}
	}
}

func withRootVolume(rv types.Volume) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		s.RootVolume = &flintlocktypes.Volume{
			Id:         rv.ID,
			IsReadOnly: rv.ReadOnly,
			Source: &flintlocktypes.VolumeSource{
				ContainerSource: &rv.Image,
			},
		}
	}
}
