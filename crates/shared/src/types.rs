use chrono::{DateTime, Utc};
use serde::{Deserialize, Deserializer, Serialize, Serializer};
use utoipa::ToSchema;

/// RFC3339 timestamp that serializes absence as null.
#[derive(Debug, Clone, Copy, PartialEq, Eq, ToSchema)]
pub struct Timestamp(pub Option<DateTime<Utc>>);

impl Timestamp {
    pub fn now() -> Self {
        Self(Some(Utc::now()))
    }

    pub fn null() -> Self {
        Self(None)
    }

    pub fn into_inner(self) -> Option<DateTime<Utc>> {
        self.0
    }
}

impl From<DateTime<Utc>> for Timestamp {
    fn from(dt: DateTime<Utc>) -> Self {
        Self(Some(dt))
    }
}

impl From<Option<DateTime<Utc>>> for Timestamp {
    fn from(o: Option<DateTime<Utc>>) -> Self {
        Self(o)
    }
}

/// Parse SQLite CURRENT_TIMESTAMP shapes: RFC3339, "YYYY-MM-DD HH:MM:SS", with optional nanos.
pub fn parse_sqlite_time(s: &str) -> Option<DateTime<Utc>> {
    if let Ok(dt) = s.parse::<DateTime<Utc>>() {
        return Some(dt);
    }
    for fmt in ["%Y-%m-%d %H:%M:%S%.f", "%Y-%m-%d %H:%M:%S"] {
        if let Ok(naive) = chrono::NaiveDateTime::parse_from_str(s, fmt) {
            return Some(naive.and_utc());
        }
    }
    None
}

pub fn sqlite_now() -> String {
    Utc::now().format("%Y-%m-%d %H:%M:%S").to_string()
}

impl Serialize for Timestamp {
    fn serialize<S: Serializer>(&self, s: S) -> Result<S::Ok, S::Error> {
        match self.0 {
            Some(dt) => s.serialize_str(&dt.to_rfc3339()),
            None => s.serialize_none(),
        }
    }
}

impl<'de> Deserialize<'de> for Timestamp {
    fn deserialize<D: Deserializer<'de>>(d: D) -> Result<Self, D::Error> {
        let opt: Option<String> = Option::deserialize(d)?;
        match opt.as_deref() {
            None | Some("") => Ok(Self(None)),
            Some(raw) => {
                if let Ok(dt) = raw.parse::<DateTime<Utc>>() {
                    return Ok(Self(Some(dt)));
                }
                Err(serde::de::Error::custom("invalid RFC3339 timestamp"))
            }
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, ToSchema)]
#[serde(rename_all = "lowercase")]
pub enum Role {
    Owner,
    Editor,
    Viewer,
}

impl Role {
    pub fn as_str(&self) -> &'static str {
        match self {
            Role::Owner => "owner",
            Role::Editor => "editor",
            Role::Viewer => "viewer",
        }
    }

    pub fn parse(s: &str) -> Option<Self> {
        match s {
            "owner" => Some(Role::Owner),
            "editor" => Some(Role::Editor),
            "viewer" => Some(Role::Viewer),
            _ => None,
        }
    }

    /// Roles assignable via sharing endpoints (owner excluded).
    pub fn is_assignable(&self) -> bool {
        matches!(self, Role::Editor | Role::Viewer)
    }
}
