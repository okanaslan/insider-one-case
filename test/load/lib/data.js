import exec from 'k6/execution';
import { getNowUnixSeconds } from './config.js';

const EVENT_NAMES = ['purchase', 'page_view', 'signup', 'add_to_cart'];
const CHANNELS = ['web', 'mobile', 'email'];
const CAMPAIGNS = ['cmp_1', 'cmp_2', 'cmp_3'];
const TAG_SETS = [
  ['promo'],
  ['summer'],
  ['new_user'],
  ['sale'],
];

function pickFrom(values, seed) {
  return values[seed % values.length];
}

export function makeEvent(sequence) {
  const now = getNowUnixSeconds();
  const vu = exec.vu.idInTest || 0;
  const iter = exec.scenario.iterationInTest || 0;
  const n = sequence || iter;

  return {
    event_name: pickFrom(EVENT_NAMES, n),
    channel: pickFrom(CHANNELS, n + vu),
    campaign_id: pickFrom(CAMPAIGNS, n + 1),
    user_id: `user_${vu}_${n}_${now}`,
    timestamp: now + n,
    tags: pickFrom(TAG_SETS, n + 2),
    metadata: {
      amount: 20 + (n % 500),
      source: pickFrom(CHANNELS, n + 3),
    },
  };
}

export function makeBulkEvents(size) {
  const events = [];
  for (let i = 0; i < size; i++) {
    events.push(makeEvent(i));
  }
  return events;
}
