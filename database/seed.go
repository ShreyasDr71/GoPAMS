package database

import (
	"log"

	"github.com/ShreyasDr71/GoPAMS/config"
	"github.com/ShreyasDr71/GoPAMS/models"
	"golang.org/x/crypto/bcrypt"
)

// SeedDatabase auto-migrates database schemas and inserts default records if not present
func SeedDatabase() {
	log.Println("Running database migrations...")
	err := DB.AutoMigrate(&models.Group{}, &models.Role{}, &models.User{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migration completed.")

	// 1. Seed Roles
	defaultRoles := []models.Role{
		{Name: "Administrator", Description: "Superuser with full system access", HierarchyLevel: 100},
		{Name: "Manager", Description: "Manages users, groups, and vaults within their team", HierarchyLevel: 80},
		{Name: "Leader", Description: "Can approve tickets and manage team passwords", HierarchyLevel: 70},
		{Name: "Engineer", Description: "Can create and read assigned passwords", HierarchyLevel: 50},
		{Name: "Employee", Description: "Can request passwords and view personal items", HierarchyLevel: 30},
		{Name: "Auditor", Description: "Read-only access to audit logs and security posture", HierarchyLevel: 20},
		{Name: "Guest", Description: "Minimal read-only access to specific assigned resources", HierarchyLevel: 10},
	}

	for _, role := range defaultRoles {
		var count int64
		DB.Model(&models.Role{}).Where("name = ?", role.Name).Count(&count)
		if count == 0 {
			if err := DB.Create(&role).Error; err != nil {
				log.Printf("Failed to seed role %s: %v", role.Name, err)
			} else {
				log.Printf("Seeded role: %s", role.Name)
			}
		}
	}

	// 2. Seed Groups
	defaultGroups := []models.Group{
		{Name: "Infrastructure", Description: "Systems, servers, virtualization, cloud, and core backend infrastructure"},
		{Name: "Networking", Description: "Switches, routers, firewalls, and network configurations"},
		{Name: "Security", Description: "Security operations, risk compliance, IAM, and incident response"},
		{Name: "Development", Description: "Software engineering, QA, DevOps, and applications"},
		{Name: "Finance", Description: "Financial platforms, banking integrations, and accounting systems"},
		{Name: "Family", Description: "Personal vaults and family shared resources"},
	}

	for _, group := range defaultGroups {
		var count int64
		DB.Model(&models.Group{}).Where("name = ?", group.Name).Count(&count)
		if count == 0 {
			if err := DB.Create(&group).Error; err != nil {
				log.Printf("Failed to seed group %s: %v", group.Name, err)
			} else {
				log.Printf("Seeded group: %s", group.Name)
			}
		}
	}

	// 3. Seed Default Admin User
	adminUser := config.AppConfig.DefaultAdminUser
	adminPass := config.AppConfig.DefaultAdminPassword

	var count int64
	DB.Model(&models.User{}).Where("username = ?", adminUser).Count(&count)
	if count == 0 {
		// Find Administrator role
		var adminRole models.Role
		if err := DB.Where("name = ?", "Administrator").First(&adminRole).Error; err != nil {
			log.Fatalf("Could not find Administrator role for seeding: %v", err)
		}

		// Find Security group to assign default admin
		var securityGroup models.Group
		if err := DB.Where("name = ?", "Security").First(&securityGroup).Error; err != nil {
			log.Printf("Could not find Security group for seeding default admin: %v", err)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash default admin password: %v", err)
		}

		admin := models.User{
			FullName:           "Default Administrator",
			Username:           adminUser,
			PasswordHash:       string(hashedPassword),
			PhoneNumber:        "+1000000000",
			RoleID:             &adminRole.ID,
			GroupID:            &securityGroup.ID,
			MustChangePassword: true,
			IsAdmin:            true,
		}

		if err := DB.Create(&admin).Error; err != nil {
			log.Fatalf("Failed to seed default administrator: %v", err)
		}
		log.Printf("Seeded default administrator: %s with temporary password.", adminUser)
	} else {
		log.Println("Administrator account already exists. Skipping seeding.")
	}
}
