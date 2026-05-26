import { check } from 'k6';

export function checkStatus(response, status, label) {
	return check(response, {
		[`${label}: status ${status}`]: (res) => res.status === status,
	});
}

export function checkJSON(response, label) {
	return check(response, {
		[`${label}: valid json`]: (res) => {
			try {
				res.json();
				return true;
			} catch (_) {
				return false;
			}
		},
	});
}
