// The app's page script (/app.js), run against a stand-in for the job page's
// live result card: what the card draws of the line's speed once PacketPulse
// has graded the test - its dials and its tiles. Run by the field app's tests
// with the script's own text and the /result.json answer to draw, as JSON.
const fs = require('fs');
const js = fs.readFileSync(process.argv[2], 'utf8');
const answer = JSON.parse(process.argv[3]);

function node(tag) {
  return {
    tag, className: '', children: [], attributes: {}, _text: '',
    appendChild(child) { this.children.push(child); return child; },
    setAttribute(name, value) { this.attributes[name] = String(value); },
    set textContent(text) { this._text = text; this.children = []; },
    get textContent() { return this._text + this.children.map(c => c.textContent).join(''); },
  };
}
const body = node('div');
const live = { querySelector: s => s === '.live-body' ? body : null, getAttribute: a => a === 'data-tt' ? 'TT-2026-01458' : null };
const document = {
  hidden: false,
  getElementById: id => id === 'live-result' ? live : null,
  querySelectorAll: () => [],
  createElement: node,
  createElementNS: (namespace, tag) => (namespace === 'http://www.w3.org/2000/svg' ? node(tag) : null),
  addEventListener: () => {},
};
const window = { addEventListener: () => {} };
const fetch = () => Promise.resolve({ json: () => Promise.resolve(answer) });
new Function('document', 'window', 'location', 'setTimeout', 'fetch', js)(document, window, {}, () => 0, fetch);

// Each tile as "Way: what it says", and each dial as what it draws.
setImmediate(() => {
  const tiles = body.children.filter(c => c.className === 'tiles');
  const read = tiles.length === 0 ? [] : tiles[0].children.map(t => t.children.map(c => c.textContent).join(' '));
  const dials = body.children.filter(c => c.className === 'dials');
  const drawn = dials.length === 0 ? [] : dials[0].children.map((wrap) => {
    const svg = wrap.children[0];
    const part = name => svg.children.find(c => c.attributes.class === name);
    return {
      class: wrap.className,
      way: wrap.children[1].textContent,
      figure: part('dial-figure').textContent,
      needle: [part('dial-needle').attributes.x2, part('dial-needle').attributes.y2],
      arc: part('dial-arc') ? part('dial-arc').attributes.d : '',
      track: part('dial-track').attributes.d,
      ticks: svg.children.filter(c => c.attributes.class === 'dial-tick').map(c => [c.attributes.x1, c.attributes.y1, c.attributes.x2, c.attributes.y2]),
      hidden: svg.attributes['aria-hidden'],
    };
  });
  console.log(JSON.stringify({ tiles: read, dials: drawn }));
});
