// fx — Fragment eXchange
// Fragment navigation for server-rendered applications.
// https://fx.ciuffolo.com — MIT

window.fx = (() => {
  // Every in-flight fetch, so a new navigation can cancel the ones it made stale.
  let NAV_ABORT_CONTROLLERS_SET = new Set();

  // Aborts carry an Error so they can be told apart from a real failure: an
  // FxError means "we cancelled this on purpose, stay quiet", anything else
  // reaches the fallback and puts the browser back in charge.
  function error(message, name = 'FxError') {
    let err = new Error(message);
    err.name = name;
    return err;
  }

  function cancelAllFetches() {
    for (let controller of NAV_ABORT_CONTROLLERS_SET.values()) {
      controller.abort(error('url changed'));
    }

    NAV_ABORT_CONTROLLERS_SET.clear();
  }

  function fetchWithAbort({ url, method = 'GET', body, targetSelectors = '', abortController }) {
    let timeoutMs = parseInt(fx.timeout, 10) || 10_000;
    let controller = abortController || new AbortController();
    let timeout = setTimeout(
      () => controller.abort(error(`fetch timeout after ${timeoutMs}ms`, 'FxTimeoutError')),
      timeoutMs,
    );

    NAV_ABORT_CONTROLLERS_SET.add(controller);

    return fetch(url, {
      method,
      body,
      signal: controller.signal,
      headers: { 'FX-Target': targetSelectors || '' },
    })
      .then((r) => {
        let redirectUrl = r.redirected ? r.url : null;
        return r.ok
          ? r.text().then((text) => ({ htmlText: text, redirectUrl }))
          : r.text().then((text) => {
              throw new Error(`HTTP ${r.status}: ${text}`);
            });
      })
      .finally(() => {
        clearTimeout(timeout);
        NAV_ABORT_CONTROLLERS_SET.delete(controller);
      });
  }

  async function runFxNavigation({
    url,
    method = 'GET',
    body,
    targetSelectors = '',
    loadingSelectors = '',
    scrollBehavior = '',
    pushHistory = true,
    label = 'nav',
    abortController,
    fallback,
  }) {
    fx.logInfo(`Navigation(${label}): started`, url);

    let startTime = performance.now();

    let loadingElements = [];
    if (loadingSelectors) {
      loadingElements = Array.from(document.querySelectorAll(loadingSelectors));
      loadingElements.forEach((el) => el.classList.add('fx-loading'));
    }

    let originalUrl = window.location.href;
    let isUrlChange = pushHistory && url !== originalUrl;

    if (isUrlChange) {
      cancelAllTimers();
      cancelAllFetches();

      // The entry we are leaving may have no state at all — a full page load
      // never sets any. Stamp it before pushing, so going back knows what to
      // swap instead of falling back to a reload, and where the page was.
      history.replaceState({ targetSelectors, loadingSelectors, scrollY: window.scrollY }, '');
      history.pushState({ targetSelectors, loadingSelectors }, '', url);
    }

    try {
      // The hungry elements have to be worked out before the request, not
      // after it. They are part of what this navigation is going to swap, so
      // they belong in the header — otherwise a server that honours FX-Target
      // skips them, the response does not contain them, and they silently stop
      // updating.
      let targets = getTargets(targetSelectors);

      let { htmlText, redirectUrl } = await fetchWithAbort({
        url,
        method,
        body,
        targetSelectors: targets.join(', '),
        abortController,
      });

      updateFragments(targets, htmlText);
      setupMetaRefresh();

      if (redirectUrl) {
        history.replaceState({ targetSelectors, loadingSelectors }, '', redirectUrl);
        fx.logDebug(`Navigation(${label}): redirected to`, redirectUrl);
      }

      if (isUrlChange) {
        applyScroll(scrollBehavior, redirectUrl || url, originalUrl);
      }

      let duration = Math.round(performance.now() - startTime);
      fx.logInfo(`Navigation(${label}): complete in ${duration}ms`, redirectUrl || url);
    } catch (err) {
      if (err.name === 'FxError') {
        fx.logDebug(`Navigation(${label}): fetch aborted:`, err.message);
        return;
      }

      fx.logError(`Navigation(${label}): failed:`, url, err.message);

      // The URL was optimistically updated before the fetch. Put it back, so
      // the fallback starts from the address the user is actually looking at.
      if (pushHistory) {
        history.replaceState(null, '', originalUrl);
      }

      fallback?.(url, err);
    } finally {
      loadingElements.forEach((el) => el.classList.remove('fx-loading'));
    }
  }

  // A path change is a new page, and a new page starts at the top. A query
  // string change is the same page refined — sorted, filtered, paged — where
  // the reading position still means something, so it stays. fx-scroll settles
  // the cases the URL cannot: "top", "preserve", or a selector to scroll to.
  function applyScroll(scrollBehavior, url, previousUrl) {
    if (scrollBehavior === 'preserve') return;

    if (scrollBehavior && scrollBehavior !== 'top') {
      let element = document.querySelector(scrollBehavior);
      element
        ? element.scrollIntoView()
        : fx.logWarn('Nothing matches fx-scroll, staying put:', scrollBehavior);
      return;
    }

    let next = new URL(url, previousUrl);
    let previous = new URL(previousUrl);
    let anchor = next.hash && document.getElementById(decodeURIComponent(next.hash.slice(1)));

    if (anchor) {
      anchor.scrollIntoView();
    } else if (scrollBehavior === 'top' || next.origin !== previous.origin || next.pathname !== previous.pathname) {
      window.scrollTo(0, 0);
    }
  }

  function getTargets(targetSelectors) {
    let targets = [];

    if (targetSelectors) {
      targets.push(...targetSelectors.split(',').map((s) => s.trim()));
    }

    let hungryElements = document.querySelectorAll('[fx-hungry]');
    for (let el of hungryElements) {
      if (!el.id) continue;
      targets.push(`#${el.id}`);
    }

    return targets;
  }

  let REFRESH_TIMERS_MAP = new Map();

  function cancelAllTimers() {
    for (let entry of REFRESH_TIMERS_MAP.values()) {
      clearTimeout(entry.timer);
      entry.abortController.abort(error('url changed'));
    }
    REFRESH_TIMERS_MAP.clear();
    fx.logDebug('Polling: cancelled all timers');
  }

  function setupMetaRefresh() {
    for (let [el, entry] of REFRESH_TIMERS_MAP.entries()) {
      if (document.contains(el)) continue;

      clearTimeout(entry.timer);
      entry.abortController.abort(error('element removed from dom'));
      REFRESH_TIMERS_MAP.delete(el);

      fx.logDebug(`Polling[${el.id}]: removed timer, element no longer in dom`);
    }

    let metaElements = document.querySelectorAll('meta[name="fx-refresh"][fx-interval]');
    for (let el of metaElements) {
      if (REFRESH_TIMERS_MAP.has(el)) {
        continue;
      }

      let interval = parseInt(el.getAttribute('fx-interval'), 10);
      if (isNaN(interval)) continue;
      if (interval <= 0) continue;

      let url = window.location.href;
      let targetSelectors = el.getAttribute('fx-target');
      if (!targetSelectors) continue;
      let loadingSelectors = el.getAttribute('fx-loading-target');
      let abortController = new AbortController();

      fx.logDebug(`Polling[${el.id}]: added timer for ${interval}ms for ${targetSelectors}`);

      // the guard against stale timer updates is in 3 parts:
      // 1. [optional] if aborted during the timer's wait period, cancel the timer
      // 2. [mandatory] after we update the fragments, abort all timers for elements that are no longer in the DOM
      // 3. [mandatory] at the end of the refresh function, if aborted, exit early and don't add a new timer

      let refresh = async () => {
        fx.logDebug(`Polling[${el.id}]: running ${interval}ms for ${targetSelectors}`);

        await runFxNavigation({
          url,
          targetSelectors,
          loadingSelectors,
          pushHistory: false,
          label: `fx-refresh[${el.id}]`,
          abortController,
        });

        if (!abortController.signal.aborted) {
          let timer = setTimeout(refresh, interval);
          REFRESH_TIMERS_MAP.set(el, { timer, abortController });
        }
      };

      let timer = setTimeout(refresh, interval);
      REFRESH_TIMERS_MAP.set(el, { timer, abortController });
    }
  }

  function updateFragments(targetSelectors, htmlText) {
    let newDocument = new DOMParser().parseFromString(htmlText, 'text/html');
    fx.__dev_mode__validateDom?.(newDocument);

    let hungryElements = newDocument.querySelectorAll('[fx-hungry]');
    for (let el of hungryElements) {
      if (!el.id) continue;
      targetSelectors.push(`#${el.id}`);
    }

    let uniqueTargetSelectors = [...new Set(targetSelectors)];
    fx.logDebug(
      `updateFragments: ${uniqueTargetSelectors.length} unique target selectors: [${uniqueTargetSelectors.join(', ')}]`,
    );

    for (let selector of uniqueTargetSelectors) {
      let oldElement = document.querySelector(selector);
      if (!oldElement) {
        fx.logWarn('Target not in the current page, skipping:', selector);
        continue;
      }

      let newElement = newDocument.querySelector(selector);
      if (!newElement) {
        fx.logWarn('Target not in the response, skipping:', selector);
        continue;
      }

      // Save scroll positions by index. A scrolled table inside a fragment is
      // common enough that losing it on every poll is unusable; matching by
      // index is crude, and it is right whenever the shape did not change.
      let scrolls = [];
      let all = oldElement.querySelectorAll('*');
      for (let i = 0; i < all.length; i++) {
        if (all[i].scrollTop || all[i].scrollLeft) {
          scrolls.push([i, all[i].scrollTop, all[i].scrollLeft]);
        }
      }

      oldElement.replaceWith(newElement);

      let newAll = newElement.querySelectorAll('*');
      for (let [i, top, left] of scrolls) {
        if (newAll[i]) {
          newAll[i].scrollTop = top;
          newAll[i].scrollLeft = left;
        }
      }

      runScripts(newElement.querySelectorAll('script')).catch((err) => {
        fx.logError(`Script error in ${selector}:`, err.message);
        console.error(err);
      });
    }
  }

  // Scripts inserted by innerHTML/replaceWith never run. Re-create them so they
  // do, one after another, so a fragment can load a library and then use it.
  async function runScripts(scripts) {
    for (let script of scripts) {
      let attributes = script.attributes;
      let isExternal = script.src;

      let newScript = document.createElement('script');
      newScript.async = false;

      for (let attribute of attributes) {
        newScript.setAttribute(attribute.name, attribute.value);
      }

      await new Promise((resolve, reject) => {
        if (!isExternal) {
          newScript.textContent = script.textContent;
          script.replaceWith(newScript);
          resolve();
          return;
        }

        newScript.onload = resolve;
        newScript.onerror = () => reject(error(`failed to load ${newScript.src}`, 'FxScriptError'));
        script.replaceWith(newScript);
      });
    }
  }

  async function handleClick(event) {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;

    let link = event.target.closest('a[href][fx-target]');
    if (!link) return;

    // Leave the browser alone when it is being asked to do something other
    // than navigate this tab.
    if (link.hasAttribute('download')) return;
    if (link.target && link.target !== '_self') return;

    event.preventDefault();
    fx.logDebug('Click:', link);

    let targetSelectors = link.getAttribute('fx-target');
    let loadingSelectors = link.getAttribute('fx-loading-target');
    let scrollBehavior = link.getAttribute('fx-scroll');

    await runFxNavigation({
      url: link.href,
      method: 'GET',
      targetSelectors,
      loadingSelectors,
      scrollBehavior,
      pushHistory: true,
      label: 'click',
      fallback: fx.clickFallback,
    });
  }

  async function handleSubmit(event) {
    let form = event.target;
    if (!form.hasAttribute('fx-target')) return;

    fx.logDebug(`Submit:`, form);
    event.preventDefault();

    let submitter = event.submitter;
    let url = submitter?.getAttribute('formaction') || form.action || window.location.href;
    let targetSelectors = form.getAttribute('fx-target');
    let loadingSelectors = form.getAttribute('fx-loading-target');
    let scrollBehavior = submitter?.getAttribute('fx-scroll') || form.getAttribute('fx-scroll');
    let method = (submitter?.getAttribute('formmethod') || form.method || 'GET').toUpperCase();
    let body = new FormData(form);

    // A submit button's own name/value is part of the submission, and only the
    // button that was actually pressed.
    if (submitter && submitter.name) {
      body.append(submitter.name, submitter.value || '');
    }

    let submitButton = submitter || form.querySelector('button[type="submit"], input[type="submit"]');
    if (submitButton) {
      submitButton.disabled = true;
      submitButton.classList.add('fx-loading');
    }

    if (method === 'GET') {
      url = `${url}?${new URLSearchParams(body)}`;
      body = null;
    } else if (!isMultipart(form)) {
      // Same choice a browser makes: urlencoded unless there is a file to send.
      // Handed to fetch as FormData it would go out as multipart, which most
      // server-side form parsers do not read by default.
      body = new URLSearchParams(body);
    }

    await runFxNavigation({
      url,
      method,
      body,
      targetSelectors,
      loadingSelectors,
      scrollBehavior,
      pushHistory: true,
      label: 'submit',
      fallback: (failedUrl, err) => fx.submitFallback(form, failedUrl, err),
    });

    if (submitButton) {
      submitButton.disabled = false;
      submitButton.classList.remove('fx-loading');
    }
  }

  function isMultipart(form) {
    if (form.enctype === 'multipart/form-data') return true;
    return Array.from(form.elements).some((el) => el.type === 'file' && el.files?.length);
  }

  let noop = () => {};

  // The fallbacks are the whole safety story: when fx cannot do its job, it
  // hands the navigation back to the browser, which has always known how.
  let clickFallback = (url) => {
    window.location.replace(url);
  };

  let submitFallback = (form) => {
    form.submit();
  };

  let historyFallback = () => {
    window.location.reload();
  };

  let fx = {
    version: '1.1.0',
    logInfo: noop,
    logDebug: noop,
    logWarn: noop,
    logError: noop,
    clickFallback,
    submitFallback,
    historyFallback,
  };

  document.addEventListener('click', handleClick);
  document.addEventListener('submit', handleSubmit);
  document.addEventListener('DOMContentLoaded', () => setupMetaRefresh());

  window.addEventListener('popstate', async (event) => {
    fx.logDebug('Popstate:', event);

    let targetSelectors = event.state?.targetSelectors;
    let loadingSelectors = event.state?.loadingSelectors;

    cancelAllTimers();
    cancelAllFetches();

    await runFxNavigation({
      url: window.location.href,
      method: 'GET',
      targetSelectors,
      loadingSelectors,
      pushHistory: false,
      label: 'history',
      fallback: fx.historyFallback,
    });

    // The browser restores scroll against the page being left, before the
    // fragments arrive, so its guess is wrong as often as not.
    if (typeof event.state?.scrollY === 'number') {
      window.scrollTo(0, event.state.scrollY);
    }
  });

  return fx;
})();
