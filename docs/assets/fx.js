// fx - Fragment eXchange library for SSR apps
// Instant URL updates, fragment swapping, and progressive enhancement
// 2025-12-04

window.fx = (() => {
  function fetchWithAbort({ url, method = 'GET', body, targetSelectors = '', abortController }) {
    let timeoutMs = parseInt(fx.timeout, 10) || 10_000;
    let controller = abortController || new AbortController();
    let timeout = setTimeout(() => controller.abort(`fetch timeout after ${timeoutMs}ms`), timeoutMs);

    return fetch(url, {
      method,
      body,
      signal: controller.signal,
      headers: { 'FX-Target': targetSelectors || '' },
    })
      .then((r) => {
        return r.ok
          ? r.text()
          : r.text().then((text) => {
              throw new Error(`HTTP ${r.status}: ${text}`);
            });
      })
      .finally(() => clearTimeout(timeout));
  }

  async function runFxNavigation({
    url,
    method = 'GET',
    body,
    targetSelectors = '',
    loadingSelectors = '',
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

    if (pushHistory && url !== window.location.href) {
      cancelAllTimers();
      history.pushState({ targetSelectors, loadingSelectors }, '', url);
    }

    try {
      let htmlText = await fetchWithAbort({
        url,
        method,
        body,
        targetSelectors,
        abortController,
      });

      let targets = getTargets(targetSelectors);
      updateFragments(targets, htmlText);
      setupMetaRefresh();

      let duration = Math.round(performance.now() - startTime);
      fx.logInfo(`Navigation(${label}): complete in ${duration}ms`, url);
    } catch (err) {
      if (err.name === 'AbortError' && abortController?.signal?.reason?.startsWith('[graceful]')) {
        fx.logDebug(`Navigation(${label}): fetch aborted:`, abortController.signal.reason);
        return;
      }

      fx.logError(`Navigation(${label}): failed:`, url, err.message);
      console.error(err);
      fallback?.(url, err);
    } finally {
      loadingElements.forEach((el) => el.classList.remove('fx-loading'));
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
      entry.abortController.abort('[graceful] url changed');
    }
    REFRESH_TIMERS_MAP.clear();
    fx.logDebug('Polling: cancelled all timers');
  }

  function setupMetaRefresh() {
    for (let [el, entry] of REFRESH_TIMERS_MAP.entries()) {
      if (document.contains(el)) continue;

      clearTimeout(entry.timer);
      entry.abortController.abort('[graceful] element removed from dom');
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

    window.fx.__dev_mode__validateDom?.(newDocument);

    let hungryElements = newDocument.querySelectorAll('[fx-hungry]');
    for (let el of hungryElements) {
      if (!el.id) continue;
      targetSelectors.push(`#${el.id}`);
    }

    let uniqueTargetSelectors = [...new Set(targetSelectors)];

    for (let selector of uniqueTargetSelectors) {
      let oldElement = document.querySelector(selector);
      if (!oldElement) continue;

      let newElement = newDocument.querySelector(selector);
      if (!newElement) continue;

      oldElement.replaceWith(newElement);

      let scripts = newElement.querySelectorAll('script');
      for (let script of scripts) {
        fx.logDebug('Run script:', script);
        let attributes = script.attributes;
        let newScript = document.createElement('script');
        for (let attribute of attributes) {
          newScript.setAttribute(attribute.name, attribute.value);
        }
        newScript.textContent = script.textContent;
        script.replaceWith(newScript);
      }
    }
  }

  async function handleClick(event) {
    let link = event.target.closest('a[href][fx-target]');
    if (!link || event.metaKey || event.ctrlKey || event.shiftKey) return;

    event.preventDefault();
    fx.logDebug('Click:', link);

    let targetSelectors = link.getAttribute('fx-target');
    let loadingSelectors = link.getAttribute('fx-loading-target');

    await runFxNavigation({
      url: link.href,
      method: 'GET',
      targetSelectors,
      loadingSelectors,
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

    let url = form.action || window.location.href;
    let targetSelectors = form.getAttribute('fx-target');
    let loadingSelectors = form.getAttribute('fx-loading-target');
    let method = (form.method || 'GET').toUpperCase();
    let body = new FormData(form);

    let submitButton = form.querySelector('button[type="submit"], input[type="submit"]');
    if (submitButton) {
      submitButton.disabled = true;
    }

    if (method === 'GET') {
      url = `${url}?${new URLSearchParams(body)}`;
      body = null;
    }

    await runFxNavigation({
      url,
      method,
      body,
      targetSelectors,
      loadingSelectors,
      pushHistory: true,
      label: 'submit',
      fallback: fx.submitFallback,
    });

    if (submitButton) {
      submitButton.disabled = false;
    }
  }

  let noop = () => {};

  let clickFallback = (url) => {
    window.location.href = url;
  };

  let submitFallback = () => {
    form.submit();
  };

  let historyFallback = () => {
    window.location.reload();
  };

  let fx = {
    version: '0.1.0',
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

    await runFxNavigation({
      url: window.location.href,
      method: 'GET',
      targetSelectors,
      loadingSelectors,
      pushHistory: false,
      label: 'history',
      fallback: fx.historyFallback,
    });
  });

  return fx;
})();
