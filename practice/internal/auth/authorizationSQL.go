package authorization

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Permission represents a specific action that can be performed
type Permission string

// Predefined set of permissions
const (
	// User-related permissions
	PermUserRead   Permission = "user:read"
	PermUserWrite  Permission = "user:write"
	PermUserDelete Permission = "user:delete"

	// Project-related permissions
	PermProjectCreate Permission = "project:create"
	PermProjectRead   Permission = "project:read"
	PermProjectUpdate Permission = "project:update"
	PermProjectDelete Permission = "project:delete"

	// Resource-related permissions
	PermResourceView   Permission = "resource:view"
	PermResourceModify Permission = "resource:modify"
)

// Role defines a set of permissions
type Role struct {
	ID          int64
	Name        string
	Permissions []Permission
	CreatedAt   time.Time
}

// User represents a system user with authorization details
type User struct {
	ID         int64
	Username   string
	Email      string
	RoleID     int64
	RoleName   string
	CreatedAt  time.Time
}

// AuthorizationService manages role-based access control with PostgreSQL
type AuthorizationService struct {
	db *sql.DB
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(db *sql.DB) *AuthorizationService {
	return &AuthorizationService{db: db}
}

// CreateRole establishes a new role with specific permissions
func (s *AuthorizationService) CreateRole(ctx context.Context, name string, permissions []Permission) (*Role, error) {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Rollback in case of error

	// Insert role
	var roleID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO roles (name, created_at) 
		VALUES ($1, NOW()) 
		RETURNING id
	`, name).Scan(&roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %v", err)
	}

	// Convert permissions to string array for PostgreSQL
	permStrings := make([]string, len(permissions))
	for i, perm := range permissions {
		permStrings[i] = string(perm)
	}

	// Insert role permissions
	_, err = tx.ExecContext(ctx, `
		INSERT INTO role_permissions (role_id, permission)
		VALUES ($1, unnest($2::text[]))
	`, roleID, pq.Array(permStrings))
	if err != nil {
		return nil, fmt.Errorf("failed to assign permissions: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction commit failed: %v", err)
	}

	return &Role{
		ID:          roleID,
		Name:        name,
		Permissions: permissions,
		CreatedAt:   time.Now(),
	}, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *AuthorizationService) GetUserPermissions(ctx context.Context, userID int64) ([]Permission, error) {
	query := `
		SELECT DISTINCT rp.permission
		FROM role_permissions rp
		JOIN roles r ON rp.role_id = r.id
		JOIN users u ON u.role_id = r.id
		WHERE u.id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch permissions: %v", err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err != nil {
			return nil, fmt.Errorf("error scanning permission: %v", err)
		}
		permissions = append(permissions, Permission(perm))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %v", err)
	}

	return permissions, nil
}

// CheckPermission verifies if a user has a specific permission
func (s *AuthorizationService) CheckPermission(ctx context.Context, userID int64, permission Permission) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM role_permissions rp
			JOIN roles r ON rp.role_id = r.id
			JOIN users u ON u.role_id = r.id
			WHERE u.id = $1 AND rp.permission = $2
		)
	`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID, string(permission)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("permission check failed: %v", err)
	}

	return exists, nil
}

// UpdateUserRole changes a user's role
func (s *AuthorizationService) UpdateUserRole(ctx context.Context, userID, newRoleID int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users 
		SET role_id = $1, updated_at = NOW() 
		WHERE id = $2
	`, newRoleID, userID)
	
	if err != nil {
		return fmt.Errorf("failed to update user role: %v", err)
	}

	return nil
}

// PolicyRule defines a dynamic authorization rule
type PolicyRule struct {
	Resource   string
	Action     string
	Condition  func(ctx context.Context, db *sql.DB, userID int64) (bool, error)
}

// DynamicAuthorization implements context-aware authorization
type DynamicAuthorization struct {
	db           *sql.DB
	policyRules  []PolicyRule
}

// NewDynamicAuthorization creates a new dynamic authorization service
func NewDynamicAuthorization(db *sql.DB) *DynamicAuthorization {
	return &DynamicAuthorization{db: db}
}

// AddPolicyRule creates a dynamic authorization rule
func (d *DynamicAuthorization) AddPolicyRule(
	resource, 
	action string, 
	condition func(ctx context.Context, db *sql.DB, userID int64) (bool, error)
) {
	rule := PolicyRule{
		Resource:   resource,
		Action:     action,
		Condition:  condition,
	}
	d.policyRules = append(d.policyRules, rule)
}

// Authorize checks if a user can perform an action on a specific resource
func (d *DynamicAuthorization) Authorize(
	ctx context.Context, 
	userID int64, 
	resource, 
	action string
) (bool, error) {
	for _, rule := range d.policyRules {
		if rule.Resource == resource && rule.Action == action {
			return rule.Condition(ctx, d.db, userID)
		}
	}
	return false, nil
}

// SQL Migration Scripts
const migrationScripts = `
-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create role_permissions table
CREATE TABLE IF NOT EXISTS role_permissions (
	role_id BIGINT REFERENCES roles(id),
	permission VARCHAR(255) NOT NULL,
	PRIMARY KEY (role_id, permission)
);

-- Create users table (assuming it exists, but adding for completeness)
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	username VARCHAR(255) UNIQUE NOT NULL,
	email VARCHAR(255) UNIQUE NOT NULL,
	role_id BIGINT REFERENCES roles(id),
	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE
);

-- Create an index for performance
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
`

// RunMigrations sets up the necessary database schema
func (s *AuthorizationService) RunMigrations(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, migrationScripts)
	if err != nil {
		return fmt.Errorf("migration failed: %v", err)
	}
	return nil
}

// Init auth service
authService := NewAuthorizationService(db)

// Create an admin role
adminRole, err := authService.CreateRole("admin", []Permission{
    PermUserRead, 
    PermUserWrite, 
    PermUserDelete,
    PermProjectCreate,
    PermProjectRead,
    PermProjectUpdate,
    PermProjectDelete
})

// Create a standard user role
userRole, err := authService.CreateRole("user", []Permission{
    PermProjectRead,
    PermResourceView
})

// Check if a user can delete a user
canDelete := authService.CheckPermission(ctx, user, PermUserDelete)

// Use in a handler
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
    user := getCurrentUser(r)
    if !authService.CheckPermission(r.Context(), user, PermUserDelete) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    // Proceed with deletion
}

dynamicAuth := &DynamicAuthorization{}

// Rule: Only project owners can delete their project
dynamicAuth.AddPolicyRule("project", "delete", func(ctx context.Context, user *User) bool {
    projectOwnerId := ctx.Value("projectOwnerId").(uint)
    return projectOwnerId == user.ID
})

// Check authorization
isAuthorized := dynamicAuth.Authorize(ctx, user, "project", "delete")

// Protect routes with specific permissions
deleteUserHandler := authService.AuthorizationMiddleware(PermUserDelete)(http.HandlerFunc(handleUserDeletion))

// Or in a router setup
router.Handle("/users/{id}", 
    authService.AuthorizationMiddleware(PermUserDelete)(userHandler)
)
