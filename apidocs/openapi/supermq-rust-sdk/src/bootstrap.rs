//! Ubootstrap service client

use crate::{Error, Result, Config};
use reqwest::Client as HttpClient;

pub struct UbootstrapClient {
    http_client: HttpClient,
    base_url: String,
    config: Config,
}

impl UbootstrapClient {
    pub(crate) fn new(http_client: HttpClient, config: Config) -> Self {
        Self {
            http_client,
            base_url: config.base_url.clone(),
            config,
        }
    }

    pub async fn health(&self) -> Result<bool> {
        let url = format!("{}/health", self.base_url);
        let response = self.http_client.get(&url).send().await?;
        Ok(response.status().is_success())
    }
}
