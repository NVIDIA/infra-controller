/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/gogo/status"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	flowv1 "github.com/NVIDIA/infra-controller-rest/workflow-schema/flow/protobuf/v1"
	wflows "github.com/NVIDIA/infra-controller-rest/workflow-schema/schema/site-agent/workflows/v1"
)

var runes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Add utlity methods here
// randSeq generates a random sequence of runes
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

// generateSiteVersion generates a version in the format of "V1-T<timestamp>"
func generateSiteVersion() string {
	// Get the current time
	now := time.Now()
	// Get microseconds since epoch
	microseconds := now.UnixMicro()
	return fmt.Sprintf("V1-T%d", microseconds)
}

// incrementMAC takes a hardware address (MAC address) and increments it by one.
// It handles carrying over to the next byte when a byte overflows (reaches 255).
func incrementMAC(mac net.HardwareAddr) {
	// Iterate from the last byte to the first.
	for i := len(mac) - 1; i >= 0; i-- {
		// Increment the current byte.
		mac[i]++
		// If the byte is not 0, it means there was no overflow, so we can stop.
		if mac[i] != 0 {
			break
		}
		// If the byte is 0, it means it overflowed from 255, so we continue to the next
		// byte to handle the "carry-over".
	}
}

// MockNICoClient is a mock implementation of NICo gRPC protobuf Client
type MockNICoClient struct {
	wflows.ForgeClient
}

/* Version mock methods */
func (c *MockNICoClient) Version(ctx context.Context, in *wflows.VersionRequest, opts ...grpc.CallOption) (*wflows.BuildInfo, error) {
	out := new(wflows.BuildInfo)
	out.BuildVersion = "1.0.0"
	return out, nil
}

/* VPC mock methods */
func (c *MockNICoClient) CreateVpc(ctx context.Context, in *wflows.VpcCreationRequest, opts ...grpc.CallOption) (*wflows.Vpc, error) {
	out := new(wflows.Vpc)
	out.Id = &wflows.VpcId{Value: uuid.NewString()}
	return out, nil
}

func (c *MockNICoClient) UpdateVpc(ctx context.Context, in *wflows.VpcUpdateRequest, opts ...grpc.CallOption) (*wflows.VpcUpdateResult, error) {
	out := new(wflows.VpcUpdateResult)
	return out, nil
}

func (c *MockNICoClient) UpdateVpcVirtualization(ctx context.Context, in *wflows.VpcUpdateVirtualizationRequest, opts ...grpc.CallOption) (*wflows.VpcUpdateVirtualizationResult, error) {
	out := new(wflows.VpcUpdateVirtualizationResult)
	return out, nil
}

func (c *MockNICoClient) DeleteVpc(ctx context.Context, in *wflows.VpcDeletionRequest, opts ...grpc.CallOption) (*wflows.VpcDeletionResult, error) {
	out := new(wflows.VpcDeletionResult)
	return out, nil
}

func (c *MockNICoClient) FindVpcIds(ctx context.Context, in *wflows.VpcSearchFilter, opts ...grpc.CallOption) (*wflows.VpcIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve vpc ids")
	}

	out := &wflows.VpcIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.VpcIds = append(out.VpcIds, &wflows.VpcId{Value: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindVpcsByIds(ctx context.Context, in *wflows.VpcsByIdsRequest, opts ...grpc.CallOption) (*wflows.VpcList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve vpcs")
	}

	out := &wflows.VpcList{}
	if in != nil {
		for _, id := range in.VpcIds {
			out.Vpcs = append(out.Vpcs, &wflows.Vpc{
				Id: id,
			})
		}
	}

	return out, nil
}

/* Network Segment mock methods */

func (c *MockNICoClient) CreateNetworkSegment(ctx context.Context, in *wflows.NetworkSegmentCreationRequest, opts ...grpc.CallOption) (*wflows.NetworkSegment, error) {
	out := new(wflows.NetworkSegment)
	out.Id = &wflows.NetworkSegmentId{Value: uuid.NewString()}
	return out, nil
}

func (c *MockNICoClient) DeleteNetworkSegment(ctx context.Context, in *wflows.NetworkSegmentDeletionRequest, opts ...grpc.CallOption) (*wflows.NetworkSegmentDeletionResult, error) {
	out := new(wflows.NetworkSegmentDeletionResult)
	return out, nil
}

func (c *MockNICoClient) FindNetworkSegmentIds(ctx context.Context, in *wflows.NetworkSegmentSearchFilter, opts ...grpc.CallOption) (*wflows.NetworkSegmentIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve network segment ids")
	}

	out := &wflows.NetworkSegmentIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.NetworkSegmentsIds = append(out.NetworkSegmentsIds, &wflows.NetworkSegmentId{Value: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindNetworkSegmentsByIds(ctx context.Context, in *wflows.NetworkSegmentsByIdsRequest, opts ...grpc.CallOption) (*wflows.NetworkSegmentList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve network segments")
	}

	out := &wflows.NetworkSegmentList{}
	if in != nil {
		for _, id := range in.NetworkSegmentsIds {
			out.NetworkSegments = append(out.NetworkSegments, &wflows.NetworkSegment{
				Id: id,
			})
		}
	}

	return out, nil
}

/* InfiniBand Partition mock methods */
func (c *MockNICoClient) CreateIBPartition(ctx context.Context, in *wflows.IBPartitionCreationRequest, opts ...grpc.CallOption) (*wflows.IBPartition, error) {
	out := new(wflows.IBPartition)
	out.Id = &wflows.IBPartitionId{Value: uuid.NewString()}
	return out, nil
}

func (c *MockNICoClient) UpdateIBPartition(ctx context.Context, in *wflows.IBPartitionUpdateRequest, opts ...grpc.CallOption) (*wflows.IBPartition, error) {
	out := new(wflows.IBPartition)
	if in != nil && in.Id != nil {
		out.Id = in.Id
	} else {
		out.Id = &wflows.IBPartitionId{Value: uuid.NewString()}
	}
	if in != nil {
		out.Config = in.GetConfig()
		out.Metadata = in.GetMetadata()
	}
	return out, nil
}

func (c *MockNICoClient) DeleteIBPartition(ctx context.Context, in *wflows.IBPartitionDeletionRequest, opts ...grpc.CallOption) (*wflows.IBPartitionDeletionResult, error) {
	out := new(wflows.IBPartitionDeletionResult)
	return out, nil
}

func (c *MockNICoClient) FindIBPartitionIds(ctx context.Context, in *wflows.IBPartitionSearchFilter, opts ...grpc.CallOption) (*wflows.IBPartitionIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve ib partition ids")
	}

	out := &wflows.IBPartitionIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.IbPartitionIds = append(out.IbPartitionIds, &wflows.IBPartitionId{Value: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindIBPartitionsByIds(ctx context.Context, in *wflows.IBPartitionsByIdsRequest, opts ...grpc.CallOption) (*wflows.IBPartitionList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve ib partitions")
	}

	out := &wflows.IBPartitionList{}
	if in != nil {
		for _, id := range in.IbPartitionIds {
			out.IbPartitions = append(out.IbPartitions, &wflows.IBPartition{
				Id: id,
			})
		}
	}

	return out, nil
}

/* Instance mock methods */
func (c *MockNICoClient) AllocateInstance(ctx context.Context, in *wflows.InstanceAllocationRequest, opts ...grpc.CallOption) (*wflows.Instance, error) {
	out := new(wflows.Instance)
	out.Id = &wflows.InstanceId{Value: uuid.NewString()}
	return out, nil
}

func (c *MockNICoClient) AllocateInstances(ctx context.Context, in *wflows.BatchInstanceAllocationRequest, opts ...grpc.CallOption) (*wflows.BatchInstanceAllocationResponse, error) {
	out := &wflows.BatchInstanceAllocationResponse{
		Instances: make([]*wflows.Instance, len(in.InstanceRequests)),
	}
	for i := range in.InstanceRequests {
		out.Instances[i] = &wflows.Instance{
			Id: &wflows.InstanceId{Value: uuid.NewString()},
		}
	}
	return out, nil
}

func (c *MockNICoClient) UpdateInstanceConfig(ctx context.Context, in *wflows.InstanceConfigUpdateRequest, opts ...grpc.CallOption) (*wflows.Instance, error) {
	out := new(wflows.Instance)
	out.Id = in.InstanceId
	out.Metadata = in.Metadata
	out.Config = in.Config
	return out, nil
}

func (c *MockNICoClient) ReleaseInstance(ctx context.Context, in *wflows.InstanceReleaseRequest, opts ...grpc.CallOption) (*wflows.InstanceReleaseResult, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.NotFound {
			return nil, status.Error(codes.NotFound, "instance not found: ")
		}
	}
	out := new(wflows.InstanceReleaseResult)
	return out, nil
}

func (c *MockNICoClient) FindInstanceIds(ctx context.Context, in *wflows.InstanceSearchFilter, opts ...grpc.CallOption) (*wflows.InstanceIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve instance ids")
	}

	out := &wflows.InstanceIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.InstanceIds = append(out.InstanceIds, &wflows.InstanceId{Value: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindInstancesByIds(ctx context.Context, in *wflows.InstancesByIdsRequest, opts ...grpc.CallOption) (*wflows.InstanceList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve instances")
	}

	out := &wflows.InstanceList{}
	if in != nil {
		for _, id := range in.InstanceIds {
			out.Instances = append(out.Instances, &wflows.Instance{
				Id: id,
			})
		}
	}

	return out, nil
}

func (c *MockNICoClient) InvokeInstancePower(ctx context.Context, in *wflows.InstancePowerRequest, opts ...grpc.CallOption) (*wflows.InstancePowerResult, error) {
	out := new(wflows.InstancePowerResult)
	return out, nil
}

/* Machine mock methods */
func (c *MockNICoClient) SetMaintenance(ctx context.Context, in *wflows.MaintenanceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) UpdateMachineMetadata(ctx context.Context, in *wflows.MachineMetadataUpdateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to update machine metadata")
	}

	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) FindMachineIds(ctx context.Context, in *wflows.MachineSearchConfig, opts ...grpc.CallOption) (*wflows.MachineIdList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve machine ids")
		}
	}

	out := &wflows.MachineIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.MachineIds = append(out.MachineIds, &wflows.MachineId{Id: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindMachinesByIds(ctx context.Context, in *wflows.MachinesByIdsRequest, opts ...grpc.CallOption) (*wflows.MachineList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve machines by ids")
		}
	}

	out := &wflows.MachineList{}
	if in != nil {
		for _, id := range in.MachineIds {
			out.Machines = append(out.Machines, &wflows.Machine{
				Id:    id,
				State: "Ready",
			})
		}
	}

	return out, nil
}

/* Tenant Keyset mock methods */
func (c *MockNICoClient) CreateTenantKeyset(ctx context.Context, in *wflows.CreateTenantKeysetRequest, opts ...grpc.CallOption) (*wflows.CreateTenantKeysetResponse, error) {
	out := new(wflows.CreateTenantKeysetResponse)
	out.Keyset = &wflows.TenantKeyset{
		KeysetIdentifier: &wflows.TenantKeysetIdentifier{
			OrganizationId: in.KeysetIdentifier.OrganizationId,
			KeysetId:       uuid.NewString(),
		},
	}
	out.Keyset.KeysetContent = in.KeysetContent
	out.Keyset.Version = in.Version
	return out, nil
}

func (c *MockNICoClient) UpdateTenantKeyset(ctx context.Context, in *wflows.UpdateTenantKeysetRequest, opts ...grpc.CallOption) (*wflows.UpdateTenantKeysetResponse, error) {
	out := new(wflows.UpdateTenantKeysetResponse)
	return out, nil
}

func (c *MockNICoClient) DeleteTenantKeyset(ctx context.Context, in *wflows.DeleteTenantKeysetRequest, opts ...grpc.CallOption) (*wflows.DeleteTenantKeysetResponse, error) {
	out := new(wflows.DeleteTenantKeysetResponse)
	return out, nil
}

func (c *MockNICoClient) FindTenantKeysetIds(ctx context.Context, in *wflows.TenantKeysetSearchFilter, opts ...grpc.CallOption) (*wflows.TenantKeysetIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve tenant keyset ids")
	}

	out := &wflows.TenantKeysetIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		orgID := uuid.NewString()
		for i := 0; i < count; i++ {
			out.KeysetIds = append(out.KeysetIds, &wflows.TenantKeysetIdentifier{OrganizationId: orgID, KeysetId: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindTenantKeysetsByIds(ctx context.Context, in *wflows.TenantKeysetsByIdsRequest, opts ...grpc.CallOption) (*wflows.TenantKeySetList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve tenant keysets")
	}

	out := &wflows.TenantKeySetList{}
	if in != nil {
		for _, id := range in.KeysetIds {
			out.Keyset = append(out.Keyset, &wflows.TenantKeyset{
				KeysetIdentifier: &wflows.TenantKeysetIdentifier{
					OrganizationId: id.OrganizationId,
					KeysetId:       id.KeysetId,
				},
			})
		}
	}

	return out, nil
}

/* OS Image mock methods */
func (c *MockNICoClient) CreateOsImage(ctx context.Context, in *wflows.OsImageAttributes, opts ...grpc.CallOption) (*wflows.OsImage, error) {
	out := new(wflows.OsImage)
	out.Attributes = &wflows.OsImageAttributes{Id: &wflows.UUID{Value: uuid.NewString()}}
	return out, nil
}

func (c *MockNICoClient) UpdateOsImage(ctx context.Context, in *wflows.OsImageAttributes, opts ...grpc.CallOption) (*wflows.OsImage, error) {
	out := new(wflows.OsImage)
	return out, nil
}

func (c *MockNICoClient) DeleteOsImage(ctx context.Context, in *wflows.DeleteOsImageRequest, opts ...grpc.CallOption) (*wflows.DeleteOsImageResponse, error) {
	out := new(wflows.DeleteOsImageResponse)
	return out, nil
}

func (c *MockNICoClient) ListOsImage(ctx context.Context, in *wflows.ListOsImageRequest, opts ...grpc.CallOption) (*wflows.ListOsImageResponse, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve os image list")
	}

	out := &wflows.ListOsImageResponse{}
	count, ok := ctx.Value("wantCount").(int)
	if ok {
		id := uuid.NewString()
		for i := 0; i < count; i++ {
			out.Images = append(out.Images, &wflows.OsImage{Attributes: &wflows.OsImageAttributes{Id: &wflows.UUID{Value: id}}})
		}
	}
	return out, nil
}

/* Tenant mock methods */
func (c *MockNICoClient) CreateTenant(ctx context.Context, in *wflows.CreateTenantRequest, opts ...grpc.CallOption) (*wflows.CreateTenantResponse, error) {
	out := new(wflows.CreateTenantResponse)
	out.Tenant = &wflows.Tenant{
		OrganizationId: in.OrganizationId,
	}
	if in.Metadata != nil {
		out.Tenant.Metadata = &wflows.Metadata{
			Name: in.Metadata.Name,
		}
	}
	return out, nil
}

func (c *MockNICoClient) FindTenant(ctx context.Context, in *wflows.FindTenantRequest, opts ...grpc.CallOption) (*wflows.FindTenantResponse, error) {
	out := new(wflows.FindTenantResponse)
	return out, nil
}

func (c *MockNICoClient) UpdateTenant(ctx context.Context, in *wflows.UpdateTenantRequest, opts ...grpc.CallOption) (*wflows.UpdateTenantResponse, error) {
	out := new(wflows.UpdateTenantResponse)
	return out, nil
}

func (c *MockNICoClient) FindTenantOrganizationIds(ctx context.Context, in *wflows.TenantSearchFilter, opts ...grpc.CallOption) (*wflows.TenantOrganizationIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve Tenant organization ids")
	}

	out := &wflows.TenantOrganizationIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.TenantOrganizationIds = append(out.TenantOrganizationIds, randSeq(10))
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindTenantsByOrganizationIds(ctx context.Context, in *wflows.TenantByOrganizationIdsRequest, opts ...grpc.CallOption) (*wflows.TenantList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve Tenants")
	}

	out := &wflows.TenantList{}
	if in != nil {
		for _, id := range in.OrganizationIds {
			out.Tenants = append(out.Tenants, &wflows.Tenant{
				OrganizationId: id,
			})
		}
	}

	return out, nil
}

/* Instance Type mock methods */
func (c *MockNICoClient) CreateInstanceType(ctx context.Context, in *wflows.CreateInstanceTypeRequest, opts ...grpc.CallOption) (*wflows.CreateInstanceTypeResponse, error) {
	out := &wflows.CreateInstanceTypeResponse{InstanceType: &wflows.InstanceType{}}
	out.InstanceType.Id = uuid.NewString()
	return out, nil
}

func (c *MockNICoClient) UpdateInstanceType(ctx context.Context, in *wflows.UpdateInstanceTypeRequest, opts ...grpc.CallOption) (*wflows.UpdateInstanceTypeResponse, error) {
	out := &wflows.UpdateInstanceTypeResponse{InstanceType: &wflows.InstanceType{}}
	out.InstanceType.Id = in.Id
	out.InstanceType.Metadata = in.Metadata
	out.InstanceType.Attributes = in.InstanceTypeAttributes
	return out, nil
}

func (c *MockNICoClient) DeleteInstanceType(ctx context.Context, in *wflows.DeleteInstanceTypeRequest, opts ...grpc.CallOption) (*wflows.DeleteInstanceTypeResponse, error) {
	out := &wflows.DeleteInstanceTypeResponse{}
	return out, nil
}

func (c *MockNICoClient) AssociateMachinesWithInstanceType(ctx context.Context, in *wflows.AssociateMachinesWithInstanceTypeRequest, opts ...grpc.CallOption) (*wflows.AssociateMachinesWithInstanceTypeResponse, error) {
	out := &wflows.AssociateMachinesWithInstanceTypeResponse{}
	return out, nil
}

func (c *MockNICoClient) RemoveMachineInstanceTypeAssociation(ctx context.Context, in *wflows.RemoveMachineInstanceTypeAssociationRequest, opts ...grpc.CallOption) (*wflows.RemoveMachineInstanceTypeAssociationResponse, error) {
	out := &wflows.RemoveMachineInstanceTypeAssociationResponse{}
	return out, nil
}

func (c *MockNICoClient) FindInstanceTypeIds(ctx context.Context, in *wflows.FindInstanceTypeIdsRequest, opts ...grpc.CallOption) (*wflows.FindInstanceTypeIdsResponse, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve InstanceType ids")
	}

	out := &wflows.FindInstanceTypeIdsResponse{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.InstanceTypeIds = append(out.InstanceTypeIds, randSeq(10))
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindInstanceTypesByIds(ctx context.Context, in *wflows.FindInstanceTypesByIdsRequest, opts ...grpc.CallOption) (*wflows.FindInstanceTypesByIdsResponse, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve InstanceTypes")
	}

	out := &wflows.FindInstanceTypesByIdsResponse{}
	if in != nil {
		for _, id := range in.InstanceTypeIds {
			out.InstanceTypes = append(out.InstanceTypes, &wflows.InstanceType{
				Id: id,
			})
		}
	}
	return out, nil
}

/* VPC Prefix mock methods */
func (c *MockNICoClient) CreateVpcPrefix(ctx context.Context, in *wflows.VpcPrefixCreationRequest, opts ...grpc.CallOption) (*wflows.VpcPrefix, error) {
	out := new(wflows.VpcPrefix)
	return out, nil
}

func (c *MockNICoClient) UpdateVpcPrefix(ctx context.Context, in *wflows.VpcPrefixUpdateRequest, opts ...grpc.CallOption) (*wflows.VpcPrefix, error) {
	out := new(wflows.VpcPrefix)
	return out, nil
}

func (c *MockNICoClient) DeleteVpcPrefix(ctx context.Context, in *wflows.VpcPrefixDeletionRequest, opts ...grpc.CallOption) (*wflows.VpcPrefixDeletionResult, error) {
	out := new(wflows.VpcPrefixDeletionResult)
	return out, nil
}

func (c *MockNICoClient) SearchVpcPrefixes(ctx context.Context, in *wflows.VpcPrefixSearchQuery, opts ...grpc.CallOption) (*wflows.VpcPrefixIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve vpcprefix ids")
	}

	out := &wflows.VpcPrefixIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.VpcPrefixIds = append(out.VpcPrefixIds, &wflows.VpcPrefixId{Value: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) GetVpcPrefixes(ctx context.Context, in *wflows.VpcPrefixGetRequest, opts ...grpc.CallOption) (*wflows.VpcPrefixList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve vpcprefixes")
	}

	out := &wflows.VpcPrefixList{}
	if in != nil {
		for _, id := range in.VpcPrefixIds {
			out.VpcPrefixes = append(out.VpcPrefixes, &wflows.VpcPrefix{
				Id: id,
			})
		}
	}

	return out, nil
}

/* VPC Peering mock methods */
func (c *MockNICoClient) CreateVpcPeering(ctx context.Context, in *wflows.VpcPeeringCreationRequest, opts ...grpc.CallOption) (*wflows.VpcPeering, error) {
	out := new(wflows.VpcPeering)
	out.Id = &wflows.VpcPeeringId{Value: uuid.NewString()}
	return out, nil
}

func (c *MockNICoClient) DeleteVpcPeering(ctx context.Context, in *wflows.VpcPeeringDeletionRequest, opts ...grpc.CallOption) (*wflows.VpcPeeringDeletionResult, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to delete vpc peering")
	}

	return &wflows.VpcPeeringDeletionResult{}, nil
}

func (c *MockNICoClient) FindVpcPeeringIds(ctx context.Context, in *wflows.VpcPeeringSearchFilter, opts ...grpc.CallOption) (*wflows.VpcPeeringIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve vpc peering ids")
	}

	out := &wflows.VpcPeeringIdList{}

	count, ok := ctx.Value("WantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.VpcPeeringIds = append(out.VpcPeeringIds, &wflows.VpcPeeringId{Value: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindVpcPeeringsByIds(ctx context.Context, in *wflows.VpcPeeringsByIdsRequest, opts ...grpc.CallOption) (*wflows.VpcPeeringList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve vpc peerings")
	}

	out := &wflows.VpcPeeringList{}
	for _, id := range in.VpcPeeringIds {
		out.VpcPeerings = append(out.VpcPeerings, &wflows.VpcPeering{
			Id:        id,
			VpcId:     &wflows.VpcId{Value: uuid.NewString()},
			PeerVpcId: &wflows.VpcId{Value: uuid.NewString()},
		})
	}

	return out, nil
}

/* Machine Validation Test mock methods */
func (c *MockNICoClient) AddMachineValidationTest(ctx context.Context, in *wflows.MachineValidationTestAddRequest, opts ...grpc.CallOption) (*wflows.MachineValidationTestAddUpdateResponse, error) {
	out := new(wflows.MachineValidationTestAddUpdateResponse)
	id, ok := ctx.Value("wantID").(string)
	if ok {
		out.TestId = id
		out.Version = "version-1"
	}
	return out, nil
}

func (c *MockNICoClient) UpdateMachineValidationTest(ctx context.Context, in *wflows.MachineValidationTestUpdateRequest, opts ...grpc.CallOption) (*wflows.MachineValidationTestAddUpdateResponse, error) {
	out := new(wflows.MachineValidationTestAddUpdateResponse)
	out.TestId = in.TestId
	out.Version = in.Version
	return out, nil
}

func (c *MockNICoClient) GetMachineValidationTests(ctx context.Context, in *wflows.MachineValidationTestsGetRequest, opts ...grpc.CallOption) (*wflows.MachineValidationTestsGetResponse, error) {
	out := new(wflows.MachineValidationTestsGetResponse)
	return out, nil
}

func (c *MockNICoClient) MachineValidationTestEnableDisableTest(ctx context.Context, in *wflows.MachineValidationTestEnableDisableTestRequest, opts ...grpc.CallOption) (*wflows.MachineValidationTestEnableDisableTestResponse, error) {
	out := new(wflows.MachineValidationTestEnableDisableTestResponse)
	return out, nil
}

func (c *MockNICoClient) AddUpdateMachineValidationExternalConfig(ctx context.Context, in *wflows.AddUpdateMachineValidationExternalConfigRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) RemoveMachineValidationExternalConfig(ctx context.Context, in *wflows.RemoveMachineValidationExternalConfigRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) GetMachineValidationExternalConfigs(ctx context.Context, in *wflows.GetMachineValidationExternalConfigsRequest, opts ...grpc.CallOption) (*wflows.GetMachineValidationExternalConfigsResponse, error) {
	out := new(wflows.GetMachineValidationExternalConfigsResponse)
	return out, nil
}

func (c *MockNICoClient) GetMachineValidationRuns(ctx context.Context, in *wflows.MachineValidationRunListGetRequest, opts ...grpc.CallOption) (*wflows.MachineValidationRunList, error) {
	out := new(wflows.MachineValidationRunList)
	return out, nil
}

func (c *MockNICoClient) GetMachineValidationResults(ctx context.Context, in *wflows.MachineValidationGetRequest, opts ...grpc.CallOption) (*wflows.MachineValidationResultList, error) {
	out := new(wflows.MachineValidationResultList)
	return out, nil
}

func (c *MockNICoClient) PersistValidationResult(ctx context.Context, in *wflows.MachineValidationResultPostRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

/* Network Security Group mock methods */
func (c *MockNICoClient) UpdateMachineValidationRun(ctx context.Context, in *wflows.MachineValidationRunRequest, opts ...grpc.CallOption) (*wflows.MachineValidationRunResponse, error) {
	out := new(wflows.MachineValidationRunResponse)
	return out, nil
}

func (c *MockNICoClient) CreateNetworkSecurityGroup(ctx context.Context, in *wflows.CreateNetworkSecurityGroupRequest, opts ...grpc.CallOption) (*wflows.CreateNetworkSecurityGroupResponse, error) {
	out := &wflows.CreateNetworkSecurityGroupResponse{NetworkSecurityGroup: &wflows.NetworkSecurityGroup{}}
	return out, nil
}

func (c *MockNICoClient) UpdateNetworkSecurityGroup(ctx context.Context, in *wflows.UpdateNetworkSecurityGroupRequest, opts ...grpc.CallOption) (*wflows.UpdateNetworkSecurityGroupResponse, error) {
	out := &wflows.UpdateNetworkSecurityGroupResponse{NetworkSecurityGroup: &wflows.NetworkSecurityGroup{}}
	return out, nil
}

func (c *MockNICoClient) DeleteNetworkSecurityGroup(ctx context.Context, in *wflows.DeleteNetworkSecurityGroupRequest, opts ...grpc.CallOption) (*wflows.DeleteNetworkSecurityGroupResponse, error) {
	out := &wflows.DeleteNetworkSecurityGroupResponse{}
	return out, nil
}

func (c *MockNICoClient) GetNetworkSecurityGroupAttachments(ctx context.Context, in *wflows.GetNetworkSecurityGroupAttachmentsRequest, opts ...grpc.CallOption) (*wflows.GetNetworkSecurityGroupAttachmentsResponse, error) {
	out := &wflows.GetNetworkSecurityGroupAttachmentsResponse{}
	return out, nil
}

func (c *MockNICoClient) GetNetworkSecurityGroupPropagationStatus(ctx context.Context, in *wflows.GetNetworkSecurityGroupPropagationStatusRequest, opts ...grpc.CallOption) (*wflows.GetNetworkSecurityGroupPropagationStatusResponse, error) {
	out := &wflows.GetNetworkSecurityGroupPropagationStatusResponse{}
	return out, nil
}

func (c *MockNICoClient) FindNetworkSecurityGroupIds(ctx context.Context, in *wflows.FindNetworkSecurityGroupIdsRequest, opts ...grpc.CallOption) (*wflows.FindNetworkSecurityGroupIdsResponse, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve NetworkSecurityGroup ids")
	}

	out := &wflows.FindNetworkSecurityGroupIdsResponse{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.NetworkSecurityGroupIds = append(out.NetworkSecurityGroupIds, randSeq(10))
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindNetworkSecurityGroupsByIds(ctx context.Context, in *wflows.FindNetworkSecurityGroupsByIdsRequest, opts ...grpc.CallOption) (*wflows.FindNetworkSecurityGroupsByIdsResponse, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve NetworkSecurityGroups")
	}

	out := &wflows.FindNetworkSecurityGroupsByIdsResponse{}
	if in != nil {
		for _, id := range in.NetworkSecurityGroupIds {
			out.NetworkSecurityGroups = append(out.NetworkSecurityGroups, &wflows.NetworkSecurityGroup{
				Id: id,
			})
		}
	}
	return out, nil
}

/* Expected Machine mock methods */
func (c *MockNICoClient) AddExpectedMachine(ctx context.Context, in *wflows.ExpectedMachine, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.Id == nil || in.Id.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for AddExpectedMachine")
	}
	if in.BmcMacAddress == "" {
		return nil, status.Error(codes.Internal, "MAC address not provided for AddExpectedMachine")
	}
	if in.ChassisSerialNumber == "" {
		return nil, status.Error(codes.Internal, "Chassis Serial Number not provided for AddExpectedMachine")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) DeleteExpectedMachine(ctx context.Context, in *wflows.ExpectedMachineRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.Id == nil || in.Id.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for DeleteExpectedMachine")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) UpdateExpectedMachine(ctx context.Context, in *wflows.ExpectedMachine, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.Id == nil || in.Id.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for UpdateExpectedMachine")
	}
	if in.BmcMacAddress == "" {
		return nil, status.Error(codes.Internal, "MAC address not provided for UpdateExpectedMachine")
	}
	if in.ChassisSerialNumber == "" {
		return nil, status.Error(codes.Internal, "Chassis Serial Number not provided for UpdateExpectedMachine")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) CreateExpectedMachines(ctx context.Context, in *wflows.BatchExpectedMachineOperationRequest, opts ...grpc.CallOption) (*wflows.BatchExpectedMachineOperationResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	out := &wflows.BatchExpectedMachineOperationResponse{
		Results: make([]*wflows.ExpectedMachineOperationResult, 0, len(in.GetExpectedMachines().GetExpectedMachines())),
	}

	// Simulate individual processing of each ExpectedMachine
	for _, em := range in.GetExpectedMachines().GetExpectedMachines() {
		result := &wflows.ExpectedMachineOperationResult{
			Id:              em.GetId(),
			Success:         true,
			ExpectedMachine: em,
		}

		// Validate required fields
		if em.GetId() == nil || em.GetId().GetValue() == "" {
			result.Success = false
			msg := "ID not provided"
			result.ErrorMessage = &msg
			result.ExpectedMachine = nil
		} else if em.GetBmcMacAddress() == "" {
			result.Success = false
			msg := "MAC address not provided"
			result.ErrorMessage = &msg
			result.ExpectedMachine = nil
		} else if em.GetChassisSerialNumber() == "" {
			result.Success = false
			msg := "Chassis Serial Number not provided"
			result.ErrorMessage = &msg
			result.ExpectedMachine = nil
		}

		out.Results = append(out.Results, result)
	}

	return out, nil
}

func (c *MockNICoClient) UpdateExpectedMachines(ctx context.Context, in *wflows.BatchExpectedMachineOperationRequest, opts ...grpc.CallOption) (*wflows.BatchExpectedMachineOperationResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	out := &wflows.BatchExpectedMachineOperationResponse{
		Results: make([]*wflows.ExpectedMachineOperationResult, 0, len(in.GetExpectedMachines().GetExpectedMachines())),
	}

	// Simulate individual processing of each ExpectedMachine
	for _, em := range in.GetExpectedMachines().GetExpectedMachines() {
		result := &wflows.ExpectedMachineOperationResult{
			Id:              em.GetId(),
			Success:         true,
			ExpectedMachine: em,
		}

		// Validate required fields
		if em.GetId() == nil || em.GetId().GetValue() == "" {
			result.Success = false
			msg := "ID not provided"
			result.ErrorMessage = &msg
			result.ExpectedMachine = nil
		} else if em.GetBmcMacAddress() == "" {
			result.Success = false
			msg := "MAC address not provided"
			result.ErrorMessage = &msg
			result.ExpectedMachine = nil
		} else if em.GetChassisSerialNumber() == "" {
			result.Success = false
			msg := "Chassis Serial Number not provided"
			result.ErrorMessage = &msg
			result.ExpectedMachine = nil
		}

		out.Results = append(out.Results, result)
	}

	return out, nil
}

func (c *MockNICoClient) GetAllExpectedMachines(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.ExpectedMachineList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve machine ids")
		}
	}

	out := &wflows.ExpectedMachineList{}

	// we generate predictable unique IDs and values
	count, ok := ctx.Value("wantCount").(int)
	if ok {
		mac, _ := net.ParseMAC("02:00:00:00:00:00")
		for i := 0; i < count; i++ {
			// Create a 16-byte array for UUID from MAC address (6 bytes) + padding
			var uuidBytes [16]byte
			copy(uuidBytes[:6], mac)
			emID, _ := uuid.FromBytes(uuidBytes[:])
			out.ExpectedMachines = append(out.ExpectedMachines, &wflows.ExpectedMachine{
				Id:                  &wflows.UUID{Value: emID.String()},
				BmcMacAddress:       mac.String(),
				ChassisSerialNumber: "serial-" + mac.String()})
			incrementMAC(mac)
		}
	}

	return out, nil
}

func (c *MockNICoClient) GetExpectedMachine(ctx context.Context, in *wflows.ExpectedMachineRequest, opts ...grpc.CallOption) (*wflows.ExpectedMachine, error) {
	if in.Id == nil || in.Id.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for GetExpectedMachine")
	}
	out := new(wflows.ExpectedMachine)
	return out, nil
}

func (c *MockNICoClient) GetAllExpectedMachinesLinked(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.LinkedExpectedMachineList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve linked expected machines")
		}
	}

	out := &wflows.LinkedExpectedMachineList{}

	// Generate linked machines based on the count in context
	count, ok := ctx.Value("wantCount").(int)
	if ok {
		mac, _ := net.ParseMAC("02:00:00:00:00:00")
		for i := 0; i < count; i++ {
			// Create a 16-byte array for UUID from MAC address (6 bytes) + padding
			var uuidBytes [16]byte
			copy(uuidBytes[:6], mac)
			machineID, _ := uuid.FromBytes(uuidBytes[:])

			out.ExpectedMachines = append(out.ExpectedMachines, &wflows.LinkedExpectedMachine{
				ChassisSerialNumber: "serial-" + mac.String(),
				BmcMacAddress:       mac.String(),
				MachineId:           &wflows.MachineId{Id: machineID.String()},
			})
			incrementMAC(mac)
		}
	}

	return out, nil
}

/* Expected Power Shelf mock methods */
func (c *MockNICoClient) AddExpectedPowerShelf(ctx context.Context, in *wflows.ExpectedPowerShelf, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.ExpectedPowerShelfId == nil || in.ExpectedPowerShelfId.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for AddExpectedPowerShelf")
	}
	if in.BmcMacAddress == "" {
		return nil, status.Error(codes.Internal, "MAC address not provided for AddExpectedPowerShelf")
	}
	if in.ShelfSerialNumber == "" {
		return nil, status.Error(codes.Internal, "Shelf Serial Number not provided for AddExpectedPowerShelf")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) DeleteExpectedPowerShelf(ctx context.Context, in *wflows.ExpectedPowerShelfRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.ExpectedPowerShelfId == nil || in.ExpectedPowerShelfId.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for DeleteExpectedPowerShelf")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) UpdateExpectedPowerShelf(ctx context.Context, in *wflows.ExpectedPowerShelf, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.ExpectedPowerShelfId == nil || in.ExpectedPowerShelfId.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for UpdateExpectedPowerShelf")
	}
	if in.BmcMacAddress == "" {
		return nil, status.Error(codes.Internal, "MAC address not provided for UpdateExpectedPowerShelf")
	}
	if in.ShelfSerialNumber == "" {
		return nil, status.Error(codes.Internal, "Shelf Serial Number not provided for UpdateExpectedPowerShelf")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) GetAllExpectedPowerShelves(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.ExpectedPowerShelfList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve expected power shelves")
		}
	}

	out := &wflows.ExpectedPowerShelfList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		mac, _ := net.ParseMAC("02:00:00:00:00:00")
		for i := 0; i < count; i++ {
			var uuidBytes [16]byte
			copy(uuidBytes[:6], mac)
			epsID, _ := uuid.FromBytes(uuidBytes[:])
			out.ExpectedPowerShelves = append(out.ExpectedPowerShelves, &wflows.ExpectedPowerShelf{
				ExpectedPowerShelfId: &wflows.UUID{Value: epsID.String()},
				BmcMacAddress:        mac.String(),
				ShelfSerialNumber:    "shelf-serial-" + mac.String()})
			incrementMAC(mac)
		}
	}

	return out, nil
}

func (c *MockNICoClient) GetAllExpectedPowerShelvesLinked(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.LinkedExpectedPowerShelfList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve linked expected power shelves")
		}
	}

	out := &wflows.LinkedExpectedPowerShelfList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		mac, _ := net.ParseMAC("02:00:00:00:00:00")
		for i := 0; i < count; i++ {
			var uuidBytes [16]byte
			copy(uuidBytes[:6], mac)
			powerShelfID, _ := uuid.FromBytes(uuidBytes[:])

			out.ExpectedPowerShelves = append(out.ExpectedPowerShelves, &wflows.LinkedExpectedPowerShelf{
				ShelfSerialNumber: "shelf-serial-" + mac.String(),
				BmcMacAddress:     mac.String(),
				PowerShelfId:      &wflows.PowerShelfId{Id: powerShelfID.String()},
			})
			incrementMAC(mac)
		}
	}

	return out, nil
}

/* Expected Switch mock methods */
func (c *MockNICoClient) AddExpectedSwitch(ctx context.Context, in *wflows.ExpectedSwitch, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.ExpectedSwitchId == nil || in.ExpectedSwitchId.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for AddExpectedSwitch")
	}
	if in.BmcMacAddress == "" {
		return nil, status.Error(codes.Internal, "MAC address not provided for AddExpectedSwitch")
	}
	if in.SwitchSerialNumber == "" {
		return nil, status.Error(codes.Internal, "Switch Serial Number not provided for AddExpectedSwitch")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) DeleteExpectedSwitch(ctx context.Context, in *wflows.ExpectedSwitchRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.ExpectedSwitchId == nil || in.ExpectedSwitchId.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for DeleteExpectedSwitch")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) UpdateExpectedSwitch(ctx context.Context, in *wflows.ExpectedSwitch, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.ExpectedSwitchId == nil || in.ExpectedSwitchId.Value == "" {
		return nil, status.Error(codes.Internal, "ID not provided for UpdateExpectedSwitch")
	}
	if in.BmcMacAddress == "" {
		return nil, status.Error(codes.Internal, "MAC address not provided for UpdateExpectedSwitch")
	}
	if in.SwitchSerialNumber == "" {
		return nil, status.Error(codes.Internal, "Switch Serial Number not provided for UpdateExpectedSwitch")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) GetAllExpectedSwitches(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.ExpectedSwitchList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve expected switches")
		}
	}

	out := &wflows.ExpectedSwitchList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		mac, _ := net.ParseMAC("02:00:00:00:00:00")
		for i := 0; i < count; i++ {
			var uuidBytes [16]byte
			copy(uuidBytes[:6], mac)
			esID, _ := uuid.FromBytes(uuidBytes[:])
			out.ExpectedSwitches = append(out.ExpectedSwitches, &wflows.ExpectedSwitch{
				ExpectedSwitchId:   &wflows.UUID{Value: esID.String()},
				BmcMacAddress:      mac.String(),
				SwitchSerialNumber: "switch-serial-" + mac.String()})
			incrementMAC(mac)
		}
	}

	return out, nil
}

func (c *MockNICoClient) GetAllExpectedSwitchesLinked(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.LinkedExpectedSwitchList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve linked expected switches")
		}
	}

	out := &wflows.LinkedExpectedSwitchList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		mac, _ := net.ParseMAC("02:00:00:00:00:00")
		for i := 0; i < count; i++ {
			var uuidBytes [16]byte
			copy(uuidBytes[:6], mac)
			switchID, _ := uuid.FromBytes(uuidBytes[:])

			out.ExpectedSwitches = append(out.ExpectedSwitches, &wflows.LinkedExpectedSwitch{
				SwitchSerialNumber: "switch-serial-" + mac.String(),
				BmcMacAddress:      mac.String(),
				SwitchId:           &wflows.SwitchId{Id: switchID.String()},
			})
			incrementMAC(mac)
		}
	}

	return out, nil
}

/* Expected Rack mock methods */
func (c *MockNICoClient) AddExpectedRack(ctx context.Context, in *wflows.ExpectedRack, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.RackId == nil || in.RackId.Id == "" {
		return nil, status.Error(codes.Internal, "ID not provided for AddExpectedRack")
	}
	if in.RackType == "" {
		return nil, status.Error(codes.Internal, "Rack Profile ID not provided for AddExpectedRack")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) UpdateExpectedRack(ctx context.Context, in *wflows.ExpectedRack, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.RackId == nil || in.RackId.Id == "" {
		return nil, status.Error(codes.Internal, "ID not provided for UpdateExpectedRack")
	}
	if in.RackType == "" {
		return nil, status.Error(codes.Internal, "Rack Profile ID not provided for UpdateExpectedRack")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) DeleteExpectedRack(ctx context.Context, in *wflows.ExpectedRackRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in.RackId == "" {
		return nil, status.Error(codes.Internal, "ID not provided for DeleteExpectedRack")
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) GetExpectedRack(ctx context.Context, in *wflows.ExpectedRackRequest, opts ...grpc.CallOption) (*wflows.ExpectedRack, error) {
	if in.RackId == "" {
		return nil, status.Error(codes.Internal, "ID not provided for GetExpectedRack")
	}
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve expected rack")
		}
	}
	out := &wflows.ExpectedRack{
		RackId:   &wflows.RackId{Id: in.RackId},
		RackType: uuid.NewString(),
	}
	return out, nil
}

func (c *MockNICoClient) GetAllExpectedRacks(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.ExpectedRackList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		if status.Code(err) == codes.Internal {
			return nil, status.Error(codes.Internal, "failed to retrieve expected racks")
		}
	}

	out := &wflows.ExpectedRackList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.ExpectedRacks = append(out.ExpectedRacks, &wflows.ExpectedRack{
				RackId:   &wflows.RackId{Id: uuid.NewString()},
				RackType: uuid.NewString(),
			})
		}
	}

	return out, nil
}

func (c *MockNICoClient) ReplaceAllExpectedRacks(ctx context.Context, in *wflows.ExpectedRackList, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	if in == nil {
		return nil, status.Error(codes.Internal, "Invalid request argument")
	}
	for _, er := range in.ExpectedRacks {
		if er == nil || er.RackId == nil || er.RackId.Id == "" {
			return nil, status.Error(codes.Internal, "ID not provided for ReplaceAllExpectedRacks")
		}
		if er.RackType == "" {
			return nil, status.Error(codes.Internal, "Rack Profile ID not provided for ReplaceAllExpectedRacks")
		}
	}
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockNICoClient) DeleteAllExpectedRacks(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

/* SKU mock methods */
func (c *MockNICoClient) FindSkusByIds(ctx context.Context, in *wflows.SkusByIdsRequest, opts ...grpc.CallOption) (*wflows.SkuList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve skus")
	}

	out := &wflows.SkuList{}
	if in != nil {
		for _, id := range in.Ids {
			out.Skus = append(out.Skus, &wflows.Sku{
				Id: id,
			})
		}
	}

	return out, nil
}

func (c *MockNICoClient) GetAllSkuIds(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*wflows.SkuIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve sku ids")
	}

	out := &wflows.SkuIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.Ids = append(out.Ids, uuid.NewString())
		}
	}

	return out, nil
}

/* DPU Extension Service mock methods */
func (c *MockNICoClient) CreateDpuExtensionService(ctx context.Context, in *wflows.CreateDpuExtensionServiceRequest, opts ...grpc.CallOption) (*wflows.DpuExtensionService, error) {
	versionInfo := &wflows.DpuExtensionServiceVersionInfo{
		Version:       generateSiteVersion(),
		Data:          "test data",
		HasCredential: false,
		Observability: in.Observability,
	}

	serviceID := uuid.NewString()
	if in.ServiceId != nil {
		serviceID = *in.ServiceId
	}

	out := &wflows.DpuExtensionService{
		ServiceId:            serviceID,
		ServiceName:          in.ServiceName,
		ServiceType:          in.ServiceType,
		TenantOrganizationId: in.TenantOrganizationId,
		LatestVersionInfo:    versionInfo,
		ActiveVersions:       []string{versionInfo.Version},
	}

	if in.Description != nil {
		out.Description = *in.Description
	}

	return out, nil
}

func (c *MockNICoClient) UpdateDpuExtensionService(ctx context.Context, in *wflows.UpdateDpuExtensionServiceRequest, opts ...grpc.CallOption) (*wflows.DpuExtensionService, error) {
	versionInfo := &wflows.DpuExtensionServiceVersionInfo{
		Version:       generateSiteVersion(),
		Data:          "test data",
		HasCredential: false,
		Observability: in.Observability,
	}

	out := &wflows.DpuExtensionService{
		ServiceId:         in.ServiceId,
		LatestVersionInfo: versionInfo,
		ActiveVersions:    []string{versionInfo.Version},
	}

	if in.ServiceName != nil {
		out.ServiceName = *in.ServiceName
	}

	if in.Description != nil {
		out.Description = *in.Description
	}

	return out, nil
}

func (c *MockNICoClient) DeleteDpuExtensionService(ctx context.Context, in *wflows.DeleteDpuExtensionServiceRequest, opts ...grpc.CallOption) (*wflows.DeleteDpuExtensionServiceResponse, error) {
	out := new(wflows.DeleteDpuExtensionServiceResponse)
	return out, nil
}

func (c *MockNICoClient) FindDpuExtensionServiceIds(ctx context.Context, in *wflows.DpuExtensionServiceSearchFilter, opts ...grpc.CallOption) (*wflows.DpuExtensionServiceIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve dpu extension service ids")
	}

	out := &wflows.DpuExtensionServiceIdList{}
	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.ServiceIds = append(out.ServiceIds, uuid.NewString())
		}
	}
	return out, nil
}

func (c *MockNICoClient) FindDpuExtensionServicesByIds(ctx context.Context, in *wflows.DpuExtensionServicesByIdsRequest, opts ...grpc.CallOption) (*wflows.DpuExtensionServiceList, error) {
	out := &wflows.DpuExtensionServiceList{}
	if in != nil {
		for _, id := range in.ServiceIds {
			out.Services = append(out.Services, &wflows.DpuExtensionService{
				ServiceId: id,
			})
		}
	}
	return out, nil
}

func (c *MockNICoClient) GetDpuExtensionServiceVersionsInfo(ctx context.Context, in *wflows.GetDpuExtensionServiceVersionsInfoRequest, opts ...grpc.CallOption) (*wflows.DpuExtensionServiceVersionInfoList, error) {
	out := &wflows.DpuExtensionServiceVersionInfoList{
		VersionInfos: []*wflows.DpuExtensionServiceVersionInfo{},
	}
	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.VersionInfos = append(out.VersionInfos, &wflows.DpuExtensionServiceVersionInfo{
				Version:       generateSiteVersion(),
				Data:          "test data",
				HasCredential: false,
			})
		}
	}
	return out, nil
}

// NVLink Logical Partition Mocks
func (c *MockNICoClient) CreateNVLinkLogicalPartition(ctx context.Context, in *wflows.NVLinkLogicalPartitionCreationRequest, opts ...grpc.CallOption) (*wflows.NVLinkLogicalPartition, error) {
	out := new(wflows.NVLinkLogicalPartition)
	if in != nil {
		out.Id = in.Id
		out.Config = in.Config
		out.Config.Metadata = in.Config.Metadata
		out.Config.TenantOrganizationId = in.Config.TenantOrganizationId
		out.Status = &wflows.NVLinkLogicalPartitionStatus{
			State: wflows.TenantState_READY,
		}
	}
	return out, nil
}

func (c *MockNICoClient) UpdateNVLinkLogicalPartition(ctx context.Context, in *wflows.NVLinkLogicalPartitionUpdateRequest, opts ...grpc.CallOption) (*wflows.NVLinkLogicalPartitionUpdateResult, error) {
	out := new(wflows.NVLinkLogicalPartitionUpdateResult)
	return out, nil
}

func (c *MockNICoClient) DeleteNVLinkLogicalPartition(ctx context.Context, in *wflows.NVLinkLogicalPartitionDeletionRequest, opts ...grpc.CallOption) (*wflows.NVLinkLogicalPartitionDeletionResult, error) {
	out := new(wflows.NVLinkLogicalPartitionDeletionResult)
	return out, nil
}

func (c *MockNICoClient) FindNVLinkLogicalPartitionIds(ctx context.Context, in *wflows.NVLinkLogicalPartitionSearchFilter, opts ...grpc.CallOption) (*wflows.NVLinkLogicalPartitionIdList, error) {
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, status.Error(status.Code(err), "failed to retrieve nvlink logical partition ids")
	}

	out := &wflows.NVLinkLogicalPartitionIdList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.PartitionIds = append(out.PartitionIds, &wflows.NVLinkLogicalPartitionId{Value: uuid.NewString()})
		}
	}

	return out, nil
}

func (c *MockNICoClient) FindNVLinkLogicalPartitionsByIds(ctx context.Context, in *wflows.NVLinkLogicalPartitionsByIdsRequest, opts ...grpc.CallOption) (*wflows.NVLinkLogicalPartitionList, error) {
	err, ok := ctx.Value("wantError").(error)
	if ok {
		return nil, status.Error(status.Code(err), "failed to retrieve nvlink logical partitions")
	}

	out := &wflows.NVLinkLogicalPartitionList{}
	if in != nil {
		for _, id := range in.PartitionIds {
			out.Partitions = append(out.Partitions, &wflows.NVLinkLogicalPartition{
				Id: id,
			})
		}
	}

	return out, nil
}

func (c *MockNICoClient) NVLinkLogicalPartitionsForTenant(ctx context.Context, in *wflows.TenantSearchQuery, opts ...grpc.CallOption) (*wflows.NVLinkLogicalPartitionList, error) {
	out := &wflows.NVLinkLogicalPartitionList{}

	count, ok := ctx.Value("wantCount").(int)
	if ok {
		for i := 0; i < count; i++ {
			out.Partitions = append(out.Partitions, &wflows.NVLinkLogicalPartition{
				Id: &wflows.NVLinkLogicalPartitionId{Value: uuid.NewString()},
			})
		}
	}

	return out, nil
}

// NewMockNICoClient creates a new mock NICoCoreClient
func NewMockNICoClient() *NICoCoreClient {
	return &NICoCoreClient{
		nico: &MockNICoClient{},
	}
}

// MockFlowClient is a mock implementation of Flow gRPC protobuf Client
type MockFlowClient struct {
	flowv1.FlowClient
}

/* Version mock methods */
func (c *MockFlowClient) Version(ctx context.Context, in *flowv1.VersionRequest, opts ...grpc.CallOption) (*flowv1.BuildInfo, error) {
	out := &flowv1.BuildInfo{
		Version:   "1.0.0",
		BuildTime: time.Now().Format(time.RFC3339),
		GitCommit: "test-commit",
	}
	return out, nil
}

/* Rack mock methods */
func (c *MockFlowClient) CreateExpectedRack(ctx context.Context, in *flowv1.CreateExpectedRackRequest, opts ...grpc.CallOption) (*flowv1.CreateExpectedRackResponse, error) {
	out := &flowv1.CreateExpectedRackResponse{
		Id: &flowv1.UUID{Id: uuid.NewString()},
	}
	return out, nil
}

func (c *MockFlowClient) PatchRack(ctx context.Context, in *flowv1.PatchRackRequest, opts ...grpc.CallOption) (*flowv1.PatchRackResponse, error) {
	out := new(flowv1.PatchRackResponse)
	return out, nil
}

func (c *MockFlowClient) GetRackInfoByID(ctx context.Context, in *flowv1.GetRackInfoByIDRequest, opts ...grpc.CallOption) (*flowv1.GetRackInfoResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	// Check for custom response via context
	if resp, ok := ctx.Value("wantResponse").(*flowv1.GetRackInfoResponse); ok {
		return resp, nil
	}

	out := &flowv1.GetRackInfoResponse{
		Rack: &flowv1.Rack{
			Info: &flowv1.DeviceInfo{
				Id: in.GetId(),
			},
		},
	}
	return out, nil
}

func (c *MockFlowClient) GetRackInfoBySerial(ctx context.Context, in *flowv1.GetRackInfoBySerialRequest, opts ...grpc.CallOption) (*flowv1.GetRackInfoResponse, error) {
	out := &flowv1.GetRackInfoResponse{
		Rack: &flowv1.Rack{
			Info: &flowv1.DeviceInfo{
				SerialNumber: in.GetSerialInfo().GetSerialNumber(),
			},
		},
	}
	return out, nil
}

func (c *MockFlowClient) GetListOfRacks(ctx context.Context, in *flowv1.GetListOfRacksRequest, opts ...grpc.CallOption) (*flowv1.GetListOfRacksResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	// Check for custom response via context
	if resp, ok := ctx.Value("wantResponse").(*flowv1.GetListOfRacksResponse); ok {
		return resp, nil
	}

	out := &flowv1.GetListOfRacksResponse{
		Racks: []*flowv1.Rack{},
	}
	return out, nil
}

/* Component mock methods */
func (c *MockFlowClient) GetComponentInfoByID(ctx context.Context, in *flowv1.GetComponentInfoByIDRequest, opts ...grpc.CallOption) (*flowv1.GetComponentInfoResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	// Check for custom response via context
	if resp, ok := ctx.Value("wantResponse").(*flowv1.GetComponentInfoResponse); ok {
		return resp, nil
	}

	out := &flowv1.GetComponentInfoResponse{
		Component: &flowv1.Component{
			ComponentId: in.GetId().GetId(),
		},
	}
	return out, nil
}

func (c *MockFlowClient) GetComponentInfoBySerial(ctx context.Context, in *flowv1.GetComponentInfoBySerialRequest, opts ...grpc.CallOption) (*flowv1.GetComponentInfoResponse, error) {
	out := &flowv1.GetComponentInfoResponse{
		Component: &flowv1.Component{
			Info: &flowv1.DeviceInfo{
				SerialNumber: in.GetSerialInfo().GetSerialNumber(),
			},
		},
	}
	return out, nil
}

func (c *MockFlowClient) GetComponents(ctx context.Context, in *flowv1.GetComponentsRequest, opts ...grpc.CallOption) (*flowv1.GetComponentsResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	// Check for custom response via context
	if resp, ok := ctx.Value("wantResponse").(*flowv1.GetComponentsResponse); ok {
		return resp, nil
	}

	out := &flowv1.GetComponentsResponse{
		Components: []*flowv1.Component{},
		Total:      0,
	}
	return out, nil
}

func (c *MockFlowClient) ValidateComponents(ctx context.Context, in *flowv1.ValidateComponentsRequest, opts ...grpc.CallOption) (*flowv1.ValidateComponentsResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	// Check for custom response via context
	if resp, ok := ctx.Value("wantResponse").(*flowv1.ValidateComponentsResponse); ok {
		return resp, nil
	}

	out := &flowv1.ValidateComponentsResponse{
		Diffs:           []*flowv1.ComponentDiff{},
		TotalDiffs:      0,
		MissingCount:    0,
		UnexpectedCount: 0,
		DriftCount:      0,
		MatchCount:      0,
	}
	return out, nil
}

/* Component mutation mock methods */
func (c *MockFlowClient) AddComponent(ctx context.Context, in *flowv1.AddComponentRequest, opts ...grpc.CallOption) (*flowv1.AddComponentResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	// Check for custom response via context
	if resp, ok := ctx.Value("wantResponse").(*flowv1.AddComponentResponse); ok {
		return resp, nil
	}

	out := &flowv1.AddComponentResponse{
		Component: &flowv1.Component{},
	}
	return out, nil
}

func (c *MockFlowClient) PatchComponent(ctx context.Context, in *flowv1.PatchComponentRequest, opts ...grpc.CallOption) (*flowv1.PatchComponentResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	// Check for custom response via context
	if resp, ok := ctx.Value("wantResponse").(*flowv1.PatchComponentResponse); ok {
		return resp, nil
	}

	out := &flowv1.PatchComponentResponse{
		Component: &flowv1.Component{},
	}
	return out, nil
}

func (c *MockFlowClient) DeleteComponent(ctx context.Context, in *flowv1.DeleteComponentRequest, opts ...grpc.CallOption) (*flowv1.DeleteComponentResponse, error) {
	// Check for error injection via context
	if err, ok := ctx.Value("wantError").(error); ok {
		return nil, err
	}

	out := &flowv1.DeleteComponentResponse{}
	return out, nil
}

/* NVL Domain mock methods */
func (c *MockFlowClient) CreateNVLDomain(ctx context.Context, in *flowv1.CreateNVLDomainRequest, opts ...grpc.CallOption) (*flowv1.CreateNVLDomainResponse, error) {
	out := &flowv1.CreateNVLDomainResponse{
		Id: &flowv1.UUID{Id: uuid.NewString()},
	}
	return out, nil
}

func (c *MockFlowClient) AttachRacksToNVLDomain(ctx context.Context, in *flowv1.AttachRacksToNVLDomainRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockFlowClient) DetachRacksFromNVLDomain(ctx context.Context, in *flowv1.DetachRacksFromNVLDomainRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockFlowClient) GetListOfNVLDomains(ctx context.Context, in *flowv1.GetListOfNVLDomainsRequest, opts ...grpc.CallOption) (*flowv1.GetListOfNVLDomainsResponse, error) {
	out := &flowv1.GetListOfNVLDomainsResponse{
		NvlDomains: []*flowv1.NVLDomain{},
		Total:      0,
	}
	return out, nil
}

func (c *MockFlowClient) GetRacksForNVLDomain(ctx context.Context, in *flowv1.GetRacksForNVLDomainRequest, opts ...grpc.CallOption) (*flowv1.GetRacksForNVLDomainResponse, error) {
	out := &flowv1.GetRacksForNVLDomainResponse{
		Racks: []*flowv1.Rack{},
	}
	return out, nil
}

/* Task mock methods */
func (c *MockFlowClient) UpgradeFirmware(ctx context.Context, in *flowv1.UpgradeFirmwareRequest, opts ...grpc.CallOption) (*flowv1.SubmitTaskResponse, error) {
	out := &flowv1.SubmitTaskResponse{
		TaskIds: []*flowv1.UUID{{Id: uuid.NewString()}},
	}
	return out, nil
}

func (c *MockFlowClient) PowerOnRack(ctx context.Context, in *flowv1.PowerOnRackRequest, opts ...grpc.CallOption) (*flowv1.SubmitTaskResponse, error) {
	out := &flowv1.SubmitTaskResponse{
		TaskIds: []*flowv1.UUID{{Id: uuid.NewString()}},
	}
	return out, nil
}

func (c *MockFlowClient) PowerOffRack(ctx context.Context, in *flowv1.PowerOffRackRequest, opts ...grpc.CallOption) (*flowv1.SubmitTaskResponse, error) {
	out := &flowv1.SubmitTaskResponse{
		TaskIds: []*flowv1.UUID{{Id: uuid.NewString()}},
	}
	return out, nil
}

func (c *MockFlowClient) PowerResetRack(ctx context.Context, in *flowv1.PowerResetRackRequest, opts ...grpc.CallOption) (*flowv1.SubmitTaskResponse, error) {
	out := &flowv1.SubmitTaskResponse{
		TaskIds: []*flowv1.UUID{{Id: uuid.NewString()}},
	}
	return out, nil
}

func (c *MockFlowClient) BringUpRack(ctx context.Context, in *flowv1.BringUpRackRequest, opts ...grpc.CallOption) (*flowv1.SubmitTaskResponse, error) {
	out := &flowv1.SubmitTaskResponse{
		TaskIds: []*flowv1.UUID{{Id: uuid.NewString()}},
	}
	return out, nil
}

func (c *MockFlowClient) IngestRack(ctx context.Context, in *flowv1.IngestRackRequest, opts ...grpc.CallOption) (*flowv1.SubmitTaskResponse, error) {
	out := &flowv1.SubmitTaskResponse{
		TaskIds: []*flowv1.UUID{{Id: uuid.NewString()}},
	}
	return out, nil
}

func (c *MockFlowClient) ListTasks(ctx context.Context, in *flowv1.ListTasksRequest, opts ...grpc.CallOption) (*flowv1.ListTasksResponse, error) {
	out := &flowv1.ListTasksResponse{
		Tasks: []*flowv1.Task{},
	}
	return out, nil
}

func (c *MockFlowClient) GetTasksByIDs(ctx context.Context, in *flowv1.GetTasksByIDsRequest, opts ...grpc.CallOption) (*flowv1.GetTasksByIDsResponse, error) {
	out := &flowv1.GetTasksByIDsResponse{
		Tasks: []*flowv1.Task{},
	}
	if in != nil {
		for _, taskID := range in.GetTaskIds() {
			out.Tasks = append(out.Tasks, &flowv1.Task{
				Id: taskID,
			})
		}
	}
	return out, nil
}

func (c *MockFlowClient) CancelTask(ctx context.Context, in *flowv1.CancelTaskRequest, opts ...grpc.CallOption) (*flowv1.CancelTaskResponse, error) {
	out := &flowv1.CancelTaskResponse{}
	if in != nil && in.GetTaskId() != nil {
		out.Task = &flowv1.Task{
			Id:     in.GetTaskId(),
			Status: flowv1.TaskStatus_TASK_STATUS_TERMINATED,
		}
	}
	return out, nil
}

/* Operation rule mock methods */
func (c *MockFlowClient) CreateOperationRule(ctx context.Context, in *flowv1.CreateOperationRuleRequest, opts ...grpc.CallOption) (*flowv1.CreateOperationRuleResponse, error) {
	out := &flowv1.CreateOperationRuleResponse{
		Id: &flowv1.UUID{Id: uuid.NewString()},
	}
	return out, nil
}

func (c *MockFlowClient) UpdateOperationRule(ctx context.Context, in *flowv1.UpdateOperationRuleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockFlowClient) DeleteOperationRule(ctx context.Context, in *flowv1.DeleteOperationRuleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockFlowClient) GetOperationRule(ctx context.Context, in *flowv1.GetOperationRuleRequest, opts ...grpc.CallOption) (*flowv1.OperationRule, error) {
	out := &flowv1.OperationRule{}
	return out, nil
}

func (c *MockFlowClient) ListOperationRules(ctx context.Context, in *flowv1.ListOperationRulesRequest, opts ...grpc.CallOption) (*flowv1.ListOperationRulesResponse, error) {
	out := &flowv1.ListOperationRulesResponse{
		Rules:      []*flowv1.OperationRule{},
		TotalCount: 0,
	}
	return out, nil
}

func (c *MockFlowClient) SetRuleAsDefault(ctx context.Context, in *flowv1.SetRuleAsDefaultRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

/* Rack-rule association mock methods */
func (c *MockFlowClient) AssociateRuleWithRack(ctx context.Context, in *flowv1.AssociateRuleWithRackRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockFlowClient) DisassociateRuleFromRack(ctx context.Context, in *flowv1.DisassociateRuleFromRackRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	return out, nil
}

func (c *MockFlowClient) GetRackRuleAssociation(ctx context.Context, in *flowv1.GetRackRuleAssociationRequest, opts ...grpc.CallOption) (*flowv1.GetRackRuleAssociationResponse, error) {
	out := &flowv1.GetRackRuleAssociationResponse{}
	return out, nil
}

func (c *MockFlowClient) ListRackRuleAssociations(ctx context.Context, in *flowv1.ListRackRuleAssociationsRequest, opts ...grpc.CallOption) (*flowv1.ListRackRuleAssociationsResponse, error) {
	out := &flowv1.ListRackRuleAssociationsResponse{
		Associations: []*flowv1.RackRuleAssociation{},
	}
	return out, nil
}

// NewMockFlowClient creates a new mock FlowClient that can be used with FlowAtomicClient.SwapClient
func NewMockFlowClient() *FlowClient {
	return &FlowClient{
		flow: &MockFlowClient{},
	}
}
