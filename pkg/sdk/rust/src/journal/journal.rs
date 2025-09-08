// Cargo.toml dependencies needed:
// [dependencies]
// serde = { version = "1.0", features = ["derive"] }
// serde_json = "1.0"
// reqwest = { version = "0.11", features = ["json"] }
// chrono = { version = "0.4", features = ["serde"] }
// uuid = { version = "1.0", features = ["v4", "serde"] }
// thiserror = "1.0"
// tokio = { version = "1", features = ["full"] }

use chrono::{DateTime, Utc};
use reqwest::{Client, Response};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;
use uuid::Uuid;

#[derive(Error, Debug)]
pub enum JournalError {
    #[error("HTTP request failed: {0}")]
    RequestError(#[from] reqwest::Error),
    #[error("Serialization error: {0}")]
    SerializationError(#[from] serde_json::Error),
    #[error("API error: {status}: {message}")]
    ApiError { status: u16, message: String },
    #[error("Authentication failed")]
    AuthenticationError,
    #[error("Authorization failed")]
    AuthorizationError,
    #[error("Entity not found")]
    NotFoundError,
    #[error("Invalid request parameters")]
    BadRequestError,
    #[error("Database processing error")]
    DatabaseError,
    #[error("Server error")]
    ServerError,
}

pub type Result<T> = std::result::Result<T, JournalError>;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Telemetry {
    pub client_id: Uuid,
    pub domain_id: Uuid,
    pub subscriptions: i64,
    pub inbound_messages: i64,
    pub outbound_messages: i64,
    pub first_seen: DateTime<Utc>,
    pub last_seen: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Journal {
    pub operation: String,
    pub occurred_at: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub attributes: Option<HashMap<String, serde_json::Value>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JournalPage {
    pub journals: Vec<Journal>,
    pub total: u64,
    pub offset: u64,
    pub limit: Option<u64>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum EntityType {
    Group,
    Client,
    Channel,
}

impl std::fmt::Display for EntityType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            EntityType::Group => write!(f, "group"),
            EntityType::Client => write!(f, "client"),
            EntityType::Channel => write!(f, "channel"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum SortDirection {
    Asc,
    Desc,
}

impl std::fmt::Display for SortDirection {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SortDirection::Asc => write!(f, "asc"),
            SortDirection::Desc => write!(f, "desc"),
        }
    }
}

#[derive(Debug, Clone, Default)]
pub struct JournalQuery {
    pub offset: Option<u64>,
    pub limit: Option<u64>,
    pub operation: Option<String>,
    pub with_attributes: Option<bool>,
    pub with_metadata: Option<bool>,
    pub from: Option<i64>,
    pub to: Option<i64>,
    pub dir: Option<SortDirection>,
}

impl JournalQuery {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn offset(mut self, offset: u64) -> Self {
        self.offset = Some(offset);
        self
    }

    pub fn limit(mut self, limit: u64) -> Self {
        self.limit = Some(limit.min(10)); // API maximum is 10
        self
    }

    pub fn operation(mut self, operation: impl Into<String>) -> Self {
        self.operation = Some(operation.into());
        self
    }

    pub fn with_attributes(mut self, with_attributes: bool) -> Self {
        self.with_attributes = Some(with_attributes);
        self
    }

    pub fn with_metadata(mut self, with_metadata: bool) -> Self {
        self.with_metadata = Some(with_metadata);
        self
    }

    pub fn from_timestamp(mut self, from: i64) -> Self {
        self.from = Some(from);
        self
    }

    pub fn to_timestamp(mut self, to: i64) -> Self {
        self.to = Some(to);
        self
    }

    pub fn sort_direction(mut self, dir: SortDirection) -> Self {
        self.dir = Some(dir);
        self
    }

    pub fn to_query_params(&self) -> Vec<(String, String)> {
        let mut params = Vec::new();

        if let Some(offset) = self.offset {
            params.push(("offset".to_string(), offset.to_string()));
        }
        if let Some(limit) = self.limit {
            params.push(("limit".to_string(), limit.to_string()));
        }
        if let Some(ref operation) = self.operation {
            params.push(("operation".to_string(), operation.clone()));
        }
        if let Some(with_attributes) = self.with_attributes {
            params.push(("with_attributes".to_string(), with_attributes.to_string()));
        }
        if let Some(with_metadata) = self.with_metadata {
            params.push(("with_metadata".to_string(), with_metadata.to_string()));
        }
        if let Some(from) = self.from {
            params.push(("from".to_string(), from.to_string()));
        }
        if let Some(to) = self.to {
            params.push(("to".to_string(), to.to_string()));
        }
        if let Some(dir) = self.dir {
            params.push(("dir".to_string(), dir.to_string()));
        }

        params
    }
}

#[derive(Debug, Clone)]
pub struct JournalClient {
    client: Client,
    base_url: String,
    token: Option<String>,
}

impl JournalClient {
    pub fn new(base_url: impl Into<String>) -> Self {
        Self {
            client: Client::new(),
            base_url: base_url.into(),
            token: None,
        }
    }

    pub fn with_token(mut self, token: impl Into<String>) -> Self {
        self.token = Some(token.into());
        self
    }

    pub fn set_token(&mut self, token: impl Into<String>) {
        self.token = Some(token.into());
    }

    async fn handle_response<T>(&self, response: Response) -> Result<T>
    where
        T: for<'de> Deserialize<'de>,
    {
        let status = response.status();

        if status.is_success() {
            let json = response.json().await?;
            Ok(json)
        } else {
            let error_text = response.text().await.unwrap_or_default();

            match status.as_u16() {
                400 => Err(JournalError::BadRequestError),
                401 => Err(JournalError::AuthenticationError),
                403 => Err(JournalError::AuthorizationError),
                404 => Err(JournalError::NotFoundError),
                422 => Err(JournalError::DatabaseError),
                500..=599 => Err(JournalError::ServerError),
                _ => Err(JournalError::ApiError {
                    status: status.as_u16(),
                    message: error_text,
                }),
            }
        }
    }

    fn build_request(&self, method: reqwest::Method, url: &str) -> reqwest::RequestBuilder {
        let mut builder = self.client.request(method, url);

        if let Some(ref token) = self.token {
            builder = builder.bearer_auth(token);
        }

        builder
    }

    /// List user journal log
    pub async fn list_user_journal(
        &self,
        user_id: Uuid,
        query: Option<JournalQuery>,
    ) -> Result<JournalPage> {
        let mut url = format!("{}/journal/user/{}", self.base_url, user_id);

        if let Some(query) = query {
            let params = query.to_query_params();
            if !params.is_empty() {
                let query_string = params
                    .iter()
                    .map(|(k, v)| format!("{}={}", k, v))
                    .collect::<Vec<_>>()
                    .join("&");
                url.push_str(&format!("?{}", query_string));
            }
        }

        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// View client telemetry
    pub async fn get_client_telemetry(
        &self,
        domain_id: Uuid,
        client_id: Uuid,
    ) -> Result<Telemetry> {
        let url = format!(
            "{}/{}/journal/client/{}/telemetry",
            self.base_url, domain_id, client_id
        );

        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// List entity journal log
    pub async fn list_entity_journal(
        &self,
        domain_id: Uuid,
        entity_type: EntityType,
        entity_id: Uuid,
        query: Option<JournalQuery>,
    ) -> Result<JournalPage> {
        let mut url = format!(
            "{}/{}/journal/{}/{}",
            self.base_url, domain_id, entity_type, entity_id
        );

        if let Some(query) = query {
            let params = query.to_query_params();
            if !params.is_empty() {
                let query_string = params
                    .iter()
                    .map(|(k, v)| format!("{}={}", k, v))
                    .collect::<Vec<_>>()
                    .join("&");
                url.push_str(&format!("?{}", query_string));
            }
        }

        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Check service health
    pub async fn health_check(&self) -> Result<serde_json::Value> {
        let url = format!("{}/health", self.base_url);

        let response = self.client.get(&url).send().await?;

        self.handle_response(response).await
    }
}

// Convenience methods for common operations
impl JournalClient {
    /// Get all journals for a user with pagination
    pub async fn get_all_user_journals(&self, user_id: Uuid) -> Result<Vec<Journal>> {
        let mut all_journals = Vec::new();
        let mut offset = 0;
        let limit = 10; // API maximum

        loop {
            let query = JournalQuery::new().offset(offset).limit(limit);

            let page = self.list_user_journal(user_id, Some(query)).await?;

            if page.journals.is_empty() {
                break;
            }

            all_journals.extend(page.journals);

            if page.journals.len() < limit as usize {
                break;
            }

            offset += limit;
        }

        Ok(all_journals)
    }

    /// Get all journals for an entity with pagination
    pub async fn get_all_entity_journals(
        &self,
        domain_id: Uuid,
        entity_type: EntityType,
        entity_id: Uuid,
    ) -> Result<Vec<Journal>> {
        let mut all_journals = Vec::new();
        let mut offset = 0;
        let limit = 10; // API maximum

        loop {
            let query = JournalQuery::new().offset(offset).limit(limit);

            let page = self
                .list_entity_journal(domain_id, entity_type, entity_id, Some(query))
                .await?;

            if page.journals.is_empty() {
                break;
            }

            all_journals.extend(page.journals);

            if page.journals.len() < limit as usize {
                break;
            }

            offset += limit;
        }

        Ok(all_journals)
    }

    /// Search for journals by operation
    pub async fn search_user_journals_by_operation(
        &self,
        user_id: Uuid,
        operation: impl Into<String>,
    ) -> Result<Vec<Journal>> {
        let query = JournalQuery::new()
            .operation(operation)
            .with_attributes(true)
            .with_metadata(true);

        let page = self.list_user_journal(user_id, Some(query)).await?;
        Ok(page.journals)
    }

    /// Search for entity journals by operation
    pub async fn search_entity_journals_by_operation(
        &self,
        domain_id: Uuid,
        entity_type: EntityType,
        entity_id: Uuid,
        operation: impl Into<String>,
    ) -> Result<Vec<Journal>> {
        let query = JournalQuery::new()
            .operation(operation)
            .with_attributes(true)
            .with_metadata(true);

        let page = self
            .list_entity_journal(domain_id, entity_type, entity_id, Some(query))
            .await?;
        Ok(page.journals)
    }

    /// Get journals within a time range
    pub async fn get_user_journals_in_range(
        &self,
        user_id: Uuid,
        from: i64,
        to: i64,
    ) -> Result<Vec<Journal>> {
        let query = JournalQuery::new()
            .from_timestamp(from)
            .to_timestamp(to)
            .sort_direction(SortDirection::Desc);

        let page = self.list_user_journal(user_id, Some(query)).await?;
        Ok(page.journals)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_journal_client_creation() {
        let client = JournalClient::new("http://localhost:9021");
        assert_eq!(client.base_url, "http://localhost:9021");
        assert!(client.token.is_none());
    }

    #[test]
    fn test_journal_query_builder() {
        let query = JournalQuery::new()
            .offset(10)
            .limit(5)
            .operation("user.create")
            .with_attributes(true)
            .sort_direction(SortDirection::Desc);

        let params = query.to_query_params();
        assert_eq!(params.len(), 5);

        // Check that all parameters are present
        let param_map: HashMap<String, String> = params.into_iter().collect();
        assert_eq!(param_map.get("offset").unwrap(), "10");
        assert_eq!(param_map.get("limit").unwrap(), "5");
        assert_eq!(param_map.get("operation").unwrap(), "user.create");
        assert_eq!(param_map.get("with_attributes").unwrap(), "true");
        assert_eq!(param_map.get("dir").unwrap(), "desc");
    }

    #[test]
    fn test_entity_type_display() {
        assert_eq!(EntityType::Group.to_string(), "group");
        assert_eq!(EntityType::Client.to_string(), "client");
        assert_eq!(EntityType::Channel.to_string(), "channel");
    }

    #[test]
    fn test_sort_direction_display() {
        assert_eq!(SortDirection::Asc.to_string(), "asc");
        assert_eq!(SortDirection::Desc.to_string(), "desc");
    }
}
