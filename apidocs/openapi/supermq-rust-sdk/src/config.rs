#[derive(Debug, Clone)]
pub struct Config {
    pub base_url: String,
    pub timeout: std::time::Duration,
    pub bearer_access_token: Option<String>,
    pub basic_auth: Option<(String, Option<String>)>,
    pub user_agent: Option<String>,
}

impl Config {
    pub fn new(base_url: impl Into<String>) -> Self {
        Self {
            base_url: base_url.into(),
            timeout: std::time::Duration::from_secs(30),
            bearer_access_token: None,
            basic_auth: None,
            user_agent: None,
        }
    }

    pub fn with_timeout(mut self, timeout: std::time::Duration) -> Self {
        self.timeout = timeout;
        self
    }

    pub fn with_bearer_token(mut self, token: impl Into<String>) -> Self {
        self.bearer_access_token = Some(token.into());
        self
    }

    pub fn with_basic_auth(mut self, username: impl Into<String>, password: Option<String>) -> Self {
        self.basic_auth = Some((username.into(), password));
        self
    }

    pub fn with_user_agent(mut self, user_agent: impl Into<String>) -> Self {
        self.user_agent = Some(user_agent.into());
        self
    }
}
