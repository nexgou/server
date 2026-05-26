import { crudCycle, health } from '../lib/client.js';

export const options = {
	vus: 1,
	iterations: 1,
	thresholds: {
		checks: ['rate==1.0'],
		http_req_failed: ['rate==0'],
	},
};

export default function () {
	health();
	crudCycle(Date.now());
}
