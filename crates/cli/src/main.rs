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
    about = "Envoy CLI - secure environment variables"
)]
struct Cli {
    #[arg(long, global = true, default_value_t = false)]
    verbose: bool,
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Print the CLI version.
    Version,
    /// First-time setup: connect to a server, bootstrap admin, login/logout.
    Setup {
        #[command(subcommand)]
        cmd: SetupCmd,
    },
    /// Instance administration: invites, user search, project members.
    Admin {
        #[command(subcommand)]
        cmd: AdminCmd,
    },
    /// Project management: create and list projects (rarely needed daily).
    Projects {
        #[command(subcommand)]
        cmd: ProjectCmd,
    },
    /// Environment management: create and list environments (daily work uses `vars` instead).
    Envs {
        #[command(subcommand)]
        cmd: EnvsCmd,
    },
    /// Daily variables: pull, push, set, get, unset and list (uses linked project/env).
    Vars {
        #[command(subcommand)]
        cmd: VarsCmd,
    },
    /// Link this directory to a project (stored globally in ~/.envoy/config.json, defaults env to `dev`).
    Link {
        /// Project name or ID (defaults to git remote match or prompt).
        #[arg(long, short = 'p')]
        project: Option<String>,
    },
    /// Set the default environment for this directory (switch-only; fresh links already use `dev`).
    Use {
        /// Environment name or ID.
        env: Option<String>,
        #[arg(long, short = 'p')]
        project: Option<String>,
    },
    /// List all directory links stored in the global config.
    Links,
    /// Remove the link for the current directory.
    Unlink,
    /// Show server URL, link for this directory, and login status.
    Status,
}

#[derive(Subcommand)]
enum VarsCmd {
    /// Download variables to a .env file (defaults to `.env.<env>`).
    Pull {
        /// Environment name or ID (defaults to linked env, then `dev`).
        #[arg(long, short = 'e')]
        env: Option<String>,
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
        /// Output file (defaults to `.env.<env>`).
        #[arg(short, long)]
        file: Option<String>,
    },
    /// Upload a .env file, creating or overwriting matching keys (never creates environments).
    Push {
        /// Environment name or ID (defaults to linked env, then `dev`).
        #[arg(long, short = 'e')]
        env: Option<String>,
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
        /// Input file to upload.
        #[arg(short, long, default_value = ".env")]
        file: String,
    },
    /// Create or update a single variable by key (upsert).
    Set {
        /// Variable key (e.g. API_KEY).
        key: String,
        /// Variable value.
        value: String,
        /// Environment name or ID (defaults to linked env, then `dev`).
        #[arg(long, short = 'e')]
        env: Option<String>,
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
    },
    /// Print a single variable value (script-friendly, value only).
    Get {
        /// Variable key (e.g. API_KEY).
        key: String,
        /// Environment name or ID (defaults to linked env, then `dev`).
        #[arg(long, short = 'e')]
        env: Option<String>,
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
    },
    /// Delete a variable by key (asks to confirm unless `--yes`).
    Unset {
        /// Variable key (e.g. API_KEY).
        key: String,
        /// Environment name or ID (defaults to linked env, then `dev`).
        #[arg(long, short = 'e')]
        env: Option<String>,
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
        /// Skip the confirmation prompt.
        #[arg(long)]
        yes: bool,
    },
    /// List variables as KEY=VALUE lines (safe to pipe to files).
    List {
        /// Environment name or ID (defaults to linked env, then `dev`).
        #[arg(long, short = 'e')]
        env: Option<String>,
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
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
    /// Log in with email and password (stores token globally).
    Login,
    /// Log out (clears the stored token).
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
        /// Invitee email address.
        email: Option<String>,
    },
    /// List pending invites.
    Invites,
    /// Revoke a pending invite.
    Revoke {
        /// Invite ID to revoke.
        invite_id: Option<String>,
    },
    /// Search users by email (for adding project members).
    Search {
        /// Email to search for.
        email: Option<String>,
    },
    /// Manage project members (list, add, remove).
    Members {
        #[command(subcommand)]
        cmd: MemberCmd,
    },
}

#[derive(Subcommand)]
enum ProjectCmd {
    /// Create a new project (auto-creates a `dev` environment).
    Create,
    /// List projects you have access to.
    List,
}

#[derive(Subcommand)]
enum EnvsCmd {
    /// Create a new environment in a project (the ONLY way to add environments).
    Create {
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
    },
    /// List environments in a project.
    List {
        /// Project name or ID (defaults to linked project).
        #[arg(long, short = 'p')]
        project: Option<String>,
    },
}

#[derive(Subcommand)]
enum MemberCmd {
    /// List members of a project.
    List {
        /// Project name.
        project_name: Option<String>,
    },
    /// Add a user to a project as viewer or editor (prompts for role).
    Add {
        /// Project name.
        project_name: Option<String>,
        /// User email.
        email: Option<String>,
    },
    /// Remove a user from a project.
    Remove {
        /// Project name.
        project_name: Option<String>,
        /// User email.
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
        let owner = trimmed[0].split(':').next_back().unwrap_or(trimmed[0]);
        return format!("{owner}/{}", trimmed[1]);
    }
    String::new()
}

fn env_override(key: &str) -> Option<String> {
    std::env::var(key)
        .ok()
        .map(|v| v.trim().to_string())
        .filter(|v| !v.is_empty())
}

async fn resolve_project_ctx(
    client: &Client,
    flag: Option<String>,
) -> envoy_shared::responses::ProjectResponse {
    if let Some(input) = flag.filter(|v| !v.trim().is_empty()) {
        return client.resolve_project(&input).await.unwrap_or_else(err);
    }
    if let Some(input) = env_override("ENVOY_PROJECT") {
        return client.resolve_project(&input).await.unwrap_or_else(err);
    }
    if let Some(link) = config::link_for_cwd() {
        if !link.project_id.is_empty() {
            if let Ok(p) = client.get_project(&link.project_id).await {
                return p;
            }
        }
        if !link.project_name.is_empty() {
            if let Ok(p) = client.resolve_project(&link.project_name).await {
                return p;
            }
        }
    }
    let detected = git_repo_detect();
    if !detected.is_empty() {
        let projects = client.list_projects().await.unwrap_or_else(err);
        if let Some(p) = projects
            .iter()
            .find(|p| p.git_repo.as_deref() == Some(detected.as_str()))
        {
            return p.clone();
        }
    }
    let id = prompts::select_project(client)
        .await
        .unwrap_or_else(|e| fail(&e.to_string()));
    client.get_project(&id).await.unwrap_or_else(err)
}

async fn resolve_env_ctx(
    client: &Client,
    project: &envoy_shared::responses::ProjectResponse,
    flag: Option<String>,
) -> envoy_shared::responses::EnvironmentResponse {
    if let Some(input) = flag.filter(|v| !v.trim().is_empty()) {
        return client
            .resolve_environment(&project.id, &input)
            .await
            .unwrap_or_else(err);
    }
    if let Some(input) = env_override("ENVOY_ENV") {
        return client
            .resolve_environment(&project.id, &input)
            .await
            .unwrap_or_else(err);
    }
    let link = config::link_for_cwd();
    if let Some(l) = link {
        if l.project_id == project.id && !l.env_name.is_empty() {
            if let Ok(e) = client
                .resolve_environment(&project.id, &l.env_name)
                .await
            {
                return e;
            }
        }
    }
    // Default: linked-but-no-env (or stale link env) resolves to `dev` when it
    // exists, so `link` is immediately usable without a mandatory `use`.
    if let Ok(e) = client.resolve_environment(&project.id, "dev").await {
        return e;
    }
    // No explicit env: single-env projects resolve without prompting.
    // Empty projects fail with a pointer to creation (never auto-create on typo).
    let envs = client
        .list_environments(&project.id)
        .await
        .unwrap_or_else(err);
    if envs.is_empty() {
        fail(&format!(
            "No environments in project '{}'. Run 'envoy envs create --project {}'",
            project.name, project.name
        ));
    }
    if envs.len() == 1 {
        return envs.into_iter().next().unwrap_or_else(|| fail("No environments found"));
    }
    let eid = prompts::select_environment(client, &project.id)
        .await
        .unwrap_or_else(|e| fail(&e.to_string()));
    client.get_environment(&eid).await.unwrap_or_else(err)
}

async fn resolve_ctx(
    client: &Client,
    project_flag: Option<String>,
    env_flag: Option<String>,
) -> (
    envoy_shared::responses::ProjectResponse,
    envoy_shared::responses::EnvironmentResponse,
) {
    let proj = resolve_project_ctx(client, project_flag).await;
    let env = resolve_env_ctx(client, &proj, env_flag).await;
    (proj, env)
}

async fn print_status(client: &Client) {
    println!("Server: {}", config::server_url());
    match config::link_for_cwd() {
        None => println!("Link: not linked here (run 'envoy link')"),
        Some(l) => {
            let env = if l.env_name.is_empty() {
                "(no default env, run 'envoy use <env>')".to_string()
            } else {
                l.env_name.clone()
            };
            println!("Link: {} ({}) env: {}", l.project_name, l.project_id, env);
        }
    }
    match client.profile().await {
        Ok(p) => println!("Logged in as: {}", p.email),
        Err(_) => println!("Not logged in. Use 'envoy setup login'."),
    }
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    match cli.command {
        Commands::Version => println!("v{}", env!("CARGO_PKG_VERSION")),
        Commands::Status => {
            let client = Client::new().unwrap_or_else(|e| fail(&e.to_string()));
            // Best-effort: profile may fail when logged out; print_status handles it.
            print_status(&client).await;
        }
        Commands::Link { project } => {
            let client = Client::require_token().unwrap_or_else(err);
            let proj = resolve_project_ctx(&client, project).await;
            let prev = config::link_for_cwd();
            let env_name = match prev {
                Some(l) if l.project_id == proj.id && !l.env_name.is_empty() => l.env_name,
                _ => "dev".to_string(),
            };
            let entry = config::LinkEntry {
                project_id: proj.id.clone(),
                project_name: proj.name.clone(),
                env_name,
                git_repo: proj.git_repo.clone().unwrap_or_default(),
            };
            match config::set_link(&entry) {
                Ok(key) => {
                    println!("Linked {} to project '{}' ({}) env: dev", key, proj.name, proj.id);
                    println!("Default environment is 'dev'. Switch with 'envoy use <env>'");
                }
                Err(e) => fail(&e.to_string()),
            }
        }
        Commands::Use { env, project } => {
            let client = Client::require_token().unwrap_or_else(err);
            let proj = resolve_project_ctx(&client, project).await;
            let selected = match env {
                Some(e) if !e.trim().is_empty() => client
                    .resolve_environment(&proj.id, &e)
                    .await
                    .unwrap_or_else(err),
                _ => resolve_env_ctx(&client, &proj, None).await,
            };
            let ctx = config::LinkEntry {
                project_id: proj.id.clone(),
                project_name: proj.name.clone(),
                env_name: selected.name.clone(),
                git_repo: proj.git_repo.clone().unwrap_or_default(),
            };
            if let Err(e) = config::set_link(&ctx) {
                fail(&e.to_string());
            }
            println!(
                "Default env for '{}' is now '{}'",
                proj.name, selected.name
            );
        }
        Commands::Vars { cmd } => match cmd {
            VarsCmd::Pull { env, project, file } => {
            let client = Client::require_token().unwrap_or_else(err);
            let (_proj, env_res) = resolve_ctx(&client, project, env).await;
            let out = file.unwrap_or_else(|| {
                format!(".env.{}", envfile::sanitize_filename(&env_res.name))
            });
            let vars = client
                .list_variables(&env_res.id)
                .await
                .unwrap_or_else(err);
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
            let n = map.len();
            if let Err(e) = envfile::write_env_file(&out, &map) {
                fail(&format!("Failed to write file '{out}': {e}"));
            }
            println!("Pulled {n} variable(s) to {out}");
        }
            VarsCmd::Push { env, project, file } => {
            let client = Client::require_token().unwrap_or_else(err);
            let (proj, env_res) = resolve_ctx(&client, project, env).await;
            if std::path::Path::new(&file).metadata().is_err() {
                fail(&format!("File '{file}' not found"));
            }
            let vars = envfile::parse_env_file(&file).unwrap_or_else(|e| fail(&e.to_string()));
            if vars.is_empty() {
                println!("No variables found in {file}");
                return;
            }
            let existing = client
                .list_variables(&env_res.id)
                .await
                .unwrap_or_else(err);
            let mut created = 0;
            let mut updated = 0;
            for (k, v) in &vars {
                match existing.iter().find(|e| e.key == *k) {
                    Some(cur) if cur.value == *v => {}
                    Some(cur) => {
                        match client.update_variable(&cur.id, k, v).await {
                            Ok(_) => updated += 1,
                            Err(e) => eprintln!("Failed to update {k}: {e}"),
                        }
                    }
                    None => match client.create_variable(&env_res.id, k, v).await {
                        Ok(_) => created += 1,
                        Err(e) => eprintln!("Failed to create {k}: {e}"),
                    },
                }
            }
            println!("Pushed to '{}/{}': {created} created, {updated} updated", proj.name, env_res.name);
        }
            VarsCmd::Set {
            key,
            value,
            env,
            project,
        } => {
            let client = Client::require_token().unwrap_or_else(err);
            let (_proj, env_res) = resolve_ctx(&client, project, env).await;
            if key.trim().is_empty() || value.is_empty() {
                fail("key and value are required");
            }
            let existing = client
                .list_variables(&env_res.id)
                .await
                .unwrap_or_else(err);
            match existing.iter().find(|v| v.key == key) {
                Some(cur) => match client.update_variable(&cur.id, &key, &value).await {
                    Ok(v) => println!("Updated {}={} in '{}'", v.key, v.value, env_res.name),
                    Err(e) => err(e),
                },
                None => match client.create_variable(&env_res.id, &key, &value).await {
                    Ok(v) => println!("Set {}={} in '{}'", v.key, v.value, env_res.name),
                    Err(e) => err(e),
                },
            }
        }
            VarsCmd::Get { key, env, project } => {
            let client = Client::require_token().unwrap_or_else(err);
            let (_proj, env_res) = resolve_ctx(&client, project, env).await;
            match client
                .resolve_variable(&env_res.id, &key)
                .await
            {
                Ok(v) => println!("{}", v.value),
                Err(e) => err(e),
            }
        }
            VarsCmd::Unset {
            key,
            env,
            project,
            yes,
        } => {
            let client = Client::require_token().unwrap_or_else(err);
            let (_proj, env_res) = resolve_ctx(&client, project, env).await;
            let cur = client
                .resolve_variable(&env_res.id, &key)
                .await
                .unwrap_or_else(err);
            if !yes {
                println!("Delete '{key}' from '{}'?", env_res.name);
                let ok = prompts::confirm("This action cannot be undone")
                    .unwrap_or_else(|e| fail(&e.to_string()));
                if !ok {
                    println!("Operation cancelled");
                    return;
                }
            }
            match client.delete_variable(&cur.id).await {
                Ok(()) => println!("Deleted '{key}'"),
                Err(e) => err(e),
            }
        }
            VarsCmd::List { env, project } => {
                let client = Client::require_token().unwrap_or_else(err);
                let (_proj, env_res) = resolve_ctx(&client, project, env).await;
                match client.list_variables(&env_res.id).await {
                    Ok(list) => {
                        if list.is_empty() {
                            println!("No variables found");
                        } else {
                            for v in list {
                                println!("{}={}", v.key, v.value);
                            }
                        }
                    }
                    Err(e) => err(e),
                }
            }
        }
        Commands::Links => {
            let links = config::list_links();
            if links.is_empty() {
                println!("No links yet. Run 'envoy link' in a project directory.");
            } else {
                for (dir, l) in links {
                    let env = if l.env_name.is_empty() { "-" } else { &l.env_name };
                    println!("{dir} -> {} ({}) env: {env}", l.project_name, l.project_id);
                }
            }
        }
        Commands::Unlink => match config::remove_link() {
            Ok(true) => println!("Unlinked current directory"),
            Ok(false) => println!("Not linked here"),
            Err(e) => fail(&e.to_string()),
        },
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
                let client = Client::new().unwrap_or_else(|e| fail(&e.to_string()));
                print_status(&client).await;
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
                let git_repo: String;
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
                        println!("  Default environment 'dev' created. Add more with 'envoy envs create --project {}'", p.name);
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
                                println!();
                            }
                        }
                    }
                    Err(e) => err(e),
                }
            }
        },
        Commands::Envs { cmd } => match cmd {
            EnvsCmd::Create { project } => {
                let client = Client::require_token().unwrap_or_else(err);
                let proj = resolve_project_ctx(&client, project).await;
                let name =
                    prompts::text("Environment name", true).unwrap_or_else(|e| fail(&e.to_string()));
                match client.create_environment(&proj.id, &name, "").await {
                    Ok(e) => {
                        println!("Environment '{}' created in '{}'!", e.name, proj.name);
                        println!("  ID: {}", e.id);
                    }
                    Err(e) => err(e),
                }
            }
            EnvsCmd::List { project } => {
                let client = Client::require_token().unwrap_or_else(err);
                let proj = resolve_project_ctx(&client, project).await;
                match client.list_environments(&proj.id).await {
                    Ok(list) => {
                        if list.is_empty() {
                            println!("No environments found");
                        } else {
                            println!("Environments in '{}':\n", proj.name);
                            for e in list {
                                println!("  {} ({})", e.name, e.id);
                            }
                        }
                    }
                    Err(e) => err(e),
                }
            }
        },
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
                    let proj = client.resolve_project(&name).await.unwrap_or_else(err);
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
                    let proj = client.resolve_project(&name).await.unwrap_or_else(err);
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
                    let proj = client.resolve_project(&name).await.unwrap_or_else(err);
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
