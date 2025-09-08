// Re-export consolidated modules
mod auth;
mod common;
mod domain;
mod identity;

// Re-export commonly used types
pub use auth::{Key, PersonalAccessToken, Scope, EntityType, Operation, PatsPage, ScopesPage};
pub use common::{Error, HealthInfo, Page};
pub use domain::{Domain, DomainUpdate, DomainsPage};
pub use identity::{User, UserCredentials, UserUpdate, UserProfilePicture, UserTags, UserRole, Role, UsersPage};
