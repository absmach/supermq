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
pub enum TwinsError {
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
    #[error("Invalid content type")]
    InvalidContentType,
    #[error("Database processing error")]
    DatabaseError,
    #[error("Server error")]
    ServerError,
}

pub type Result<T> = std::result::Result<T, TwinsError>;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attribute {
    pub name: String,
    pub channel: String,
    pub subtopic: String,
    pub persist_state: bool,
}

impl Attribute {
    pub fn new(
        name: impl Into<String>,
        channel: impl Into<String>,
        subtopic: impl Into<String>,
        persist_state: bool,
    ) -> Self {
        Self {
            name: name.into(),
            channel: channel.into(),
            subtopic: subtopic.into(),
            persist_state,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Definition {
    pub delta: f64,
    pub attributes: Vec<Attribute>,
}

impl Definition {
    pub fn new(delta: f64) -> Self {
        Self {
            delta,
            attributes: Vec::new(),
        }
    }

    pub fn with_attributes(mut self, attributes: Vec<Attribute>) -> Self {
        self.attributes = attributes;
        self
    }

    pub fn add_attribute(mut self, attribute: Attribute) -> Self {
        self.attributes.push(attribute);
        self
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TwinRequest {
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub definition: Option<Definition>,
}

impl TwinRequest {
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            name: name.into(),
            metadata: None,
            definition: None,
        }
    }

    pub fn with_metadata(mut self, metadata: HashMap<String, serde_json::Value>) -> Self {
        self.metadata = Some(metadata);
        self
    }

    pub fn with_definition(mut self, definition: Definition) -> Self {
        self.definition = Some(definition);
        self
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Twin {
    pub owner: String,
    pub id: Uuid,
    pub name: String,
    pub revision: f64,
    pub created: DateTime<Utc>,
    pub updated: DateTime<Utc>,
    pub definitions: Vec<Definition>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TwinsPage {
    pub twins: Vec<Twin>,
    pub total: u64,
    pub offset: u64,
    pub limit: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct State {
    pub twin_id: Uuid,
    pub id: u64,
    pub created: DateTime<Utc>,
    pub payload: HashMap<String, serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StatesPage {
    pub states: Vec<State>,
    pub total: u64,
    pub offset: u64,
    pub limit: u64,
}

#[derive(Debug, Clone, Default)]
pub struct TwinsQuery {
    pub limit: Option<u64>,
    pub offset: Option<u64>,
    pub name: Option<String>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
}

impl TwinsQuery {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn limit(mut self, limit: u64) -> Self {
        self.limit = Some(limit.min(100)); // API maximum is 100
        self
    }

    pub fn offset(mut self, offset: u64) -> Self {
        self.offset = Some(offset);
        self
    }

    pub fn name(mut self, name: impl Into<String>) -> Self {
        self.name = Some(name.into());
        self
    }

    pub fn metadata(mut self, metadata: HashMap<String, serde_json::Value>) -> Self {
        self.metadata = Some(metadata);
        self
    }

    pub fn to_query_params(&self) -> Result<Vec<(String, String)>> {
        let mut params = Vec::new();

        if let Some(limit) = self.limit {
            params.push(("limit".to_string(), limit.to_string()));
        }
        if let Some(offset) = self.offset {
            params.push(("offset".to_string(), offset.to_string()));
        }
        if let Some(ref name) = self.name {
            params.push(("name".to_string(), name.clone()));
        }
        if let Some(ref metadata) = self.metadata {
            let metadata_str = serde_json::to_string(metadata)?;
            params.push(("metadata".to_string(), metadata_str));
        }

        Ok(params)
    }
}

#[derive(Debug, Clone, Default)]
pub struct StatesQuery {
    pub limit: Option<u64>,
    pub offset: Option<u64>,
}

impl StatesQuery {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn limit(mut self, limit: u64) -> Self {
        self.limit = Some(limit.min(100)); // API maximum is 100
        self
    }

    pub fn offset(mut self, offset: u64) -> Self {
        self.offset = Some(offset);
        self
    }

    pub fn to_query_params(&self) -> Vec<(String, String)> {
        let mut params = Vec::new();

        if let Some(limit) = self.limit {
            params.push(("limit".to_string(), limit.to_string()));
        }
        if let Some(offset) = self.offset {
            params.push(("offset".to_string(), offset.to_string()));
        }

        params
    }
}

#[derive(Debug, Clone)]
pub struct TwinsClient {
    client: Client,
    base_url: String,
    token: Option<String>,
}

impl TwinsClient {
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
                400 => Err(TwinsError::BadRequestError),
                401 => Err(TwinsError::AuthenticationError),
                404 => Err(TwinsError::NotFoundError),
                415 => Err(TwinsError::InvalidContentType),
                422 => Err(TwinsError::DatabaseError),
                500..=599 => Err(TwinsError::ServerError),
                _ => Err(TwinsError::ApiError {
                    status: status.as_u16(),
                    message: error_text,
                }),
            }
        }
    }

    async fn handle_empty_response(&self, response: Response) -> Result<()> {
        let status = response.status();

        if status.is_success() {
            Ok(())
        } else {
            let error_text = response.text().await.unwrap_or_default();

            match status.as_u16() {
                400 => Err(TwinsError::BadRequestError),
                401 => Err(TwinsError::AuthenticationError),
                404 => Err(TwinsError::NotFoundError),
                415 => Err(TwinsError::InvalidContentType),
                422 => Err(TwinsError::DatabaseError),
                500..=599 => Err(TwinsError::ServerError),
                _ => Err(TwinsError::ApiError {
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

    /// Create a new twin
    pub async fn create_twin(&self, twin_req: TwinRequest) -> Result<String> {
        let url = format!("{}/twins", self.base_url);

        let response = self
            .build_request(reqwest::Method::POST, &url)
            .header("Content-Type", "application/json")
            .json(&twin_req)
            .send()
            .await?;

        if response.status().is_success() {
            let location = response
                .headers()
                .get("Location")
                .and_then(|h| h.to_str().ok())
                .unwrap_or_default()
                .to_string();
            Ok(location)
        } else {
            let error_text = response.text().await.unwrap_or_default();
            match response.status().as_u16() {
                400 => Err(TwinsError::BadRequestError),
                401 => Err(TwinsError::AuthenticationError),
                415 => Err(TwinsError::InvalidContentType),
                422 => Err(TwinsError::DatabaseError),
                500..=599 => Err(TwinsError::ServerError),
                status => Err(TwinsError::ApiError {
                    status,
                    message: error_text,
                }),
            }
        }
    }

    /// Retrieve a paginated list of twins
    pub async fn get_twins(&self, query: Option<TwinsQuery>) -> Result<TwinsPage> {
        let mut url = format!("{}/twins", self.base_url);

        if let Some(query) = query {
            let params = query.to_query_params()?;
            if !params.is_empty() {
                let query_string = params
                    .iter()
                    .map(|(k, v)| format!("{}={}", urlencoding::encode(k), urlencoding::encode(v)))
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

    /// Retrieve a specific twin by ID
    pub async fn get_twin(&self, twin_id: Uuid) -> Result<Twin> {
        let url = format!("{}/twins/{}", self.base_url, twin_id);

        let response = self
            .build_request(reqwest::Method::GET, &url)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Update an existing twin
    pub async fn update_twin(&self, twin_id: Uuid, twin_req: TwinRequest) -> Result<()> {
        let url = format!("{}/twins/{}", self.base_url, twin_id);

        let response = self
            .build_request(reqwest::Method::PUT, &url)
            .header("Content-Type", "application/json")
            .json(&twin_req)
            .send()
            .await?;

        self.handle_empty_response(response).await
    }

    /// Delete a twin
    pub async fn delete_twin(&self, twin_id: Uuid) -> Result<()> {
        let url = format!("{}/twins/{}", self.base_url, twin_id);

        let response = self
            .build_request(reqwest::Method::DELETE, &url)
            .send()
            .await?;

        self.handle_empty_response(response).await
    }

    /// Retrieve states for a specific twin
    pub async fn get_states(
        &self,
        twin_id: Uuid,
        query: Option<StatesQuery>,
    ) -> Result<StatesPage> {
        let mut url = format!("{}/states/{}", self.base_url, twin_id);

        if let Some(query) = query {
            let params = query.to_query_params();
            if !params.is_empty() {
                let query_string = params
                    .iter()
                    .map(|(k, v)| format!("{}={}", urlencoding::encode(k), urlencoding::encode(v)))
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
impl TwinsClient {
    /// Get all twins with automatic pagination
    pub async fn get_all_twins(&self) -> Result<Vec<Twin>> {
        let mut all_twins = Vec::new();
        let mut offset = 0;
        let limit = 100; // API maximum

        loop {
            let query = TwinsQuery::new().offset(offset).limit(limit);

            let page = self.get_twins(Some(query)).await?;

            if page.twins.is_empty() {
                break;
            }

            all_twins.extend(page.twins);

            if page.twins.len() < limit as usize {
                break;
            }

            offset += limit;
        }

        Ok(all_twins)
    }

    /// Get all states for a twin with automatic pagination
    pub async fn get_all_states(&self, twin_id: Uuid) -> Result<Vec<State>> {
        let mut all_states = Vec::new();
        let mut offset = 0;
        let limit = 100; // API maximum

        loop {
            let query = StatesQuery::new().offset(offset).limit(limit);

            let page = self.get_states(twin_id, Some(query)).await?;

            if page.states.is_empty() {
                break;
            }

            all_states.extend(page.states);

            if page.states.len() < limit as usize {
                break;
            }

            offset += limit;
        }

        Ok(all_states)
    }

    /// Search twins by name
    pub async fn search_twins_by_name(&self, name: impl Into<String>) -> Result<Vec<Twin>> {
        let query = TwinsQuery::new().name(name);
        let page = self.get_twins(Some(query)).await?;
        Ok(page.twins)
    }

    /// Search twins by metadata
    pub async fn search_twins_by_metadata(
        &self,
        metadata: HashMap<String, serde_json::Value>,
    ) -> Result<Vec<Twin>> {
        let query = TwinsQuery::new().metadata(metadata);
        let page = self.get_twins(Some(query)).await?;
        Ok(page.twins)
    }

    /// Create a simple twin with just a name
    pub async fn create_simple_twin(&self, name: impl Into<String>) -> Result<String> {
        let twin_req = TwinRequest::new(name);
        self.create_twin(twin_req).await
    }

    /// Create a twin with metadata
    pub async fn create_twin_with_metadata(
        &self,
        name: impl Into<String>,
        metadata: HashMap<String, serde_json::Value>,
    ) -> Result<String> {
        let twin_req = TwinRequest::new(name).with_metadata(metadata);
        self.create_twin(twin_req).await
    }

    /// Update twin name
    pub async fn update_twin_name(&self, twin_id: Uuid, new_name: impl Into<String>) -> Result<()> {
        // First get the existing twin to preserve other fields
        let existing_twin = self.get_twin(twin_id).await?;

        let twin_req = TwinRequest {
            name: new_name.into(),
            metadata: existing_twin.metadata,
            definition: existing_twin.definitions.into_iter().next(),
        };

        self.update_twin(twin_id, twin_req).await
    }

    /// Update twin metadata
    pub async fn update_twin_metadata(
        &self,
        twin_id: Uuid,
        metadata: HashMap<String, serde_json::Value>,
    ) -> Result<()> {
        // First get the existing twin to preserve other fields
        let existing_twin = self.get_twin(twin_id).await?;

        let twin_req = TwinRequest {
            name: existing_twin.name,
            metadata: Some(metadata),
            definition: existing_twin.definitions.into_iter().next(),
        };

        self.update_twin(twin_id, twin_req).await
    }

    /// Check if a twin exists
    pub async fn twin_exists(&self, twin_id: Uuid) -> Result<bool> {
        match self.get_twin(twin_id).await {
            Ok(_) => Ok(true),
            Err(TwinsError::NotFoundError) => Ok(false),
            Err(e) => Err(e),
        }
    }

    /// Get twin count (using the total from first page)
    pub async fn get_twin_count(&self) -> Result<u64> {
        let query = TwinsQuery::new().limit(1);
        let page = self.get_twins(Some(query)).await?;
        Ok(page.total)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_twins_client_creation() {
        let client = TwinsClient::new("http://localhost:9018");
        assert_eq!(client.base_url, "http://localhost:9018");
        assert!(client.token.is_none());
    }

    #[test]
    fn test_twin_request_builder() {
        let mut metadata = HashMap::new();
        metadata.insert(
            "key1".to_string(),
            serde_json::Value::String("value1".to_string()),
        );

        let definition = Definition::new(1.5).add_attribute(Attribute::new(
            "temp",
            "channel1",
            "temperature",
            true,
        ));

        let twin_req = TwinRequest::new("test-twin")
            .with_metadata(metadata.clone())
            .with_definition(definition);

        assert_eq!(twin_req.name, "test-twin");
        assert!(twin_req.metadata.is_some());
        assert!(twin_req.definition.is_some());
    }

    #[test]
    fn test_twins_query_builder() {
        let mut metadata = HashMap::new();
        metadata.insert(
            "type".to_string(),
            serde_json::Value::String("sensor".to_string()),
        );

        let query = TwinsQuery::new()
            .limit(50)
            .offset(10)
            .name("test")
            .metadata(metadata);

        let params = query.to_query_params().unwrap();
        assert_eq!(params.len(), 4);

        // Check that all parameters are present
        let param_map: HashMap<String, String> = params.into_iter().collect();
        assert_eq!(param_map.get("limit").unwrap(), "50");
        assert_eq!(param_map.get("offset").unwrap(), "10");
        assert_eq!(param_map.get("name").unwrap(), "test");
        assert!(param_map.contains_key("metadata"));
    }

    #[test]
    fn test_states_query_builder() {
        let query = StatesQuery::new().limit(25).offset(5);

        let params = query.to_query_params();
        assert_eq!(params.len(), 2);

        let param_map: HashMap<String, String> = params.into_iter().collect();
        assert_eq!(param_map.get("limit").unwrap(), "25");
        assert_eq!(param_map.get("offset").unwrap(), "5");
    }

    #[test]
    fn test_attribute_creation() {
        let attr = Attribute::new("temperature", "temp_channel", "sensors/temp", true);
        assert_eq!(attr.name, "temperature");
        assert_eq!(attr.channel, "temp_channel");
        assert_eq!(attr.subtopic, "sensors/temp");
        assert!(attr.persist_state);
    }

    #[test]
    fn test_definition_builder() {
        let attr1 = Attribute::new("temp", "ch1", "temp", true);
        let attr2 = Attribute::new("humidity", "ch2", "hum", false);

        let definition = Definition::new(2.0)
            .add_attribute(attr1)
            .add_attribute(attr2);

        assert_eq!(definition.delta, 2.0);
        assert_eq!(definition.attributes.len(), 2);
        assert_eq!(definition.attributes[0].name, "temp");
        assert_eq!(definition.attributes[1].name, "humidity");
    }
}
