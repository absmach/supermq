//! # SuperMQ Rust SDK
//! 
//! Complete Rust SDK for SuperMQ services with full OpenAPI-generated implementations.
//! 
//! ## Usage
//! 
//! ```rust
//! use supermq_rust_sdk::{Client, Config};
//! 
//! let config = Config::new("http://localhost:8080")
//!     .with_bearer_token("your-token-here");
//! let client = Client::new(config);
//! 
//! // Use any service
//! let users = client.users();
//! ```

pub mod client;
pub mod config;
pub mod error;
pub mod types;

pub mod auth;
pub mod users;
pub mod domains;
pub mod things;
pub mod channels;
pub mod groups;
pub mod bootstrap;
pub mod certs;
pub mod provision;
pub mod journal;

// Re-export main types
pub use client::Client;
pub use config::Config;
pub use error::{Error, Result, ResponseContent};
pub use types::*;
