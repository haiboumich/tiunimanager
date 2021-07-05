package models

import (
	"gorm.io/gorm"
)

type Tenant struct {
	gorm.Model

	Name   string		`gorm:"size:255"`
	Type   int8			`gorm:"size:255"`
	Status int8			`gorm:"size:255"`
}

func AddTenant(name string, tenantType, status int8) (tenant Tenant, err error) {
	tenant.Status = tenantType
	tenant.Type = tenantType
	tenant.Name = name

	MetaDB.Create(&tenant)
	// 返回ID
	return
}

func FindTenantById(tenantId int) (tenant Tenant, err error){
	MetaDB.First(&tenant, tenantId)
	return
}

func FindTenantByName(name string) (tenant Tenant, err error){
	MetaDB.Where("name = ?", name).First(&tenant)
	return
}