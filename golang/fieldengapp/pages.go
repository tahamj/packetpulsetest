package fieldengapp

// pages are the app's screens: an operator's own field app, deliberately
// unlike PacketPulse's look, so a demonstration shows plainly where the
// operator's app ends and PacketPulse begins. Server pages with their styles
// inline and one small script beside them (/app.js), which every page works
// without; nothing is loaded from anywhere else.
const pages = `
{{define "logo"}}<svg class="logo-mark" viewBox="0 0 64 64" aria-hidden="true"><defs><linearGradient id="lg" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#0e7fa6"/><stop offset="1" stop-color="#073b55"/></linearGradient></defs><rect width="64" height="64" rx="14" fill="url(#lg)"/><polyline points="10,36 22,36 27,22 33,46 38,30 42,36 54,36" fill="none" stroke="#fff" stroke-width="4.5" stroke-linecap="round" stroke-linejoin="round"/></svg>{{end}}
{{define "i-back"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M15 18l-6-6 6-6"/></svg>{{end}}
{{define "i-chevron"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M9 18l6-6-6-6"/></svg>{{end}}
{{define "i-clock"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg>{{end}}
{{define "i-alarm"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="13" r="8"/><path d="M12 9v4l2 2M5 3L2 6M19 3l3 3"/></svg>{{end}}
{{define "i-pin"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 21s-7-6.1-7-11.5A7 7 0 0 1 19 9.5C19 14.9 12 21 12 21z"/><circle cx="12" cy="9.5" r="2.5"/></svg>{{end}}
{{define "i-search"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="11" cy="11" r="7"/><path d="M20 20l-3.5-3.5"/></svg>{{end}}
{{define "i-check"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M8 12.5l2.5 2.5L16 9.5"/></svg>{{end}}
{{define "i-x"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M9 9l6 6M15 9l-6 6"/></svg>{{end}}
{{define "i-alert"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3l9.5 17h-19z"/><path d="M12 10v4M12 17.5v.5"/></svg>{{end}}
{{define "i-info"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M12 11v6M12 7.5v.5"/></svg>{{end}}
{{define "i-refresh"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M20 12a8 8 0 1 1-2.3-5.6M20 4v5h-5"/></svg>{{end}}
{{define "i-user"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="8" r="4"/><path d="M4 21c1.5-4 4.5-6 8-6s6.5 2 8 6"/></svg>{{end}}
{{define "i-lock"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><rect x="5" y="11" width="14" height="10" rx="2"/><path d="M8 11V8a4 4 0 0 1 8 0v3"/></svg>{{end}}
{{define "i-pulse"}}<svg class="i" viewBox="0 0 24 24" aria-hidden="true"><path d="M2 13h4l2-6 4 12 3-9 2 3h5"/></svg>{{end}}

{{define "head"}}<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
<meta name="theme-color" content="#0b5d7a">
<meta name="color-scheme" content="light dark">
<meta name="mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="apple-mobile-web-app-title" content="{{.App}}">
<link rel="manifest" href="/manifest.webmanifest">
<link rel="icon" href="/icon.svg" type="image/svg+xml">
<link rel="apple-touch-icon" href="/icon-192.png">
<title>{{if .Title}}{{.Title}} · {{end}}{{.App}}</title>
<style>
:root{--bg:#eef2f6;--backdrop:#dfe6ee;--card:#fff;--ink:#0f2233;--muted:#5b6d7e;--line:#dde5ec;--soft:#f5f8fa;
--brand:#0b5d7a;--brand-2:#0e7fa6;--brand-deep:#073b55;--on-brand:#fff;--go:#0f7b45;--go-2:#12965a;
--good:#0f7b45;--good-bg:#e5f5ec;--warn:#8f5200;--warn-bg:#fff3dc;--bad:#b42318;--bad-bg:#fdeceb;--info:#0b5d7a;--info-bg:#e4f1f7;--neutral:#5b6d7e;--neutral-bg:#edf1f4;
--radius:16px;--shadow:0 1px 2px rgba(15,34,51,.06),0 4px 14px rgba(15,34,51,.06);--dial-down:#0e7fa6;--dial-up:#7c4dcc}
@media (prefers-color-scheme:dark){:root{--bg:#0b141c;--backdrop:#060c11;--card:#131f2a;--ink:#e7eef4;--muted:#9aaebf;--line:#22323f;--soft:#182633;
--brand:#38a3cc;--brand-2:#5bb8dc;--brand-deep:#0b3a52;--on-brand:#03131c;--go:#3cc17f;--go-2:#5fd497;
--good:#5fd497;--good-bg:#11301f;--warn:#ffc266;--warn-bg:#33240a;--bad:#ff8a80;--bad-bg:#3a1512;--info:#7cc8e8;--info-bg:#0f2a38;--neutral:#9aaebf;--neutral-bg:#1b2833;--shadow:none;--dial-down:#5bb8dc;--dial-up:#b39cff}}
*{box-sizing:border-box}html{-webkit-text-size-adjust:100%}
body{margin:0;background:var(--bg);color:var(--ink);font:16px/1.45 system-ui,-apple-system,"Segoe UI",Roboto,"Noto Sans",sans-serif;-webkit-font-smoothing:antialiased}
.app{max-width:520px;margin:0 auto;min-height:100vh;display:flex;flex-direction:column;background:var(--bg)}
@media (min-width:640px){body{background:var(--backdrop)}.app{margin:24px auto;min-height:calc(100vh - 48px);border-radius:28px;overflow:hidden;box-shadow:0 20px 60px rgba(0,0,0,.18)}}
.i{width:20px;height:20px;flex:none;fill:none;stroke:currentColor;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.logo-mark{width:32px;height:32px;display:block;border-radius:9px}
.bar{position:sticky;top:0;z-index:5;display:flex;align-items:center;gap:10px;padding:calc(10px + env(safe-area-inset-top)) 14px 10px;
background:linear-gradient(135deg,var(--brand-2),var(--brand-deep));color:#fff}
.bar .title{flex:1;min-width:0}.bar .title strong{display:block;font-size:17px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.iconbutton{display:grid;place-items:center;width:40px;height:40px;border-radius:12px;color:#fff;text-decoration:none}
.iconbutton:hover,.iconbutton:focus-visible{background:rgba(255,255,255,.15)}
.me{position:relative}.me summary{list-style:none;cursor:pointer}.me summary::-webkit-details-marker{display:none}
.avatar{display:grid;place-items:center;width:38px;height:38px;border-radius:50%;background:rgba(255,255,255,.18);color:#fff;font-weight:700;font-size:14px;flex:none}
.menu{position:absolute;right:0;top:48px;width:250px;background:var(--card);color:var(--ink);border:1px solid var(--line);border-radius:14px;padding:14px;box-shadow:0 12px 32px rgba(0,0,0,.18);display:grid;gap:4px}
.menu small{color:var(--muted);overflow-wrap:anywhere}.menu form{margin-top:8px}
main{padding:16px 16px calc(16px + env(safe-area-inset-bottom));display:grid;gap:14px;align-content:start;flex:1}
.card{background:var(--card);border:1px solid var(--line);border-radius:var(--radius);padding:16px;box-shadow:var(--shadow)}
h1,h2,h3{margin:0;line-height:1.25}h1{font-size:22px}h2{font-size:17px}h3{font-size:13px;text-transform:uppercase;letter-spacing:.06em;color:var(--muted);margin-bottom:10px}
p{margin:0}.muted{color:var(--muted)}.small,small{font-size:13px}.center{text-align:center}.mono,.tt{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}
.tt{font-size:13px;font-weight:600;color:var(--brand);letter-spacing:.02em}
.row{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.chip{display:inline-flex;align-items:center;gap:4px;padding:2px 9px;border-radius:999px;font-size:12px;font-weight:700;letter-spacing:.02em}
.chip.bad{background:var(--bad-bg);color:var(--bad)}.chip.warn{background:var(--warn-bg);color:var(--warn)}.chip.info{background:var(--info-bg);color:var(--info)}.chip.good{background:var(--good-bg);color:var(--good)}
.chip.status-assigned{background:var(--neutral-bg);color:var(--neutral)}.chip.status-testing{background:var(--info-bg);color:var(--info)}.chip.status-closed{background:var(--good-bg);color:var(--good)}
.button{display:flex;align-items:center;justify-content:center;gap:8px;width:100%;min-height:52px;padding:12px 16px;border:0;border-radius:14px;font:inherit;font-weight:700;font-size:16px;cursor:pointer;text-decoration:none;transition:transform .08s ease,filter .15s ease}
.button:active{transform:scale(.985)}.button:disabled{filter:grayscale(.4) opacity(.7);cursor:progress}
.primary{background:var(--brand);color:var(--on-brand)}.go{background:linear-gradient(135deg,var(--go-2),var(--go));color:#fff;box-shadow:0 6px 18px rgba(15,123,69,.28)}
.secondary{background:transparent;color:var(--brand);border:1.5px solid var(--line)}.secondary:hover{border-color:var(--brand)}
.button.small{min-height:40px;font-size:14px}
:focus-visible{outline:3px solid var(--brand-2);outline-offset:2px}
.welcome{margin:-16px -16px 0;padding:calc(48px + env(safe-area-inset-top)) 24px 56px;text-align:center;color:#fff;background:radial-gradient(120% 90% at 50% 0%,var(--brand-2),var(--brand-deep))}
.welcome .logo-mark{width:72px;height:72px;margin:0 auto 16px;border-radius:20px;box-shadow:0 10px 30px rgba(0,0,0,.25)}
.welcome h1{font-size:26px}.welcome p{opacity:.88;margin-top:6px}
.form{margin-top:-40px}
label{display:block;font-size:14px;font-weight:600;margin:14px 0 6px}
.field{display:flex;align-items:center;gap:10px;border:1.5px solid var(--line);border-radius:14px;padding:0 14px;background:var(--soft)}
.field:focus-within{border-color:var(--brand)}.field .i{color:var(--muted)}
.field input{flex:1;min-width:0;border:0;background:transparent;color:var(--ink);font:inherit;padding:14px 0;outline:none}
.form .button{margin-top:20px}.form > .small{margin-top:14px}
.notice{display:flex;gap:12px;align-items:flex-start;border-radius:14px;padding:14px}.notice p{margin-top:4px}.notice small{display:block;margin-top:6px;opacity:.8}
.notice.bad{background:var(--bad-bg);color:var(--bad)}.notice.good{background:var(--good-bg);color:var(--good)}
.chip.neutral{background:var(--neutral-bg);color:var(--neutral)}
.live{display:grid;gap:10px}.live a{color:var(--brand);font-weight:600;font-size:14px}
.waiting{display:flex;gap:12px;align-items:center;color:var(--muted)}.spinner.small{width:22px;height:22px;border-width:3px}
.live .verdict h2{font-size:18px}.live .verdict small{display:block;margin-top:6px;opacity:.85}.live .tiles{margin-top:0}
.demo summary{cursor:pointer;font-weight:700;color:var(--brand)}
.accounts{list-style:none;margin:12px 0;padding:0;display:grid;gap:8px}
.accounts a{display:flex;align-items:center;gap:12px;padding:10px 12px;border:1px solid var(--line);border-radius:14px;color:inherit;text-decoration:none;background:var(--soft)}
.accounts a:hover{border-color:var(--brand)}.accounts .avatar{background:var(--info-bg);color:var(--info)}
.accounts span:nth-child(2){flex:1;display:grid}.accounts small{color:var(--muted)}
.pill{font-size:13px;font-weight:700;color:var(--brand)}
.hello .date{font-size:13px;color:var(--muted);font-weight:600;text-transform:uppercase;letter-spacing:.06em}
.hello h1{margin:2px 0 14px}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:10px}
.stats div{background:var(--card);border:1px solid var(--line);border-radius:14px;padding:12px;display:grid;box-shadow:var(--shadow)}
.stats b{font-size:24px;line-height:1.1}.stats span{font-size:12px;color:var(--muted)}.stats .hot b{color:var(--bad)}
.search{display:flex;align-items:center;gap:10px;background:var(--card);border:1.5px solid var(--line);border-radius:14px;padding:0 14px}
.search:focus-within{border-color:var(--brand)}.search .i{color:var(--muted)}
.search input{flex:1;border:0;background:transparent;color:var(--ink);font:inherit;padding:13px 0;outline:none}
.chips{display:flex;gap:8px;overflow-x:auto;scrollbar-width:none}
.chips button{flex:none;border:1.5px solid var(--line);background:var(--card);color:var(--ink);border-radius:999px;padding:7px 14px;font:inherit;font-size:14px;font-weight:600;cursor:pointer}
.chips button[aria-pressed=true]{background:var(--brand);border-color:var(--brand);color:var(--on-brand)}
.jobs{list-style:none;margin:0;padding:0;display:grid;gap:12px}
.job{display:grid;gap:6px;color:inherit;text-decoration:none;position:relative;padding-left:20px;overflow:hidden}
.job::before{content:"";position:absolute;left:0;top:0;bottom:0;width:5px;background:var(--neutral)}
.tone-bad.job::before{background:var(--bad)}.tone-warn.job::before{background:#e8a33d}.tone-info.job::before{background:var(--brand-2)}
.job:hover{border-color:var(--brand)}
.job h2{margin-top:2px}.fault{font-size:15px}.fault.big{font-size:16px;margin-top:6px}
.meta{color:var(--muted);font-size:13px;gap:14px;margin-top:4px}.meta span{display:inline-flex;align-items:center;gap:5px}.meta .i{width:16px;height:16px}
.due.hot{color:var(--bad);font-weight:700}.go-arrow{margin-left:auto;color:var(--muted)}
.empty{text-align:center;color:var(--muted);padding:24px}
.hero{display:grid;gap:4px}.hero h1{margin-top:6px}
.action{display:grid;gap:12px;border:1.5px solid color-mix(in srgb,var(--go) 30%,var(--line))}
.pp{display:flex;gap:12px;align-items:flex-start}.pp .i{width:28px;height:28px;color:var(--go);margin-top:2px}.pp p{margin-top:2px}
dl{display:grid;grid-template-columns:auto 1fr;gap:8px 16px;margin:0}dt{color:var(--muted);font-size:14px}dd{margin:0;overflow-wrap:anywhere}
.card dl + .button{margin-top:14px}
.timeline{list-style:none;margin:0;padding:0 0 0 6px}
.timeline li{position:relative;padding:0 0 14px 20px;border-left:2px solid var(--line)}.timeline li:last-child{padding-bottom:0;border-color:transparent}
.timeline li::before{content:"";position:absolute;left:-7px;top:3px;width:12px;height:12px;border-radius:50%;background:var(--card);border:2px solid var(--muted)}
.timeline li.done::before{background:var(--go);border-color:var(--go)}
.timeline b{display:block;font-size:15px}.timeline small{color:var(--muted)}
.close summary{cursor:pointer;font-weight:700;color:var(--brand)}
fieldset{border:0;margin:12px 0;padding:0;display:grid;gap:8px}legend{font-size:14px;color:var(--muted);margin-bottom:6px}
.option{display:flex;gap:10px;align-items:center;margin:0;padding:12px;border:1.5px solid var(--line);border-radius:12px;font-weight:500}
.option:has(input:checked){border-color:var(--brand);background:var(--info-bg)}
.opening{position:fixed;inset:0;z-index:20;display:grid;place-content:center;justify-items:center;gap:8px;text-align:center;padding:24px;
background:color-mix(in srgb,var(--bg) 88%,transparent);backdrop-filter:blur(6px);-webkit-backdrop-filter:blur(6px)}
.opening[hidden]{display:none}.opening small{color:var(--muted)}
.opening .slow-note{display:none}.opening.slow .slow-note{display:grid;gap:12px;justify-items:center;max-width:320px;margin-top:10px}
.spinner{width:44px;height:44px;border-radius:50%;border:4px solid var(--line);border-top-color:var(--go);animation:spin .9s linear infinite}
@keyframes spin{to{transform:rotate(360deg)}}
@media (prefers-reduced-motion:reduce){.spinner{animation-duration:3s}.button{transition:none}}
.crumb{color:var(--muted);font-size:14px}
.verdict{display:flex;gap:14px;align-items:flex-start;border-radius:var(--radius);padding:18px}
.verdict .icon .i{width:36px;height:36px}.verdict p{margin-top:4px}
.verdict.tone-good{background:var(--good-bg);color:var(--good)}.verdict.tone-warn{background:var(--warn-bg);color:var(--warn)}
.verdict.tone-bad{background:var(--bad-bg);color:var(--bad)}.verdict.tone-info{background:var(--info-bg);color:var(--info)}.verdict.tone-neutral{background:var(--neutral-bg);color:var(--neutral)}
.tiles{display:grid;grid-template-columns:1fr 1fr;gap:10px;margin-top:14px}
.tiles div{background:var(--soft);border:1px solid var(--line);border-radius:14px;padding:12px;display:grid}
.tiles b{font-size:26px;line-height:1.15}.tiles small{color:var(--muted)}.tiles b.unrecorded{font-size:17px;font-weight:600;color:var(--muted);padding:6px 0}
.dials{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:10px;margin-top:14px}
.dial{display:grid;justify-items:center;gap:2px;min-width:0;color:var(--dial-down)}.dial.up{color:var(--dial-up)}.dial.none{color:var(--neutral)}
.dial svg{width:100%;max-width:200px;height:auto;display:block}.dial small{font-size:13px;font-weight:700}
.dial-track{fill:none;stroke:var(--line);stroke-width:10;stroke-linecap:round}.dial-arc{fill:none;stroke:currentColor;stroke-width:10;stroke-linecap:round}
.dial-tick{stroke:var(--muted);stroke-width:1.5}.dial-needle{stroke:currentColor;stroke-width:2.5;stroke-linecap:round}.dial-hub{fill:var(--ink)}
.dial-figure{fill:var(--ink);font:700 22px ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;text-anchor:middle}.dial-unit{fill:var(--muted);font:600 11px system-ui,sans-serif;text-anchor:middle}
.sites{list-style:none;margin:0;padding:0;display:grid;gap:10px}
.sites li{border:1px solid var(--line);border-left:4px solid var(--neutral);border-radius:12px;padding:12px;display:grid;gap:8px}
.sites li.tone-good{border-left-color:var(--good)}.sites li.tone-warn{border-left-color:#e8a33d}.sites li.tone-bad{border-left-color:var(--bad)}
.figures{display:flex;gap:18px;flex-wrap:wrap}.figures span{display:grid;font-weight:700}.figures small{font-weight:500;color:var(--muted)}
.empty-state{display:grid;gap:8px;justify-items:center;text-align:center}.empty-state .i{width:36px;height:36px;color:var(--muted)}
.raw summary{cursor:pointer;color:var(--muted);font-size:14px}pre{white-space:pre-wrap;font-size:12px;overflow-wrap:anywhere;margin:10px 0 0}
footer{padding:18px 20px calc(18px + env(safe-area-inset-bottom));text-align:center;font-size:12px;color:var(--muted)}
</style></head><body><div class="app">
{{if .Engineer}}<header class="bar">{{if .Back}}<a class="iconbutton" href="{{.Back}}" aria-label="Back">{{template "i-back"}}</a>{{else}}{{template "logo"}}{{end}}
<div class="title"><strong>{{if .Back}}{{.Title}}{{else}}{{.App}}{{end}}</strong></div>
<details class="me"><summary aria-label="Account"><span class="avatar">{{initials .Engineer.Name}}</span></summary>
<div class="menu"><strong>{{.Engineer.Name}}</strong><small>{{.Engineer.Id}} · {{.Engineer.Email}}</small>
<form method="post" action="/signout"><button class="button secondary small" type="submit">Sign out</button></form></div></details></header>{{end}}
<main>{{end}}

{{define "foot"}}</main><footer>{{.App}} is a demonstration of an operator's own field app. It opens PacketPulse without a second sign-in.</footer></div><script src="{{script}}" defer></script></body></html>{{end}}

{{define "signin"}}{{template "head" .}}
<section class="welcome">{{template "logo"}}<h1>{{.App}}</h1><p>Your jobs for the day, and a one-tap line test in PacketPulse.</p></section>
<section class="card form">
{{if .Problem}}<div class="notice bad" role="alert">{{template "i-alert"}}<div>{{.Problem}}</div></div>{{end}}
<form method="post" action="/signin">
<label for="username">Employee ID or email</label>
<div class="field">{{template "i-user"}}<input id="username" name="username" value="{{.Username}}" autocomplete="username" autocapitalize="none" spellcheck="false" required></div>
<label for="password">Password</label>
<div class="field">{{template "i-lock"}}<input id="password" name="password" type="password" value="{{.Password}}" autocomplete="current-password" required></div>
<button class="button primary" type="submit">Sign in</button>
</form>
<p class="muted center small">Forgotten your password? Ask your supervisor.</p></section>
<details class="card demo" open><summary>Demo accounts</summary>
<ul class="accounts">{{range .Engineers}}<li><a href="/?as={{.Id}}"><span class="avatar">{{initials .Name}}</span><span><strong>{{.Name}}</strong><small>{{.Id}} · {{.Password}}</small></span><span class="pill">Use</span></a></li>{{end}}</ul>
<p class="muted small">Mock credentials for this demonstration only: the form is filled in with the first; tap a name to sign in as another. The email works in place of the ID.</p></details>
{{template "foot" .}}{{end}}

{{define "tickets"}}{{template "head" .}}
<section class="hello"><p class="date">{{.Today}}</p><h1>{{.Greeting}}</h1>
<div class="stats"><div><b>{{.Open}}</b><span>open jobs</span></div><div{{if .Urgent}} class="hot"{{end}}><b>{{.Urgent}}</b><span>due within the hour</span></div><div><b>{{.Closed}}</b><span>closed today</span></div></div></section>
<div class="search">{{template "i-search"}}<input id="q" type="search" placeholder="Search jobs, customers, areas" aria-label="Search jobs" autocomplete="off"></div>
<div class="chips" role="group" aria-label="Show"><button type="button" data-filter="open" aria-pressed="true">Open</button><button type="button" data-filter="p1" aria-pressed="false">P1</button><button type="button" data-filter="closed" aria-pressed="false">Closed</button><button type="button" data-filter="all" aria-pressed="false">All</button></div>
<p class="muted small">Signed in as <strong>{{.Engineer.Name}}</strong> ({{.Engineer.Id}})</p>
<ul class="jobs">{{range .Jobs}}<li data-job data-status="{{.Status}}" data-priority="{{.Priority}}" data-search="{{.Search}}"><a class="job card tone-{{.PriorityTone}}" href="/ticket?tt={{.TtNumber}}">
<div class="row"><span class="tt">{{.TtNumber}}</span>{{if .Priority}}<span class="chip {{.PriorityTone}}">{{.Priority}}</span>{{end}}<span class="chip status-{{.Status}}">{{.StatusLabel}}</span>{{if .Verdict}}<span class="chip {{.VerdictTone}}">{{.Verdict}}</span>{{end}}</div>
<h2>{{.Customer}}</h2>
<p class="muted small">{{if .Area}}{{.Area}}{{else}}{{.Address}}{{end}}{{if .Plan}} · {{.Plan}}{{end}}</p>
<p class="fault">{{.Fault}}</p>
<div class="row meta">{{if .Window}}<span>{{template "i-clock"}}{{.Window}}</span>{{end}}{{if .Due}}<span class="due{{if .Urgent}} hot{{end}}">{{template "i-alarm"}}{{.Due}}</span>{{end}}{{if .Resolution}}<span>{{template "i-check"}}{{.Resolution}}</span>{{end}}<span class="go-arrow">{{template "i-chevron"}}</span></div>
</a></li>{{end}}</ul>
<p id="none" class="empty" hidden>No jobs match. Try another search or filter.</p>
{{template "foot" .}}{{end}}

{{define "ticket"}}{{template "head" .}}
{{with .Job}}<section class="card hero">
<div class="row"><span class="tt">{{.TtNumber}}</span>{{if .Priority}}<span class="chip {{.PriorityTone}}">{{.Priority}}</span>{{end}}<span class="chip status-{{.Status}}">{{.StatusLabel}}</span></div>
<h1>{{.Customer}}</h1><p class="fault big">{{.Fault}}</p>
<div class="row meta">{{if .Window}}<span>{{template "i-clock"}}Visit {{.Window}}</span>{{end}}{{if .Due}}<span class="due{{if .Urgent}} hot{{end}}">{{template "i-alarm"}}{{.Due}}</span>{{end}}</div>
</section>{{end}}
{{if .Refusal}}<section class="notice bad" role="alert">{{template "i-alert"}}<div><strong>PacketPulse could not open the test.</strong><p>{{.Refusal.Message}}</p><small>Reference: {{.Refusal.Status}}{{if .Refusal.Code}} {{.Refusal.Code}}{{end}}</small></div></section>{{end}}
{{if .Returned}}<section class="notice good" role="status">{{template "i-check"}}<div><strong>Back from PacketPulse</strong><p>The result appears below as soon as PacketPulse has it.</p></div></section>{{end}}
{{with .Job}}
{{if or $.Returned .TestedAt}}<section class="card live" id="live-result" data-tt="{{.TtNumber}}" aria-live="polite"><h3>Line test</h3>
<div class="live-body"><div class="waiting"><div class="spinner small"></div><span>Waiting for PacketPulse's result…</span></div></div>
<a href="/result?tt={{.TtNumber}}">See the full result</a></section>{{end}}
<section class="card action">
<div class="pp">{{template "i-pulse"}}<div><strong>Test this customer's line</strong><p class="muted small">PacketPulse opens signed in as you, with this ticket filled in. You press Start there.</p></div></div>
{{if ne .Status "closed"}}<form method="post" action="/launch" data-launch><input type="hidden" name="tt" value="{{.TtNumber}}"><button class="button go" type="submit">Test this line in PacketPulse</button></form>{{end}}
<a class="button secondary" href="/result?tt={{.TtNumber}}">See the result</a>
</section>
<section class="card"><h3>Customer</h3><dl><dt>Name</dt><dd>{{.Customer}}</dd><dt>Customer ID</dt><dd class="mono">{{.CustomerId}}</dd>{{if .Phone}}<dt>Phone</dt><dd>{{.Phone}}</dd>{{end}}<dt>Address</dt><dd>{{.Address}}</dd></dl>
<a class="button secondary" href="https://www.openstreetmap.org/search?query={{.Address}}" target="_blank" rel="noopener">{{template "i-pin"}}Navigate</a></section>
{{if or .Plan .Ont .Opened}}<section class="card"><h3>Line</h3><dl>{{if .Plan}}<dt>Plan</dt><dd>{{.Plan}}</dd>{{end}}{{if .Ont}}<dt>ONT serial</dt><dd class="mono">{{.Ont}}</dd>{{end}}{{if .Opened}}<dt>Reported</dt><dd>{{.Opened}}</dd>{{end}}</dl></section>{{end}}
<section class="card"><h3>Timeline</h3><ol class="timeline">
{{if .Opened}}<li class="done"><b>Reported by the customer</b><small>{{.Opened}}</small></li>{{end}}
<li class="done"><b>Assigned to you</b><small>Today</small></li>
{{if .TestedAt}}<li class="done"><b>Line test opened in PacketPulse</b><small>Today, {{.TestedAt}}</small></li>{{else}}<li><b>Test the line in PacketPulse</b><small>Next</small></li>{{end}}
{{if .ClosedAt}}<li class="done"><b>Closed: {{.Resolution}}</b><small>Today, {{.ClosedAt}}</small></li>{{end}}
</ol></section>
{{if ne .Status "closed"}}<details class="card close"><summary>Close this job</summary>
<form method="post" action="/close"><input type="hidden" name="tt" value="{{.TtNumber}}">
<fieldset><legend>How was it resolved?</legend>{{range $index, $resolution := $.Resolutions}}<label class="option"><input type="radio" name="resolution" value="{{$resolution}}" required{{if eq $index 0}} checked{{end}}>{{$resolution}}</label>{{end}}</fieldset>
<button class="button primary" type="submit">Close job</button></form></details>{{end}}
{{end}}
<div id="opening" class="opening" hidden role="status" aria-live="polite"><div class="spinner"></div><strong>Opening PacketPulse…</strong><small>Signing you in, with this ticket filled in.</small>
<div class="slow-note"><small>Taking a while? If PacketPulse opened over the app, close it when you are done - you come back to this job, where the result appears.</small><button type="button" id="opening-dismiss" class="button secondary small">Back to the job</button></div></div>
{{template "foot" .}}{{end}}

{{/* One way of the line's speed. A way with no figure says so: PacketPulse files no figure for a
way it could not measure fully, and a missing tile would read as a fault in this app. */}}
{{define "speed"}}<div><small>{{.Way}}</small>{{if .Figure}}<b>{{.Figure}}</b><small>Mbit/s</small>{{else}}<b class="unrecorded">Not recorded</b>{{end}}</div>{{end}}
{{/* One way of the line's speed as a dial, resting at its figure - subtle, as PacketPulse draws a report's. The tiles
beside it say the same figures in words, so a screen reader passes the drawing by. */}}
{{define "dial"}}<div class="dial{{if .Upload}} up{{end}}{{if not .Figure}} none{{end}}"><svg viewBox="0 0 200 150" aria-hidden="true" focusable="false"><path class="dial-track" d="{{.Track}}"/>{{if .Arc}}<path class="dial-arc" d="{{.Arc}}"/>{{end}}{{range .Ticks}}<line class="dial-tick" x1="{{.X1}}" y1="{{.Y1}}" x2="{{.X2}}" y2="{{.Y2}}"/>{{end}}<line class="dial-needle" x1="100" y1="86" x2="{{.NeedleX}}" y2="{{.NeedleY}}"/><circle class="dial-hub" cx="100" cy="86" r="6"/><text class="dial-figure" x="100" y="124">{{if .Figure}}{{.Figure}}{{else}}—{{end}}</text><text class="dial-unit" x="100" y="141">Mbit/s</text></svg><small>{{if .Upload}}↑{{else}}↓{{end}} {{.Way}}</small></div>{{end}}
{{define "result"}}{{template "head" .}}
{{with .Job}}<p class="crumb"><span class="tt">{{.TtNumber}}</span> · {{.Customer}}</p>{{end}}
{{if .Refusal}}<section class="card empty-state">{{template "i-info"}}<h2>No result to show yet</h2><p>{{.Refusal.Message}}</p><small class="muted">Reference: {{.Refusal.Status}}{{if .Refusal.Code}} {{.Refusal.Code}}{{end}}</small>
<a class="button secondary" href="/result?tt={{.Job.TtNumber}}">{{template "i-refresh"}}Try again</a></section>{{end}}
{{with .Result}}
<section class="verdict tone-{{.Tone}}"><div class="icon">{{if eq .Tone "good"}}{{template "i-check"}}{{else if eq .Tone "bad"}}{{template "i-x"}}{{else if eq .Tone "warn"}}{{template "i-alert"}}{{else}}{{template "i-info"}}{{end}}</div>
<div><h1>{{.Verdict}}</h1><p>{{.VerdictNote}}</p></div></section>
{{if or .TestedBy .When .Download .Upload}}<section class="card">{{if or .TestedBy .When}}<dl>{{if .TestedBy}}<dt>Tested by</dt><dd>{{.TestedBy}}</dd>{{end}}{{if .When}}<dt>Tested on</dt><dd>{{.When}}</dd>{{end}}</dl>{{end}}
{{if or .Download .Upload}}<div class="dials">{{template "dial" (dial "Download" .Download false)}}{{template "dial" (dial "Upload" .Upload true)}}</div><div class="tiles">{{template "speed" (speed "Download" .Download)}}{{template "speed" (speed "Upload" .Upload)}}</div>{{end}}</section>{{end}}
{{if .Sites}}<section class="card"><h3>Endpoints</h3><ul class="sites">{{range .Sites}}<li class="tone-{{.Tone}}"><div class="row"><strong>{{.Name}}</strong><span class="chip {{if .Reached}}good{{else}}bad{{end}}">{{if .Reached}}Reached{{else}}Not reached{{end}}</span></div>
<div class="figures">{{if .Loss}}<span><small>Loss</small>{{.Loss}} %</span>{{end}}{{if .Latency}}<span><small>Latency</small>{{.Latency}} ms</span>{{end}}{{if .Jitter}}<span><small>Jitter</small>{{.Jitter}} ms</span>{{end}}</div></li>{{end}}</ul></section>{{end}}
<a class="button secondary" href="/result?tt={{$.Job.TtNumber}}">{{template "i-refresh"}}Refresh</a>
<details class="card raw"><summary>For the support team: everything PacketPulse returned</summary><pre>{{.Raw}}</pre></details>
{{end}}
{{template "foot" .}}{{end}}
`
