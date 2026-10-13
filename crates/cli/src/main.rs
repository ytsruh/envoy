mod client;
mod config;
mod envfile;
mod prompts;

use clap::{Parser, Subcommand};
use client::{Client, ClientError};
use prompts::print_client_error;
use std::collections::BTreeMap;

#[derive(Parser)]
#[command(
    name = "envoy",
    version,
    about = "Envoy CLI - manage projects and environments"
)]
struct Cli {
    #[arg(long, global = true, default_value_t = false)]
    verbose: bool,
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Version,
    Setup {
        #[command(subcommand)]
        cmd: SetupCmd,
    },
    Admin {
        #[command(subcommand)]
        cmd: AdminCmd,
    },
    Projects {
        #[command(subcommand)]
        cmd: ProjectCmd,
    },
    Environments {
        #[command(subcommand)]
        cmd: EnvCmd,
    },
    Variables {
        #[command(subcommand)]
        cmd: VarCmd,
    },
}

#[derive(Subcommand)]
enum SetupCmd {
    /// Point the CLI at a self-hosted server (clears any stale token).
    Connect {
        url: Option<String>,
    },
    /// One-time bootstrap of a fresh server; creates the first admin.
    Init,
    /// Accept an admin invite token and choose your password.
    Accept {
        token: Option<String>,
    },
    Login,
    Logout,
    /// Show the currently logged-in user.
    Whoami,
    /// Show server URL and login status.
    Status,
}

#[derive(Subcommand)]
enum AdminCmd {
    /// Create a single-use invite (prints a token to share).
    Invite {
        email: Option<String>,
    },
    /// List pending invites.
    Invites,
    /// Revoke a pending invite.
    Revoke {
        invite_id: Option<String>,
    },
    Search {
        email: Option<String>,
    },
    Members {
        #[command(subcommand)]
        cmd: MemberCmd,
    },
}

#[derive(Subcommand)]
enum ProjectCmd {
    Create,
    List,
    Get { id: Option<String> },
    Update { id: Option<String> },
    Delete { id: Option<String> },
}

#[derive(Subcommand)]
enum EnvCmd {
    Create {
        project_id: Option<String>,
    },
    List {
        project_id: Option<String>,
    },
    Get {
        project_id: Option<String>,
        environment_id: Option<String>,
    },
    Update {
        project_id: Option<String>,
        environment_id: Option<String>,
    },
    Delete {
        project_id: Option<String>,
        environment_id: Option<String>,
    },
}

#[derive(Subcommand)]
enum VarCmd {
    Import {
        #[arg(short, long, default_value = ".env")]
        file: String,
    },
    Export {
        #[arg(short, long)]
        file: Option<String>,
    },
    Create {
        project_id: Option<String>,
        environment_id: Option<String>,
    },
    List {
        project_id: Option<String>,
        environment_id: Option<String>,
    },
    Get {
        variable_id: Option<String>,
        project_id: Option<String>,
        environment_id: Option<String>,
    },
    Update {
        variable_id: Option<String>,
        project_id: Option<String>,
        environment_id: Option<String>,
    },
    Delete {
        variable_id: Option<String>,
        project_id: Option<String>,
        environment_id: Option<String>,
    },
}

#[derive(Subcommand)]
enum MemberCmd {
    List {
        project_name: Option<String>,
    },
    Add {
        project_name: Option<String>,
        email: Option<String>,
    },
    Remove {
        project_name: Option<String>,
        email: Option<String>,
    },
}

fn err<T>(e: ClientError) -> T {
    print_client_error(e);
    std::process::exit(1);
}

fn fail(msg: &str) -> ! {
    eprintln!("Error: {msg}");
    std::process::exit(1);
}

fn git_repo_detect() -> String {
    let out = std::process::Command::new("git")
        .args(["config", "--get", "remote.origin.url"])
        .output();
    let url = out
        .ok()
        .and_then(|o| String::from_utf8(o.stdout).ok())
        .unwrap_or_default();
    let url = url.trim().to_string();
    if url.is_empty() {
        return String::new();
    }
    // owner/repo from github URLs
    let trimmed = url
        .trim_end_matches(".git")
        .rsplit('/')
        .take(2)
        .collect::<Vec<_>>()
        .into_iter()
        .rev()
        .collect::<Vec<_>>();
    if trimmed.len() == 2 {
        let owner = trimmed[0].split(':').last().unwrap_or(trimmed[0]);
        return format!("{owner}/{}", trimmed[1]);
    }
    String::new()
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    match cli.command {
        Commands::Version => println!("v{}", env!("CARGO_PKG_VERSION")),
        Commands::Setup { cmd } => match cmd {
            SetupCmd::Connect { url } => {
                let url = match url {
                    Some(u) if !u.trim().is_empty() => u,
                    _ => prompts::text("Server URL", true).unwrap_or_else(|e| fail(&e.to_string())),
                };
                if let Err(e) = config::connect(&url) {
                    fail(&e.to_string());
                }
                let client = Client::new().unwrap_or_else(|e| fail(&e.to_string()));
                match client.health().await {
                    Ok(_) => println!("Connected to {}", url.trim().trim_end_matches('/')),
                    Err(e) => {
                        eprintln!("Warning: saved {url}, but health check failed: {e}");
                        eprintln!("Is the server running?");
                    }
                }
            }
            SetupCmd::Init => {
                let name = prompts::text("Name", true).unwrap_or_else(|e| fail(&e.to_string()));
                let email = prompts::email("Email").unwrap_or_else(|e| fail(&e.to_string()));
                let password = prompts::password("Password (min 8 characters)")
                    .unwrap_or_else(|e| fail(&e.to_string()));
                if password.len() < 8 {
                    fail("Password must be at least 8 characters");
                }
                let client = Client::new().unwrap_or_else(|e| fail(&e.to_string()));
                match client.setup_init(&name, &email, &password).await {
                    Ok(r) => {
                        println!("Server set up successfully!");
                        println!("Welcome, {}! You are the instance admin.", r.user.name);
                    }
                    Err(e) => err(e),
                }
            }
            SetupCmd::Accept { token } => {
                let token = match token {
                    Some(t) if !t.trim().is_empty() => t,
                    _ => {
                        prompts::text("Invite token", true).unwrap_or_else(|e| fail(&e.to_string()))
                    }
                };
                let name = prompts::text("Name", true).unwrap_or_else(|e| fail(&e.to_string()));
                let password = prompts::password("Password (min 8 characters)")
                    .unwrap_or_else(|e| fail(&e.to_string()));
                if password.len() < 8 {
                    fail("Password must be at least 8 characters");
                }
                let client = Client::new().unwrap_or_else(|e| fail(&e.to_string()));
                match client.accept_invite(&token, &name, &password).await {
                    Ok(r) => {
                        println!("Invite accepted!");
                        println!("Welcome, {}!", r.user.name);
                    }
                    Err(e) => err(e),
                }
            }
            SetupCmd::Login => {
                let email = prompts::email("Email").unwrap_or_else(|e| fail(&e.to_string()));
                let password =
                    prompts::password("Password").unwrap_or_else(|e| fail(&e.to_string()));
                let client = Client::new().unwrap_or_else(|e| fail(&e.to_string()));
                match client.login(&email, &password).await {
                    Ok(r) => {
                        println!("Login successful!");
                        println!("Welcome back, {}!", r.user.name);
                    }
                    Err(e) => err(e),
                }
            }
            SetupCmd::Logout => {
                if let Err(e) = config::clear_token() {
                    eprintln!("Error: {e:?}");
                    std::process::exit(1);
                }
                println!("Logged out successfully");
            }
            SetupCmd::Whoami => {
                let client = Client::require_token().unwrap_or_else(err);
                match client.profile().await {
                    Ok(p) => {
                        println!("Profile Information:");
                        println!("  User ID: {}", p.user_id);
                        println!("  Email: {}", p.email);
                        println!("  Token issued at: {}", p.issued_at);
                        println!("  Token expires at: {}", p.expires_at);
                    }
                    Err(e) => err(e),
                }
            }
            SetupCmd::Status => {
                println!("Server: {}", config::server_url());
                let client = match Client::require_token() {
                    Ok(c) => c,
                    Err(_) => {
                        println!("Not logged in. Use 'envoy setup login'.");
                        return;
                    }
                };
                match client.profile().await {
                    Ok(p) => {
                        println!("Logged in as: {}", p.email);
                    }
                    Err(e) => err(e),
                }
            }
        },
        Commands::Projects { cmd } => match cmd {
            ProjectCmd::Create => {
                let client = Client::require_token().unwrap_or_else(err);
                let name =
                    prompts::text("Project name", true).unwrap_or_else(|e| fail(&e.to_string()));
                let description = prompts::text("Description (optional)", false)
                    .unwrap_or_else(|e| fail(&e.to_string()));
                let detected = git_repo_detect();
                let mut git_repo = String::new();
                if detected.is_empty() {
                    git_repo = prompts::text("Git repository (owner/repo, optional)", false)
                        .unwrap_or_else(|e| fail(&e.to_string()));
                } else {
                    println!("Detected git repository: {detected}");
                    let use_it = prompts::confirm("Use this git repository?")
                        .unwrap_or_else(|e| fail(&e.to_string()));
                    if use_it {
                        git_repo = detected;
                    } else {
                        git_repo = prompts::text("Git repository (owner/repo, optional)", false)
                            .unwrap_or_else(|e| fail(&e.to_string()));
                    }
                }
                match client.create_project(&name, &description, &git_repo).await {
                    Ok(p) => {
                        println!("Project created successfully!");
                        println!("  ID: {}", p.id);
                        println!("  Name: {}", p.name);
                    }
                    Err(e) => err(e),
                }
            }
            ProjectCmd::List => {
                let client = Client::require_token().unwrap_or_else(err);
                match client.list_projects().await {
                    Ok(list) => {
                        if list.is_empty() {
                            println!("No projects found");
                        } else {
                            println!("Found {} project(s):\n", list.len());
                            for p in list {
                                println!("  ID: {}", p.id);
                                println!("  Name: {}", p.name);
                                if let Some(d) = &p.description {
                                    if !d.is_empty() {
                                        println!("  Description: {d}");
                                    }
                                }
                                println!("  Created: {}", p.created_at_string());
                                println!();
                            }
                        }
                    }
                    Err(e) => err(e),
                }
            }
            ProjectCmd::Get { id } => {
                let client = Client::require_token().unwrap_or_else(err);
                let pid = match id {
                    Some(v) => v,
                    None => prompts::select_project(&client)
                        .await
                        .unwrap_or_else(|e| fail(&e.to_string())),
                };
                match client.get_project(&pid).await {
                    Ok(p) => {
                        println!("Project Details:");
                        println!("  ID: {}", p.id);
                        println!("  Name: {}", p.name);
                        println!("  Owner ID: {}", p.owner_id);
                    }
                    Err(e) => err(e),
                }
            }
            ProjectCmd::Update { id } => {
                let client = Client::require_token().unwrap_or_else(err);
                let pid = match id {
                    Some(v) => v,
                    None => prompts::select_project(&client)
                        .await
                        .unwrap_or_else(|e| fail(&e.to_string())),
                };
                let cur = client.get_project(&pid).await.unwrap_or_else(err);
                let name = prompts::text_default("Project name", &cur.name)
                    .unwrap_or_else(|e| fail(&e.to_string()));
                let description = prompts::text("Description (leave empty to keep current)", false)
                    .unwrap_or_else(|e| fail(&e.to_string()));
                let description = if description.is_empty() {
                    cur.description.clone().unwrap_or_default()
                } else {
                    description
                };
                let git_repo = prompts::text(
                    "Git repository owner/repo (leave empty to keep current)",
                    false,
                )
                .unwrap_or_else(|e| fail(&e.to_string()));
                let git_repo = if git_repo.is_empty() {
                    cur.git_repo.clone().unwrap_or_default()
                } else {
                    git_repo
                };
                match client
                    .update_project(&pid, &name, &description, &git_repo)
                    .await
                {
                    Ok(p) => {
                        println!("Project updated successfully!");
                        println!("  Name: {}", p.name);
                    }
                    Err(e) => err(e),
                }
            }
            ProjectCmd::Delete { id } => {
                let client = Client::require_token().unwrap_or_else(err);
                let pid = match id {
                    Some(v) => v,
                    None => prompts::select_project(&client)
                        .await
                        .unwrap_or_else(|e| fail(&e.to_string())),
                };
                let cur = client.get_project(&pid).await.unwrap_or_else(err);
                println!(
                    "Are you sure you want to delete project '{}' (ID: {})?",
                    cur.name, cur.id
                );
                let ok = prompts::confirm("This action cannot be undone")
                    .unwrap_or_else(|e| fail(&e.to_string()));
                if !ok {
                    println!("Operation cancelled");
                    return;
                }
                match client.delete_project(&pid).await {
                    Ok(()) => println!("Project deleted successfully"),
                    Err(e) => err(e),
                }
            }
        },
        Commands::Environments { cmd } => handle_env(cmd).await,
        Commands::Variables { cmd } => handle_vars(cmd).await,
        Commands::Admin { cmd } => match cmd {
            AdminCmd::Invite { email } => {
                let email = match email {
                    Some(e) if !e.trim().is_empty() => e,
                    _ => prompts::email("Invitee email").unwrap_or_else(|e| fail(&e.to_string())),
                };
                let name = prompts::text("Invitee name (optional)", false)
                    .unwrap_or_else(|e| fail(&e.to_string()));
                let client = Client::require_token().unwrap_or_else(err);
                match client.create_invite(&email, &name).await {
                    Ok(inv) => {
                        println!("Invite created for {}:", inv.email);
                        println!("  Token: {}", inv.token);
                        println!("  Expires: {}", inv.expires_at);
                        println!("\nShare this token. It is shown once.");
                        println!("Invitee runs: envoy setup accept --token {}", inv.token);
                    }
                    Err(e) => err(e),
                }
            }
            AdminCmd::Invites => {
                let client = Client::require_token().unwrap_or_else(err);
                match client.list_invites().await {
                    Ok(list) => {
                        if list.is_empty() {
                            println!("No pending invites");
                        } else {
                            println!("Pending invites ({}):\n", list.len());
                            for i in list {
                                println!("  ID: {}", i.invite_id);
                                println!("  Email: {}", i.email);
                                if let Some(n) = &i.name {
                                    if !n.is_empty() {
                                        println!("  Name: {n}");
                                    }
                                }
                                println!("  Expires: {}", i.expires_at);
                                println!();
                            }
                        }
                    }
                    Err(e) => err(e),
                }
            }
            AdminCmd::Revoke { invite_id } => {
                let id = match invite_id {
                    Some(v) if !v.trim().is_empty() => v,
                    _ => prompts::text("Invite ID", true).unwrap_or_else(|e| fail(&e.to_string())),
                };
                let client = Client::require_token().unwrap_or_else(err);
                match client.revoke_invite(&id).await {
                    Ok(()) => println!("Invite revoked successfully"),
                    Err(e) => err(e),
                }
            }
            AdminCmd::Search { email } => {
                let email = email.unwrap_or_else(|| {
                    fail("email is required\nUsage: envoy admin search <email>")
                });
                let client = Client::require_token().unwrap_or_else(err);
                match client.search_users(&email).await {
                    Ok(users) => {
                        if users.is_empty() {
                            println!("No users found with that email.");
                        } else {
                            println!("Found {} user(s):\n", users.len());
                            for u in users {
                                println!("  Name:  {}", u.name);
                                println!("  Email: {}", u.email);
                                println!("  ID:    {}\n", u.user_id);
                            }
                        }
                    }
                    Err(e) => err(e),
                }
            }
            AdminCmd::Members { cmd } => match cmd {
                MemberCmd::List { project_name } => {
                    let name = project_name.unwrap_or_else(|| fail("project-name is required"));
                    let client = Client::require_token().unwrap_or_else(err);
                    let projects = client.list_projects().await.unwrap_or_else(err);
                    let proj = projects
                        .into_iter()
                        .find(|p| p.name == name)
                        .unwrap_or_else(|| fail(&format!("project '{name}' not found")));
                    match client.list_members(&proj.id).await {
                        Ok(members) => {
                            if members.is_empty() {
                                println!("No members found in this project.");
                            } else {
                                println!("Members of '{name}':\n");
                                for m in members {
                                    println!("  User ID: {}", m.user_id);
                                    println!("  Role:    {}\n", m.role);
                                }
                            }
                        }
                        Err(e) => err(e),
                    }
                }
                MemberCmd::Add {
                    project_name,
                    email,
                } => {
                    let name =
                        project_name.unwrap_or_else(|| fail("project-name and email are required"));
                    let email =
                        email.unwrap_or_else(|| fail("project-name and email are required"));
                    let client = Client::require_token().unwrap_or_else(err);
                    let projects = client.list_projects().await.unwrap_or_else(err);
                    let proj = projects
                        .into_iter()
                        .find(|p| p.name == name)
                        .unwrap_or_else(|| fail(&format!("project '{name}' not found")));
                    let users = client.search_users(&email).await.unwrap_or_else(err);
                    if users.is_empty() {
                        fail(&format!("no user found with email '{email}'"));
                    }
                    if users.len() > 1 {
                        fail("multiple users found with similar email. Use 'envoy admin search' to find the exact user.");
                    }
                    let role = prompts::select_role().unwrap_or_else(|e| fail(&e.to_string()));
                    match client.add_member(&proj.id, &users[0].user_id, &role).await {
                        Ok(()) => println!(
                            "Successfully added {} ({}) to project '{name}' as {role}",
                            users[0].name, users[0].email
                        ),
                        Err(e) => err(e),
                    }
                }
                MemberCmd::Remove {
                    project_name,
                    email,
                } => {
                    let name =
                        project_name.unwrap_or_else(|| fail("project-name and email are required"));
                    let email =
                        email.unwrap_or_else(|| fail("project-name and email are required"));
                    let client = Client::require_token().unwrap_or_else(err);
                    let projects = client.list_projects().await.unwrap_or_else(err);
                    let proj = projects
                        .into_iter()
                        .find(|p| p.name == name)
                        .unwrap_or_else(|| fail(&format!("project '{name}' not found")));
                    let users = client.search_users(&email).await.unwrap_or_else(err);
                    if users.is_empty() {
                        fail(&format!("no user found with email '{email}'"));
                    }
                    if users.len() > 1 {
                        fail("multiple users found with similar email.");
                    }
                    match client.remove_member(&proj.id, &users[0].user_id).await {
                        Ok(()) => println!(
                            "Successfully removed {} ({}) from project '{name}'",
                            users[0].name, users[0].email
                        ),
                        Err(e) => err(e),
                    }
                }
            },
        },
    }
}

trait CreatedAtString {
    fn created_at_string(&self) -> String;
}

impl CreatedAtString for envoy_shared::responses::ProjectResponse {
    fn created_at_string(&self) -> String {
        self.created_at
            .0
            .map(|d| d.to_rfc3339())
            .unwrap_or_default()
    }
}

async fn resolve_project_env(
    client: &Client,
    project_id: Option<String>,
    environment_id: Option<String>,
) -> (String, String) {
    let pid = match project_id {
        Some(v) => v,
        None => prompts::select_project(client)
            .await
            .unwrap_or_else(|e| fail(&e.to_string())),
    };
    let eid = match environment_id {
        Some(v) => v,
        None => prompts::select_environment(client, &pid)
            .await
            .unwrap_or_else(|e| fail(&e.to_string())),
    };
    (pid, eid)
}

async fn handle_env(cmd: EnvCmd) {
    let client = Client::require_token().unwrap_or_else(err);
    match cmd {
        EnvCmd::Create { project_id } => {
            let pid = match project_id {
                Some(v) => v,
                None => prompts::select_project(&client)
                    .await
                    .unwrap_or_else(|e| fail(&e.to_string())),
            };
            let name =
                prompts::text("Environment name", true).unwrap_or_else(|e| fail(&e.to_string()));
            let description = prompts::text("Description (optional)", false)
                .unwrap_or_else(|e| fail(&e.to_string()));
            match client.create_environment(&pid, &name, &description).await {
                Ok(e) => {
                    println!("Environment created successfully!");
                    println!("  ID: {}", e.id);
                    println!("  Name: {}", e.name);
                }
                Err(e) => err(e),
            }
        }
        EnvCmd::List { project_id } => {
            let pid = match project_id {
                Some(v) => v,
                None => prompts::select_project(&client)
                    .await
                    .unwrap_or_else(|e| fail(&e.to_string())),
            };
            match client.list_environments(&pid).await {
                Ok(list) => {
                    if list.is_empty() {
                        println!("No environments found");
                    } else {
                        println!("Found {} environment(s):\n", list.len());
                        for e in list {
                            println!("  ID: {}", e.id);
                            println!("  Name: {}", e.name);
                            println!();
                        }
                    }
                }
                Err(e) => err(e),
            }
        }
        EnvCmd::Get {
            project_id,
            environment_id,
        } => {
            if project_id.is_some() ^ environment_id.is_some() {
                fail("Both project_id and environment_id are required\nUsage: envoy environments get <project_id> <environment_id>");
            }
            let (pid, eid) = resolve_project_env(&client, project_id, environment_id).await;
            match client.get_environment(&pid, &eid).await {
                Ok(e) => {
                    println!("Environment Details:");
                    println!("  ID: {}", e.id);
                    println!("  Name: {}", e.name);
                    println!("  Project ID: {}", e.project_id);
                }
                Err(e) => err(e),
            }
        }
        EnvCmd::Update {
            project_id,
            environment_id,
        } => {
            if project_id.is_some() ^ environment_id.is_some() {
                fail("Both project_id and environment_id are required");
            }
            let (pid, eid) = resolve_project_env(&client, project_id, environment_id).await;
            let cur = client.get_environment(&pid, &eid).await.unwrap_or_else(err);
            let name = prompts::text_default("Environment name", &cur.name)
                .unwrap_or_else(|e| fail(&e.to_string()));
            let description = prompts::text("Description (leave empty to keep current)", false)
                .unwrap_or_else(|e| fail(&e.to_string()));
            let description = if description.is_empty() {
                cur.description.unwrap_or_default()
            } else {
                description
            };
            match client
                .update_environment(&pid, &eid, &name, &description)
                .await
            {
                Ok(e) => {
                    println!("Environment updated successfully!");
                    println!("  Name: {}", e.name);
                }
                Err(e) => err(e),
            }
        }
        EnvCmd::Delete {
            project_id,
            environment_id,
        } => {
            if project_id.is_some() ^ environment_id.is_some() {
                fail("Both project_id and environment_id are required");
            }
            let (pid, eid) = resolve_project_env(&client, project_id, environment_id).await;
            let cur = client.get_environment(&pid, &eid).await.unwrap_or_else(err);
            println!(
                "Are you sure you want to delete environment '{}' (ID: {})?",
                cur.name, cur.id
            );
            let ok = prompts::confirm("This action cannot be undone")
                .unwrap_or_else(|e| fail(&e.to_string()));
            if !ok {
                println!("Operation cancelled");
                return;
            }
            match client.delete_environment(&pid, &eid).await {
                Ok(()) => println!("Environment deleted successfully"),
                Err(e) => err(e),
            }
        }
    }
}

async fn resolve_var(
    client: &Client,
    vid: Option<String>,
    pid: Option<String>,
    eid: Option<String>,
) -> (String, String, String) {
    let provided = vid.is_some() as u8 + pid.is_some() as u8 + eid.is_some() as u8;
    if provided == 1 || provided == 2 {
        fail("All three arguments are required: variable_id, project_id, and environment_id\nUsage: envoy variables get <variable_id> <project_id> <environment_id>");
    }
    if provided == 3 {
        return (vid.unwrap(), pid.unwrap(), eid.unwrap());
    }
    let pid = prompts::select_project(client)
        .await
        .unwrap_or_else(|e| fail(&e.to_string()));
    let eid = prompts::select_environment(client, &pid)
        .await
        .unwrap_or_else(|e| fail(&e.to_string()));
    let vid = prompts::select_variable(client, &pid, &eid)
        .await
        .unwrap_or_else(|e| fail(&e.to_string()));
    (vid, pid, eid)
}

async fn handle_vars(cmd: VarCmd) {
    let client = Client::require_token().unwrap_or_else(err);
    match cmd {
        VarCmd::Import { file } => {
            let pid = prompts::select_project(&client)
                .await
                .unwrap_or_else(|e| fail(&e.to_string()));
            let eid = prompts::select_environment(&client, &pid)
                .await
                .unwrap_or_else(|e| fail(&e.to_string()));
            if std::path::Path::new(&file).metadata().is_err() {
                fail(&format!("File '{file}' not found"));
            }
            let vars = envfile::parse_env_file(&file).unwrap_or_else(|e| fail(&e.to_string()));
            if vars.is_empty() {
                println!("No variables found in {file}");
                return;
            }
            println!("Found {} variable(s) in {file}:\n", vars.len());
            for (k, v) in &vars {
                println!("  {k}={v}");
            }
            println!();
            let ok = prompts::confirm("Import these variables?")
                .unwrap_or_else(|e| fail(&e.to_string()));
            if !ok {
                println!("Import cancelled");
                return;
            }
            let mut created = 0;
            for (k, v) in &vars {
                match client.create_variable(&pid, &eid, k, v).await {
                    Ok(_) => created += 1,
                    Err(e) => eprintln!("Failed to import variable {k}: {e}"),
                }
            }
            println!("Successfully imported {created} variable(s)");
        }
        VarCmd::Export { file } => {
            let pid = prompts::select_project(&client)
                .await
                .unwrap_or_else(|e| fail(&e.to_string()));
            let eid = prompts::select_environment(&client, &pid)
                .await
                .unwrap_or_else(|e| fail(&e.to_string()));
            let out = match file {
                Some(f) => f,
                None => {
                    let env = client.get_environment(&pid, &eid).await.unwrap_or_else(err);
                    format!(".env.{}", envfile::sanitize_filename(&env.name))
                }
            };
            let vars = client.list_variables(&pid, &eid).await.unwrap_or_else(err);
            if vars.is_empty() {
                println!("No variables to export");
                return;
            }
            if std::path::Path::new(&out).exists() {
                eprintln!("Warning: File '{out}' already exists in current directory");
                let ok = prompts::confirm("Overwrite existing file?")
                    .unwrap_or_else(|e| fail(&e.to_string()));
                if !ok {
                    println!("Export cancelled");
                    return;
                }
            }
            let map: BTreeMap<String, String> =
                vars.into_iter().map(|v| (v.key, v.value)).collect();
            if let Err(e) = envfile::write_env_file(&out, &map) {
                fail(&format!("Failed to write file '{out}': {e}"));
            }
            println!("Exported {} variable(s) to {out}", map.len());
        }
        VarCmd::Create {
            project_id,
            environment_id,
        } => {
            if project_id.is_some() ^ environment_id.is_some() {
                fail("Both project_id and environment_id are required\nUsage: envoy variables create <project_id> <environment_id>");
            }
            let (pid, eid) = resolve_project_env(&client, project_id, environment_id).await;
            let key = prompts::text("Variable key", true).unwrap_or_else(|e| fail(&e.to_string()));
            let value =
                prompts::text("Variable value", true).unwrap_or_else(|e| fail(&e.to_string()));
            match client.create_variable(&pid, &eid, &key, &value).await {
                Ok(v) => {
                    println!("Variable created successfully!");
                    println!("  ID: {}", v.id);
                    println!("  Key: {}", v.key);
                }
                Err(e) => err(e),
            }
        }
        VarCmd::List {
            project_id,
            environment_id,
        } => {
            if project_id.is_some() ^ environment_id.is_some() {
                fail("Both project_id and environment_id are required");
            }
            let (pid, eid) = resolve_project_env(&client, project_id, environment_id).await;
            match client.list_variables(&pid, &eid).await {
                Ok(list) => {
                    if list.is_empty() {
                        println!("No variables found");
                    } else {
                        println!("Found {} variable(s):\n", list.len());
                        for v in list {
                            println!("  ID: {}", v.id);
                            println!("  Key: {}", v.key);
                            println!("  Value: {}", v.value);
                            println!();
                        }
                    }
                }
                Err(e) => err(e),
            }
        }
        VarCmd::Get {
            variable_id,
            project_id,
            environment_id,
        } => {
            let (vid, pid, eid) =
                resolve_var(&client, variable_id, project_id, environment_id).await;
            match client.get_variable(&pid, &eid, &vid).await {
                Ok(v) => {
                    println!("Variable Details:");
                    println!("  ID: {}", v.id);
                    println!("  Key: {}", v.key);
                    println!("  Value: {}", v.value);
                }
                Err(e) => err(e),
            }
        }
        VarCmd::Update {
            variable_id,
            project_id,
            environment_id,
        } => {
            let (vid, pid, eid) =
                resolve_var(&client, variable_id, project_id, environment_id).await;
            let cur = client
                .get_variable(&pid, &eid, &vid)
                .await
                .unwrap_or_else(err);
            let key = prompts::text_default("Variable key", &cur.key)
                .unwrap_or_else(|e| fail(&e.to_string()));
            let value =
                prompts::text("Variable value", true).unwrap_or_else(|e| fail(&e.to_string()));
            match client.update_variable(&pid, &eid, &vid, &key, &value).await {
                Ok(v) => {
                    println!("Variable updated successfully!");
                    println!("  Key: {}", v.key);
                    println!("  Value: {}", v.value);
                }
                Err(e) => err(e),
            }
        }
        VarCmd::Delete {
            variable_id,
            project_id,
            environment_id,
        } => {
            let (vid, pid, eid) =
                resolve_var(&client, variable_id, project_id, environment_id).await;
            let cur = client
                .get_variable(&pid, &eid, &vid)
                .await
                .unwrap_or_else(err);
            println!(
                "Are you sure you want to delete variable '{}' (ID: {})?",
                cur.key, cur.id
            );
            let ok = prompts::confirm("This action cannot be undone")
                .unwrap_or_else(|e| fail(&e.to_string()));
            if !ok {
                println!("Operation cancelled");
                return;
            }
            match client.delete_variable(&pid, &eid, &vid).await {
                Ok(()) => println!("Variable deleted successfully"),
                Err(e) => err(e),
            }
        }
    }
}
