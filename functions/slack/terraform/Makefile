# Terraform Makefile for Slack Functions
# Usage: make plan, make apply, make destroy

# Default environment
ENV ?= dev

# Terraform variables
TF_VAR_FILE := environments/$(ENV).tfvars
TF_PLAN_FILE := terraform-$(ENV).tfplan
TF_BACKEND_FILE := backend-$(ENV).hcl

# Colors for output
YELLOW := \033[1;33m
GREEN := \033[1;32m
RED := \033[1;31m
NC := \033[0m # No Color

.PHONY: help init plan apply destroy clean fmt validate

help: ## Show this help message
	@echo "$(YELLOW)Available commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

init: ## Initialize Terraform with GCS backend
	@echo "$(YELLOW)Initializing Terraform with GCS backend for $(ENV) environment...$(NC)"
	@if [ ! -f $(TF_BACKEND_FILE) ]; then \
		echo "$(RED)Error: $(TF_BACKEND_FILE) not found$(NC)"; \
		echo "$(YELLOW)Please create $(TF_BACKEND_FILE) with your bucket configuration$(NC)"; \
		exit 1; \
	fi
	terraform init -backend-config=$(TF_BACKEND_FILE) -reconfigure

init-local: ## Initialize Terraform without backend (local state)
	@echo "$(YELLOW)Initializing Terraform with local state...$(NC)"
	terraform init

migrate-state: ## Migrate local state to GCS backend
	@echo "$(YELLOW)Migrating state to GCS backend for $(ENV) environment...$(NC)"
	@if [ ! -f $(TF_BACKEND_FILE) ]; then \
		echo "$(RED)Error: $(TF_BACKEND_FILE) not found$(NC)"; \
		exit 1; \
	fi
	terraform init -backend-config=$(TF_BACKEND_FILE) -migrate-state

plan: ## Create execution plan for the specified environment
	@echo "$(YELLOW)Creating Terraform plan for $(ENV) environment...$(NC)"
	@if [ ! -f $(TF_VAR_FILE) ]; then \
		echo "$(RED)Error: $(TF_VAR_FILE) not found$(NC)"; \
		exit 1; \
	fi
	terraform plan -var-file=$(TF_VAR_FILE) -out=$(TF_PLAN_FILE)
	@echo "$(GREEN)Plan saved to $(TF_PLAN_FILE)$(NC)"

apply: ## Apply the Terraform plan
	@echo "$(YELLOW)Applying Terraform changes for $(ENV) environment...$(NC)"
	@if [ ! -f $(TF_PLAN_FILE) ]; then \
		echo "$(RED)Error: $(TF_PLAN_FILE) not found. Run 'make plan' first.$(NC)"; \
		exit 1; \
	fi
	terraform apply $(TF_PLAN_FILE)
	@echo "$(GREEN)Successfully applied changes for $(ENV) environment$(NC)"

apply-auto: ## Apply without plan file (auto-approve)
	@echo "$(YELLOW)Auto-applying Terraform changes for $(ENV) environment...$(NC)"
	@if [ ! -f $(TF_VAR_FILE) ]; then \
		echo "$(RED)Error: $(TF_VAR_FILE) not found$(NC)"; \
		exit 1; \
	fi
	terraform apply -var-file=$(TF_VAR_FILE) -auto-approve
	@echo "$(GREEN)Successfully applied changes for $(ENV) environment$(NC)"

destroy: ## Destroy the infrastructure
	@echo "$(RED)Destroying infrastructure for $(ENV) environment...$(NC)"
	@if [ ! -f $(TF_VAR_FILE) ]; then \
		echo "$(RED)Error: $(TF_VAR_FILE) not found$(NC)"; \
		exit 1; \
	fi
	terraform destroy -var-file=$(TF_VAR_FILE)

fmt: ## Format Terraform files
	@echo "$(YELLOW)Formatting Terraform files...$(NC)"
	terraform fmt -recursive

validate: ## Validate Terraform configuration
	@echo "$(YELLOW)Validating Terraform configuration...$(NC)"
	terraform validate

clean: ## Clean up plan files
	@echo "$(YELLOW)Cleaning up plan files...$(NC)"
	rm -f terraform-*.tfplan

show: ## Show current state
	terraform show

output: ## Show output values
	terraform output

# Environment-specific shortcuts
plan-dev: ## Plan for dev environment
	@$(MAKE) plan ENV=dev

apply-dev: ## Apply for dev environment
	@$(MAKE) apply ENV=dev

plan-staging: ## Plan for staging environment
	@$(MAKE) plan ENV=staging

apply-staging: ## Apply for staging environment
	@$(MAKE) apply ENV=staging

plan-prod: ## Plan for production environment
	@$(MAKE) plan ENV=prod

apply-prod: ## Apply for production environment
	@$(MAKE) apply ENV=prod

# Quick workflow
dev: init plan-dev apply-dev ## Complete dev workflow: init + plan + apply

# Show current environment
env: ## Show current environment
	@echo "$(GREEN)Current environment: $(ENV)$(NC)"
	@echo "$(GREEN)Using var file: $(TF_VAR_FILE)$(NC)"
	@echo "$(GREEN)Using backend file: $(TF_BACKEND_FILE)$(NC)"