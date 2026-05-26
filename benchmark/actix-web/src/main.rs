use actix_web::{delete, get, post, put, web, App, HttpResponse, HttpServer, Responder};
use rusqlite::{params, Connection, OptionalExtension};
use serde::{Deserialize, Serialize};
use std::env;
use std::fs;
use std::path::Path;
use std::sync::Mutex;
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Clone)]
struct Config {
    service: String,
    version: String,
}

struct State {
    db: Mutex<Connection>,
    config: Config,
}

#[derive(Serialize)]
struct Health {
    status: &'static str,
    service: String,
    version: String,
}

#[derive(Serialize)]
struct ErrorBody<'a> {
    #[serde(rename = "statusCode")]
    status_code: u16,
    message: &'a str,
}

#[derive(Deserialize)]
struct UserPayload {
    name: String,
    email: String,
    age: i64,
}

#[derive(Serialize)]
struct User {
    id: i64,
    name: String,
    email: String,
    age: i64,
    #[serde(rename = "createdAt")]
    created_at: String,
    #[serde(rename = "updatedAt")]
    updated_at: String,
}

#[derive(Serialize)]
struct UserList {
    items: Vec<User>,
    limit: i64,
    offset: i64,
    total: i64,
}

#[derive(Deserialize)]
struct ListQuery {
    limit: Option<i64>,
    offset: Option<i64>,
}

#[get("/health")]
async fn health(state: web::Data<State>) -> impl Responder {
    web::Json(Health { status: "ok", service: state.config.service.clone(), version: state.config.version.clone() })
}

#[post("/users")]
async fn create(state: web::Data<State>, payload: web::Json<UserPayload>) -> impl Responder {
    if !valid_payload(&payload) {
        return error(400, "name, email and age are required");
    }
    let db = state.db.lock().unwrap();
    let now = timestamp();
    let result = db.execute(
        "INSERT INTO users (name, email, age, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5)",
        params![payload.name.trim(), payload.email.trim(), payload.age, now, now],
    );
    if result.is_err() {
        return error(400, "user could not be created");
    }
    let id = db.last_insert_rowid();
    match find_user(&db, id).unwrap_or(None) {
        Some(user) => HttpResponse::Created().json(user),
        None => error(500, "user could not be fetched"),
    }
}

#[get("/users")]
async fn list(state: web::Data<State>, query: web::Query<ListQuery>) -> impl Responder {
    let limit = query.limit.unwrap_or(20).max(1);
    let offset = query.offset.unwrap_or(0).max(0);
    let db = state.db.lock().unwrap();
    let mut statement = match db.prepare("SELECT id, name, email, age, created_at, updated_at FROM users ORDER BY id DESC LIMIT ?1 OFFSET ?2") {
        Ok(statement) => statement,
        Err(_) => return error(500, "users could not be listed"),
    };
    let rows = match statement.query_map(params![limit, offset], read_user) {
        Ok(rows) => rows,
        Err(_) => return error(500, "users could not be listed"),
    };
    let mut items = Vec::new();
    for row in rows {
        match row {
            Ok(user) => items.push(user),
            Err(_) => return error(500, "users could not be listed"),
        }
    }
    let total = db.query_row("SELECT COUNT(*) FROM users", [], |row| row.get(0)).unwrap_or(0);
    HttpResponse::Ok().json(UserList { items, limit, offset, total })
}

#[get("/users/{id}")]
async fn get_user(state: web::Data<State>, path: web::Path<i64>) -> impl Responder {
    let db = state.db.lock().unwrap();
    match find_user(&db, path.into_inner()).unwrap_or(None) {
        Some(user) => HttpResponse::Ok().json(user),
        None => error(404, "user not found"),
    }
}

#[put("/users/{id}")]
async fn update(state: web::Data<State>, path: web::Path<i64>, payload: web::Json<UserPayload>) -> impl Responder {
    if !valid_payload(&payload) {
        return error(400, "name, email and age are required");
    }
    let id = path.into_inner();
    let db = state.db.lock().unwrap();
    let result = db.execute(
        "UPDATE users SET name = ?1, email = ?2, age = ?3, updated_at = ?4 WHERE id = ?5",
        params![payload.name.trim(), payload.email.trim(), payload.age, timestamp(), id],
    );
    match result {
        Ok(0) => error(404, "user not found"),
        Ok(_) => match find_user(&db, id).unwrap_or(None) {
            Some(user) => HttpResponse::Ok().json(user),
            None => error(404, "user not found"),
        },
        Err(_) => error(400, "user could not be updated"),
    }
}

#[delete("/users/{id}")]
async fn delete_user(state: web::Data<State>, path: web::Path<i64>) -> impl Responder {
    let db = state.db.lock().unwrap();
    match db.execute("DELETE FROM users WHERE id = ?1", params![path.into_inner()]) {
        Ok(0) => error(404, "user not found"),
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({ "deleted": true })),
        Err(_) => error(500, "user could not be deleted"),
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port = env::var("PORT").unwrap_or_else(|_| "3004".to_string());
    let db_path = env::var("DB_PATH").unwrap_or_else(|_| "benchmark/actix-web/data/db.sqlite".to_string());
    if let Some(parent) = Path::new(&db_path).parent() {
        fs::create_dir_all(parent)?;
    }
    let db = Connection::open(db_path).expect("open sqlite database");
    migrate(&db).expect("migrate sqlite database");
    let state = web::Data::new(State {
        db: Mutex::new(db),
        config: Config {
            service: env::var("SERVICE_NAME").unwrap_or_else(|_| "actix-web".to_string()),
            version: env::var("SERVICE_VERSION").unwrap_or_else(|_| "2.0.0".to_string()),
        },
    });
    HttpServer::new(move || App::new().app_data(state.clone()).service(health).service(create).service(list).service(get_user).service(update).service(delete_user))
        .bind(("0.0.0.0", port.parse::<u16>().unwrap_or(3004)))?
        .run()
        .await
}

fn migrate(db: &Connection) -> rusqlite::Result<()> {
    db.execute_batch(
        "PRAGMA journal_mode = WAL;
         PRAGMA synchronous = NORMAL;
         PRAGMA temp_store = MEMORY;
         PRAGMA busy_timeout = 5000;
         PRAGMA foreign_keys = ON;
         CREATE TABLE IF NOT EXISTS users (
           id INTEGER PRIMARY KEY AUTOINCREMENT,
           name TEXT NOT NULL,
           email TEXT NOT NULL UNIQUE,
           age INTEGER NOT NULL,
           created_at TEXT NOT NULL,
           updated_at TEXT NOT NULL
         );
         CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
         CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);",
    )
}

fn find_user(db: &Connection, id: i64) -> rusqlite::Result<Option<User>> {
    db.query_row("SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?1", params![id], read_user).optional()
}

fn read_user(row: &rusqlite::Row<'_>) -> rusqlite::Result<User> {
    Ok(User {
        id: row.get(0)?,
        name: row.get(1)?,
        email: row.get(2)?,
        age: row.get(3)?,
        created_at: row.get(4)?,
        updated_at: row.get(5)?,
    })
}

fn valid_payload(payload: &UserPayload) -> bool {
    !payload.name.trim().is_empty() && !payload.email.trim().is_empty() && payload.age > 0
}

fn error(status_code: u16, message: &'static str) -> HttpResponse {
    HttpResponse::build(actix_web::http::StatusCode::from_u16(status_code).unwrap()).json(ErrorBody { status_code, message })
}

fn timestamp() -> String {
    let seconds = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
    let datetime = chrono_like(seconds);
    format!("{}Z", datetime)
}

fn chrono_like(seconds: u64) -> String {
    let days = (seconds / 86_400) as i64;
    let secs = seconds % 86_400;
    let (year, month, day) = civil_from_days(days);
    format!("{:04}-{:02}-{:02}T{:02}:{:02}:{:02}", year, month, day, secs / 3600, (secs % 3600) / 60, secs % 60)
}

fn civil_from_days(days_since_epoch: i64) -> (i64, i64, i64) {
    let z = days_since_epoch + 719468;
    let era = if z >= 0 { z } else { z - 146096 } / 146097;
    let doe = z - era * 146097;
    let yoe = (doe - doe / 1460 + doe / 36524 - doe / 146096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = mp + if mp < 10 { 3 } else { -9 };
    (y + if m <= 2 { 1 } else { 0 }, m, d)
}
