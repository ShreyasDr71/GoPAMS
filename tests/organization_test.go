package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ShreyasDr71/GoPAMS/config"
	"github.com/ShreyasDr71/GoPAMS/database"
	"github.com/ShreyasDr71/GoPAMS/models"
	"github.com/ShreyasDr71/GoPAMS/routes"
	"github.com/ShreyasDr71/GoPAMS/services"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	var err error
	database.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Migrate schemas
	err = database.DB.AutoMigrate(&models.Group{}, &models.Role{}, &models.User{})
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Seed roles and groups (needed for tests)
	adminRole := models.Role{Name: "Administrator", Description: "Admin role", HierarchyLevel: 100}
	engineerRole := models.Role{Name: "Engineer", Description: "Engineer role", HierarchyLevel: 50}
	database.DB.Create(&adminRole)
	database.DB.Create(&engineerRole)

	devGroup := models.Group{Name: "Development", Description: "Dev team"}
	database.DB.Create(&devGroup)

	// Create an admin user and a regular user
	adminUser := models.User{
		FullName:           "Admin User",
		Username:           "admin_test",
		PasswordHash:       "dummy_hash",
		IsAdmin:            true,
		RoleID:             &adminRole.ID,
		GroupID:            &devGroup.ID,
		MustChangePassword: false,
	}
	regularUser := models.User{
		FullName:           "Regular User",
		Username:           "user_test",
		PasswordHash:       "dummy_hash",
		IsAdmin:            false,
		RoleID:             &engineerRole.ID,
		GroupID:            &devGroup.ID,
		MustChangePassword: false,
	}
	database.DB.Create(&adminUser)
	database.DB.Model(&adminUser).Update("must_change_password", false)
	
	database.DB.Create(&regularUser)
	database.DB.Model(&regularUser).Update("must_change_password", false)

	// Set test configs
	config.AppConfig = &config.Config{
		JWTSecret:             "test_secret_key_12345",
		SessionTimeoutMinutes: 10,
		EnterpriseMode:        false,
	}
}

func performRequest(r http.Handler, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getTokens(t *testing.T) (string, string) {
	adminToken, err := services.GenerateJWT(1, "admin_test", true, false, "Administrator")
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}
	userToken, err := services.GenerateJWT(2, "user_test", false, false, "Engineer")
	if err != nil {
		t.Fatalf("failed to generate user token: %v", err)
	}
	return adminToken, userToken
}

func TestGroupCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)
	router := routes.SetupRouter()
	adminToken, userToken := getTokens(t)

	// 1. List groups (AuthRequired)
	w := performRequest(router, "GET", "/api/groups", nil, userToken)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	// 2. Create group as regular user (should be Forbidden)
	newGroup := map[string]string{
		"name":        "Finance",
		"description": "Finance team",
	}
	w = performRequest(router, "POST", "/api/groups", newGroup, userToken)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-admin group creation, got %d", w.Code)
	}

	// 3. Create group as admin (should succeed)
	w = performRequest(router, "POST", "/api/groups", newGroup, adminToken)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", w.Code)
	}

	var group models.Group
	_ = json.Unmarshal(w.Body.Bytes(), &group)
	if group.Name != "Finance" {
		t.Errorf("expected group name to be Finance, got %s", group.Name)
	}

	// 4. Update group as admin
	updatedGroup := map[string]string{
		"name":        "Finance & Accounting",
		"description": "Finance department",
	}
	w = performRequest(router, "PUT", "/api/groups/"+strconv.Itoa(int(group.ID)), updatedGroup, adminToken)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK on update, got %d", w.Code)
	}

	// 5. Delete group as admin
	w = performRequest(router, "DELETE", "/api/groups/"+strconv.Itoa(int(group.ID)), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK on delete, got %d", w.Code)
	}
}

func TestRoleCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)
	router := routes.SetupRouter()
	adminToken, userToken := getTokens(t)

	// 1. Create role as admin
	newRole := map[string]interface{}{
		"name":            "Auditor",
		"description":     "Audit access",
		"hierarchy_level": 20,
	}
	w := performRequest(router, "POST", "/api/roles", newRole, adminToken)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", w.Code)
	}

	var role models.Role
	_ = json.Unmarshal(w.Body.Bytes(), &role)

	// 2. Prevent deleting Administrator role
	var adminRole models.Role
	database.DB.Where("name = ?", "Administrator").First(&adminRole)
	w = performRequest(router, "DELETE", "/api/roles/"+strconv.Itoa(int(adminRole.ID)), nil, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when deleting Admin role, got %d", w.Code)
	}

	// 3. Delete normal role as user (should be Forbidden)
	w = performRequest(router, "DELETE", "/api/roles/"+strconv.Itoa(int(role.ID)), nil, userToken)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden, got %d", w.Code)
	}

	// 4. Delete normal role as admin (should succeed)
	w = performRequest(router, "DELETE", "/api/roles/"+strconv.Itoa(int(role.ID)), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}

func TestUserCRUDAndGuardrails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupTestDB(t)
	router := routes.SetupRouter()
	adminToken, userToken := getTokens(t)

	// 1. List users as regular user (should be Forbidden)
	w := performRequest(router, "GET", "/api/users", nil, userToken)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-admin users list, got %d", w.Code)
	}

	// 2. Create user as admin (normal mode)
	newUser := map[string]interface{}{
		"full_name":     "New Engineer",
		"username":      "engineer_new",
		"phone_number":  "+1999999999",
		"email":         "new@example.com",
		"role_id":       2, // Engineer role
		"group_id":      1, // Development group
		"is_admin":      false,
		"temp_password": "TempPassword123!",
	}
	w = performRequest(router, "POST", "/api/users", newUser, adminToken)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created user, got %d %s", w.Code, w.Body.String())
	}

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["temporary_password"] != "TempPassword123!" {
		t.Errorf("expected temporary password to match input, got %v", res["temporary_password"])
	}

	// 3. Enterprise mode checks
	config.AppConfig.EnterpriseMode = true
	// Create user without Employee ID (should fail)
	newUserNoEmp := map[string]interface{}{
		"full_name": "No Emp ID User",
		"username":  "no_emp_id",
	}
	w = performRequest(router, "POST", "/api/users", newUserNoEmp, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request in Enterprise mode without employee_id, got %d", w.Code)
	}

	// Create user WITH Employee ID in Enterprise mode (should succeed)
	newUserWithEmp := map[string]interface{}{
		"full_name":   "Emp ID User",
		"username":    "emp_user",
		"employee_id": "EMP-100",
	}
	w = performRequest(router, "POST", "/api/users", newUserWithEmp, adminToken)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created with employee ID, got %d %s", w.Code, w.Body.String())
	}
	config.AppConfig.EnterpriseMode = false // revert

	// 4. Guardrail: Demoting the last administrator (should fail)
	// Currently we have 1 admin user (ID 1). Let's attempt to update user ID 1 and set is_admin to false
	demoteReq := map[string]interface{}{
		"full_name": "Demoted Admin",
		"is_admin":  false,
	}
	w = performRequest(router, "PUT", "/api/users/1", demoteReq, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when demoting last admin, got %d", w.Code)
	}

	// 5. Guardrail: Deleting oneself (should fail)
	w = performRequest(router, "DELETE", "/api/users/1", nil, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request when deleting oneself, got %d", w.Code)
	}

	// 6. Reset password (should succeed and generate one if empty)
	w = performRequest(router, "POST", "/api/users/2/reset-password", map[string]string{}, adminToken)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK on password reset, got %d", w.Code)
	}
	var resetRes map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resetRes)
	if resetRes["temporary_password"] == "" {
		t.Errorf("expected generated temporary password on reset, got empty")
	}
}
