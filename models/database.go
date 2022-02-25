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

package models

import (
	"context"
	"github.com/asim/go-micro/v3/util/file"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/models/platform/check"
	"github.com/pingcap-inc/tiem/models/platform/product"
	"github.com/pingcap-inc/tiem/models/platform/system"
	mm "github.com/pingcap-inc/tiem/models/resource/management"
	resourcePool "github.com/pingcap-inc/tiem/models/resource/resourcepool"
	"github.com/pingcap-inc/tiem/models/tiup"
	"github.com/pingcap-inc/tiem/models/user/account"
	"github.com/pingcap-inc/tiem/models/user/identification"
	"github.com/pingcap-inc/tiem/models/user/rbac"
	gormopentracing "gorm.io/plugin/opentracing"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/cluster/backuprestore"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/cluster/parameter"
	"github.com/pingcap-inc/tiem/models/cluster/upgrade"
	"github.com/pingcap-inc/tiem/models/datatransfer/importexport"
	"github.com/pingcap-inc/tiem/models/parametergroup"
	"github.com/pingcap-inc/tiem/models/platform/config"
	"github.com/pingcap-inc/tiem/models/resource"
	resource_rw "github.com/pingcap-inc/tiem/models/resource/gormreadwrite"
	"github.com/pingcap-inc/tiem/models/workflow"
	"github.com/pingcap-inc/tiem/models/workflow/secondparty"
	"gorm.io/driver/sqlite"

	"gorm.io/gorm"
)

var defaultDb *database

type database struct {
	base                             *gorm.DB
	workFlowReaderWriter             workflow.ReaderWriter
	importExportReaderWriter         importexport.ReaderWriter
	brReaderWriter                   backuprestore.ReaderWriter
	changeFeedReaderWriter           changefeed.ReaderWriter
	upgradeReadWriter                upgrade.ReaderWriter
	clusterReaderWriter              management.ReaderWriter
	parameterGroupReaderWriter       parametergroup.ReaderWriter
	clusterParameterReaderWriter     parameter.ReaderWriter
	configReaderWriter               config.ReaderWriter
	secondPartyOperationReaderWriter secondparty.ReaderWriter
	resourceReaderWriter             resource.ReaderWriter
	rbacReaderWriter                 rbac.ReaderWriter
	accountReaderWriter              account.ReaderWriter
	tokenReaderWriter                identification.ReaderWriter
	productReaderWriter              product.ProductReadWriterInterface
	tiUPConfigReaderWriter           tiup.ReaderWriter
	reportReaderWriter               check.ReaderWriter
	systemReaderWriter               system.ReaderWriter
}

func Open(fw *framework.BaseFramework) error {
	dbFilePath := fw.GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName

	dbFileExisted, err := file.Exists(dbFilePath)
	if err != nil {
		return err
	}

	logins := framework.LogForkFile(constants.LogFileSystem)

	db, err := gorm.Open(sqlite.Open(dbFilePath+"?_busy_timeout=60000"), &gorm.Config{})

	if err != nil || db.Error != nil {
		logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFilePath, err, db.Error)
		return err
	} else {
		logins.Infof("open database succeed, filepath: %s", dbFilePath)
	}
	db.Use(gormopentracing.New(
		gormopentracing.WithSqlParameters(false),
		gormopentracing.WithCreateOpName("em.db.create"),
		gormopentracing.WithDeleteOpName("em.db.delete"),
		gormopentracing.WithQueryOpName("em.db.query"),
		gormopentracing.WithRawOpName("em.db.raw"),
		gormopentracing.WithRowOpName("em.db.row"),
		gormopentracing.WithUpdateOpName("em.db.update"),
	))

	defaultDb = &database{
		base: db,
	}

	defaultDb.initReaderWriters()

	err = defaultDb.migrateTables()
	if err != nil {
		return err
	}

	// init data for empty database
	if !dbFileExisted {
		logins.Infof("init default data for new database")
		return allVersionInitializers[0].DataInitializer()
	} else {
		// compatible
		if e := defaultDb.base.First(&system.SystemInfo{}).Error; e != nil {
			defaultDb.base.Create(&system.SystemInfo{
				SystemName:       "EM",
				SystemLogo:       "",
				CurrentVersionID: "",
				LastVersionID:    "",
				State:            constants.SystemInitialing,
			})
		}
		logins.Infof("database is existed, skip init data")
	}

	return nil
}

// IncrementVersionData
// @Description: execute data initializer between originalVersion and targetVersion
// @Parameter originalVersion
// @Parameter targetVersion
// @return error
func IncrementVersionData(originalVersion string, targetVersion string) error {
	if len(targetVersion) == 0 {
		return errors.NewErrorf(errors.TIEM_SYSTEM_INVALID_VERSION, "invalid version %s", targetVersion)
	}

	if originalVersion == targetVersion {
		return nil
	}
	originalVersionIndex := -1
	for i, eachVersion := range allVersionInitializers {
		// match target version before originalVersion, return err
		if originalVersionIndex == -1 && targetVersion == eachVersion.VersionID {
			return errors.NewErrorf(errors.TIEM_SYSTEM_INVALID_VERSION, "unable to upgrade version from %s to %s", originalVersion, targetVersion)
		}
		if originalVersionIndex == -1 && originalVersion == eachVersion.VersionID {
			originalVersionIndex = i
		}

		// execute DataInitializer for versions between originalVersion and targetVersion
		if originalVersionIndex != -1 && i > originalVersionIndex {
			err := eachVersion.DataInitializer()
			if err != nil {
				return err
			}
		}
		// target version reached, break
		if targetVersion == eachVersion.VersionID {
			break
		}
	}

	return nil
}

func (p *database) migrateStream(models ...interface{}) (err error) {
	for _, model := range models {
		err = p.base.AutoMigrate(model)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("init table failed, model = %v, err = %s", models, err.Error())
			return err
		}
	}
	return nil
}

func (p *database) migrateTables() (err error) {
	return p.migrateStream(
		new(system.SystemInfo),
		new(system.VersionInfo),
		new(changefeed.ChangeFeedTask),
		new(workflow.WorkFlow),
		new(workflow.WorkFlowNode),
		new(upgrade.ProductUpgradePath),
		new(management.Cluster),
		new(management.ClusterInstance),
		new(management.ClusterRelation),
		new(management.ClusterTopologySnapshot),
		new(management.DBUser),
		new(importexport.DataTransportRecord),
		new(backuprestore.BackupRecord),
		new(backuprestore.BackupStrategy),
		new(config.SystemConfig),
		new(secondparty.SecondPartyOperation),
		new(parametergroup.Parameter),
		new(parametergroup.ParameterGroup),
		new(parametergroup.ParameterGroupMapping),
		new(parameter.ClusterParameterMapping),
		new(identification.Token),
		new(tiup.TiupConfig),
		new(resourcePool.Host),
		new(resourcePool.Disk),
		new(resourcePool.Label),
		new(mm.UsedCompute),
		new(mm.UsedPort),
		new(mm.UsedDisk),
		new(product.Zone),
		new(product.Spec),
		new(product.Product),
		new(product.ProductComponent),
		new(account.User),
		new(account.Tenant),
		new(account.UserLogin),
		new(account.UserTenantRelation),
		new(check.CheckReport),
	)
}

func (p *database) initReaderWriters() {
	defaultDb.changeFeedReaderWriter = changefeed.NewGormChangeFeedReadWrite(defaultDb.base)
	defaultDb.workFlowReaderWriter = workflow.NewFlowReadWrite(defaultDb.base)
	defaultDb.importExportReaderWriter = importexport.NewImportExportReadWrite(defaultDb.base)
	defaultDb.brReaderWriter = backuprestore.NewBRReadWrite(defaultDb.base)
	defaultDb.upgradeReadWriter = upgrade.NewGormProductUpgradePath(defaultDb.base)
	defaultDb.resourceReaderWriter = resource_rw.NewGormResourceReadWrite(defaultDb.base)
	defaultDb.parameterGroupReaderWriter = parametergroup.NewParameterGroupReadWrite(defaultDb.base)
	defaultDb.clusterParameterReaderWriter = parameter.NewClusterParameterReadWrite(defaultDb.base)
	defaultDb.configReaderWriter = config.NewConfigReadWrite(defaultDb.base)
	defaultDb.secondPartyOperationReaderWriter = secondparty.NewGormSecondPartyOperationReadWrite(defaultDb.base)
	defaultDb.clusterReaderWriter = management.NewClusterReadWrite(defaultDb.base)
	defaultDb.rbacReaderWriter = rbac.NewRBACReadWrite(defaultDb.base)
	defaultDb.accountReaderWriter = account.NewAccountReadWrite(defaultDb.base)
	defaultDb.tokenReaderWriter = identification.NewTokenReadWrite(defaultDb.base)
	defaultDb.productReaderWriter = product.NewProductReadWriter(defaultDb.base)
	defaultDb.tiUPConfigReaderWriter = tiup.NewGormTiupConfigReadWrite(defaultDb.base)
	defaultDb.reportReaderWriter = check.NewReportReadWrite(defaultDb.base)
	defaultDb.systemReaderWriter = system.NewSystemReadWrite(defaultDb.base)
}

func (p *database) initSystemConfig() {
	// system config
	framework.LogWithContext(context.TODO()).Info("begin init system configs to database...")
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupStorageType, ConfigValue: string(constants.StorageTypeS3)})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupStoragePath, ConfigValue: constants.DefaultBackupStoragePath})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3AccessKey, ConfigValue: constants.DefaultBackupS3AccessKey})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3SecretAccessKey, ConfigValue: constants.DefaultBackupS3SecretAccessKey})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupS3Endpoint, ConfigValue: constants.DefaultBackupS3Endpoint})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupRateLimit, ConfigValue: constants.DefaultBackupRateLimit})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyRestoreRateLimit, ConfigValue: constants.DefaultRestoreRateLimit})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyBackupConcurrency, ConfigValue: constants.DefaultBackupConcurrency})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyRestoreConcurrency, ConfigValue: constants.DefaultRestoreConcurrency})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyExportShareStoragePath, ConfigValue: constants.DefaultExportPath})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyImportShareStoragePath, ConfigValue: constants.DefaultImportPath})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyDumplingThreadNum, ConfigValue: constants.DefaultDumplingThreadNum})
	defaultDb.configReaderWriter.CreateConfig(context.TODO(), &config.SystemConfig{ConfigKey: constants.ConfigKeyRetainedPortRange, ConfigValue: constants.DefaultRetainedPortRange})
}

func (p *database) initSystemData() {
	tenant, err := defaultDb.accountReaderWriter.CreateTenant(context.TODO(),
		&account.Tenant{
			ID:               "admin",
			Name:             "EM system administration",
			Creator:          "System",
			Status:           string(constants.TenantStatusNormal),
			OnBoardingStatus: string(constants.TenantOnBoarding)})
	if err != nil {
		framework.LogWithContext(context.TODO()).Errorf("create 'admin' tenant error: %v", err)
		return
	}

	// todo determine if default data needed
	// system admin account
	user := &account.User{
		DefaultTenantID: tenant.ID,
		Name:            "admin",
		Creator:         "System",
	}
	user.GenSaltAndHash("admin")
	_, _, _, err = defaultDb.accountReaderWriter.CreateUser(context.TODO(), user, "admin")
	if err != nil {
		framework.LogWithContext(context.TODO()).Errorf("create 'admin' user error: %v", err)
		return
	}

	// label
	for _, v := range structs.DefaultLabelTypes {
		labelRecord := new(resourcePool.Label)
		labelRecord.ConstructLabelRecord(&v)
		if err = defaultDb.base.Create(labelRecord).Error; err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("create label error: %s", err.Error())
			return
		}
	}

	// batch import parameters & default parameter group sql
	parameterSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/parameters.sql"
	err = syscall.Access(parameterSqlFile, syscall.F_OK)
	if !os.IsNotExist(err) {
		sqls, err := ioutil.ReadFile(parameterSqlFile)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("batch import parameters failed, err = %s", err.Error())
			return
		}
		sqlArr := strings.Split(string(sqls), ";")
		for _, sql := range sqlArr {
			if strings.TrimSpace(sql) == "" {
				continue
			}
			// exec import sql
			defaultDb.base.Exec(sql)
		}
	}

	// import TiUP configs
	tiUPSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/tiup_configs.sql"
	err = syscall.Access(tiUPSqlFile, syscall.F_OK)
	if !os.IsNotExist(err) {
		sqls, err := ioutil.ReadFile(tiUPSqlFile)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("import tiupconfigs failed, err = %s", err.Error())
			return
		}
		sqlArr := strings.Split(string(sqls), ";")
		for _, sql := range sqlArr {
			if strings.TrimSpace(sql) == "" {
				continue
			}
			// exec import sql
			defaultDb.base.Exec(sql)
		}
	}

	// import upgrade paths
	upgradeSqlFile := framework.Current.GetClientArgs().DeployDir + "/sqls/upgrades.sql"
	err = syscall.Access(tiUPSqlFile, syscall.F_OK)
	if !os.IsNotExist(err) {
		sqls, err := ioutil.ReadFile(upgradeSqlFile)
		if err != nil {
			framework.LogForkFile(constants.LogFileSystem).Errorf("import upgrades failed, err = %s", err.Error())
			return
		}
		sqlArr := strings.Split(string(sqls), ";")
		for _, sql := range sqlArr {
			if strings.TrimSpace(sql) == "" {
				continue
			}
			// exec import sql
			defaultDb.base.Exec(sql)
		}
	}
}

func GetChangeFeedReaderWriter() changefeed.ReaderWriter {
	return defaultDb.changeFeedReaderWriter
}

func SetChangeFeedReaderWriter(rw changefeed.ReaderWriter) {
	defaultDb.changeFeedReaderWriter = rw
}

func GetWorkFlowReaderWriter() workflow.ReaderWriter {
	return defaultDb.workFlowReaderWriter
}

func SetWorkFlowReaderWriter(rw workflow.ReaderWriter) {
	defaultDb.workFlowReaderWriter = rw
}

func GetImportExportReaderWriter() importexport.ReaderWriter {
	return defaultDb.importExportReaderWriter
}

func SetImportExportReaderWriter(rw importexport.ReaderWriter) {
	defaultDb.importExportReaderWriter = rw
}

func GetBRReaderWriter() backuprestore.ReaderWriter {
	return defaultDb.brReaderWriter
}

func GetUpgradeReaderWriter() upgrade.ReaderWriter {
	return defaultDb.upgradeReadWriter
}

func SetUpgradeReaderWriter(rw upgrade.ReaderWriter) {
	defaultDb.upgradeReadWriter = rw
}

func SetResourceReaderWriter(rw resource.ReaderWriter) {
	defaultDb.resourceReaderWriter = rw
}

func GetResourceReaderWriter() resource.ReaderWriter {
	return defaultDb.resourceReaderWriter
}

func SetBRReaderWriter(rw backuprestore.ReaderWriter) {
	defaultDb.brReaderWriter = rw
}

func GetClusterReaderWriter() management.ReaderWriter {
	return defaultDb.clusterReaderWriter
}

func SetClusterReaderWriter(rw management.ReaderWriter) {
	defaultDb.clusterReaderWriter = rw
}

func GetConfigReaderWriter() config.ReaderWriter {
	return defaultDb.configReaderWriter
}

func SetConfigReaderWriter(rw config.ReaderWriter) {
	defaultDb.configReaderWriter = rw
}

func GetSecondPartyOperationReaderWriter() secondparty.ReaderWriter {
	return defaultDb.secondPartyOperationReaderWriter
}

func SetSecondPartyOperationReaderWriter(rw secondparty.ReaderWriter) {
	defaultDb.secondPartyOperationReaderWriter = rw
}

func GetParameterGroupReaderWriter() parametergroup.ReaderWriter {
	return defaultDb.parameterGroupReaderWriter
}

func SetParameterGroupReaderWriter(rw parametergroup.ReaderWriter) {
	defaultDb.parameterGroupReaderWriter = rw
}

func GetClusterParameterReaderWriter() parameter.ReaderWriter {
	return defaultDb.clusterParameterReaderWriter
}

func SetClusterParameterReaderWriter(rw parameter.ReaderWriter) {
	defaultDb.clusterParameterReaderWriter = rw
}

func GetRBACReaderWriter() rbac.ReaderWriter {
	return defaultDb.rbacReaderWriter
}

func SetRBACReaderWriter(rw rbac.ReaderWriter) {
	defaultDb.rbacReaderWriter = rw
}

func GetAccountReaderWriter() account.ReaderWriter {
	return defaultDb.accountReaderWriter
}

func SetAccountReaderWriter(rw account.ReaderWriter) {
	defaultDb.accountReaderWriter = rw
}

func GetTokenReaderWriter() identification.ReaderWriter {
	return defaultDb.tokenReaderWriter
}

func SetTokenReaderWriter(rw identification.ReaderWriter) {
	defaultDb.tokenReaderWriter = rw
}

func GetProductReaderWriter() product.ProductReadWriterInterface {
	return defaultDb.productReaderWriter
}

func SetProductReaderWriter(rw product.ProductReadWriterInterface) {
	defaultDb.productReaderWriter = rw
}

func GetTiUPConfigReaderWriter() tiup.ReaderWriter {
	return defaultDb.tiUPConfigReaderWriter
}

func SetTiUPConfigReaderWriter(rw tiup.ReaderWriter) {
	defaultDb.tiUPConfigReaderWriter = rw
}

func SetReportReaderWriter(rw check.ReaderWriter) {
	defaultDb.reportReaderWriter = rw
}

func GetReportReaderWriter() check.ReaderWriter {
	return defaultDb.reportReaderWriter
}

func GetSystemReaderWriter() system.ReaderWriter {
	return defaultDb.systemReaderWriter
}

func SetSystemReaderWriter(rw system.ReaderWriter) {
	defaultDb.systemReaderWriter = rw
}

func MockDB() {
	defaultDb = &database{}
}
