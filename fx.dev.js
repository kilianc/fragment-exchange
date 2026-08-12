// fx.dev.js — development companion for fx.js
// Turns on logging and checks every document fx parses for the mistakes that
// are otherwise silent. Do not ship it: fx.js alone is the library.
// https://fx.ciuffolo.com — MIT

'use strict';

(function () {
  if (!window.fx) {
    console.warn('fx.dev.js: fx.js not loaded');
    return;
  }

  let FX_LOG_KEY = 'fx:log-enabled';
  let FX_LOG_ENABLED = true;
  try {
    FX_LOG_ENABLED = localStorage.getItem(FX_LOG_KEY) === 'true';
  } catch (_) {}

  fx.toggleLog = (value) => {
    FX_LOG_ENABLED = value;
    try {
      localStorage.setItem(FX_LOG_KEY, value.toString());
    } catch (_) {}
    console.log(`fx: logging ${value ? 'enabled' : 'disabled'}`);
  };

  fx.createLogFunction = (color, textColor = 'white') => {
    return (...args) => {
      if (!FX_LOG_ENABLED) return;
      console.log(
        `%c𝒇𝒙%c ${args[0]}`,
        `background:${color};color:white;padding:1px 4px;border-radius:3px;`,
        `color:${textColor};`,
        ...args.slice(1),
      );
    };
  };

  fx.logInfo = fx.createLogFunction('#2563eb');
  fx.logDebug = fx.createLogFunction('#4b5563', '#4b5563');
  fx.logWarn = fx.createLogFunction('#f59e0b');
  fx.logError = fx.createLogFunction('#dc2626');

  // ?fx-log=true survives a reload, so you can turn logging on for one tab and
  // leave it on while you work.
  if (window.location.search.includes('fx-log=')) {
    fx.toggleLog(window.location.search.includes('fx-log=true'));
  }

  // fx-target holds a selector list. Checking the list as one selector would
  // pass as soon as any single selector matched.
  let selectors = (value) =>
    (value || '')
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);

  fx.__dev_mode__validateDom = (doc) => {
    // A hungry element is found by id. Without one there is nothing to swap.
    for (let el of doc.querySelectorAll('[fx-hungry]')) {
      if (!el.id) fx.logWarn('fx-hungry without an id, it will never update:', el);
    }

    // Two elements with the same id make every target ambiguous — the swap
    // takes the first match, which is rarely the one that changed.
    let idCounts = new Map();
    for (let el of doc.querySelectorAll('[id]')) {
      idCounts.set(el.id, (idCounts.get(el.id) || 0) + 1);
    }
    for (let [id, count] of idCounts.entries()) {
      if (count > 1) fx.logWarn(`Duplicate id "${id}" (${count} elements), targets are ambiguous`);
    }

    for (let el of doc.querySelectorAll('meta[name="fx-refresh"]')) {
      let interval = parseInt(el.getAttribute('fx-interval'), 10);
      if (isNaN(interval) || interval <= 0) {
        fx.logWarn('fx-refresh: fx-interval missing or not a positive number, no polling:', el);
      }
      if (!el.getAttribute('fx-target')) {
        fx.logWarn('fx-refresh: fx-target missing, no polling:', el);
      }
      if (!el.id) {
        fx.logWarn('fx-refresh: no id, its log lines will be hard to tell apart:', el);
      }
    }

    for (let el of doc.querySelectorAll('[fx-target]')) {
      for (let sel of selectors(el.getAttribute('fx-target'))) {
        if (!doc.querySelector(sel)) fx.logWarn(`fx-target "${sel}" matches nothing in this document:`, el);
      }
    }

    for (let el of doc.querySelectorAll('[fx-loading-target]')) {
      for (let sel of selectors(el.getAttribute('fx-loading-target'))) {
        if (!doc.querySelector(sel)) fx.logWarn(`fx-loading-target "${sel}" matches nothing in this document:`, el);
      }
    }

    // A form without an action posts to the current URL. That is often what
    // you meant, and when it is not the bug is invisible.
    for (let form of doc.querySelectorAll('form[fx-target]')) {
      if (!form.getAttribute('action')) {
        fx.logWarn('form[fx-target] has no action, it will use the current url:', form);
      }
    }
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => fx.__dev_mode__validateDom(document));
  } else {
    fx.__dev_mode__validateDom(document);
  }
})();
