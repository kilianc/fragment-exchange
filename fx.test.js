'use strict';

test('a[fx-target] navigates and updates fragments', async () => {
  let { fetchResponses } = mockFetch(50, [
    `
      <div id="status">
        updated
        <script>window._fxTestScript = 'ran';</script>
      </div>
      <div id="hungry" fx-hungry>hungry-updated</div>
    `,
  ]);

  setPageHTML('/initial', `
    <div id="status">initial</div>
    <div id="hungry" fx-hungry>hungry-initial</div>
    <a id="link" href="/next" fx-target="#status" fx-loading-target="#loader">go</a>
    <div id="loader">loader</div>
  `);

  document.getElementById('link').click();

  assertClass('fx-loading', '#loader');
  await waitForText('updated', '#status');
  await waitForText('hungry-updated', '#hungry');
  assert(window._fxTestScript === 'ran', 'inline script not executed');
  assertClass('!fx-loading', '#loader');
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('form[method="GET"][fx-target] submits and updates target', async () => {
  let { calls, fetchResponses } = mockFetch(20, [
    '<div id="result">got<script>window._fxTestScript = "ran";</script></div>'
  ]);

  setPageHTML('/initial', `
    <div id="result">initial</div>
    <div id="loader">loader</div>
    <form id="form" action="/search" fx-target="#result" fx-loading-target="#loader">
      <input name="q" value="hello">
      <button type="submit">submit</button>
    </form>
  `);

  document.getElementById('form').requestSubmit();

  assertClass('fx-loading', '#loader');
  await waitForText('got', '#result');
  assertClass('!fx-loading', '#loader');
  assert(window._fxTestScript === 'ran', 'inline script not executed');

  let call = calls[0];
  assert(call.url.includes('q=hello'), 'query param missing', call.url);
  assert(call.options.method === 'GET', 'method not GET', call.options.method);
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('form[method="POST"][fx-target] submits and updates target', async () => {
  let { calls, fetchResponses } = mockFetch(20, [
    '<div id="result">posted<script>window._fxPostScript = "ran-post";</script></div>',
  ]);

  setPageHTML('/initial', `
    <div id="result">initial</div>
    <div id="loader">loader</div>
    <form id="form" action="/submit" method="POST" fx-target="#result" fx-loading-target="#loader">
      <input name="q" value="hello">
      <button type="submit">submit</button>
    </form>
  `);

  document.getElementById('form').requestSubmit();

  assertClass('fx-loading', '#loader');
  await waitForText('posted', '#result');
  assertClass('!fx-loading', '#loader');
  assert(window._fxPostScript === 'ran-post', 'inline script not executed');

  let call = calls[0];
  assert(call.options.method === 'POST', 'method not POST', call.options.method);
  assert(!call.url.includes('q=hello'), 'query param leaked to URL', call.url);
  assert(call.options.body.get('q') === 'hello', 'body missing q', call.options.body);
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('browser navigation restores fragments on popstate', async () => {
  let { fetchResponses } = mockFetch(10, [
    `
      <div id="content">
        <div id="title" fx-hungry>title a</div>
        <a id="link-to-b" href="/page-b" fx-target="#content">link to b</a>
      </div>
    `,
    `
      <div id="content">
        <div id="title" fx-hungry>title b</div>
        <a id="link-to-a" href="/page-a" fx-target="#content">link to a</a>
      </div>
    `,
    `
      <div id="content">
        <div id="title" fx-hungry>title a restored</div>
        <a id="link-to-b" href="/page-b" fx-target="#content">link to b restored</a>
      </div>
    `,
  ]);

  setPageHTML('/initial',`
    <div id="content">
      <div id="title" fx-hungry>title initial</div>
      <a id="link-to-a" href="/page-a" fx-target="#content">link to a</a>
    </div>
  `);

  document.getElementById('link-to-a').click();
  await waitForText('title a', '#title');
  assertPathname('/page-a');

  document.getElementById('link-to-b').click();
  await waitForText('title b', '#title');
  assertPathname('/page-b');

  history.back();
  await waitForText('title a restored', '#title');
  assertPathname('/page-a');
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('meta[name="fx-refresh"][fx-interval] should poll and self-replace', async () => {
  let { fetchResponses } = mockFetch(100, [
    `
      <!-- disable meta-1 timer -->
      <meta id="meta-1" fx-hungry>
      <div id="poll-container">
        poll-count=1
      </div>
    `,
    `
      <!-- disable meta-2 timer -->
      <meta id="meta-2" fx-hungry>
      <div id="poll-container">
        all-timers-disabled
      </div>
    `
  ]);

  setPageHTML('/initial', `
    <meta id="meta-1" name="fx-refresh" fx-interval="50" fx-target="#poll-container">
    <meta id="meta-2" name="fx-refresh" fx-interval="55" fx-target="#poll-container">
    <div id="poll-container">
      poll-count=0
    </div>
  `);

  await waitForText('all-timers-disabled', '#poll-container');
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('meta[name="fx-refresh"][fx-interval] polling timers are cleared on url change', async () => {
  let { calls, fetchResponses } = mockFetch(10, [
    '<a id="nav" href="/page-a" fx-target="#noop" fx-hungry>to-page-b</a>',
    '<a id="nav" href="/page-b" fx-target="#noop" fx-hungry>to-page-a</a>',
    '<a id="nav" href="/page-c" fx-target="#noop" fx-hungry>to-page-c</a>',
    '<meta id="meta" name="fx-refresh" fx-hungry>',
  ]);

  setPageHTML('/initial', `
    <meta id="meta" name="fx-refresh" fx-interval="100" fx-target="#nav">
    <a id="nav" href="/no" fx-target="#nav">nav</a>
  `);

  await waitForText('to-page-b', '#nav');

  await wait(25);
  document.getElementById('nav').click();
  assertPathname('/page-a');

  await wait(25);
  document.getElementById('nav').click();
  assertPathname('/page-b');

  await wait(25);
  document.getElementById('nav').click();
  assertPathname('/page-c');

  await wait(250);

  assert(calls.length === 4, 'fetch should have been called 3 times', calls.length);
  assert(calls[0].url.includes('/initial'), 'fetch should have been called with /initial', calls[0].url);
  assert(calls[1].url.includes('/page-a'), 'fetch should have been called with /page-b', calls[1].url);
  assert(calls[2].url.includes('/page-b'), 'fetch should have been called with /page-b', calls[1].url);
  assert(calls[3].url.includes('/page-c'), 'fetch should have been called with /page-c', calls[3].url);
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('fetch errors are handled gracefully', async () => {
  mockFetch(10, [
    'error: Mock Server Error',
    '<a id="nav" href="/no" fx-target="#nav">ok</a>',
  ]);

  setPageHTML('/initial', `
    <a id="nav" href="/no" fx-target="#nav">nav</a>
  `);

  let clickFallback = fx.clickFallback;
  fx.clickFallback = () => {
    fx.clickFallback = clickFallback;
  };

  document.getElementById('nav').click();
  document.getElementById('nav').click();
  await waitForText('ok', '#nav');
  assert(fx.clickFallback === clickFallback, 'fallback should have been called', fx.clickFallback);
});

test('fetch timeouts are handled gracefully', async () => {
  mockFetch(60, [
    '<a id="nav" href="/no" fx-target="#nav">ok</a>',
  ]);

  setPageHTML('/initial', `
    <a id="nav" href="/no" fx-target="#nav">nav</a>
  `);

  let clickFallback = fx.clickFallback;
  fx.clickFallback = () => {
    fx.clickFallback = clickFallback;
  };

  fx.timeout = '50';
  document.getElementById('nav').click();
  await wait(70);
  assert(fx.clickFallback === clickFallback, 'fallback should have been called', fx.clickFallback);

  fx.timeout = 70;
  document.getElementById('nav').click();
  await waitForText('ok', '#nav');
});

function assert(condition, message, value) {
  if (!condition) {
    throw new Error(`${message} ${value ? `(${value})` : ''}`);
  }
}

function assertClass(className, selector) {
  let element = document.querySelector(selector);
  if (!element) {
    throw new Error(`assertClass: element ${selector} not found`);
  }

  if (className.startsWith('!')) {
    className = className.slice(1);
    if (element.classList.contains(className)) {
      throw new Error(`assertClass: ${className} found in ${element.className}`);
    }
    return;
  }

  if (!element.classList.contains(className)) {
    throw new Error(`assertClass: ${className} not found in ${element.className}`);
  }
}

function assertText(text, selector) {
  let element = document.querySelector(selector);
  if (!element) {
    throw new Error(`assertText: element ${selector} not found`);
  }
  if (!element.textContent.includes(text)) {
    throw new Error(`assertText: text "${text}" not found in element ${selector}: ${element.textContent}`);
  }
}

function assertPathname(pathname) {
  if (window.location.pathname !== pathname) {
    throw new Error(
      `assertPathname: pathname "${pathname}" not found, current pathname is ${window.location.pathname}`,
    );
  }
}

async function waitForText(text, selector) {
  for (let i = 0; i < 100; i++) {
    let element = document.querySelector(selector);
    if (element.textContent.includes(text)) return;
    await wait(50);
  }

  throw new Error(`waitFor: timed out waiting for "${text}" in ${selector}`);
}

function wait(ms = 50) {
  return new Promise((r) => setTimeout(r, ms));
}

function mockFetch(delayMs, responses) {
  let fetchResponses = responses;
  let calls = [];

  window._fetch = window.fetch;

  window.fetch = async (url, options) => {
    calls.push({ url, options });

    if (options?.signal?.aborted) {
      throw new DOMException(options.signal.reason, 'AbortError');
    }

    await wait(delayMs);

    if (options?.signal?.aborted) {
      throw new DOMException(options.signal.reason, 'AbortError');
    }

    let response = fetchResponses.shift();
    assert(response, 'no response configured for fetch at index ' + calls.length);

    if (response.startsWith('error:')) {
      return {
        ok: false,
        status: 500,
        text: async () => response.slice(7),
      };
    }

    return {
      ok: true,
      text: async () => response,
    };
  };

  return { calls, fetchResponses };
}

function restoreFetch() {
  window.fetch = window._fetch;
  window._fetch = undefined;
}

function setPageHTML(path = '/initial', html = '') {
  history.replaceState(null, '', path);
  document.body.innerHTML = html;
  document.dispatchEvent(new Event('DOMContentLoaded'));
}

function test(name, fn) {
  if (!globalThis.tests) {
    globalThis.tests = [];
  }
  globalThis.tests.push({ name, fn });
}

let originalPathname = window.location.pathname;
let originalUrl = window.location.href;

function reset() {
  history.replaceState(null, '', originalPathname);
  setPageHTML(originalPathname);
  restoreFetch();
  window._fxTestScript = undefined;
}

async function runTests() {
  let originalUrl = window.location.href;
  let logPassed = fx.createLogFunction('#22c55e');
  let logFailed = fx.createLogFunction('#dc2626');
  let logSkipped = fx.createLogFunction('#6b7280');

  let passedTests = [];
  let failedTests = [];
  let skippedTests = [];
  let onlyTests = [];

  if (!globalThis.tests) {
    globalThis.tests = [];
  }

  for (let test of globalThis.tests) {
    if (test.name.startsWith('only:')) {
      onlyTests.push(test);
    } else {
      skippedTests.push(test);
    }
  }

  if (onlyTests.length > 0) {
    globalThis.tests = onlyTests;
  } else {
    skippedTests = [];
  }

  for (let test of globalThis.tests) {
    if (test.name.startsWith('skip:')) {
      skippedTests.push(test);
      continue;
    }

    try {
      await test.fn();
      reset();
      passedTests.push(test);
    } catch (err) {
      failedTests.push(test);
      console.error(err);
      break;
    }
  }

  history.replaceState(null, '', originalUrl);

  if (failedTests.length === 0 && onlyTests.length === 0) {
    console.clear();
  }

  for (let test of skippedTests) {
    logSkipped(`test: ${test.name} skipped`);
  }

  for (let test of passedTests) {
    logPassed(`test: ${test.name} passed`);
  }

  for (let test of failedTests) {
    logFailed(`test: ${test.name} failed`);
  }

  if (failedTests.length === 0) {
    logPassed(`all ${passedTests.length} tests passed`);
  } else {
    logFailed(`${failedTests.length} tests failed`);
  }
}
