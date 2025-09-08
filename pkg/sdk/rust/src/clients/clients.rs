// SuperMQ Clients Service Rust SDK
// This SDK provides a complete interface to the SuperMQ Clients Service API

use reqwest::{Client as HttpClient, Error as ReqwestError, Response};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;
use uuid::Uuid;

#[derive(Error, Debug)]
pub enum ClientsError {
    #[error("HTTP request failed: {0}")]
    Http(#[from] ReqwestError),
    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),
    #[error("API error: {0}")]
    Api(String),
    #[error("Authentication failed: {message}")]
    Auth { message: String },
    #[error("Not found: {resource}")]
    NotFound { resource: String },
    #[error("Validation error: {message}")]
    Validation { message: String },
    #[error("Conflict: {message}")]
    Conflict { message: String },
}

// Core Data Structures

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Credentials {
    pub identity: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub secret: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientRequest {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub tags: Vec<String>,
    pub credentials: Credentials,
    #[serde(default, skip_serializing_if = "HashMap::is_empty")]
    pub metadata: HashMap<String, serde_json::Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Client {
    pub id: Uuid,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default)]
    pub tags: Vec<String>,
    pub domain_id: Uuid,
    pub credentials: Credentials,
    #[serde(default)]
    pub metadata: HashMap<String, serde_json::Value>,
    pub status: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientWithEmptySecret {
    pub id: Uuid,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(default)]
    pub tags: Vec<String>,
    pub domain_id: Uuid,
    pub credentials: Credentials,
    #[serde(default)]
    pub metadata: HashMap<String, serde_json::Value>,
    pub status: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientsPage {
    pub clients: Vec<ClientWithEmptySecret>,
    pub total: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub offset: Option<u64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub limit: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientUpdate {
    pub name: String,
    pub metadata: HashMap<String, serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientTags {
    pub tags: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientSecret {
    pub secret: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParentGroupRequest {
    pub parent_group_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: String,
    pub commit: String,
    pub description: String,
    pub build_time: String,
}

// Role-related structures
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Role {
    pub id: String,
    pub name: String,
    #[serde(default)]
    pub description: Option<String>,
    #[serde(default)]
    pub actions: Vec<String>,
    #[serde(default)]
    pub members: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRoleRequest {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub actions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateRoleRequest {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoleActionsRequest {
    pub actions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RoleMembersRequest {
    pub members: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RolesPage {
    pub roles: Vec<Role>,
    pub total: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub offset: Option<u64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub limit: Option<u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AvailableActionsResponse {
    pub actions: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EntityMembers {
    pub members: Vec<String>,
}

// Query parameters for listing clients
#[derive(Debug, Clone, Default)]
pub struct ClientListParams {
    pub limit: Option<u64>,
    pub offset: Option<u64>,
    pub metadata: Option<String>,
    pub status: Option<String>,
    pub name: Option<String>,
    pub tags: Option<Vec<String>>,
}

// Main SDK Client
pub struct ClientsServiceClient {
    http_client: HttpClient,
    base_url: String,
    auth_token: Option<String>,
}

impl ClientsServiceClient {
    /// Create a new client instance
    pub fn new(base_url: String) -> Self {
        Self {
            http_client: HttpClient::new(),
            base_url: base_url.trim_end_matches('/').to_string(),
            auth_token: None,
        }
    }

    /// Set the authentication token
    pub fn with_auth_token(mut self, token: String) -> Self {
        self.auth_token = Some(token);
        self
    }

    /// Set the authentication token (mutable)
    pub fn set_auth_token(&mut self, token: String) {
        self.auth_token = Some(token);
    }

    /// Build request with authentication headers
    fn build_request(&self, method: reqwest::Method, url: &str) -> reqwest::RequestBuilder {
        let mut request = self.http_client.request(method, url);

        if let Some(ref token) = self.auth_token {
            request = request.bearer_auth(token);
        }

        request.header("Content-Type", "application/json")
    }

    /// Handle API response and convert to appropriate error
    async fn handle_response<T: for<'de> Deserialize<'de>>(
        response: Response,
    ) -> Result<T, ClientsError> {
        let status = response.status();

        if status.is_success() {
            let body = response.json::<T>().await?;
            Ok(body)
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());

            match status.as_u16() {
                401 => Err(ClientsError::Auth {
                    message: error_text,
                }),
                404 => Err(ClientsError::NotFound {
                    resource: error_text,
                }),
                409 => Err(ClientsError::Conflict {
                    message: error_text,
                }),
                400 | 422 => Err(ClientsError::Validation {
                    message: error_text,
                }),
                _ => Err(ClientsError::Api(error_text)),
            }
        }
    }

    // Client CRUD Operations

    /// Create a new client
    pub async fn create_client(
        &self,
        domain_id: Uuid,
        client: ClientRequest,
    ) -> Result<Client, ClientsError> {
        let url = format!("{}/{}/clients", self.base_url, domain_id);
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&client)
            .send()
            .await?;

        Self::handle_response(response).await
    }

    /// List clients with optional filtering
    pub async fn list_clients(
        &self,
        domain_id: Uuid,
        params: Option<ClientListParams>,
    ) -> Result<ClientsPage, ClientsError> {
        let mut url = format!("{}/{}/clients", self.base_url, domain_id);

        if let Some(params) = params {
            let mut query_params = Vec::new();

            if let Some(limit) = params.limit {
                query_params.push(format!("limit={}", limit));
            }
            if let Some(offset) = params.offset {
                query_params.push(format!("offset={}", offset));
            }
            if let Some(metadata) = params.metadata {
                query_params.push(format!("metadata={}", urlencoding::encode(&metadata)));
            }
            if let Some(status) = params.status {
                query_params.push(format!("status={}", status));
            }
            if let Some(name) = params.name {
                query_params.push(format!("name={}", urlencoding::encode(&name)));
            }
            if let Some(tags) = params.tags {
                for tag in tags {
                    query_params.push(format!("tags={}", urlencoding::encode(&tag)));
                }
            }

            if !query_params.is_empty() {
                url.push_str(&format!("?{}", query_params.join("&")));
            }
        }

        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;
        Self::handle_response(response).await
    }

    /// Bulk create clients
    pub async fn bulk_create_clients(
        &self,
        domain_id: Uuid,
        clients: Vec<ClientRequest>,
    ) -> Result<ClientsPage, ClientsError> {
        let url = format!("{}/{}/clients/bulk", self.base_url, domain_id);
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&clients)
            .send()
            .await?;

        Self::handle_response(response).await
    }

    /// Get a specific client by ID
    pub async fn get_client(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
    ) -> Result<Client, ClientsError> {
        let url = format!("{}/{}/clients/{}", self.base_url, domain_id, client_id);
        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;
        Self::handle_response(response).await
    }

    /// Update a client
    pub async fn update_client(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        update: ClientUpdate,
    ) -> Result<Client, ClientsError> {
        let url = format!("{}/{}/clients/{}", self.base_url, domain_id, client_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &url)
            .json(&update)
            .send()
            .await?;

        Self::handle_response(response).await
    }

    /// Delete a client
    pub async fn delete_client(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
    ) -> Result<(), ClientsError> {
        let url = format!("{}/{}/clients/{}", self.base_url, domain_id, client_id);
        let response = self
            .build_request(reqwest::Method::DELETE, &url)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    /// Update client tags
    pub async fn update_client_tags(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        tags: ClientTags,
    ) -> Result<Client, ClientsError> {
        let url = format!("{}/{}/clients/{}/tags", self.base_url, domain_id, client_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &url)
            .json(&tags)
            .send()
            .await?;

        Self::handle_response(response).await
    }

    /// Update client secret
    pub async fn update_client_secret(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        secret: ClientSecret,
    ) -> Result<Client, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/secret",
            self.base_url, domain_id, client_id
        );
        let response = self
            .build_request(reqwest::Method::PATCH, &url)
            .json(&secret)
            .send()
            .await?;

        Self::handle_response(response).await
    }

    /// Enable a client
    pub async fn enable_client(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
    ) -> Result<Client, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/enable",
            self.base_url, domain_id, client_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .send()
            .await?;
        Self::handle_response(response).await
    }

    /// Disable a client
    pub async fn disable_client(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
    ) -> Result<Client, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/disable",
            self.base_url, domain_id, client_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .send()
            .await?;
        Self::handle_response(response).await
    }

    /// Set parent group for a client
    pub async fn set_client_parent_group(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        parent_group: ParentGroupRequest,
    ) -> Result<(), ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/parent",
            self.base_url, domain_id, client_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&parent_group)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    /// Remove parent group from a client
    pub async fn remove_client_parent_group(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        parent_group: ParentGroupRequest,
    ) -> Result<(), ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/parent",
            self.base_url, domain_id, client_id
        );
        let response = self
            .build_request(reqwest::Method::DELETE, &url)
            .json(&parent_group)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    // Role Management Operations

    /// Create a role for a client
    pub async fn create_client_role(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role: CreateRoleRequest,
    ) -> Result<Role, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles",
            self.base_url, domain_id, client_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&role)
            .send()
            .await?;

        Self::handle_response(response).await
    }

    /// List client roles
    pub async fn list_client_roles(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        limit: Option<u64>,
        offset: Option<u64>,
    ) -> Result<RolesPage, ClientsError> {
        let mut url = format!(
            "{}/{}/clients/{}/roles",
            self.base_url, domain_id, client_id
        );

        let mut query_params = Vec::new();
        if let Some(limit) = limit {
            query_params.push(format!("limit={}", limit));
        }
        if let Some(offset) = offset {
            query_params.push(format!("offset={}", offset));
        }

        if !query_params.is_empty() {
            url.push_str(&format!("?{}", query_params.join("&")));
        }

        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;
        Self::handle_response(response).await
    }

    /// Get client members from all roles
    pub async fn get_client_members(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
    ) -> Result<EntityMembers, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/members",
            self.base_url, domain_id, client_id
        );
        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;
        Self::handle_response(response).await
    }

    /// Get a specific client role
    pub async fn get_client_role(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
    ) -> Result<Role, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;
        Self::handle_response(response).await
    }

    /// Update a client role
    pub async fn update_client_role(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
        update: UpdateRoleRequest,
    ) -> Result<Role, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::PUT, &url)
            .json(&update)
            .send()
            .await?;

        Self::handle_response(response).await
    }

    /// Delete a client role
    pub async fn delete_client_role(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
    ) -> Result<(), ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::DELETE, &url)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    /// Add actions to a client role
    pub async fn add_client_role_actions(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
        actions: RoleActionsRequest,
    ) -> Result<Vec<String>, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/actions",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&actions)
            .send()
            .await?;

        #[derive(Deserialize)]
        struct ActionsResponse {
            actions: Vec<String>,
        }

        let result: ActionsResponse = Self::handle_response(response).await?;
        Ok(result.actions)
    }

    /// List client role actions
    pub async fn list_client_role_actions(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
    ) -> Result<Vec<String>, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/actions",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;

        #[derive(Deserialize)]
        struct ActionsResponse {
            actions: Vec<String>,
        }

        let result: ActionsResponse = Self::handle_response(response).await?;
        Ok(result.actions)
    }

    /// Delete specific actions from a client role
    pub async fn delete_client_role_actions(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
        actions: RoleActionsRequest,
    ) -> Result<(), ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/actions/delete",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&actions)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    /// Delete all actions from a client role
    pub async fn delete_all_client_role_actions(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
    ) -> Result<(), ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/actions/delete-all",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    /// Add members to a client role
    pub async fn add_client_role_members(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
        members: RoleMembersRequest,
    ) -> Result<Vec<String>, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/members",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&members)
            .send()
            .await?;

        #[derive(Deserialize)]
        struct MembersResponse {
            members: Vec<String>,
        }

        let result: MembersResponse = Self::handle_response(response).await?;
        Ok(result.members)
    }

    /// List client role members
    pub async fn list_client_role_members(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
    ) -> Result<Vec<String>, ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/members",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;

        #[derive(Deserialize)]
        struct MembersResponse {
            members: Vec<String>,
        }

        let result: MembersResponse = Self::handle_response(response).await?;
        Ok(result.members)
    }

    /// Delete specific members from a client role
    pub async fn delete_client_role_members(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
        members: RoleMembersRequest,
    ) -> Result<(), ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/members/delete",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .json(&members)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    /// Delete all members from a client role
    pub async fn delete_all_client_role_members(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
        role_id: &str,
    ) -> Result<(), ClientsError> {
        let url = format!(
            "{}/{}/clients/{}/roles/{}/members/delete-all",
            self.base_url, domain_id, client_id, role_id
        );
        let response = self
            .build_request(reqwest::Method::POST, &url)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            let error_text = response
                .text()
                .await
                .unwrap_or_else(|_| "Unknown error".to_string());
            Err(ClientsError::Api(error_text))
        }
    }

    /// List available actions
    pub async fn list_available_actions(
        &self,
        domain_id: Uuid,
    ) -> Result<Vec<String>, ClientsError> {
        let url = format!(
            "{}/{}/clients/roles/available-actions",
            self.base_url, domain_id
        );
        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;

        let result: AvailableActionsResponse = Self::handle_response(response).await?;
        Ok(result.actions)
    }

    /// Health check
    pub async fn health(&self) -> Result<HealthResponse, ClientsError> {
        let url = format!("{}/health", self.base_url);
        let response = self.http_client.get(&url).send().await?;
        Self::handle_response(response).await
    }
}

// Convenience methods and builders
impl ClientRequest {
    pub fn new(identity: String, secret: String) -> Self {
        Self {
            name: None,
            tags: Vec::new(),
            credentials: Credentials {
                identity,
                secret: Some(secret),
            },
            metadata: HashMap::new(),
            status: None,
        }
    }

    pub fn with_name(mut self, name: String) -> Self {
        self.name = Some(name);
        self
    }

    pub fn with_tags(mut self, tags: Vec<String>) -> Self {
        self.tags = tags;
        self
    }

    pub fn with_metadata(mut self, metadata: HashMap<String, serde_json::Value>) -> Self {
        self.metadata = metadata;
        self
    }

    pub fn with_status(mut self, status: String) -> Self {
        self.status = Some(status);
        self
    }
}

impl ClientListParams {
    pub fn new() -> Self {
        Default::default()
    }

    pub fn with_limit(mut self, limit: u64) -> Self {
        self.limit = Some(limit);
        self
    }

    pub fn with_offset(mut self, offset: u64) -> Self {
        self.offset = Some(offset);
        self
    }

    pub fn with_status(mut self, status: String) -> Self {
        self.status = Some(status);
        self
    }

    pub fn with_name(mut self, name: String) -> Self {
        self.name = Some(name);
        self
    }

    pub fn with_tags(mut self, tags: Vec<String>) -> Self {
        self.tags = Some(tags);
        self
    }

    pub fn with_metadata_filter(mut self, metadata: String) -> Self {
        self.metadata = Some(metadata);
        self
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_client_creation() {
        // This is a basic test structure - you would need to set up a test server
        // or mock the HTTP client for actual testing

        let client = ClientsServiceClient::new("http://localhost:9006".to_string())
            .with_auth_token("test-token".to_string());

        let client_request =
            ClientRequest::new("test@example.com".to_string(), "test-secret".to_string())
                .with_name("Test Client".to_string())
                .with_tags(vec!["test".to_string(), "example".to_string()]);

        // In a real test, you would make actual API calls or mock them
        assert_eq!(client_request.credentials.identity, "test@example.com");
        assert_eq!(client_request.name, Some("Test Client".to_string()));
    }

    #[test]
    fn test_client_list_params_builder() {
        let params = ClientListParams::new()
            .with_limit(50)
            .with_offset(10)
            .with_status("enabled".to_string())
            .with_name("test".to_string())
            .with_tags(vec!["tag1".to_string(), "tag2".to_string()]);

        assert_eq!(params.limit, Some(50));
        assert_eq!(params.offset, Some(10));
        assert_eq!(params.status, Some("enabled".to_string()));
        assert_eq!(params.name, Some("test".to_string()));
        assert_eq!(
            params.tags,
            Some(vec!["tag1".to_string(), "tag2".to_string()])
        );
    }
}
