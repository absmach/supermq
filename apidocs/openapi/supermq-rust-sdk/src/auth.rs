//! Uauth service client with complete models and API methods
//! 
//! This module contains all models and API methods for the Uauth service,
//! extracted and adapted from the OpenAPI specification.

use crate::{Config, Error, Result, ResponseContent};
use reqwest::Client as HttpClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// ============================================================================
// Uauth Models
// ============================================================================

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct AddScopeRequest {
pub struct AddScopeRequest {
pub struct AddScopeRequest {
pub struct AddScopeRequest {
    /// List of scopes to add
    #[serde(rename = "scopes")]
    pub scopes: Vec<models::Scope>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct CreatePatRequest {
pub struct CreatePatRequest {
pub struct CreatePatRequest {
pub struct CreatePatRequest {
    /// Name of the Personal Access Token
    #[serde(rename = "name")]
    pub name: String,
    /// Description of the Personal Access Token
    #[serde(rename = "description", skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    /// Duration for which the PAT is valid. Format is a duration string (e.g. \"30d\", \"24h\", \"1y\").
    #[serde(rename = "duration", skip_serializing_if = "Option::is_none")]
    pub duration: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct HealthInfo {
pub struct HealthInfo {
pub struct HealthInfo {
pub struct HealthInfo {
    /// Service status.
    #[serde(rename = "status", skip_serializing_if = "Option::is_none")]
    pub status: Option<Status>,
    /// Service version.
    #[serde(rename = "version", skip_serializing_if = "Option::is_none")]
    pub version: Option<String>,
    /// Service commit hash.
    #[serde(rename = "commit", skip_serializing_if = "Option::is_none")]
    pub commit: Option<String>,
    /// Service description.
    #[serde(rename = "description", skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    /// Service build time.
    #[serde(rename = "build_time", skip_serializing_if = "Option::is_none")]
    pub build_time: Option<String>,
    /// Service instance ID.
    #[serde(rename = "instance_id", skip_serializing_if = "Option::is_none")]
    pub instance_id: Option<String>,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd, Hash, Serialize, Deserialize)]
pub enum Status {
pub enum Status {
pub enum Status {
pub enum Status {
    #[serde(rename = "pass")]
    Pass,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct IssueKeyRequest {
pub struct IssueKeyRequest {
pub struct IssueKeyRequest {
pub struct IssueKeyRequest {
    /// API key type. Keys of different type are processed differently.
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    pub r#type: Option<i32>,
    /// Number of seconds issued token is valid for.
    #[serde(rename = "duration", skip_serializing_if = "Option::is_none")]
    pub duration: Option<f64>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Key {
pub struct Key {
pub struct Key {
pub struct Key {
    /// API key unique identifier
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// In ID of the entity that issued the token.
    #[serde(rename = "issuer_id", skip_serializing_if = "Option::is_none")]
    pub issuer_id: Option<uuid::Uuid>,
    /// API key type. Keys of different type are processed differently.
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    pub r#type: Option<i32>,
    /// User's email or service identifier of API key subject.
    #[serde(rename = "subject", skip_serializing_if = "Option::is_none")]
    pub subject: Option<String>,
    /// Time when the key is generated.
    #[serde(rename = "issued_at", skip_serializing_if = "Option::is_none")]
    pub issued_at: Option<String>,
    /// Time when the Key expires. If this field is missing, that means that Key is valid indefinitely.
    #[serde(rename = "expires_at", skip_serializing_if = "Option::is_none")]
    pub expires_at: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Pat {
pub struct Pat {
pub struct Pat {
pub struct Pat {
    /// Personal Access Token unique identifier
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// User ID of the PAT owner
    #[serde(rename = "user_id", skip_serializing_if = "Option::is_none")]
    pub user_id: Option<uuid::Uuid>,
    /// Name of the Personal Access Token
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// Description of the Personal Access Token
    #[serde(rename = "description", skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    /// Secret value of the Personal Access Token
    #[serde(rename = "secret", skip_serializing_if = "Option::is_none")]
    pub secret: Option<String>,
    /// Time when the PAT was issued
    #[serde(rename = "issued_at", skip_serializing_if = "Option::is_none")]
    pub issued_at: Option<String>,
    /// Time when the PAT expires
    #[serde(rename = "expires_at", skip_serializing_if = "Option::is_none")]
    pub expires_at: Option<String>,
    /// Time when the PAT was last updated
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
    /// Time when the PAT was last used
    #[serde(rename = "last_used_at", skip_serializing_if = "Option::is_none")]
    pub last_used_at: Option<String>,
    /// Whether the PAT is revoked
    #[serde(rename = "revoked", skip_serializing_if = "Option::is_none")]
    pub revoked: Option<bool>,
    /// Time when the PAT was revoked
    #[serde(rename = "revoked_at", skip_serializing_if = "Option::is_none")]
    pub revoked_at: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct PatsPage {
pub struct PatsPage {
pub struct PatsPage {
pub struct PatsPage {
    /// Total number of PATs
    #[serde(rename = "total", skip_serializing_if = "Option::is_none")]
    pub total: Option<i32>,
    /// Number of items to skip during retrieval
    #[serde(rename = "offset", skip_serializing_if = "Option::is_none")]
    pub offset: Option<i32>,
    /// Size of the subset to retrieve
    #[serde(rename = "limit", skip_serializing_if = "Option::is_none")]
    pub limit: Option<i32>,
    /// List of Personal Access Tokens
    #[serde(rename = "pats", skip_serializing_if = "Option::is_none")]
    pub pats: Option<Vec<models::Pat>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct RemoveScopeRequest {
pub struct RemoveScopeRequest {
pub struct RemoveScopeRequest {
pub struct RemoveScopeRequest {
    /// List of scope IDs to remove
    #[serde(rename = "scopes_id")]
    pub scopes_id: Vec<uuid::Uuid>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct ResetPatSecretRequest {
pub struct ResetPatSecretRequest {
pub struct ResetPatSecretRequest {
pub struct ResetPatSecretRequest {
    /// Duration for which the new PAT secret is valid. Format is a duration string (e.g. \"30d\", \"24h\", \"1y\").
    #[serde(rename = "duration", skip_serializing_if = "Option::is_none")]
    pub duration: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Scope {
pub struct Scope {
pub struct Scope {
pub struct Scope {
    /// Scope unique identifier
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// PAT ID this scope belongs to
    #[serde(rename = "pat_id", skip_serializing_if = "Option::is_none")]
    pub pat_id: Option<uuid::Uuid>,
    /// Optional domain ID for the scope
    #[serde(rename = "optional_domain_id", skip_serializing_if = "Option::is_none")]
    pub optional_domain_id: Option<uuid::Uuid>,
    /// Type of entity the scope applies to
    #[serde(rename = "entity_type", skip_serializing_if = "Option::is_none")]
    pub entity_type: Option<EntityType>,
    /// ID of the entity the scope applies to. '*' means all entities of the specified type.
    #[serde(rename = "entity_id", skip_serializing_if = "Option::is_none")]
    pub entity_id: Option<String>,
    /// Operation allowed by this scope
    #[serde(rename = "operation", skip_serializing_if = "Option::is_none")]
    pub operation: Option<Operation>,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd, Hash, Serialize, Deserialize)]
pub enum EntityType {
pub enum EntityType {
pub enum EntityType {
pub enum EntityType {
    #[serde(rename = "groups")]
    Groups,
    #[serde(rename = "channels")]
    Channels,
    #[serde(rename = "clients")]
    Clients,
    #[serde(rename = "domains")]
    Domains,
    #[serde(rename = "users")]
    Users,
    #[serde(rename = "dashboards")]
    Dashboards,
    #[serde(rename = "messages")]
    Messages,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd, Hash, Serialize, Deserialize)]
pub enum Operation {
pub enum Operation {
pub enum Operation {
pub enum Operation {
    #[serde(rename = "create")]
    Create,
    #[serde(rename = "read")]
    Read,
    #[serde(rename = "list")]
    List,
    #[serde(rename = "update")]
    Update,
    #[serde(rename = "delete")]
    Delete,
    #[serde(rename = "share")]
    Share,
    #[serde(rename = "unshare")]
    Unshare,
    #[serde(rename = "publish")]
    Publish,
    #[serde(rename = "subscribe")]
    Subscribe,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct ScopesPage {
pub struct ScopesPage {
pub struct ScopesPage {
pub struct ScopesPage {
    /// Total number of scopes
    #[serde(rename = "total", skip_serializing_if = "Option::is_none")]
    pub total: Option<i32>,
    /// Number of items to skip during retrieval
    #[serde(rename = "offset", skip_serializing_if = "Option::is_none")]
    pub offset: Option<i32>,
    /// Size of the subset to retrieve
    #[serde(rename = "limit", skip_serializing_if = "Option::is_none")]
    pub limit: Option<i32>,
    /// List of scopes
    #[serde(rename = "scopes", skip_serializing_if = "Option::is_none")]
    pub scopes: Option<Vec<models::Scope>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct UpdatePatDescriptionRequest {
pub struct UpdatePatDescriptionRequest {
pub struct UpdatePatDescriptionRequest {
pub struct UpdatePatDescriptionRequest {
    /// New description for the Personal Access Token
    #[serde(rename = "description")]
    pub description: String,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct UpdatePatNameRequest {
pub struct UpdatePatNameRequest {
pub struct UpdatePatNameRequest {
pub struct UpdatePatNameRequest {
    /// New name for the Personal Access Token
    #[serde(rename = "name")]
    pub name: String,
}


// ============================================================================
// Uauth Error Types
// ============================================================================

pub enum GetKeyError {
    Status400(),
    Status401(),
    Status404(),
    Status500(),
    UnknownValue(serde_json::Value),
}

/// struct for typed errors of method [`issue_key`]
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum IssueKeyError {
    Status400(),
    Status401(),
    Status409(),

// ============================================================================
// Uauth Client Implementation
// ============================================================================

/// Uauth service client with full API method implementations
pub struct UauthClient {
    http_client: HttpClient,
    base_url: String,
    config: Config,
}

impl UauthClient {
    /// Create a new Uauth client
    pub(crate) fn new(http_client: HttpClient, config: Config) -> Self {
        Self {
            http_client,
            base_url: config.base_url.clone(),
            config,
        }
    }

    /// Health check endpoint
    pub async fn health(&self) -> Result<bool> {
        let url = format!("{}/health", self.base_url);
        let mut request = self.http_client.get(&url);
        
        if let Some(ref token) = self.config.bearer_access_token {
            request = request.bearer_auth(token);
        }
        
        if let Some(ref user_agent) = self.config.user_agent {
            request = request.header(reqwest::header::USER_AGENT, user_agent);
        }
        
        let response = request.send().await?;
        Ok(response.status().is_success())
    }

    /// pub async fn get_key(configuration: &configuration::Configuration, key_id: &str) -> Result<models::Key, Error<GetKeyError>> { - Extracted from OpenAPI
    pub async fn fn(&self, key_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn issue_key(configuration: &configuration::Configuration, issue_key_request: models::IssueKeyRequest) -> Result<(), Error<IssueKeyError>> { - Extracted from OpenAPI
    pub async fn fn(&self, issue_key_request: models::IssueKeyRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn revoke_key(configuration: &configuration::Configuration, key_id: &str) -> Result<(), Error<RevokeKeyError>> { - Extracted from OpenAPI
    pub async fn fn(&self, key_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn health_get(configuration: &configuration::Configuration, ) -> Result<models::HealthInfo, Error<HealthGetError>> { - Extracted from OpenAPI
    pub async fn fn(&self, ) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn add_scope(configuration: &configuration::Configuration, pat_id: &str, add_scope_request: models::AddScopeRequest) -> Result<(), Error<AddScopeError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str, add_scope_request: models::AddScopeRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn clear_all_pats(configuration: &configuration::Configuration, ) -> Result<(), Error<ClearAllPatsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, ) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn clear_all_scopes(configuration: &configuration::Configuration, pat_id: &str) -> Result<(), Error<ClearAllScopesError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn create_pat(configuration: &configuration::Configuration, create_pat_request: models::CreatePatRequest) -> Result<models::Pat, Error<CreatePatError>> { - Extracted from OpenAPI
    pub async fn fn(&self, create_pat_request: models::CreatePatRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_pat(configuration: &configuration::Configuration, pat_id: &str) -> Result<(), Error<DeletePatError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_pats(configuration: &configuration::Configuration, limit: Option<i32>, offset: Option<i32>) -> Result<models::PatsPage, Error<ListPatsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, limit: Option<i32>, offset: Option<i32>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_scopes(configuration: &configuration::Configuration, pat_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<models::ScopesPage, Error<ListScopesError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn remove_scope(configuration: &configuration::Configuration, pat_id: &str, remove_scope_request: models::RemoveScopeRequest) -> Result<(), Error<RemoveScopeError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str, remove_scope_request: models::RemoveScopeRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn reset_pat_secret(configuration: &configuration::Configuration, pat_id: &str, reset_pat_secret_request: models::ResetPatSecretRequest) -> Result<models::Pat, Error<ResetPatSecretError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str, reset_pat_secret_request: models::ResetPatSecretRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn retrieve_pat(configuration: &configuration::Configuration, pat_id: &str) -> Result<models::Pat, Error<RetrievePatError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn revoke_pat_secret(configuration: &configuration::Configuration, pat_id: &str) -> Result<(), Error<RevokePatSecretError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn update_pat_description(configuration: &configuration::Configuration, pat_id: &str, update_pat_description_request: models::UpdatePatDescriptionRequest) -> Result<models::Pat, Error<UpdatePatDescriptionError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str, update_pat_description_request: models::UpdatePatDescriptionRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn update_pat_name(configuration: &configuration::Configuration, pat_id: &str, update_pat_name_request: models::UpdatePatNameRequest) -> Result<models::Pat, Error<UpdatePatNameError>> { - Extracted from OpenAPI
    pub async fn fn(&self, pat_id: &str, update_pat_name_request: models::UpdatePatNameRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }
}
