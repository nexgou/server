package com.nexgou.benchmark;

import io.vertx.core.Vertx;
import io.vertx.core.json.JsonArray;
import io.vertx.core.json.JsonObject;
import io.vertx.ext.web.Router;
import io.vertx.ext.web.RoutingContext;
import io.vertx.ext.web.handler.BodyHandler;

import java.nio.file.Files;
import java.nio.file.Path;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.sql.Statement;
import java.time.Instant;

public final class App {
    private final Connection db;
    private final String service;
    private final String version;

    private App(Connection db, String service, String version) {
        this.db = db;
        this.service = service;
        this.version = version;
    }

    public static void main(String[] args) throws Exception {
        int port = Integer.parseInt(env("PORT", "3006"));
        String dbPath = env("DB_PATH", "benchmark/vert-x/data/db.sqlite");
        Path parent = Path.of(dbPath).toAbsolutePath().getParent();
        if (parent != null) {
            Files.createDirectories(parent);
        }
        Connection db = DriverManager.getConnection("jdbc:sqlite:" + dbPath);
        App app = new App(db, env("SERVICE_NAME", "vert-x"), env("SERVICE_VERSION", "2.0.0"));
        app.migrate();

        Vertx vertx = Vertx.vertx();
        Router router = Router.router(vertx);
        router.route().handler(BodyHandler.create());
        router.get("/health").handler(app::health);
        router.post("/users").handler(app::create);
        router.get("/users").handler(app::list);
        router.get("/users/:id").handler(app::find);
        router.put("/users/:id").handler(app::update);
        router.delete("/users/:id").handler(app::delete);
        vertx.createHttpServer().requestHandler(router).listen(port, "0.0.0.0").onFailure(Throwable::printStackTrace);
    }

    private void health(RoutingContext context) {
        json(context, 200, new JsonObject().put("status", "ok").put("service", service).put("version", version));
    }

    private void create(RoutingContext context) {
        JsonObject payload = context.body().asJsonObject();
        if (!valid(payload)) {
            error(context, 400, "name, email and age are required");
            return;
        }
        synchronized (db) {
            try {
                String now = timestamp();
                try (PreparedStatement statement = db.prepareStatement("INSERT INTO users (name, email, age, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", Statement.RETURN_GENERATED_KEYS)) {
                    statement.setString(1, payload.getString("name").trim());
                    statement.setString(2, payload.getString("email").trim());
                    statement.setInt(3, payload.getInteger("age"));
                    statement.setString(4, now);
                    statement.setString(5, now);
                    statement.executeUpdate();
                    try (ResultSet keys = statement.getGeneratedKeys()) {
                        keys.next();
                        json(context, 201, findUser(keys.getLong(1)));
                    }
                }
            } catch (SQLException exception) {
                error(context, 400, "user could not be created");
            }
        }
    }

    private void list(RoutingContext context) {
        int limit = Math.max(parseInt(context.queryParam("limit").stream().findFirst().orElse("20"), 20), 1);
        int offset = Math.max(parseInt(context.queryParam("offset").stream().findFirst().orElse("0"), 0), 0);
        synchronized (db) {
            try (PreparedStatement statement = db.prepareStatement("SELECT id, name, email, age, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?")) {
                statement.setInt(1, limit);
                statement.setInt(2, offset);
                JsonArray items = new JsonArray();
                try (ResultSet rows = statement.executeQuery()) {
                    while (rows.next()) {
                        items.add(readUser(rows));
                    }
                }
                int total;
                try (Statement count = db.createStatement(); ResultSet row = count.executeQuery("SELECT COUNT(*) FROM users")) {
                    row.next();
                    total = row.getInt(1);
                }
                json(context, 200, new JsonObject().put("items", items).put("limit", limit).put("offset", offset).put("total", total));
            } catch (SQLException exception) {
                error(context, 500, "users could not be listed");
            }
        }
    }

    private void find(RoutingContext context) {
        long id = Long.parseLong(context.pathParam("id"));
        synchronized (db) {
            try {
                JsonObject user = findUser(id);
                if (user == null) {
                    error(context, 404, "user not found");
                    return;
                }
                json(context, 200, user);
            } catch (SQLException exception) {
                error(context, 500, "user could not be fetched");
            }
        }
    }

    private void update(RoutingContext context) {
        JsonObject payload = context.body().asJsonObject();
        if (!valid(payload)) {
            error(context, 400, "name, email and age are required");
            return;
        }
        long id = Long.parseLong(context.pathParam("id"));
        synchronized (db) {
            try (PreparedStatement statement = db.prepareStatement("UPDATE users SET name = ?, email = ?, age = ?, updated_at = ? WHERE id = ?")) {
                statement.setString(1, payload.getString("name").trim());
                statement.setString(2, payload.getString("email").trim());
                statement.setInt(3, payload.getInteger("age"));
                statement.setString(4, timestamp());
                statement.setLong(5, id);
                if (statement.executeUpdate() == 0) {
                    error(context, 404, "user not found");
                    return;
                }
                json(context, 200, findUser(id));
            } catch (SQLException exception) {
                error(context, 400, "user could not be updated");
            }
        }
    }

    private void delete(RoutingContext context) {
        long id = Long.parseLong(context.pathParam("id"));
        synchronized (db) {
            try (PreparedStatement statement = db.prepareStatement("DELETE FROM users WHERE id = ?")) {
                statement.setLong(1, id);
                if (statement.executeUpdate() == 0) {
                    error(context, 404, "user not found");
                    return;
                }
                json(context, 200, new JsonObject().put("deleted", true));
            } catch (SQLException exception) {
                error(context, 500, "user could not be deleted");
            }
        }
    }

    private void migrate() throws SQLException {
        try (Statement statement = db.createStatement()) {
            statement.execute("PRAGMA journal_mode = WAL");
            statement.execute("PRAGMA synchronous = NORMAL");
            statement.execute("PRAGMA temp_store = MEMORY");
            statement.execute("PRAGMA busy_timeout = 5000");
            statement.execute("PRAGMA foreign_keys = ON");
            statement.execute("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, email TEXT NOT NULL UNIQUE, age INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)");
            statement.execute("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)");
            statement.execute("CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)");
        }
    }

    private JsonObject findUser(long id) throws SQLException {
        try (PreparedStatement statement = db.prepareStatement("SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?")) {
            statement.setLong(1, id);
            try (ResultSet row = statement.executeQuery()) {
                return row.next() ? readUser(row) : null;
            }
        }
    }

    private static JsonObject readUser(ResultSet row) throws SQLException {
        return new JsonObject()
            .put("id", row.getLong("id"))
            .put("name", row.getString("name"))
            .put("email", row.getString("email"))
            .put("age", row.getInt("age"))
            .put("createdAt", row.getString("created_at"))
            .put("updatedAt", row.getString("updated_at"));
    }

    private static boolean valid(JsonObject payload) {
        return payload != null
            && payload.getString("name", "").trim().length() > 0
            && payload.getString("email", "").trim().length() > 0
            && payload.getInteger("age", 0) > 0;
    }

    private static void json(RoutingContext context, int statusCode, Object body) {
        context.response().setStatusCode(statusCode).putHeader("Content-Type", "application/json").end(body.toString());
    }

    private static void error(RoutingContext context, int statusCode, String message) {
        json(context, statusCode, new JsonObject().put("statusCode", statusCode).put("message", message));
    }

    private static String timestamp() {
        return Instant.now().toString().replaceFirst("\\.\\d+Z$", "Z");
    }

    private static String env(String key, String fallback) {
        String value = System.getenv(key);
        return value == null || value.isBlank() ? fallback : value;
    }

    private static int parseInt(String value, int fallback) {
        try {
            return Integer.parseInt(value);
        } catch (NumberFormatException exception) {
            return fallback;
        }
    }
}
