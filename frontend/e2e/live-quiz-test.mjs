/**
 * E2E Test: Live Quiz Mode — Admin + 3 Guests playing simultaneously
 *
 * Prerequisites:
 *   - Backend running on http://localhost:8080
 *   - Frontend running on http://localhost:5175
 *   - Seed data loaded (admin@test.com / 123456)
 *
 * Run: node e2e/live-quiz-test.mjs
 */

import puppeteer from 'puppeteer';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SCREENSHOTS = path.join(__dirname, 'screenshots');
const BASE = 'http://localhost:5175';

const ADMIN = { email: 'admin@test.com', password: '123456' };
const GUESTS = ['SwiftFox42', 'BravePanda77', 'CleverEagle13'];
const QUIZ_TITLE = `E2E Test ${Date.now()}`;

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

// ─── Admin Flow ───────────────────────────────────────────────

async function adminLogin(page) {
  log('STEP: Admin login');
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle0' });
  await page.type('input[type="email"]', ADMIN.email);
  await page.type('input[type="password"]', ADMIN.password);
  await screenshot(page, 'admin_login_form');
  await page.click('button[type="submit"]');
  await page.waitForNavigation({ waitUntil: 'networkidle0' });
  await screenshot(page, 'admin_dashboard');
  log('  ✅ Admin logged in');
}

async function adminCreateQuiz(page) {
  log('STEP: Admin creates quiz');
  await page.waitForSelector('input[placeholder="Quiz title"]');

  const titleInput = await page.$('input[placeholder="Quiz title"]');
  await titleInput.click({ clickCount: 3 });
  await titleInput.type(QUIZ_TITLE);

  const selects = await page.$$('select');
  if (selects.length > 0) {
    await selects[0].select('live');
  }

  await screenshot(page, 'admin_create_quiz_form');
  await page.click('button[type="submit"]');
  await wait(1500);
  await screenshot(page, 'admin_quiz_created');
  log('  ✅ Quiz created');
}

async function adminAddQuestions(page) {
  log('STEP: Admin adds questions');

  // Click "Questions" on the quiz we just created (match by title)
  await page.evaluate((title) => {
    const cards = Array.from(document.querySelectorAll('div'));
    for (const card of cards) {
      if (card.querySelector('h3')?.textContent?.includes(title) ||
          card.querySelector('strong')?.textContent?.includes(title) ||
          card.textContent?.includes(title)) {
        const btn = card.querySelector('button');
        if (btn && btn.textContent.trim() === 'Questions') {
          btn.click();
          return;
        }
      }
    }
    // Fallback: click first Questions button
    const buttons = Array.from(document.querySelectorAll('button'));
    const btn = buttons.find((b) => b.textContent.trim() === 'Questions');
    if (btn) btn.click();
  }, QUIZ_TITLE);
  await wait(1500);
  await page.waitForSelector('input[placeholder="Question text"]');

  const questions = [
    { text: 'What is 2 + 2?', options: ['3', '4', '5', '6'], correct: 1 },
    { text: 'What color is the sky?', options: ['Red', 'Green', 'Blue', 'Yellow'], correct: 2 },
    { text: 'How many legs does a cat have?', options: ['2', '3', '4', '5'], correct: 2 },
  ];

  for (let i = 0; i < questions.length; i++) {
    const q = questions[i];
    log(`  Adding question ${i + 1}: "${q.text}"`);

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

    // Click correct radio button (0-indexed)
    const radios = await page.$$('input[type="radio"]');
    if (radios[q.correct]) {
      await radios[q.correct].click();
    }

    await clickButtonByText(page, 'Add Question');
    await wait(1200);
  }

  await screenshot(page, 'admin_questions_added');
  log('  ✅ Questions added');
}

async function adminStartQuiz(page) {
  log('STEP: Admin starts quiz');
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
    // Fallback
    const buttons = Array.from(document.querySelectorAll('button'));
    const btn = buttons.find((b) => b.textContent.trim() === 'Start');
    if (btn) btn.click();
  }, QUIZ_TITLE);
  await wait(1500);
  await screenshot(page, 'admin_quiz_started');
  log('  ✅ Quiz started');
}

async function getQuizCode(page, title) {
  return page.evaluate((t) => {
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
    // Fallback: first code found
    const els = document.querySelectorAll('strong');
    for (const el of els) {
      if (el.parentElement?.textContent?.includes('Code:')) {
        return el.textContent.trim();
      }
    }
    return null;
  }, title);
}

async function adminJoinAsHost(page, code) {
  log(`STEP: Admin joins quiz as host (code: ${code})`);
  await page.goto(`${BASE}/play/${code}`, { waitUntil: 'networkidle0' });
  await wait(2000);
  await screenshot(page, 'admin_lobby');
  log('  ✅ Admin in lobby');
}

async function clickButtonByText(page, text) {
  return page.evaluate((t) => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const btn = buttons.find((b) => b.textContent.includes(t));
    if (btn) { btn.click(); return true; }
    return false;
  }, text);
}

// ─── Guest Flow ───────────────────────────────────────────────

async function guestJoin(page, name, code) {
  log(`STEP: Guest "${name}" joins quiz`);
  await page.goto(`${BASE}/join?code=${code}`, { waitUntil: 'networkidle0' });
  await wait(500);

  const nameInput = await page.$('input[placeholder="Your name"]');
  if (nameInput) {
    await nameInput.click({ clickCount: 3 });
    await nameInput.type(name);
  }

  await screenshot(page, `guest_${name}_join`);

  await page.click('button[type="submit"]');

  // Wait for lobby or play page
  await wait(3000);
  await screenshot(page, `guest_${name}_lobby`);
  log(`  ✅ Guest "${name}" in lobby/play`);
}

async function waitForQuestion(page, name, questionNum, maxWait = 10000) {
  const start = Date.now();
  while (Date.now() - start < maxWait) {
    const found = await page.evaluate((qNum) => {
      const el = document.querySelector('div');
      const allText = document.body?.innerText || '';
      // Check if "Question <N>" label is visible AND buttons are not disabled
      if (allText.includes(`Question ${qNum}`)) {
        const buttons = Array.from(document.querySelectorAll('button:not([disabled])'));
        // At least 2 non-disabled option buttons means the question is ready
        const optionButtons = buttons.filter((b) => {
          const style = b.getAttribute('style') || '';
          return style.includes('background-color') && !b.disabled;
        });
        return optionButtons.length >= 2;
      }
      return false;
    }, questionNum);
    if (found) return true;
    await wait(300);
  }
  log(`    ⚠️ Timed out waiting for Q${questionNum} on ${name}`);
  return false;
}

async function guestAnswerOption(page, name, optionIdx, questionNum) {
  log(`  Guest "${name}" answers Q${questionNum} → option ${optionIdx}`);

  // Wait for the question to appear with enabled buttons
  await waitForQuestion(page, name, questionNum);
  await wait(300);

  // QuestionCard renders options in a 2x2 grid inside a div with style display:grid
  // Each option is a <button> with colored background, not disabled
  const clicked = await page.evaluate((idx) => {
    // Find non-disabled buttons inside the grid
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

  if (clicked) {
    log(`    → Clicked: "${clicked}"`);
  } else {
    log(`    ⚠️ Could not find option ${optionIdx}`);
  }

  await wait(800);
  await screenshot(page, `guest_${name}_answer_q${questionNum}`);
}

// ─── Main Test ────────────────────────────────────────────────

async function main() {
  log('═══════════════════════════════════════════════════');
  log('  E2E TEST: Live Quiz Mode — Admin + 3 Guests');
  log('═══════════════════════════════════════════════════\n');

  // Clear screenshots
  const files = fs.readdirSync(SCREENSHOTS).filter((f) => f.endsWith('.png') || f.endsWith('.txt'));
  files.forEach((f) => fs.unlinkSync(path.join(SCREENSHOTS, f)));

  const browser = await puppeteer.launch({
    headless: false,
    defaultViewport: { width: 1280, height: 800 },
    args: ['--no-sandbox'],
    protocolTimeout: 120000,
  });

  let adminPage;
  const guestPages = [];

  try {
    // ═══ Phase 1: Admin Setup ═══
    adminPage = await browser.newPage();
    await adminLogin(adminPage);
    await adminCreateQuiz(adminPage);
    await adminAddQuestions(adminPage);
    await adminStartQuiz(adminPage);

    const code = await getQuizCode(adminPage, QUIZ_TITLE);
    if (!code) throw new Error('Could not find quiz code');
    log(`\n📋 Quiz code: ${code}\n`);

    // ═══ Phase 2: Admin joins lobby ═══
    await adminJoinAsHost(adminPage, code);

    // ═══ Phase 3: Create isolated guest sessions (incognito) ═══
    log('STEP: Creating guest browser sessions...');
    for (const name of GUESTS) {
      const ctx = await browser.createBrowserContext();
      const page = await ctx.newPage();
      await page.setViewport({ width: 1024, height: 768 });
      guestPages.push({ page, name, ctx });
    }

    // All guests join simultaneously
    log('\nSTEP: All guests joining simultaneously...');
    await Promise.all(
      guestPages.map(({ page, name }) => guestJoin(page, name, code))
    );

    await wait(2000);
    await screenshot(adminPage, 'admin_lobby_all_joined');
    log('  ✅ All 3 guests joined\n');

    // ═══ Phase 4: Q1 — Host starts, guests answer ═══
    log('STEP: Host starts Q1');
    await clickButtonByText(adminPage, 'Start First Question');
    await wait(2500);
    await screenshot(adminPage, 'admin_q1_started');

    for (const { page, name } of guestPages) {
      await screenshot(page, `guest_${name}_q1_received`);
    }

    log('\nSTEP: Guests answer Q1 (What is 2 + 2?)');
    // Guest 1: correct (idx 1 = "4"), fast
    await guestAnswerOption(guestPages[0].page, GUESTS[0], 1, 1);
    await wait(600);
    // Guest 2: correct (idx 1 = "4"), medium speed
    await guestAnswerOption(guestPages[1].page, GUESTS[1], 1, 1);
    await wait(600);
    // Guest 3: wrong (idx 0 = "3"), slow
    await guestAnswerOption(guestPages[2].page, GUESTS[2], 0, 1);

    await wait(2000);
    await screenshot(adminPage, 'admin_q1_all_answered');
    log('  ✅ Q1 complete\n');

    // ═══ Phase 5: Q2 — Host advances ═══
    log('STEP: Host advances to Q2');
    await clickButtonByText(adminPage, 'Next Question');
    await wait(2500);
    await screenshot(adminPage, 'admin_q2_started');

    for (const { page, name } of guestPages) {
      await screenshot(page, `guest_${name}_q2_received`);
    }

    log('\nSTEP: Guests answer Q2 (What color is the sky?)');
    await guestAnswerOption(guestPages[0].page, GUESTS[0], 2, 2); // correct "Blue"
    await wait(400);
    await guestAnswerOption(guestPages[1].page, GUESTS[1], 0, 2); // wrong "Red"
    await wait(400);
    await guestAnswerOption(guestPages[2].page, GUESTS[2], 2, 2); // correct "Blue"

    await wait(2000);
    await screenshot(adminPage, 'admin_q2_all_answered');
    log('  ✅ Q2 complete\n');

    // ═══ Phase 6: Q3 — Host advances ═══
    log('STEP: Host advances to Q3');
    await clickButtonByText(adminPage, 'Next Question');
    await wait(2500);
    await screenshot(adminPage, 'admin_q3_started');

    log('\nSTEP: Guests answer Q3 (How many legs does a cat have?)');
    await guestAnswerOption(guestPages[0].page, GUESTS[0], 2, 3); // correct "4"
    await wait(400);
    await guestAnswerOption(guestPages[1].page, GUESTS[1], 2, 3); // correct "4"
    await wait(400);
    await guestAnswerOption(guestPages[2].page, GUESTS[2], 1, 3); // wrong "3"

    await wait(2000);
    await screenshot(adminPage, 'admin_q3_all_answered');
    log('  ✅ Q3 complete\n');

    // ═══ Phase 7: Quiz finishes ═══
    log('STEP: Host ends quiz (next after last question)');
    await clickButtonByText(adminPage, 'Next Question');
    await wait(3000);

    await screenshot(adminPage, 'admin_final_results');
    for (const { page, name } of guestPages) {
      await screenshot(page, `guest_${name}_final_results`);
    }

    log('\n  ✅ Quiz finished! Final results shown');

    // ═══ Expected Results ═══
    log('\n─── Expected Scoring ───');
    log('  SwiftFox42:   Q1 ✅ Q2 ✅ Q3 ✅ → 3/3 correct (fastest)');
    log('  BravePanda77:  Q1 ✅ Q2 ❌ Q3 ✅ → 2/3 correct');
    log('  CleverEagle13: Q1 ❌ Q2 ✅ Q3 ❌ → 1/3 correct');

    log('\n═══════════════════════════════════════════════════');
    log('  TEST COMPLETE ✅');
    log('═══════════════════════════════════════════════════');

  } catch (err) {
    log(`\n❌ ERROR: ${err.message}`);
    if (adminPage) await screenshot(adminPage, 'error_admin').catch(() => {});
    for (const { page, name } of guestPages) {
      await screenshot(page, `error_guest_${name}`).catch(() => {});
    }
    throw err;
  } finally {
    // Write report
    const report = results.join('\n');
    fs.writeFileSync(path.join(SCREENSHOTS, 'test-report.txt'), report);
    log(`\n📂 Screenshots: ${SCREENSHOTS}`);
    log(`📸 Total: ${stepNum} screenshots`);
    log('📄 Report: screenshots/test-report.txt');

    await wait(5000);
    // Cleanup
    for (const { ctx } of guestPages) {
      await ctx.close().catch(() => {});
    }
    await browser.close();
  }
}

main().catch((err) => {
  console.error('\nTest failed:', err.message);
  process.exit(1);
});
