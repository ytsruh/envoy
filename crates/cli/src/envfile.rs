use std::{collections::BTreeMap, fs};

pub fn parse_env_file(path: &str) -> anyhow::Result<BTreeMap<String, String>> {
    let content = fs::read_to_string(path)?;
    let mut map = BTreeMap::new();
    for (idx, line) in content.lines().enumerate() {
        let line = line.trim();
        if line.is_empty() || line.starts_with('#') {
            continue;
        }
        let (k, v) = line.split_once('=').ok_or_else(|| {
            anyhow::anyhow!("invalid format on line {}: {}", idx + 1, line)
        })?;
        let key = k.trim().to_string();
        let mut value = v.trim().to_string();
        if value.len() >= 2
            && ((value.starts_with('"') && value.ends_with('"'))
                || (value.starts_with('\'') && value.ends_with('\'')))
        {
            value = value[1..value.len() - 1].to_string();
        }
        map.insert(key, value);
    }
    Ok(map)
}

pub fn write_env_file(path: &str, vars: &BTreeMap<String, String>) -> anyhow::Result<()> {
    let mut out = String::new();
    for (k, v) in vars {
        out.push_str(&format!("{k}={v}\n"));
    }
    fs::write(path, out)?;
    Ok(())
}

pub fn sanitize_filename(name: &str) -> String {
    name.chars()
        .map(|c| {
            if c.is_ascii_alphanumeric() || c == '_' || c == '-' || c == '.' {
                c.to_ascii_lowercase()
            } else {
                '_'
            }
        })
        .collect()
}
