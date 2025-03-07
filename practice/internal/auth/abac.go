package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Core ABAC Types
type Policy struct {
	ID          string
	Description string
	Effect      string // "allow" or "deny"
	Conditions  []Condition
}

type Condition struct {
	Attribute string
	Operator  string
	Value     interface{}
}

type Request struct {
	User     map[string]interface{}
	Resource map[string]interface{}
	Action   string
	Context  map[string]interface{}
}

// Policy Repository
type PolicyRepository interface {
	GetPolicies() ([]Policy, error)
	GetPolicyByID(id string) (Policy, error)
	SavePolicy(policy Policy) error
}

type InMemoryPolicyRepository struct {
	policies map[string]Policy
}

func NewInMemoryPolicyRepository() *InMemoryPolicyRepository {
	return &InMemoryPolicyRepository{
		policies: make(map[string]Policy),
	}
}

func (repo *InMemoryPolicyRepository) GetPolicies() ([]Policy, error) {
	result := make([]Policy, 0, len(repo.policies))
	for _, policy := range repo.policies {
		result = append(result, policy)
	}
	return result, nil
}

func (repo *InMemoryPolicyRepository) GetPolicyByID(id string) (Policy, error) {
	if policy, ok := repo.policies[id]; ok {
		return policy, nil
	}
	return Policy{}, fmt.Errorf("policy not found: %s", id)
}

func (repo *InMemoryPolicyRepository) SavePolicy(policy Policy) error {
	repo.policies[policy.ID] = policy
	return nil
}

// Policy Evaluation
func EvaluatePolicy(policy Policy, request Request) bool {
	matches := evaluateConditions(policy.Conditions, request)

	if policy.Effect == "allow" {
		return matches
	} else {
		return matches // For deny policies, return true when conditions match (to deny)
	}
}

func evaluateConditions(conditions []Condition, request Request) bool {
	for _, condition := range conditions {
		if !evaluateCondition(condition, request) {
			return false
		}
	}
	return true
}

func evaluateCondition(condition Condition, request Request) bool {
	parts := strings.Split(condition.Attribute, ".")
	if len(parts) != 2 {
		log.Printf("Invalid attribute format: %s", condition.Attribute)
		return false
	}

	var value interface{}
	switch parts[0] {
	case "user":
		value = request.User[parts[1]]
	case "resource":
		value = request.Resource[parts[1]]
	case "action":
		value = request.Action
	case "context":
		value = request.Context[parts[1]]
	default:
		log.Printf("Unknown attribute category: %s", parts[0])
		return false
	}

	// Handle nil values
	if value == nil {
		return false
	}

	switch condition.Operator {
	case "equals":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", condition.Value)
	case "contains":
		strValue, ok := value.(string)
		if !ok {
			return false
		}
		strCondValue, ok := condition.Value.(string)
		if !ok {
			return false
		}
		return strings.Contains(strValue, strCondValue)
	case "startsWith":
		strValue, ok := value.(string)
		if !ok {
			return false
		}
		strCondValue, ok := condition.Value.(string)
		if !ok {
			return false
		}
		return strings.HasPrefix(strValue, strCondValue)
	case "greaterThan":
		numValue, ok := value.(float64)
		if !ok {
			return false
		}
		numCondValue, ok := condition.Value.(float64)
		if !ok {
			return false
		}
		return numValue > numCondValue
	}

	log.Printf("Unsupported operator: %s", condition.Operator)
	return false
}

// ABAC Service
type AbacService struct {
	policyRepo PolicyRepository
}

func NewAbacService(repo PolicyRepository) *AbacService {
	return &AbacService{policyRepo: repo}
}

func (s *AbacService) Authorize(request Request) bool {
	policies, err := s.policyRepo.GetPolicies()
	if err != nil {
		log.Printf("Error fetching policies: %v", err)
		return false // Fail closed
	}

	// Apply deny policies first
	for _, policy := range policies {
		if policy.Effect == "deny" && EvaluatePolicy(policy, request) {
			return false
		}
	}

	// Then check for allow policies
	for _, policy := range policies {
		if policy.Effect == "allow" && EvaluatePolicy(policy, request) {
			return true
		}
	}

	return false // Deny by default
}

// HTTP Integration
func extractUserFromContext(ctx http.Request) map[string]interface{} {
	// In a real app, you'd extract user from JWT or session
	// This is just a placeholder
	return map[string]interface{}{
		"id":         "user123",
		"role":       "admin",
		"department": "engineering",
	}
}

func getResourceType(path string) string {
	// Simple implementation - in reality, you might map paths to resource types
	if strings.HasPrefix(path, "/api/documents") {
		return "document"
	} else if strings.HasPrefix(path, "/api/users") {
		return "user"
	}
	return "unknown"
}

func AbacMiddleware(abacService *AbacService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := extractUserFromContext(*r)
			resource := map[string]interface{}{
				"path": r.URL.Path,
				"type": getResourceType(r.URL.Path),
			}
			action := r.Method
			context := map[string]interface{}{
				"ip":        r.RemoteAddr,
				"timestamp": time.Now().Unix(),
			}

			request := Request{
				User:     user,
				Resource: resource,
				Action:   action,
				Context:  context,
			}

			log.Printf("Evaluating request: %+v", request)

			if abacService.Authorize(request) {
				next.ServeHTTP(w, r)
			} else {
				log.Printf("Access denied for request: %+v", request)
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		})
	}
}

// Policy Administration API
func handleGetPolicies(repo PolicyRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		policies, err := repo.GetPolicies()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(policies)
	}
}

func handleCreatePolicy(repo PolicyRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var policy Policy
		if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := repo.SavePolicy(policy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// Example usage
func main() {
	// Initialize repositories
	policyRepo := NewInMemoryPolicyRepository()

	// Create sample policies
	adminDocumentPolicy := Policy{
		ID:          "admin-document-access",
		Description: "Admins can access all documents",
		Effect:      "allow",
		Conditions: []Condition{
			{Attribute: "user.role", Operator: "equals", Value: "admin"},
			{Attribute: "resource.type", Operator: "equals", Value: "document"},
		},
	}

	denyFinanceDocumentsPolicy := Policy{
		ID:          "deny-finance-documents",
		Description: "Only finance department can access finance documents",
		Effect:      "deny",
		Conditions: []Condition{
			{Attribute: "resource.type", Operator: "equals", Value: "document"},
			{Attribute: "resource.path", Operator: "contains", Value: "finance"},
			{Attribute: "user.department", Operator: "equals", Value: "engineering"},
		},
	}

	// Save policies
	policyRepo.SavePolicy(adminDocumentPolicy)
	policyRepo.SavePolicy(denyFinanceDocumentsPolicy)

	// Create ABAC service
	abacService := NewAbacService(policyRepo)

	// Create API handlers
	mux := http.NewServeMux()

	// Protected resources
	documentsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Authorized access to documents")
	})

	financeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Authorized access to finance documents")
	})

	// Policy administration
	mux.HandleFunc("/policies", handleGetPolicies(policyRepo))
	mux.HandleFunc("/policies/create", handleCreatePolicy(policyRepo))

	// Protected endpoints with ABAC
	mux.Handle("/api/documents", AbacMiddleware(abacService)(documentsHandler))
	mux.Handle("/api/documents/finance", AbacMiddleware(abacService)(financeHandler))

	// Start server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
