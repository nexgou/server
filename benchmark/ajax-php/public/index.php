<?php

$port = getenv('PORT') ?: '3007';
$dbPath = getenv('DB_PATH') ?: __DIR__ . '/../data/db.sqlite';
$serviceName = getenv('SERVICE_NAME') ?: 'ajax-php';
$serviceVersion = getenv('SERVICE_VERSION') ?: '2.0.0';

if (!is_dir(dirname($dbPath))) {
    mkdir(dirname($dbPath), 0777, true);
}

$db = new PDO('sqlite:' . $dbPath, null, null, [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]);
$db->exec('PRAGMA journal_mode = WAL');
$db->exec('PRAGMA synchronous = NORMAL');
$db->exec('PRAGMA temp_store = MEMORY');
$db->exec('PRAGMA busy_timeout = 5000');
$db->exec('PRAGMA foreign_keys = ON');
$db->exec('
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  age INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
');

function send_json(int $statusCode, array $body): void
{
    http_response_code($statusCode);
    header('Content-Type: application/json');
    echo json_encode($body, JSON_UNESCAPED_SLASHES);
}

function error_json(int $statusCode, string $message): void
{
    send_json($statusCode, ['statusCode' => $statusCode, 'message' => $message]);
}

function read_payload(): ?array
{
    $payload = json_decode(file_get_contents('php://input'), true);
    if (!is_array($payload) || !isset($payload['name'], $payload['email'], $payload['age'])) {
        return null;
    }
    $name = trim((string)$payload['name']);
    $email = trim((string)$payload['email']);
    $age = (int)$payload['age'];
    if ($name === '' || $email === '' || $age <= 0) {
        return null;
    }
    return ['name' => $name, 'email' => $email, 'age' => $age];
}

function user_row(array $row): array
{
    return [
        'id' => (int)$row['id'],
        'name' => $row['name'],
        'email' => $row['email'],
        'age' => (int)$row['age'],
        'createdAt' => $row['created_at'],
        'updatedAt' => $row['updated_at'],
    ];
}

$method = $_SERVER['REQUEST_METHOD'];
$path = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH) ?: '/';

if ($method === 'GET' && $path === '/health') {
    send_json(200, ['status' => 'ok', 'service' => $serviceName, 'version' => $serviceVersion]);
    return;
}

if ($method === 'POST' && $path === '/users') {
    $payload = read_payload();
    if ($payload === null) {
        error_json(400, 'name, email and age are required');
        return;
    }
    $now = gmdate('Y-m-d\TH:i:s\Z');
    try {
        $statement = $db->prepare('INSERT INTO users (name, email, age, created_at, updated_at) VALUES (?, ?, ?, ?, ?)');
        $statement->execute([$payload['name'], $payload['email'], $payload['age'], $now, $now]);
        $row = $db->query('SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ' . $db->lastInsertId())->fetch(PDO::FETCH_ASSOC);
        send_json(201, user_row($row));
    } catch (Throwable) {
        error_json(400, 'user could not be created');
    }
    return;
}

if ($method === 'GET' && $path === '/users') {
    $limit = max((int)($_GET['limit'] ?? 20), 1);
    $offset = max((int)($_GET['offset'] ?? 0), 0);
    $statement = $db->prepare('SELECT id, name, email, age, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?');
    $statement->bindValue(1, $limit, PDO::PARAM_INT);
    $statement->bindValue(2, $offset, PDO::PARAM_INT);
    $statement->execute();
    $items = array_map('user_row', $statement->fetchAll(PDO::FETCH_ASSOC));
    $total = (int)$db->query('SELECT COUNT(*) FROM users')->fetchColumn();
    send_json(200, ['items' => $items, 'limit' => $limit, 'offset' => $offset, 'total' => $total]);
    return;
}

if (preg_match('#^/users/(\d+)$#', $path, $matches)) {
    $id = $matches[1];
    if ($method === 'GET') {
        $statement = $db->prepare('SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?');
        $statement->execute([$id]);
        $row = $statement->fetch(PDO::FETCH_ASSOC);
        if (!$row) {
            error_json(404, 'user not found');
            return;
        }
        send_json(200, user_row($row));
        return;
    }
    if ($method === 'PUT') {
        $payload = read_payload();
        if ($payload === null) {
            error_json(400, 'name, email and age are required');
            return;
        }
        $now = gmdate('Y-m-d\TH:i:s\Z');
        try {
            $statement = $db->prepare('UPDATE users SET name = ?, email = ?, age = ?, updated_at = ? WHERE id = ?');
            $statement->execute([$payload['name'], $payload['email'], $payload['age'], $now, $id]);
            if ($statement->rowCount() === 0) {
                error_json(404, 'user not found');
                return;
            }
            $statement = $db->prepare('SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?');
            $statement->execute([$id]);
            send_json(200, user_row($statement->fetch(PDO::FETCH_ASSOC)));
        } catch (Throwable) {
            error_json(400, 'user could not be updated');
        }
        return;
    }
    if ($method === 'DELETE') {
        $statement = $db->prepare('DELETE FROM users WHERE id = ?');
        $statement->execute([$id]);
        if ($statement->rowCount() === 0) {
            error_json(404, 'user not found');
            return;
        }
        send_json(200, ['deleted' => true]);
        return;
    }
}

error_json(404, 'not found');
