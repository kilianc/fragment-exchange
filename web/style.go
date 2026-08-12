package web

// stylesheet is inlined into every page.
//
// The site argues that you do not need a build step. Linking a compiled
// stylesheet would be a strange way to make that case, and one <style> is
// faster than one more request anyway.
//
// The palette is light in every environment. The page is mostly prose and
// code samples whose highlighting has to stay in tune with the diagram beside
// it, and one palette is one thing to keep in tune.
const stylesheet = `
:root {
  --bg: #fcfcfb;
  --bg-raised: #ffffff;
  --bg-code: #f5f5f2;
  --bg-inset: #f0f0ec;
  --fg: #16161a;
  --fg-muted: #62626e;
  --fg-faint: #8f8f9c;
  --border: #e4e4de;
  --border-strong: #d2d2ca;
  /* The blue fx.dev.js prints its log badge in. The accent on the page and
     the accent in your console are the same colour on purpose. */
  --accent: #2563eb;
  --accent-fg: #ffffff;
  --accent-soft: #e9effd;
  --ok: #0f7a52;
  --warn: #a35700;
  --bad: #b42318;

  --hl-comment: #8b8b96;
  --hl-string: #0f7a52;
  --hl-keyword: #8b3fa8;
  --hl-tag: #b42318;
  --hl-attr: #a35700;
  --hl-fx: #2563eb;
  --hl-number: #a35700;

  --radius: 10px;
  --mono: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
  --sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Inter, Roboto, sans-serif;
  /* System serif only. The site has no external requests and is not about to
     grow one for a font. ui-serif is New York on Apple platforms; Charter and
     Georgia cover the rest. */
  --serif: ui-serif, Charter, "Bitstream Charter", "Iowan Old Style", Georgia, "Times New Roman", serif;
}

* { box-sizing: border-box; }

/* Form controls, scrollbars and the diagram follow this, so they stay light
   under a dark OS instead of being half-inverted. */
html {
  color-scheme: light;
  -webkit-text-size-adjust: 100%;
  scroll-behavior: smooth;
  scroll-padding-top: 80px;
}

body {
  margin: 0;
  background: var(--bg);
  color: var(--fg);
  font-family: var(--serif);
  font-size: 18px;
  line-height: 1.72;
  -webkit-font-smoothing: antialiased;
}

/* The paper is serif. The application is not — the chrome, the tables and the
   demo stay sans, so a control never reads as prose. */
header.top, nav.primary, .btn, table, footer.bottom,
.panel, .filters, button, select, input, .pill, .log, .empty,
.snippet-label, .kicker, .diagram, .steps li::before {
  font-family: var(--sans);
}

a { color: var(--accent); text-decoration: none; }
a:hover { text-decoration: underline; }

h1, h2, h3 { line-height: 1.22; letter-spacing: -0.01em; font-weight: 680; }
h1 { font-size: 2.3rem; margin: 0 0 14px; }
h2 { font-size: 1.55rem; margin: 60px 0 16px; }
h3 { font-size: 1.15rem; margin: 34px 0 10px; }

p { margin: 0 0 18px; }
ul, ol { margin: 0 0 18px; padding-left: 22px; }
li { margin: 6px 0; }
strong { font-weight: 640; }

/* ---- progress bar -------------------------------------------------- */

#progress-bar {
  position: fixed;
  top: 0; left: 0;
  height: 2px;
  width: 0;
  background: var(--accent);
  z-index: 100;
  transition: width 240ms ease-out, opacity 200ms 120ms;
  opacity: 0;
}
#progress-bar.fx-loading {
  width: 70%;
  opacity: 1;
  transition: width 900ms cubic-bezier(.1,.8,.3,1);
}

/* ---- chrome -------------------------------------------------------- */

header.top {
  position: sticky;
  top: 0;
  z-index: 50;
  background: color-mix(in srgb, var(--bg) 88%, transparent);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
}

.top-inner {
  max-width: 1180px;
  margin: 0 auto;
  padding: 11px 24px;
  display: flex;
  align-items: center;
  gap: 24px;
}

.brand {
  font-family: var(--mono);
  font-weight: 700;
  font-size: 15px;
  color: var(--fg);
  letter-spacing: -0.03em;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.brand:hover { text-decoration: none; }
.brand .mark {
  display: block;
  width: 22px; height: 22px;
  color: var(--accent);
}

nav.primary { display: flex; gap: 2px; margin-left: auto; flex-wrap: wrap; }
nav.primary a {
  padding: 5px 11px;
  border-radius: 7px;
  font-size: 14px;
  color: var(--fg-muted);
}
nav.primary a:hover { background: var(--bg-inset); color: var(--fg); text-decoration: none; }
nav.primary a[aria-current="page"] { background: var(--bg-inset); color: var(--fg); font-weight: 560; }

.shell { max-width: 1180px; margin: 0 auto; padding: 0 24px; }
main { padding: 48px 0 96px; max-width: 740px; }
main.wide { max-width: none; }

.page-head { margin-bottom: 36px; }
/* The deck: the claim, between the title and the explanation. Full-strength
   colour, because it is the line the page is actually making. */
.tagline {
  font-size: 1.5rem;
  line-height: 1.35;
  letter-spacing: -0.01em;
  font-weight: 600;
  color: var(--fg);
  margin: 0 0 14px;
  max-width: 30ch;
}

.lede { font-size: 1.2rem; line-height: 1.6; color: var(--fg-muted); margin: 0; max-width: 60ch; }

footer.bottom {
  border-top: 1px solid var(--border);
  padding: 28px 0 64px;
  color: var(--fg-faint);
  font-size: 14px;
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

/* ---- code ---------------------------------------------------------- */

/* Mono runs large next to a serif, and the inset needs to sit tight enough
   that a comma after it does not look like a space. */
code {
  font-family: var(--mono);
  font-size: 0.84em;
  background: var(--bg-inset);
  padding: 0.08em 0.26em;
  border-radius: 4px;
}

pre {
  margin: 0 0 20px;
  background: var(--bg-code);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 16px 18px;
  overflow-x: auto;
  line-height: 1.6;
}
pre code {
  background: none;
  padding: 0;
  font-size: 13.5px;
  white-space: pre;
}

.snippet { margin: 0 0 20px; }
.snippet > pre { margin: 0; border-radius: 0 0 var(--radius) var(--radius); border-top: 0; }
.snippet-label {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--fg-faint);
  background: var(--bg-inset);
  border: 1px solid var(--border);
  border-radius: var(--radius) var(--radius) 0 0;
  padding: 6px 14px;
}

.c-comment { color: var(--hl-comment); font-style: italic; }
.c-string  { color: var(--hl-string); }
.c-keyword { color: var(--hl-keyword); }
.c-tag     { color: var(--hl-tag); }
.c-attr    { color: var(--hl-attr); }
.c-number  { color: var(--hl-number); }
.c-fx      { color: var(--hl-fx); font-weight: 620; }

/* ---- blocks -------------------------------------------------------- */

.note {
  border-left: 3px solid var(--accent);
  background: var(--accent-soft);
  padding: 14px 18px;
  border-radius: 0 var(--radius) var(--radius) 0;
  margin: 0 0 20px;
}
.note > :last-child { margin-bottom: 0; }
.note-title { font-weight: 640; display: block; margin-bottom: 4px; }

.cols { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 20px; }
.cols > * { min-width: 0; }
@media (max-width: 760px) { .cols { grid-template-columns: 1fr; } }

table { border-collapse: collapse; width: 100%; margin: 0 0 22px; font-size: 15px; line-height: 1.6; }
th, td { text-align: left; padding: 9px 12px; border-bottom: 1px solid var(--border); vertical-align: top; }
th { font-size: 12px; letter-spacing: 0.05em; text-transform: uppercase; color: var(--fg-faint); font-weight: 600; }
td code { white-space: nowrap; }

.hero-cta { display: flex; gap: 10px; flex-wrap: wrap; margin: 0 0 32px; }
.btn {
  display: inline-block;
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  font-size: 14.5px;
  font-weight: 520;
  color: var(--fg);
  background: var(--bg-raised);
}
.btn:hover { text-decoration: none; border-color: var(--fg-faint); }
.btn-primary { background: var(--accent); border-color: var(--accent); color: var(--accent-fg); }
.btn-primary:hover { filter: brightness(1.08); }

.steps { counter-reset: step; list-style: none; padding: 0; margin: 0 0 22px; }
.steps li {
  counter-increment: step;
  position: relative;
  padding-left: 38px;
  margin: 0 0 12px;
}
.steps li::before {
  content: counter(step);
  position: absolute;
  left: 0; top: 2px;
  width: 24px; height: 24px;
  display: grid; place-items: center;
  border-radius: 50%;
  background: var(--bg-inset);
  border: 1px solid var(--border);
  font-family: var(--mono);
  font-size: 12px;
  color: var(--fg-muted);
}

.kicker {
  font-family: var(--mono);
  font-size: 11.5px;
  letter-spacing: 0.09em;
  text-transform: uppercase;
  color: var(--fg-faint);
  margin: 0 0 8px;
}

.pull {
  font-size: 1.28rem;
  line-height: 1.5;
  font-style: italic;
  border-left: 3px solid var(--border-strong);
  padding-left: 18px;
  margin: 28px 0;
  color: var(--fg);
}

/* ---- demo ---------------------------------------------------------- */

.demo-grid { display: grid; grid-template-columns: minmax(0,1fr) 320px; gap: 20px; align-items: start; }
@media (max-width: 900px) { .demo-grid { grid-template-columns: 1fr; } }

.panel {
  background: var(--bg-raised);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}
.panel-head {
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
  font-size: 12px;
  font-family: var(--mono);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--fg-faint);
  display: flex;
  align-items: center;
  gap: 8px;
}
.panel-body { padding: 14px; }
.panel-body > :last-child { margin-bottom: 0; }

.filters { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; margin: 0 0 16px; }
.filters input[type="search"], .filters select {
  font: inherit;
  font-size: 14px;
  padding: 6px 10px;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--bg-raised);
  color: var(--fg);
}
.filters input[type="search"] { flex: 1; min-width: 140px; }

button {
  font: inherit;
  font-size: 14px;
  padding: 6px 13px;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
  background: var(--bg-raised);
  color: var(--fg);
  cursor: pointer;
}
button:hover:not(:disabled) { border-color: var(--fg-faint); }
button:disabled { opacity: 0.55; cursor: default; }
button.primary { background: var(--accent); border-color: var(--accent); color: var(--accent-fg); }

.jobs { width: 100%; font-size: 14px; margin: 0; }
.jobs td, .jobs th { padding: 8px 10px; }
.jobs tr[aria-selected="true"] td { background: var(--accent-soft); }
.jobs .name { font-family: var(--mono); font-size: 13px; }

.pill {
  display: inline-block;
  font-size: 11.5px;
  font-family: var(--mono);
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--bg-inset);
  color: var(--fg-muted);
  white-space: nowrap;
}
.pill-running { background: color-mix(in srgb, var(--accent) 18%, transparent); color: var(--accent); }
.pill-done { background: color-mix(in srgb, var(--ok) 16%, transparent); color: var(--ok); }
.pill-failed { background: color-mix(in srgb, var(--bad) 16%, transparent); color: var(--bad); }
.pill-queued { background: var(--bg-inset); color: var(--fg-muted); }

.bar { height: 5px; border-radius: 999px; background: var(--bg-inset); overflow: hidden; margin-top: 6px; }
.bar > span { display: block; height: 100%; background: var(--accent); }

.log { font-family: var(--mono); font-size: 11.5px; line-height: 1.75; margin: 0; }
.log li { list-style: none; margin: 0; padding: 5px 0; border-bottom: 1px solid var(--border); }
.log li:last-child { border-bottom: 0; }
.log ul { padding: 0; }
.log .m { color: var(--fg-faint); }
.log .t { color: var(--accent); }
.log .skip { color: var(--ok); }
.log .full { color: var(--warn); }

.muted { color: var(--fg-muted); }
.small { font-size: 14px; }
.mono { font-family: var(--mono); }

.empty { color: var(--fg-faint); text-align: center; padding: 28px 0; }

/* A figure wider than the prose column. main is left-aligned inside a wider
   shell, so it grows into the space already there instead of being centred. */
.figure-wide {
  width: min(1080px, calc(100vw - 48px));
  max-width: none;
  margin: 12px 0 32px;
}
.figure-caption {
  font-size: 15px;
  color: var(--fg-faint);
  margin: 0;
  max-width: 68ch;
}

/* ---- hook diagram --------------------------------------------------- */

.hook { margin: 0 0 6px; }

/* The sheet the base scene sits on, and the boxes on it. The boxes are white
   so they read as objects; the sheet is the page colour so it reads as ground. */
.k-sheet      { fill: var(--bg); stroke: var(--border); stroke-width: 1; }
.k-sheetlabel { fill: var(--fg-faint); font: 600 11px var(--mono); letter-spacing: 0.11em; }
.k-sheetsub   { fill: var(--fg-faint); font: 13px var(--sans); }
/* The whole idea in one line, so the picture has a spine to hang on. */
.k-equation   { fill: var(--fg); font: 700 17px var(--mono); letter-spacing: -0.01em; }

.k-box    { fill: var(--bg-raised); stroke: var(--border-strong); stroke-width: 1; }
.k-chrome { fill: var(--bg-inset); }
.k-dot    { fill: var(--border-strong); }
.k-cap    { fill: var(--fg-faint); font: 600 10px var(--mono); letter-spacing: 0.12em; }

/* The address bar is the only thing in the base drawing that gets the accent,
   because it is the thing the base drawing is about. */
.k-url  { fill: var(--accent-soft); stroke: var(--accent); stroke-width: 1.2; }
.k-urlt { fill: var(--accent); font: 600 12.5px var(--mono); }

.k-page    { fill: var(--bg); stroke: var(--border); stroke-width: 1; }
.k-bar     { fill: var(--bg-inset); }
.k-frag    { fill: none; stroke: var(--accent); stroke-width: 1.4; stroke-dasharray: 4 3; }
.k-fragt   { fill: var(--accent); font: 600 11px var(--mono); }
.k-fragbar { fill: var(--accent); opacity: 0.28; }

.k-code     { fill: var(--fg-muted); font: 12px var(--mono); }
.k-code-hot { fill: var(--accent); font-weight: 600; }
.k-rule     { stroke: var(--border); stroke-width: 1; }
.k-note     { fill: var(--fg); font: 13px var(--sans); }

.k-wire     { fill: var(--fg-muted); font: 12.5px var(--mono); }
.k-wire-fx  { fill: var(--accent); font-weight: 600; }
.k-say      { fill: var(--fg-faint); font: 13px var(--sans); }

.k-line      { stroke: var(--border-strong); stroke-width: 1.5; fill: none; }
.k-line-dash { stroke-dasharray: 5 4; }
.k-line-fx   { stroke: var(--accent); stroke-width: 1.8; }
.k-head      { fill: var(--border-strong); }
.k-head-fx   { fill: var(--accent); }

/* The layer. Lifted with a shadow, breaking the top edge of the sheet below
   it, so it reads as sitting on top rather than sitting beside. */
.k-layer      { fill: var(--accent-soft); stroke: var(--accent); stroke-width: 1.4; }
.k-layerlabel { fill: var(--accent); font: 700 11.5px var(--mono); letter-spacing: 0.11em; }
.k-layertext  { fill: var(--fg); font: 13.5px var(--sans); }
.k-header     { fill: var(--accent); font: 700 12.5px var(--mono); }
.k-leader     { stroke: var(--accent); stroke-width: 1.4; stroke-dasharray: 4 4; fill: none; opacity: 0.7; }

.k-chip  { fill: var(--accent); }
.k-chipt { fill: var(--accent-fg); font: 600 12px var(--mono); }

/* ---- pattern diagrams ----------------------------------------------- */

.pattern { margin: 0; }

.p-box   { fill: var(--bg-raised); stroke: var(--border-strong); stroke-width: 1; }
.p-doc   { fill: var(--bg); stroke: var(--border); stroke-width: 1; }
.p-cap   { fill: var(--fg-faint); font: 600 10px var(--mono); letter-spacing: 0.12em; }
.p-doclabel { fill: var(--fg-faint); font: 13px var(--sans); }

/* Hungry pieces sit outside the hole and get swapped anyway. */
.p-hungry    { fill: var(--bg-raised); stroke: var(--border-strong); stroke-width: 1; stroke-dasharray: 5 4; }
.p-hungrytag { fill: var(--accent); font: 600 10px var(--mono); letter-spacing: 0.1em; }

/* The hole every link targets. */
.p-hole  { fill: var(--accent-soft); stroke: var(--accent); stroke-width: 1.3; }
.p-ghost { fill: var(--bg-inset); }
.p-gone  { fill: var(--bg-inset); stroke: var(--border); stroke-width: 1; stroke-dasharray: 4 4; }
.p-gonet { fill: var(--fg-faint); font: 12px var(--sans); }

.p-tag     { fill: var(--fg-muted); font: 12px var(--mono); }
.p-tag-hot { fill: var(--accent); font-weight: 600; }
.p-code     { fill: var(--fg-muted); font: 12.5px var(--mono); }
.p-code-hot { fill: var(--accent); font-weight: 600; }
.p-note    { fill: var(--fg-muted); font: 13px var(--sans); }
.p-caption { fill: var(--fg-faint); font: 13.5px var(--sans); }

.p-url      { fill: var(--bg-inset); }
.p-url-hot  { fill: var(--accent-soft); stroke: var(--accent); stroke-width: 1.2; }
.p-urlt     { fill: var(--fg-muted); font: 12.5px var(--mono); }
.p-urlt-hot { fill: var(--accent); font-weight: 600; }

.p-field   { fill: var(--bg); stroke: var(--border-strong); stroke-width: 1; }
.p-fieldt  { fill: var(--fg-muted); font: 12.5px var(--mono); }
.p-submit  { fill: var(--accent); }
.p-submitt { fill: var(--accent-fg); font: 600 12.5px var(--sans); }

.p-pane  { fill: var(--accent-soft); stroke: var(--accent); stroke-width: 1.3; }
.p-panet { fill: var(--accent); font: 600 12px var(--mono); }

.p-hint    { fill: var(--fg-faint); font: 11.5px var(--mono); }
.p-hint-fx { fill: var(--accent); font-weight: 600; }

.p-line    { stroke: var(--border-strong); stroke-width: 1.5; fill: none; }
.p-line-fx { stroke: var(--accent); stroke-width: 1.8; }
.p-head    { fill: var(--border-strong); }
.p-head-fx { fill: var(--accent); }

/* Rendered against skipped, in the selective-rendering diagram. */
.p-run   { fill: var(--accent); opacity: 0.14; }
.p-runt  { fill: var(--fg); font: 12px var(--mono); }
.p-skip  { fill: var(--bg-inset); }
.p-skipt { fill: var(--fg-faint); font: 12px var(--mono); }
.p-state { font-size: 11.5px; }
.p-total { fill: var(--fg); font: 700 15px var(--mono); }

/* ---- lifecycle diagram --------------------------------------------- */

.diagram {
  display: block;
  width: 100%;
  height: auto;
  margin: 8px 0 28px;
  font-family: var(--sans);
  overflow: visible;
}

.d-kicker { fill: var(--fg-faint); font: 600 11px var(--mono); letter-spacing: 0.1em; }
.d-note   { fill: var(--fg-muted); font: 13px var(--sans); }
.d-url    { fill: var(--accent-soft); stroke: var(--accent); stroke-opacity: 0.35; }
.d-mono   { fill: var(--accent); font: 600 13px var(--mono); }

.d-badge   { fill: var(--bg-inset); stroke: var(--border-strong); }
.d-badge-n { fill: var(--fg-muted); font: 600 12px var(--mono); text-anchor: middle; }
.d-title   { fill: var(--fg); font: 640 15px var(--sans); }
.d-sub     { fill: var(--fg-muted); font: 12.5px var(--sans); }

.d-box     { fill: var(--bg-raised); stroke: var(--border-strong); }
.d-box-hot { stroke: var(--accent); stroke-dasharray: 5 3; }
.d-cap     { fill: var(--fg-faint); font: 600 9.5px var(--mono); letter-spacing: 0.12em; }

.d-frag        { fill: var(--bg-inset); stroke: none; }
.d-frag-swap   { fill: var(--accent); }
.d-frag-run    { fill: color-mix(in srgb, var(--ok) 20%, transparent); }
.d-frag-skip   { fill: none; stroke: var(--border-strong); stroke-dasharray: 3 3; }
.d-frag-t      { fill: var(--fg-muted); font: 11px var(--mono); }
.d-frag-t-swap { fill: var(--accent-fg); font-weight: 600; }
.d-frag-t-skip { fill: var(--fg-faint); }

.d-swap-tag { fill: var(--accent); font: 600 10.5px var(--mono); }
.d-swap-in  { fill: var(--accent-fg); font: 600 9.5px var(--mono); opacity: 0.85; }
.d-state      { fill: var(--ok); font: 10px var(--mono); }
.d-state-skip { fill: var(--fg-faint); }
.d-total      { fill: var(--fg-muted); font: 600 11px var(--mono); }

.d-line      { stroke: var(--border-strong); stroke-width: 1.5; }
.d-line-dash { stroke-dasharray: 5 4; }
.d-head      { fill: var(--border-strong); }
.d-wire      { fill: var(--fg-muted); font: 11px var(--mono); }
.d-wire-hot  { fill: var(--accent); font-weight: 600; }

.d-out      { fill: var(--fg); font: 640 15px var(--sans); }
.d-out-fast { fill: var(--accent); font-family: var(--mono); font-size: 14px; }
.d-out-sub  { fill: var(--fg-muted); font: 12px var(--sans); }

@media (max-width: 720px) {
  .diagram { min-width: 640px; }
  .diagram-scroll { overflow-x: auto; }
}

.skip-link {
  position: absolute;
  left: -9999px;
}
.skip-link:focus {
  left: 12px; top: 12px;
  z-index: 200;
  background: var(--bg-raised);
  padding: 8px 14px;
  border-radius: 8px;
  border: 1px solid var(--border-strong);
}
`
