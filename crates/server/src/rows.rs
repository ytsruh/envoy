use envoy_shared::types::{Timestamp, parse_sqlite_time};

pub fn opt_text(v: libsql::Value) -> Option<String> {
    match v {
        libsql::Value::Null => None,
        libsql::Value::Text(s) => {
            if s.is_empty() {
                None
            } else {
                Some(s)
            }
        }
        libsql::Value::Integer(i) => Some(i.to_string()),
        libsql::Value::Real(f) => Some(f.to_string()),
        libsql::Value::Blob(b) => String::from_utf8(b).ok(),
    }
}

pub fn ts_from_row_string(s: Option<String>) -> Timestamp {
    match s {
        Some(raw) => Timestamp(parse_sqlite_time(&raw)),
        None => Timestamp(None),
    }
}

pub fn get_string(row: &libsql::Row, idx: i32) -> anyhow::Result<String> {
    Ok(row.get::<String>(idx)?)
}

pub fn get_opt_string(row: &libsql::Row, idx: i32) -> anyhow::Result<Option<String>> {
    let v: Option<String> = row.get(idx)?;
    Ok(v.filter(|s| !s.is_empty()))
}
