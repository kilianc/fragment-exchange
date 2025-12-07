'use strict';

(function () {
  if (!window.fx) {
    console.fx.logWarn('fx.dev.js: fx.js not loaded');
    return;
  }

  let FX_LOG_KEY = 'fx:log-enabled';
  let FX_LOG_ENABLED = true;
  try { FX_LOG_ENABLED = localStorage.getItem(FX_LOG_KEY) === 'true'; } catch (_) { }

  fx.toggleLog = (value) => {
    FX_LOG_ENABLED = value;
    try { localStorage.setItem(FX_LOG_KEY, value.toString()); } catch (_) { }
    info(`logging ${value ? 'enabled' : 'disabled'}`);
  }

  fx.createLogFunction = (color, textColor = 'white') => {
    return (...args) => {
      if (!FX_LOG_ENABLED) return;
      console.log(
        `%cf⚡%c ${args[0]}`,
        `background:${color};color:white;padding:1px 4px;border-radius:3px;`,
        `color:${textColor};`,
        ...args.slice(1),
      );
    };
  }

  fx.logInfo = fx.createLogFunction('#2563eb');
  fx.logDebug = fx.createLogFunction('#4b5563', '#4b5563');
  fx.logWarn = fx.createLogFunction('#f59e0b');
  fx.logError = fx.createLogFunction('#dc2626');

  if (window.location.search.includes('fx-log=')) {
    fx.toggleLog(window.location.search.includes('fx-log=true'));
  }

  fx.__dev_mode__validateDom = (doc) => {
    // hungry without id
    for (let el of doc.querySelectorAll('[fx-hungry]')) {
      if (!el.id) fx.logWarn('Hungry element without id, skipping:', el);
    }

    // duplicate ids
    let idCounts = new Map();
    for (let el of doc.querySelectorAll('[id]')) {
      let id = el.id;
      idCounts.set(id, (idCounts.get(id) || 0) + 1);
    }
    for (let [id, count] of idCounts.entries()) {
      if (count > 1) {
        fx.logWarn('Duplicate id in document, target may be ambiguous:', id);
      }
    }

    // meta refresh checks
    for (let el of doc.querySelectorAll('meta[name="fx-refresh"]')) {
      let interval = parseInt(el.getAttribute('fx-interval'), 10);
      if (isNaN(interval) || interval <= 0) {
        fx.logWarn('fx-refresh: invalid fx-interval, skipping:', el);
      }
      let targetSelectors = el.getAttribute('fx-target');
      if (!targetSelectors) {
        fx.logWarn('fx-refresh: missing fx-target, skipping:', el);
      }
    }

    // fx-target missing in current DOM
    for (let el of doc.querySelectorAll('[fx-target]')) {
      let sel = el.getAttribute('fx-target');
      if (!sel) continue;
      let found = doc.querySelector(sel);
      if (!found) fx.logWarn('fx-target not found in document:', sel);
    }

    // fx-loading-target missing in current DOM
    for (let el of doc.querySelectorAll('[fx-loading-target]')) {
      let sel = el.getAttribute('fx-loading-target');
      if (!sel) continue;
      let found = doc.querySelector(sel);
      if (!found) fx.logWarn('fx-loading-target not found in document:', sel);
    }

    // form fx-target missing action
    for (let form of doc.querySelectorAll('form[fx-target]')) {
      if (!form.action) fx.logWarn('form with fx-target missing action, defaulting to current url');
    }

    // nested fx-targets (parent contains another target)
    let targets = Array.from(doc.querySelectorAll('[fx-target]')).map(el => el.getAttribute('fx-target')).filter(Boolean);
    let uniqueTargets = [...new Set(targets)];
    for (let sel of uniqueTargets) {
      let el = doc.querySelector(sel);
      if (!el) continue;
      for (let other of uniqueTargets) {
        if (other === sel) continue;
        let otherEl = doc.querySelector(other);
        if (otherEl && el.contains(otherEl)) {
          fx.logWarn('Nested fx-target detected; updates may be ambiguous:', sel, other);
        }
      }
    }
  }

  document.addEventListener('DOMContentLoaded', () => fx.__dev_mode__validateDom(document));
})();
