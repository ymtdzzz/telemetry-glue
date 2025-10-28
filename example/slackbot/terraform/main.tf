module "secret_manager" {
  source = "./modules/secret-manager"

  project_id  = var.project_id
  environment = var.environment
  prefix      = var.prefix
}

module "cloud_function" {
  source = "./modules/cloud-function"

  project_id  = var.project_id
  region      = var.region
  environment = var.environment
  prefix      = var.prefix

  secret_ids = {
    slack_bot_token          = module.secret_manager.slack_bot_token_secret_id
    slack_verification_token = module.secret_manager.slack_verification_token_secret_id
    new_relic_api_key        = module.secret_manager.new_relic_api_key_secret_id
    new_relic_account_id     = module.secret_manager.new_relic_account_id_secret_id
  }
}
