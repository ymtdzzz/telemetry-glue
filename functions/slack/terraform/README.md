# Terraform State Management for Slack Functions

This is a simplified Terraform configuration for setting up state management and basic infrastructure.

## Current Resources

- Storage bucket for future Terraform state (if needed)
- Test storage bucket to verify Terraform is working

## Setup

1. Create `backend-dev.hcl` with your actual GCS bucket name:

   ```hcl
   bucket = "your-actual-terraform-state-bucket-name"
   prefix = "slack-functions/dev"
   ```

2. Create `environments/dev.tfvars` with your actual project ID:

   ```hcl
   project_id = "your-actual-project-id"
   ```

3. Initialize Terraform with GCS backend:

   ```bash
   make init
   ```

4. Plan the deployment:

   ```bash
   make plan
   ```

5. Apply the configuration:
   ```bash
   make apply
   ```

## State Migration

If you already have local state and want to migrate to GCS:

```bash
# First, ensure backend-dev.hcl is configured
make migrate-state
```

## Available Make Commands

```bash
make help           # Show all available commands
make init           # Initialize with GCS backend
make init-local     # Initialize with local state
make migrate-state  # Migrate local state to GCS
make plan           # Create execution plan
make apply          # Apply the plan
make destroy        # Destroy infrastructure
make fmt            # Format Terraform files
make validate       # Validate configuration
```

## Required Permissions

The user or service account running Terraform needs:

- Storage Admin role
- Basic permissions to create/manage storage buckets

## Next Steps

After this basic setup works:

1. Add Secret Manager resources
2. Add IAM resources
3. Add Cloud Functions resources
