import http from 'k6/http';
import { checkJSON, checkStatus } from './checks.js';
import { updatedUserPayload, userPayload } from './payloads.js';

export const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';

const JSON_HEADERS = { 'Content-Type': 'application/json' };

export function health() {
	const response = http.get(`${BASE_URL}/health`);
	checkStatus(response, 200, 'health');
	checkJSON(response, 'health');
	return response;
}

export function createUser(seed = Date.now()) {
	const response = http.post(
		`${BASE_URL}/users`,
		JSON.stringify(userPayload(seed)),
		{ headers: JSON_HEADERS },
	);
	checkStatus(response, 201, 'create user');
	checkJSON(response, 'create user');
	return response.json();
}

export function getUser(id) {
	const response = http.get(`${BASE_URL}/users/${id}`);
	checkStatus(response, 200, 'get user');
	checkJSON(response, 'get user');
	return response;
}

export function listUsers() {
	const response = http.get(`${BASE_URL}/users?limit=20&offset=0`);
	checkStatus(response, 200, 'list users');
	checkJSON(response, 'list users');
	return response;
}

export function updateUser(id) {
	const response = http.put(
		`${BASE_URL}/users/${id}`,
		JSON.stringify(updatedUserPayload(id)),
		{ headers: JSON_HEADERS },
	);
	checkStatus(response, 200, 'update user');
	checkJSON(response, 'update user');
	return response;
}

export function deleteUser(id) {
	const response = http.del(`${BASE_URL}/users/${id}`);
	checkStatus(response, 200, 'delete user');
	checkJSON(response, 'delete user');
	return response;
}

export function crudCycle(seed = Date.now()) {
	const created = createUser(seed);
	getUser(created.id);
	listUsers();
	updateUser(created.id);
	deleteUser(created.id);
}
