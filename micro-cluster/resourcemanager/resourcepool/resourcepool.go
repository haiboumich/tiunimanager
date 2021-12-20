/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 ******************************************************************************/

package resourcepool

import (
	"context"
	"sync"

	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostinitiator"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/resourcepool/hostprovider"
)

type ResourcePool struct {
	hostProvider  hostprovider.HostProvider
	hostInitiator hostinitiator.HostInitiator
	// cloudHostProvider hostprovider.HostProvider
}

var resourcePool *ResourcePool
var once sync.Once

func GetResourcePool() *ResourcePool {
	once.Do(func() {
		if resourcePool == nil {
			resourcePool = new(ResourcePool)
			resourcePool.InitResourcePool()
		}
	})
	return resourcePool
}

func (p *ResourcePool) InitResourcePool() {
	p.hostProvider = hostprovider.NewFileHostProvider()
	p.hostInitiator = hostinitiator.NewFileHostInitiator()
}

func (p *ResourcePool) GetHostProvider() hostprovider.HostProvider {
	return p.hostProvider
}

func (p *ResourcePool) SetHostInitiator(initiator hostinitiator.HostInitiator) {
	p.hostInitiator = initiator
}

func (p *ResourcePool) ImportHosts(ctx context.Context, hosts []structs.HostInfo) (hostIds []string, err error) {
	for _, host := range hosts {
		err = p.verify(ctx, &host)
		if err != nil {
			return nil, err
		}
	}
	return p.hostProvider.ImportHosts(ctx, hosts)
}

func (p *ResourcePool) DeleteHosts(ctx context.Context, hostIds []string) (err error) {
	return p.hostProvider.DeleteHosts(ctx, hostIds)
}

func (p *ResourcePool) QueryHosts(ctx context.Context, filter *structs.HostFilter, page *structs.PageRequest) (hosts []structs.HostInfo, err error) {
	return p.hostProvider.QueryHosts(ctx, filter, page)
}

func (p *ResourcePool) UpdateHostStatus(ctx context.Context, hostIds []string, status string) (err error) {
	return p.hostProvider.UpdateHostStatus(ctx, hostIds, status)
}

func (p *ResourcePool) UpdateHostReserved(ctx context.Context, hostIds []string, reserved bool) (err error) {
	return p.hostProvider.UpdateHostReserved(ctx, hostIds, reserved)
}

func (p *ResourcePool) GetHierarchy(ctx context.Context, filter *structs.HostFilter, level int, depth int) (root *structs.HierarchyTreeNode, err error) {
	return p.hostProvider.GetHierarchy(ctx, filter, level, depth)
}

func (p *ResourcePool) GetStocks(ctx context.Context, location *structs.Location, hostFilter *structs.HostFilter, diskFilter *structs.DiskFilter) (stocks *structs.Stocks, err error) {
	return p.hostProvider.GetStocks(ctx, location, hostFilter, diskFilter)
}

func (p *ResourcePool) verify(ctx context.Context, h *structs.HostInfo) (err error) {
	client, err := p.hostInitiator.VerifyConnect(ctx, h)
	if err != nil {
		return err
	}
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	if err = p.hostInitiator.VerifyCpuMem(ctx, client, h); err != nil {
		return err
	}

	if err = p.hostInitiator.VerifyDisks(ctx, client, h); err != nil {
		return err
	}

	if err = p.hostInitiator.VerifyFS(ctx, client, h); err != nil {
		return err
	}

	if err = p.hostInitiator.VerifySwap(ctx, client, h); err != nil {
		return err
	}

	if err = p.hostInitiator.VerifyEnv(ctx, client, h); err != nil {
		return err
	}

	if err = p.hostInitiator.VerifyOSEnv(ctx, client, h); err != nil {
		return err
	}

	return nil
}
