/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: status
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/20
*******************************************************************************/

package deployment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/pingcap-inc/tiem/util/disk"

	"gorm.io/gorm"
)

type Status string

const (
	Init       Status = "init"
	Processing Status = "processing"
	Finished   Status = "finished"
	Error      Status = "error"
)

// Operation Record information about each TiUP operation
type Operation struct {
	ID         string    `gorm:"primaryKey;"`
	Type       string    `gorm:"not null;comment:'second party operation of type, eg: deploy, start, stop...'"`
	WorkFlowID string    `gorm:"not null;index;comment:'workflow ID of operation'"`
	Status     Status    `gorm:"default:null"`
	Result     string    `gorm:"default:null"`
	ErrorStr   string    `gorm:"size:8192;comment:'operation error msg'"`
	CreatedAt  time.Time `gorm:"<-:create"`
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func Create(tiUPHome string, op Operation) (fileName string, err error) {
	b, err := json.Marshal(op)
	if err != nil {
		return "", err
	}

	fileName, err = disk.CreateWithContent(fmt.Sprintf("%s/storage", tiUPHome), "operation", "json", b)
	return
}

func Update(fileName string, op Operation) error {
	b, err := json.Marshal(op)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, b, 0644)
}

func Read(fileName string) (op Operation, err error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &op)
	return
}

func Delete(filename string) error {
	return os.Remove(filename)
}
