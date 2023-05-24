// Copyright 2023 Weaveworks or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MPL-2.0

package microvm

import (
	"fmt"

	types "github.com/weaveworks-liquidmetal/controller-pkg/types/microvm"
	flintlocktypes "github.com/weaveworks-liquidmetal/flintlock/api/types"
	"k8s.io/utils/pointer"
)

const (
	platformLiquidMetal        = "liquid_metal"
	rootVolumeID               = "root"
	modulesVolumeID            = "modules"
	modulesMountPoint          = "/lib/modules/%s"
	kernelFilename             = "boot/vmlinux"
	defaultOsImagePath         = "ghcr.io/weaveworks-liquidmetal/capmvm-k8s-os:%s"
	defaultKernelBinImagePath  = "ghcr.io/weaveworks-liquidmetal/firecracker-kernel-bin:%s"
	defaultKernelModsImagePath = "ghcr.io/weaveworks-liquidmetal/firecracker-kernel-modules:%s"
)

func convertToFlintlockAPI(mvmScope Scope) *flintlocktypes.MicroVMSpec {
	mvmSpec := mvmScope.GetMicrovmSpec()

	apiVM := newVM(
		withNamespaceName(mvmScope.Name(), mvmScope.Namespace()),
		withLabels(mvmScope.GetLabels()),
		withResources(mvmSpec.VCPU, mvmSpec.MemoryMb),
		withInitRD(mvmSpec.Initrd),
		withNetworkInterfaces(mvmSpec.NetworkInterfaces),
		withAdditionalVolumes(mvmSpec.AdditionalVolumes),
		withKernel(mvmSpec.Kernel, mvmSpec.KernelCmdLine, mvmSpec.KernelVersion),
		withRootVolume(mvmSpec.RootVolume, mvmSpec.OsVersion),
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

func withKernel(k types.ContainerFileSource, cmd map[string]string, version string) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		s.Kernel = &flintlocktypes.Kernel{
			Cmdline:          cmd,
			AddNetworkConfig: true,
		}

		if kernelImageConfigured(k) {
			s.Kernel.Image = k.Image
			s.Kernel.Filename = &k.Filename
		}

		// If a user has set both the image and the version, the version wins so
		// here we override some stuff.
		// Switch these round if we want to change the override.
		if isSet(version) {
			s.Kernel.Image = kernelBinImage(version)
			s.Kernel.Filename = pointer.String(kernelFilename)

			addVol := &flintlocktypes.Volume{
				Id:         modulesVolumeID,
				IsReadOnly: false,
				Source: &flintlocktypes.VolumeSource{
					ContainerSource: kernelModsImage(version),
				},
				MountPoint: kernelModsMount(version),
			}

			s.AdditionalVolumes = append(s.AdditionalVolumes, addVol)
		}
	}
}

func withRootVolume(rv types.Volume, os string) specOption {
	return func(s *flintlocktypes.MicroVMSpec) {
		if rootVolumeConfigured(rv) {
			s.RootVolume = &flintlocktypes.Volume{
				Id:         rv.ID,
				IsReadOnly: rv.ReadOnly,
				Source: &flintlocktypes.VolumeSource{
					ContainerSource: &rv.Image,
				},
			}
		}

		// If a user has set both the image and the os version, the os version wins
		// so here we override some stuff.
		// Switch these round if we want to change the override.
		if isSet(os) {
			s.RootVolume = &flintlocktypes.Volume{
				Id:         rootVolumeID,
				IsReadOnly: false,
				Source: &flintlocktypes.VolumeSource{
					ContainerSource: rootImage(os),
				},
			}
		}
	}
}

func rootVolumeConfigured(vol types.Volume) bool {
	return vol.Image != "" && vol.ID != ""
}

func kernelImageConfigured(k types.ContainerFileSource) bool {
	return k.Image != "" && k.Filename != ""
}

func isSet(value string) bool {
	return value != ""
}

func rootImage(version string) *string {
	return pointer.String(fmt.Sprintf(defaultOsImagePath, version))
}

func kernelBinImage(version string) string {
	return fmt.Sprintf(defaultKernelBinImagePath, version)
}

func kernelModsImage(version string) *string {
	return pointer.String(fmt.Sprintf(defaultKernelModsImagePath, version))
}

func kernelModsMount(version string) *string {
	return pointer.String(fmt.Sprintf(modulesMountPoint, version))
}
