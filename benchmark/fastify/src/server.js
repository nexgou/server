import Fastify from 'fastify';
import Database from 'better-sqlite3';
import { dirname, resolve } from 'node:path';
import { mkdirSync } from 'node:fs';

const port = Number(process.env.PORT || 3002);
const dbPath = process.env.DB_PATH || 'benchmark/fastify/data/db.sqlite';
const serviceName = process.env.SERVICE_NAME || 'fastify';
const serviceVersion = process.env.SERVICE_VERSION || '2.0.0';
const logLevel = process.env.LOG_LEVEL === 'silent' ? false : true;

mkdirSync(dirname(resolve(dbPath)), { recursive: true });

const db = new Database(dbPath);
db.pragma('journal_mode = WAL');
db.pragma('synchronous = NORMAL');
db.pragma('temp_store = MEMORY');
db.pragma('busy_timeout = 5000');
db.pragma('foreign_keys = ON');
db.exec(`
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
`);

const app = Fastify({ logger: logLevel });

const toUser = (row) => ({
  id: row.id,
  name: row.name,
  email: row.email,
  age: row.age,
  createdAt: row.created_at,
  updatedAt: row.updated_at,
});

const errorBody = (statusCode, message) => ({ statusCode, message });

function readPayload(body) {
  if (!body || typeof body.name !== 'string' || typeof body.email !== 'string' || !Number.isInteger(Number(body.age))) {
    return null;
  }
  const payload = { name: body.name.trim(), email: body.email.trim(), age: Number(body.age) };
  if (!payload.name || !payload.email || payload.age <= 0) {
    return null;
  }
  return payload;
}

app.get('/health', async () => ({ status: 'ok', service: serviceName, version: serviceVersion }));

app.post('/users', async (request, reply) => {
  const payload = readPayload(request.body);
  if (!payload) {
    return reply.code(400).send(errorBody(400, 'name, email and age are required'));
  }
  const now = new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
  try {
    const result = db.prepare('INSERT INTO users (name, email, age, created_at, updated_at) VALUES (?, ?, ?, ?, ?)')
      .run(payload.name, payload.email, payload.age, now, now);
    const row = db.prepare('SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?').get(result.lastInsertRowid);
    return reply.code(201).send(toUser(row));
  } catch {
    return reply.code(400).send(errorBody(400, 'user could not be created'));
  }
});

app.get('/users', async (request, reply) => {
  const limit = Math.max(Number(request.query.limit || 20), 1);
  const offset = Math.max(Number(request.query.offset || 0), 0);
  try {
    const rows = db.prepare('SELECT id, name, email, age, created_at, updated_at FROM users ORDER BY id DESC LIMIT ? OFFSET ?').all(limit, offset);
    const total = db.prepare('SELECT COUNT(*) AS total FROM users').get().total;
    return { items: rows.map(toUser), limit, offset, total };
  } catch {
    return reply.code(500).send(errorBody(500, 'users could not be listed'));
  }
});

app.get('/users/:id', async (request, reply) => {
  const row = db.prepare('SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?').get(request.params.id);
  if (!row) {
    return reply.code(404).send(errorBody(404, 'user not found'));
  }
  return toUser(row);
});

app.put('/users/:id', async (request, reply) => {
  const payload = readPayload(request.body);
  if (!payload) {
    return reply.code(400).send(errorBody(400, 'name, email and age are required'));
  }
  const now = new Date().toISOString().replace(/\.\d{3}Z$/, 'Z');
  try {
    const result = db.prepare('UPDATE users SET name = ?, email = ?, age = ?, updated_at = ? WHERE id = ?')
      .run(payload.name, payload.email, payload.age, now, request.params.id);
    if (result.changes === 0) {
      return reply.code(404).send(errorBody(404, 'user not found'));
    }
    const row = db.prepare('SELECT id, name, email, age, created_at, updated_at FROM users WHERE id = ?').get(request.params.id);
    return toUser(row);
  } catch {
    return reply.code(400).send(errorBody(400, 'user could not be updated'));
  }
});

app.delete('/users/:id', async (request, reply) => {
  const result = db.prepare('DELETE FROM users WHERE id = ?').run(request.params.id);
  if (result.changes === 0) {
    return reply.code(404).send(errorBody(404, 'user not found'));
  }
  return { deleted: true };
});

await app.listen({ host: '0.0.0.0', port });
