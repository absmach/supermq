use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct User {
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<Uuid>,
    #[serde(rename = "first_name", skip_serializing_if = "Option::is_none")]
    pub first_name: Option<String>,
    #[serde(rename = "last_name", skip_serializing_if = "Option::is_none")]
    pub last_name: Option<String>,
    #[serde(rename = "email", skip_serializing_if = "Option::is_none")]
    pub email: Option<String>,
    #[serde(rename = "credentials", skip_serializing_if = "Option::is_none")]
    pub credentials: Option<UserCredentials>,
    #[serde(rename = "profile_picture", skip_serializing_if = "Option::is_none")]
    pub profile_picture: Option<String>,
    #[serde(rename = "metadata", skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
    #[serde(rename = "tags", skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    #[serde(rename = "status", skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    #[serde(rename = "created_at", skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UserCredentials {
    #[serde(rename = "username")]
    pub username: String,
    #[serde(rename = "secret", skip_serializing_if = "Option::is_none")]
    pub secret: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UserUpdate {
    #[serde(rename = "first_name")]
    pub first_name: String,
    #[serde(rename = "last_name")]
    pub last_name: String,
    #[serde(rename = "metadata")]
    pub metadata: serde_json::Value,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UserProfilePicture {
    #[serde(rename = "profile_picture")]
    pub profile_picture: String,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UserTags {
    #[serde(rename = "tags")]
    pub tags: Vec<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UserRole {
    #[serde(rename = "role")]
    pub role: Role,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd, Hash, Serialize, Deserialize)]
pub enum Role {
    #[serde(rename = "admin")]
    Admin,
    #[serde(rename = "user")]
    User,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct UsersPage {
    #[serde(rename = "users")]
    pub users: Vec<User>,
    #[serde(rename = "total")]
    pub total: i32,
    #[serde(rename = "offset")]
    pub offset: i32,
    #[serde(rename = "limit")]
    pub limit: i32,
}
