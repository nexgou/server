use chrono::Utc;
use hyper::body::to_bytes;
use hyper::header::CONTENT_TYPE;
use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Method, Request, Response, Server, StatusCode};
use rusqlite::{params, Connection, OptionalExtension};
use serde::{Deserialize, Serialize};
use std::convert::Infallible;
use std::env;
use std::fs;
use std::net::SocketAddr;
use std::path::Path;
use std::sync::{Arc, Mutex};
use url::form_urlencoded;

#[derive(Clone)]
struct Config {
    service: String,
    version: String,
}

struct State {
    db: Mutex<Connection>,
    config: Config,
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

#[derive(Serialize)]
struct ErrorBody<'a> {
    #[serde(rename = "statusCode")]
    status_code: u16,
    message: &'a str,
}

#[tokio::main]
async fn main() {
    let port = env::var("PORT").unwrap_or_else(|_| "3005".to_string()).parse::<u16>().unwrap_or(3005);
    let db_path = env::var("DB_PATH").unwrap_or_else(|_| "benchmark/hyper/data/db.sqlite".to_string());
    if let Some(parent) = Path::new(&db_path).parent() {
        fs::create_dir_all(parent).expect("create sqlite directory");
    }
    let db = Connection::open(db_path).expect("open sqlite database");
    migrate(&db).expect("migrate sqlite database");
    let state = Arc::new(State {
        db: Mutex::new(db),
        config: Config {
            service: env::var("SERVICE_NAME").unwrap_or_else(|_| "hyper".to_string()),
            version: env::var("SERVICE_VERSION").unwrap_or_else(|_| "2.0.0".to_string()),
        },
    });
    let make_service = make_service_fn(move |_| {
        let state = state.clone();
        async move { Ok::<_, Infallible>(service_fn(move |request| route(request, state.clone()))) }
    });
    let address = SocketAddr::from(([0, 0, 0, 0], port));
    Server::bind(&address).serve(make_service).await.expect("run hyper server");
}

async fn route(request: Request<Body>, state: Arc<State>) -> Result<Response<Body>, Infallible> {
    let method = request.method().clone();
    let uri = request.uri().clone();
    let path = uri.path().to_string();
    let response = match (method, path.as_str()) {
        (Method::GET, "/health") => json(StatusCode::OK, serde_json::json!({ "status": "ok", "service": state.config.service, "version": state.config.version })),
        (Method::GET, "/plaintext") => plaintext(),
        (Method::GET, "/json") => json(StatusCode::OK, serde_json::json!({ "message": "Hello, World!" })),
        (Method::GET, "/middleware") => raw_middleware(state),
        (Method::POST, "/users") => create(request, state).await,
        (Method::GET, "/users") => list(&uri, state),
        _ if path.starts_with("/params/") => raw_params(&path),
        _ => match parse_user_id(&path) {
            Some(id) => match *request.method() {
                Method::GET => get_user(id, state),
                Method::PUT => update(id, request, state).await,
                Method::DELETE => delete_user(id, state),
                _ => error(StatusCode::NOT_FOUND, "not found"),
            },
            None => error(StatusCode::NOT_FOUND, "not found"),
        },
    };
    Ok(response)
}

async fn create(request: Request<Body>, state: Arc<State>) -> Response<Body> {
    let payload = match parse_payload(request).await {
        Some(payload) if valid_payload(&payload) => payload,
        _ => return error(StatusCode::BAD_REQUEST, "name, email and age are required"),
    };
    let db = state.db.lock().unwrap();
    let now = timestamp();
    let inserted = db.execute(
        "INSERT INTO users (name, email, age, created_at, updated_at) VALUES (?1, ?2, ?3, ?4, ?5)",
        params![payload.name.trim(), payload.email.trim(), payload.age, now, now],
    );
    if inserted.is_err() {
        return error(StatusCode::BAD_REQUEST, "user could not be created");
    }
    match find_user(&db, db.last_insert_rowid()).unwrap_or(None) {
        Some(user) => json(StatusCode::CREATED, user),
        None => error(StatusCode::INTERNAL_SERVER_ERROR, "user could not be fetched"),
    }
}

fn list(uri: &hyper::Uri, state: Arc<State>) -> Response<Body> {
    let mut limit = 20;
    let mut offset = 0;
    if let Some(query) = uri.query() {
        for (key, value) in form_urlencoded::parse(query.as_bytes()) {
            if key == "limit" {
                limit = value.parse::<i64>().unwrap_or(20).max(1);
            }
            if key == "offset" {
                offset = value.parse::<i64>().unwrap_or(0).max(0);
            }
        }
    }
    let db = state.db.lock().unwrap();
    let mut statement = match db.prepare("SELECT id, name, email, age, created_at, updated_at FROM users ORDER BY id DESC LIMIT ?1 OFFSET ?2") {
        Ok(statement) => statement,
        Err(_) => return error(StatusCode::INTERNAL_SERVER_ERROR, "users could not be listed"),
    };
    let rows = match statement.query_map(params![limit, offset], read_user) {
        Ok(rows) => rows,
        Err(_) => return error(StatusCode::INTERNAL_SERVER_ERROR, "users could not be listed"),
    };
    let mut items = Vec::new();
    for row in rows {
        match row {
            Ok(user) => items.push(user),
            Err(_) => return error(StatusCode::INTERNAL_SERVER_ERROR, "users could not be listed"),
        }
    }
    let total = db.query_row("SELECT COUNT(*) FROM users", [], |row| row.get(0)).unwrap_or(0);
    json(StatusCode::OK, UserList { items, limit, offset, total })
}

fn get_user(id: i64, state: Arc<State>) -> Response<Body> {
    let db = state.db.lock().unwrap();
    match find_user(&db, id).unwrap_or(None) {
        Some(user) => json(StatusCode::OK, user),
        None => error(StatusCode::NOT_FOUND, "user not found"),
    }
}

async fn update(id: i64, request: Request<Body>, state: Arc<State>) -> Response<Body> {
    let payload = match parse_payload(request).await {
        Some(payload) if valid_payload(&payload) => payload,
        _ => return error(StatusCode::BAD_REQUEST, "name, email and age are required"),
    };
    let db = state.db.lock().unwrap();
    match db.execute(
        "UPDATE users SET name = ?1, email = ?2, age = ?3, updated_at = ?4 WHERE id = ?5",
        params![payload.name.trim(), payload.email.trim(), payload.age, timestamp(), id],
    ) {
        Ok(0) => error(StatusCode::NOT_FOUND, "user not found"),
        Ok(_) => match find_user(&db, id).unwrap_or(None) {
            Some(user) => json(StatusCode::OK, user),
            None => error(StatusCode::NOT_FOUND, "user not found"),
        },
        Err(_) => error(StatusCode::BAD_REQUEST, "user could not be updated"),
    }
}

fn delete_user(id: i64, state: Arc<State>) -> Response<Body> {
    let db = state.db.lock().unwrap();
    match db.execute("DELETE FROM users WHERE id = ?1", params![id]) {
        Ok(0) => error(StatusCode::NOT_FOUND, "user not found"),
        Ok(_) => json(StatusCode::OK, serde_json::json!({ "deleted": true })),
        Err(_) => error(StatusCode::INTERNAL_SERVER_ERROR, "user could not be deleted"),
    }
}

async fn parse_payload(request: Request<Body>) -> Option<UserPayload> {
    let body = to_bytes(request.into_body()).await.ok()?;
    serde_json::from_slice::<UserPayload>(&body).ok()
}

fn parse_user_id(path: &str) -> Option<i64> {
    let prefix = "/users/";
    path.strip_prefix(prefix).and_then(|value| value.parse::<i64>().ok())
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

fn json<T: Serialize>(status: StatusCode, body: T) -> Response<Body> {
    Response::builder()
        .status(status)
        .header(CONTENT_TYPE, "application/json")
        .body(Body::from(serde_json::to_vec(&body).unwrap()))
        .unwrap()
}

fn plaintext() -> Response<Body> {
    Response::builder()
        .status(StatusCode::OK)
        .header(CONTENT_TYPE, "text/plain; charset=utf-8")
        .body(Body::from("Hello, World!"))
        .unwrap()
}

fn raw_params(path: &str) -> Response<Body> {
    let id = path.trim_start_matches("/params/");
    json(StatusCode::OK, serde_json::json!({ "id": id, "echo": "value" }))
}

fn raw_middleware(state: Arc<State>) -> Response<Body> {
    Response::builder()
        .status(StatusCode::OK)
        .header(CONTENT_TYPE, "application/json")
        .header("X-Raw-Middleware", "true")
        .body(Body::from(serde_json::to_vec(&serde_json::json!({
            "service": state.config.service,
            "version": state.config.version,
            "guard": true,
            "interceptor": true
        })).unwrap()))
        .unwrap()
}

fn error(status: StatusCode, message: &'static str) -> Response<Body> {
    json(status, ErrorBody { status_code: status.as_u16(), message })
}

fn timestamp() -> String {
    Utc::now().format("%Y-%m-%dT%H:%M:%SZ").to_string()
}
