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

package constants

import (
	"github.com/pingcap-inc/tiem/common/errors"
)

type ChangeFeedStatus string

const (
	Initial  ChangeFeedStatus = "Initial"
	Normal   ChangeFeedStatus = "Normal"
	Stopped  ChangeFeedStatus = "Stopped"
	Finished ChangeFeedStatus = "Finished"
	Error    ChangeFeedStatus = "Error"
	Failed   ChangeFeedStatus = "Failed"
	Unknown  ChangeFeedStatus = "Unknown"
)

func (s ChangeFeedStatus) IsFinal() bool {
	return Finished == s || Failed == s
}

func (s ChangeFeedStatus) ToString() string {
	return string(s)
}

func IsValidStatus(s string) bool {
	return Initial.ToString() == s ||
		Normal.ToString() == s ||
		Stopped.ToString() == s ||
		Finished.ToString() == s ||
		Error.ToString() == s ||
		Failed.ToString() == s
}

func ConvertStatus(s string) (status ChangeFeedStatus, err error) {
	if IsValidStatus(s) {
		return ChangeFeedStatus(s), nil
	} else {
		return Unknown, errors.NewError(errors.TIEM_PARAMETER_INVALID, "unexpected change feed status")
	}
}

type DownstreamType string

const (
	DownstreamTypeTiDB  DownstreamType = "tidb"
	DownstreamTypeKafka DownstreamType = "kafka"
	DownstreamTypeMysql DownstreamType = "mysql"
)
