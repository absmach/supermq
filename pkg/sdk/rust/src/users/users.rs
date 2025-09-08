// users.rs - Rust SDK for SuperMQ Users API
use reqwest::{Client, Error as ReqwestError, Response};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// Error types
#[derive(Debug, thiserror::Error)]
pub enum UsersError {
    #[error("HTTP request failed: {0}")]
    Http(#[from] ReqwestError),
    #[error("Serialization failed: {0}")]
    Serialization(#[from] serde_json::Error),
    #[error("API error: {status} - {message}")]
    Api { status: u16, message: String },
    #[error("Authentication failed")]
    Authentication,
    #[error("User not found")]
    NotFound,
    #[error("Validation failed")]
    Validation,
}

pub type Result<T> = std::result::Result<T, UsersError>;

// User models matching OpenAPI spec
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Credentials {
    pub username: String,
    pub secret: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserReqObj {
    pub first_name: String,
    pub last_name: String,
    pub email: String,
    pub tags: Option<Vec<String>>,
    pub credentials: Credentials,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    pub profile_picture: Option<String>,
    pub status: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    pub first_name: String,
    pub last_name: String,
    pub tags: Option<Vec<String>>,
    pub email: String,
    pub credentials: Option<Credentials>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    pub profile_picture: Option<String>,
    pub status: Option<String>,
    pub created_at: Option<String>,
    pub updated_at: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UsersPage {
    pub users: Vec<User>,
    pub total: u64,
    pub offset: u64,
    pub limit: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserUpdate {
    pub first_name: String,
    pub last_name: String,
    pub metadata: HashMap<String, serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserTags {
    pub tags: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserProfilePicture {
    pub profile_picture: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Email {
    pub email: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserSecret {
    pub old_secret: String,
    pub new_secret: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UserRole {
    pub role: String, // "admin" or "user"
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Username {
    pub username: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IssueToken {
    pub username: String,
    pub password: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenRes {
    pub access_token: String,
    pub refresh_token: String,
    pub access_type: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RequestPasswordReset {
    pub email: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PasswordReset {
    pub password: String,
    pub confirm_password: String,
    pub token: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthRes {
    pub status: String,
    pub version: String,
    pub commit: String,
    pub description: String,
    pub build_time: String,
}

// Query parameters for listing/searching
#[derive(Debug, Clone, Default)]
pub struct UserListParams {
    pub limit: Option<u64>,
    pub offset: Option<u64>,
    pub metadata: Option<HashMap<String, serde_json::Value>>,
    pub status: Option<String>,
    pub first_name: Option<String>,
    pub last_name: Option<String>,
    pub username: Option<String>,
    pub email: Option<String>,
    pub tag: Option<String>,
}

#[derive(Debug, Clone, Default)]
pub struct UserSearchParams {
    pub limit: Option<u64>,
    pub offset: Option<u64>,
    pub username: Option<String>,
    pub first_name: Option<String>,
    pub last_name: Option<String>,
    pub email: Option<String>,
    pub user_id: Option<String>,
}

// SDK Configuration
#[derive(Debug, Clone)]
pub struct UsersConfig {
    pub base_url: String,
    pub token: Option<String>,
    pub timeout: std::time::Duration,
}

impl Default for UsersConfig {
    fn default() -> Self {
        Self {
            base_url: "http://localhost:9002".to_string(),
            token: None,
            timeout: std::time::Duration::from_secs(30),
        }
    }
}

// Main SDK client
#[derive(Debug)]
pub struct UsersSDK {
    client: Client,
    config: UsersConfig,
}

impl UsersSDK {
    pub fn new(config: UsersConfig) -> Result<Self> {
        let client = Client::builder().timeout(config.timeout).build()?;

        Ok(Self { client, config })
    }

    pub fn with_base_url(base_url: &str) -> Result<Self> {
        let config = UsersConfig {
            base_url: base_url.to_string(),
            ..Default::default()
        };
        Self::new(config)
    }

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

    // Helper to handle API responses
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
            Err(UsersError::Api {
                status: status.as_u16(),
                message: body,
            })
        }
    }

    // Helper to handle responses that return no content (204)
    async fn handle_empty_response(&self, response: Response) -> Result<()> {
        let status = response.status();

        if status.is_success() {
            Ok(())
        } else {
            let body = response.text().await.unwrap_or_default();
            Err(UsersError::Api {
                status: status.as_u16(),
                message: body,
            })
        }
    }

    // Helper to build query parameters
    fn build_query_params<T: Serialize>(&self, params: &T) -> Vec<(String, String)> {
        // This is a simplified version - in a real implementation you'd want proper query serialization
        Vec::new()
    }

    // User Management Methods

    // Create user (POST /users)
    pub async fn create_user(&self, user: UserReqObj) -> Result<User> {
        let response = self
            .build_request(reqwest::Method::POST, "/users")
            .json(&user)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // List users (GET /users)
    pub async fn list_users(&self, params: Option<UserListParams>) -> Result<UsersPage> {
        let mut path = "/users".to_string();

        if let Some(params) = params {
            let mut query_params = Vec::new();
            if let Some(limit) = params.limit {
                query_params.push(format!("limit={}", limit));
            }
            if let Some(offset) = params.offset {
                query_params.push(format!("offset={}", offset));
            }
            if let Some(status) = params.status {
                query_params.push(format!("status={}", status));
            }
            if let Some(first_name) = params.first_name {
                query_params.push(format!("first_name={}", first_name));
            }
            if let Some(last_name) = params.last_name {
                query_params.push(format!("last_name={}", last_name));
            }
            if let Some(username) = params.username {
                query_params.push(format!("username={}", username));
            }
            if let Some(email) = params.email {
                query_params.push(format!("email={}", email));
            }
            if let Some(tag) = params.tag {
                query_params.push(format!("tag={}", tag));
            }

            if !query_params.is_empty() {
                path.push('?');
                path.push_str(&query_params.join("&"));
            }
        }

        let response = self
            .build_request(reqwest::Method::GET, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Get profile (GET /users/profile)
    pub async fn get_profile(&self) -> Result<User> {
        let response = self
            .build_request(reqwest::Method::GET, "/users/profile")
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Get user by ID (GET /users/{userID})
    pub async fn get_user(&self, user_id: &str) -> Result<User> {
        let path = format!("/users/{}", user_id);
        let response = self
            .build_request(reqwest::Method::GET, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Update user (PATCH /users/{userID})
    pub async fn update_user(&self, user_id: &str, update: UserUpdate) -> Result<User> {
        let path = format!("/users/{}", user_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &path)
            .json(&update)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Delete user (DELETE /users/{userID})
    pub async fn delete_user(&self, user_id: &str) -> Result<()> {
        let path = format!("/users/{}", user_id);
        let response = self
            .build_request(reqwest::Method::DELETE, &path)
            .send()
            .await?;

        self.handle_empty_response(response).await
    }

    // Update username (PATCH /users/{userID}/username)
    pub async fn update_username(&self, user_id: &str, username: Username) -> Result<User> {
        let path = format!("/users/{}/username", user_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &path)
            .json(&username)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Update tags (PATCH /users/{userID}/tags)
    pub async fn update_tags(&self, user_id: &str, tags: UserTags) -> Result<User> {
        let path = format!("/users/{}/tags", user_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &path)
            .json(&tags)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Update profile picture (PATCH /users/{userID}/picture)
    pub async fn update_profile_picture(
        &self,
        user_id: &str,
        picture: UserProfilePicture,
    ) -> Result<User> {
        let path = format!("/users/{}/picture", user_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &path)
            .json(&picture)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Update email (PATCH /users/{userID}/email)
    pub async fn update_email(&self, user_id: &str, email: Email) -> Result<User> {
        let path = format!("/users/{}/email", user_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &path)
            .json(&email)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Update role (PATCH /users/{userID}/role)
    pub async fn update_role(&self, user_id: &str, role: UserRole) -> Result<User> {
        let path = format!("/users/{}/role", user_id);
        let response = self
            .build_request(reqwest::Method::PATCH, &path)
            .json(&role)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Disable user (POST /users/{userID}/disable)
    pub async fn disable_user(&self, user_id: &str) -> Result<User> {
        let path = format!("/users/{}/disable", user_id);
        let response = self
            .build_request(reqwest::Method::POST, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Enable user (POST /users/{userID}/enable)
    pub async fn enable_user(&self, user_id: &str) -> Result<User> {
        let path = format!("/users/{}/enable", user_id);
        let response = self
            .build_request(reqwest::Method::POST, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Update secret/password (PATCH /users/secret)
    pub async fn update_secret(&self, secret: UserSecret) -> Result<User> {
        let response = self
            .build_request(reqwest::Method::PATCH, "/users/secret")
            .json(&secret)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Search users (GET /users/search)
    pub async fn search_users(&self, params: Option<UserSearchParams>) -> Result<UsersPage> {
        let mut path = "/users/search".to_string();

        if let Some(params) = params {
            let mut query_params = Vec::new();
            if let Some(limit) = params.limit {
                query_params.push(format!("limit={}", limit));
            }
            if let Some(offset) = params.offset {
                query_params.push(format!("offset={}", offset));
            }
            if let Some(username) = params.username {
                query_params.push(format!("username={}", username));
            }
            if let Some(first_name) = params.first_name {
                query_params.push(format!("first_name={}", first_name));
            }
            if let Some(last_name) = params.last_name {
                query_params.push(format!("last_name={}", last_name));
            }
            if let Some(email) = params.email {
                query_params.push(format!("email={}", email));
            }
            if let Some(user_id) = params.user_id {
                query_params.push(format!("userID={}", user_id));
            }

            if !query_params.is_empty() {
                path.push('?');
                path.push_str(&query_params.join("&"));
            }
        }

        let response = self
            .build_request(reqwest::Method::GET, &path)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Authentication Methods

    // Issue token (POST /users/tokens/issue)
    pub async fn issue_token(&self, credentials: IssueToken) -> Result<TokenRes> {
        let response = self
            .build_request(reqwest::Method::POST, "/users/tokens/issue")
            .json(&credentials)
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Refresh token (POST /users/tokens/refresh)
    pub async fn refresh_token(&self) -> Result<TokenRes> {
        let response = self
            .build_request(reqwest::Method::POST, "/users/tokens/refresh")
            .send()
            .await?;

        self.handle_response(response).await
    }

    // Password Management

    // Request password reset (POST /password/reset-request)
    pub async fn request_password_reset(
        &self,
        email_req: RequestPasswordReset,
        referer: Option<&str>,
    ) -> Result<()> {
        let mut request = self.build_request(reqwest::Method::POST, "/password/reset-request");

        if let Some(referer) = referer {
            request = request.header("Referer", referer);
        }

        let response = request.json(&email_req).send().await?;

        self.handle_empty_response(response).await
    }

    // Reset password (PUT /password/reset)
    pub async fn reset_password(&self, reset_data: PasswordReset) -> Result<()> {
        let response = self
            .build_request(reqwest::Method::PUT, "/password/reset")
            .json(&reset_data)
            .send()
            .await?;

        self.handle_empty_response(response).await
    }

    // Email Verification

    // Send verification email (POST /users/send-verification)
    pub async fn send_verification(&self) -> Result<()> {
        let response = self
            .build_request(reqwest::Method::POST, "/users/send-verification")
            .send()
            .await?;

        self.handle_empty_response(response).await
    }

    // Verify email (GET /verify-email)
    pub async fn verify_email(&self, token: &str) -> Result<()> {
        let path = format!("/verify-email?token={}", token);
        let response = self
            .build_request(reqwest::Method::GET, &path)
            .send()
            .await?;

        self.handle_empty_response(response).await
    }

    // Health check (GET /health)
    pub async fn health(&self) -> Result<HealthRes> {
        let response = self
            .build_request(reqwest::Method::GET, "/health")
            .send()
            .await?;

        self.handle_response(response).await
    }
}

// Convenience methods
impl UsersSDK {
    // Create user with minimal required fields
    pub async fn create_simple_user(
        &self,
        first_name: &str,
        last_name: &str,
        email: &str,
        username: &str,
        password: &str,
    ) -> Result<User> {
        let user = UserReqObj {
            first_name: first_name.to_string(),
            last_name: last_name.to_string(),
            email: email.to_string(),
            tags: None,
            credentials: Credentials {
                username: username.to_string(),
                secret: Some(password.to_string()),
            },
            metadata: None,
            profile_picture: None,
            status: None,
        };
        self.create_user(user).await
    }

    // Login and get tokens
    pub async fn login(&self, username: &str, password: &str) -> Result<TokenRes> {
        let credentials = IssueToken {
            username: username.to_string(),
            password: password.to_string(),
        };
        self.issue_token(credentials).await
    }

    // Get all users with pagination
    pub async fn list_all_users(&self) -> Result<Vec<User>> {
        let mut all_users = Vec::new();
        let mut offset = 0;
        let limit = 100;

        loop {
            let params = UserListParams {
                limit: Some(limit),
                offset: Some(offset),
                ..Default::default()
            };

            let page = self.list_users(Some(params)).await?;
            let fetched = page.users.len();
            all_users.extend(page.users);

            if fetched < limit as usize {
                break;
            }
            offset += limit;
        }

        Ok(all_users)
    }

    // Check if service is healthy
    pub async fn is_healthy(&self) -> bool {
        self.health().await.is_ok()
    }

    // Update user tags helper
    pub async fn set_user_tags(&self, user_id: &str, tags: Vec<String>) -> Result<User> {
        let tag_update = UserTags { tags };
        self.update_tags(user_id, tag_update).await
    }

    // Make user admin
    pub async fn make_admin(&self, user_id: &str) -> Result<User> {
        let role = UserRole {
            role: "admin".to_string(),
        };
        self.update_role(user_id, role).await
    }

    // Make user regular user
    pub async fn make_user(&self, user_id: &str) -> Result<User> {
        let role = UserRole {
            role: "user".to_string(),
        };
        self.update_role(user_id, role).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_sdk_creation() {
        let config = UsersConfig::default();
        let sdk = UsersSDK::new(config).unwrap();
        assert!(sdk.config.base_url.contains("localhost:9002"));
    }

    #[tokio::test]
    async fn test_with_token() {
        let sdk = UsersSDK::with_base_url("http://localhost:9002")
            .unwrap()
            .with_token("test-token");

        assert!(sdk.config.token.is_some());
        assert_eq!(sdk.config.token.unwrap(), "test-token");
    }

    #[test]
    fn test_user_creation_struct() {
        let user = UserReqObj {
            first_name: "John".to_string(),
            last_name: "Doe".to_string(),
            email: "john@example.com".to_string(),
            tags: Some(vec!["test".to_string()]),
            credentials: Credentials {
                username: "john_doe".to_string(),
                secret: Some("password123".to_string()),
            },
            metadata: None,
            profile_picture: None,
            status: None,
        };

        assert_eq!(user.email, "john@example.com");
        assert_eq!(user.credentials.username, "john_doe");
    }
}
