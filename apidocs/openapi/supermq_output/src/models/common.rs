use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Error {
    #[serde(rename = "error")]
    pub error: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct HealthInfo {
    #[serde(rename = "status")]
    pub status: String,
    #[serde(rename = "version")]
    pub version: String,
    #[serde(rename = "commit")]
    pub commit: String,
    #[serde(rename = "description")]
    pub description: String,
    #[serde(rename = "build_time")]
    pub build_time: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Page<T> {
    pub items: Vec<T>,
    pub total: i32,
    pub offset: i32,
    pub limit: i32,
}
