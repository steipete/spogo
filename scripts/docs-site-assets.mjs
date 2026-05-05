export function css() {
  return `
:root{
  --ink:#0f1115;
  --text:#1f2328;
  --muted:#6b7280;
  --subtle:#9aa1ab;
  --bg:#fafafa;
  --paper:#ffffff;
  --paper-2:#f3f4f6;
  --accent:#169c46;
  --accent-soft:rgba(30,215,96,.12);
  --accent-strong:#0f7c34;
  --brand:#1ed760;
  --brand-deep:#0f7c34;
  --brand-glow:rgba(30,215,96,.22);
  --accent-warm:#ff7a3d;
  --accent-pink:#ff6ea1;
  --accent-violet:#8a6dff;
  --accent-blue:#4ea0ff;
  --accent-yellow:#ffd95e;
  --line:#e5e7eb;
  --line-soft:#eef0f3;
  --code-bg:#0c1014;
  --code-fg:#e6f7ec;
  --shadow-card:0 1px 2px rgba(15,17,21,.04),0 8px 24px rgba(15,17,21,.05);
  --shadow-glow:0 0 0 1px rgba(30,215,96,.18),0 18px 60px rgba(30,215,96,.12);
}
[data-theme="dark"]{
  --ink:#f5f8f6;
  --text:#d6dedc;
  --muted:#8c9994;
  --subtle:#5e6c68;
  --bg:#0a0e12;
  --paper:#11171c;
  --paper-2:#161e23;
  --accent:#1ed760;
  --accent-soft:rgba(30,215,96,.18);
  --accent-strong:#4ee07c;
  --brand:#1ed760;
  --brand-deep:#13a248;
  --brand-glow:rgba(30,215,96,.32);
  --line:#1f2a30;
  --line-soft:#161e23;
  --code-bg:#06090c;
  --code-fg:#dfeee6;
  --shadow-card:0 1px 2px rgba(0,0,0,.4),0 12px 36px rgba(0,0,0,.45);
  --shadow-glow:0 0 0 1px rgba(30,215,96,.28),0 22px 80px rgba(30,215,96,.22);
}
*{box-sizing:border-box}
html{scroll-behavior:smooth;scroll-padding-top:24px}
html,body{background:var(--bg);color:var(--text)}
body{margin:0;font-family:"Inter",ui-sans-serif,system-ui,-apple-system,Segoe UI,sans-serif;line-height:1.65;overflow-x:hidden;-webkit-font-smoothing:antialiased;font-feature-settings:"cv02","cv03","cv04","cv11";transition:background .25s ease,color .25s ease}
::selection{background:var(--brand);color:#07120b}
a{color:var(--accent);text-decoration:none;transition:color .12s}
a:hover{text-decoration:underline;text-underline-offset:.2em}
[data-theme="dark"] a{color:var(--accent-strong)}
.shell{display:grid;grid-template-columns:268px minmax(0,1fr);min-height:100vh}
.sidebar{position:sticky;top:0;height:100vh;overflow:auto;padding:22px 22px 22px;background:var(--paper);border-right:1px solid var(--line);scrollbar-width:thin;scrollbar-color:var(--line) transparent;transition:background .25s ease,border-color .25s ease}
.sidebar::-webkit-scrollbar{width:6px}
.sidebar::-webkit-scrollbar-thumb{background:var(--line);border-radius:6px}
.sidebar-top{display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:22px}
.brand{display:flex;align-items:center;gap:11px;color:var(--ink);text-decoration:none;flex:1;min-width:0}
.brand:hover{text-decoration:none}
.brand .mark{display:flex;align-items:flex-end;justify-content:center;gap:3px;flex:0 0 28px;height:28px;padding:0}
.brand .mark i{display:block;width:5px;border-radius:2px;background:var(--brand);transform-origin:bottom;animation:eq 1.4s ease-in-out infinite}
.brand .mark i:nth-child(1){height:55%;background:var(--brand-deep);animation-delay:-.2s}
.brand .mark i:nth-child(2){height:100%;animation-delay:-.5s}
.brand .mark i:nth-child(3){height:75%;background:var(--accent-warm);animation-delay:-.9s}
@keyframes eq{
  0%,100%{transform:scaleY(.55)}
  20%{transform:scaleY(1)}
  40%{transform:scaleY(.4)}
  60%{transform:scaleY(.85)}
  80%{transform:scaleY(.5)}
}
@media(prefers-reduced-motion:reduce){
  .brand .mark i{animation:none}
}
.brand strong{display:block;font-size:1.05rem;line-height:1.1;font-weight:700;letter-spacing:0;color:var(--ink)}
.brand small{display:block;color:var(--muted);font-size:.74rem;margin-top:3px;font-weight:400}
.theme-toggle{display:inline-flex;align-items:center;justify-content:center;width:34px;height:34px;border-radius:9px;background:var(--paper-2);border:1px solid var(--line);color:var(--ink);cursor:pointer;padding:0;transition:background .15s,border-color .15s,color .15s,transform .15s}
.theme-toggle:hover{border-color:var(--brand);color:var(--accent-strong);transform:translateY(-1px)}
.theme-toggle svg{width:16px;height:16px;display:block}
.theme-toggle .icon-sun{display:none}
.theme-toggle .icon-moon{display:block}
[data-theme="dark"] .theme-toggle .icon-sun{display:block}
[data-theme="dark"] .theme-toggle .icon-moon{display:none}
.search{display:block;margin:0 0 22px}
.search span{display:block;color:var(--muted);font-size:.7rem;font-weight:600;text-transform:uppercase;letter-spacing:0;margin-bottom:7px}
.search input{width:100%;border:1px solid var(--line);background:var(--paper);color:var(--text);border-radius:8px;padding:9px 12px;font:inherit;font-size:.9rem;outline:none;transition:border-color .15s,box-shadow .15s,background .25s ease}
.search input:focus{border-color:var(--accent);box-shadow:0 0 0 3px var(--accent-soft)}
nav section{margin:0 0 18px}
nav h2{font-size:.68rem;color:var(--muted);text-transform:uppercase;letter-spacing:0;margin:0 0 6px;font-weight:600}
.nav-link{display:block;color:var(--text);text-decoration:none;border-radius:6px;padding:5px 10px;margin:1px 0;font-size:.9rem;line-height:1.4;transition:background .12s,color .12s}
.nav-link:hover{background:var(--line-soft);color:var(--ink);text-decoration:none}
.nav-link.active{background:var(--accent-soft);color:var(--accent-strong);font-weight:600}
main{min-width:0;padding:32px clamp(20px,4.5vw,56px) 80px;max-width:1180px;margin:0 auto;width:100%}
.hero{display:flex;align-items:flex-end;justify-content:space-between;gap:22px;border-bottom:1px solid var(--line);padding:8px 0 22px;margin-bottom:8px;flex-wrap:wrap}
.hero-text{min-width:0;flex:1 1 320px}
.eyebrow{margin:0 0 8px;color:var(--accent-strong);font-weight:700;text-transform:uppercase;letter-spacing:.06em;font-size:.7rem}
.hero h1{font-size:2.25rem;line-height:1.1;letter-spacing:-.01em;margin:0;font-weight:700;color:var(--ink)}
.hero-meta{display:flex;gap:8px;flex:0 0 auto;flex-wrap:wrap}
.repo,.edit,.btn-ghost{border:1px solid var(--line);color:var(--text);text-decoration:none;border-radius:7px;padding:6px 11px;font-weight:500;font-size:.83rem;background:var(--paper);transition:border-color .15s,color .15s,background .15s}
.repo:hover,.edit:hover,.btn-ghost:hover{border-color:var(--brand);color:var(--ink);text-decoration:none}
.edit{color:var(--muted)}
.home-hero{
  position:relative;
  padding:44px 32px 40px;
  margin:0 -8px 32px;
  border-radius:20px;
  border:1px solid var(--line);
  background-color:var(--paper);
  background-image:
    radial-gradient(ellipse 70% 110% at 100% -10%, var(--brand-glow) 0%, transparent 60%),
    radial-gradient(ellipse 60% 90% at -5% 110%, rgba(255,122,61,.18) 0%, transparent 60%),
    linear-gradient(160deg, var(--paper) 0%, var(--paper-2) 100%);
  box-shadow:var(--shadow-card);
  overflow:hidden;
  isolation:isolate;
}
[data-theme="dark"] .home-hero{
  background-image:
    radial-gradient(ellipse 75% 110% at 100% -10%, rgba(30,215,96,.28) 0%, transparent 65%),
    radial-gradient(ellipse 60% 90% at -5% 110%, rgba(255,122,61,.16) 0%, transparent 65%),
    linear-gradient(160deg, #0e151a 0%, #0a1014 100%);
  border-color:#1a262d;
}
.home-hero>*{position:relative;z-index:1}
.home-hero h1{font-size:3.4rem;line-height:1.04;letter-spacing:-.02em;margin:0 0 .35em;font-weight:800;color:var(--ink)}
.home-hero h1 .accent{background:linear-gradient(120deg,var(--brand) 0%,var(--accent-warm) 65%,var(--accent-pink) 100%);-webkit-background-clip:text;background-clip:text;-webkit-text-fill-color:transparent;color:transparent}
.home-hero .lede{font-size:1.18rem;line-height:1.55;color:var(--text);opacity:.85;margin:0 0 1.2em;max-width:60ch}
.home-cta{display:flex;flex-wrap:wrap;gap:10px;align-items:center;margin:0 0 18px}
.home-cta .btn{display:inline-flex;align-items:center;gap:7px;border-radius:999px;padding:11px 22px;font-weight:700;font-size:.92rem;text-decoration:none;transition:background .15s,border-color .15s,color .15s,transform .15s,box-shadow .15s}
.home-cta .btn-primary{background:var(--brand);color:#07120b;border:1px solid var(--brand);box-shadow:0 8px 22px var(--brand-glow)}
.home-cta .btn-primary:hover{background:#19c557;border-color:#19c557;text-decoration:none;transform:translateY(-1px);box-shadow:0 12px 28px var(--brand-glow)}
.home-cta .btn-ghost{padding:11px 22px;border-radius:999px;background:var(--paper);border:1px solid var(--line);color:var(--text)}
.home-cta .btn-ghost:hover{border-color:var(--brand);color:var(--ink);transform:translateY(-1px)}
.home-install{display:flex;align-items:center;gap:14px;background:var(--code-bg);color:var(--code-fg);border-radius:10px;padding:11px 14px;font:500 .9rem/1.2 "JetBrains Mono","SF Mono",ui-monospace,monospace;max-width:38em;border:1px solid #1f2937;box-shadow:0 6px 18px rgba(15,17,21,.08)}
.home-install .prompt{color:var(--brand);user-select:none;font-weight:700;flex:0 0 auto}
.home-install code{flex:1;min-width:0;background:transparent;border:0;color:var(--code-fg);font:inherit;padding:0;white-space:pre;overflow:hidden;text-overflow:ellipsis}
.home-install .copy{flex:0 0 auto;margin-left:auto;background:rgba(255,255,255,.08);color:var(--code-fg);border:1px solid rgba(255,255,255,.18);border-radius:7px;padding:6px 12px;font:600 .72rem/1 "Inter",sans-serif;letter-spacing:.02em;cursor:pointer;transition:background .15s,border-color .15s,transform .12s}
.home-install .copy:hover{background:rgba(255,255,255,.16);border-color:rgba(255,255,255,.28);transform:translateY(-1px)}
.home-install .copy.copied{background:var(--brand);border-color:var(--brand);color:#07120b}
.home-services{display:flex;flex-wrap:wrap;gap:7px;margin:18px 0 14px}
.home-services span{display:inline-flex;align-items:center;gap:6px;padding:5px 12px;border:1px solid var(--line);border-radius:999px;font-size:.78rem;color:var(--text);background:var(--paper);font-weight:500;transition:transform .15s,border-color .15s,color .15s}
.home-services span::before{content:"";width:7px;height:7px;border-radius:50%;background:var(--brand)}
.home-services span:nth-child(7n+1)::before{background:var(--brand)}
.home-services span:nth-child(7n+2)::before{background:var(--accent-warm)}
.home-services span:nth-child(7n+3)::before{background:var(--accent-pink)}
.home-services span:nth-child(7n+4)::before{background:var(--accent-violet)}
.home-services span:nth-child(7n+5)::before{background:var(--accent-blue)}
.home-services span:nth-child(7n+6)::before{background:var(--accent-yellow)}
.home-services span:nth-child(7n+7)::before{background:var(--brand-deep)}
.home-services span:hover{transform:translateY(-1px);border-color:var(--brand)}
.home-hero .muted{color:var(--muted);font-size:.92rem;margin:6px 0 0}
.home-hero .muted a{color:var(--accent-strong);font-weight:600}
.doc-grid{display:grid;grid-template-columns:minmax(0,1fr);gap:48px;margin-top:24px}
.doc-grid-home{margin-top:8px}
@media(min-width:1180px){.doc-grid{grid-template-columns:minmax(0,72ch) 200px;justify-content:start}.doc-grid-home{grid-template-columns:minmax(0,76ch);justify-content:start}}
.doc{min-width:0;max-width:72ch;overflow-wrap:break-word}
.doc-home{max-width:78ch}
.doc h1{font-size:2.6rem;line-height:1.08;letter-spacing:-.01em;margin:0 0 .4em;font-weight:700;color:var(--ink)}
body:not(.home) .doc>h1:first-child{display:none}
.doc h2{font-size:1.45rem;line-height:1.2;margin:2em 0 .5em;font-weight:700;letter-spacing:-.01em;color:var(--ink);position:relative;padding-left:14px}
.doc h2::before{content:"";position:absolute;left:0;top:.35em;bottom:.3em;width:4px;border-radius:3px;background:linear-gradient(180deg,var(--brand) 0%,var(--brand-deep) 100%)}
.doc h3{font-size:1.1rem;margin:1.7em 0 .35em;position:relative;font-weight:600;color:var(--ink);letter-spacing:0}
.doc h4{font-size:.98rem;margin:1.4em 0 .25em;color:var(--ink);position:relative;font-weight:600}
.doc h2:first-child,.doc h3:first-child,.doc h4:first-child{margin-top:.2em}
.doc :is(h2,h3,h4) .anchor{position:absolute;left:-1.05em;top:0;color:var(--subtle);opacity:0;text-decoration:none;font-weight:400;padding-right:.3em;transition:opacity .12s,color .12s}
.doc h2 .anchor{left:-.7em}
.doc :is(h2,h3,h4):hover .anchor{opacity:.7}
.doc :is(h2,h3,h4) .anchor:hover{opacity:1;color:var(--accent-strong);text-decoration:none}
.doc p{margin:0 0 1.05em}
.doc ul,.doc ol{padding-left:1.3rem;margin:0 0 1.15em}
.doc li{margin:.25em 0}
.doc li>p{margin:0 0 .4em}
.doc strong{font-weight:600;color:var(--ink)}
.doc em{font-style:italic}
.doc code{font-family:"JetBrains Mono","SF Mono",ui-monospace,monospace;font-size:.84em;background:var(--accent-soft);border:1px solid var(--line);border-radius:5px;padding:.08em .4em;color:var(--accent-strong)}
[data-theme="dark"] .doc code{color:var(--brand);background:rgba(30,215,96,.12)}
.doc pre{position:relative;overflow:auto;background:var(--code-bg);color:var(--code-fg);border-radius:10px;padding:16px 18px 14px;margin:1.3em 0;font-size:.85em;line-height:1.6;scrollbar-width:thin;scrollbar-color:#334155 transparent;border:1px solid #1f2937;box-shadow:0 6px 18px rgba(15,17,21,.06)}
.doc pre::before{content:"";position:absolute;top:10px;left:14px;width:9px;height:9px;border-radius:50%;background:var(--brand);box-shadow:14px 0 0 #ffd95e,28px 0 0 #ff6ea1;opacity:.85}
.doc pre>code{padding-top:14px}
.doc pre::-webkit-scrollbar{height:8px;width:8px}
.doc pre::-webkit-scrollbar-thumb{background:#334155;border-radius:8px}
.doc pre code{display:block;background:transparent;border:0;color:inherit;padding:0;font-size:1em;white-space:pre}
.doc pre .copy{position:absolute;top:10px;right:10px;background:rgba(255,255,255,.08);color:var(--code-fg);border:1px solid rgba(255,255,255,.18);border-radius:7px;padding:5px 12px;font:600 .7rem/1 "Inter",sans-serif;letter-spacing:.02em;cursor:pointer;opacity:0;transition:opacity .15s,background .15s,border-color .15s}
.doc pre:hover .copy,.doc pre .copy:focus{opacity:1}
.doc pre .copy:hover{background:rgba(255,255,255,.12)}
.doc pre .copy.copied{background:var(--brand);border-color:var(--brand);color:#07120b;opacity:1}
.doc blockquote{margin:1.4em 0;padding:12px 18px;border-left:4px solid var(--brand);background:var(--accent-soft);border-radius:0 10px 10px 0;color:var(--text)}
.doc blockquote p:last-child{margin-bottom:0}
.doc table{width:100%;border-collapse:collapse;margin:1.2em 0;font-size:.92em;background:var(--paper);border-radius:8px;overflow:hidden;border:1px solid var(--line)}
.doc th,.doc td{border-bottom:1px solid var(--line);padding:9px 12px;text-align:left;vertical-align:top}
.doc tr:last-child td{border-bottom:0}
.doc th{font-weight:700;color:var(--ink);background:var(--paper-2);border-bottom:1px solid var(--line);font-size:.86em;text-transform:uppercase;letter-spacing:.04em}
.doc hr{border:0;border-top:1px solid var(--line);margin:2.2em 0}
.toc{position:sticky;top:24px;align-self:start;font-size:.84rem;padding-left:14px;border-left:1px solid var(--line);max-height:calc(100vh - 48px);overflow:auto;scrollbar-width:thin;scrollbar-color:var(--line) transparent}
.toc::-webkit-scrollbar{width:5px}
.toc::-webkit-scrollbar-thumb{background:var(--line);border-radius:5px}
.toc h2{font-size:.66rem;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin:0 0 10px;font-weight:700}
.toc a{display:block;color:var(--muted);text-decoration:none;padding:4px 0 4px 10px;line-height:1.35;border-left:2px solid transparent;margin-left:-12px;transition:color .12s,border-color .12s}
.toc a:hover{color:var(--ink);text-decoration:none}
.toc a.active{color:var(--accent-strong);border-left-color:var(--brand);font-weight:600}
.toc-l3{padding-left:22px!important;font-size:.94em}
@media(max-width:1179px){.toc{display:none}}
.page-nav{display:grid;grid-template-columns:1fr 1fr;gap:14px;margin-top:48px;border-top:1px solid var(--line);padding-top:20px}
.page-nav>a{display:block;border:1px solid var(--line);background:var(--paper);border-radius:10px;padding:14px 16px;text-decoration:none;color:var(--text);transition:border-color .15s,transform .15s,box-shadow .15s}
.page-nav>a:hover{border-color:var(--brand);text-decoration:none;color:var(--ink);transform:translateY(-1px);box-shadow:var(--shadow-card)}
.page-nav small{display:block;color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.05em;margin-bottom:5px;font-weight:700}
.page-nav span{display:block;font-weight:600;line-height:1.3;color:var(--ink)}
.page-nav-prev{text-align:left}
.page-nav-next{text-align:right;grid-column:2}
.page-nav-prev:only-child{grid-column:1}
.nav-toggle{display:none;position:fixed;top:14px;right:14px;top:calc(14px + env(safe-area-inset-top, 0px));right:calc(14px + env(safe-area-inset-right, 0px));z-index:20;width:40px;height:40px;border-radius:10px;background:var(--paper);border:1px solid var(--line);color:var(--ink);cursor:pointer;padding:10px 9px;flex-direction:column;align-items:stretch;justify-content:space-between;box-shadow:0 4px 14px rgba(15,17,21,.12)}
.nav-toggle span{display:block;width:100%;height:2px;flex:0 0 2px;background:currentColor;border-radius:2px;transition:transform .2s,opacity .2s}
.nav-toggle[aria-expanded="true"] span:nth-child(1){transform:translateY(8px) rotate(45deg)}
.nav-toggle[aria-expanded="true"] span:nth-child(2){opacity:0}
.nav-toggle[aria-expanded="true"] span:nth-child(3){transform:translateY(-8px) rotate(-45deg)}
@media(max-width:900px){
  .shell{display:block}
  .sidebar{position:fixed;inset:0 30% 0 0;max-width:320px;height:100vh;z-index:15;transform:translateX(-100%);transition:transform .25s ease,background .25s ease,border-color .25s ease;box-shadow:0 18px 40px rgba(15,17,21,.18);background:var(--paper);pointer-events:none}
  .sidebar.open{transform:translateX(0);pointer-events:auto}
  .nav-toggle{display:flex}
  main{padding:64px 18px 56px}
  .hero{padding-top:6px}
  .hero h1{font-size:1.8rem}
  .home-hero{padding:28px 18px 26px;margin:0 0 24px;border-radius:16px}
  .home-hero h1{font-size:2.5rem}
  .doc h1{font-size:2.1rem}
  .hero-meta{width:100%;justify-content:flex-start}
  .doc{padding:0}
  .doc-grid{margin-top:18px;gap:24px}
  .doc :is(h2,h3,h4) .anchor{display:none}
  .doc h2{padding-left:12px}
}
@media(max-width:520px){
  main{padding:60px 14px 48px}
  .doc pre{margin-left:-14px;margin-right:-14px;border-radius:0;border-left:0;border-right:0}
  .home-install{flex-wrap:wrap}
}
`;
}

export function js() {
  return `
const root=document.documentElement;
const themeBtn=document.querySelector('.theme-toggle');
const storedThemeKey='spogo-theme';
function applyTheme(t,persist){
  if(t!=='light'&&t!=='dark')t='light';
  root.dataset.theme=t;
  if(themeBtn){
    const next=t==='dark'?'light':'dark';
    themeBtn.setAttribute('aria-label','Switch to '+next+' mode');
    themeBtn.setAttribute('title','Switch to '+next+' mode');
  }
  if(persist){try{localStorage.setItem(storedThemeKey,t)}catch{}}
}
function currentTheme(){
  if(root.dataset.theme==='dark'||root.dataset.theme==='light')return root.dataset.theme;
  return 'light';
}
themeBtn?.addEventListener('click',()=>{applyTheme(currentTheme()==='dark'?'light':'dark',true)});
applyTheme(currentTheme(),false);
const systemDark=window.matchMedia('(prefers-color-scheme: dark)');
const onSystemChange=(e)=>{let stored=null;try{stored=localStorage.getItem(storedThemeKey)}catch{}if(stored)return;applyTheme(e.matches?'dark':'light',false)};
systemDark.addEventListener?.('change',onSystemChange);

const sidebar=document.querySelector('.sidebar');
const toggle=document.querySelector('.nav-toggle');
const mobileNav=window.matchMedia('(max-width: 900px)');
const sidebarFocusable='a[href],button,input,select,textarea,[tabindex]';
function setSidebarFocusable(enabled){
  sidebar?.querySelectorAll(sidebarFocusable).forEach((el)=>{
    if(enabled){
      if(el.dataset.sidebarTabindex!==undefined){
        if(el.dataset.sidebarTabindex)el.setAttribute('tabindex',el.dataset.sidebarTabindex);
        else el.removeAttribute('tabindex');
        delete el.dataset.sidebarTabindex;
      }
    }else if(el.dataset.sidebarTabindex===undefined){
      el.dataset.sidebarTabindex=el.getAttribute('tabindex')??'';
      el.setAttribute('tabindex','-1');
    }
  });
}
function setSidebarOpen(open){
  if(!sidebar||!toggle)return;
  sidebar.classList.toggle('open',open);
  toggle.setAttribute('aria-expanded',open?'true':'false');
  if(mobileNav.matches){
    sidebar.inert=!open;
    if(open)sidebar.removeAttribute('aria-hidden');
    else sidebar.setAttribute('aria-hidden','true');
    setSidebarFocusable(open);
  }else{
    sidebar.inert=false;
    sidebar.removeAttribute('aria-hidden');
    setSidebarFocusable(true);
  }
}
setSidebarOpen(false);
toggle?.addEventListener('click',()=>setSidebarOpen(!sidebar?.classList.contains('open')));
document.addEventListener('click',(e)=>{if(!sidebar?.classList.contains('open'))return;if(sidebar.contains(e.target)||toggle?.contains(e.target))return;setSidebarOpen(false)});
document.addEventListener('keydown',(e)=>{if(e.key==='Escape')setSidebarOpen(false)});
const syncSidebarForViewport=()=>setSidebarOpen(sidebar?.classList.contains('open')??false);
if(mobileNav.addEventListener)mobileNav.addEventListener('change',syncSidebarForViewport);
else mobileNav.addListener?.(syncSidebarForViewport);

const input=document.getElementById('doc-search');
input?.addEventListener('input',()=>{const q=input.value.trim().toLowerCase();document.querySelectorAll('nav section').forEach(sec=>{let any=false;sec.querySelectorAll('.nav-link').forEach(a=>{const m=!q||a.textContent.toLowerCase().includes(q);a.style.display=m?'block':'none';if(m)any=true});sec.style.display=any?'block':'none'})});
function attachCopy(target,getText){const btn=document.createElement('button');btn.type='button';btn.className='copy';btn.textContent='Copy';btn.addEventListener('click',async()=>{try{await navigator.clipboard.writeText(getText());btn.textContent='Copied';btn.classList.add('copied');setTimeout(()=>{btn.textContent='Copy';btn.classList.remove('copied')},1400)}catch{btn.textContent='Failed';setTimeout(()=>{btn.textContent='Copy'},1400)}});target.appendChild(btn)}
document.querySelectorAll('.doc pre').forEach(pre=>attachCopy(pre,()=>pre.querySelector('code')?.textContent??''));
document.querySelectorAll('.home-install').forEach(el=>attachCopy(el,()=>el.querySelector('code')?.textContent??''));
const tocLinks=document.querySelectorAll('.toc a');
if(tocLinks.length){const map=new Map();tocLinks.forEach(a=>{const id=a.getAttribute('href').slice(1);const el=document.getElementById(id);if(el)map.set(el,a)});const setActive=l=>{tocLinks.forEach(x=>x.classList.remove('active'));l.classList.add('active')};const obs=new IntersectionObserver(entries=>{const visible=entries.filter(e=>e.isIntersecting).sort((a,b)=>a.boundingClientRect.top-b.boundingClientRect.top);if(visible.length){const link=map.get(visible[0].target);if(link)setActive(link)}},{rootMargin:'-15% 0px -65% 0px',threshold:0});map.forEach((_,el)=>obs.observe(el))}
`;
}

export function themeBootstrap() {
  // Inline-able tiny script that runs before render, no FOUC.
  return `(function(){try{var k='spogo-theme';var s=localStorage.getItem(k);var d=window.matchMedia&&window.matchMedia('(prefers-color-scheme: dark)').matches;var t=s==='dark'||s==='light'?s:(d?'dark':'light');document.documentElement.dataset.theme=t}catch(e){}})();`;
}

export function faviconSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" role="img" aria-label="spogo">
<defs>
  <radialGradient id="g" cx="50%" cy="50%" r="50%">
    <stop offset="0%" stop-color="#1ed760"/>
    <stop offset="65%" stop-color="#0f7c34"/>
    <stop offset="100%" stop-color="#06281a"/>
  </radialGradient>
</defs>
<rect width="64" height="64" rx="14" fill="#0c1014"/>
<circle cx="32" cy="32" r="22" fill="url(#g)"/>
<circle cx="32" cy="32" r="22" fill="none" stroke="rgba(255,255,255,.08)" stroke-width="1"/>
<circle cx="32" cy="32" r="14" fill="none" stroke="rgba(0,0,0,.35)" stroke-width="1"/>
<circle cx="32" cy="32" r="9" fill="none" stroke="rgba(0,0,0,.45)" stroke-width="1"/>
<circle cx="32" cy="32" r="4" fill="#ff7a3d"/>
<circle cx="32" cy="32" r="1.4" fill="#0c1014"/>
</svg>`;
}

export function themeToggleSvg() {
  return `<svg class="icon-moon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg><svg class="icon-sun" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>`;
}
