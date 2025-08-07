// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package omniapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/resource/protobuf"
	"github.com/cosi-project/runtime/pkg/resource/typed"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"

	// "google.golang.org/protobuf/types/known/emptypb"

	"github.com/siderolabs/omni/client/api/omni/specs"
	"github.com/siderolabs/omni/client/pkg/client"
	"github.com/siderolabs/omni/client/pkg/client/management"
	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/siderolabs/omni/client/pkg/template"
	"github.com/siderolabs/omni/client/pkg/template/operations"

	api_management "github.com/siderolabs/omni/client/api/omni/management"
)

type OmniClient struct {
	omniURL            string
	omniServiceAccount string
	omniClient         *client.Client
	context            context.Context
	state              state.State
}

func NewClient(url, sa string) *OmniClient {

	return &OmniClient{
		omniURL:            url,
		omniServiceAccount: sa,
	}
}

func (o *OmniClient) Open() error {
	client, err := client.New(o.omniURL, client.WithServiceAccount(o.omniServiceAccount))
	if err != nil {
		return err
	}
	o.omniClient = client
	o.context = context.Background()
	o.state = o.omniClient.Omni().State()
	return nil
}

func (o *OmniClient) GetMachines() (safe.List[*typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension]], error) {

	ctx := o.context
	st := o.state
	// Getting the resources from the Omni state.
	machines, err := safe.StateList[*omni.MachineStatus](ctx, st, omni.NewMachineStatus(resources.DefaultNamespace, "").Metadata())

	if err != nil {
		return machines, err
	}

	return machines, nil
}

func (o *OmniClient) GetClusters() (safe.List[*typed.Resource[protobuf.ResourceSpec[specs.ClusterStatusSpec, *specs.ClusterStatusSpec], omni.ClusterStatusExtension]], error) {
	ctx := o.context

	st := o.state
	// Getting the resources from the Omni state.
	clusters, err := safe.StateList[*omni.ClusterStatus](ctx, st, omni.NewClusterStatus(resources.DefaultNamespace, "").Metadata())

	if err != nil {
		return safe.List[*typed.Resource[protobuf.ResourceSpec[specs.ClusterStatusSpec, *specs.ClusterStatusSpec], omni.ClusterStatusExtension]]{}, err
	}
	return clusters, nil
}

func (o *OmniClient) SyncCluster(input io.Reader) error {
	ctx := o.context
	st := o.state

	err := operations.SyncTemplate(ctx, input, io.Discard, st, operations.SyncOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (o *OmniClient) SyncClusterAndWaitForReady(input io.Reader) error {
	ctx := o.context
	st := o.state

	buf := &bytes.Buffer{}
	tee := io.TeeReader(input, buf)

	err := operations.SyncTemplate(ctx, tee, io.Discard, st, operations.SyncOptions{})
	if err != nil {
		return err
	}

	t, err := template.Load(buf)
	if err != nil {
		return fmt.Errorf("template.Load : %v", err)
	}

	name, err := t.ClusterName()
	if err != nil {
		return fmt.Errorf("t.ClusterName() : %v", err)
	}

	// Waiting for the cluster to be Ready
	for {
		cluster, err := safe.StateGetByID[*omni.ClusterStatus](ctx, st, name)
		if err != nil {
			return err
		}
		if cluster.TypedSpec().Value.Ready {
			break
		}
	}

	return nil
}

func (o *OmniClient) DeleteCluster(name string) error {
	ctx := o.context
	st := o.state

	err := operations.DeleteCluster(ctx, name, io.Discard, st, operations.SyncOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (o *OmniClient) FindMachineByUuid(uuid string) (*typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension], bool) {
	machines, err := o.GetMachines()
	if err != nil {
		return nil, false
	}

	if machine, ok := machines.Find(func(r *typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension]) bool {
		return uuid != "" && uuid == r.Metadata().ID()
	}); ok {
		return machine, true
	}
	return nil, false
}

func (o *OmniClient) FindMachineByHardwareAddress(mac string) (*typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension], bool) {
	machines, err := o.GetMachines()
	if err != nil {
		return nil, false
	}

	if machine, ok := machines.Find(func(r *typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension]) bool {
		return mac != "" && r.TypedSpec().Value.Network != nil && len(r.TypedSpec().Value.Network.NetworkLinks) > 0 && mac == r.TypedSpec().Value.Network.NetworkLinks[0].HardwareAddress
	}); ok {
		return machine, true
	}
	return nil, false
}

func (o *OmniClient) GetClusterMachines(clustername string) (safe.List[*typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension]], error) {
	machines, err := o.GetMachines()
	if err != nil {
		return machines, err
	}

	return machines.FilterLabelQuery(resource.LabelEqual("omni.sidero.dev/cluster", clustername)), nil
}

func (o *OmniClient) GetClusterNameFromTemplate(r io.Reader) (string, error) {
	t, err := template.Load(r)
	if err != nil {
		return "", err
	}

	name, err := t.ClusterName()
	if err != nil {
		return "", err
	}

	return name, nil
}

func (o *OmniClient) GetKubeconfig(cluster string) (string, error) {
	ctx := o.context

	k, err := o.omniClient.Management().WithCluster(cluster).Kubeconfig(ctx)
	if err != nil {
		return "", err
	}

	return string(k), nil
}

func (o *OmniClient) GetKubeconfigWithoutOIDC(cluster, user string, groups ...string) (string, error) {
	ctx := o.context

	if user == "" {
		user = "admin"
	}

	if groups == nil {
		groups = []string{"system:masters"}
	}

	k, err := o.omniClient.Management().WithCluster(cluster).Kubeconfig(ctx, management.WithServiceAccount(365*24*time.Hour, user, groups...))
	if err != nil {
		return "", err
	}

	return string(k), nil
}

func (o *OmniClient) GetTalosconfigWithBreakGlass(cluster string) (string, error) {
	ctx := o.context

	k, err := o.omniClient.Management().WithCluster(cluster).Talosconfig(ctx, management.WithBreakGlassTalosconfig(true))
	if err != nil {
		return "", err
	}

	return string(k), nil
}

func (o *OmniClient) GetTalosconfig(cluster string) (string, error) {
	ctx := o.context

	k, err := o.omniClient.Management().WithCluster(cluster).Talosconfig(ctx, management.WithRawTalosconfig(false))
	if err != nil {
		return "", err
	}

	return string(k), nil
}

func (o *OmniClient) GetTalosconfigWithRawTalosconfig(cluster string) (string, error) {
	ctx := o.context

	k, err := o.omniClient.Management().WithCluster(cluster).Talosconfig(ctx, management.WithRawTalosconfig(true))
	if err != nil {
		return "", err
	}

	return string(k), nil
}

func (o *OmniClient) GetTemplateFromClusterName(cluster string) (string, error) {
	buf := &bytes.Buffer{}

	_, err := operations.ExportTemplate(o.context, o.state, cluster, buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (o *OmniClient) SyncManifests(cluster string) error {
	ctx := o.context
	return o.omniClient.Management().WithCluster(cluster).KubernetesSyncManifests(ctx, false,
		func(resp *api_management.KubernetesSyncManifestResponse) error {
			switch resp.ResponseType {
			case api_management.KubernetesSyncManifestResponse_UNKNOWN:
			case api_management.KubernetesSyncManifestResponse_MANIFEST:
				log.Printf("[INFO] > processing manifest %s\n", resp.Path)

				switch {
				case resp.Skipped:
					log.Println("[INFO] < no changes")
				default:
					log.Println(resp.Diff)
					log.Println("[INFO] < applied successfully")
				}
			case api_management.KubernetesSyncManifestResponse_ROLLOUT:
				log.Printf("[INFO] > waiting for %s\n", resp.Path)
			}
			return nil
		})
}

func (o *OmniClient) DeleteClusterMachines(machines safe.List[*typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension]]) error {
	ctx := o.context
	st := o.state

	err := machines.ForEachErr(func(r *typed.Resource[protobuf.ResourceSpec[specs.MachineStatusSpec, *specs.MachineStatusSpec], omni.MachineStatusExtension]) error {
		//destroyReady, err := st.Teardown(o.context, r.Metadata())
		destroyReady, err := st.Teardown(o.context, resource.NewMetadata(r.Metadata().Namespace(), "Links.omni.sidero.dev", r.Metadata().ID(), r.Metadata().Version()))
		if err != nil {
			return err
		}
		if destroyReady {
			if err = st.Destroy(ctx, resource.NewMetadata(r.Metadata().Namespace(), "Links.omni.sidero.dev", r.Metadata().ID(), r.Metadata().Version())); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
