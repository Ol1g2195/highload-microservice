import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
  vus: Number(__ENV.VUS || 10),
  duration: __ENV.DURATION || '1m',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const ADMIN_EMAIL = __ENV.ADMIN_EMAIL || 'admin@highload-microservice.local';
const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || 'admin123456';

export default function () {
  // Health
  const health = http.get(`${BASE_URL}/health`);
  check(health, {
    'health 200': (r) => r.status === 200,
  });

  // Try login to get token (best effort)
  let token = __ENV.TOKEN || '';
  if (!token) {
    const res = http.post(
      `${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
      { headers: { 'Content-Type': 'application/json' } },
    );
    if (res.status === 200) {
      try {
        token = res.json('access_token');
      } catch (_) {
        // ignore
      }
    }
  }

  // Users list (authorized if token present)
  group('users', () => {
    const params = token ? { headers: { Authorization: `Bearer ${token}` } } : {};
    const r = http.get(`${BASE_URL}/api/v1/users/?limit=10`, params);
    check(r, {
      'users list status ok': (resp) => [200, 401].includes(resp.status),
    });
  });

  // Events list (authorized if token present)
  group('events', () => {
    const params = token ? { headers: { Authorization: `Bearer ${token}` } } : {};
    const r = http.get(`${BASE_URL}/api/v1/events/?limit=10`, params);
    check(r, {
      'events list status ok': (resp) => [200, 401].includes(resp.status),
    });
  });

  sleep(1);
}


