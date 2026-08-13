'use strict';

// The tests run inside an iframe (see fx.test.html) because they rewrite
// document.body and push history entries. Each one sets up a page, drives a
// real event, and waits for the DOM to catch up.

test('a[fx-target] navigates and updates fragments', async () => {
  let { calls, fetchResponses } = mockFetch(50, [
    `
      <div id="status">
        updated
        <script>window._fxTestScript = 'ran';</script>
      </div>
      <div id="hungry" fx-hungry>hungry-updated</div>
    `,
  ]);

  setPageHTML(
    '/initial',
    `
    <div id="status">initial</div>
    <div id="hungry" fx-hungry>hungry-initial</div>
    <a id="link" href="/next" fx-target="#status" fx-loading-target="#loader">go</a>
    <div id="loader">loader</div>
  `,
  );

  document.getElementById('link').click();

  assertClass('fx-loading', '#loader');
  await waitForText('updated', '#status');
  await waitForText('hungry-updated', '#hungry');
  assert(window._fxTestScript === 'ran', 'inline script not executed');
  assertClass('!fx-loading', '#loader');
  // The hungry element is part of this navigation, so the server has to be
  // told about it — otherwise honouring the header silently breaks fx-hungry.
  assert(
    calls[0].options.headers['FX-Target'] === '#status, #hungry',
    'FX-Target header must carry the hungry selectors too',
    calls[0].options.headers['FX-Target'],
  );
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('fx-target accepts a list of selectors', async () => {
  mockFetch(10, [
    `
      <div id="one">one-updated</div>
      <div id="two">two-updated</div>
      <div id="three">three-updated</div>
    `,
  ]);

  setPageHTML(
    '/initial',
    `
    <div id="one">one</div>
    <div id="two">two</div>
    <div id="three">three</div>
    <a id="link" href="/next" fx-target="#one, #two">go</a>
  `,
  );

  document.getElementById('link').click();

  await waitForText('one-updated', '#one');
  await waitForText('two-updated', '#two');
  assertText('three', '#three');
});

test('form[method="GET"][fx-target] submits and updates target', async () => {
  let { calls, fetchResponses } = mockFetch(20, [
    '<div id="result">got<script>window._fxTestScript = "ran";</script></div>',
  ]);

  setPageHTML(
    '/initial',
    `
    <div id="result">initial</div>
    <div id="loader">loader</div>
    <form id="form" action="/search" fx-target="#result" fx-loading-target="#loader">
      <input name="q" value="hello">
      <button type="submit">submit</button>
    </form>
  `,
  );

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

  setPageHTML(
    '/initial',
    `
    <div id="result">initial</div>
    <div id="loader">loader</div>
    <form id="form" action="/submit" method="POST" fx-target="#result" fx-loading-target="#loader">
      <input name="q" value="hello">
      <button type="submit">submit</button>
    </form>
  `,
  );

  document.getElementById('form').requestSubmit();

  assertClass('fx-loading', '#loader');
  await waitForText('posted', '#result');
  assertClass('!fx-loading', '#loader');
  assert(window._fxPostScript === 'ran-post', 'inline script not executed');

  let call = calls[0];
  assert(call.options.method === 'POST', 'method not POST', call.options.method);
  assert(!call.url.includes('q=hello'), 'query param leaked to URL', call.url);
  assert(call.options.body instanceof URLSearchParams, 'body should be urlencoded, not multipart');
  assert(call.options.body.get('q') === 'hello', 'body missing q', call.options.body);
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('form with a file keeps multipart encoding', async () => {
  let { calls } = mockFetch(10, ['<div id="result">uploaded</div>']);

  setPageHTML(
    '/initial',
    `
    <div id="result">initial</div>
    <form id="form" action="/upload" method="POST" enctype="multipart/form-data" fx-target="#result">
      <input type="file" name="doc">
      <button type="submit">submit</button>
    </form>
  `,
  );

  document.getElementById('form').requestSubmit();

  await waitForText('uploaded', '#result');
  assert(calls[0].options.body instanceof FormData, 'body should stay FormData for multipart forms');
});

test('the submitter decides the action, the method and its own value', async () => {
  let { calls } = mockFetch(10, ['<div id="result">deleted</div>']);

  setPageHTML(
    '/initial',
    `
    <div id="result">initial</div>
    <form id="form" action="/save" method="GET" fx-target="#result">
      <input name="id" value="7">
      <button type="submit" name="op" value="save">save</button>
      <button id="delete" type="submit" name="op" value="delete" formaction="/delete" formmethod="POST">delete</button>
    </form>
  `,
  );

  document.getElementById('delete').click();

  await waitForText('deleted', '#result');

  let call = calls[0];
  assert(call.options.method === 'POST', 'formmethod ignored', call.options.method);
  assert(call.url.includes('/delete'), 'formaction ignored', call.url);
  assert(call.options.body.get('op') === 'delete', 'submitter value missing', call.options.body.get('op'));
  assert(call.options.body.get('id') === '7', 'other fields missing');
});

test('a redirect replaces the url with where the server landed', async () => {
  mockFetch(10, [{ html: '<div id="result">created</div>', redirectUrl: '/things/42' }]);

  setPageHTML(
    '/initial',
    `
    <div id="result">initial</div>
    <a id="link" href="/things/new" fx-target="#result">go</a>
  `,
  );

  document.getElementById('link').click();

  await waitForText('created', '#result');
  assertPathname('/things/42');
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

  setPageHTML(
    '/initial',
    `
    <div id="content">
      <div id="title" fx-hungry>title initial</div>
      <a id="link-to-a" href="/page-a" fx-target="#content">link to a</a>
    </div>
  `,
  );

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

test('history state the app set survives an fx navigation', async () => {
  let { fetchResponses } = mockFetch(10, [
    '<div id="content">page a</div>',
    '<div id="content">initial restored</div>',
  ]);

  setPageHTML(
    '/initial',
    `
    <div id="content">
      <a id="link-to-a" href="/page-a" fx-target="#content">initial</a>
    </div>
  `,
  );

  // The entry belongs to the page, not to fx. An app may have put its own state
  // on it long before any fx navigation, and fx has to share rather than take.
  history.replaceState({ appState: 'keep me' }, '');

  document.getElementById('link-to-a').click();
  await waitForText('page a', '#content');
  assertPathname('/page-a');

  // Getting the fragment back proves fx still stamped the entry on the way out,
  // so this cannot be passed by simply leaving the entry alone.
  history.back();
  await waitForText('initial restored', '#content');
  assertPathname('/initial');

  assert(
    history.state?.appState === 'keep me',
    'fx overwrote the state the app put on the entry',
    JSON.stringify(history.state),
  );
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('a navigation resets the scroll position and popstate restores it', async () => {
  let { fetchResponses } = mockFetch(10, [
    '<div id="content" style="height: 4000px"><a id="link-to-b" href="/page-b" fx-target="#content">page a</a></div>',
    '<div id="content" style="height: 100px">page b</div>',
    '<div id="content" style="height: 4000px">page a restored</div>',
  ]);

  setPageHTML(
    '/initial',
    `
    <div id="content" style="height: 4000px">
      <a id="link-to-a" href="/page-a" fx-target="#content">initial</a>
    </div>
  `,
  );

  window.scrollTo(0, 500);
  await waitUntil(() => Math.round(window.scrollY) === 500, 'the initial page never scrolled');

  document.getElementById('link-to-a').click();
  await waitForText('page a', '#content');
  await waitUntil(() => Math.round(window.scrollY) === 0, 'a navigation should scroll back to the top');

  window.scrollTo(0, 300);
  await waitUntil(() => Math.round(window.scrollY) === 300, 'page a never scrolled');

  document.getElementById('link-to-b').click();
  await waitForText('page b', '#content');

  // Page b is short on purpose: the browser clamps its own restore against it,
  // so only the position fx saved on the entry can put page a back.
  history.back();
  await waitForText('page a restored', '#content');
  await waitUntil(() => Math.round(window.scrollY) === 300, 'popstate should restore the scroll position');
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('a query string change keeps the scroll position', async () => {
  let { fetchResponses } = mockFetch(10, [
    '<div id="content" style="height: 4000px">sorted by name</div>',
  ]);

  setPageHTML(
    '/reports',
    `
    <div id="content" style="height: 4000px">
      <a id="sort" href="/reports?sort=name" fx-target="#content">sort</a>
    </div>
  `,
  );

  window.scrollTo(0, 500);
  await waitUntil(() => Math.round(window.scrollY) === 500, 'the page never scrolled');

  document.getElementById('sort').click();
  await waitForText('sorted by name', '#content');
  await waitUntil(() => Math.round(window.scrollY) === 500, 'a refined view should stay where it was');
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('fx-scroll="top" resets a query string change', async () => {
  let { fetchResponses } = mockFetch(10, [
    '<div id="content" style="height: 4000px">page 2</div>',
  ]);

  setPageHTML(
    '/reports',
    `
    <div id="content" style="height: 4000px">
      <a id="pager" href="/reports?page=2" fx-target="#content" fx-scroll="top">next</a>
    </div>
  `,
  );

  window.scrollTo(0, 500);
  await waitUntil(() => Math.round(window.scrollY) === 500, 'the page never scrolled');

  document.getElementById('pager').click();
  await waitForText('page 2', '#content');
  await waitUntil(() => Math.round(window.scrollY) === 0, 'fx-scroll="top" should override the default');
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('fx-scroll="preserve" keeps the position across a path change', async () => {
  let { fetchResponses } = mockFetch(10, [
    '<div id="content" style="height: 4000px">page b</div>',
  ]);

  setPageHTML(
    '/page-a',
    `
    <div id="content" style="height: 4000px">
      <a id="link" href="/page-b" fx-target="#content" fx-scroll="preserve">to b</a>
    </div>
  `,
  );

  window.scrollTo(0, 500);
  await waitUntil(() => Math.round(window.scrollY) === 500, 'the page never scrolled');

  document.getElementById('link').click();
  await waitForText('page b', '#content');
  await waitUntil(() => Math.round(window.scrollY) === 500, 'fx-scroll="preserve" should override the default');
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('fx-scroll takes a selector and brings that element into view', async () => {
  let { fetchResponses } = mockFetch(10, [
    `
      <div id="content">
        <div style="height: 2000px">filters</div>
        <div id="results" style="height: 2000px">page 2</div>
      </div>
    `,
  ]);

  setPageHTML(
    '/reports',
    `
    <div id="content" style="height: 4000px">
      <a id="pager" href="/reports?page=2" fx-target="#content" fx-scroll="#results">next</a>
    </div>
  `,
  );

  document.getElementById('pager').click();
  await waitForText('page 2', '#results');
  await waitUntil(
    () => Math.round(document.getElementById('results').getBoundingClientRect().top) === 0,
    'fx-scroll should bring #results to the top of the viewport',
  );
  assert(window.scrollY > 0, 'the page should have scrolled down to it', window.scrollY);
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
    `,
  ]);

  setPageHTML(
    '/initial',
    `
    <meta id="meta-1" name="fx-refresh" fx-interval="50" fx-target="#poll-container">
    <meta id="meta-2" name="fx-refresh" fx-interval="55" fx-target="#poll-container">
    <div id="poll-container">
      poll-count=0
    </div>
  `,
  );

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

  setPageHTML(
    '/initial',
    `
    <meta id="meta" name="fx-refresh" fx-interval="100" fx-target="#nav">
    <a id="nav" href="/no" fx-target="#nav">nav</a>
  `,
  );

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

  assert(calls.length === 4, 'fetch should have been called 4 times', calls.length);
  assert(calls[0].url.includes('/initial'), 'fetch should have been called with /initial', calls[0].url);
  assert(calls[1].url.includes('/page-a'), 'fetch should have been called with /page-a', calls[1].url);
  assert(calls[2].url.includes('/page-b'), 'fetch should have been called with /page-b', calls[2].url);
  assert(calls[3].url.includes('/page-c'), 'fetch should have been called with /page-c', calls[3].url);
  assert(fetchResponses.length === 0, 'no fetch responses should be left', fetchResponses.length);
});

test('fetch errors are handled gracefully', async () => {
  mockFetch(10, ['error: Mock Server Error', '<a id="nav" href="/no" fx-target="#nav">ok</a>']);

  setPageHTML(
    '/initial',
    `
    <a id="nav" href="/no" fx-target="#nav">nav</a>
  `,
  );

  let called = false;
  fx.clickFallback = () => {
    called = true;
  };

  document.getElementById('nav').click();
  await waitUntil(() => called, 'clickFallback was never called');

  // The url was pushed optimistically before the request; a failure has to put
  // it back, or the address bar describes a page that was never rendered.
  assertPathname('/initial');

  fx.clickFallback = ORIGINAL_FALLBACKS.clickFallback;
  document.getElementById('nav').click();
  await waitForText('ok', '#nav');
});

test('fetch timeouts are handled gracefully', async () => {
  mockFetch(60, ['<a id="nav" href="/no" fx-target="#nav">ok</a>']);

  setPageHTML(
    '/initial',
    `
    <a id="nav" href="/no" fx-target="#nav">nav</a>
  `,
  );

  let called = false;
  fx.clickFallback = () => {
    called = true;
  };

  fx.timeout = '50';
  document.getElementById('nav').click();
  await waitUntil(() => called, 'clickFallback was never called on timeout');

  fx.clickFallback = ORIGINAL_FALLBACKS.clickFallback;
  fx.timeout = 500;
  document.getElementById('nav').click();
  await waitForText('ok', '#nav');
});

test('a failing form hands the submission back to the browser', async () => {
  mockFetch(10, ['error: Mock Server Error']);

  setPageHTML(
    '/initial',
    `
    <div id="result">initial</div>
    <form id="form" action="/submit" method="POST" fx-target="#result">
      <button type="submit">submit</button>
    </form>
  `,
  );

  let submitted = null;
  fx.submitFallback = (form) => {
    submitted = form;
  };

  document.getElementById('form').requestSubmit();
  await waitUntil(() => submitted !== null, 'submitFallback was never called');
  assert(submitted.id === 'form', 'submitFallback got the wrong form', submitted.id);

  // The button has to come back, or a failed submit leaves a dead form.
  assert(!document.querySelector('#form button').disabled, 'submit button left disabled');
});

test('a form whose controls shadow its own properties still submits', async () => {
  let { calls } = mockFetch(10, ['error: Mock Server Error']);

  // A form exposes its controls as properties, and those win over the ones the
  // platform defines — including its methods. Every name here is one a real
  // form uses, and each one used to break fx in a different way.
  //
  // The submission is aimed at an iframe so the real form.submit() the fallback
  // performs lands there instead of navigating the page running the tests.
  setPageHTML(
    '/initial',
    `
    <div id="result">initial</div>
    <iframe name="sink" style="display:none"></iframe>
    <form id="form" action="/submit" method="POST" target="sink" fx-target="#result">
      <input name="action" value="clobbered">
      <input name="method" value="clobbered">
      <button type="submit" name="submit" value="go">submit</button>
    </form>
  `,
  );

  document.getElementById('form').requestSubmit();
  await waitUntil(() => calls.length > 0, 'the form never reached the network');

  let call = calls[0];
  assert(call.url.includes('/submit'), 'the action attribute lost to the control named action', call.url);
  assert(call.options.method === 'POST', 'the method attribute lost to the control named method', call.options.method);

  // The fallback runs form.submit(), which is shadowed by the button named
  // "submit". When it throws, the await never returns and the form stays dead.
  await waitUntil(
    () => !document.querySelector('#form button').disabled,
    'submit button left disabled, the fallback threw',
  );
});

test('clicks the browser should keep are left alone', async () => {
  let { calls } = mockFetch(10, []);

  setPageHTML(
    '/initial',
    `
    <a id="modified" href="/no" fx-target="#modified">modified</a>
    <a id="blank" href="/no" target="_blank" fx-target="#blank">new tab</a>
    <a id="download" href="/no" download fx-target="#download">download</a>
    <a id="foreign" href="https://example.com/page" fx-target="#foreign">another origin</a>
    <a id="mailto" href="mailto:someone@example.com" fx-target="#mailto">mail</a>
  `,
  );

  // fx registered its listener at load, so this one runs after it and can both
  // read what fx decided and stop the browser from really navigating.
  let prevented = [];
  let spy = (event) => {
    prevented.push(event.defaultPrevented);
    event.preventDefault();
  };
  document.addEventListener('click', spy);

  document.getElementById('modified').dispatchEvent(
    new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }),
  );
  document.getElementById('blank').dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
  document.getElementById('download').dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));

  // fx cannot read another origin's document, and history refuses a url from
  // one. Taking the click would leave the user on a page that never moved.
  document.getElementById('foreign').dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
  document.getElementById('mailto').dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));

  await wait(30);
  document.removeEventListener('click', spy);

  assert(
    prevented.join() === 'false,false,false,false,false',
    'fx hijacked a click it should have ignored',
    prevented.join(),
  );
  assert(calls.length === 0, 'fx fetched for a click it should have ignored', calls.length);
});

// --- harness ------------------------------------------------------------

function assert(condition, message, value) {
  if (!condition) {
    throw new Error(`${message}${value === undefined ? '' : ` (${value})`}`);
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
    throw new Error(`assertPathname: expected "${pathname}", got "${window.location.pathname}"`);
  }
}

async function waitForText(text, selector) {
  await waitUntil(
    () => document.querySelector(selector)?.textContent.includes(text),
    `timed out waiting for "${text}" in ${selector}`,
  );
}

async function waitUntil(predicate, message) {
  for (let i = 0; i < 100; i++) {
    if (predicate()) return;
    await wait(25);
  }
  throw new Error(`waitUntil: ${message}`);
}

function wait(ms = 50) {
  return new Promise((r) => setTimeout(r, ms));
}

// mockFetch replaces window.fetch with a queue. A response is either an HTML
// string, "error: message" for a 500, or { html, redirectUrl, status }.
function mockFetch(delayMs, responses) {
  let fetchResponses = responses;
  let calls = [];

  window._fetch = window.fetch;

  window.fetch = async (url, options) => {
    calls.push({ url, options });

    if (options?.signal?.aborted) {
      throw options.signal.reason;
    }

    await wait(delayMs);

    if (options?.signal?.aborted) {
      throw options.signal.reason;
    }

    let response = fetchResponses.shift();
    assert(response, 'no response configured for fetch at index ' + calls.length);

    if (typeof response === 'string' && response.startsWith('error:')) {
      return { ok: false, status: 500, url, redirected: false, text: async () => response.slice(7) };
    }

    if (typeof response === 'string') {
      return { ok: true, status: 200, url, redirected: false, text: async () => response };
    }

    return {
      ok: response.status ? response.status < 400 : true,
      status: response.status || 200,
      url: response.redirectUrl || url,
      redirected: Boolean(response.redirectUrl),
      text: async () => response.html || '',
    };
  };

  return { calls, fetchResponses };
}

function restoreFetch() {
  if (window._fetch) window.fetch = window._fetch;
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

let ORIGINAL_PATHNAME = window.location.pathname;
let ORIGINAL_FALLBACKS = {
  clickFallback: fx.clickFallback,
  submitFallback: fx.submitFallback,
  historyFallback: fx.historyFallback,
};

function reset() {
  history.replaceState(null, '', ORIGINAL_PATHNAME);
  setPageHTML(ORIGINAL_PATHNAME);
  restoreFetch();
  Object.assign(fx, ORIGINAL_FALLBACKS);
  fx.timeout = undefined;
  window._fxTestScript = undefined;
  window._fxPostScript = undefined;
}

// runTests runs every test, keeps going after a failure, and returns a plain
// object — fx.test.html hands that to the Go test runner.
async function runTests() {
  let started = performance.now();

  let passed = [];
  let failed = [];
  let skipped = [];

  if (!globalThis.tests) globalThis.tests = [];

  // ?filter=… narrows the run, so `go test -run` and opening the page by hand
  // can both ask for one case.
  let filter = new URLSearchParams(window.location.search).get('filter') || '';
  let candidates = globalThis.tests;
  if (filter) {
    skipped.push(...candidates.filter((t) => !t.name.includes(filter)).map((t) => t.name));
    candidates = candidates.filter((t) => t.name.includes(filter));
  }

  let only = candidates.filter((t) => t.name.startsWith('only:'));
  let selected = only.length > 0 ? only : candidates;
  if (only.length > 0) {
    skipped.push(...candidates.filter((t) => !t.name.startsWith('only:')).map((t) => t.name));
  }

  for (let test of selected) {
    if (test.name.startsWith('skip:')) {
      skipped.push(test.name);
      continue;
    }

    let testStarted = performance.now();
    try {
      await test.fn();
      passed.push({ name: test.name, durationMs: Math.round(performance.now() - testStarted) });
    } catch (err) {
      failed.push({ name: test.name, error: err.message, stack: err.stack });
      console.error(`fx test failed: ${test.name}`, err);
    }
    reset();
  }

  return {
    passed,
    failed,
    skipped,
    only: only.length > 0,
    durationMs: Math.round(performance.now() - started),
  };
}
