package domain

import (
	"archive/zip"
	"fmt"
	"github.com/BurntSushi/toml"
	proto "github.com/pingcap/ticp/micro-cluster/proto"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"time"
)

type TransportType string
const (
	TransportTypeExport TransportType = "export"
	TransportTypeImport TransportType = "import"
)

type TransportRecord struct {
	Id 				uint
	ClusterId 		string
	StartTime 		time.Time
	EndTime 		time.Time
	Operator 		Operator
	TransportType   TransportType
	FilePath 		string
}

type ImportInfo struct {
	ClusterId 		string
	UserName		string
	Password 		string
	FilePath 		string
	StartTime 		time.Time
}

type ExportInfo struct {
	ClusterId 		string
	UserName		string
	Password 		string
	FileType 		string
	StartTime 		time.Time
}

/*
	data import toml config for lighting
	https://docs.pingcap.com/zh/tidb/dev/deploy-tidb-lightning
 */
type DataImportConfig struct {
	Lighting 		LightingCfg			`toml:"lighting"`
	TikvImporter	TikvImporterCfg 	`toml:"tikv-importer"`
	MyDumper		MyDumperCfg 		`toml:"mydumper"`
	Tidb 			TidbCfg 			`toml:"tidb"`
}

type LightingCfg struct {
	Level			string	`toml:"level"`	//lighting log level
	File 			string 	`toml:"file"`	//lighting log path
}

type TikvImporterCfg struct {
	Backend 		string 	`toml:"backend"`		//backend mode: local/normal
	SortedKvDir		string 	`toml:"sorted-kv-dir"`	//temp store path
}

type MyDumperCfg struct {
	DataSourceDir	string	`toml:"data-source-dir"`	//import data filepath
}

type TidbCfg struct {
	Host 			string 	`toml:"host"`
	Port 			int		`toml:"port"`
	User 			string 	`toml:"user"`
	Password		string 	`toml:"password"`
	StatusPort		int 	`toml:"status-port"`	//table infomation from tidb status port
	PdAddr 			string 	`toml:"pd-addr"`
}

var contextDataTransportKey = "dataTransportInfo"
var dataTransportDirPrefix = "/tmp/tiem/datatransport"

func ExportData(ope *proto.OperatorDTO, clusterId string, userName string, password string, fileType string) (uint32, error) {
	log.Infof("[domain] begin exportdata clusterId: %s, userName: %s, password: %s, fileType: %s", clusterId, userName, password, fileType)
	defer log.Infof("[domain] end exportdata")
	//todo: check operator
	operator := parseOperatorFromDTO(ope)
	log.Info(operator)
	clusterAggregation, err := ClusterRepo.Load(clusterId)

	info := &ExportInfo{
		ClusterId: clusterId,
		UserName: userName,
		Password: password,//todo: need encrypt
		FileType: fileType,
		StartTime: time.Now(),
	}

	// Start the workflow
	flow, err := CreateFlowWork(clusterId, FlowExportData)
	if err != nil {
		return 0, nil
	}
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextDataTransportKey, info)
	flow.Start()

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	ClusterRepo.Persist(clusterAggregation)
	return uint32(flow.FlowWork.Id), nil
}

func ImportData(ope *proto.OperatorDTO, clusterId string, userName string, password string, filepath string) (uint32, error) {
	log.Infof("[domain] begin importdata clusterId: %s, userName: %s, password: %s, datadIR: %s", clusterId, userName, password, filepath)
	defer log.Infof("[domain] end importdata")
	//todo: check operator
	operator := parseOperatorFromDTO(ope)
	log.Info(operator)
	clusterAggregation, err := ClusterRepo.Load(clusterId)

	info := &ImportInfo{
		ClusterId: clusterId,
		UserName: userName,
		Password: password,//todo: need encrypt
		FilePath: filepath,
		StartTime: time.Now(),
	}

	// Start the workflow
	flow, err := CreateFlowWork(clusterId, FlowImportData)
	if err != nil {
		return 0, nil
	}
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextDataTransportKey, info)
	flow.Start()

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	ClusterRepo.Persist(clusterAggregation)
	return uint32(flow.FlowWork.Id), nil
}

func convertTomlConfig(clusterAggregation *ClusterAggregation, info *ImportInfo) *DataImportConfig {
	if clusterAggregation == nil || clusterAggregation.CurrentTiUPConfigRecord == nil {
		return nil
	}
	cluster := clusterAggregation.Cluster
	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	if configModel == nil || configModel.TiDBServers == nil || configModel.PDServers == nil {
		return nil
	}
	tidbServer := configModel.TiDBServers[0]
	pdServer := configModel.PDServers[0]

	config := &DataImportConfig{
		Lighting: LightingCfg{
			Level: "info",
			File: fmt.Sprintf("%s/tidb-lighting.log", getDataTransportDir(cluster.Id, TransportTypeImport)),
		},
		TikvImporter: TikvImporterCfg{
			Backend: "local",
			SortedKvDir: "/mnt/ssd/sorted-kv-dir", //todo: replace config item
		},
		MyDumper: MyDumperCfg{
			DataSourceDir: info.FilePath,
		},
		Tidb: TidbCfg{
			Host: tidbServer.Host,
			Port: tidbServer.Port,
			User: info.UserName,
			Password: info.Password,
			StatusPort: tidbServer.StatusPort,
			PdAddr: fmt.Sprintf("%s:%d", pdServer.Host, pdServer.ClientPort),
		},
	}
	return config
}

/**
data import && export dir
└── dataTransportDirPrefix/[cluster-id]
	├── import
	|	├── data
	|	├── data.zip
	|	└── log
	└── export
		├── data
		├── data.zip
		└── log
*/
func getDataTransportDir(clusterId string, transportType TransportType) string {
	return fmt.Sprintf("%s/%s/%s", dataTransportDirPrefix, clusterId, transportType)
}

func cleanDataTransportDir(clusterId string, transportType TransportType) error {
	if err := os.RemoveAll(getDataTransportDir(clusterId, transportType)); err != nil {
		return err
	}

	if err := os.Mkdir(getDataTransportDir(clusterId, transportType), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func zipDir(dir string, zipFile string) error {
	fz, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("Create zip file failed: %s", err.Error())
	}
	defer fz.Close()

	w := zip.NewWriter(fz)
	defer w.Close()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fDest, err := w.Create(path)
			if err != nil {
				return fmt.Errorf("zip Create failed: %s", err.Error())
			}
			fSrc, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("zip Open failed: %s", err.Error())
			}
			defer fSrc.Close()
			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				return fmt.Errorf("zip Copy failed: %s", err.Error())
			}
		}
		return nil
	})

	return nil
}

func unzipDir(zipFile string, dir string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("Open zip file failed: %s", err.Error())
	}
	defer r.Close()

	for _, f := range r.File {
		func() {
			path := dir + string(filepath.Separator) + f.Name
			os.MkdirAll(filepath.Dir(path), 0755)
			fDest, err := os.Create(path)
			if err != nil {
				log.Errorf("unzip Create failed: %s", err.Error())
				return
			}
			defer fDest.Close()

			fSrc, err := f.Open()
			if err != nil {
				log.Errorf("unzip Open failed: %s", err.Error())
				return
			}
			defer fSrc.Close()

			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				log.Errorf("unzip Copy failed: %s", err.Error())
				return
			}
		}()
	}
	return nil
}

func buildDataImportConfig(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	info := context.value(contextDataTransportKey).(ImportInfo)
	cluster := clusterAggregation.Cluster

	if err := cleanDataTransportDir(cluster.Id, TransportTypeImport); err != nil {
		log.Errorf("[domain] clean import directory failed, %s", err.Error())
		return false
	}

	config := convertTomlConfig(&clusterAggregation, &info)
	if config == nil {
		log.Errorf("[domain] convert toml config failed, cluster: %v", clusterAggregation)
		return false
	}
	filePath := fmt.Sprintf("%s/tidb-lighting.toml", getDataTransportDir(cluster.Id, TransportTypeImport))
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		log.Errorf("[domain] decode data import toml config failed, %s", err.Error())
		return false
	}
	return true
}

func importDataToCluster(task *TaskEntity, context *FlowContext) bool {
	//todo: call tiupmgr
	return true
}

func updateDataImportRecord(task *TaskEntity, context *FlowContext) bool {
	//todo: implement
	return true
}

func exportDataFromCluster(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	cluster := clusterAggregation.Cluster

	if err := cleanDataTransportDir(cluster.Id, TransportTypeExport); err != nil {
		log.Errorf("[domain] clean export directory failed, %s", err.Error())
		return false
	}
	//todo: call tiupmgr

	return true
}

func updateDataExportRecord(task *TaskEntity, context *FlowContext) bool {
	//todo: implement
	return true
}

func compressExportData(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	cluster := clusterAggregation.Cluster

	dataDir := fmt.Sprintf("%s/data", getDataTransportDir(cluster.Id, TransportTypeExport))
	dataZipDir := fmt.Sprintf("%s/data.zip", getDataTransportDir(cluster.Id, TransportTypeExport))
	if err := zipDir(dataDir, dataZipDir); err != nil {
		log.Errorf("[domain] compress export data failed, %s", err.Error())
		return false
	}

	return true
}

func deCompressImportData(task *TaskEntity, context *FlowContext) bool {
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	info := context.value(contextDataTransportKey).(ImportInfo)
	cluster := clusterAggregation.Cluster

	dataDir := fmt.Sprintf("%s/", getDataTransportDir(cluster.Id, TransportTypeImport))
	dataZipDir := info.FilePath
	if err := unzipDir(dataZipDir, dataDir); err != nil {
		log.Errorf("[domain] deCompress import data failed, %s", err.Error())
		return false
	}

	return true
}