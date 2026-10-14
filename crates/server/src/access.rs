use libsql::Connection;

use crate::error::ApiError;

async fn count(conn: &Connection, sql: &str, args: Vec<libsql::Value>) -> Result<i64, ApiError> {
    let mut rows = conn
        .query(sql, args)
        .await
        .map_err(|e| ApiError::Internal(e.into()))?;
    let row = rows
        .next()
        .await
        .map_err(|e| ApiError::Internal(e.into()))?
        .ok_or_else(|| ApiError::Internal(anyhow::anyhow!("count query returned no rows")))?;
    let n: i64 = row.get(0).map_err(|e| ApiError::Internal(e.into()))?;
    Ok(n)
}

pub async fn require_admin(conn: &Connection, user_id: &str) -> Result<(), ApiError> {
    let n = count(
        conn,
        "SELECT COUNT(*) FROM users WHERE id = ?1 AND is_admin = 1 AND deleted_at IS NULL",
        vec![user_id.into()],
    )
    .await?;
    if n == 0 {
        return Err(ApiError::forbidden("admin access required"));
    }
    Ok(())
}

pub async fn require_owner(
    conn: &Connection,
    project_id: &str,
    user_id: &str,
) -> Result<(), ApiError> {
    let n = count(
        conn,
        "SELECT COUNT(*) FROM projects WHERE id = ?1 AND owner_id = ?2 AND deleted_at IS NULL",
        vec![project_id.into(), user_id.into()],
    )
    .await?;
    if n == 0 {
        return Err(ApiError::forbidden("access denied"));
    }
    Ok(())
}

pub async fn require_editor(
    conn: &Connection,
    project_id: &str,
    user_id: &str,
) -> Result<(), ApiError> {
    // owner OR member with role=editor
    let owner = count(
        conn,
        "SELECT COUNT(*) FROM projects WHERE id = ?1 AND owner_id = ?2 AND deleted_at IS NULL",
        vec![project_id.into(), user_id.into()],
    )
    .await?;
    if owner > 0 {
        return Ok(());
    }
    let editor = count(
        conn,
        "SELECT COUNT(*) FROM project_users WHERE project_id = ?1 AND user_id = ?2 AND role = 'editor'",
        vec![project_id.into(), user_id.into()],
    )
    .await?;
    if editor == 0 {
        return Err(ApiError::forbidden("access denied"));
    }
    Ok(())
}

pub async fn require_viewer(
    conn: &Connection,
    project_id: &str,
    user_id: &str,
) -> Result<(), ApiError> {
    let owner = count(
        conn,
        "SELECT COUNT(*) FROM projects WHERE id = ?1 AND owner_id = ?2 AND deleted_at IS NULL",
        vec![project_id.into(), user_id.into()],
    )
    .await?;
    if owner > 0 {
        return Ok(());
    }
    let member = count(
        conn,
        "SELECT COUNT(*) FROM project_users WHERE project_id = ?1 AND user_id = ?2",
        vec![project_id.into(), user_id.into()],
    )
    .await?;
    if member == 0 {
        return Err(ApiError::forbidden("access denied"));
    }
    Ok(())
}
