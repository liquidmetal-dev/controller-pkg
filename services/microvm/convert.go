// Copyright 2022 Weaveworks or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MPL-2.0

package microvm

import (
	types "github.com/weaveworks-liquidmetal/controller-pkg/types/microvm"
	flintlocktypes "github.com/weaveworks-liquidmetal/flintlock/api/types"
)

const platformLiquidMetal = "liquid_metal"

func convertToFlintlockAPI(mvmScope Scope) *flintlocktypes.MicroVMSpec {
	mvmSpec := mvmScope.GetMicrovmSpec()

	apiVM := &flintlocktypes.MicroVMSpec{
		Id:         mvmScope.Name(),
		Namespace:  mvmScope.Namespace(),
		Labels:     mvmScope.GetLabels(),
		Vcpu:       int32(mvmSpec.VCPU),
		MemoryInMb: int32(mvmSpec.MemoryMb),
		Kernel: &flintlocktypes.Kernel{
			Image:            mvmSpec.Kernel.Image,
			Cmdline:          mvmSpec.KernelCmdLine,
			AddNetworkConfig: true,
			Filename:         &mvmSpec.Kernel.Filename,
		},
		RootVolume: &flintlocktypes.Volume{
			Id:         "root",
			IsReadOnly: mvmSpec.RootVolume.ReadOnly,
			Source: &flintlocktypes.VolumeSource{
				ContainerSource: &mvmSpec.RootVolume.Image,
			},
		},
		Metadata: map[string]string{},
	}

	if mvmSpec.Initrd != nil {
		apiVM.Initrd = &flintlocktypes.Initrd{
			Image:    mvmSpec.Initrd.Image,
			Filename: &mvmSpec.Initrd.Filename,
		}
	}

	apiVM.AdditionalVolumes = []*flintlocktypes.Volume{}

	for i := range mvmSpec.AdditionalVolumes {
		volume := mvmSpec.AdditionalVolumes[i]

		apiVM.AdditionalVolumes = append(apiVM.AdditionalVolumes, &flintlocktypes.Volume{
			Id:         volume.ID,
			IsReadOnly: volume.ReadOnly,
			Source: &flintlocktypes.VolumeSource{
				ContainerSource: &volume.Image,
			},
		})
	}

	apiVM.Interfaces = []*flintlocktypes.NetworkInterface{}

	for i := range mvmSpec.NetworkInterfaces {
		iface := mvmSpec.NetworkInterfaces[i]

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

		apiVM.Interfaces = append(apiVM.Interfaces, apiIface)
	}

	return apiVM
}
