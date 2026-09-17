package seed

import (
	"log"
	"time"

	"digcatalog/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func hash(pw string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func date(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}

func fp(v float64) *float64 { return &v }

func at(s string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", s)
	return t
}

func Run(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	users := []models.User{
		{Username: "admin", PasswordHash: hash("123456"), Role: "admin"},
		{Username: "recorder", PasswordHash: hash("123456"), Role: "recorder"},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			log.Printf("seed user error: %v", err)
		}
	}

	materials := []models.Material{
		{Name: "陶器", Description: "泥质或夹砂陶质器物及残片"},
		{Name: "青铜", Description: "铜锡合金礼器、兵器、工具等"},
		{Name: "骨质", Description: "动物骨骼加工器具或遗存"},
		{Name: "玉石", Description: "玉器、石器及石质工具"},
		{Name: "铁质", Description: "铁器及锈蚀残件"},
	}
	for i := range materials {
		db.Create(&materials[i])
	}

	sites := []models.Site{
		{Name: "二里头遗址发掘区A", Period: "夏商周", Latitude: 34.7035, Longitude: 112.7892, Manager: "张明远"},
		{Name: "良渚古城外围探区", Period: "新石器", Latitude: 30.3936, Longitude: 120.0398, Manager: "李清荷"},
		{Name: "殷墟王陵区东段", Period: "商周", Latitude: 36.1218, Longitude: 114.3165, Manager: "王启程"},
	}
	for i := range sites {
		db.Create(&sites[i])
	}

	units := []models.Unit{
		{SiteID: sites[0].ID, Code: "T1", DepthMin: 0.2, DepthMax: 1.8, StratumDesc: "第①层表土；第②层灰褐文化层含陶片"},
		{SiteID: sites[0].ID, Code: "T2", DepthMin: 0.3, DepthMax: 2.1, StratumDesc: "第③层黑灰层，见夯土迹象"},
		{SiteID: sites[1].ID, Code: "T1", DepthMin: 0.1, DepthMax: 1.5, StratumDesc: "淤泥层下见玉琮残片伴随石核"},
		{SiteID: sites[2].ID, Code: "T3", DepthMin: 0.5, DepthMax: 2.4, StratumDesc: "灰坑开口于第②层下，填土含卜骨"},
	}
	for i := range units {
		db.Create(&units[i])
	}

	m0, m1, m2, m3 := materials[0].ID, materials[1].ID, materials[2].ID, materials[3].ID
	finds := []models.Find{
		{
			UnitID: units[0].ID, MaterialID: &m0, RegisterNo: "EL-2024-0001",
			ArtifactType: "陶片", MaterialName: "陶器", Completeness: "碎片",
			FindDate: date("2024-03-12"), Description: "泥质灰陶口沿残片，可见弦纹", StorageLoc: "库房A-架01",
		},
		{
			UnitID: units[0].ID, MaterialID: &m1, RegisterNo: "EL-2024-0002",
			ArtifactType: "青铜器", MaterialName: "青铜", Completeness: "残缺",
			FindDate: date("2024-03-15"), Description: "青铜爵足残段，表面绿锈", StorageLoc: "库房B-柜03",
		},
		{
			UnitID: units[1].ID, MaterialID: &m0, RegisterNo: "EL-2024-0003",
			ArtifactType: "陶片", MaterialName: "陶器", Completeness: "完整",
			FindDate: date("2024-04-02"), Description: "小型陶纺轮，穿孔完好", StorageLoc: "库房A-架02",
		},
		{
			UnitID: units[2].ID, MaterialID: &m3, RegisterNo: "LZ-2024-0010",
			ArtifactType: "玉器", MaterialName: "玉石", Completeness: "残缺",
			FindDate: date("2024-05-08"), Description: "玉琮角部残片，刻纹清晰", StorageLoc: "珍品柜-01",
		},
		{
			UnitID: units[3].ID, MaterialID: &m2, RegisterNo: "YX-2024-0021",
			ArtifactType: "骨器", MaterialName: "骨质", Completeness: "完整",
			FindDate: date("2024-06-18"), Description: "骨笄一件，磨光精细", StorageLoc: "库房C-屉05",
		},
		{
			UnitID: units[3].ID, MaterialID: &m1, RegisterNo: "YX-2024-0022",
			ArtifactType: "青铜器", MaterialName: "青铜", Completeness: "碎片",
			FindDate: date("2024-06-20"), Description: "青铜戈援部碎片", StorageLoc: "库房B-柜07",
		},
	}
	for i := range finds {
		db.Create(&finds[i])
	}

	// 度量单：EL-2024-0001 与 LZ-2024-0010 各留多份历史，按时间倒序可追溯
	sheets := []models.MeasurementSheet{
		// 陶片 EL-2024-0001：出土后初次测量 + 修复后复测
		{
			FindID: finds[0].ID, MeasuredAt: at("2024-03-13 09:30:00"),
			LengthMm: 42.5, WidthMm: 28.2, HeightMm: 7.8, WeightG: fp(18.6),
			CaliperNote: "数显卡尺沿口沿最长轴测量；边缘有新碴口", OperatorName: "周仪",
		},
		{
			FindID: finds[0].ID, MeasuredAt: at("2024-04-20 14:10:00"),
			LengthMm: 42.3, WidthMm: 28.0, HeightMm: 7.9, WeightG: fp(18.4),
			CaliperNote: "清理加固后复测，尺寸微缩属正常失水", OperatorName: "周仪",
		},
		// 玉琮残片 LZ-2024-0010：两次测量 + 一次高精度复称
		{
			FindID: finds[3].ID, MeasuredAt: at("2024-05-09 10:05:00"),
			LengthMm: 63.4, WidthMm: 55.1, HeightMm: 22.6, WeightG: fp(96.2),
			CaliperNote: "残角外廓取最大长、宽；厚度在射部中段", OperatorName: "沈衡",
		},
		{
			FindID: finds[3].ID, MeasuredAt: at("2024-06-01 11:20:00"),
			LengthMm: 63.4, WidthMm: 55.0, HeightMm: 22.6, WeightG: fp(96.0),
			CaliperNote: "电子天平复称；刻纹深度未计入高度", OperatorName: "沈衡",
		},
	}
	for i := range sheets {
		db.Create(&sheets[i])
	}

	// 将各 Find 最新一份度量摘要回写展示字段（保持 Description 原样）
	latestByFind := map[uint]models.MeasurementSheet{}
	for _, s := range sheets {
		cur, ok := latestByFind[s.FindID]
		if !ok || s.MeasuredAt.After(cur.MeasuredAt) {
			latestByFind[s.FindID] = s
		}
	}
	for findID, s := range latestByFind {
		db.Model(&models.Find{}).Where("id = ?", findID).Updates(map[string]interface{}{
			"last_measured_at": s.MeasuredAt,
			"last_length_mm":   s.LengthMm,
			"last_width_mm":    s.WidthMm,
			"last_height_mm":   s.HeightMm,
			"last_weight_g":    s.WeightG,
		})
	}

	log.Println("seed data inserted")
}
