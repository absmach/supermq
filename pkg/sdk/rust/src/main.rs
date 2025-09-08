use rust_demo::CertsSDK;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let sdk = CertsSDK::with_base_url("http://localhost:9019")?.with_token("your-auth-token");

    let entity_id = "entity-123";

    // Issue a new certificate
    let cert = sdk.issue_default(entity_id).await?;
    println!("Created certificate: {}", cert.serial);

    // View the certificate
    let retrieved = sdk.view(&cert.serial).await?;
    println!("Certificate expires: {}", retrieved.expires_at);

    // List all certificates
    let all = sdk.list_all().await?;
    println!("Total certificates: {}", all.len());

    // Revoke the certificate
    let revoked = sdk.revoke(&cert.serial).await?;
    println!("Certificate revoked: {}", revoked.revoked);

    // Check if revoked
    let is_revoked = sdk.is_revoked(&cert.serial).await?;
    println!("Is revoked? {}", is_revoked);

    Ok(())
}

// Example usage
/*
#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Create SDK instance
    let sdk = UsersSDK::with_base_url("http://localhost:9002")?;

    // Create a new user
    let user = sdk.create_simple_user(
        "John",
        "Doe",
        "john@example.com",
        "john_doe",
        "password123"
    ).await?;
    println!("Created user: {}", user.id);

    // Login to get tokens
    let tokens = sdk.login("john_doe", "password123").await?;
    println!("Got access token: {}", tokens.access_token);

    // Create authenticated SDK instance
    let auth_sdk = UsersSDK::with_base_url("http://localhost:9002")?
        .with_token(&tokens.access_token);

    // Get user profile
    let profile = auth_sdk.get_profile().await?;
    println!("Profile: {} {}", profile.first_name, profile.last_name);

    // List users
    let users_page = auth_sdk.list_users(None).await?;
    println!("Found {} users", users_page.users.len());

    // Update user
    let update = UserUpdate {
        first_name: "Jane".to_string(),
        last_name: "Doe".to_string(),
        metadata: HashMap::new(),
    };
    let updated_user = auth_sdk.update_user(&user.id, update).await?;
    println!("Updated user name: {}", updated_user.first_name);

    // Make user admin
    let admin_user = auth_sdk.make_admin(&user.id).await?;
    println!("User is now admin");

    // Check health
    if auth_sdk.is_healthy().await {
        println!("Service is healthy");
    }

    Ok(())
}
*/

/*use std::collections::HashMap;
use supermq_channels_sdk::*;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Create client
    let client = ChannelsClient::new("http://localhost:9005")?
        .with_token("your-jwt-token".to_string());

    // Create a channel
    let mut metadata = HashMap::new();
    metadata.insert("location".to_string(), serde_json::Value::String("datacenter-1".to_string()));

    let create_request = ChannelCreateRequest::new("sensor-data".to_string())
        .with_route("sensors/temperature")
        .with_metadata(metadata)
        .with_status("enabled".to_string());

    let channel = client.create_channel("domain-123", create_request).await?;
    println!("Created channel: {}", channel.id);

    // List channels with pagination
    let params = ListChannelsParams::new()
        .with_limit(10)
        .with_offset(0)
        .with_name("sensor");

    let channels_page = client.list_channels("domain-123", Some(params)).await?;
    println!("Found {} channels", channels_page.total);

    // Connect clients to a channel
    let connection_request = ChannelConnectionRequest::new(
        vec!["client-1".to_string(), "client-2".to_string()]
    ).with_types(vec!["publish".to_string(), "subscribe".to_string()]);

    client.connect_clients_to_channel("domain-123", &channel.id, connection_request).await?;

    Ok(())
} */

// Example usage:
/*
use uuid::Uuid;
use std::collections::HashMap;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Create client
    let mut client = ClientsServiceClient::new("http://localhost:9006".to_string());
    client.set_auth_token("your-jwt-token".to_string());

    let domain_id = Uuid::parse_str("123e4567-e89b-12d3-a456-426614174000")?;

    // Create a new client
    let new_client_request = ClientRequest::new(
        "device001@iot.example.com".to_string(),
        "super-secret-password".to_string(),
    )
    .with_name("IoT Device 001".to_string())
    .with_tags(vec!["iot".to_string(), "sensor".to_string(), "production".to_string()])
    .with_status("enabled".to_string());

    let created_client = client.create_client(domain_id, new_client_request).await?;
    println!("Created client with ID: {}", created_client.id);

    // List all clients with pagination
    let list_params = ClientListParams::new()
        .with_limit(50)
        .with_offset(0)
        .with_status("enabled".to_string());

    let clients_page = client.list_clients(domain_id, Some(list_params)).await?;
    println!("Found {} clients", clients_page.total);

    // Get a specific client
    let retrieved_client = client.get_client(domain_id, created_client.id).await?;
    println!("Retrieved client: {:?}", retrieved_client.name);

    // Update client metadata
    let mut metadata = HashMap::new();
    metadata.insert("location".to_string(), serde_json::Value::String("warehouse-a".to_string()));
    metadata.insert("firmware_version".to_string(), serde_json::Value::String("1.2.3".to_string()));

    let update = ClientUpdate {
        name: "IoT Device 001 - Updated".to_string(),
        metadata,
    };

    let updated_client = client.update_client(domain_id, created_client.id, update).await?;
    println!("Updated client name: {:?}", updated_client.name);

    // Update client tags
    let new_tags = ClientTags {
        tags: vec!["iot".to_string(), "sensor".to_string(), "warehouse-a".to_string()],
    };
    client.update_client_tags(domain_id, created_client.id, new_tags).await?;

    // Create a role for the client
    let role_request = CreateRoleRequest {
        name: "device_operator".to_string(),
        description: Some("Allows device operation commands".to_string()),
        actions: vec!["read".to_string(), "write".to_string()],
    };

    let created_role = client.create_client_role(domain_id, created_client.id, role_request).await?;
    println!("Created role: {}", created_role.name);

    // Add actions to the role
    let actions_request = RoleActionsRequest {
        actions: vec!["delete".to_string(), "update".to_string()],
    };
    client.add_client_role_actions(domain_id, created_client.id, &created_role.id, actions_request).await?;

    // Add members to the role
    let members_request = RoleMembersRequest {
        members: vec!["user123".to_string(), "user456".to_string()],
    };
    client.add_client_role_members(domain_id, created_client.id, &created_role.id, members_request).await?;

    // List all available actions
    let available_actions = client.list_available_actions(domain_id).await?;
    println!("Available actions: {:?}", available_actions);

    // Bulk create multiple clients
    let bulk_clients = vec![
        ClientRequest::new("device002@iot.example.com".to_string(), "secret2".to_string())
            .with_name("IoT Device 002".to_string()),
        ClientRequest::new("device003@iot.example.com".to_string(), "secret3".to_string())
            .with_name("IoT Device 003".to_string()),
    ];

    let bulk_result = client.bulk_create_clients(domain_id, bulk_clients).await?;
    println!("Bulk created {} clients", bulk_result.clients.len());

    // Set parent group for a client
    let parent_group_id = Uuid::parse_str("456e7890-e89b-12d3-a456-426614174001")?;
    let parent_group_req = ParentGroupRequest {
        parent_group_id,
    };
    client.set_client_parent_group(domain_id, created_client.id, parent_group_req).await?;
    println!("Set parent group for client");

    // Disable a client
    let disabled_client = client.disable_client(domain_id, created_client.id).await?;
    println!("Client disabled, status: {}", disabled_client.status);

    // Enable the client again
    let enabled_client = client.enable_client(domain_id, created_client.id).await?;
    println!("Client enabled, status: {}", enabled_client.status);

    // Health check
    let health = client.health().await?;
    println!("Service status: {} (version: {})", health.status, health.version);

    Ok(())
}
*/

// Example usage
/*
#[cfg(feature = "examples")]
pub mod examples {
    use super::*;

    pub async fn example_usage() -> Result<()> {
        // Create a client
        let mut client =
            JournalClient::new("http://localhost:9021").with_token("your-jwt-token-here");

        let user_id = Uuid::new_v4();
        let domain_id = Uuid::new_v4();
        let client_id = Uuid::new_v4();

        // Get user journals with pagination
        let query = JournalQuery::new()
            .limit(5)
            .with_attributes(true)
            .with_metadata(true)
            .sort_direction(SortDirection::Desc);

        let journals = client.list_user_journal(user_id, Some(query)).await?;
        println!("Found {} journals", journals.journals.len());

        // Get client telemetry
        let telemetry = client.get_client_telemetry(domain_id, client_id).await?;
        println!("Client telemetry: {:?}", telemetry);

        // Get entity journals
        let entity_journals = client
            .list_entity_journal(
                domain_id,
                EntityType::Client,
                client_id,
                Some(JournalQuery::new().with_attributes(true)),
            )
            .await?;
        println!("Found {} entity journals", entity_journals.journals.len());

        // Search by operation
        let create_journals = client
            .search_user_journals_by_operation(user_id, "user.create")
            .await?;
        println!("Found {} create operations", create_journals.len());

        // Health check
        let health = client.health_check().await?;
        println!("Service health: {:?}", health);

        Ok(())
    }
}
    */

// Example usage
/* #[cfg(feature = "examples")]
pub mod examples {
    use super::*;

    pub async fn example_usage() -> Result<()> {
        // Create a client
        let client = TwinsClient::new("http://localhost:9018")
            .with_token("your-jwt-token-here");

        // Create a simple twin
        let location = client.create_simple_twin("my-sensor").await?;
        println!("Created twin at: {}", location);

        // Create a twin with metadata and definition
        let mut metadata = HashMap::new();
        metadata.insert("type".to_string(), serde_json::Value::String("temperature_sensor".to_string()));
        metadata.insert("location".to_string(), serde_json::Value::String("living_room".to_string()));

        let attribute = Attribute::new("temperature", "temp_channel", "sensors/temp", true);
        let definition = Definition::new(1.0).add_attribute(attribute);

        let twin_req = TwinRequest::new("smart-thermostat")
            .with_metadata(metadata)
            .with_definition(definition);

        let location = client.create_twin(twin_req).await?;
        println!("Created advanced twin at: {}", location);

        // Get twins with pagination
        let query = TwinsQuery::new()
            .limit(10)
            .name("sensor");

        let twins_page = client.get_twins(Some(query)).await?;
        println!("Found {} twins", twins_page.twins.len());

        // Get a specific twin
        if let Some(first_twin) = twins_page.twins.first() {
            let twin = client.get_twin(first_twin.id).await?;
            println!("Twin: {} (revision: {})", twin.name, twin.revision);

            // Get states for the twin
            let states_query = StatesQuery::new().limit(5);
            let states_page = client.get_states(twin.id, Some(states_query)).await?;
            println!("Found {} states for twin", states_page.states.len());

            // Update twin name
            client.update_twin_name(twin.id, "updated-name").await?;
            println!("Updated twin name");
        }

        // Search twins by metadata
        let mut search_metadata = HashMap::new();
        search_metadata.insert("type".to_string(), serde_json::Value::String("sensor".to_string()));

        let sensor_twins = client.search_twins_by_metadata(search_metadata).await?;
        println!("Found {} sensor twins", sensor_twins.len());

        // Get all twins (with automatic pagination)
        let all_twins = client.get_all_twins().await?;
        println!("Total twins: {}", all_twins.len());

        // Health check
        let health = client.health_check().await?;
        println!("Service health: {:?}", health);

        Ok(())
    }
}
*/

/*
// Create client
let client = DomainsClient::new("http://localhost:9003")
    .with_token("your-token");

// Create domain
let domain = client.create_domain(
    CreateDomainRequest::new("my-domain", "my-route")
        .with_tags(vec!["tag1".to_string()])
).await?;

// List domains with filtering
let domains = client.list_domains(ListDomainsParams {
    limit: Some(50),
    status: Some("enabled".to_string()),
    ..Default::default()
}).await?;
*/

/*
// Example usage
#[cfg(feature = "examples")]
mod examples {
    use super::*;

    /// Example: Basic group operations
    pub async fn basic_group_operations() -> Result<(), GroupsError> {
        // Create client
        let client = GroupsClient::with_token("your-jwt-token")?;
        let domain_id = Uuid::new_v4();

        // Create a group
        let create_request = CreateGroupRequest::new("Engineering Team")
            .description("Software engineering department")
            .status(GroupStatus::Enabled);

        let group = client.create_group(&domain_id, create_request).await?;
        println!("Created group: {:?}", group);

        // List all groups
        let groups_page = client
            .list_groups(&domain_id, Some(ListGroupsParams::new().limit(10)))
            .await?;
        println!("Groups count: {}", groups_page.total);

        // Get specific group
        let retrieved_group = client.get_group(&domain_id, &group.id).await?;
        println!("Retrieved group: {:?}", retrieved_group);

        // Update group
        let mut metadata = HashMap::new();
        metadata.insert("department".to_string(), serde_json::Value::String("IT".to_string()));

        let update_request = UpdateGroupRequest::new(
            "Updated Engineering Team",
            "Updated description",
            metadata,
        );

        let updated_group = client
            .update_group(&domain_id, &group.id, update_request)
            .await?;
        println!("Updated group: {:?}", updated_group);

        Ok(())
    }

    /// Example: Group hierarchy operations
    pub async fn group_hierarchy_operations() -> Result<(), GroupsError> {
        let client = GroupsClient::with_token("your-jwt-token")?;
        let domain_id = Uuid::new_v4();

        // Create parent group
        let parent_request = CreateGroupRequest::new("Company")
            .description("Top-level organization");
        let parent_group = client.create_group(&domain_id, parent_request).await?;

        // Create child group
        let child_request = CreateGroupRequest::new("Engineering")
            .description("Engineering department")
            .parent_id(parent_group.id);
        let child_group = client.create_group(&domain_id, child_request).await?;

        // List children of parent group
        let children = client
            .list_children_groups(&domain_id, &parent_group.id, None)
            .await?;
        println!("Children count: {}", children.total);

        // Get group hierarchy
        let hierarchy = client
            .list_group_hierarchy(
                &domain_id,
                &parent_group.id,
                Some(GroupHierarchyParams::new().level(2).tree(true)),
            )
            .await?;
        println!("Hierarchy: {:?}", hierarchy);

        Ok(())
    }

    /// Example: Role management
    pub async fn role_management_operations() -> Result<(), GroupsError> {
        let client = GroupsClient::with_token("your-jwt-token")?;
        let domain_id = Uuid::new_v4();
        let group_id = Uuid::new_v4(); // Assuming group exists

        // Create a role
        let create_role_request = CreateRoleRequest {
            name: "admin".to_string(),
            description: Some("Administrator role".to_string()),
            actions: vec!["read".to_string(), "write".to_string(), "delete".to_string()],
        };

        let role = client
            .create_group_role(&domain_id, &group_id, create_role_request)
            .await?;
        println!("Created role: {:?}", role);

        // Add members to role
        let add_members_request = AddRoleMembersRequest {
            members: vec!["user1".to_string(), "user2".to_string()],
        };

        client
            .add_group_role_members(&domain_id, &group_id, &role.id, add_members_request)
            .await?;

        // List role members
        let members = client
            .list_group_role_members(&domain_id, &group_id, &role.id)
            .await?;
        println!("Role members: {:?}", members);

        Ok(())
    }
}

impl GroupHierarchyParams {
    /// Create new group hierarchy parameters
    pub fn new() -> Self {
        Self::default()
    }

    /// Set level
    pub fn level(mut self, level: u32) -> Self {
        self.level = Some(level);
        self
    }

    /// Set tree format
    pub fn tree(mut self, tree: bool) -> Self {
        self.tree = Some(tree);
        self
    }

    /// Set direction
    pub fn direction(mut self, direction: i32) -> Self {
        self.direction = Some(direction);
        self
    }
}
*/
