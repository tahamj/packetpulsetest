// The app's page script (/app.js), run against a stand-in for a phone's
// page: what happens to the "Opening PacketPulse" overlay when PacketPulse
// opens over the app, and the engineer comes back. Run by
// TestTheAppNeverLeavesTheEngineerOnItsOverlay with the script's own text.
const fs = require('fs');
const js = fs.readFileSync(process.argv[2], 'utf8');
function world() {
  const listeners = {}, docListeners = {}, timers = [];
  const classes = new Set();
  const overlay = { hidden: true, classList: { add: c => classes.add(c), remove: c => classes.delete(c), has: c => classes.has(c) } };
  const button = { disabled: false };
  const input = { value: 'TT-2026-01458' };
  const formListeners = {};
  const form = { querySelector: s => s === 'button' ? button : input, addEventListener: (e, f) => { formListeners[e] = f; } };
  const dismissListeners = {};
  const dismiss = { addEventListener: (e, f) => { dismissListeners[e] = f; } };
  const replaced = [];
  const document = {
    hidden: false,
    getElementById: id => ({ opening: overlay, 'opening-dismiss': dismiss }[id] || null),
    querySelectorAll: s => s === 'form[data-launch]' ? [form] : s === 'form[data-launch] button' ? [button] : [],
    addEventListener: (e, f) => { (docListeners[e] = docListeners[e] || []).push(f); },
  };
  const window = { addEventListener: (e, f) => { listeners[e] = f; } };
  const location = { replace: url => replaced.push(url) };
  const setTimeout = (f, ms) => timers.push([f, ms]);
  new Function('document', 'window', 'location', 'setTimeout', 'fetch', js)(document, window, location, setTimeout, null);
  return { overlay, classes, button, formListeners, dismissListeners, docListeners, listeners, timers, replaced, document };
}
const assert = (ok, what) => { if (!ok) { console.error('FAIL ' + what); process.exitCode = 1; } else console.log('ok   ' + what); };

let w = world();
w.formListeners.submit();
assert(!w.overlay.hidden && w.button.disabled, 'a tap shows the overlay and holds the button');
w.document.hidden = true; w.docListeners.visibilitychange.forEach(f => f());
w.document.hidden = false; w.docListeners.visibilitychange.forEach(f => f());
assert(w.replaced[0] === '/ticket?tt=TT-2026-01458&back=1', 'back in the app, the job waits for its result: ' + w.replaced[0]);

w = world();
w.formListeners.submit();
w.timers.find(t => t[1] === 20000)[0]();
assert(w.classes.has('slow'), 'twenty seconds on, the overlay says what to do');
w.dismissListeners.click();
assert(w.replaced[0] === '/ticket?tt=TT-2026-01458&back=1', 'Back to the job takes the engineer to it');

w = world();
w.formListeners.submit();
w.listeners.pageshow({ persisted: true });
assert(w.replaced[0] === '/ticket?tt=TT-2026-01458&back=1', 'the browser Back lands on the job, not the overlay');

w = world();
w.document.hidden = false; w.docListeners.visibilitychange.forEach(f => f());
assert(w.replaced.length === 0 && w.overlay.hidden, 'without a tap, coming back changes nothing');
