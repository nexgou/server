import { crudCycle, health, listUsers } from '../lib/client.js';

export const options = {
	scenarios: {
		mixed: {
			executor: 'constant-vus',
			vus: Number(__ENV.VUS || 20),
			duration: __ENV.DURATION || '30s',
		},
	},
	thresholds: {
		checks: ['rate>0.99'],
		http_req_failed: ['rate<0.01'],
		http_req_duration: ['p(95)<500'],
	},
};

export default function () {
	health();
	if (__ITER % 3 === 0) {
		listUsers();
		return;
	}
	crudCycle(Date.now());
}
