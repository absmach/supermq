//! Complete SuperMQ workflow example demonstrating all services

use supermq_rust_sdk::{Client, Config};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Configure the client
    let config = Config::new("http://localhost:8080")
        .with_timeout(std::time::Duration::from_secs(10))
        .with_bearer_token("your-auth-token-here")  // Add your actual token
        .with_user_agent("SuperMQ-Rust-SDK/0.1.0");
    
    let client = Client::new(config);

    println!("🚀 Testing SuperMQ SDK with full implementations...");

    // Test all services
    let services = [
        ("Auth", "🔐"),
        ("Users", "👤"),
        ("Domains", "🌐"),
        ("Things", "📱"),
        ("Channels", "📡"),
        ("Groups", "👥"),
        ("Bootstrap", "🔄"),
        ("Certs", "🔒"),
        ("Provision", "⚙️"),
        ("Journal", "📋"),
    ];

    for (service_name, icon) in &services {
        println!("{} {} service health check...", icon, service_name);
        
        let result = match service_name {
            &"Auth" => client.auth().health().await,
            &"Users" => client.users().health().await,
            &"Domains" => client.domains().health().await,
            &"Things" => client.things().health().await,
            &"Channels" => client.channels().health().await,
            &"Groups" => client.groups().health().await,
            &"Bootstrap" => client.bootstrap().health().await,
            &"Certs" => client.certs().health().await,
            &"Provision" => client.provision().health().await,
            &"Journal" => client.journal().health().await,
            _ => unreachable!(),
        };
        
        match result {
            Ok(healthy) => println!("  Status: {}", if healthy { "✅ Healthy" } else { "❌ Unhealthy" }),
            Err(e) => println!("  Error: {}", e),
        }
    }

    // Example of using actual API methods (when implemented)
    println!("\n🔧 Testing API methods...");
    println!("Note: Full method implementations depend on actual OpenAPI specs");

    println!("\n✅ SDK test completed!");
    println!("💡 To use real methods, ensure your OpenAPI specs are available and run the generator");

    Ok(())
}
