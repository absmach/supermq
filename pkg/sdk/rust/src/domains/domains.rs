use reqwest::{Client, Response};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

/// SuperMQ Domains Service Client
#[derive(Clone)]
pub struct DomainsClient {
    base_url: String,
    client: Client,
    token: Option<String>,
}

/// Domain representation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Domain {
    pub id: Uuid,
    pub name: String,
    pub tags: Vec<String>,
    pub metadata: HashMap<String, serde_json::Value>,
    pub route: String,
    pub status: String,
    pub created_by: Uuid,
    pub created_at: String,
    pub updated_by: Option<Uuid>,
    pub updated_at: Option<String>,
}

/// Request to create a domain
#[derive(Debug, Serialize)]
pub struct CreateDomainRequest {
    pub name: String,
    pub route: String,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    pub tags: Vec<String>,
    #[serde(skip_serializing_if = "HashMap::is_empty")]
    pub metadata: HashMap<String, serde_json::Value>,
}

/// Request to update a domain
#[derive(Debug, Serialize)]
pub struct UpdateDomainRequest {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

/// Paginated domains response
#[derive(Debug, Deserialize)]
pub struct DomainsPage {
    pub domains: Vec<Domain>,
    pub total: u64,
    pub offset: u64,
    pub limit: u64,
}

/// Invitation representation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Invitation {
    pub invited_by: Uuid,
    pub invitee_user_id: Uuid,
    pub domain_id: Uuid,
    pub role_id: Uuid,
    pub role_name: String,
    pub actions: Vec<String>,
    pub created_at: String,
    pub updated_at: Option<String>,
    pub confirmed_at: Option<String>,
}

/// Request to send an invitation
#[derive(Debug, Serialize)]
pub struct SendInvitationRequest {
    pub invitee_user_id: Uuid,
    pub role_id: Uuid,
}

/// Paginated invitations response
#[derive(Debug, Deserialize)]
pub struct InvitationsPage {
    pub invitations: Vec<Invitation>,
    pub total: u64,
    pub offset: u64,
    pub limit: u64,
}

/// Query parameters for listing domains
#[derive(Debug, Default)]
pub struct ListDomainsParams {
    pub limit: Option<u64>,
    pub offset: Option<u64>,
    pub name: Option<String>,
    pub status: Option<String>,
    pub permission: Option<String>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

/// Query parameters for listing invitations
#[derive(Debug, Default)]
pub struct ListInvitationsParams {
    pub limit: Option<u64>,
    pub offset: Option<u64>,
    pub user_id: Option<Uuid>,
    pub invited_by: Option<Uuid>,
    pub state: Option<String>,
}

/// Service health information
#[derive(Debug, Deserialize)]
pub struct HealthInfo {
    pub status: String,
    pub version: String,
    pub commit: String,
    pub description: String,
    pub build_time: String,
}

/// API Error response
#[derive(Debug, Deserialize)]
pub struct ApiError {
    pub error: String,
}

/// Result type for API operations
pub type Result<T> = std::result::Result<T, Box<dyn std::error::Error + Send + Sync>>;

impl DomainsClient {
    /// Create a new domains client
    pub fn new<S: Into<String>>(base_url: S) -> Self {
        Self {
            base_url: base_url.into(),
            client: Client::new(),
            token: None,
        }
    }

    /// Set authentication token
    pub fn with_token<S: Into<String>>(mut self, token: S) -> Self {
        self.token = Some(token.into());
        self
    }

    /// Set authentication token after construction
    pub fn set_token<S: Into<String>>(&mut self, token: S) {
        self.token = Some(token.into());
    }

    /// Build request with authentication
    fn request(&self, method: reqwest::Method, path: &str) -> reqwest::RequestBuilder {
        let url = format!("{}{}", self.base_url, path);
        let mut req = self.client.request(method, &url);

        if let Some(ref token) = self.token {
            req = req.bearer_auth(token);
        }

        req
    }

    /// Handle API response and extract JSON
    async fn handle_response<T: for<'de> Deserialize<'de>>(&self, response: Response) -> Result<T> {
        if response.status().is_success() {
            Ok(response.json::<T>().await?)
        } else {
            let error = response
                .json::<ApiError>()
                .await
                .unwrap_or_else(|_| ApiError {
                    error: "Unknown error".to_string(),
                });
            Err(format!("API error: {}", error.error).into())
        }
    }

    // === DOMAIN OPERATIONS ===

    /// Create a new domain
    pub async fn create_domain(&self, req: CreateDomainRequest) -> Result<Domain> {
        let response = self
            .request(reqwest::Method::POST, "/domains")
            .json(&req)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// List domains with optional filtering
    pub async fn list_domains(&self, params: ListDomainsParams) -> Result<DomainsPage> {
        let mut req = self.request(reqwest::Method::GET, "/domains");

        if let Some(limit) = params.limit {
            req = req.query(&[("limit", limit.to_string())]);
        }
        if let Some(offset) = params.offset {
            req = req.query(&[("offset", offset.to_string())]);
        }
        if let Some(name) = params.name {
            req = req.query(&[("name", name)]);
        }
        if let Some(status) = params.status {
            req = req.query(&[("status", status)]);
        }
        if let Some(permission) = params.permission {
            req = req.query(&[("permission", permission)]);
        }
        if let Some(metadata) = params.metadata {
            let metadata_str = serde_json::to_string(&metadata)?;
            req = req.query(&[("metadata", metadata_str)]);
        }

        let response = req.send().await?;
        self.handle_response(response).await
    }

    /// Get a specific domain by ID
    pub async fn get_domain(&self, domain_id: Uuid) -> Result<Domain> {
        let response = self
            .request(reqwest::Method::GET, &format!("/domains/{}", domain_id))
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Update a domain
    pub async fn update_domain(&self, domain_id: Uuid, req: UpdateDomainRequest) -> Result<Domain> {
        let response = self
            .request(reqwest::Method::PATCH, &format!("/domains/{}", domain_id))
            .json(&req)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Enable a domain
    pub async fn enable_domain(&self, domain_id: Uuid) -> Result<()> {
        let response = self
            .request(
                reqwest::Method::POST,
                &format!("/domains/{}/enable", domain_id),
            )
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(format!("Failed to enable domain: {}", response.status()).into())
        }
    }

    /// Disable a domain
    pub async fn disable_domain(&self, domain_id: Uuid) -> Result<()> {
        let response = self
            .request(
                reqwest::Method::POST,
                &format!("/domains/{}/disable", domain_id),
            )
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(format!("Failed to disable domain: {}", response.status()).into())
        }
    }

    /// Freeze a domain
    pub async fn freeze_domain(&self, domain_id: Uuid) -> Result<()> {
        let response = self
            .request(
                reqwest::Method::POST,
                &format!("/domains/{}/freeze", domain_id),
            )
            .send()
            .await?;

        if response.status().is_success() {
            Ok(())
        } else {
            Err(format!("Failed to freeze domain: {}", response.status()).into())
        }
    }

    // === INVITATION OPERATIONS ===

    /// Send an invitation to join a domain
    pub async fn send_invitation(&self, domain_id: Uuid, req: SendInvitationRequest) -> Result<()> {
        let response = self
            .request(
                reqwest::Method::POST,
                &format!("/domains/{}/invitations", domain_id),
            )
            .json(&req)
            .send()
            .await?;

        if response.status().as_u16() == 201 {
            Ok(())
        } else {
            Err(format!("Failed to send invitation: {}", response.status()).into())
        }
    }

    /// List invitations for a specific domain
    pub async fn list_domain_invitations(
        &self,
        domain_id: Uuid,
        params: ListInvitationsParams,
    ) -> Result<InvitationsPage> {
        let mut req = self.request(
            reqwest::Method::GET,
            &format!("/domains/{}/invitations", domain_id),
        );

        if let Some(limit) = params.limit {
            req = req.query(&[("limit", limit.to_string())]);
        }
        if let Some(offset) = params.offset {
            req = req.query(&[("offset", offset.to_string())]);
        }
        if let Some(user_id) = params.user_id {
            req = req.query(&[("user_id", user_id.to_string())]);
        }
        if let Some(invited_by) = params.invited_by {
            req = req.query(&[("invited_by", invited_by.to_string())]);
        }
        if let Some(state) = params.state {
            req = req.query(&[("state", state)]);
        }

        let response = req.send().await?;
        self.handle_response(response).await
    }

    /// List invitations for the current user
    pub async fn list_user_invitations(
        &self,
        params: ListInvitationsParams,
    ) -> Result<InvitationsPage> {
        let mut req = self.request(reqwest::Method::GET, "/invitations");

        if let Some(limit) = params.limit {
            req = req.query(&[("limit", limit.to_string())]);
        }
        if let Some(offset) = params.offset {
            req = req.query(&[("offset", offset.to_string())]);
        }
        if let Some(user_id) = params.user_id {
            req = req.query(&[("user_id", user_id.to_string())]);
        }
        if let Some(invited_by) = params.invited_by {
            req = req.query(&[("invited_by", invited_by.to_string())]);
        }
        if let Some(state) = params.state {
            req = req.query(&[("state", state)]);
        }

        let response = req.send().await?;
        self.handle_response(response).await
    }

    /// Accept an invitation
    pub async fn accept_invitation(&self, domain_id: Uuid) -> Result<()> {
        let req_body = serde_json::json!({"domain_id": domain_id});

        let response = self
            .request(reqwest::Method::POST, "/invitations/accept")
            .json(&req_body)
            .send()
            .await?;

        if response.status().as_u16() == 204 {
            Ok(())
        } else {
            Err(format!("Failed to accept invitation: {}", response.status()).into())
        }
    }

    /// Reject an invitation
    pub async fn reject_invitation(&self, domain_id: Uuid) -> Result<()> {
        let req_body = serde_json::json!({"domain_id": domain_id});

        let response = self
            .request(reqwest::Method::POST, "/invitations/reject")
            .json(&req_body)
            .send()
            .await?;

        if response.status().as_u16() == 204 {
            Ok(())
        } else {
            Err(format!("Failed to reject invitation: {}", response.status()).into())
        }
    }

    /// Delete an invitation
    pub async fn delete_invitation(&self, domain_id: Uuid, user_id: Uuid) -> Result<()> {
        let req_body = serde_json::json!({"user_id": user_id});

        let response = self
            .request(
                reqwest::Method::DELETE,
                &format!("/domains/{}/invitations", domain_id),
            )
            .json(&req_body)
            .send()
            .await?;

        if response.status().as_u16() == 204 {
            Ok(())
        } else {
            Err(format!("Failed to delete invitation: {}", response.status()).into())
        }
    }

    // === HEALTH CHECK ===

    /// Get service health information
    pub async fn health(&self) -> Result<HealthInfo> {
        let response = self.request(reqwest::Method::GET, "/health").send().await?;

        self.handle_response(response).await
    }
}

// === CONVENIENCE BUILDERS ===

impl CreateDomainRequest {
    pub fn new<S: Into<String>>(name: S, route: S) -> Self {
        Self {
            name: name.into(),
            route: route.into(),
            tags: Vec::new(),
            metadata: HashMap::new(),
        }
    }

    pub fn with_tags(mut self, tags: Vec<String>) -> Self {
        self.tags = tags;
        self
    }

    pub fn with_metadata(mut self, metadata: HashMap<String, serde_json::Value>) -> Self {
        self.metadata = metadata;
        self
    }
}

impl UpdateDomainRequest {
    pub fn new() -> Self {
        Self {
            name: None,
            tags: None,
            metadata: None,
        }
    }

    pub fn with_name<S: Into<String>>(mut self, name: S) -> Self {
        self.name = Some(name.into());
        self
    }

    pub fn with_tags(mut self, tags: Vec<String>) -> Self {
        self.tags = Some(tags);
        self
    }

    pub fn with_metadata(mut self, metadata: HashMap<String, serde_json::Value>) -> Self {
        self.metadata = Some(metadata);
        self
    }
}

impl SendInvitationRequest {
    pub fn new(invitee_user_id: Uuid, role_id: Uuid) -> Self {
        Self {
            invitee_user_id,
            role_id,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_domain_request_builder() {
        let req = CreateDomainRequest::new("test-domain", "test-route")
            .with_tags(vec!["tag1".to_string(), "tag2".to_string()]);

        assert_eq!(req.name, "test-domain");
        assert_eq!(req.route, "test-route");
        assert_eq!(req.tags, vec!["tag1", "tag2"]);
    }

    #[test]
    fn test_client_creation() {
        let client = DomainsClient::new("http://localhost:9003").with_token("test-token");

        assert_eq!(client.base_url, "http://localhost:9003");
        assert_eq!(client.token, Some("test-token".to_string()));
    }
}
