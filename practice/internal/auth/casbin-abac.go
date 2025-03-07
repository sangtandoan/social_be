# Policy definitions for ABAC model
# Format: p, sub_rule, obj_rule, act_rule, eft

# Admin role has full access to all resources
p, r.role == "admin", *, *, allow

# Department-specific document access
p, r.department == "finance", r.resource_type == "financial_document", r.action == "read", allow
p, r.department == "finance" && r.role == "manager", r.resource_type == "financial_document", "write", allow
p, r.department == "hr", r.resource_type == "employee_record", r.action == "read", allow
p, r.department == "hr" && r.role == "manager", r.resource_type == "employee_record", r.action == "write", allow

# Project-based access
p, r.role == "developer" && r.project_id == r.resource_project_id, r.resource_type == "code_repository", "read", allow
p, r.role == "senior_developer" && r.project_id == r.resource_project_id, r.resource_type == "code_repository", "write", allow
p, r.role == "project_manager" && r.project_id == r.resource_project_id, r.resource_type == "project_plan", "write", allow

# Time-based access restrictions
p, r.role == "contractor" && r.login_time >= 9 && r.login_time <= 17, r.resource_type == "company_network", "access", allow

# Geolocation-based access
p, r.role == "remote_worker" && r.ip_prefix == "10.0.0", r.resource_type == "vpn", "connect", allow
p, r.location == "office", r.resource_type == "internal_system", "access", allow

# Multi-factor authentication conditions
p, r.role == "executive" && r.mfa_verified == true, r.resource_type == "financial_report", "read", allow
p, r.role == "executive" && r.mfa_verified == true && r.security_clearance >= 3, r.resource_type == "strategic_plan", "read", allow

# Content-based restrictions
p, r.department == "legal", r.resource_type == "contract" && r.resource_status == "draft", "write", allow
p, r.department == "legal" && r.role == "paralegal", r.resource_type == "contract" && r.resource_status == "approved", "read", allow
p, r.department == "legal" && r.role == "attorney", r.resource_type == "contract", "approve", allow

# Customer data access with regional compliance
p, r.department == "support" && r.region == r.customer_region, r.resource_type == "customer_data", "read", allow
p, r.department == "support" && r.role == "supervisor" && r.gdpr_trained == true, r.resource_type == "customer_data" && r.customer_region == "EU", "write", allow

# Explicit deny policies (these take precedence over allow policies)
p, r.security_clearance < r.resource_classification, *, *, deny
p, r.role == "external_auditor" && r.audit_period_active == false, *, *, deny
p, r.compliance_training_completed == false, r.resource_type =~ "regulated_.*", *, deny

# Dynamic resource ownership
p, r.user_id == r.resource_owner_id, r.resource_type == "personal_document", "read", allow
p, r.user_id == r.resource_owner_id, r.resource_type == "personal_document", "write", allow
p, r.user_id == r.resource_owner_id, r.resource_type == "personal_document", "delete", allow

# Team-based access control
p, r.team_id == r.resource_team_id, r.resource_type == "team_document", "read", allow
p, r.team_id == r.resource_team_id && r.role == "team_leader", r.resource_type == "team_document", "write", allow

# Attribute combinations for API access
p, r.role == "api_user" && r.api_key_valid == true && r.rate_limit_exceeded == false, r.resource_type == "api" && r.resource_tier <= r.subscription_tier, "call", allow

# Temporary access grants
p, r.role == "temporary_consultant" && r.contract_end_date > r.current_date, r.resource_type == "project_data", "read", allow

# Compliance and audit requirements
p, r.role == "auditor" && r.audit_scope == r.resource_department, r.resource_type == "audit_log", "read", allow
p, r.department == "compliance" && r.certification == "iso27001", r.resource_type == "compliance_document", "write", allow

# Emergency access patterns
p, r.role == "emergency_responder" && r.emergency_mode == true, *, *, allow
