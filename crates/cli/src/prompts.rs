use anyhow::Context;

use crate::client::{Client, ClientError};

pub fn text(prompt: &str, required: bool) -> anyhow::Result<String> {
    loop {
        let ans = inquire::Text::new(prompt)
            .prompt()
            .context("prompt cancelled")?;
        let v = ans.trim().to_string();
        if required && v.is_empty() {
            println!("This field is required");
            continue;
        }
        return Ok(v);
    }
}

pub fn text_default(prompt: &str, default: &str) -> anyhow::Result<String> {
    let ans = inquire::Text::new(prompt)
        .with_default(default)
        .prompt()
        .context("prompt cancelled")?;
    Ok(ans.trim().to_string())
}

pub fn password(prompt: &str) -> anyhow::Result<String> {
    Ok(rpassword::prompt_password(format!("{prompt}: "))?)
}

pub fn email(prompt: &str) -> anyhow::Result<String> {
    loop {
        let v = text(prompt, true)?;
        if !v.contains('@') || !v.contains('.') {
            println!("invalid email format");
            continue;
        }
        return Ok(v);
    }
}

pub fn confirm(prompt: &str) -> anyhow::Result<bool> {
    inquire::Confirm::new(prompt)
        .with_default(false)
        .prompt()
        .context("prompt cancelled")
}

pub async fn select_project(client: &Client) -> anyhow::Result<String> {
    let projects = client.list_projects().await?;
    if projects.is_empty() {
        anyhow::bail!("No projects found. Create one first with 'envoy projects create'");
    }
    let options: Vec<String> = projects
        .iter()
        .map(|p| {
            if let Some(d) = &p.description {
                if !d.is_empty() {
                    return format!("{} - {d} ({})", p.name, p.id);
                }
            }
            format!("{} ({})", p.name, p.id)
        })
        .collect();
    let ans = inquire::Select::new("Select a project", options).prompt()?;
    // Extract id between last parens
    let start = ans.rfind('(').map(|i| i + 1).unwrap_or(0);
    let end = ans.rfind(')').unwrap_or(ans.len());
    Ok(ans[start..end].to_string())
}

pub async fn select_environment(client: &Client, project_id: &str) -> anyhow::Result<String> {
    let envs = client.list_environments(project_id).await?;
    if envs.is_empty() {
        anyhow::bail!("No environments found. Create one first.");
    }
    let options: Vec<String> = envs
        .iter()
        .map(|e| format!("{} ({})", e.name, e.id))
        .collect();
    let ans = inquire::Select::new("Select an environment", options).prompt()?;
    let start = ans.rfind('(').map(|i| i + 1).unwrap_or(0);
    let end = ans.rfind(')').unwrap_or(ans.len());
    Ok(ans[start..end].to_string())
}

pub async fn select_variable(
    client: &Client,
    project_id: &str,
    env_id: &str,
) -> anyhow::Result<String> {
    let vars = client.list_variables(project_id, env_id).await?;
    if vars.is_empty() {
        anyhow::bail!("No variables found.");
    }
    let options: Vec<String> = vars
        .iter()
        .map(|v| format!("{} = {} ({})", v.key, v.value, v.id))
        .collect();
    let ans = inquire::Select::new("Select a variable", options).prompt()?;
    let start = ans.rfind('(').map(|i| i + 1).unwrap_or(0);
    let end = ans.rfind(')').unwrap_or(ans.len());
    Ok(ans[start..end].to_string())
}

pub fn select_role() -> anyhow::Result<String> {
    let ans = inquire::Select::new(
        "Enter role (viewer/editor)",
        vec!["viewer".to_string(), "editor".to_string()],
    )
    .prompt()?;
    Ok(ans)
}

pub fn print_client_error(e: ClientError) {
    match e {
        ClientError::NoToken => {
            eprintln!("Error: no token");
            println!("Please login first using 'envoy setup login'");
        }
        ClientError::Expired => {
            eprintln!("Error: expired token");
            println!("Your session has expired. Please login again using 'envoy setup login'");
        }
        ClientError::Server(msg) => eprintln!("Error: {msg}"),
        ClientError::Transport(err) => eprintln!("Error: {err}"),
        ClientError::Other(err) => eprintln!("Error: {err:?}"),
    }
}
