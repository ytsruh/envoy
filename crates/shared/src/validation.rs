/// Project/environment names are 1-100 chars, charset [a-zA-Z0-9 \-_];
/// descriptions max 500 chars.
pub fn validate_name(value: &str) -> Result<(), String> {
    if value.is_empty() || value.len() > 100 {
        return Err("name must be between 1 and 100 characters".to_string());
    }
    if !value
        .chars()
        .all(|c| c.is_ascii_alphanumeric() || c == ' ' || c == '-' || c == '_')
    {
        return Err("name may only contain letters, numbers, spaces, - and _".to_string());
    }
    Ok(())
}

pub fn validate_description(value: &str) -> Result<(), String> {
    if value.len() > 500 {
        return Err("description must be at most 500 characters".to_string());
    }
    Ok(())
}

pub fn validate_git_repo(value: &str) -> Result<(), String> {
    if value.len() > 500 {
        return Err("git_repo must be at most 500 characters".to_string());
    }
    Ok(())
}

pub fn validate_register(name: &str, email: &str, password: &str) -> Result<(), String> {
    if name.is_empty() || name.len() > 50 {
        return Err("name must be between 1 and 50 characters".to_string());
    }
    validate_email(email)?;
    if password.len() < 8 {
        return Err("password must be at least 8 characters".to_string());
    }
    Ok(())
}

pub fn validate_login(email: &str, password: &str) -> Result<(), String> {
    validate_email(email)?;
    if password.len() < 8 {
        return Err("password must be at least 8 characters".to_string());
    }
    Ok(())
}

pub fn validate_email(email: &str) -> Result<(), String> {
    if email.is_empty() || email.len() > 100 {
        return Err("invalid email".to_string());
    }
    let parts: Vec<&str> = email.split('@').collect();
    if parts.len() != 2 || parts[0].is_empty() || parts[1].is_empty() || !parts[1].contains('.') {
        return Err("invalid email".to_string());
    }
    Ok(())
}
