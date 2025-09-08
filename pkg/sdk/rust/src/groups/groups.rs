// SuperMQ Groups Service Rust SDK
// Generated from OpenAPI specification v0.18.0

use chrono::{DateTime, Utc};
use reqwest::{Client, Response, StatusCode};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;
use uuid::Uuid;

/// SDK Error Types
#[derive(Error, Debug)]
pub enum GroupsError {
    #[error("HTTP request failed: {0}")]
    Http(#[from] reqwest::Error),

    #[error("Serialization error: {0}")]
    Serialization(#[from] serde_json::Error),

    #[error("Bad Request: {message}")]
    BadRequest { message: String },

    #[error("Unauthorized: Missing or invalid access token")]
    Unauthorized,

    #[error("Forbidden: Failed to perform authorization over the entity")]
    Forbidden,

    #[error("Not Found: {message}")]
    NotFound { message: String },

    #[error("Conflict: {message}")]
    Conflict { message: String },

    #[error("Unprocessable Entity: Database can't process request")]
    UnprocessableEntity,

    #[error("Internal Server Error: {message}")]
    InternalServerError { message: String },

    #[error("Unsupported Media Type: Missing or invalid content type")]
    UnsupportedMediaType,
}

/// API Error Response
#[derive(Debug, Deserialize)]
pub struct ApiError {
    pub error: String,
}

/// Group Status
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum GroupStatus {
    Enabled,
    Disabled,
}

impl Default for GroupStatus {
    fn default() -> Self {
        Self::Enabled
    }
}

/// Group Creation Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateGroupRequest {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub parent_id: Option<Uuid>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub status: Option<GroupStatus>,
}

/// Group Update Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateGroupRequest {
    pub name: String,
    pub description: String,
    pub metadata: HashMap<String, serde_json::Value>,
}

/// Group Tags Update Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateGroupTagsRequest {
    pub tags: Vec<String>,
}

/// Parent Group Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParentGroupRequest {
    pub group_id: Uuid,
}

/// Children Groups Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChildrenGroupsRequest {
    pub groups: Vec<Uuid>,
}

/// Group Model
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Group {
    pub id: Uuid,
    pub name: String,
    pub domain_id: Uuid,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub parent_id: Option<Uuid>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    #[serde(default)]
    pub tags: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub path: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub level: Option<i32>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub status: GroupStatus,
}

/// Member Model
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Member {
    pub id: Uuid,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub first_name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub last_name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub email: Option<String>,
    #[serde(default)]
    pub tags: Vec<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub credentials: Option<HashMap<String, String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    pub status: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// Groups Page Response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GroupsPage {
    pub groups: Vec<Group>,
    pub total: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub offset: Option<u64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub limit: Option<u64>,
}

/// Groups Hierarchy Page Response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GroupsHierarchyPage {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub level: Option<i32>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub direction: Option<i32>,
    pub groups: Vec<Group>,
}

/// Members Page Response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MembersPage {
    pub members: Vec<Member>,
    pub total: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub offset: Option<u64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub limit: Option<u64>,
}

/// Role Model
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Role {
    pub id: String,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    #[serde(default)]
    pub actions: Vec<String>,
    #[serde(default)]
    pub members: Vec<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

/// Create Role Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRoleRequest {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    #[serde(default)]
    pub actions: Vec<String>,
}

/// Update Role Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateRoleRequest {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
}

/// Add Role Actions Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddRoleActionsRequest {
    pub actions: Vec<String>,
}

/// Add Role Members Request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddRoleMembersRequest {
    pub members: Vec<String>,
}

/// Roles Page Response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RolesPage {
    pub roles: Vec<Role>,
    pub total: u64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub offset: Option<u64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub limit: Option<u64>,
}

/// Available Actions Response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AvailableActionsResponse {
    pub actions: Vec<String>,
}

/// Health Check Response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: String,
    pub commit: String,
    pub description: String,
    pub build_time: String,
}

/// Query Parameters for listing groups
#[derive(Debug, Clone, Default)]
pub struct ListGroupsParams {
    pub limit: Option<u32>,
    pub offset: Option<u32>,
    pub level: Option<u32>,
    pub tree: Option<bool>,
    pub metadata: Option<String>,
    pub name: Option<String>,
    pub root_group: Option<bool>,
}

/// Query Parameters for listing children groups
#[derive(Debug, Clone, Default)]
pub struct ListChildrenGroupsParams {
    pub limit: Option<u32>,
    pub offset: Option<u32>,
    pub start_level: Option<u32>,
    pub end_level: Option<u32>,
    pub tree: Option<bool>,
    pub metadata: Option<String>,
    pub name: Option<String>,
}

/// Query Parameters for group hierarchy
#[derive(Debug, Clone, Default)]
pub struct GroupHierarchyParams {
    pub level: Option<u32>,
    pub tree: Option<bool>,
    pub direction: Option<i32>,
}

/// Query Parameters for pagination
#[derive(Debug, Clone, Default)]
pub struct PaginationParams {
    pub limit: Option<u32>,
    pub offset: Option<u32>,
}

/// SuperMQ Groups Client Configuration
#[derive(Debug, Clone)]
pub struct ClientConfig {
    pub base_url: String,
    pub token: String,
    pub timeout: Option<std::time::Duration>,
}

impl Default for ClientConfig {
    fn default() -> Self {
        Self {
            base_url: "http://localhost:9004".to_string(),
            token: String::new(),
            timeout: Some(std::time::Duration::from_secs(30)),
        }
    }
}

/// SuperMQ Groups Service Client
#[derive(Debug)]
pub struct GroupsClient {
    client: Client,
    config: ClientConfig,
}

impl GroupsClient {
    /// Create a new Groups Client
    pub fn new(config: ClientConfig) -> Result<Self, GroupsError> {
        let mut client_builder = Client::builder();

        if let Some(timeout) = config.timeout {
            client_builder = client_builder.timeout(timeout);
        }

        let client = client_builder.build()?;

        Ok(Self { client, config })
    }

    /// Create a client with default configuration
    pub fn with_token(token: impl Into<String>) -> Result<Self, GroupsError> {
        let mut config = ClientConfig::default();
        config.token = token.into();
        Self::new(config)
    }

    /// Handle HTTP response and convert to appropriate error
    async fn handle_response<T>(&self, response: Response) -> Result<T, GroupsError>
    where
        T: for<'de> Deserialize<'de>,
    {
        let status = response.status();
        let body = response.text().await?;

        match status {
            StatusCode::OK | StatusCode::CREATED => {
                serde_json::from_str(&body).map_err(GroupsError::from)
            }
            StatusCode::NO_CONTENT => {
                // For 204 responses, return empty JSON object
                serde_json::from_str("{}").map_err(GroupsError::from)
            }
            StatusCode::BAD_REQUEST => {
                let error: ApiError = serde_json::from_str(&body).unwrap_or_else(|_| ApiError {
                    error: "Bad Request".to_string(),
                });
                Err(GroupsError::BadRequest {
                    message: error.error,
                })
            }
            StatusCode::UNAUTHORIZED => Err(GroupsError::Unauthorized),
            StatusCode::FORBIDDEN => Err(GroupsError::Forbidden),
            StatusCode::NOT_FOUND => {
                let error: ApiError = serde_json::from_str(&body).unwrap_or_else(|_| ApiError {
                    error: "Not Found".to_string(),
                });
                Err(GroupsError::NotFound {
                    message: error.error,
                })
            }
            StatusCode::CONFLICT => {
                let error: ApiError = serde_json::from_str(&body).unwrap_or_else(|_| ApiError {
                    error: "Conflict".to_string(),
                });
                Err(GroupsError::Conflict {
                    message: error.error,
                })
            }
            StatusCode::UNSUPPORTED_MEDIA_TYPE => Err(GroupsError::UnsupportedMediaType),
            StatusCode::UNPROCESSABLE_ENTITY => Err(GroupsError::UnprocessableEntity),
            StatusCode::INTERNAL_SERVER_ERROR => {
                let error: ApiError = serde_json::from_str(&body).unwrap_or_else(|_| ApiError {
                    error: "Internal Server Error".to_string(),
                });
                Err(GroupsError::InternalServerError {
                    message: error.error,
                })
            }
            _ => {
                let error: ApiError = serde_json::from_str(&body).unwrap_or_else(|_| ApiError {
                    error: format!("HTTP {}", status),
                });
                Err(GroupsError::InternalServerError {
                    message: error.error,
                })
            }
        }
    }

    /// Build URL with domain ID
    fn build_url(&self, domain_id: &Uuid, path: &str) -> String {
        format!("{}/{}{}", self.config.base_url, domain_id, path)
    }

    /// Add query parameters to request
    fn add_query_params<T: Serialize>(
        &self,
        mut builder: reqwest::RequestBuilder,
        params: &T,
    ) -> reqwest::RequestBuilder {
        // Convert to serde_json::Value and extract as query params
        if let Ok(value) = serde_json::to_value(params) {
            if let Some(obj) = value.as_object() {
                for (key, val) in obj {
                    if !val.is_null() {
                        let param_value = match val {
                            serde_json::Value::String(s) => s.clone(),
                            serde_json::Value::Number(n) => n.to_string(),
                            serde_json::Value::Bool(b) => b.to_string(),
                            _ => continue,
                        };
                        builder = builder.query(&[(key, param_value)]);
                    }
                }
            }
        }
        builder
    }

    /// Create a new group
    pub async fn create_group(
        &self,
        domain_id: &Uuid,
        request: CreateGroupRequest,
    ) -> Result<Group, GroupsError> {
        let url = self.build_url(domain_id, "/groups");

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// List groups
    pub async fn list_groups(
        &self,
        domain_id: &Uuid,
        params: Option<ListGroupsParams>,
    ) -> Result<GroupsPage, GroupsError> {
        let url = self.build_url(domain_id, "/groups");

        let mut builder = self.client.get(&url).bearer_auth(&self.config.token);

        if let Some(params) = params {
            builder = self.add_query_params(builder, &params);
        }

        let response = builder.send().await?;
        self.handle_response(response).await
    }

    /// Get group by ID
    pub async fn get_group(&self, domain_id: &Uuid, group_id: &Uuid) -> Result<Group, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}", group_id));

        let response = self
            .client
            .get(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Update group
    pub async fn update_group(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        request: UpdateGroupRequest,
    ) -> Result<Group, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}", group_id));

        let response = self
            .client
            .put(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Delete group
    pub async fn delete_group(&self, domain_id: &Uuid, group_id: &Uuid) -> Result<(), GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}", group_id));

        let response = self
            .client
            .delete(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Update group tags
    pub async fn update_group_tags(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        request: UpdateGroupTagsRequest,
    ) -> Result<Group, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/tags", group_id));

        let response = self
            .client
            .patch(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Enable group
    pub async fn enable_group(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
    ) -> Result<Group, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/enable", group_id));

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Disable group
    pub async fn disable_group(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
    ) -> Result<Group, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/disable", group_id));

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// List group hierarchy
    pub async fn list_group_hierarchy(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        params: Option<GroupHierarchyParams>,
    ) -> Result<GroupsHierarchyPage, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/hierarchy", group_id));

        let mut builder = self.client.get(&url).bearer_auth(&self.config.token);

        if let Some(params) = params {
            builder = self.add_query_params(builder, &params);
        }

        let response = builder.send().await?;
        self.handle_response(response).await
    }

    /// Set parent group
    pub async fn set_parent_group(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        request: ParentGroupRequest,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/parent", group_id));

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Remove parent group
    pub async fn remove_parent_group(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        request: ParentGroupRequest,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/parent", group_id));

        let response = self
            .client
            .delete(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Add children groups
    pub async fn add_children_groups(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        request: ChildrenGroupsRequest,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/children", group_id));

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Remove children groups
    pub async fn remove_children_groups(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        request: ChildrenGroupsRequest,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/children", group_id));

        let response = self
            .client
            .delete(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// List children groups
    pub async fn list_children_groups(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        params: Option<ListChildrenGroupsParams>,
    ) -> Result<GroupsPage, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/children", group_id));

        let mut builder = self.client.get(&url).bearer_auth(&self.config.token);

        if let Some(params) = params {
            builder = self.add_query_params(builder, &params);
        }

        let response = builder.send().await?;
        self.handle_response(response).await
    }

    /// Remove all children groups
    pub async fn remove_all_children_groups(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/children/all", group_id));

        let response = self
            .client
            .delete(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Create group role
    pub async fn create_group_role(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        request: CreateRoleRequest,
    ) -> Result<Role, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/roles", group_id));

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// List group roles
    pub async fn list_group_roles(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        params: Option<PaginationParams>,
    ) -> Result<RolesPage, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/roles", group_id));

        let mut builder = self.client.get(&url).bearer_auth(&self.config.token);

        if let Some(params) = params {
            builder = self.add_query_params(builder, &params);
        }

        let response = builder.send().await?;
        self.handle_response(response).await
    }

    /// Get group members from all roles
    pub async fn get_group_members(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
    ) -> Result<MembersPage, GroupsError> {
        let url = self.build_url(domain_id, &format!("/groups/{}/roles/members", group_id));

        let response = self
            .client
            .get(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Get group role
    pub async fn get_group_role(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
    ) -> Result<Role, GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}", group_id, role_id),
        );

        let response = self
            .client
            .get(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Update group role
    pub async fn update_group_role(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
        request: UpdateRoleRequest,
    ) -> Result<Role, GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}", group_id, role_id),
        );

        let response = self
            .client
            .put(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Delete group role
    pub async fn delete_group_role(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}", group_id, role_id),
        );

        let response = self
            .client
            .delete(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Add group role actions
    pub async fn add_group_role_actions(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
        request: AddRoleActionsRequest,
    ) -> Result<Vec<String>, GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/actions", group_id, role_id),
        );

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let result: serde_json::Value = self.handle_response(response).await?;
        Ok(result
            .get("actions")
            .and_then(|v| v.as_array())
            .map(|arr| {
                arr.iter()
                    .filter_map(|v| v.as_str().map(String::from))
                    .collect()
            })
            .unwrap_or_default())
    }

    /// List group role actions
    pub async fn list_group_role_actions(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
    ) -> Result<Vec<String>, GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/actions", group_id, role_id),
        );

        let response = self
            .client
            .get(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        let result: serde_json::Value = self.handle_response(response).await?;
        Ok(result
            .get("actions")
            .and_then(|v| v.as_array())
            .map(|arr| {
                arr.iter()
                    .filter_map(|v| v.as_str().map(String::from))
                    .collect()
            })
            .unwrap_or_default())
    }

    /// Delete group role actions
    pub async fn delete_group_role_actions(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
        request: AddRoleActionsRequest,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/actions/delete", group_id, role_id),
        );

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Delete all group role actions
    pub async fn delete_all_group_role_actions(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/actions/delete-all", group_id, role_id),
        );

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Add group role members
    pub async fn add_group_role_members(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
        request: AddRoleMembersRequest,
    ) -> Result<Vec<String>, GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/members", group_id, role_id),
        );

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let result: serde_json::Value = self.handle_response(response).await?;
        Ok(result
            .get("members")
            .and_then(|v| v.as_array())
            .map(|arr| {
                arr.iter()
                    .filter_map(|v| v.as_str().map(String::from))
                    .collect()
            })
            .unwrap_or_default())
    }

    /// List group role members
    pub async fn list_group_role_members(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
    ) -> Result<MembersPage, GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/members", group_id, role_id),
        );

        let response = self
            .client
            .get(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Delete group role members
    pub async fn delete_group_role_members(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
        request: AddRoleMembersRequest,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/members/delete", group_id, role_id),
        );

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .json(&request)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// Delete all group role members
    pub async fn delete_all_group_role_members(
        &self,
        domain_id: &Uuid,
        group_id: &Uuid,
        role_id: &str,
    ) -> Result<(), GroupsError> {
        let url = self.build_url(
            domain_id,
            &format!("/groups/{}/roles/{}/members/delete-all", group_id, role_id),
        );

        let response = self
            .client
            .post(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        let _: serde_json::Value = self.handle_response(response).await?;
        Ok(())
    }

    /// List available actions
    pub async fn list_available_actions(
        &self,
        domain_id: &Uuid,
    ) -> Result<AvailableActionsResponse, GroupsError> {
        let url = self.build_url(domain_id, "/groups/roles/available-actions");

        let response = self
            .client
            .get(&url)
            .bearer_auth(&self.config.token)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Health check
    pub async fn health(&self) -> Result<HealthResponse, GroupsError> {
        let url = format!("{}/health", self.config.base_url);

        let response = self.client.get(&url).send().await?;

        self.handle_response(response).await
    }
}

// Convenience builder methods
impl CreateGroupRequest {
    /// Create a new group request builder
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            description: None,
            parent_id: None,
            metadata: None,
            status: None,
        }
    }

    /// Set description
    pub fn description(mut self, description: impl Into<String>) -> Self {
        self.description = Some(description.into());
        self
    }

    /// Set parent ID
    pub fn parent_id(mut self, parent_id: Uuid) -> Self {
        self.parent_id = Some(parent_id);
        self
    }

    /// Set metadata
    pub fn metadata(mut self, metadata: HashMap<String, serde_json::Value>) -> Self {
        self.metadata = Some(metadata);
        self
    }

    /// Set status
    pub fn status(mut self, status: GroupStatus) -> Self {
        self.status = Some(status);
        self
    }
}

impl UpdateGroupRequest {
    /// Create a new update group request
    pub fn new(
        name: impl Into<String>,
        description: impl Into<String>,
        metadata: HashMap<String, serde_json::Value>,
    ) -> Self {
        Self {
            name: name.into(),
            description: description.into(),
            metadata,
        }
    }
}

impl ListGroupsParams {
    /// Create new list groups parameters
    pub fn new() -> Self {
        Self::default()
    }

    /// Set limit
    pub fn limit(mut self, limit: u32) -> Self {
        self.limit = Some(limit);
        self
    }

    /// Set offset
    pub fn offset(mut self, offset: u32) -> Self {
        self.offset = Some(offset);
        self
    }

    /// Set level
    pub fn level(mut self, level: u32) -> Self {
        self.level = Some(level);
        self
    }

    /// Set tree format
    pub fn tree(mut self, tree: bool) -> Self {
        self.tree = Some(tree);
        self
    }

    /// Set metadata filter
    pub fn metadata(mut self, metadata: impl Into<String>) -> Self {
        self.metadata = Some(metadata.into());
        self
    }

    /// Set name filter
    pub fn name(mut self, name: impl Into<String>) -> Self {
        self.name = Some(name.into());
        self
    }

    /// Set root group filter
    pub fn root_group(mut self, root_group: bool) -> Self {
        self.root_group = Some(root_group);
        self
    }
}

// Tests module
#[cfg(test)]
mod tests {
    use super::*;
    use uuid::Uuid;

    #[test]
    fn test_create_group_request_builder() {
        let request = CreateGroupRequest::new("test-group")
            .description("A test group")
            .status(GroupStatus::Enabled);

        assert_eq!(request.name, "test-group");
        assert_eq!(request.description, Some("A test group".to_string()));
        assert_eq!(request.status, Some(GroupStatus::Enabled));
    }

    #[test]
    fn test_list_groups_params_builder() {
        let params = ListGroupsParams::new()
            .limit(50)
            .offset(10)
            .tree(true)
            .name("test");

        assert_eq!(params.limit, Some(50));
        assert_eq!(params.offset, Some(10));
        assert_eq!(params.tree, Some(true));
        assert_eq!(params.name, Some("test".to_string()));
    }

    #[test]
    fn test_client_config_default() {
        let config = ClientConfig::default();
        assert_eq!(config.base_url, "http://localhost:9004");
        assert!(config.token.is_empty());
        assert!(config.timeout.is_some());
    }

    #[tokio::test]
    async fn test_client_creation() {
        let config = ClientConfig {
            base_url: "http://localhost:9004".to_string(),
            token: "test-token".to_string(),
            timeout: Some(std::time::Duration::from_secs(10)),
        };

        let client = GroupsClient::new(config);
        assert!(client.is_ok());
    }

    #[tokio::test]
    async fn test_client_with_token() {
        let client = GroupsClient::with_token("test-token");
        assert!(client.is_ok());
    }

    #[test]
    fn test_group_status_serialization() {
        assert_eq!(
            serde_json::to_string(&GroupStatus::Enabled).unwrap(),
            "\"enabled\""
        );
        assert_eq!(
            serde_json::to_string(&GroupStatus::Disabled).unwrap(),
            "\"disabled\""
        );
    }

    #[test]
    fn test_group_status_default() {
        assert_eq!(GroupStatus::default(), GroupStatus::Enabled);
    }
}
