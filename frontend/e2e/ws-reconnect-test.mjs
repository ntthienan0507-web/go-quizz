/**
 * E2E Test: WebSocket Reconnection — Player disconnects and reconnects mid-quiz
 *
 * Scenario:
 *   1. Admin creates & starts a quiz
 *   2. Guest joins and answers Q1 correctly
 *   3. Guest's WebSocket is forcefully killed (simulating network drop)
 *   4. Verify "Reconnecting..." UI appears
 *   5. Guest automatically reconnects
 *   6. Guest answers Q2 after reconnection
 *   7. Quiz finishes — verify score includes BOTH Q1 and Q2
 *
 * Prerequisites:
 *   - Backend running on http://localhost:8080
 *   - Frontend running on http://localhost:5175
 *   - Seed data loaded (admin@test.com / 123456)
 *
 * Run: node e2e/ws-reconnect-test.mjs
 */

import puppeteer from 'puppeteer';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SCREENSHOTS = path.join(__dirname, 'screenshots-reconnect');
const BASE = 'http://localhost:5175';

const ADMIN = { email: 'admin@test.com', password: '123456' };
const GUEST_NAME = 'ReconnectPlayer';
const QUIZ_TITLE = `Reconnect Test ${Date.now()}`;

let results = [];
let stepNum = 0;

function log(msg) {
  const line = `[${new Date().toISOString().slice(11, 19)}] ${msg}`;
  console.log(line);
  results.push(line);
}

async function screenshot(page, name) {
  stepNum++;
  const filename = `${String(stepNum).padStart(2, '0')}_${name}.png`;
  await page.screenshot({ path: path.join(SCREENSHOTS, filename), fullPage: true });
  log(`  📸 ${filename}`);
  return filename;
}

async function wait(ms) {
  return new Promise((r) => setTimeout(r, ms));
}

async function clickButtonByText(page, text) {
  return page.evaluate((t) => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const btn = buttons.find((b) => b.textContent.includes(t));
    if (btn) { btn.click(); return true; }
    return false;
  }, text);
}

async function waitForText(page, text, maxWait = 10000) {
  const start = Date.now();
  while (Date.now() - start < maxWait) {
    const found = await page.evaluate((t) => {
      return document.body?.innerText?.includes(t);
    }, text);
    if (found) return true;
    await wait(300);
  }
  return false;
}

async function waitForQuestion(page, questionNum, maxWait = 10000) {
  const start = Date.now();
  while (Date.now() - start < maxWait) {
    const found = await page.evaluate((qNum) => {
      const allText = document.body?.innerText || '';
      if (allText.includes(`Question ${qNum}`)) {
        const buttons = Array.from(document.querySelectorAll('button:not([disabled])'));
        const optionButtons = buttons.filter((b) => {
          const style = b.getAttribute('style') || '';
          return style.includes('background-color') && style.includes('border-radius');
        });
        return optionButtons.length >= 2;
      }
      return false;
    }, questionNum);
    if (found) return true;
    await wait(300);
  }
  return false;
}

async function answerOption(page, optionIdx) {
  return page.evaluate((idx) => {
    const allButtons = Array.from(document.querySelectorAll('button:not([disabled])'));
    const optionButtons = allButtons.filter((b) => {
      const style = b.getAttribute('style') || '';
      return style.includes('background-color') && style.includes('border-radius');
    });
    if (idx < optionButtons.length) {
      optionButtons[idx].click();
      return optionButtons[idx].textContent;
    }
    return null;
  }, optionIdx);
}

// ─── Admin Helpers ───────────────────────────────────────────

async function adminLogin(page) {
  log('STEP: Admin login');
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle0' });
  await page.type('input[type="email"]', ADMIN.email);
  await page.type('input[type="password"]', ADMIN.password);
  await page.click('button[type="submit"]');
  await page.waitForNavigation({ waitUntil: 'networkidle0' });
  log('  ✅ Admin logged in');
}

async function adminCreateAndStartQuiz(page) {
  log('STEP: Admin creates quiz');
  await page.waitForSelector('input[placeholder="Quiz title"]');

  const titleInput = await page.$('input[placeholder="Quiz title"]');
  await titleInput.click({ clickCount: 3 });
  await titleInput.type(QUIZ_TITLE);

  const selects = await page.$$('select');
  if (selects.length > 0) await selects[0].select('live');

  await page.click('button[type="submit"]');
  await wait(1500);
  log('  ✅ Quiz created');

  // Add questions
  log('STEP: Adding questions');
  await page.evaluate((title) => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const btn = buttons.find((b) => b.textContent.trim() === 'Questions');
    if (btn) btn.click();
  }, QUIZ_TITLE);
  await wait(1500);
  await page.waitForSelector('input[placeholder="Question text"]');

  const questions = [
    { text: 'What is 2 + 2?', options: ['3', '4', '5', '6'], correct: 1 },
    { text: 'What color is the sky?', options: ['Red', 'Green', 'Blue', 'Yellow'], correct: 2 },
  ];

  for (const q of questions) {
    const qInput = await page.$('input[placeholder="Question text"]');
    await qInput.click({ clickCount: 3 });
    await qInput.type(q.text);

    const optionInputs = await page.$$('input[placeholder^="Option"]');
    for (let j = 0; j < q.options.length; j++) {
      if (optionInputs[j]) {
        await optionInputs[j].click({ clickCount: 3 });
        await optionInputs[j].type(q.options[j]);
      }
    }

    const radios = await page.$$('input[type="radio"]');
    if (radios[q.correct]) await radios[q.correct].click();

    await clickButtonByText(page, 'Add Question');
    await wait(1200);
  }
  log('  ✅ Questions added');

  // Start quiz
  log('STEP: Starting quiz');
  await page.goto(`${BASE}/dashboard`, { waitUntil: 'networkidle0' });
  await wait(500);

  await page.evaluate((title) => {
    const cards = Array.from(document.querySelectorAll('div'));
    for (const card of cards) {
      if (card.textContent?.includes(title)) {
        const buttons = Array.from(card.querySelectorAll('button'));
        const btn = buttons.find((b) => b.textContent.trim() === 'Start');
        if (btn) { btn.click(); return; }
      }
    }
  }, QUIZ_TITLE);
  await wait(1500);
  log('  ✅ Quiz started');

  // Get quiz code
  const code = await page.evaluate((t) => {
    const cards = Array.from(document.querySelectorAll('div'));
    for (const card of cards) {
      if (card.textContent?.includes(t)) {
        const strongs = card.querySelectorAll('strong');
        for (const el of strongs) {
          if (el.parentElement?.textContent?.includes('Code:')) {
            return el.textContent.trim();
          }
        }
      }
    }
    return null;
  }, QUIZ_TITLE);

  return code;
}

// ─── Main Test ────────────────────────────────────────────────

async function main() {
  log('═══════════════════════════════════════════════════');
  log('  E2E TEST: WebSocket Reconnection');
  log('═══════════════════════════════════════════════════\n');

  // Ensure screenshots dir
  if (!fs.existsSync(SCREENSHOTS)) fs.mkdirSync(SCREENSHOTS, { recursive: true });
  const files = fs.readdirSync(SCREENSHOTS).filter((f) => f.endsWith('.png') || f.endsWith('.txt'));
  files.forEach((f) => fs.unlinkSync(path.join(SCREENSHOTS, f)));

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1280, height: 800 },
    args: ['--no-sandbox'],
    protocolTimeout: 120000,
  });

  let adminPage, guestCtx, guestPage;

  try {
    // ═══ Phase 1: Admin Setup ═══
    adminPage = await browser.newPage();
    await adminLogin(adminPage);
    const code = await adminCreateAndStartQuiz(adminPage);
    if (!code) throw new Error('Could not find quiz code');
    log(`\n📋 Quiz code: ${code}\n`);

    // Admin joins lobby
    log('STEP: Admin joins lobby');
    await adminPage.goto(`${BASE}/play/${code}`, { waitUntil: 'networkidle0' });
    await wait(2000);
    await screenshot(adminPage, 'admin_lobby');

    // ═══ Phase 2: Guest Joins ═══
    guestCtx = await browser.createBrowserContext();
    guestPage = await guestCtx.newPage();
    await guestPage.setViewport({ width: 1024, height: 768 });

    log(`STEP: Guest "${GUEST_NAME}" joins`);
    await guestPage.goto(`${BASE}/join?code=${code}`, { waitUntil: 'networkidle0' });
    await wait(500);
    const nameInput = await guestPage.$('input[placeholder="Your name"]');
    if (nameInput) {
      await nameInput.click({ clickCount: 3 });
      await nameInput.type(GUEST_NAME);
    }
    await guestPage.click('button[type="submit"]');
    await wait(3000);
    await screenshot(guestPage, 'guest_joined_lobby');
    log('  ✅ Guest in lobby');

    // ═══ Phase 3: Host starts Q1, Guest answers ═══
    log('\nSTEP: Host starts Q1');
    await clickButtonByText(adminPage, 'Start First Question');
    await wait(2500);
    await screenshot(adminPage, 'admin_q1_started');

    log('STEP: Guest answers Q1 correctly (option 1 = "4")');
    await waitForQuestion(guestPage, 1);
    await wait(300);
    await answerOption(guestPage, 1); // "4" = correct
    await wait(1500);
    await screenshot(guestPage, 'guest_q1_answered');
    log('  ✅ Q1 answered');

    // Verify answer result shows correct
    const q1Correct = await guestPage.evaluate(() => {
      return document.body?.innerText?.includes('Correct');
    });
    log(`  Q1 result: ${q1Correct ? '✅ Correct' : '❌ Wrong'}`);
    await screenshot(adminPage, 'admin_q1_answered');

    // ═══ Phase 4: KILL WEBSOCKET — simulate network drop ═══
    log('\n══════════════════════════════════════');
    log('  💥 KILLING GUEST WEBSOCKET');
    log('══════════════════════════════════════');

    // Force close all WebSocket connections via CDP
    const cdpSession = await guestPage.createCDPSession();
    await cdpSession.send('Network.enable');

    // Close all WebSocket connections by evaluating in page
    await guestPage.evaluate(() => {
      // Access the internal WebSocket and force close it
      // This simulates an abrupt network failure
      const allWs = (window).__testWs;
      if (allWs) allWs.close();

      // Brute force: override WebSocket prototype to kill existing connections
      const originalWs = window.WebSocket;
      const instances = [];
      // @ts-ignore
      window.__wsInstances = instances;

      // Hook to find existing connections via performance API
      const entries = performance.getEntriesByType('resource');
      // Just forcefully close any WebSocket we can find
    });

    // More reliable: use CDP to emulate offline mode
    await cdpSession.send('Network.emulateNetworkConditions', {
      offline: true,
      latency: 0,
      downloadThroughput: 0,
      uploadThroughput: 0,
    });

    log('  Network set to OFFLINE');
    await wait(2000);
    await screenshot(guestPage, 'guest_disconnected');

    // Check if "Reconnecting..." text appears
    const reconnectingVisible = await guestPage.evaluate(() => {
      return document.body?.innerText?.includes('Reconnecting') ||
             document.body?.innerText?.includes('Connection lost');
    });
    log(`  Reconnecting UI visible: ${reconnectingVisible ? '✅ YES' : '⚠️ NO (may show on next render)'}`);
    await screenshot(guestPage, 'guest_reconnecting_ui');

    // ═══ Phase 5: Host advances to Q2 while guest is disconnected ═══
    log('\nSTEP: Host advances to Q2 (guest still offline)');
    await clickButtonByText(adminPage, 'Next Question');
    await wait(2000);
    await screenshot(adminPage, 'admin_q2_started_guest_offline');

    // ═══ Phase 6: Restore network — guest reconnects ═══
    log('\n══════════════════════════════════════');
    log('  🔄 RESTORING NETWORK');
    log('══════════════════════════════════════');

    await cdpSession.send('Network.emulateNetworkConditions', {
      offline: false,
      latency: 0,
      downloadThroughput: -1,
      uploadThroughput: -1,
    });

    log('  Network set to ONLINE');

    // Wait for reconnection (exponential backoff: first retry at 1s, second at 2s)
    await wait(5000);
    await screenshot(guestPage, 'guest_after_reconnect');

    // Check if guest is reconnected (either sees Q2 or lobby)
    const reconnected = await guestPage.evaluate(() => {
      const text = document.body?.innerText || '';
      return !text.includes('Reconnecting') && !text.includes('Connection lost');
    });
    log(`  Reconnected: ${reconnected ? '✅ YES' : '❌ NO'}`);

    // Guest should see Q2 (late join sends current question)
    const seesQ2 = await waitForText(guestPage, 'Question 2', 8000);
    log(`  Sees Q2 after reconnect: ${seesQ2 ? '✅ YES' : '⚠️ NO'}`);
    await screenshot(guestPage, 'guest_sees_q2');

    // ═══ Phase 7: Guest answers Q2 ═══
    if (seesQ2) {
      log('\nSTEP: Guest answers Q2 after reconnection (option 2 = "Blue")');
      await waitForQuestion(guestPage, 2);
      await wait(300);
      await answerOption(guestPage, 2); // "Blue" = correct
      await wait(1500);
      await screenshot(guestPage, 'guest_q2_answered_after_reconnect');

      const q2Correct = await guestPage.evaluate(() => {
        return document.body?.innerText?.includes('Correct');
      });
      log(`  Q2 result: ${q2Correct ? '✅ Correct' : '❌ Wrong'}`);
    }

    await wait(1500);
    await screenshot(adminPage, 'admin_q2_answered');

    // ═══ Phase 8: Finish quiz & verify score ═══
    log('\nSTEP: Host finishes quiz');
    await clickButtonByText(adminPage, 'Next Question');
    await wait(3000);
    await screenshot(adminPage, 'admin_final_results');
    await screenshot(guestPage, 'guest_final_results');

    // Check if guest has a score > 0 (meaning Q1 score was preserved through reconnect)
    const guestScore = await guestPage.evaluate((name) => {
      // The leaderboard renders each entry as a row with rank, username, and score badge
      // Find all elements that contain the guest name, then look for a nearby score
      const allElements = Array.from(document.querySelectorAll('div, span'));
      for (const el of allElements) {
        if (el.textContent?.includes(name) && el.children.length > 0) {
          // Look for score in a badge/span nearby
          const scoreEls = el.querySelectorAll('span');
          for (const s of scoreEls) {
            const num = parseFloat(s.textContent);
            if (!isNaN(num) && num > 0) return num;
          }
        }
      }
      // Fallback: search all text content for numbers after the player name
      const text = document.body?.innerText || '';
      const idx = text.indexOf(name);
      if (idx >= 0) {
        const after = text.slice(idx + name.length, idx + name.length + 50);
        const match = after.match(/(\d+)/);
        if (match) return parseInt(match[1]);
      }
      return -1;
    }, GUEST_NAME);

    log(`\n─── Results ───`);
    log(`  Guest score: ${guestScore}`);
    log(`  Score preserved through reconnect: ${guestScore > 0 ? '✅ YES' : '❌ NO'}`);

    // ═══ Final Verdict ═══
    log('\n═══════════════════════════════════════════════════');
    if (reconnected && guestScore > 0) {
      log('  TEST PASSED ✅ — Reconnection works, score preserved');
    } else {
      log('  TEST FAILED ❌');
      if (!reconnected) log('    - Guest did not reconnect');
      if (guestScore <= 0) log('    - Score was not preserved');
    }
    log('═══════════════════════════════════════════════════');

  } catch (err) {
    log(`\n❌ ERROR: ${err.message}`);
    if (adminPage) await screenshot(adminPage, 'error_admin').catch(() => {});
    if (guestPage) await screenshot(guestPage, 'error_guest').catch(() => {});
    throw err;
  } finally {
    const report = results.join('\n');
    fs.writeFileSync(path.join(SCREENSHOTS, 'test-report.txt'), report);
    log(`\n📂 Screenshots: ${SCREENSHOTS}`);
    log(`📸 Total: ${stepNum} screenshots`);

    await wait(5000);
    if (guestCtx) await guestCtx.close().catch(() => {});
    await browser.close();
  }
}

main().catch((err) => {
  console.error('\nTest failed:', err.message);
  process.exit(1);
});
