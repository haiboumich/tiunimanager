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

package hostresource

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"

	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/micro-api/controller"

	"github.com/pingcap-inc/tiem/message"
)

func importExcelFile(r io.Reader, reserved bool) ([]structs.HostInfo, error) {
	xlsx, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	rows := xlsx.GetRows("Host Information")
	var hosts []structs.HostInfo
	for irow, row := range rows {
		if irow > 0 {
			var host structs.HostInfo
			host.Reserved = reserved
			host.HostName = row[HOSTNAME_FIELD]
			addr := net.ParseIP(row[IP_FILED])
			if addr == nil {
				errMsg := fmt.Sprintf("Row %d has a Invalid IP Address %s", irow, row[IP_FILED])
				return nil, errors.New(errMsg)
			}
			host.IP = addr.String()
			host.UserName = row[USERNAME_FIELD]
			host.Passwd = row[PASSWD_FIELD]
			host.Region = row[REGION_FIELD]
			host.AZ = row[ZONE_FIELD]
			host.Rack = row[RACK_FIELD]
			if err = constants.ValidArchType(row[ARCH_FIELD]); err != nil {
				errMsg := fmt.Sprintf("Row %d get arch(%s) failed, %v", irow, row[ARCH_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.Arch = row[ARCH_FIELD]
			host.OS = row[OS_FIELD]
			host.Kernel = row[KERNEL_FIELD]
			coreNum, err := (strconv.Atoi(row[CPU_FIELD]))
			if err != nil {
				errMsg := fmt.Sprintf("Row %d get coreNum(%s) failed, %v", irow, row[CPU_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.CpuCores = int32(coreNum)
			host.FreeCpuCores = host.CpuCores
			mem, err := (strconv.Atoi(row[MEM_FIELD]))
			if err != nil {
				errMsg := fmt.Sprintf("Row %d get memory(%s) failed, %v", irow, row[MEM_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.Memory = int32(mem)
			host.FreeMemory = host.Memory
			host.Nic = row[NIC_FIELD]

			if err = constants.ValidProductName(row[CLUSTER_TYPE_FIELD]); err != nil {
				errMsg := fmt.Sprintf("Row %d get cluster type(%s) failed, %v", irow, row[CLUSTER_TYPE_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.ClusterType = row[CLUSTER_TYPE_FIELD]
			if err = host.AddTraits(host.ClusterType); err != nil {
				return nil, err
			}

			host.Purpose = row[PURPOSE_FIELD]
			purposes := host.GetPurposes()
			for _, p := range purposes {
				if err = constants.ValidPurposeType(p); err != nil {
					errMsg := fmt.Sprintf("Row %d get purpose(%s) failed, %v", irow, p, err)
					return nil, errors.New(errMsg)
				}
				if err = host.AddTraits(p); err != nil {
					return nil, err
				}
			}

			if err = constants.ValidDiskType(row[DISKTYPE_FIELD]); err != nil {
				errMsg := fmt.Sprintf("Row %d get disk type(%s) failed, %v", irow, row[DISKTYPE_FIELD], err)
				return nil, errors.New(errMsg)
			}
			host.DiskType = row[DISKTYPE_FIELD]
			if err = host.AddTraits(host.DiskType); err != nil {
				return nil, err
			}
			host.Status = string(constants.HostOnline)
			host.Stat = string(constants.HostLoadLoadLess)
			disksStr := row[DISKS_FIELD]
			if err = json.Unmarshal([]byte(disksStr), &host.Disks); err != nil {
				errMsg := fmt.Sprintf("Row %d has a Invalid Disk Json Format, %v", irow, err)
				return nil, errors.New(errMsg)
			}
			for i := range host.Disks {
				if host.Disks[i].Type == "" {
					host.Disks[i].Type = host.DiskType
				}
			}
			hosts = append(hosts, host)
		}
	}
	return hosts, nil
}

// ImportHosts godoc
// @Summary Import a batch of hosts to TiEM
// @Description import hosts by xlsx file
// @Tags resource
// @Accept mpfd
// @Produce json
// @Security ApiKeyAuth
// @Param hostReserved formData string false "whether hosts are reserved(won't be allocated) after import" default(false)
// @Param file formData file true "hosts information in a xlsx file"
// @Success 200 {object} controller.CommonResult{data=message.ImportHostsResp}
// @Router /resources/hosts [post]
func ImportHosts(c *gin.Context) {
	reservedStr := c.DefaultPostForm("hostReserved", "false")
	reserved, err := strconv.ParseBool(reservedStr)
	if err != nil {
		errmsg := fmt.Sprintf("GetFormData Error: %v", err)
		c.JSON(http.StatusBadRequest, controller.Fail(common.TIEM_PARAMETER_INVALID.GetHttpCode(), errmsg))
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		errmsg := fmt.Sprintf("GetFormFile Error: %v", err)
		c.JSON(http.StatusBadRequest, controller.Fail(common.TIEM_PARAMETER_INVALID.GetHttpCode(), errmsg))
		return
	}
	hosts, err := importExcelFile(file, reserved)
	if err != nil {
		errmsg := fmt.Sprintf("Import File Error: %v", err)
		c.JSON(http.StatusInternalServerError, controller.Fail(common.TIEM_PARAMETER_INVALID.GetHttpCode(), errmsg))
		return
	}

	requestBody, err := controller.HandleJsonRequestWithBuiltReq(c, message.ImportHostsReq{
		Hosts: hosts,
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.ImportHosts,
			&message.ImportHostsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// ListHost godoc
// @Summary Show all hosts list in TiEM
// @Description get hosts list
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param hostQuery query message.QueryHostsReq false "list condition"
// @Success 200 {object} controller.ResultWithPage{data=message.QueryHostsResp}
// @Router /resources/hosts [get]
func QueryHosts(c *gin.Context) {
	var req message.QueryHostsReq

	requestBody, err := controller.HandleJsonRequestFromQuery(c, &req)

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryHosts, &message.QueryHostsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

func detectDuplicateElement(hostIds []string) (string, bool) {
	temp := map[string]struct{}{}
	hasDuplicate := false
	var duplicateStr string
	for _, item := range hostIds {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
		} else {
			hasDuplicate = true
			duplicateStr = item
			break
		}
	}
	return duplicateStr, hasDuplicate
}

// RemoveHosts godoc
// @Summary Remove a batch of hosts
// @Description remove hosts by a list
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param hostIds body message.DeleteHostsReq true "list of host IDs"
// @Success 200 {object} controller.CommonResult{data=message.DeleteHostsResp}
// @Router /resources/hosts/ [delete]
func RemoveHosts(c *gin.Context) {
	var req message.DeleteHostsReq

	requestBody, err := controller.HandleJsonRequestFromBody(c, &req)
	if str, dup := detectDuplicateElement(req.HostIDs); dup {
		c.JSON(http.StatusBadRequest, controller.Fail(common.TIEM_PARAMETER_INVALID.GetHttpCode(), str+" Is Duplicated in request"))
		return
	}

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteHosts, &message.DeleteHostsResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// DownloadHostTemplateFile godoc
// @Summary Download the host information template file for importing
// @Description get host template xlsx file
// @Tags resource
// @Accept json
// @Produce octet-stream
// @Security ApiKeyAuth
// @Success 200 {file} file
// @Router /resources/hosts-template/ [get]
func DownloadHostTemplateFile(c *gin.Context) {
	curDir, _ := os.Getwd()
	templateName := common.TemplateFileName
	// The template file should be on tiem/etc/hostInfo_template.xlsx
	filePath := filepath.Join(curDir, common.TemplateFilePath, templateName)

	_, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		c.JSON(http.StatusInternalServerError, controller.Fail(common.TIEM_RESOURCE_TEMPLATE_FILE_NOT_FOUND.GetHttpCode(), err.Error()))
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+templateName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.File(filePath)
}

// UpdateHostReserved godoc
// @Summary Update host reserved
// @Description update host reserved by a list
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param updateReq body message.UpdateHostReservedReq true "do update in host list"
// @Success 200 {object} controller.CommonResult{data=message.UpdateHostReservedResp}
// @Router /resources/host-reserved [put]
func UpdateHostReserved(c *gin.Context) {
	var req message.UpdateHostReservedReq

	requestBody, err := controller.HandleJsonRequestFromBody(c, &req)
	if str, dup := detectDuplicateElement(req.HostIDs); dup {
		c.JSON(http.StatusBadRequest, controller.Fail(common.TIEM_PARAMETER_INVALID.GetHttpCode(), str+" Is Duplicated in request"))
		return
	}

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateHostReserved, &message.UpdateHostReservedResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// UpdateHostStatus godoc
// @Summary Update host status
// @Description update host status by a list
// @Tags resource
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param updateReq body message.UpdateHostStatusReq true "do update in host list"
// @Success 200 {object} controller.CommonResult{data=message.UpdateHostStatusResp}
// @Router /resources/host-status [put]
func UpdateHostStatus(c *gin.Context) {
	var req message.UpdateHostStatusReq

	requestBody, err := controller.HandleJsonRequestFromBody(c, &req)
	if str, dup := detectDuplicateElement(req.HostIDs); dup {
		c.JSON(http.StatusBadRequest, controller.Fail(common.TIEM_PARAMETER_INVALID.GetHttpCode(), str+" Is Duplicated in request"))
		return
	}

	if !constants.HostStatus(req.Status).IsValidStatus() {
		errmsg := fmt.Sprintf("input status %s is invalid, [Online,Offline,Deleted]", req.Status)
		c.JSON(http.StatusBadRequest, controller.Fail(common.TIEM_PARAMETER_INVALID.GetHttpCode(), errmsg))
		return
	}

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateHostStatus, &message.UpdateHostStatusResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
