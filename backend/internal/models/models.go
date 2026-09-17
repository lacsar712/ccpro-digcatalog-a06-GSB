package models

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"uniqueIndex;size:64;not null"`
	PasswordHash string         `json:"-" gorm:"size:255;not null"`
	Role         string         `json:"role" gorm:"size:32;not null"` // admin | recorder
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Site struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:128;not null"`
	Period    string         `json:"period" gorm:"size:64;not null"` // 新石器/商周等
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Manager   string         `json:"manager" gorm:"size:64"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Units     []Unit         `json:"units,omitempty" gorm:"foreignKey:SiteID"`
}

type Unit struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	SiteID           uint           `json:"siteId" gorm:"not null;index"`
	Code             string         `json:"code" gorm:"size:64;not null"` // T1, T2...
	DepthMin         float64        `json:"depthMin"`
	DepthMax         float64        `json:"depthMax"`
	StratumDesc      string         `json:"stratumDesc" gorm:"type:text"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
	Site             *Site          `json:"site,omitempty" gorm:"foreignKey:SiteID"`
	Finds            []Find         `json:"finds,omitempty" gorm:"foreignKey:UnitID"`
}

type Material struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;size:64;not null"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type Find struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UnitID       uint           `json:"unitId" gorm:"not null;index"`
	MaterialID   *uint          `json:"materialId" gorm:"index"`
	RegisterNo   string         `json:"registerNo" gorm:"uniqueIndex;size:64;not null"`
	ArtifactType   string         `json:"artifactType" gorm:"size:64;not null"` // 陶片/青铜器/骨器
	MaterialName string         `json:"materialName" gorm:"size:64"`          // 冗余展示字段
	Completeness string         `json:"completeness" gorm:"size:32"`          // 完整/残缺/碎片
	FindDate     *time.Time     `json:"findDate" gorm:"type:date"`
	Description  string         `json:"description" gorm:"type:text"`
	StorageLoc   string         `json:"storageLoc" gorm:"size:128"`
	// 最新一份度量单的冗余展示字段（由系统回写，不参与编辑，不影响 Description）
	LatestMeasuredAt  *time.Time       `json:"latestMeasuredAt"`
	LatestDimsSummary string           `json:"latestDimsSummary" gorm:"size:160"`
	LatestWeightG     *float64         `json:"latestWeightG"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt   `json:"-" gorm:"index"`
	Unit              *Unit            `json:"unit,omitempty" gorm:"foreignKey:UnitID"`
	Material          *Material        `json:"material,omitempty" gorm:"foreignKey:MaterialID"`
	Measurements      []MeasurementSheet `json:"measurements,omitempty" gorm:"foreignKey:FindID"`
}

// ApplyLatestMeasurement 将最新一份度量摘要回写到 Find 的冗余展示字段；
// 传入 nil 表示清空。不会改动 Description 等业务字段。
func (f *Find) ApplyLatestMeasurement(m *MeasurementSheet) {
	if m == nil {
		f.LatestMeasuredAt = nil
		f.LatestDimsSummary = ""
		f.LatestWeightG = nil
		return
	}
	f.LatestMeasuredAt = &m.MeasuredAt
	f.LatestDimsSummary = m.DimsSummary()
	f.LatestWeightG = m.WeightG
}

// MeasurementSheet 器物度量单：一次度量挂一个 Find，同一 Find 可有多份历史记录
type MeasurementSheet struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	FindID       uint           `json:"findId" gorm:"not null;index"`
	MeasuredAt   time.Time      `json:"measuredAt" gorm:"not null;index"`
	LengthMm     float64        `json:"lengthMm"`                // 非负
	WidthMm      float64        `json:"widthMm"`                 // 非负
	HeightMm     float64        `json:"heightMm"`                // 非负
	WeightG      *float64       `json:"weightG"`                 // 可空，非负
	CaliperNote  string         `json:"caliperNote" gorm:"type:text"`
	OperatorName string         `json:"operatorName" gorm:"size:64;not null"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Find         *Find          `json:"find,omitempty" gorm:"foreignKey:FindID"`
}

// DimsSummary 生成 "长 x × 宽 x × 高 x mm" 的展示摘要
func (m *MeasurementSheet) DimsSummary() string {
	return fmt.Sprintf("长 %s × 宽 %s × 高 %s mm",
		trimNum(m.LengthMm), trimNum(m.WidthMm), trimNum(m.HeightMm))
}

func trimNum(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
