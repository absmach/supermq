use crate::config::Config;
use reqwest::Client as HttpClient;

/// Main SuperMQ SDK client that provides access to all services
/// 
/// Each service client receives a copy of the configuration and HTTP client
/// for consistent behavior across all services.
pub struct Client {
    http_client: HttpClient,
    config: Config,
}

impl Client {
    /// Create a new SuperMQ client with the provided configuration
    pub fn new(config: Config) -> Self {
        let http_client = HttpClient::builder()
            .timeout(config.timeout)
            .build()
            .expect("Failed to create HTTP client");
            
        Self {
            http_client,
            config,
        }
    }

    /// Update the bearer token for authentication
    pub fn with_bearer_token(mut self, token: impl Into<String>) -> Self {
        self.config.bearer_access_token = Some(token.into());
        self
    }

    /// Get auth service client
    pub fn auth(&self) -> crate::auth::UauthClient {
        crate::auth::UauthClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get users service client
    pub fn users(&self) -> crate::users::UusersClient {
        crate::users::UusersClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get domains service client
    pub fn domains(&self) -> crate::domains::UdomainsClient {
        crate::domains::UdomainsClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get things service client
    pub fn things(&self) -> crate::things::UthingsClient {
        crate::things::UthingsClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get channels service client
    pub fn channels(&self) -> crate::channels::UchannelsClient {
        crate::channels::UchannelsClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get groups service client
    pub fn groups(&self) -> crate::groups::UgroupsClient {
        crate::groups::UgroupsClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get bootstrap service client
    pub fn bootstrap(&self) -> crate::bootstrap::UbootstrapClient {
        crate::bootstrap::UbootstrapClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get certs service client
    pub fn certs(&self) -> crate::certs::UcertsClient {
        crate::certs::UcertsClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get provision service client
    pub fn provision(&self) -> crate::provision::UprovisionClient {
        crate::provision::UprovisionClient::new(self.http_client.clone(), self.config.clone())
    }

    /// Get journal service client
    pub fn journal(&self) -> crate::journal::UjournalClient {
        crate::journal::UjournalClient::new(self.http_client.clone(), self.config.clone())
    }
}
