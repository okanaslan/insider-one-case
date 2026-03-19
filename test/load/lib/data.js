import exec from 'k6/execution';
import { getNowUnixSeconds } from './config.js';

const EVENT_NAMES = ['purchase', 'page_view', 'signup', 'add_to_cart', 'click', 'view_content', 'search'];
const CHANNELS = ['web', 'mobile', 'email', 'social', 'ads'];
const CAMPAIGNS = ['cmp_1', 'cmp_2', 'cmp_3', 'cmp_4', 'cmp_5', 'cmp_6', 'cmp_7', 'cmp_8', 'cmp_9', 'cmp_10'];
const TAG_SETS = [
    ['promo', 'summer', 'new_user'],
    ['summer', 'clearance'],
    ['new_user', 'referral'],
    ['sale', 'promo'],
    ['clearance', 'referral'],
    ['summer', 'sale'],
    ['promo', 'referral'],
    ['new_user', 'summer'],
    ['sale', 'clearance'],
    ['referral', 'promo'],
];

function pickFrom(values, seed) {
    return values[seed % values.length];
}

// Make event with intentional duplicates
export function makeEvent(sequence, forceDedup = true) {
    const now = getNowUnixSeconds();
    const vu = exec.vu.idInTest || 0;
    const iter = exec.scenario.iterationInTest || 0;
    const n = sequence || iter;

    // If dedup testing: reuse user_id and same timestamp within windows
    const userId = forceDedup
        ? `user_${vu}_${Math.floor(n / 10)}` // Same ID for every 10 iterations
        : `user_${vu}_${n}_${now}`;

    return {
        event_name: pickFrom(EVENT_NAMES, n),
        channel: pickFrom(CHANNELS, n + vu),
        campaign_id: pickFrom(CAMPAIGNS, n + 1),
        user_id: userId,
        timestamp: forceDedup ? now : now + n, // Same timestamp for dedup test
        tags: pickFrom(TAG_SETS, n + 2),
        metadata: { amount: 20 + (n % 500), source: pickFrom(CHANNELS, n + 3) },
    };
}

export function makeBulkEvents(size, forceDedup = true) {
    const events = [];
    for (let i = 0; i < size; i++) {
        events.push(makeEvent(i, forceDedup));
    }
    return events;
}
