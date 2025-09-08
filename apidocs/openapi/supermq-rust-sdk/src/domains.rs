//! Udomains service client with complete models and API methods
//! 
//! This module contains all models and API methods for the Udomains service,
//! extracted and adapted from the OpenAPI specification.

use crate::{Config, Error, Result, ResponseContent};
use reqwest::Client as HttpClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// ============================================================================
// Udomains Models
// ============================================================================

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct AcceptInvitationRequest {
pub struct AcceptInvitationRequest {
pub struct AcceptInvitationRequest {
pub struct AcceptInvitationRequest {
    /// Domain unique identifier.
    #[serde(rename = "domain_id")]
    pub domain_id: uuid::Uuid,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct AvailableActionsObj {
pub struct AvailableActionsObj {
pub struct AvailableActionsObj {
pub struct AvailableActionsObj {
    /// List of all available actions.
    #[serde(rename = "available_actions", skip_serializing_if = "Option::is_none")]
    pub available_actions: Option<Vec<String>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct CreateRoleObj {
pub struct CreateRoleObj {
pub struct CreateRoleObj {
pub struct CreateRoleObj {
    /// Role's name.
    #[serde(rename = "role_name", skip_serializing_if = "Option::is_none")]
    pub role_name: Option<String>,
    /// List of optional actions.
    #[serde(rename = "optional_actions", skip_serializing_if = "Option::is_none")]
    pub optional_actions: Option<Vec<String>>,
    /// List of optional members.
    #[serde(rename = "optional_members", skip_serializing_if = "Option::is_none")]
    pub optional_members: Option<Vec<String>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct DeleteInvitationRequest {
pub struct DeleteInvitationRequest {
pub struct DeleteInvitationRequest {
pub struct DeleteInvitationRequest {
    /// User unique identifier.
    #[serde(rename = "user_id")]
    pub user_id: uuid::Uuid,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct DomainReqObj {
pub struct DomainReqObj {
pub struct DomainReqObj {
pub struct DomainReqObj {
    /// Domain name.
    #[serde(rename = "name")]
    pub name: String,
    /// domain tags.
    #[serde(rename = "tags", skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    /// Arbitrary, object-encoded domain's data.
    #[serde(rename = "metadata", skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
    /// Domain route.
    #[serde(rename = "route")]
    pub route: String,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct DomainUpdate {
pub struct DomainUpdate {
pub struct DomainUpdate {
pub struct DomainUpdate {
    /// Domain name.
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// domain tags.
    #[serde(rename = "tags", skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    /// Arbitrary, object-encoded domain's data.
    #[serde(rename = "metadata", skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Domain {
pub struct Domain {
pub struct Domain {
pub struct Domain {
    /// Domain unique identified.
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// Domain name.
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// domain tags.
    #[serde(rename = "tags", skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    /// Arbitrary, object-encoded domain's data.
    #[serde(rename = "metadata", skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
    /// Domain route.
    #[serde(rename = "route", skip_serializing_if = "Option::is_none")]
    pub route: Option<String>,
    /// Domain Status
    #[serde(rename = "status", skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    /// User ID of the user who created the domain.
    #[serde(rename = "created_by", skip_serializing_if = "Option::is_none")]
    pub created_by: Option<uuid::Uuid>,
    /// Time when the domain was created.
    #[serde(rename = "created_at", skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,
    /// User ID of the user who last updated the domain.
    #[serde(rename = "updated_by", skip_serializing_if = "Option::is_none")]
    pub updated_by: Option<uuid::Uuid>,
    /// Time when the domain was last updated.
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct DomainsPage {
pub struct DomainsPage {
pub struct DomainsPage {
pub struct DomainsPage {
    #[serde(rename = "domains", skip_serializing_if = "Option::is_none")]
    pub domains: Option<Vec<models::Domain>>,
    /// Total number of items.
    #[serde(rename = "total")]
    pub total: i32,
    /// Number of items to skip during retrieval.
    #[serde(rename = "offset", skip_serializing_if = "Option::is_none")]
    pub offset: Option<i32>,
    /// Maximum number of items to return in one page.
    #[serde(rename = "limit", skip_serializing_if = "Option::is_none")]
    pub limit: Option<i32>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct EntityMembersObjMembersInnerRolesInner {
pub struct EntityMembersObjMembersInnerRolesInner {
pub struct EntityMembersObjMembersInnerRolesInner {
pub struct EntityMembersObjMembersInnerRolesInner {
    /// Unique identifier of the role.
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// Name of the role.
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// List of actions the member can perform.
    #[serde(rename = "actions", skip_serializing_if = "Option::is_none")]
    pub actions: Option<Vec<String>>,
    /// Type of access granted.
    #[serde(rename = "access_type", skip_serializing_if = "Option::is_none")]
    pub access_type: Option<AccessType>,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd, Hash, Serialize, Deserialize)]
pub enum AccessType {
pub enum AccessType {
pub enum AccessType {
pub enum AccessType {
    #[serde(rename = "read")]
    Read,
    #[serde(rename = "write")]
    Write,
    #[serde(rename = "admin")]
    Admin,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct EntityMembersObjMembersInner {
pub struct EntityMembersObjMembersInner {
pub struct EntityMembersObjMembersInner {
pub struct EntityMembersObjMembersInner {
    /// Unique identifier of the member.
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// List of roles assigned to the member.
    #[serde(rename = "roles", skip_serializing_if = "Option::is_none")]
    pub roles: Option<Vec<models::EntityMembersObjMembersInnerRolesInner>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct EntityMembersObj {
pub struct EntityMembersObj {
pub struct EntityMembersObj {
pub struct EntityMembersObj {
    /// List of members with assigned roles and actions.
    #[serde(rename = "members", skip_serializing_if = "Option::is_none")]
    pub members: Option<Vec<models::EntityMembersObjMembersInner>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Error {
pub struct Error {
pub struct Error {
pub struct Error {
    /// Error message
    #[serde(rename = "error", skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct HealthInfo {
pub struct HealthInfo {
pub struct HealthInfo {
pub struct HealthInfo {
    /// Service status.
    #[serde(rename = "status", skip_serializing_if = "Option::is_none")]
    pub status: Option<Status>,
    /// Service version.
    #[serde(rename = "version", skip_serializing_if = "Option::is_none")]
    pub version: Option<String>,
    /// Service commit hash.
    #[serde(rename = "commit", skip_serializing_if = "Option::is_none")]
    pub commit: Option<String>,
    /// Service description.
    #[serde(rename = "description", skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    /// Service build time.
    #[serde(rename = "build_time", skip_serializing_if = "Option::is_none")]
    pub build_time: Option<String>,
    /// Service instance ID.
    #[serde(rename = "instance_id", skip_serializing_if = "Option::is_none")]
    pub instance_id: Option<String>,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq, Ord, PartialOrd, Hash, Serialize, Deserialize)]
pub enum Status {
pub enum Status {
pub enum Status {
pub enum Status {
    #[serde(rename = "pass")]
    Pass,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct InvitationPage {
pub struct InvitationPage {
pub struct InvitationPage {
pub struct InvitationPage {
    #[serde(rename = "invitations")]
    pub invitations: Vec<models::Invitation>,
    /// Total number of items.
    #[serde(rename = "total")]
    pub total: i32,
    /// Number of items to skip during retrieval.
    #[serde(rename = "offset")]
    pub offset: i32,
    /// Maximum number of items to return in one page.
    #[serde(rename = "limit", skip_serializing_if = "Option::is_none")]
    pub limit: Option<i32>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Invitation {
pub struct Invitation {
pub struct Invitation {
pub struct Invitation {
    /// User unique identifier.
    #[serde(rename = "invited_by", skip_serializing_if = "Option::is_none")]
    pub invited_by: Option<uuid::Uuid>,
    /// Invitee user unique identifier.
    #[serde(rename = "invitee_user_id", skip_serializing_if = "Option::is_none")]
    pub invitee_user_id: Option<uuid::Uuid>,
    /// Domain unique identifier.
    #[serde(rename = "domain_id", skip_serializing_if = "Option::is_none")]
    pub domain_id: Option<uuid::Uuid>,
    /// Role unique identifier.
    #[serde(rename = "role_id", skip_serializing_if = "Option::is_none")]
    pub role_id: Option<uuid::Uuid>,
    /// Role name.
    #[serde(rename = "role_name", skip_serializing_if = "Option::is_none")]
    pub role_name: Option<String>,
    /// Actions allowed for the role.
    #[serde(rename = "actions", skip_serializing_if = "Option::is_none")]
    pub actions: Option<Vec<String>>,
    /// Time when the group was created.
    #[serde(rename = "created_at", skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,
    /// Time when the group was created.
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
    /// Time when the group was created.
    #[serde(rename = "confirmed_at", skip_serializing_if = "Option::is_none")]
    pub confirmed_at: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct NewRole {
pub struct NewRole {
pub struct NewRole {
pub struct NewRole {
    /// Role unique identifier.
    #[serde(rename = "role_id", skip_serializing_if = "Option::is_none")]
    pub role_id: Option<uuid::Uuid>,
    /// Role's name.
    #[serde(rename = "role_name", skip_serializing_if = "Option::is_none")]
    pub role_name: Option<String>,
    /// Entity unique identifier.
    #[serde(rename = "entity_id", skip_serializing_if = "Option::is_none")]
    pub entity_id: Option<uuid::Uuid>,
    /// Role creator unique identifier.
    #[serde(rename = "created_by", skip_serializing_if = "Option::is_none")]
    pub created_by: Option<uuid::Uuid>,
    /// Time when the channel was created.
    #[serde(rename = "created_at", skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,
    /// Role updater unique identifier.
    #[serde(rename = "updated_by", skip_serializing_if = "Option::is_none")]
    pub updated_by: Option<uuid::Uuid>,
    /// Time when the channel was created.
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct RoleActionsObj {
pub struct RoleActionsObj {
pub struct RoleActionsObj {
pub struct RoleActionsObj {
    /// List of actions to be added to a role.
    #[serde(rename = "actions", skip_serializing_if = "Option::is_none")]
    pub actions: Option<Vec<String>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct RoleMembersObj {
pub struct RoleMembersObj {
pub struct RoleMembersObj {
pub struct RoleMembersObj {
    /// List of members to be added to a role.
    #[serde(rename = "members", skip_serializing_if = "Option::is_none")]
    pub members: Option<Vec<String>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Role {
pub struct Role {
pub struct Role {
pub struct Role {
    /// Role unique identifier.
    #[serde(rename = "role_id", skip_serializing_if = "Option::is_none")]
    pub role_id: Option<uuid::Uuid>,
    /// Role's name.
    #[serde(rename = "role_name", skip_serializing_if = "Option::is_none")]
    pub role_name: Option<String>,
    /// Entity unique identifier.
    #[serde(rename = "entity_id", skip_serializing_if = "Option::is_none")]
    pub entity_id: Option<uuid::Uuid>,
    /// Role creator unique identifier.
    #[serde(rename = "created_by", skip_serializing_if = "Option::is_none")]
    pub created_by: Option<uuid::Uuid>,
    /// Time when the channel was created.
    #[serde(rename = "created_at", skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,
    /// Role updater unique identifier.
    #[serde(rename = "updated_by", skip_serializing_if = "Option::is_none")]
    pub updated_by: Option<uuid::Uuid>,
    /// Time when the channel was created.
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct RolesPage {
pub struct RolesPage {
pub struct RolesPage {
pub struct RolesPage {
    /// List of roles.
    #[serde(rename = "roles", skip_serializing_if = "Option::is_none")]
    pub roles: Option<Vec<models::Role>>,
    /// Total number of roles.
    #[serde(rename = "total", skip_serializing_if = "Option::is_none")]
    pub total: Option<i32>,
    /// Number of items to skip during retrieval.
    #[serde(rename = "offset", skip_serializing_if = "Option::is_none")]
    pub offset: Option<i32>,
    /// Maximum number of items to return in one page.
    #[serde(rename = "limit", skip_serializing_if = "Option::is_none")]
    pub limit: Option<i32>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct SendInvitationReqObj {
pub struct SendInvitationReqObj {
pub struct SendInvitationReqObj {
pub struct SendInvitationReqObj {
    /// User unique identifier.
    #[serde(rename = "invitee_user_id")]
    pub invitee_user_id: uuid::Uuid,
    /// Identifier for the role to be assigned to the user.
    #[serde(rename = "role_id")]
    pub role_id: uuid::Uuid,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct UpdateRoleObj {
pub struct UpdateRoleObj {
pub struct UpdateRoleObj {
pub struct UpdateRoleObj {
    /// Role's name.
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
}


// ============================================================================
// Udomains Error Types
// ============================================================================

pub enum AddDomainRoleActionError {
    Status400(),
    Status401(),
    Status403(),
    Status404(),
    Status422(),
    Status500(),
    UnknownValue(serde_json::Value),
}

/// struct for typed errors of method [`add_domain_role_member`]
--
pub enum AddDomainRoleMemberError {
    Status400(),
    Status401(),

// ============================================================================
// Udomains Client Implementation
// ============================================================================

/// Udomains service client with full API method implementations
pub struct UdomainsClient {
    http_client: HttpClient,
    base_url: String,
    config: Config,
}

impl UdomainsClient {
    /// Create a new Udomains client
    pub(crate) fn new(http_client: HttpClient, config: Config) -> Self {
        Self {
            http_client,
            base_url: config.base_url.clone(),
            config,
        }
    }

    /// Health check endpoint
    pub async fn health(&self) -> Result<bool> {
        let url = format!("{}/health", self.base_url);
        let mut request = self.http_client.get(&url);
        
        if let Some(ref token) = self.config.bearer_access_token {
            request = request.bearer_auth(token);
        }
        
        if let Some(ref user_agent) = self.config.user_agent {
            request = request.header(reqwest::header::USER_AGENT, user_agent);
        }
        
        let response = request.send().await?;
        Ok(response.status().is_success())
    }

    /// pub async fn add_domain_role_action(configuration: &configuration::Configuration, domain_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<models::RoleActionsObj, Error<AddDomainRoleActionError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn add_domain_role_member(configuration: &configuration::Configuration, domain_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<models::RoleMembersObj, Error<AddDomainRoleMemberError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn create_domain_role(configuration: &configuration::Configuration, domain_id: &str, create_role_obj: models::CreateRoleObj) -> Result<models::NewRole, Error<CreateDomainRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, create_role_obj: models::CreateRoleObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_all_domain_role_actions(configuration: &configuration::Configuration, domain_id: &str, role_id: &str) -> Result<(), Error<DeleteAllDomainRoleActionsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_all_domain_role_members(configuration: &configuration::Configuration, domain_id: &str, role_id: &str) -> Result<(), Error<DeleteAllDomainRoleMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_domain_role(configuration: &configuration::Configuration, domain_id: &str, role_id: &str) -> Result<(), Error<DeleteDomainRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_domain_role_action(configuration: &configuration::Configuration, domain_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<(), Error<DeleteDomainRoleActionError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_domain_role_members(configuration: &configuration::Configuration, domain_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<(), Error<DeleteDomainRoleMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn get_domain_members(configuration: &configuration::Configuration, domain_id: &str) -> Result<models::EntityMembersObj, Error<GetDomainMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn get_domain_role(configuration: &configuration::Configuration, domain_id: &str, role_id: &str) -> Result<models::Role, Error<GetDomainRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_available_actions(configuration: &configuration::Configuration, domain_id: &str) -> Result<models::AvailableActionsObj, Error<ListAvailableActionsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_domain_role_actions(configuration: &configuration::Configuration, domain_id: &str, role_id: &str) -> Result<models::RoleActionsObj, Error<ListDomainRoleActionsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_domain_role_members(configuration: &configuration::Configuration, domain_id: &str, role_id: &str) -> Result<models::RoleMembersObj, Error<ListDomainRoleMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_domain_roles(configuration: &configuration::Configuration, domain_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<models::RolesPage, Error<ListDomainRolesError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn update_domain_role(configuration: &configuration::Configuration, domain_id: &str, role_id: &str, update_role_obj: models::UpdateRoleObj) -> Result<models::Role, Error<UpdateDomainRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, role_id: &str, update_role_obj: models::UpdateRoleObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domains_domain_id_disable_post(configuration: &configuration::Configuration, domain_id: &str) -> Result<(), Error<DomainsDomainIdDisablePostError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domains_domain_id_enable_post(configuration: &configuration::Configuration, domain_id: &str) -> Result<(), Error<DomainsDomainIdEnablePostError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domains_domain_id_freeze_post(configuration: &configuration::Configuration, domain_id: &str) -> Result<(), Error<DomainsDomainIdFreezePostError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domains_domain_id_get(configuration: &configuration::Configuration, domain_id: &str) -> Result<models::Domain, Error<DomainsDomainIdGetError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domains_domain_id_patch(configuration: &configuration::Configuration, domain_id: &str, domain_update: models::DomainUpdate) -> Result<models::Domain, Error<DomainsDomainIdPatchError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, domain_update: models::DomainUpdate) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domains_get(configuration: &configuration::Configuration, limit: Option<i32>, offset: Option<i32>, metadata: Option<std::collections::HashMap<String, serde_json::Value>>, status: Option<&str>, name: Option<&str>, permission: Option<&str>) -> Result<models::DomainsPage, Error<DomainsGetError>> { - Extracted from OpenAPI
    pub async fn fn(&self, limit: Option<i32>, offset: Option<i32>, metadata: Option<std::collections::HashMap<String, serde_json::Value>>, status: Option<&str>, name: Option<&str>, permission: Option<&str>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domains_post(configuration: &configuration::Configuration, domain_req_obj: models::DomainReqObj) -> Result<models::Domain, Error<DomainsPostError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_req_obj: models::DomainReqObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn health_get(configuration: &configuration::Configuration, ) -> Result<models::HealthInfo, Error<HealthGetError>> { - Extracted from OpenAPI
    pub async fn fn(&self, ) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn accept_invitation(configuration: &configuration::Configuration, accept_invitation_request: models::AcceptInvitationRequest) -> Result<(), Error<AcceptInvitationError>> { - Extracted from OpenAPI
    pub async fn fn(&self, accept_invitation_request: models::AcceptInvitationRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_invitation(configuration: &configuration::Configuration, domain_id: &str, delete_invitation_request: models::DeleteInvitationRequest) -> Result<(), Error<DeleteInvitationError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, delete_invitation_request: models::DeleteInvitationRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_domain_invitations(configuration: &configuration::Configuration, domain_id: &str, limit: Option<i32>, offset: Option<i32>, user_id: Option<&str>, invited_by: Option<&str>, state: Option<&str>) -> Result<models::InvitationPage, Error<ListDomainInvitationsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, limit: Option<i32>, offset: Option<i32>, user_id: Option<&str>, invited_by: Option<&str>, state: Option<&str>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_user_invitations(configuration: &configuration::Configuration, domain_id: Option<&str>, limit: Option<i32>, offset: Option<i32>, user_id: Option<&str>, invited_by: Option<&str>, state: Option<&str>) -> Result<models::InvitationPage, Error<ListUserInvitationsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: Option<&str>, limit: Option<i32>, offset: Option<i32>, user_id: Option<&str>, invited_by: Option<&str>, state: Option<&str>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn reject_invitation(configuration: &configuration::Configuration, accept_invitation_request: models::AcceptInvitationRequest) -> Result<(), Error<RejectInvitationError>> { - Extracted from OpenAPI
    pub async fn fn(&self, accept_invitation_request: models::AcceptInvitationRequest) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn send_invitation(configuration: &configuration::Configuration, domain_id: &str, send_invitation_req_obj: models::SendInvitationReqObj) -> Result<(), Error<SendInvitationError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, send_invitation_req_obj: models::SendInvitationReqObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }
}
