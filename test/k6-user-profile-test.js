import http from 'k6/http';
import { check } from 'k6';

// ============================================================
// CONFIG — override any of these at run time, e.g.:
//   k6 run -e RPS=200 -e DURATION=2m k6-user-profile-test.js
// ============================================================
const BASE_URL = __ENV.BASE_URL || 'https://deployd-app-storage.desain.gratis';
const RPS = Number(__ENV.RPS) || 100;                 // <-- constant requests/sec, change this
const DURATION = __ENV.DURATION || '1m';               // test duration
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS) || Math.min(Math.max(RPS * 2, 10), 1000);
const MAX_VUS = Number(__ENV.MAX_VUS) || Math.min(Math.max(RPS * 4, 20), 2000);

export const options = {
  scenarios: {
    constant_rps: {
      executor: 'constant-arrival-rate',
      rate: RPS,
      timeUnit: '1s',
      duration: DURATION,
      preAllocatedVUs: PRE_ALLOCATED_VUS, // VUs pre-spun to sustain the rate
      maxVUs: MAX_VUS,                    // hard ceiling k6 can scale up to
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1000'],
  },
};

// ============================================================
// SAMPLE DATA GENERATION — 1000 unique UserProfile records
// Matches:
//   type UserProfile struct {
//     Ns          string   `json:"namespace"`
//     Id          string   `json:"id"`
//     Name        string   `json:"name"`
//     NickName    string   `json:"nick_name"`
//     Hobby       []string `json:"hobby"`
//     PublishedAt time.Time`json:"published_at"`
//     URLx        string   `json:"url"`
//   }
// ============================================================

// Deterministic pseudo-random generator so the dataset is
// reproducible across runs (same seed -> same 1000 records).
function seededRandom(seed) {
  let s = seed;
  return function () {
    s = (s * 9301 + 49297) % 233280;
    return s / 233280;
  };
}

function uuidv4(rand) {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (rand() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

const FIRST_NAMES = [
  'Aditya', 'Bella', 'Citra', 'Dimas', 'Eka', 'Farah', 'Gilang', 'Hana',
  'Indra', 'Joko', 'Kirana', 'Lestari', 'Made', 'Nadia', 'Oscar', 'Putri',
  'Qori', 'Rangga', 'Sari', 'Tono', 'Umar', 'Vina', 'Wibowo', 'Yuni', 'Zaki',
];
const LAST_NAMES = [
  'Saputra', 'Wijaya', 'Kusuma', 'Pratama', 'Santoso', 'Hidayat',
  'Permata', 'Nugraha', 'Setiawan', 'Wulandari',
];
const HOBBIES = [
  'reading', 'gaming', 'cycling', 'swimming', 'cooking', 'photography',
  'hiking', 'painting', 'coding', 'traveling', 'fishing', 'running',
  'yoga', 'chess', 'gardening', 'dancing', 'writing', 'singing',
  'diving', 'climbing',
];
const NAMESPACES = [
  'prod', 'staging', 'dev', 'qa', 'sandbox',
  'test-ns-1', 'test-ns-2', 'internal', 'partner-a', 'partner-b',
];
const DOMAINS = [
  'example.com', 'profile-cdn.com', 'userdata.io', 'assets.dev', 'media.local',
];

function randomItem(rand, arr) {
  return arr[Math.floor(rand() * arr.length)];
}

function randomHobbies(rand) {
  const count = 1 + Math.floor(rand() * 3); // 1-3 hobbies
  const shuffled = [...HOBBIES].sort(() => rand() - 0.5);
  return shuffled.slice(0, count);
}

function randomPastISODate(rand) {
  const now = Date.now();
  const threeYearsMs = 1000 * 60 * 60 * 24 * 365 * 3;
  const past = now - Math.floor(rand() * threeYearsMs);
  return new Date(past).toISOString(); // RFC3339 - unmarshals fine into Go time.Time
}

function generateProfiles(count) {
  const rand = seededRandom(42);
  const profiles = [];
  for (let i = 0; i < count; i++) {
    const first = randomItem(rand, FIRST_NAMES);
    const last = randomItem(rand, LAST_NAMES);
    const name = `${first} ${last}`;
    const nick = `${first.toLowerCase()}${i}`;
    profiles.push({
      namespace: randomItem(rand, NAMESPACES),
      id: uuidv4(rand),
      name: name,
      nick_name: nick,
      hobby: randomHobbies(rand),
      published_at: randomPastISODate(rand),
      url: `https://${randomItem(rand, DOMAINS)}/u/${nick}-${i}.png`,
    });
  }
  return profiles;
}

// Generated once at init time, shared (read-only) across all VUs.
const PROFILES = generateProfiles(1000);

// ============================================================
// TEST LOGIC
// ============================================================
export default function () {
  const profile = PROFILES[Math.floor(Math.random() * PROFILES.length)];

  // Give every single request a fresh unique id so repeated
  // sampling of the same base record doesn't collide server-side.
  const payload = Object.assign({}, profile, {
    id: `${profile.id}-${__VU}-${__ITER}`,
  });

  const res = http.post(`${BASE_URL}/user-profile`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });
}