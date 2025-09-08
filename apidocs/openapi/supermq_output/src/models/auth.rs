use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Key {
    #[serde(rename = "id")]
    pub id: Uuid,
    #[serde(rename = "issuer_id")]
    pub issuer_id: Uuid,
    #[serde(rename = "type")]
    pub type_: i32,
    #[serde(rename = "subject")]
    pub subject: String,
    #[serde(rename = "issued_at")]
    pub issued_at: String,
    #[serde(rename = "expires_at", skip_serializing_if = "Option::is_none")]
    pub expires_at: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct PersonalAccessToken {
    #[serde(rename = "id")]
    pub id: Uuid,
    #[serde(rename = "user_id")]
    pub user_id: Uuid,
    #[serde(rename = "name")]
    pub name: String,
    #[serde(rename = "description", skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    #[serde(rename = "secret")]
    pub secret: String,
    #[serde(rename = "issued_at")]
    pub issued_at: String,
    #[serde(rename = "expires_at", skip_serializing_if = "Option::is_none")]
    pub expires_at: Option<String>,
    #[serde(rename = "updated_at")]
    pub updated_at: String,
    #[serde(rename = "last_used_at", skip_serializing_if = "Option::is_none")]
    pub last_used_at: Option<String>,
    #[serde(rename = "revoked")]
    pub revoked: bool,
    #[serde(rename = "revoked_at", skip_serializing_if = "Option::is_none")]
    pub revoked_at: Option<String>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Scope {
    #[serde(rename = "id")]
    pub id: Uuid,
    #[serde(rename = "pat_id")]
    pub pat_id: Uuid,
    #[serde(rename = "optional_domain_id", skip_serializing_if = "Option::is_none")]
    pub optional_domain_id: Option<Uuid>,
    #[serde(rename = "entity_type")]
    pub entity_type: EntityType,
    #[serde(rename = "entity_id")]
    pub entity_id: String,
    #[serde(rename = "operation")]
    pub operation: Operation,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum EntityType {
    #[serde(rename = "groups")]
    Groups,
    #[serde(rename = "channels")]
    Channels,
    #[serde(rename = "clients")]
    Clients,
    #[serde(rename = "domains")]
    Domains,
    #[serde(rename = "users")]
    Users,
    #[serde(rename = "dashboards")]
    Dashboards,
    #[serde(rename = "messages")]
    Messages,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum Operation {
    #[serde(rename = "create")]
    Create,
    #[serde(rename = "read")]
    Read,
    #[serde(rename = "list")]
    List,
    #[serde(rename = "update")]
    Update,
    #[serde(rename = "delete")]
    Delete,
    #[serde(rename = "share")]
    Share,
    #[serde(rename = "unshare")]
    Unshare,
    #[serde(rename = "publish")]
    Publish,
    #[serde(rename = "subscribe")]
    Subscribe,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct PatsPage {
    #[serde(rename = "total")]
    pub total: i32,
    #[serde(rename = "offset")]
    pub offset: i32,
    #[serde(rename = "limit")]
    pub limit: i32,
    #[serde(rename = "pats")]
    pub pats: Vec<PersonalAccessToken>,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct ScopesPage {
    #[serde(rename = "total")]
    pub total: i32,
    #[serde(rename = "offset")]
    pub offset: i32,
    #[serde(rename = "limit")]
    pub limit: i32,
    #[serde(rename = "scopes")]
    pub scopes: Vec<Scope>,
}
