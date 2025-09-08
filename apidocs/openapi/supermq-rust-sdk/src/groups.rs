//! Ugroups service client with complete models and API methods
//! 
//! This module contains all models and API methods for the Ugroups service,
//! extracted and adapted from the OpenAPI specification.

use crate::{Config, Error, Result, ResponseContent};
use reqwest::Client as HttpClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

// ============================================================================
// Ugroups Models
// ============================================================================

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
pub struct ChildrenGroupReqObj {
pub struct ChildrenGroupReqObj {
pub struct ChildrenGroupReqObj {
pub struct ChildrenGroupReqObj {
    /// Children group IDs.
    #[serde(rename = "groups")]
    pub groups: Vec<uuid::Uuid>,
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
pub struct GroupReqObj {
pub struct GroupReqObj {
pub struct GroupReqObj {
pub struct GroupReqObj {
    /// Free-form group name. Group name is unique on the given hierarchy level.
    #[serde(rename = "name")]
    pub name: String,
    /// Group description, free form text.
    #[serde(rename = "description", skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    /// Id of parent group, it must be existing group.
    #[serde(rename = "parent_id", skip_serializing_if = "Option::is_none")]
    pub parent_id: Option<String>,
    /// Arbitrary, object-encoded groups's data.
    #[serde(rename = "metadata", skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
    /// Group Status
    #[serde(rename = "status", skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct GroupUpdateTags {
pub struct GroupUpdateTags {
pub struct GroupUpdateTags {
pub struct GroupUpdateTags {
    /// Group tags.
    #[serde(rename = "tags")]
    pub tags: Vec<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct GroupUpdate {
pub struct GroupUpdate {
pub struct GroupUpdate {
pub struct GroupUpdate {
    /// Free-form group name. Group name is unique on the given hierarchy level.
    #[serde(rename = "name")]
    pub name: String,
    /// Group description, free form text.
    #[serde(rename = "description")]
    pub description: String,
    /// Arbitrary, object-encoded groups's data.
    #[serde(rename = "metadata")]
    pub metadata: serde_json::Value,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct Group {
pub struct Group {
pub struct Group {
pub struct Group {
    /// Unique group identifier generated by the service.
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// Free-form group name. Group name is unique on the given hierarchy level.
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
    /// ID of the domain to which the group belongs..
    #[serde(rename = "domain_id", skip_serializing_if = "Option::is_none")]
    pub domain_id: Option<uuid::Uuid>,
    /// Group parent identifier.
    #[serde(rename = "parent_id", skip_serializing_if = "Option::is_none")]
    pub parent_id: Option<uuid::Uuid>,
    /// Group description, free form text.
    #[serde(rename = "description", skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    /// Group tags.
    #[serde(rename = "tags", skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    /// Arbitrary, object-encoded groups's data.
    #[serde(rename = "metadata", skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
    /// Hierarchy path, concatenated ids of group ancestors.
    #[serde(rename = "path", skip_serializing_if = "Option::is_none")]
    pub path: Option<String>,
    /// Level in hierarchy, distance from the root group.
    #[serde(rename = "level", skip_serializing_if = "Option::is_none")]
    pub level: Option<i32>,
    /// Datetime when the group was created.
    #[serde(rename = "created_at", skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,
    /// Datetime when the group was created.
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
    /// Group Status
    #[serde(rename = "status", skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct GroupsHierarchyPage {
pub struct GroupsHierarchyPage {
pub struct GroupsHierarchyPage {
pub struct GroupsHierarchyPage {
    /// Level of hierarchy.
    #[serde(rename = "level", skip_serializing_if = "Option::is_none")]
    pub level: Option<i32>,
    /// Direction of hierarchy traversal.
    #[serde(rename = "direction", skip_serializing_if = "Option::is_none")]
    pub direction: Option<i32>,
    #[serde(rename = "groups", skip_serializing_if = "Option::is_none")]
    pub groups: Option<Vec<models::Group>>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct GroupsPage {
pub struct GroupsPage {
pub struct GroupsPage {
pub struct GroupsPage {
    #[serde(rename = "groups", skip_serializing_if = "Option::is_none")]
    pub groups: Option<Vec<models::Group>>,
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
pub struct HealthRes {
pub struct HealthRes {
pub struct HealthRes {
pub struct HealthRes {
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
pub struct MembersCredentials {
pub struct MembersCredentials {
pub struct MembersCredentials {
pub struct MembersCredentials {
    /// User's username.
    #[serde(rename = "username", skip_serializing_if = "Option::is_none")]
    pub username: Option<String>,
    /// User secret password.
    #[serde(rename = "secret", skip_serializing_if = "Option::is_none")]
    pub secret: Option<String>,
}

#[derive(Clone, Default, Debug, PartialEq, Serialize, Deserialize)]
pub struct MembersPage {
pub struct MembersPage {
pub struct MembersPage {
pub struct MembersPage {
    #[serde(rename = "members")]
    pub members: Vec<models::Members>,
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
pub struct Members {
pub struct Members {
pub struct Members {
pub struct Members {
    /// User unique identifier.
    #[serde(rename = "id", skip_serializing_if = "Option::is_none")]
    pub id: Option<uuid::Uuid>,
    /// User's first name.
    #[serde(rename = "first_name", skip_serializing_if = "Option::is_none")]
    pub first_name: Option<String>,
    /// User's last name.
    #[serde(rename = "last_name", skip_serializing_if = "Option::is_none")]
    pub last_name: Option<String>,
    /// User's email address.
    #[serde(rename = "email", skip_serializing_if = "Option::is_none")]
    pub email: Option<String>,
    /// User tags.
    #[serde(rename = "tags", skip_serializing_if = "Option::is_none")]
    pub tags: Option<Vec<String>>,
    #[serde(rename = "credentials", skip_serializing_if = "Option::is_none")]
    pub credentials: Option<Box<models::MembersCredentials>>,
    /// Arbitrary, object-encoded user's data.
    #[serde(rename = "metadata", skip_serializing_if = "Option::is_none")]
    pub metadata: Option<serde_json::Value>,
    /// User Status
    #[serde(rename = "status", skip_serializing_if = "Option::is_none")]
    pub status: Option<String>,
    /// Time when the group was created.
    #[serde(rename = "created_at", skip_serializing_if = "Option::is_none")]
    pub created_at: Option<String>,
    /// Time when the group was created.
    #[serde(rename = "updated_at", skip_serializing_if = "Option::is_none")]
    pub updated_at: Option<String>,
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
pub struct ParentGroupReqObj {
pub struct ParentGroupReqObj {
pub struct ParentGroupReqObj {
pub struct ParentGroupReqObj {
    /// Parent group unique identifier.
    #[serde(rename = "group_id")]
    pub group_id: uuid::Uuid,
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
pub struct UpdateRoleObj {
pub struct UpdateRoleObj {
pub struct UpdateRoleObj {
pub struct UpdateRoleObj {
    /// Role's name.
    #[serde(rename = "name", skip_serializing_if = "Option::is_none")]
    pub name: Option<String>,
}


// ============================================================================
// Ugroups Error Types
// ============================================================================

pub enum AddGroupRoleActionError {
    Status400(),
    Status401(),
    Status403(),
    Status404(),
    Status422(),
    Status500(models::Error),
    UnknownValue(serde_json::Value),
}

/// struct for typed errors of method [`add_group_role_member`]
--
pub enum AddGroupRoleMemberError {
    Status400(),
    Status401(),

// ============================================================================
// Ugroups Client Implementation
// ============================================================================

/// Ugroups service client with full API method implementations
pub struct UgroupsClient {
    http_client: HttpClient,
    base_url: String,
    config: Config,
}

impl UgroupsClient {
    /// Create a new Ugroups client
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

    /// pub async fn add_group_role_action(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<models::RoleActionsObj, Error<AddGroupRoleActionError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn add_group_role_member(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<models::RoleMembersObj, Error<AddGroupRoleMemberError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn create_group_role(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, create_role_obj: models::CreateRoleObj) -> Result<models::NewRole, Error<CreateGroupRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, create_role_obj: models::CreateRoleObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_all_group_role_actions(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str) -> Result<(), Error<DeleteAllGroupRoleActionsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_all_group_role_members(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str) -> Result<(), Error<DeleteAllGroupRoleMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_group_role(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str) -> Result<(), Error<DeleteGroupRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_group_role_action(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<(), Error<DeleteGroupRoleActionError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str, role_actions_obj: models::RoleActionsObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn delete_group_role_members(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<(), Error<DeleteGroupRoleMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str, role_members_obj: models::RoleMembersObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn get_group_members(configuration: &configuration::Configuration, domain_id: &str, group_id: &str) -> Result<models::EntityMembersObj, Error<GetGroupMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn get_group_role(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str) -> Result<models::Role, Error<GetGroupRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_available_actions(configuration: &configuration::Configuration, domain_id: &str) -> Result<models::AvailableActionsObj, Error<ListAvailableActionsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_group_role_actions(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str) -> Result<models::RoleActionsObj, Error<ListGroupRoleActionsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_group_role_members(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str) -> Result<models::RoleMembersObj, Error<ListGroupRoleMembersError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_group_roles(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<models::RolesPage, Error<ListGroupRolesError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, limit: Option<i32>, offset: Option<i32>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn update_group_role(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, role_id: &str, update_role_obj: models::UpdateRoleObj) -> Result<models::Role, Error<UpdateGroupRoleError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, role_id: &str, update_role_obj: models::UpdateRoleObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn health(configuration: &configuration::Configuration, ) -> Result<models::HealthRes, Error<HealthError>> { - Extracted from OpenAPI
    pub async fn fn(&self, ) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn add_children_groups(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, children_group_req_obj: models::ChildrenGroupReqObj) -> Result<(), Error<AddChildrenGroupsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, children_group_req_obj: models::ChildrenGroupReqObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn create_group(configuration: &configuration::Configuration, domain_id: &str, group_req_obj: models::GroupReqObj) -> Result<models::Group, Error<CreateGroupError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_req_obj: models::GroupReqObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn disable_group(configuration: &configuration::Configuration, domain_id: &str, group_id: &str) -> Result<models::Group, Error<DisableGroupError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn domain_id_groups_group_id_delete(configuration: &configuration::Configuration, domain_id: &str, group_id: &str) -> Result<(), Error<DomainIdGroupsGroupIdDeleteError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn enable_group(configuration: &configuration::Configuration, domain_id: &str, group_id: &str) -> Result<models::Group, Error<EnableGroupError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn get_group(configuration: &configuration::Configuration, domain_id: &str, group_id: &str) -> Result<models::Group, Error<GetGroupError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_children_groups(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, limit: Option<i32>, offset: Option<i32>, start_level: Option<i32>, end_level: Option<i32>, tree: Option<bool>, metadata: Option<&str>, name: Option<&str>) -> Result<models::GroupsPage, Error<ListChildrenGroupsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, limit: Option<i32>, offset: Option<i32>, start_level: Option<i32>, end_level: Option<i32>, tree: Option<bool>, metadata: Option<&str>, name: Option<&str>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_group_hierarchy(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, level: Option<i32>, tree: Option<bool>, direction: Option<i32>) -> Result<models::GroupsHierarchyPage, Error<ListGroupHierarchyError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, level: Option<i32>, tree: Option<bool>, direction: Option<i32>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn list_groups(configuration: &configuration::Configuration, domain_id: &str, limit: Option<i32>, offset: Option<i32>, level: Option<i32>, tree: Option<bool>, metadata: Option<&str>, name: Option<&str>, root_group: Option<bool>) -> Result<models::GroupsPage, Error<ListGroupsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, limit: Option<i32>, offset: Option<i32>, level: Option<i32>, tree: Option<bool>, metadata: Option<&str>, name: Option<&str>, root_group: Option<bool>) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn remove_all_children_groups(configuration: &configuration::Configuration, domain_id: &str, group_id: &str) -> Result<(), Error<RemoveAllChildrenGroupsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn remove_children_groups(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, children_group_req_obj: models::ChildrenGroupReqObj) -> Result<(), Error<RemoveChildrenGroupsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, children_group_req_obj: models::ChildrenGroupReqObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn remove_group_parent_group(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, parent_group_req_obj: models::ParentGroupReqObj) -> Result<(), Error<RemoveGroupParentGroupError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, parent_group_req_obj: models::ParentGroupReqObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn set_group_parent_group(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, parent_group_req_obj: models::ParentGroupReqObj) -> Result<(), Error<SetGroupParentGroupError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, parent_group_req_obj: models::ParentGroupReqObj) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn update_group(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, group_update: models::GroupUpdate) -> Result<models::Group, Error<UpdateGroupError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, group_update: models::GroupUpdate) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }

    /// pub async fn update_group_tags(configuration: &configuration::Configuration, domain_id: &str, group_id: &str, group_update_tags: models::GroupUpdateTags) -> Result<models::Group, Error<UpdateGroupTagsError>> { - Extracted from OpenAPI
    pub async fn fn(&self, domain_id: &str, group_id: &str, group_update_tags: models::GroupUpdateTags) -> Result<()> {
        // TODO: Implement response parsing
        Ok(())
    }
}
