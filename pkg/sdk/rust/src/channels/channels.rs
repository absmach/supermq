use reqwest::{Client, Error as ReqwestError, Response};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;
use url::Url;

/// Custom error type for the SuperMQ Channels SDK
#[derive(Error, Debug)]
pub enum ChannelsError {
    #[error("HTTP request failed: {0}")]
    RequestFailed(#[from] ReqwestError),

    #[error("Bad Request (400): {0}")]
    BadRequest(String),

    #[error("Unauthorized (401): Missing or invalid access token")]
    Unauthorized,

    #[error("Forbidden (403): Failed to perform authorization")]
    Forbidden,

    #[error("Not Found (404): {0}")]
    NotFound(String),

    #[error("Conflict (409): {0}")]
    Conflict(String),

    #[error("Unsupported Media Type (415): Missing or invalid content type")]
    UnsupportedMediaType,

    #[error("Unprocessable Entity (422): Database can't process request")]
    UnprocessableEntity,

    #[error("Internal Server Error (500): {0}")]
    InternalServerError(String),

    #[error("Unknown error: {status} - {message}")]
    Unknown { status: u16, message: String },
}

/// Represents a channel in the SuperMQ system
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Channel {
    pub id: String,
    pub name: String,
    pub domain_id: String,
    pub parent_id: Option<String>,
    pub route: Option<String>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    pub path: Option<String>,
    pub level: Option<i32>,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
    pub status: String,
}

/// Request object for creating a new channel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelCreateRequest {
    pub name: String,
    pub parent_id: Option<String>,
    pub route: Option<String>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    pub status: Option<String>,
}

/// Request object for updating a channel
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelUpdateRequest {
    pub name: String,
    pub metadata: HashMap<String, serde_json::Value>,
}

/// Request object for updating channel tags
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelUpdateTagsRequest {
    pub tags: Vec<String>,
}

/// Request object for setting/removing parent group
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ParentGroupRequest {
    pub parent_group_id: String,
}

/// Paginated response for channels
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelsPage {
    pub channels: Vec<Channel>,
    pub total: i64,
    pub offset: i64,
    pub limit: Option<i64>,
}

/// Connection request schema
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionRequest {
    pub channel_ids: Vec<String>,
    pub client_ids: Vec<String>,
    pub types: Option<Vec<String>>,
}

/// Channel connection request schema
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelConnectionRequest {
    pub client_ids: Vec<String>,
    pub types: Option<Vec<String>>,
}

/// Health check response
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: Option<String>,
    pub commit: Option<String>,
    pub description: Option<String>,
    pub build_time: Option<String>,
}

/// Query parameters for listing channels
#[derive(Debug, Default, Clone)]
pub struct ListChannelsParams {
    pub limit: Option<i32>,
    pub offset: Option<i32>,
    pub metadata: Option<String>,
    pub name: Option<String>,
}

/// Main client for the SuperMQ Channels Service
#[derive(Debug, Clone)]
pub struct ChannelsClient {
    client: Client,
    base_url: Url,
    bearer_token: Option<String>,
}

impl ChannelsClient {
    /// Create a new ChannelsClient instance
    pub fn new(base_url: &str) -> Result<Self, url::ParseError> {
        let base_url = Url::parse(base_url)?;
        Ok(Self {
            client: Client::new(),
            base_url,
            bearer_token: None,
        })
    }

    /// Set the bearer token for authentication
    pub fn with_token(mut self, token: String) -> Self {
        self.bearer_token = Some(token);
        self
    }

    /// Set the bearer token for authentication (mutable)
    pub fn set_token(&mut self, token: String) {
        self.bearer_token = Some(token);
    }

    /// Helper method to build request with authentication
    fn build_request(&self, method: reqwest::Method, url: Url) -> reqwest::RequestBuilder {
        let mut request = self.client.request(method, url);

        if let Some(token) = &self.bearer_token {
            request = request.bearer_auth(token);
        }

        request.header("Content-Type", "application/json")
    }

    /// Helper method to handle HTTP errors
    async fn handle_response_error(response: Response) -> ChannelsError {
        let status = response.status().as_u16();
        let error_message = response
            .text()
            .await
            .unwrap_or_else(|_| "Failed to read error response".to_string());

        match status {
            400 => ChannelsError::BadRequest(error_message),
            401 => ChannelsError::Unauthorized,
            403 => ChannelsError::Forbidden,
            404 => ChannelsError::NotFound(error_message),
            409 => ChannelsError::Conflict(error_message),
            415 => ChannelsError::UnsupportedMediaType,
            422 => ChannelsError::UnprocessableEntity,
            500 => ChannelsError::InternalServerError(error_message),
            _ => ChannelsError::Unknown {
                status,
                message: error_message,
            },
        }
    }

    /// Create a new channel
    pub async fn create_channel(
        &self,
        domain_id: &str,
        request: ChannelCreateRequest,
    ) -> Result<Channel, ChannelsError> {
        let url = self.base_url.join(&format!("/{}/channels", domain_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(response.json::<Channel>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Create multiple channels in bulk
    pub async fn create_channels(
        &self,
        domain_id: &str,
        requests: Vec<ChannelCreateRequest>,
    ) -> Result<Vec<Channel>, ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/bulk", domain_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .json(&requests)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(response.json::<Vec<Channel>>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// List channels with optional filtering
    pub async fn list_channels(
        &self,
        domain_id: &str,
        params: Option<ListChannelsParams>,
    ) -> Result<ChannelsPage, ChannelsError> {
        let mut url = self.base_url.join(&format!("/{}/channels", domain_id))?;

        if let Some(params) = params {
            let mut query_pairs = url.query_pairs_mut();

            if let Some(limit) = params.limit {
                query_pairs.append_pair("limit", &limit.to_string());
            }
            if let Some(offset) = params.offset {
                query_pairs.append_pair("offset", &offset.to_string());
            }
            if let Some(metadata) = params.metadata {
                query_pairs.append_pair("metadata", &metadata);
            }
            if let Some(name) = params.name {
                query_pairs.append_pair("name", &name);
            }
        }

        let response = self.build_request(reqwest::Method::GET, url).send().await?;

        if response.status().is_success() {
            Ok(response.json::<ChannelsPage>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Get a specific channel by ID
    pub async fn get_channel(
        &self,
        domain_id: &str,
        channel_id: &str,
    ) -> Result<Channel, ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}", domain_id, channel_id))?;
        let response = self.build_request(reqwest::Method::GET, url).send().await?;

        if response.status().is_success() {
            Ok(response.json::<Channel>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Update a channel
    pub async fn update_channel(
        &self,
        domain_id: &str,
        channel_id: &str,
        request: ChannelUpdateRequest,
    ) -> Result<Channel, ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::PATCH, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(response.json::<Channel>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Delete a channel
    pub async fn delete_channel(
        &self,
        domain_id: &str,
        channel_id: &str,
    ) -> Result<(), ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::DELETE, url)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Update channel tags
    pub async fn update_channel_tags(
        &self,
        domain_id: &str,
        channel_id: &str,
        request: ChannelUpdateTagsRequest,
    ) -> Result<Channel, ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}/tags", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::PATCH, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(response.json::<Channel>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Enable a channel
    pub async fn enable_channel(
        &self,
        domain_id: &str,
        channel_id: &str,
    ) -> Result<Channel, ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}/enable", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(response.json::<Channel>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Disable a channel
    pub async fn disable_channel(
        &self,
        domain_id: &str,
        channel_id: &str,
    ) -> Result<Channel, ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}/disable", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(response.json::<Channel>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Set a parent group for a channel
    pub async fn set_channel_parent_group(
        &self,
        domain_id: &str,
        channel_id: &str,
        request: ParentGroupRequest,
    ) -> Result<(), ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}/parent", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Remove a parent group from a channel
    pub async fn remove_channel_parent_group(
        &self,
        domain_id: &str,
        channel_id: &str,
        request: ParentGroupRequest,
    ) -> Result<(), ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}/parent", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::DELETE, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Connect clients and channels
    pub async fn connect_clients_and_channels(
        &self,
        domain_id: &str,
        request: ConnectionRequest,
    ) -> Result<(), ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/connect", domain_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Disconnect clients and channels
    pub async fn disconnect_clients_and_channels(
        &self,
        domain_id: &str,
        request: ConnectionRequest,
    ) -> Result<(), ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/disconnect", domain_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Connect clients to a specific channel
    pub async fn connect_clients_to_channel(
        &self,
        domain_id: &str,
        channel_id: &str,
        request: ChannelConnectionRequest,
    ) -> Result<(), ChannelsError> {
        let url = self
            .base_url
            .join(&format!("/{}/channels/{}/connect", domain_id, channel_id))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Disconnect clients from a specific channel
    pub async fn disconnect_clients_from_channel(
        &self,
        domain_id: &str,
        channel_id: &str,
        request: ChannelConnectionRequest,
    ) -> Result<(), ChannelsError> {
        let url = self.base_url.join(&format!(
            "/{}/channels/{}/disconnect",
            domain_id, channel_id
        ))?;
        let response = self
            .build_request(reqwest::Method::POST, url)
            .json(&request)
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }

    /// Check service health
    pub async fn health_check(&self) -> Result<HealthResponse, ChannelsError> {
        let url = self.base_url.join("/health")?;
        let response = self.client.get(url).send().await?;

        if response.status().is_success() {
            Ok(response.json::<HealthResponse>().await?)
        } else {
            Err(Self::handle_response_error(response).await)
        }
    }
}

/// Convenience methods for creating request objects
impl ChannelCreateRequest {
    pub fn new(name: String) -> Self {
        Self {
            name,
            parent_id: None,
            route: None,
            metadata: None,
            status: None,
        }
    }

    pub fn with_parent_id(mut self, parent_id: String) -> Self {
        self.parent_id = Some(parent_id);
        self
    }

    pub fn with_route(mut self, route: String) -> Self {
        self.route = Some(route);
        self
    }

    pub fn with_metadata(mut self, metadata: HashMap<String, serde_json::Value>) -> Self {
        self.metadata = Some(metadata);
        self
    }

    pub fn with_status(mut self, status: String) -> Self {
        self.status = Some(status);
        self
    }
}

impl ChannelUpdateRequest {
    pub fn new(name: String, metadata: HashMap<String, serde_json::Value>) -> Self {
        Self { name, metadata }
    }
}

impl ChannelUpdateTagsRequest {
    pub fn new(tags: Vec<String>) -> Self {
        Self { tags }
    }
}

impl ParentGroupRequest {
    pub fn new(parent_group_id: String) -> Self {
        Self { parent_group_id }
    }
}

impl ConnectionRequest {
    pub fn new(channel_ids: Vec<String>, client_ids: Vec<String>) -> Self {
        Self {
            channel_ids,
            client_ids,
            types: None,
        }
    }

    pub fn with_types(mut self, types: Vec<String>) -> Self {
        self.types = Some(types);
        self
    }
}

impl ChannelConnectionRequest {
    pub fn new(client_ids: Vec<String>) -> Self {
        Self {
            client_ids,
            types: None,
        }
    }

    pub fn with_types(mut self, types: Vec<String>) -> Self {
        self.types = Some(types);
        self
    }
}

impl ListChannelsParams {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn with_limit(mut self, limit: i32) -> Self {
        self.limit = Some(limit);
        self
    }

    pub fn with_offset(mut self, offset: i32) -> Self {
        self.offset = Some(offset);
        self
    }

    pub fn with_metadata(mut self, metadata: String) -> Self {
        self.metadata = Some(metadata);
        self
    }

    pub fn with_name(mut self, name: String) -> Self {
        self.name = Some(name);
        self
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    #[test]
    fn test_create_client() {
        let client = ChannelsClient::new("http://localhost:9005").unwrap();
        assert!(!client.base_url.as_str().is_empty());
    }

    #[test]
    fn test_channel_create_request_builder() {
        let mut metadata = HashMap::new();
        metadata.insert(
            "location".to_string(),
            serde_json::Value::String("test".to_string()),
        );

        let request = ChannelCreateRequest::new("test-channel".to_string())
            .with_parent_id("parent-123".to_string())
            .with_route("test-route".to_string())
            .with_metadata(metadata)
            .with_status("enabled".to_string());

        assert_eq!(request.name, "test-channel");
        assert_eq!(request.parent_id, Some("parent-123".to_string()));
        assert_eq!(request.route, Some("test-route".to_string()));
        assert_eq!(request.status, Some("enabled".to_string()));
        assert!(request.metadata.is_some());
    }

    #[test]
    fn test_list_channels_params_builder() {
        let params = ListChannelsParams::new()
            .with_limit(50)
            .with_offset(10)
            .with_name("test".to_string())
            .with_metadata(r#"{"location": "test"}"#.to_string());

        assert_eq!(params.limit, Some(50));
        assert_eq!(params.offset, Some(10));
        assert_eq!(params.name, Some("test".to_string()));
        assert!(params.metadata.is_some());
    }
}
