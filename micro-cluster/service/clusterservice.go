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
 *                                                                            *
 ******************************************************************************/

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/common/errors"
	"net/http"
	"strconv"
	"time"

	"github.com/pingcap-inc/tiem/micro-cluster/service/user/adapt"

	log "github.com/sirupsen/logrus"

	"github.com/pingcap-inc/tiem/micro-cluster/platform/config"

	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/changefeed"
	clusterLog "github.com/pingcap-inc/tiem/micro-cluster/cluster/log"
	clusterManager "github.com/pingcap-inc/tiem/micro-cluster/cluster/management"
	clusterParameter "github.com/pingcap-inc/tiem/micro-cluster/cluster/parameter"
	"github.com/pingcap-inc/tiem/micro-cluster/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/micro-cluster/parametergroup"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager"
	"github.com/pingcap-inc/tiem/workflow"

	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	user "github.com/pingcap-inc/tiem/micro-cluster/service/user/application"

	"github.com/pingcap-inc/tiem/library/framework"
	userDomain "github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
)

var TiEMClusterServiceName = "go.micro.tiem.cluster"

var SuccessResponseStatus = &clusterpb.ResponseStatusDTO{Code: 0}
var BizErrorResponseStatus = &clusterpb.ResponseStatusDTO{Code: 500}

type ClusterServiceHandler struct {
	resourceManager         *resourcemanager.ResourceManager
	authManager             *user.AuthManager
	tenantManager           *user.TenantManager
	userManager             *user.UserManager
	changeFeedManager       *changefeed.Manager
	parameterGroupManager   *parametergroup.Manager
	clusterParameterManager *clusterParameter.Manager
	clusterManager          *clusterManager.Manager
	systemConfigManager     *config.SystemConfigManager
	brManager               backuprestore.BRService
	importexportManager     importexport.ImportExportService
	clusterLogManager       *clusterLog.Manager
}

func handleRequest(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse, requestBody interface{}) bool {
	err := json.Unmarshal([]byte(req.GetRequest()), requestBody)
	if err != nil {
		errMsg := fmt.Sprintf("unmarshal request error, request = %s, err = %s", req.GetRequest(), err.Error())
		handleResponse(ctx, resp, framework.NewTiEMErrorf(common.TIEM_UNMARSHAL_ERROR, errMsg), nil, nil)
		return false
	} else {
		return true
	}
}

func handleResponse(ctx context.Context, resp *clusterpb.RpcResponse, err error, responseData interface{}, page *clusterpb.RpcPage) {
	if err == nil {
		data, getDataError := json.Marshal(responseData)
		if getDataError != nil {
			// deal with err uniformly later
			err = errors.WrapError(errors.TIEM_MARSHAL_ERROR, fmt.Sprintf("marshal request data error, data = %v", responseData), getDataError)
		} else {
			// handle data and page
			resp.Code = int32(common.TIEM_SUCCESS)
			resp.Response = string(data)
			if page != nil {
				resp.Page = page
			}
			return
		}
	}

	if err != nil {
		if finalError, ok := err.(framework.TiEMError); ok {
			framework.LogWithContext(ctx).Errorf("rpc method failed with error, %s", err.Error())
			resp.Code = int32(finalError.GetCode())
			resp.Message = finalError.GetMsg()
			return
		}
		if finalError, ok := err.(errors.EMError); ok {
			framework.LogWithContext(ctx).Errorf("rpc method failed with error, %s", err.Error())
			resp.Code = int32(finalError.GetCode())
			resp.Message = finalError.GetMsg()
			return
		} else {
			resp.Code = int32(errors.TIEM_UNRECOGNIZED_ERROR)
			resp.Message = err.Error()
		}

		return
	}
}

func getLoggerWithContext(ctx context.Context) *log.Entry {
	return framework.LogWithContext(ctx)
}

func handleMetrics(start time.Time, funcName string, code int) {
	duration := time.Since(start)
	framework.Current.GetMetrics().MicroDurationHistogramMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel:  funcName,
		metrics.CodeLabel:    strconv.Itoa(code)}).
		Observe(duration.Seconds())
	framework.Current.GetMetrics().MicroRequestsCounterMetric.With(prometheus.Labels{
		metrics.ServiceLabel: framework.Current.GetServiceMeta().ServiceName.ServerName(),
		metrics.MethodLabel:  funcName,
		metrics.CodeLabel:    strconv.Itoa(code)}).
		Inc()
}

func NewClusterServiceHandler(fw *framework.BaseFramework) *ClusterServiceHandler {
	handler := new(ClusterServiceHandler)
	handler.resourceManager = resourcemanager.NewResourceManager()
	handler.userManager = user.NewUserManager(adapt.MicroMetaDbRepo{})
	handler.tenantManager = user.NewTenantManager(adapt.MicroMetaDbRepo{})
	handler.authManager = user.NewAuthManager(handler.userManager, adapt.MicroMetaDbRepo{})
	handler.changeFeedManager = changefeed.NewManager()
	handler.parameterGroupManager = parametergroup.NewManager()
	handler.clusterParameterManager = clusterParameter.NewManager()
	handler.clusterManager = clusterManager.NewClusterManager()
	handler.systemConfigManager = config.NewSystemConfigManager()
	handler.brManager = backuprestore.GetBRService()
	handler.importexportManager = importexport.GetImportExportService()
	handler.clusterLogManager = clusterLog.NewManager()

	return handler
}

func (handler *ClusterServiceHandler) CreateChangeFeedTask(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CreateChangeFeedTask", int(resp.GetCode()))
	request := cluster.CreateChangeFeedTaskReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.changeFeedManager.Create(ctx, request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) PauseChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "PauseChangeFeedTask", int(response.GetCode()))
	panic("implement me")
}

func (handler *ClusterServiceHandler) ResumeChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ResumeChangeFeedTask", int(response.GetCode()))
	panic("implement me")
}

func (handler *ClusterServiceHandler) DeleteChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteChangeFeedTask", int(response.GetCode()))
	panic("implement me")
}

func (handler *ClusterServiceHandler) UpdateChangeFeedTask(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateChangeFeedTask", int(response.GetCode()))
	panic("implement me")
}

func (handler *ClusterServiceHandler) QueryChangeFeedTasks(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryChangeFeedTasks", int(response.GetCode()))
	panic("implement me")
}

func (handler *ClusterServiceHandler) CreateParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CreateParameterGroup", int(resp.GetCode()))
	request := &message.CreateParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.CreateParameterGroup(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateParameterGroup", int(resp.GetCode()))
	request := &message.UpdateParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.UpdateParameterGroup(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) DeleteParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteParameterGroup", int(resp.GetCode()))
	request := &message.DeleteParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.DeleteParameterGroup(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryParameterGroup", int(resp.GetCode()))
	request := &message.QueryParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.parameterGroupManager.QueryParameterGroup(ctx, *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (handler *ClusterServiceHandler) DetailParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DetailParameterGroup", int(resp.GetCode()))
	request := &message.DetailParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.DetailParameterGroup(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) ApplyParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ApplyParameterGroup", int(resp.GetCode()))
	request := &message.ApplyParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.ApplyParameterGroup(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) CopyParameterGroup(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CopyParameterGroup", int(resp.GetCode()))
	request := &message.CopyParameterGroupReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.parameterGroupManager.CopyParameterGroup(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryClusterParameters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryClusterParameters", int(resp.GetCode()))
	request := &cluster.QueryClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.clusterParameterManager.QueryClusterParameters(ctx, *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (handler *ClusterServiceHandler) UpdateClusterParameters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateClusterParameters", int(resp.GetCode()))
	request := &cluster.UpdateClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.UpdateClusterParameters(ctx, *request, true)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) InspectClusterParameters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "InspectClusterParameters", int(resp.GetCode()))
	request := &cluster.InspectClusterParametersReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := handler.clusterParameterManager.InspectClusterParameters(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (handler *ClusterServiceHandler) QueryClusterLog(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryClusterLog", int(resp.GetCode()))
	request := &cluster.QueryClusterLogReq{}

	if handleRequest(ctx, req, resp, request) {
		result, page, err := handler.clusterLogManager.QueryClusterLog(ctx, *request)
		handleResponse(ctx, resp, err, result, page)
	}
	return nil
}

func (c ClusterServiceHandler) CreateCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "CreateCluster", int(resp.GetCode()))

	request := cluster.CreateClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.CreateCluster(ctx, request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) RestoreNewCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "RestoreNewCluster", int(resp.GetCode()))

	request := cluster.RestoreNewClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.RestoreNewCluster(ctx, request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ScaleOutCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ScaleOutCluster", int(resp.GetCode()))

	request := cluster.ScaleOutClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.ScaleOut(ctx, request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) ScaleInCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ScaleInCluster", int(resp.GetCode()))

	request := cluster.ScaleInClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.ScaleIn(ctx, request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) CloneCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CloneCluster", int(resp.GetCode()))

	request := cluster.CloneClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := handler.clusterManager.Clone(ctx, request)

		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) TakeoverClusters(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "TakeoverClusters", int(resp.GetCode()))
	// todo takeover
	return nil
}

func (c ClusterServiceHandler) QueryCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "QueryCluster", int(resp.GetCode()))

	request := cluster.QueryClustersReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, total, err := c.clusterManager.QueryCluster(ctx, request)
		handleResponse(ctx, resp, err, result, &clusterpb.RpcPage{
			Page:     int32(request.Page),
			PageSize: int32(request.PageSize),
			Total:    int32(total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) DeleteCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteCluster", int(resp.GetCode()))

	request := cluster.DeleteClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.DeleteCluster(ctx, request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) RestartCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "RestartCluster", int(resp.GetCode()))

	request := cluster.RestartClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.RestartCluster(ctx, request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) StopCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "StopCluster", int(resp.GetCode()))

	request := cluster.StopClusterReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.StopCluster(ctx, request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) DetailCluster(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "DetailCluster", int(resp.GetCode()))

	request := cluster.QueryClusterDetailReq{}

	if handleRequest(ctx, req, resp, &request) {
		result, err := c.clusterManager.DetailCluster(ctx, request)
		handleResponse(ctx, resp, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) ExportData(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ExportData", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("export data")
	exportReq := message.DataExportReq{}

	if handleRequest(ctx, request, response, &exportReq) {
		result, err := c.importexportManager.ExportData(ctx, exportReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) ImportData(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ImportData", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("import data")
	importReq := message.DataImportReq{}

	if handleRequest(ctx, request, response, &importReq) {
		result, err := c.importexportManager.ImportData(ctx, importReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryDataTransport(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryDataTransport", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("query data transport")
	queryReq := message.QueryDataImportExportRecordsReq{}

	if handleRequest(ctx, request, response, &queryReq) {
		result, page, err := c.importexportManager.QueryDataTransportRecords(ctx, queryReq)
		handleResponse(ctx, response, err, result, &clusterpb.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) DeleteDataTransportRecord(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteDataTransportRecord", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("delete data transport record")
	deleteReq := message.DeleteImportExportRecordReq{}

	if handleRequest(ctx, request, response, &deleteReq) {
		result, err := c.importexportManager.DeleteDataTransportRecord(ctx, deleteReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c *ClusterServiceHandler) GetSystemConfig(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetSystemConfig", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("get system config")
	getReq := message.GetSystemConfigReq{}

	if handleRequest(ctx, request, response, &getReq) {
		result, err := c.systemConfigManager.GetSystemConfig(ctx, getReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) CreateBackup(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "CreateBackup", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("create backup")
	backupReq := cluster.BackupClusterDataReq{}

	if handleRequest(ctx, request, response, &backupReq) {
		result, err := c.brManager.BackupCluster(ctx, backupReq, true)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) DeleteBackupRecords(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteBackupRecord", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("delete backup records")
	deleteReq := cluster.DeleteBackupDataReq{}

	if handleRequest(ctx, request, response, &deleteReq) {
		result, err := c.brManager.DeleteBackupRecords(ctx, deleteReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) SaveBackupStrategy(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "SaveBackupStrategy", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("save backup strategy")
	saveReq := cluster.SaveBackupStrategyReq{}

	if handleRequest(ctx, request, response, &saveReq) {
		result, err := c.brManager.SaveBackupStrategy(ctx, saveReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) GetBackupStrategy(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetBackupStrategy", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("get backup strategy")
	getReq := cluster.GetBackupStrategyReq{}

	if handleRequest(ctx, request, response, &getReq) {
		result, err := c.brManager.GetBackupStrategy(ctx, getReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) QueryBackupRecords(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "QueryBackupRecords", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("query backup records")
	queryReq := cluster.QueryBackupRecordsReq{}

	if handleRequest(ctx, request, response, &queryReq) {
		result, page, err := c.brManager.QueryClusterBackupRecords(ctx, queryReq)
		handleResponse(ctx, response, err, result, &clusterpb.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c ClusterServiceHandler) GetDashboardInfo(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "DescribeDashboard", int(response.GetCode()))
	framework.LogWithContext(ctx).Info("get cluster dashboard info")
	dashboardReq := cluster.GetDashboardInfoReq{}

	if handleRequest(ctx, request, response, &dashboardReq) {
		result, err := c.clusterManager.GetClusterDashboardInfo(ctx, dashboardReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

func (c ClusterServiceHandler) GetMonitorInfo(ctx context.Context, req *clusterpb.RpcRequest, resp *clusterpb.RpcResponse) (err error) {
	start := time.Now()
	defer handleMetrics(start, "GetMonitorInfo", int(resp.GetCode()))
	request := &cluster.QueryMonitorInfoReq{}

	if handleRequest(ctx, req, resp, request) {
		result, err := c.clusterManager.GetMonitorInfo(ctx, *request)
		handleResponse(ctx, resp, err, result, nil)
	}
	return nil
}

func (c ClusterServiceHandler) ListFlows(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	framework.LogWithContext(ctx).Info("list flows")
	start := time.Now()
	defer handleMetrics(start, "ListFlows", int(response.GetCode()))

	listReq := message.QueryWorkFlowsReq{}
	if handleRequest(ctx, request, response, &listReq) {
		manager := workflow.GetWorkFlowService()
		result, page, err := manager.ListWorkFlows(ctx, listReq)
		handleResponse(ctx, response, err, result, &clusterpb.RpcPage{
			Page:     int32(page.Page),
			PageSize: int32(page.PageSize),
			Total:    int32(page.Total),
		})
	}

	return nil
}

func (c *ClusterServiceHandler) DetailFlow(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	framework.LogWithContext(ctx).Info("detail flow")
	start := time.Now()
	defer handleMetrics(start, "DetailFlow", int(response.GetCode()))

	detailReq := message.QueryWorkFlowDetailReq{}
	if handleRequest(ctx, request, response, &detailReq) {
		manager := workflow.GetWorkFlowService()
		result, err := manager.DetailWorkFlow(ctx, detailReq)
		handleResponse(ctx, response, err, result, nil)
	}

	return nil
}

var ManageSuccessResponseStatus = &clusterpb.ManagerResponseStatus{
	Code: 0,
}

func (p *ClusterServiceHandler) Login(ctx context.Context, req *clusterpb.LoginRequest, resp *clusterpb.LoginResponse) error {
	log := framework.LogWithContext(ctx).WithField("fp", "ClusterServiceHandler.Login")
	/*
		start := time.Now()
		defer handleMetrics(start, "Login", int(resp.GetCode()))*/
	log.Debug("req:", req)
	token, err := p.authManager.Login(ctx, req.GetAccountName(), req.GetPassword())

	if err != nil {
		resp.Status = &clusterpb.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
		log.Error("resp:", resp)
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.TokenString = token
		log.Debug("resp:", resp)
	}
	return nil

}

func (p *ClusterServiceHandler) Logout(ctx context.Context, req *clusterpb.LogoutRequest, resp *clusterpb.LogoutResponse) error {
	accountName, err := p.authManager.Logout(ctx, req.TokenString)
	if err != nil {
		resp.Status = &clusterpb.ManagerResponseStatus{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
		resp.Status.Message = err.Error()
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.AccountName = accountName
	}
	return nil

}

func (p *ClusterServiceHandler) VerifyIdentity(ctx context.Context, req *clusterpb.VerifyIdentityRequest, resp *clusterpb.VerifyIdentityResponse) error {
	tenantId, accountId, accountName, err := p.authManager.Accessible(ctx, req.GetAuthType(), req.GetPath(), req.GetTokenString())

	if err != nil {
		if _, ok := err.(*userDomain.UnauthorizedError); ok {
			resp.Status = &clusterpb.ManagerResponseStatus{
				Code:    http.StatusUnauthorized,
				Message: "未登录或登录失效，请重试",
			}
		} else if _, ok := err.(*userDomain.ForbiddenError); ok {
			resp.Status = &clusterpb.ManagerResponseStatus{
				Code:    http.StatusForbidden,
				Message: "无权限",
			}
		} else {
			resp.Status = &clusterpb.ManagerResponseStatus{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
	} else {
		resp.Status = ManageSuccessResponseStatus
		resp.TenantId = tenantId
		resp.AccountId = accountId
		resp.AccountName = accountName
	}

	return nil
}

func (handler *ClusterServiceHandler) ImportHosts(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "ImportHosts", int(response.GetCode()))
	reqStruct := message.ImportHostsReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		hostIds, err := handler.resourceManager.ImportHosts(ctx, reqStruct.Hosts)
		var rsp message.ImportHostsResp
		if err == nil {
			rsp.HostIDS = hostIds
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) DeleteHosts(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "DeleteHosts", int(response.GetCode()))
	reqStruct := message.DeleteHostsReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		err := handler.resourceManager.DeleteHosts(ctx, reqStruct.HostIDs)
		var rsp message.DeleteHostsResp
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) QueryHosts(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "QueryHosts", int(response.GetCode()))
	reqStruct := message.QueryHostsReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		filter := reqStruct.GetHostFilter()
		page := reqStruct.GetPage()

		hosts, err := handler.resourceManager.QueryHosts(ctx, filter, page)
		var rsp message.QueryHostsResp
		if err == nil {
			rsp.Hosts = hosts
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostReserved(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateHostReserved", int(response.GetCode()))
	reqStruct := message.UpdateHostReservedReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		err := handler.resourceManager.UpdateHostReserved(ctx, reqStruct.HostIDs, reqStruct.Reserved)
		var rsp message.UpdateHostReservedResp
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) UpdateHostStatus(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "UpdateHostStatus", int(response.GetCode()))
	reqStruct := message.UpdateHostStatusReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		err := handler.resourceManager.UpdateHostStatus(ctx, reqStruct.HostIDs, reqStruct.Status)
		var rsp message.UpdateHostStatusResp
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetHierarchy(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetHierarchy", int(response.GetCode()))
	reqStruct := message.GetHierarchyReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		filter := reqStruct.GetHostFilter()

		root, err := handler.resourceManager.GetHierarchy(ctx, filter, reqStruct.Level, reqStruct.Depth)
		var rsp message.GetHierarchyResp
		if err == nil {
			rsp.Root = *root
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}

func (handler *ClusterServiceHandler) GetStocks(ctx context.Context, request *clusterpb.RpcRequest, response *clusterpb.RpcResponse) error {
	start := time.Now()
	defer handleMetrics(start, "GetStocks", int(response.GetCode()))
	reqStruct := message.GetStocksReq{}

	if handleRequest(ctx, request, response, &reqStruct) {
		location := reqStruct.GetLocation()
		hostFilter := reqStruct.GetHostFilter()
		diskFilter := reqStruct.GetDiskFilter()

		stocks, err := handler.resourceManager.GetStocks(ctx, location, hostFilter, diskFilter)
		var rsp message.GetStocksResp
		if err == nil {
			rsp.Stocks = *stocks
		}
		handleResponse(ctx, response, err, rsp, nil)
	}

	return nil
}
