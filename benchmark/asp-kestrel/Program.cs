using Microsoft.Data.Sqlite;

var port = Environment.GetEnvironmentVariable("PORT") ?? "3003";
var dbPath = Environment.GetEnvironmentVariable("DB_PATH") ?? Path.Combine("benchmark", "asp-kestrel", "data", "db.sqlite");
var serviceName = Environment.GetEnvironmentVariable("SERVICE_NAME") ?? "asp-kestrel";
var serviceVersion = Environment.GetEnvironmentVariable("SERVICE_VERSION") ?? "2.0.0";

Directory.CreateDirectory(Path.GetDirectoryName(Path.GetFullPath(dbPath))!);

var builder = WebApplication.CreateBuilder(args);
builder.WebHost.UseUrls($"http://0.0.0.0:{port}");
var app = builder.Build();

var store = new Store(dbPath);
store.Migrate();

app.MapGet("/health", () => Results.Json(new { status = "ok", service = serviceName, version = serviceVersion }));

app.MapGet("/plaintext", () => Results.Text("Hello, World!", "text/plain"));

app.MapGet("/json", () => Results.Json(new { message = "Hello, World!" }));

app.MapGet("/params/{id}", (string id) => Results.Json(new { id, echo = "value" }));

app.MapGet("/middleware", (HttpContext context) =>
{
    context.Response.Headers["X-Raw-Middleware"] = "true";
    return Results.Json(new { service = serviceName, version = serviceVersion, guard = true, interceptor = true });
});

app.MapPost("/users", (UserPayload payload) =>
{
    if (!payload.Valid())
    {
        return Error(400, "name, email and age are required");
    }
    try
    {
        return Results.Json(store.Create(payload), statusCode: 201);
    }
    catch
    {
        return Error(400, "user could not be created");
    }
});

app.MapGet("/users", (int? limit, int? offset) =>
{
    try
    {
        return Results.Json(store.List(Math.Max(limit ?? 20, 1), Math.Max(offset ?? 0, 0)));
    }
    catch
    {
        return Error(500, "users could not be listed");
    }
});

app.MapGet("/users/{id:long}", (long id) =>
{
    var user = store.Find(id);
    return user is null ? Error(404, "user not found") : Results.Json(user);
});

app.MapPut("/users/{id:long}", (long id, UserPayload payload) =>
{
    if (!payload.Valid())
    {
        return Error(400, "name, email and age are required");
    }
    try
    {
        var user = store.Update(id, payload);
        return user is null ? Error(404, "user not found") : Results.Json(user);
    }
    catch
    {
        return Error(400, "user could not be updated");
    }
});

app.MapDelete("/users/{id:long}", (long id) => store.Delete(id) ? Results.Json(new { deleted = true }) : Error(404, "user not found"));

app.Run();

static IResult Error(int statusCode, string message) => Results.Json(new { statusCode, message }, statusCode: statusCode);

sealed record User(long Id, string Name, string Email, int Age, string CreatedAt, string UpdatedAt);
sealed record UserPayload(string Name, string Email, int Age)
{
    public bool Valid() => !string.IsNullOrWhiteSpace(Name) && !string.IsNullOrWhiteSpace(Email) && Age > 0;
}
sealed record UserList(IReadOnlyList<User> Items, int Limit, int Offset, int Total);

sealed class Store
{
    private readonly string connectionString;
    private readonly object gate = new();

    public Store(string dbPath)
    {
        connectionString = new SqliteConnectionStringBuilder { DataSource = dbPath }.ToString();
    }

    public void Migrate()
    {
        lock (gate)
        {
            using var connection = Open();
            Execute(connection, "PRAGMA journal_mode = WAL");
            Execute(connection, "PRAGMA synchronous = NORMAL");
            Execute(connection, "PRAGMA temp_store = MEMORY");
            Execute(connection, "PRAGMA busy_timeout = 5000");
            Execute(connection, "PRAGMA foreign_keys = ON");
            Execute(connection, @"
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  age INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);");
        }
    }

    public User Create(UserPayload payload)
    {
        lock (gate)
        {
            using var connection = Open();
            var now = Timestamp();
            using var command = connection.CreateCommand();
            command.CommandText = "INSERT INTO users (name, email, age, created_at, updated_at) VALUES ($name, $email, $age, $created, $updated); SELECT last_insert_rowid();";
            command.Parameters.AddWithValue("$name", payload.Name.Trim());
            command.Parameters.AddWithValue("$email", payload.Email.Trim());
            command.Parameters.AddWithValue("$age", payload.Age);
            command.Parameters.AddWithValue("$created", now);
            command.Parameters.AddWithValue("$updated", now);
            var id = (long)(command.ExecuteScalar() ?? 0L);
            return Find(connection, id)!;
        }
    }

    public User? Find(long id)
    {
        lock (gate)
        {
            using var connection = Open();
            return Find(connection, id);
        }
    }

    public UserList List(int limit, int offset)
    {
        lock (gate)
        {
            using var connection = Open();
            using var command = connection.CreateCommand();
            command.CommandText = "SELECT id, name, email, age, created_at, updated_at FROM users ORDER BY id DESC LIMIT $limit OFFSET $offset";
            command.Parameters.AddWithValue("$limit", limit);
            command.Parameters.AddWithValue("$offset", offset);
            var items = new List<User>();
            using (var reader = command.ExecuteReader())
            {
                while (reader.Read())
                {
                    items.Add(ReadUser(reader));
                }
            }
            using var count = connection.CreateCommand();
            count.CommandText = "SELECT COUNT(*) FROM users";
            return new UserList(items, limit, offset, Convert.ToInt32(count.ExecuteScalar()));
        }
    }

    public User? Update(long id, UserPayload payload)
    {
        lock (gate)
        {
            using var connection = Open();
            using var command = connection.CreateCommand();
            command.CommandText = "UPDATE users SET name = $name, email = $email, age = $age, updated_at = $updated WHERE id = $id";
            command.Parameters.AddWithValue("$name", payload.Name.Trim());
            command.Parameters.AddWithValue("$email", payload.Email.Trim());
            command.Parameters.AddWithValue("$age", payload.Age);
            command.Parameters.AddWithValue("$updated", Timestamp());
            command.Parameters.AddWithValue("$id", id);
            if (command.ExecuteNonQuery() == 0)
            {
                return null;
            }
            return Find(connection, id);
        }
    }

    public bool Delete(long id)
    {
        lock (gate)
        {
            using var connection = Open();
            using var command = connection.CreateCommand();
            command.CommandText = "DELETE FROM users WHERE id = $id";
            command.Parameters.AddWithValue("$id", id);
            return command.ExecuteNonQuery() > 0;
        }
    }

    private SqliteConnection Open()
    {
        var connection = new SqliteConnection(connectionString);
        connection.Open();
        return connection;
    }

    private static User? Find(SqliteConnection connection, long id)
    {
        using var command = connection.CreateCommand();
        command.CommandText = "SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = $id";
        command.Parameters.AddWithValue("$id", id);
        using var reader = command.ExecuteReader();
        return reader.Read() ? ReadUser(reader) : null;
    }

    private static User ReadUser(SqliteDataReader reader) => new(
        reader.GetInt64(0),
        reader.GetString(1),
        reader.GetString(2),
        reader.GetInt32(3),
        reader.GetString(4),
        reader.GetString(5));

    private static void Execute(SqliteConnection connection, string sql)
    {
        using var command = connection.CreateCommand();
        command.CommandText = sql;
        command.ExecuteNonQuery();
    }

    private static string Timestamp() => DateTime.UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ");
}
