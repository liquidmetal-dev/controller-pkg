package microvm

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"

	"github.com/weaveworks-liquidmetal/controller-pkg/services/microvm/fakes"
	"github.com/weaveworks-liquidmetal/controller-pkg/types/microvm"
	flintlocktypes "github.com/weaveworks-liquidmetal/flintlock/api/types"
)

func Test_convertToFlintlockAPI(t *testing.T) {
	g := NewWithT(t)

	var (
		machineName = "foo"
		namespace   = "baz"

		keyVal1 = "key1"

		strVal1 = "value1"
		strVal2 = "value2"
		strVal3 = "value3"
		strVal4 = "value4"

		intVal1 int64 = 1
		intVal2 int64 = 2
	)

	tt := []struct {
		name     string
		input    microvm.VMSpec
		expected func(*WithT, *flintlocktypes.MicroVMSpec)
	}{
		{
			name:  "withNamespaceName",
			input: microvm.VMSpec{},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Id).To(Equal(machineName))
				g.Expect(converted.Namespace).To(Equal(namespace))
			},
		},
		{
			name:  "withLabels",
			input: microvm.VMSpec{},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Labels).To(HaveKeyWithValue(keyVal1, strVal1))
			},
		},
		{
			name:  "withResources",
			input: microvm.VMSpec{VCPU: intVal1, MemoryMb: intVal2},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Vcpu).To(Equal(int32(intVal1)))
				g.Expect(converted.MemoryInMb).To(Equal(int32(intVal2)))
			},
		},
		{
			name: "withInitRD",
			input: microvm.VMSpec{Initrd: &microvm.ContainerFileSource{
				Image:    strVal1,
				Filename: strVal2,
			}},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Initrd.Image).To(Equal(strVal1))
				g.Expect(*converted.Initrd.Filename).To(Equal(strVal2))
			},
		},
		{
			name: "withNetworkInterfaces",
			input: microvm.VMSpec{NetworkInterfaces: []microvm.NetworkInterface{
				{
					GuestDeviceName: strVal1,
					GuestMAC:        strVal2,
					Type:            microvm.IfaceTypeMacvtap,
				},
				{
					GuestDeviceName: strVal3,
					GuestMAC:        strVal4,
					Type:            microvm.IfaceTypeTap,
				},
			}},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Interfaces).To(HaveLen(2))

				g.Expect(converted.Interfaces[0].DeviceId).To(Equal(strVal1))
				g.Expect(*converted.Interfaces[0].GuestMac).To(Equal(strVal2))
				g.Expect(converted.Interfaces[0].Type).To(Equal(flintlocktypes.NetworkInterface_MACVTAP))
				g.Expect(converted.Interfaces[1].DeviceId).To(Equal(strVal3))
				g.Expect(*converted.Interfaces[1].GuestMac).To(Equal(strVal4))
				g.Expect(converted.Interfaces[1].Type).To(Equal(flintlocktypes.NetworkInterface_TAP))
			},
		},
		{
			name: "withNetworkInterfaces, has address",
			input: microvm.VMSpec{NetworkInterfaces: []microvm.NetworkInterface{{
				GuestDeviceName: strVal1,
				GuestMAC:        strVal2,
				Type:            microvm.IfaceTypeMacvtap,
				Address:         strVal3,
			}}},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Interfaces).To(HaveLen(1))

				g.Expect(converted.Interfaces[0].DeviceId).To(Equal(strVal1))
				g.Expect(*converted.Interfaces[0].GuestMac).To(Equal(strVal2))
				g.Expect(converted.Interfaces[0].Type).To(Equal(flintlocktypes.NetworkInterface_MACVTAP))
				g.Expect(converted.Interfaces[0].Address.Address).To(Equal(strVal3))
			},
		},
		{
			name: "withAdditionalVolumes",
			input: microvm.VMSpec{
				AdditionalVolumes: []microvm.Volume{
					{
						ID:       strVal1,
						Image:    strVal2,
						ReadOnly: false,
					},
					{
						ID:       strVal3,
						Image:    strVal4,
						ReadOnly: true,
					},
				},
			},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.AdditionalVolumes).To(HaveLen(2))

				g.Expect(converted.AdditionalVolumes[0].Id).To(Equal(strVal1))
				g.Expect(*converted.AdditionalVolumes[0].Source.ContainerSource).To(Equal(strVal2))
				g.Expect(converted.AdditionalVolumes[0].IsReadOnly).To(BeFalse())
				g.Expect(converted.AdditionalVolumes[1].Id).To(Equal(strVal3))
				g.Expect(*converted.AdditionalVolumes[1].Source.ContainerSource).To(Equal(strVal4))
				g.Expect(converted.AdditionalVolumes[1].IsReadOnly).To(BeTrue())
			},
		},
		{
			name: "withAdditionalVolumes, has mountpoint",
			input: microvm.VMSpec{
				AdditionalVolumes: []microvm.Volume{{
					ID:         strVal1,
					Image:      strVal2,
					ReadOnly:   false,
					MountPoint: strVal3,
				}},
			},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.AdditionalVolumes).To(HaveLen(1))

				g.Expect(converted.AdditionalVolumes[0].Id).To(Equal(strVal1))
				g.Expect(*converted.AdditionalVolumes[0].Source.ContainerSource).To(Equal(strVal2))
				g.Expect(converted.AdditionalVolumes[0].IsReadOnly).To(BeFalse())
				g.Expect(converted.AdditionalVolumes[0].MountPoint).To(Equal(pointer.String(strVal3)))
			},
		},
		{
			name: "withKernel, image and filename configured",
			input: microvm.VMSpec{Kernel: microvm.ContainerFileSource{
				Image:    strVal1,
				Filename: strVal2,
			},
				KernelCmdLine: map[string]string{strVal3: strVal4},
			},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Kernel.Image).To(Equal(strVal1))
				g.Expect(*converted.Kernel.Filename).To(Equal(strVal2))
				g.Expect(converted.Kernel.AddNetworkConfig).To(BeTrue())
				g.Expect(converted.Kernel.Cmdline).To(HaveKeyWithValue(strVal3, strVal4))
				g.Expect(converted.AdditionalVolumes).To(BeEmpty())
			},
		},
		{
			name: "withKernel, version configured",
			input: microvm.VMSpec{KernelVersion: strVal1,
				KernelCmdLine: map[string]string{strVal2: strVal3},
			},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.Kernel.Image).To(Equal(fmt.Sprintf(defaultKernelBinImagePath, strVal1)))
				g.Expect(*converted.Kernel.Filename).To(Equal(kernelFilename))
				g.Expect(converted.Kernel.AddNetworkConfig).To(BeTrue())
				g.Expect(converted.Kernel.Cmdline).To(HaveKeyWithValue(strVal2, strVal3))

				g.Expect(converted.AdditionalVolumes).To(HaveLen(1))
				g.Expect(converted.AdditionalVolumes[0].Id).To(Equal(modulesVolumeID))
				g.Expect(*converted.AdditionalVolumes[0].Source.ContainerSource).To(Equal(fmt.Sprintf(defaultKernelModsImagePath, strVal1)))
				g.Expect(converted.AdditionalVolumes[0].IsReadOnly).To(BeFalse())
				g.Expect(converted.AdditionalVolumes[0].MountPoint).To(Equal(pointer.String(fmt.Sprintf(modulesMountPoint, strVal1))))
			},
		},
		{
			name: "withRootVolume, image and id configured",
			input: microvm.VMSpec{RootVolume: microvm.Volume{
				ID:       strVal1,
				Image:    strVal2,
				ReadOnly: false,
			}},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.RootVolume.Id).To(Equal(strVal1))
				g.Expect(*converted.RootVolume.Source.ContainerSource).To(Equal(strVal2))
				g.Expect(converted.RootVolume.IsReadOnly).To(BeFalse())
			},
		},
		{
			name:  "withRootVolume, os version configured",
			input: microvm.VMSpec{OsVersion: strVal1},
			expected: func(g *WithT, converted *flintlocktypes.MicroVMSpec) {
				g.Expect(converted.RootVolume.Id).To(Equal(rootVolumeID))
				g.Expect(*converted.RootVolume.Source.ContainerSource).To(Equal(fmt.Sprintf(defaultOsImagePath, strVal1)))
				g.Expect(converted.RootVolume.IsReadOnly).To(BeFalse())
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var mScope *fakes.FakeScope
			mScope = new(fakes.FakeScope)
			mScope.GetMicrovmSpecReturns(tc.input)
			mScope.NameReturns(machineName)
			mScope.NamespaceReturns(namespace)
			mScope.GetLabelsReturns(map[string]string{keyVal1: strVal1})

			convertedVM := convertToFlintlockAPI(mScope)

			tc.expected(g, convertedVM)
		})
	}
}
