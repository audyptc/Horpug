package database

import (
	"fmt"

	"apihorpug/config"
	menudomain "apihorpug/internal/features/menu/domain"
	permissiondomain "apihorpug/internal/features/permission/domain"
	roledomain "apihorpug/internal/features/role/domain"
	userdomain "apihorpug/internal/features/user/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Bangkok",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&menudomain.Menu{},
		&roledomain.Role{},
		&permissiondomain.Permission{},
		&userdomain.User{},
		&roledomain.RoleMenuPermission{},
	)
}

func SeedPermissions(db *gorm.DB) error {
	permissions := []permissiondomain.Permission{
		{Name: "users.read", Description: "Read users"},
		{Name: "users.write", Description: "Create or update users"},
		{Name: "users.delete", Description: "Delete users"},
		{Name: "roles.read", Description: "Read roles"},
		{Name: "roles.write", Description: "Create or update roles"},
		{Name: "roles.delete", Description: "Delete roles"},
	}

	for _, permission := range permissions {
		if err := db.Where("name = ?", permission.Name).FirstOrCreate(&permission).Error; err != nil {
			return err
		}
	}

	return nil
}

func SeedMenus(db *gorm.DB) error {
	menus := []menudomain.Menu{
		{Name: "Users", Path: "/users", Description: "User management", IsActive: true},
		{Name: "Roles", Path: "/roles", Description: "Role and access management", IsActive: true},
		{Name: "Permissions", Path: "/permissions", Description: "Permission catalog", IsActive: true},
	}

	for _, menu := range menus {
		seed := menu
		if err := db.Where("path = ?", menu.Path).
			Assign(menudomain.Menu{Name: menu.Name, Description: menu.Description, IsActive: menu.IsActive}).
			FirstOrCreate(&seed).Error; err != nil {
			return err
		}
	}

	return nil
}
