package dao

import "gorm.io/gorm"

//func InitTables(db *gorm.DB) error {
//	return db.AutoMigrate(&User{})
//}

func InitTables(db *gorm.DB) error {

	//if err := dropAllTables(db); err != nil {
	//	return err
	//}
	return db.AutoMigrate(
		&User{},
		&Student{},
		&Parent{},
		&StandHolder{},
		&Organizer{},
		&Stand{},
		&Kermesse{},
		&Stock{},
		&ChatMessage{},
	)
}

func dropAllTables(db *gorm.DB) error {
	// Disable foreign key checks
	db.Exec("SET CONSTRAINTS ALL DEFERRED;")

	// Get all table names
	var tableNames []string
	if err := db.Table("information_schema.tables").
		Where("table_schema = ?", "public"). // for PostgreSQL, use 'public' schema
		Pluck("table_name", &tableNames).Error; err != nil {
		return err
	}

	// Drop all tables
	for _, tableName := range tableNames {
		if err := db.Exec("DROP TABLE IF EXISTS " + tableName + " CASCADE").Error; err != nil {
			return err
		}
	}

	// Re-enable foreign key checks
	db.Exec("SET CONSTRAINTS ALL IMMEDIATE;")

	return nil
}
