package handlers

import (
	"net/http"
	"strconv"
	"time"

	"digcatalog/internal/models"

	"github.com/gin-gonic/gin"
)

// ---------- MeasurementSheets（器物度量单） ----------

type measurementReq struct {
	MeasuredAt   string   `json:"measuredAt"`
	LengthMm     *float64 `json:"lengthMm"`
	WidthMm      *float64 `json:"widthMm"`
	HeightMm     *float64 `json:"heightMm"`
	WeightG      *float64 `json:"weightG"`
	CaliperNote  string   `json:"caliperNote"`
	OperatorName string   `json:"operatorName"`
}

func parseMeasuredAt(s string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func validateMeasurementReq(req *measurementReq) (time.Time, string) {
	measuredAt, ok := parseMeasuredAt(req.MeasuredAt)
	if !ok {
		return measuredAt, "度量时间必填且格式无效"
	}
	if req.LengthMm == nil || req.WidthMm == nil || req.HeightMm == nil {
		return measuredAt, "长、宽、高必填"
	}
	if *req.LengthMm < 0 || *req.WidthMm < 0 || *req.HeightMm < 0 {
		return measuredAt, "长、宽、高不能为负数"
	}
	if req.WeightG != nil && *req.WeightG < 0 {
		return measuredAt, "重量不能为负数"
	}
	if req.OperatorName == "" {
		return measuredAt, "操作人必填"
	}
	return measuredAt, ""
}

// refreshFindMeasurementSummary 将最新一份度量摘要回写到 Find 的冗余展示字段
func (h *Handler) refreshFindMeasurementSummary(findID uint) {
	var find models.Find
	if err := h.DB.First(&find, findID).Error; err != nil {
		return
	}
	var latest models.MeasurementSheet
	if err := h.DB.Where("find_id = ?", findID).
		Order("measured_at desc, id desc").
		First(&latest).Error; err != nil {
		find.ApplyLatestMeasurement(nil)
	} else {
		find.ApplyLatestMeasurement(&latest)
	}
	h.DB.Model(&models.Find{}).Where("id = ?", findID).Updates(map[string]interface{}{
		"latest_measured_at":  find.LatestMeasuredAt,
		"latest_dims_summary": find.LatestDimsSummary,
		"latest_weight_g":     find.LatestWeightG,
	})
}

// ListFindMeasurements GET /finds/:id/measurements —— 同一 Find 的历史度量，按时间倒序
func (h *Handler) ListFindMeasurements(c *gin.Context) {
	findID, _ := strconv.Atoi(c.Param("id"))
	var find models.Find
	if err := h.DB.First(&find, findID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文物不存在"})
		return
	}
	var list []models.MeasurementSheet
	if err := h.DB.Where("find_id = ?", find.ID).
		Order("measured_at desc, id desc").
		Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// CreateMeasurement POST /finds/:id/measurements
func (h *Handler) CreateMeasurement(c *gin.Context) {
	findID, _ := strconv.Atoi(c.Param("id"))
	var find models.Find
	if err := h.DB.First(&find, findID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文物不存在"})
		return
	}
	var req measurementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	measuredAt, msg := validateMeasurementReq(&req)
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	m := models.MeasurementSheet{
		FindID:       find.ID,
		MeasuredAt:   measuredAt,
		LengthMm:     *req.LengthMm,
		WidthMm:      *req.WidthMm,
		HeightMm:     *req.HeightMm,
		WeightG:      req.WeightG,
		CaliperNote:  req.CaliperNote,
		OperatorName: req.OperatorName,
	}
	if err := h.DB.Create(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.refreshFindMeasurementSummary(find.ID)
	c.JSON(http.StatusCreated, m)
}

// UpdateMeasurement PUT /measurements/:id（FindID 不可改）
func (h *Handler) UpdateMeasurement(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m models.MeasurementSheet
	if err := h.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "度量单不存在"})
		return
	}
	var req measurementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	measuredAt, msg := validateMeasurementReq(&req)
	if msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	m.MeasuredAt = measuredAt
	m.LengthMm = *req.LengthMm
	m.WidthMm = *req.WidthMm
	m.HeightMm = *req.HeightMm
	m.WeightG = req.WeightG
	m.CaliperNote = req.CaliperNote
	m.OperatorName = req.OperatorName
	if err := h.DB.Save(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.refreshFindMeasurementSummary(m.FindID)
	c.JSON(http.StatusOK, m)
}

// DeleteMeasurement DELETE /measurements/:id
func (h *Handler) DeleteMeasurement(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m models.MeasurementSheet
	if err := h.DB.First(&m, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "度量单不存在"})
		return
	}
	if err := h.DB.Delete(&m).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.refreshFindMeasurementSummary(m.FindID)
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
