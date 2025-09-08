// certs.rs - Rust SDK for Certificates API
use reqwest::{Client, Error as ReqwestError, Response};
use serde::{Deserialize, Serialize};
// use std::collections::HashMap;

// Error types
#[derive(Debug, thiserror::Error)]
pub enum CertsError {
    #[error("HTTP request failed: {0}")]
    Http(#[from] ReqwestError),
    #[error("Serialization failed: {0}")]
    Serialization(#[from] serde_json::Error),
    #[error("API error: {status} - {message}")]
    Api { status: u16, message: String },
    #[error("Authentication failed")]
    Authentication,
    #[error("Certificate not found")]
    NotFound,
}

pub type Result<T> = std::result::Result<T, CertsError>;

// Certificate models
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Certificate {
    pub serial: String,
    pub certificate: String,
    pub key: Option<String>,
    pub revoked: bool,
    pub expires_at: String,
    pub entity_id: String,
    pub created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IssueCertificateRequest {
    pub entity_id: String,
    pub ttl: String,
    pub key_type: Option<String>,
    pub key_bits: Option<u32>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CertificatesPage {
    pub certificates: Vec<Certificate>,
    pub total: u64,
    pub offset: u64,
    pub limit: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PageMetadata {
    pub total: u64,
    pub offset: u64,
    pub limit: u64,
}

// SDK Configuration
#[derive(Debug, Clone)]
pub struct CertsConfig {
    pub base_url: String,
    pub token: Option<String>,
    pub timeout: std::time::Duration,
}

impl Default for CertsConfig {
    fn default() -> Self {
        Self {
            base_url: "http://localhost:9019".to_string(),
            token: None,
            timeout: std::time::Duration::from_secs(30),
        }
    }
}

// Main SDK client
#[derive(Debug)]
pub struct CertsSDK {
    client: Client,
    config: CertsConfig,
}

impl CertsSDK {
    pub fn new(config: CertsConfig) -> Result<Self> {
        let client = Client::builder().timeout(config.timeout).build()?;

        Ok(Self { client, config })
    }

    // Create a new SDK instance with default config
    pub fn with_base_url(base_url: &str) -> Result<Self> {
        let config = CertsConfig {
            base_url: base_url.to_string(),
            ..Default::default()
        };
        Self::new(config)
    }

    // Set authentication token
    pub fn with_token(mut self, token: &str) -> Self {
        self.config.token = Some(token.to_string());
        self
    }

    // Helper method to build request with auth
    fn build_request(&self, method: reqwest::Method, path: &str) -> reqwest::RequestBuilder {
        let url = format!("{}{}", self.config.base_url, path);
        let mut request = self.client.request(method, &url);

        if let Some(token) = &self.config.token {
            request = request.bearer_auth(token);
        }

        request.header("Content-Type", "application/json")
    }

    // Helper method to handle API responses
    async fn handle_response<T: serde::de::DeserializeOwned>(
        &self,
        response: Response,
    ) -> Result<T> {
        let status = response.status();

        if status.is_success() {
            let body = response.text().await?;
            Ok(serde_json::from_str(&body)?)
        } else {
            let body = response.text().await.unwrap_or_default();
            Err(CertsError::Api {
                status: status.as_u16(),
                message: body,
            })
        }
    }

    // Issue a new certificate
    pub async fn issue(&self, req: IssueCertificateRequest) -> Result<Certificate> {
        let response = self
            .build_request(reqwest::Method::POST, "/certs")
            .json(&req)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Get certificate by serial number
    pub async fn view(&self, serial: &str) -> Result<Certificate> {
        let path = format!("/certs/{}", serial);
        let response = self
            .build_request(reqwest::Method::GET, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // List certificates with pagination
    pub async fn list(&self, offset: Option<u64>, limit: Option<u64>) -> Result<CertificatesPage> {
        let mut path = "/certs".to_string();
        let mut params = Vec::new();

        if let Some(offset) = offset {
            params.push(format!("offset={}", offset));
        }
        if let Some(limit) = limit {
            params.push(format!("limit={}", limit));
        }

        if !params.is_empty() {
            path.push('?');
            path.push_str(&params.join("&"));
        }

        let response = self
            .build_request(reqwest::Method::GET, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Revoke a certificate
    pub async fn revoke(&self, serial: &str) -> Result<Certificate> {
        let path = format!("/certs/{}/revoke", serial);
        let response = self
            .build_request(reqwest::Method::DELETE, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Get certificates by entity ID
    pub async fn get_by_entity(
        &self,
        entity_id: &str,
        offset: Option<u64>,
        limit: Option<u64>,
    ) -> Result<CertificatesPage> {
        let mut path = format!("/certs?entity_id={}", entity_id);

        if let Some(offset) = offset {
            path.push_str(&format!("&offset={}", offset));
        }
        if let Some(limit) = limit {
            path.push_str(&format!("&limit={}", limit));
        }

        let response = self
            .build_request(reqwest::Method::GET, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }
}

// Convenience functions for common operations
impl CertsSDK {
    // Issue certificate with default TTL
    pub async fn issue_default(&self, entity_id: &str) -> Result<Certificate> {
        let req = IssueCertificateRequest {
            entity_id: entity_id.to_string(),
            ttl: "8760h".to_string(), // 1 year
            key_type: None,
            key_bits: None,
        };
        self.issue(req).await
    }

    // Check if certificate is revoked
    pub async fn is_revoked(&self, serial: &str) -> Result<bool> {
        let cert = self.view(serial).await?;
        Ok(cert.revoked)
    }

    // Get all certificates (with automatic pagination)
    pub async fn list_all(&self) -> Result<Vec<Certificate>> {
        let mut all_certs = Vec::new();
        let mut offset = 0;
        let limit = 100;

        loop {
            let page = self.list(Some(offset), Some(limit)).await?;
            let fetched = page.certificates.len();
            all_certs.extend(page.certificates);

            if fetched < limit as usize {
                break;
            }
            offset += limit;
        }

        Ok(all_certs)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_sdk_creation() {
        let config = CertsConfig::default();
        let sdk = CertsSDK::new(config).unwrap();
        assert!(sdk.config.base_url.contains("localhost"));
    }

    #[tokio::test]
    async fn test_with_token() {
        let sdk = CertsSDK::with_base_url("http://localhost:9019")
            .unwrap()
            .with_token("test-token");

        assert!(sdk.config.token.is_some());
        assert_eq!(sdk.config.token.unwrap(), "test-token");
    }
}
