export function userPayload(seed = Date.now()) {
	return {
		name: `Benchmark User ${seed}`,
		email: `bench-${seed}-${__VU}-${__ITER}@example.com`,
		age: 30 + (seed % 20),
	};
}

export function updatedUserPayload(id) {
	return {
		name: `Benchmark User ${id} Updated`,
		email: `bench-${id}-updated-${__VU}-${__ITER}@example.com`,
		age: 40,
	};
}
