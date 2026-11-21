package fieldengapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
)

// The files a phone installs the app from - its manifest and icons - and
// the page script. All are made here, so the app is one program with no
// files beside it, and serves nothing from anywhere else.

// brand and brandDeep are the app's colours: its icon, its bar and the
// phone's status bar.
var (
	brand     = color.NRGBA{0x0b, 0x5d, 0x7a, 0xff}
	brandDeep = color.NRGBA{0x07, 0x3b, 0x55, 0xff}
)

// iconSVG is the app's mark: a rounded square with a line's pulse across it.
const iconSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
	`<defs><linearGradient id="g" x1="0" y1="0" x2="0" y2="1">` +
	`<stop offset="0" stop-color="#0e7fa6"/><stop offset="1" stop-color="#073b55"/></linearGradient></defs>` +
	`<rect width="64" height="64" rx="14" fill="url(#g)"/>` +
	`<polyline points="10,36 22,36 27,22 33,46 38,30 42,36 54,36" fill="none" stroke="#fff" ` +
	`stroke-width="4.5" stroke-linecap="round" stroke-linejoin="round"/></svg>`

// pulse is the mark's line, on a 64-unit square.
var pulse = [][2]float64{{10, 36}, {22, 36}, {27, 22}, {33, 46}, {38, 30}, {42, 36}, {54, 36}}

// iconPNG draws the mark at size pixels: four times as large, then each
// 4 x 4 block averaged, so its edges are smooth.
func iconPNG(size int) []byte {
	const scale = 4
	big := size * scale
	unit := float64(big) / 64
	radius, stroke := 14*unit, 4.5*unit/2
	canvas := image.NewNRGBA(image.Rect(0, 0, big, big))
	for y := 0; y < big; y++ {
		shade := float64(y) / float64(big)
		fill := color.NRGBA{
			uint8(float64(0x0e)*(1-shade) + float64(brandDeep.R)*shade),
			uint8(float64(0x7f)*(1-shade) + float64(brandDeep.G)*shade),
			uint8(float64(0xa6)*(1-shade) + float64(brandDeep.B)*shade), 0xff}
		for x := 0; x < big; x++ {
			if !insideRounded(float64(x)+0.5, float64(y)+0.5, float64(big), radius) {
				continue
			}
			if nearLine(float64(x)+0.5, float64(y)+0.5, unit, stroke) {
				canvas.SetNRGBA(x, y, color.NRGBA{0xff, 0xff, 0xff, 0xff})
			} else {
				canvas.SetNRGBA(x, y, fill)
			}
		}
	}
	small := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var r, g, b, a int
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					c := canvas.NRGBAAt(x*scale+dx, y*scale+dy)
					r, g, b, a = r+int(c.R)*int(c.A), g+int(c.G)*int(c.A), b+int(c.B)*int(c.A), a+int(c.A)
				}
			}
			if a > 0 {
				small.SetNRGBA(x, y, color.NRGBA{uint8(r / a), uint8(g / a), uint8(b / a), uint8(a / (scale * scale))})
			}
		}
	}
	var out bytes.Buffer
	_ = png.Encode(&out, small)
	return out.Bytes()
}

// insideRounded is whether a point lies in a side x side square with
// corners of radius r.
func insideRounded(x, y, side, r float64) bool {
	cx := math.Max(r, math.Min(x, side-r))
	cy := math.Max(r, math.Min(y, side-r))
	return (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r
}

// nearLine is whether a point lies within half a stroke of the pulse.
func nearLine(x, y, unit, half float64) bool {
	for i := 0; i+1 < len(pulse); i++ {
		ax, ay := pulse[i][0]*unit, pulse[i][1]*unit
		bx, by := pulse[i+1][0]*unit, pulse[i+1][1]*unit
		dx, dy := bx-ax, by-ay
		t := math.Max(0, math.Min(1, ((x-ax)*dx+(y-ay)*dy)/(dx*dx+dy*dy)))
		px, py := ax+t*dx-x, ay+t*dy-y
		if px*px+py*py <= half*half {
			return true
		}
	}
	return false
}

// manifest lets a phone install the app on its home screen and open it on
// its own, full screen, as the operator's real app opens. A link it then
// follows to PacketPulse opens in the phone's in-app browser, as the real
// app's does.
func manifest(appName string) []byte {
	body, _ := json.Marshal(map[string]any{
		"name": appName, "short_name": appName, "start_url": "/", "scope": "/",
		"display": "standalone", "orientation": "portrait",
		"background_color": "#0b1e2d", "theme_color": "#0b5d7a",
		"description": "Today's jobs, and a one-tap line test in PacketPulse.",
		"icons": []map[string]string{
			{"src": "/icon-192.png", "sizes": "192x192", "type": "image/png"},
			{"src": "/icon-512.png", "sizes": "512x512", "type": "image/png", "purpose": "any maskable"},
			{"src": "/icon.svg", "sizes": "any", "type": "image/svg+xml"},
		},
	})
	return body
}

// script is the pages' only script: search and filters on the jobs list,
// and the moment between tapping the test button and PacketPulse opening.
// Every page works without it.
const script = `(function () {
  var q = document.getElementById('q');
  var jobs = Array.prototype.slice.call(document.querySelectorAll('[data-job]'));
  var chips = Array.prototype.slice.call(document.querySelectorAll('[data-filter]'));
  var none = document.getElementById('none');
  var filter = 'open';
  function show() {
    var term = q ? q.value.trim().toLowerCase() : '';
    var shown = 0;
    jobs.forEach(function (job) {
      var status = job.getAttribute('data-status');
      var keep = (filter === 'all') ||
        (filter === 'open' && status !== 'closed') ||
        (filter === 'p1' && job.getAttribute('data-priority') === 'P1' && status !== 'closed') ||
        (filter === 'closed' && status === 'closed');
      keep = keep && (!term || job.getAttribute('data-search').indexOf(term) >= 0);
      job.hidden = !keep;
      if (keep) shown++;
    });
    if (none) none.hidden = shown !== 0;
  }
  chips.forEach(function (chip) {
    chip.addEventListener('click', function () {
      filter = chip.getAttribute('data-filter');
      chips.forEach(function (c) { c.setAttribute('aria-pressed', c === chip ? 'true' : 'false'); });
      show();
    });
  });
  if (q) q.addEventListener('input', show);
  if (jobs.length) show();
  // Tapping the test button: the moment until PacketPulse opens. In a
  // browser tab PacketPulse takes this page's place; installed on a phone,
  // the app opens it in a browser layer over itself and this page stays as
  // it was, under the overlay. So the overlay goes the moment the engineer
  // is back in the app - which then shows the job as it is now, waiting for
  // its result - and after twenty seconds it says what to do.
  var opening = document.getElementById('opening');
  var launched = null;
  var launchButtons = function (disabled) {
    Array.prototype.slice.call(document.querySelectorAll('form[data-launch] button')).forEach(function (b) { b.disabled = disabled; });
  };
  var backToJob = function () {
    var tt = launched && launched.tt;
    launched = null;
    if (tt) {
      location.replace('/ticket?tt=' + encodeURIComponent(tt) + '&back=1');
      return;
    }
    if (opening) { opening.hidden = true; opening.classList.remove('slow'); }
    launchButtons(false);
  };
  Array.prototype.slice.call(document.querySelectorAll('form[data-launch]')).forEach(function (form) {
    form.addEventListener('submit', function () {
      var ticket = form.querySelector('input[name=tt]');
      launched = { tt: ticket ? ticket.value : '', away: false };
      if (opening) { opening.hidden = false; opening.classList.remove('slow'); }
      launchButtons(true);
      setTimeout(function () {
        if (launched && !launched.away && opening) opening.classList.add('slow');
      }, 20000);
    });
  });
  document.addEventListener('visibilitychange', function () {
    if (!launched) return;
    if (document.hidden) { launched.away = true; return; }
    if (launched.away) backToJob();
  });
  var dismiss = document.getElementById('opening-dismiss');
  if (dismiss) dismiss.addEventListener('click', backToJob);
  // The job's result, as it lands: asked for every few seconds once the
  // engineer is back, until PacketPulse has graded it. Built from text
  // only, never from the server's markup.
  var live = document.getElementById('live-result');
  if (live) {
    var body = live.querySelector('.live-body');
    var tt = live.getAttribute('data-tt');
    var timer = null, since = Date.now();
    var element = function (tag, className, text) {
      var node = document.createElement(tag);
      if (className) node.className = className;
      if (text) node.textContent = text;
      return node;
    };
    var waiting = function (text) {
      var row = element('div', 'waiting');
      row.appendChild(element('div', 'spinner small'));
      row.appendChild(element('span', '', text));
      return row;
    };
    // A way of the line's speed as a dial, drawn by the numbers the result
    // page's own dials are (newSpeedDial): its marks, each step between two
    // the same share of a 240-degree sweep about the hub at 100,86.
    var marks = [0, 1, 5, 10, 25, 50, 100, 250, 500];
    var position = function (mbps) {
      if (!(mbps > marks[0])) return 0;
      for (var i = 1; i < marks.length; i++) {
        if (mbps <= marks[i]) return (i - 1 + (mbps - marks[i - 1]) / (marks[i] - marks[i - 1])) / (marks.length - 1);
      }
      return 1;
    };
    var point = function (share, radius) {
      var angle = (150 + 240 * share) * Math.PI / 180;
      return [(100 + radius * Math.cos(angle)).toFixed(1), (86 + radius * Math.sin(angle)).toFixed(1)];
    };
    var arc = function (share) {
      var from = point(0, 70), to = point(share, 70);
      return 'M' + from[0] + ',' + from[1] + ' A70,70 0 ' + (240 * share > 180 ? 1 : 0) + ' 1 ' + to[0] + ',' + to[1];
    };
    var drawn = function (tag, attributes, text) {
      var node = document.createElementNS('http://www.w3.org/2000/svg', tag);
      Object.keys(attributes).forEach(function (name) { node.setAttribute(name, attributes[name]); });
      if (text) node.textContent = text;
      return node;
    };
    var dial = function (way, figure, upload) {
      var share = figure ? position(parseFloat(figure)) : 0;
      var wrap = element('div', 'dial' + (upload ? ' up' : '') + (figure ? '' : ' none'));
      var svg = drawn('svg', { viewBox: '0 0 200 150', 'aria-hidden': 'true', focusable: 'false' });
      svg.appendChild(drawn('path', { 'class': 'dial-track', d: arc(1) }));
      if (share > 0) svg.appendChild(drawn('path', { 'class': 'dial-arc', d: arc(share) }));
      marks.forEach(function (mark, i) {
        var outer = point(i / (marks.length - 1), 61), inner = point(i / (marks.length - 1), 55);
        svg.appendChild(drawn('line', { 'class': 'dial-tick', x1: outer[0], y1: outer[1], x2: inner[0], y2: inner[1] }));
      });
      var tip = point(share, 44);
      svg.appendChild(drawn('line', { 'class': 'dial-needle', x1: '100', y1: '86', x2: tip[0], y2: tip[1] }));
      svg.appendChild(drawn('circle', { 'class': 'dial-hub', cx: '100', cy: '86', r: '6' }));
      svg.appendChild(drawn('text', { 'class': 'dial-figure', x: '100', y: '124' }, figure || '—'));
      svg.appendChild(drawn('text', { 'class': 'dial-unit', x: '100', y: '141' }, 'Mbit/s'));
      wrap.appendChild(svg);
      wrap.appendChild(element('small', '', (upload ? '↑ ' : '↓ ') + way));
      return wrap;
    };
    var render = function (result) {
      body.textContent = '';
      if (result.state === 'done') {
        var verdict = element('div', 'verdict tone-' + result.tone);
        var words = element('div');
        words.appendChild(element('h2', '', result.verdict));
        words.appendChild(element('p', '', result.note));
        var who = [result.tested_by, result.when].filter(Boolean).join(' · ');
        if (who) words.appendChild(element('small', '', 'Tested by ' + who));
        verdict.appendChild(words);
        body.appendChild(verdict);
        if (result.download || result.upload) {
          // The dials, subtle, each resting at its figure - then the tiles.
          var dials = element('div', 'dials');
          dials.appendChild(dial('Download', result.download, false));
          dials.appendChild(dial('Upload', result.upload, true));
          body.appendChild(dials);
          var tiles = element('div', 'tiles');
          // Both ways once either has a figure: a way PacketPulse filed no
          // figure for reads Not recorded, as the result page says it.
          [['Download', result.download], ['Upload', result.upload]].forEach(function (pair) {
            var tile = element('div');
            tile.appendChild(element('small', '', pair[0]));
            if (pair[1]) {
              tile.appendChild(element('b', '', pair[1]));
              tile.appendChild(element('small', '', 'Mbit/s'));
            } else {
              tile.appendChild(element('b', 'unrecorded', 'Not recorded'));
            }
            tiles.appendChild(tile);
          });
          body.appendChild(tiles);
        }
      } else if (result.state === 'running') {
        body.appendChild(waiting('PacketPulse is testing the line…'));
      } else if (result.state === 'none') {
        body.appendChild(waiting('Waiting for the test to be filed in PacketPulse…'));
      } else {
        body.appendChild(element('p', 'muted small', result.message));
      }
    };
    var poll = function () {
      clearTimeout(timer);
      fetch('/result.json?tt=' + encodeURIComponent(tt), { credentials: 'same-origin', headers: { Accept: 'application/json' } })
        .then(function (response) { return response.json(); })
        .then(function (result) {
          render(result);
          if ((result.state === 'none' || result.state === 'running') && Date.now() - since < 300000) {
            timer = setTimeout(poll, 4000);
          }
        })
        .catch(function () {
          if (Date.now() - since < 300000) timer = setTimeout(poll, 8000);
        });
    };
    poll();
    document.addEventListener('visibilitychange', function () {
      if (!document.hidden) { since = Date.now(); poll(); }
    });
  }
  // Back from PacketPulse by the browser's own Back: the page comes back
  // as it was left, so it is taken to the job as it is now.
  window.addEventListener('pageshow', function (event) {
    if (event.persisted) backToJob();
  });
})();
`

// scriptPath is where pages load the script from: named by its content, so
// a release with a changed script is a new address, and a phone that kept
// the old one for its day of caching fetches the new one at once.
var scriptPath = func() string {
	sum := sha256.Sum256([]byte(script))
	return "/app.js?v=" + hex.EncodeToString(sum[:6])
}()

// asset serves one of the files above, cached for a day: they change only
// with a new release of the app - and the script's address changes with it.
func asset(w http.ResponseWriter, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(body)
}
