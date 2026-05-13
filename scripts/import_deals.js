#!/usr/bin/env node
'use strict';

/**
 * G Motion CRM — CSV deal importer
 * Usage: node import_deals.js <password> [csv_path]
 *
 * Column mapping from spreadsheet:
 * DATE, CUSTOMER NAME, MAKE, MODEL, VIN, DEALER / SALES PERSON,
 * BROKER FEE, REFFERAL, DOC FEE, REFFERAL MONEY, BROKER,
 * BROKER %, BROKER PROFIT, CHECK STATUS, PROFIT BEFORE TAX,
 * BROKER PAID, BROKER PAYMENT TYPE, COLLECT, SOURCE, NOTES
 */

const fs   = require('fs');
const https = require('https');
const http  = require('http');

const BASE     = process.env.CRM_BASE_URL || 'https://gmotioncrm.com/api';
const CSV_PATH = process.argv[3] ||
  '/Users/yevgeniygeodakyan/Downloads/Untitled spreadsheet - Sheet1.csv';

let sessionCookie = '';

// ── HTTP ─────────────────────────────────────────────────────────────────────

function apiCall(method, path, body) {
  return new Promise((resolve, reject) => {
    const url  = new URL(BASE + path);
    const data = body ? JSON.stringify(body) : null;

    const opts = {
      hostname : url.hostname,
      port     : url.port || (url.protocol === 'https:' ? 443 : 80),
      path     : url.pathname + url.search,
      method,
      headers  : {
        'Content-Type': 'application/json',
        ...(sessionCookie ? { Cookie: sessionCookie } : {}),
        ...(data ? { 'Content-Length': Buffer.byteLength(data) } : {}),
      },
    };

    const transport = url.protocol === 'https:' ? https : http;
    const req = transport.request(opts, (res) => {
      const sc = res.headers['set-cookie'];
      if (sc) sessionCookie = sc.map(c => c.split(';')[0]).join('; ');

      let buf = '';
      res.on('data', c => buf += c);
      res.on('end', () => {
        if (res.statusCode >= 400) {
          reject(new Error(`HTTP ${res.statusCode} ${method} ${path}: ${buf}`));
          return;
        }
        try { resolve(JSON.parse(buf)); } catch { resolve(buf); }
      });
    });
    req.on('error', reject);
    if (data) req.write(data);
    req.end();
  });
}

// ── CSV parsing ───────────────────────────────────────────────────────────────

function parseCSV(text) {
  const lines = text.replace(/\r/g, '').split('\n');
  return lines.map(line => {
    const fields = [];
    let field = '', inQ = false;
    for (const ch of line) {
      if (ch === '"')      { inQ = !inQ; }
      else if (ch === ',' && !inQ) { fields.push(field.trim()); field = ''; }
      else                 { field += ch; }
    }
    fields.push(field.trim());
    return fields;
  });
}

// ── Field parsers ─────────────────────────────────────────────────────────────

const parseMoney = s => {
  if (!s) return null;
  const n = parseFloat(s.replace(/[$,\s]/g, ''));
  return isNaN(n) || n === 0 ? null : n;
};

const parsePct = s => {
  if (!s) return null;
  const n = parseFloat(s.replace('%', '').trim());
  return isNaN(n) ? null : n;
};

const splitName = full => {
  const p = (full || '').trim().split(/\s+/).filter(Boolean);
  if (!p.length) return { first: '', last: '' };
  return { first: p[0], last: p.slice(1).join(' ') };
};

const dealerName = s => {
  if (!s) return null;
  const i = s.indexOf('/');
  return (i !== -1 ? s.slice(0, i) : s).trim().toUpperCase();
};

const paymentMethod = s => {
  if (!s) return null;
  const l = s.toLowerCase().trim();
  if (l.includes('personal zelle') || l === 'personal') return 'zelle_personal';
  if (l.includes('zelle g motion') || l.includes('zelle gmotion')) return 'zelle_gmotion';
  if (l.includes('zelle'))  return 'zelle_gmotion';
  if (l.includes('cash'))   return 'cash';
  return null;
};

const isDataRow = row => /^\d+\/\d+\/\d+/.test((row[0] || '').trim());

const toISODate = s => {
  // M/D/YYYY → YYYY-MM-DD
  const [m, d, y] = s.split('/');
  return `${y}-${m.padStart(2,'0')}-${d.padStart(2,'0')}`;
};

const clean = obj => {
  Object.keys(obj).forEach(k => (obj[k] == null || obj[k] === '') && delete obj[k]);
  return obj;
};

// ── Main ──────────────────────────────────────────────────────────────────────

async function main() {
  const password = process.argv[2];
  if (!password) {
    console.error('Usage: node import_deals.js <password> [csv_path]');
    process.exit(1);
  }

  // Login
  process.stdout.write('Logging in… ');
  await apiCall('POST', '/auth/login', { password });
  console.log('OK');

  // Cache existing brokers and dealers
  process.stdout.write('Loading brokers & dealers… ');
  const [brokerRes, dealerRes] = await Promise.all([
    apiCall('GET', '/brokers?limit=200'),
    apiCall('GET', '/dealers?limit=200'),
  ]);
  const brokerById  = {};   // uppercased name → id
  const dealerById  = {};   // uppercased name → id
  for (const b of (brokerRes.data || brokerRes || [])) {
    brokerById[b.first_name.toUpperCase()] = b.id;
    brokerById[(b.first_name + ' ' + b.last_name).trim().toUpperCase()] = b.id;
  }
  for (const d of (dealerRes.data || dealerRes || [])) {
    dealerById[d.name.toUpperCase()] = d.id;
  }
  console.log(`${Object.keys(brokerById).length} brokers, ${Object.keys(dealerById).length} dealers`);

  // Parse CSV
  const rows     = parseCSV(fs.readFileSync(CSV_PATH, 'utf8'));
  const dataRows = rows.slice(1).filter(isDataRow);
  console.log(`Importing ${dataRows.length} deals…\n`);

  const clientCache = {}; // "FIRST LAST" → id
  let ok = 0, fail = 0;

  for (const row of dataRows) {
    const [
      date, custName, make, model, vin, dealerStr,
      brokerFeeStr, refName, docFeeStr, refMoneyStr,
      brokerStr, brokerPctStr, brokerCommStr, checkStatus, profitStr,
      brokerPaidStr, brokerPayTypeStr, collectStr, /*source*/, notes,
    ] = row;

    const label = `${custName} — ${make} ${model}`;

    try {
      // ── Client ──────────────────────────────────────────────
      const nameKey = custName.trim().toUpperCase();
      let clientId  = clientCache[nameKey];

      if (!clientId) {
        const { first, last } = splitName(custName);
        if (!first) { console.warn(`  SKIP (no name): ${row[0]}`); continue; }
        const c = await apiCall('POST', '/clients', clean({
          first_name : first,
          last_name  : last,
          status     : 'active',
        }));
        clientId = c.id;
        clientCache[nameKey] = clientId;
      }

      // ── Dealer ──────────────────────────────────────────────
      const dName  = dealerName(dealerStr);
      let dealerId = null;
      if (dName) {
        if (dealerById[dName]) {
          dealerId = dealerById[dName];
        } else {
          const d  = await apiCall('POST', '/dealers', { name: dName });
          dealerId = d.id;
          dealerById[dName] = dealerId;
        }
      }

      // ── Broker ──────────────────────────────────────────────
      const bKey   = (brokerStr || '').trim().toUpperCase();
      let brokerId = null;
      if (bKey) {
        if (brokerById[bKey]) {
          brokerId = brokerById[bKey];
        } else {
          const b  = await apiCall('POST', '/brokers', clean({
            first_name      : bKey,
            commission_rate : parsePct(brokerPctStr) ?? 0,
          }));
          brokerId = b.id;
          brokerById[bKey] = brokerId;
        }
      }

      // ── Deal status ─────────────────────────────────────────
      const dealStatus   = (checkStatus || '').trim().toUpperCase() === 'PENDING' ? 'pending' : 'paid';
      const brokerPaid   = (brokerPaidStr || '').trim().toUpperCase() === 'YES'   ? 'paid'    : 'pending';
      const pmMethod     = paymentMethod(brokerPayTypeStr);
      const brokerFee    = parseMoney(brokerFeeStr);
      const brokerComm   = parseMoney(brokerCommStr);
      const brokerPct    = parsePct(brokerPctStr);
      const docFee       = parseMoney(docFeeStr);
      const yourNet      = parseMoney(profitStr);
      const collectAmt   = parseMoney(collectStr);

      // Validate commission math; omit broker_commission if it doesn't match
      // the fee×percent calculation (backend rejects mismatches).
      let safeComm = brokerComm;
      if (brokerFee != null && brokerPct != null && brokerComm != null) {
        const expected = Math.round(brokerFee * brokerPct) / 100;
        if (Math.abs(expected - brokerComm) > 0.02) safeComm = null;
      }

      // ── Create deal ─────────────────────────────────────────
      const deal = await apiCall('POST', '/deals', clean({
        client_id                : clientId,
        dealer_id                : dealerId,
        broker_id                : brokerId,
        deal_date                : toISODate(date),
        make                     : make?.trim()  || undefined,
        model                    : model?.trim() || undefined,
        vin                      : vin?.trim()   || undefined,
        broker_fee               : brokerFee,
        broker_percent           : brokerPct,
        broker_commission        : safeComm,
        broker_commission_status : brokerPaid,
        broker_fee_status        : dealStatus === 'paid' ? 'paid' : 'pending',
        broker_payment_method    : pmMethod,
        deal_status              : dealStatus,
        your_net                 : yourNet,
        doc_fee                  : docFee,
        collect_amount           : collectAmt,
        notes                    : notes?.trim() || undefined,
      }));

      // ── Referral ─────────────────────────────────────────────
      const ref       = (refName || '').trim();
      const refMoney  = parseMoney(refMoneyStr);
      if (ref && ref.toUpperCase() !== 'NO' && ref !== '$0.00') {
        const { first, last } = splitName(ref);
        await apiCall('POST', '/referrals', clean({
          deal_id    : deal.id,
          first_name : first,
          last_name  : last || undefined,
          amount     : refMoney || undefined,
          status     : dealStatus === 'paid' ? 'paid' : 'none',
        }));
      }

      console.log(`  ✓  ${label}`);
      ok++;
    } catch (err) {
      console.error(`  ✗  ${label}: ${err.message}`);
      fail++;
    }
  }

  console.log(`\n────────────────────────────`);
  console.log(`  Created : ${ok}`);
  console.log(`  Errors  : ${fail}`);
  console.log(`────────────────────────────`);
}

main().catch(err => { console.error(err); process.exit(1); });
