use crate::db::Db;

#[derive(Clone)]
pub struct AppState {
    pub db: Db,
    pub jwt_secret: String,
}
