package authorization

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"gorm.io/gorm"
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
	gorm.Model
	Name        string       `gorm:"unique;not null"`
	Permissions []Permission `gorm:"-"`
}

// RolePermission is a join table for many-to-many relationship
type RolePermission struct {
	RoleID       uint       `gorm:"primaryKey"`
	PermissionID Permission `gorm:"primaryKey"`
}

// User extended with authorization details
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Email    string `gorm:"unique;not null"`
	RoleID   uint
	Role     Role
}

// AuthorizationService manages role-based access control
type AuthorizationService struct {
	db    *gorm.DB
	cache *sync.Map // Caching role permissions for performance
}

// PolicyRule defines a dynamic authorization rule
type PolicyRule struct {
	Resource  string
	Action    string
	Condition func(ctx context.Context, user *User) bool
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(db *gorm.DB) *AuthorizationService {
	return &AuthorizationService{
		db:    db,
		cache: &sync.Map{},
	}
}

// CreateRole establishes a new role with specific permissions
func (s *AuthorizationService) CreateRole(name string, permissions []Permission) (*Role, error) {
	role := &Role{
		Name:        name,
		Permissions: permissions,
	}

	// Begin a transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create role
	if err := tx.Create(role).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create role: %v", err)
	}

	// Create role-permission mappings
	for _, perm := range permissions {
		rolePermission := RolePermission{
			RoleID:       role.ID,
			PermissionID: perm,
		}
		if err := tx.Create(&rolePermission).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to assign permission: %v", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %v", err)
	}

	// Cache role permissions
	s.cacheRolePermissions(role)

	return role, nil
}

// cacheRolePermissions stores role permissions in memory for quick access
func (s *AuthorizationService) cacheRolePermissions(role *Role) {
	permissionSet := make(map[Permission]bool)
	for _, perm := range role.Permissions {
		permissionSet[perm] = true
	}
	s.cache.Store(role.ID, permissionSet)
}

// CheckPermission verifies if a user has a specific permission
func (s *AuthorizationService) CheckPermission(
	ctx context.Context,
	user *User,
	permission Permission,
) bool {
	// Check cached permissions first
	cachedPerms, ok := s.cache.Load(user.RoleID)
	if ok {
		permMap := cachedPerms.(map[Permission]bool)
		return permMap[permission]
	}

	// Fallback to database check
	var count int64
	err := s.db.Model(&RolePermission{}).
		Where("role_id = ? AND permission_id = ?", user.RoleID, permission).
		Count(&count).Error

	return err == nil && count > 0
}

// DynamicAuthorization implements context-aware authorization
type DynamicAuthorization struct {
	policyRules []PolicyRule
}

// AddPolicyRule creates a dynamic authorization rule
func (d *DynamicAuthorization) AddPolicyRule(
	resource, action string,
	condition func(ctx context.Context, user *User) bool,
) {
	rule := PolicyRule{
		Resource:  resource,
		Action:    action,
		Condition: condition,
	}
	d.policyRules = append(d.policyRules, rule)
}

// Authorize checks if a user can perform an action on a specific resource
func (d *DynamicAuthorization) Authorize(
	ctx context.Context,
	user *User,
	resource, action string,
) bool {
	for _, rule := range d.policyRules {
		if rule.Resource == resource && rule.Action == action {
			return rule.Condition(ctx, user)
		}
	}
	return false
}

// Example of a complex authorization scenario
func ExampleAuthorizationUsage() {
	// Create dynamic authorization rules
	dynamicAuth := &DynamicAuthorization{}

	// Rule: Only project owners can delete their project
	dynamicAuth.AddPolicyRule("project", "delete", func(ctx context.Context, user *User) bool {
		// Hypothetical context value containing project owner ID
		projectOwnerId, ok := ctx.Value("projectOwnerId").(uint)
		return ok && projectOwnerId == user.ID
	})

	// Rule: Managers can modify resources in their department
	dynamicAuth.AddPolicyRule("resource", "modify", func(ctx context.Context, user *User) bool {
		// Assume context contains user's department
		userDepartment, ok := ctx.Value("userDepartment").(string)
		resourceDepartment, ok2 := ctx.Value("resourceDepartment").(string)
		return ok && ok2 && userDepartment == resourceDepartment
	})
}

// Middleware for authorization in HTTP handlers
func (s *AuthorizationService) AuthorizationMiddleware(
	requiredPermission Permission,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user from context (set by authentication middleware)
			user, ok := r.Context().Value("user").(*User)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check permission
			if !s.CheckPermission(r.Context(), user, requiredPermission) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
