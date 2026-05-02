import { useEffect, useState, useRef, useCallback } from 'react';
import { RotateCcw, AlertTriangle, Palette, Code, Settings, Check } from 'lucide-react';


function hexToHsv(hex: string): [number, number, number] {
  const r = parseInt(hex.slice(1, 3), 16) / 255;
  const g = parseInt(hex.slice(3, 5), 16) / 255;
  const b = parseInt(hex.slice(5, 7), 16) / 255;
  const max = Math.max(r, g, b), min = Math.min(r, g, b), d = max - min;
  let h = 0;
  if (d > 0) {
    if (max === r) h = ((g - b) / d + 6) % 6;
    else if (max === g) h = (b - r) / d + 2;
    else h = (r - g) / d + 4;
    h = Math.round(h * 60);
  }
  return [h, max > 0 ? Math.round((d / max) * 100) : 0, Math.round(max * 100)];
}

function hsvToHex(h: number, s: number, brt: number): string {
  const sv = s / 100, vv = brt / 100;
  const f = (n: number) => {
    const k = (n + h / 60) % 6;
    return Math.round((vv - vv * sv * Math.max(0, Math.min(k, 4 - k, 1))) * 255)
      .toString(16).padStart(2, '0');
  };
  return `#${f(5)}${f(3)}${f(1)}`;
}

const isValidHex = (s: string) => /^#[0-9a-fA-F]{6}$/.test(s);


function HexColorPicker({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const [open, setOpen] = useState(false);
  const [h, setH] = useState(0);
  const [s, setS] = useState(100);
  const [brt, setBrt] = useState(100);
  const [inputText, setInputText] = useState(value.toUpperCase());

  const containerRef = useRef<HTMLDivElement>(null);
  const svRef = useRef<HTMLDivElement>(null);
  const hueRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isValidHex(value)) {
      const [nh, ns, nbrt] = hexToHsv(value);
      setH(nh); setS(ns); setBrt(nbrt);
    }
    setInputText(value.toUpperCase());
  }, [value]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const emit = useCallback((nh: number, ns: number, nbrt: number) => {
    onChange(hsvToHex(nh, ns, nbrt));
  }, [onChange]);

  const startSvDrag = (e: React.MouseEvent) => {
    e.preventDefault();
    const rect = svRef.current!.getBoundingClientRect();
    const move = (ex: number, ey: number) => {
      const ns = Math.round(Math.max(0, Math.min(1, (ex - rect.left) / rect.width)) * 100);
      const nbrt = Math.round(Math.max(0, Math.min(1, 1 - (ey - rect.top) / rect.height)) * 100);
      setS(ns); setBrt(nbrt);
      emit(h, ns, nbrt);
    };
    move(e.clientX, e.clientY);
    const onMove = (ev: MouseEvent) => move(ev.clientX, ev.clientY);
    const onUp = () => { document.removeEventListener('mousemove', onMove); document.removeEventListener('mouseup', onUp); };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  };

  const startHueDrag = (e: React.MouseEvent) => {
    e.preventDefault();
    const rect = hueRef.current!.getBoundingClientRect();
    const move = (ex: number) => {
      const nh = Math.round(Math.max(0, Math.min(360, ((ex - rect.left) / rect.width) * 360)));
      setH(nh);
      emit(nh, s, brt);
    };
    move(e.clientX);
    const onMove = (ev: MouseEvent) => move(ev.clientX);
    const onUp = () => { document.removeEventListener('mousemove', onMove); document.removeEventListener('mouseup', onUp); };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  };

  const hueColor = `hsl(${h}, 100%, 50%)`;

  return (
    <div className="relative shrink-0" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        className="w-10 h-10 rounded-lg cursor-pointer border-2 border-neutral-600 hover:border-neutral-400 transition-all shadow-inner hover:scale-105"
        style={{ backgroundColor: value }}
        title="Wybierz kolor"
      />
      {open && (
        <div className="absolute left-0 top-12 z-50 bg-neutral-900 border border-neutral-700 rounded-xl p-3 shadow-2xl w-52 select-none">
          <div
            ref={svRef}
            className="relative w-full h-36 rounded-lg mb-3 cursor-crosshair overflow-hidden"
            onMouseDown={startSvDrag}
          >
            <div className="absolute inset-0" style={{ background: hueColor }} />
            <div className="absolute inset-0" style={{ background: 'linear-gradient(to right, #fff, transparent)' }} />
            <div className="absolute inset-0" style={{ background: 'linear-gradient(to bottom, transparent, #000)' }} />
            <div
              className="absolute w-3.5 h-3.5 rounded-full border-2 border-white shadow-[0_0_0_1px_rgba(0,0,0,0.4)] -translate-x-1/2 -translate-y-1/2 pointer-events-none"
              style={{ left: `${s}%`, top: `${100 - brt}%` }}
            />
          </div>

          <div
            ref={hueRef}
            className="relative h-4 rounded-full mb-3 cursor-pointer overflow-hidden"
            style={{ background: 'linear-gradient(to right,#f00,#ff0,#0f0,#0ff,#00f,#f0f,#f00)' }}
            onMouseDown={startHueDrag}
          >
            <div
              className="absolute top-0 h-full w-3.5 rounded-full border-2 border-white shadow-md -translate-x-1/2 pointer-events-none"
              style={{ left: `${(h / 360) * 100}%`, background: hueColor }}
            />
          </div>

          <input
            type="text"
            value={inputText}
            onChange={e => {
              const v = e.target.value;
              setInputText(v);
              if (isValidHex(v)) onChange(v.toLowerCase());
            }}
            onBlur={() => {
              if (!isValidHex(inputText)) setInputText(value.toUpperCase());
            }}
            className="w-full bg-neutral-800 text-neutral-200 px-2 py-1.5 rounded-lg text-sm font-mono text-center uppercase outline-none focus:ring-1 ring-emerald-500/50 transition-all"
            maxLength={7}
          />
        </div>
      )}
    </div>
  );
}

function ColorField({ label, value, defaultValue, onChange, onReset }: any) {
  return (
    <div className="flex flex-col gap-1.5 mb-3">
      <label className="text-xs font-semibold text-neutral-400 uppercase tracking-wider flex justify-between items-center">
        <span>{label}</span>
        {value !== defaultValue && (
          <button onClick={onReset} className="text-neutral-500 hover:text-neutral-300 flex items-center gap-1 transition-colors text-[10px]">
            <RotateCcw size={10} /> Reset
          </button>
        )}
      </label>
      <div className="flex gap-2 items-center">
        <HexColorPicker value={value} onChange={onChange} />
        <input
          type="text"
          value={value.toUpperCase()}
          onChange={e => {
            const v = e.target.value;
            if (/^#[0-9a-fA-F]{0,6}$/.test(v)) onChange(v.toLowerCase());
          }}
          className="flex-1 bg-neutral-800 text-neutral-200 px-3 py-2.5 rounded-lg text-sm outline-none focus:ring-1 ring-emerald-500/50 uppercase font-mono transition-all"
        />
      </div>
    </div>
  );
}

const DEFAULTS = {
  sesja_winColor: '#00ff88',
  sesja_lossColor: '#ff3366',
  sesja_streakColor: '#ffffff',
  sesja_bgColor: '#141414',
  sesja_bgOpacity: '90',
  sesja_targetNick: "Kartos'",
  sesja_customCss: '',
  ranga_theme: 'default',
  ranga_customCss: '',
  rlPort: '49123'
};



function App() {
  const [wins, setWins] = useState<number | string>('-');
  const [losses, setLosses] = useState<number | string>('-');
  const [streak, setStreak] = useState<number | string>('-');
  const [status, setStatus] = useState<'connected' | 'disconnected'>('disconnected');

  const [settings, setSettings] = useState(DEFAULTS);
  const [showModal, setShowModal] = useState(false);
  const [tab, setTab] = useState<'stats' | 'sesja' | 'ranga' | 'ustawienia'>('stats');
  const [genMode, setGenMode] = useState('2s');
  const [genNick, setGenNick] = useState('');
  const [copied, setCopied] = useState(false);
  const [copiedSesja, setCopiedSesja] = useState(false);

  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let ws: WebSocket;
    let reconnectTimer: number;

    const connect = () => {
      const host = window.location.port === '5173' ? 'localhost:8080' : window.location.host;
      ws = new WebSocket(`ws://${host}/ws/dashboard`);
      wsRef.current = ws;

      ws.onopen = () => setStatus('connected');

      ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.wins !== undefined) setWins(data.wins);
        if (data.losses !== undefined) setLosses(data.losses);
        if (data.streak !== undefined) setStreak(data.streak);
        if (data.settings) setSettings(prev => ({ ...prev, ...data.settings }));
      };

      ws.onclose = () => {
        setStatus('disconnected');
        reconnectTimer = window.setTimeout(connect, 2000);
      };
    };

    connect();
    return () => {
      clearTimeout(reconnectTimer);
      if (ws) { ws.onclose = null; ws.close(); }
    };
  }, []);

  const sendAction = (action: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN)
      wsRef.current.send(JSON.stringify({ action }));
  };

  const updateSetting = (key: string, value: string) => {
    setSettings(prev => ({ ...prev, [key]: value }));
    if (wsRef.current?.readyState === WebSocket.OPEN)
      wsRef.current.send(JSON.stringify({ action: 'update_settings', settings: { [key]: value } }));
  };

  const handleCopy = () => {
    const host = window.location.port === '5173' ? 'localhost:8080' : window.location.host;
    const url = `http://${host}/ranga/${genMode}/${encodeURIComponent(genNick || 'Nick')}`;
    navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  };

  const handleCopySesja = () => {
    const host = window.location.port === '5173' ? 'localhost:8080' : window.location.host;
    const url = `http://${host}/sesja/`;
    navigator.clipboard.writeText(url);
    setCopiedSesja(true);
    setTimeout(() => setCopiedSesja(false), 1800);
  };

  const iframeUrl = window.location.port === '5173' ? 'http://localhost:8080/sesja/' : '/sesja/';
  const iframeRangaUrl = window.location.port === '5173' ? 'http://localhost:8080/ranga/' : '/ranga/';

  return (
    <div className="bg-neutral-950 text-neutral-100 min-h-screen flex font-sans">

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
          <div className="bg-neutral-900 border border-neutral-800 p-8 rounded-2xl shadow-2xl max-w-sm w-full flex flex-col items-center text-center">
            <div className="w-14 h-14 bg-red-500/15 text-red-500 rounded-full flex items-center justify-center mb-4">
              <AlertTriangle size={28} />
            </div>
            <h2 className="text-xl font-bold mb-2">Zresetować Sesję?</h2>
            <p className="text-neutral-400 mb-8 text-sm leading-relaxed">Usunie wszystkie wygrane, przegrane, streak oraz historię unikalnych meczów.</p>
            <div className="flex w-full gap-3">
              <button onClick={() => setShowModal(false)} className="flex-1 py-3 bg-neutral-800 hover:bg-neutral-700 rounded-xl font-bold transition-colors">Anuluj</button>
              <button onClick={() => { sendAction('reset'); setShowModal(false); }} className="flex-1 py-3 bg-red-600 hover:bg-red-500 rounded-xl font-bold transition-colors shadow-lg">Zresetuj</button>
            </div>
          </div>
        </div>
      )}

      <div className="w-[490px] border-r border-neutral-800 bg-neutral-900 flex flex-col h-screen shrink-0 shadow-[4px_0_30px_rgba(0,0,0,0.4)] z-10">

        <div className="relative p-7 border-b border-neutral-800">
          <h1 className="text-2xl font-black mb-4 uppercase tracking-widest flex items-center gap-2">
            RL&nbsp;
            <span className="text-emerald-500">
              Overlay
            </span>
          </h1>
          <div className={`px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-wider flex items-center gap-2.5 w-fit transition-all ${status === 'connected' ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/25' : 'bg-red-500/10 text-red-500 border border-red-500/25'}`}>
            <span className={`w-2 h-2 rounded-full ${status === 'connected' ? 'bg-emerald-500' : 'bg-red-500'}`} />
            {status === 'connected' ? 'Połączono (Live)' : 'Rozłączono...'}
          </div>
        </div>

        <div className="flex border-b border-neutral-800 bg-neutral-950/20 shrink-0">
          {([
            { id: 'stats', label: 'Statystyki', icon: null },
            { id: 'sesja', label: 'Sesja', icon: <Palette size={14} /> },
            { id: 'ranga', label: 'Ranga', icon: <Code size={14} /> },
            { id: 'ustawienia', label: 'Ust.', icon: <Settings size={14} /> },
          ] as const).map(t => (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={`flex-1 py-3.5 text-xs font-bold uppercase tracking-wider flex justify-center items-center gap-1.5 transition-all border-b-2 ${tab === t.id ? 'text-emerald-400 border-emerald-400 bg-neutral-800/40' : 'text-neutral-500 border-transparent hover:text-neutral-300 hover:bg-neutral-800/20'}`}
            >
              {t.icon}{t.label}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto p-7">

          {tab === 'stats' && (
            <div className="space-y-5">
              <div className="bg-neutral-950 rounded-2xl p-5 border border-neutral-800 grid grid-cols-3 gap-4">
                <div className="text-center space-y-1">
                  <div className="text-[10px] font-bold uppercase tracking-widest text-neutral-500">Wygrane</div>
                  <div className="text-4xl font-black text-emerald-500">{wins}</div>
                </div>
                <div className="text-center space-y-1 border-x border-neutral-800">
                  <div className="text-[10px] font-bold uppercase tracking-widest text-neutral-500">Streak</div>
                  <div className="text-4xl font-black text-neutral-300">{streak}</div>
                </div>
                <div className="text-center space-y-1">
                  <div className="text-[10px] font-bold uppercase tracking-widest text-neutral-500">Przegrane</div>
                  <div className="text-4xl font-black text-rose-500">{losses}</div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-2.5">
                <button onClick={() => sendAction('add_win')} className="group cursor-pointer bg-neutral-800 hover:bg-emerald-500/10 border border-neutral-700 hover:border-emerald-500/40 text-neutral-200 py-4 rounded-xl font-bold text-sm transition-all flex flex-col items-center gap-1">
                  <span className="text-emerald-400 text-xl leading-none font-black group-hover:scale-110 transition-transform">+1</span>
                  <span className="text-xs text-neutral-400 uppercase tracking-wider">Wygrana</span>
                </button>
                <button onClick={() => sendAction('sub_win')} className="group cursor-pointer bg-neutral-800 hover:bg-neutral-700 border border-neutral-700 text-neutral-400 py-4 rounded-xl font-bold text-sm transition-all flex flex-col items-center gap-1">
                  <span className="text-neutral-500 text-xl leading-none font-black">−1</span>
                  <span className="text-xs text-neutral-500 uppercase tracking-wider">Wygrana</span>
                </button>
                <button onClick={() => sendAction('add_loss')} className="group cursor-pointer bg-neutral-800 hover:bg-rose-500/10 border border-neutral-700 hover:border-rose-500/40 text-neutral-200 py-4 rounded-xl font-bold text-sm transition-all flex flex-col items-center gap-1">
                  <span className="text-rose-500 text-xl leading-none font-black group-hover:scale-110 transition-transform">+1</span>
                  <span className="text-xs text-neutral-400 uppercase tracking-wider">Przegrana</span>
                </button>
                <button onClick={() => sendAction('sub_loss')} className="group cursor-pointer bg-neutral-800 hover:bg-neutral-700 border border-neutral-700 text-neutral-400 py-4 rounded-xl font-bold text-sm transition-all flex flex-col items-center gap-1">
                  <span className="text-neutral-500 text-xl leading-none font-black">−1</span>
                  <span className="text-xs text-neutral-500 uppercase tracking-wider">Przegrana</span>
                </button>
              </div>

              <button
                onClick={() => setShowModal(true)}
                className="cursor-pointer w-full bg-red-950/20 hover:bg-red-900/30 text-red-500 border border-red-900/40 py-3.5 rounded-xl font-bold text-sm transition-all flex justify-center items-center gap-2 hover:border-red-500/40"
              >
                <RotateCcw size={15} /> Resetuj Sesję
              </button>
            </div>
          )}

          {tab === 'sesja' && (
            <div className="space-y-5">
              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-4">Link dla OBS (Sesja)</h3>
                <div className="flex bg-neutral-800 border border-neutral-700 rounded-xl overflow-hidden">
                  <input
                    readOnly
                    value={`http://${window.location.port === '5173' ? 'localhost:8080' : window.location.host}/sesja/`}
                    className="flex-1 bg-transparent text-emerald-400 font-mono text-xs px-3 py-3 outline-none"
                  />
                  <button
                    onClick={handleCopySesja}
                    className={`px-4 py-3 text-xs font-bold uppercase transition-all flex items-center gap-1.5 ${copiedSesja ? 'bg-emerald-600 text-white' : 'bg-neutral-700 hover:bg-neutral-600 text-neutral-300'}`}
                  >
                    {copiedSesja ? <><Check size={12} /> Skopiowano</> : 'Kopiuj'}
                  </button>
                </div>
                <p className="text-[10px] text-neutral-600 mt-2">Skopiuj ten link i wklej jako źródło przeglądarki w OBS.</p>
              </div>

              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-4">Ustawienia Śledzenia</h3>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-neutral-400 uppercase tracking-wider flex justify-between items-center">
                    <span>Nick do śledzenia</span>
                    {settings.sesja_targetNick !== DEFAULTS.sesja_targetNick && (
                      <button onClick={() => updateSetting('sesja_targetNick', DEFAULTS.sesja_targetNick)} className="text-neutral-500 hover:text-neutral-300 flex items-center gap-1 transition-colors text-[10px]">
                        <RotateCcw size={10} /> Reset
                      </button>
                    )}
                  </label>
                  <input
                    type="text"
                    value={settings.sesja_targetNick}
                    onChange={e => updateSetting('sesja_targetNick', e.target.value)}
                    className="w-full bg-neutral-800 border border-neutral-700 text-neutral-200 px-3 py-2.5 rounded-xl text-sm outline-none focus:border-emerald-500/50 transition-all font-mono"
                    placeholder="Np. Kartos'"
                  />
                  <p className="text-[10px] text-neutral-600 mt-1">Wpisz swój nick z gry (identyczny jak w danych z RL).</p>
                </div>
              </div>

              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-4">Kolory Tekstu</h3>
                <ColorField label="Wygrane" value={settings.sesja_winColor} defaultValue={DEFAULTS.sesja_winColor} onChange={(v: string) => updateSetting('sesja_winColor', v)} onReset={() => updateSetting('sesja_winColor', DEFAULTS.sesja_winColor)} />
                <ColorField label="Przegrane" value={settings.sesja_lossColor} defaultValue={DEFAULTS.sesja_lossColor} onChange={(v: string) => updateSetting('sesja_lossColor', v)} onReset={() => updateSetting('sesja_lossColor', DEFAULTS.sesja_lossColor)} />
                <ColorField label="Streak" value={settings.sesja_streakColor} defaultValue={DEFAULTS.sesja_streakColor} onChange={(v: string) => updateSetting('sesja_streakColor', v)} onReset={() => updateSetting('sesja_streakColor', DEFAULTS.sesja_streakColor)} />
              </div>

              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-4">Tło Nakładki</h3>
                <ColorField label="Kolor Tła" value={settings.sesja_bgColor} defaultValue={DEFAULTS.sesja_bgColor} onChange={(v: string) => updateSetting('sesja_bgColor', v)} onReset={() => updateSetting('sesja_bgColor', DEFAULTS.sesja_bgColor)} />
                <div className="flex flex-col gap-2 mt-1">
                  <label className="text-xs font-semibold text-neutral-400 uppercase tracking-wider flex justify-between items-center">
                    <span>Krycie: <span className="text-neutral-200 font-mono">{settings.sesja_bgOpacity}%</span></span>
                    {settings.sesja_bgOpacity !== DEFAULTS.sesja_bgOpacity && (
                      <button onClick={() => updateSetting('sesja_bgOpacity', DEFAULTS.sesja_bgOpacity)} className="text-neutral-500 hover:text-neutral-300 flex items-center gap-1 transition-colors text-[10px]">
                        <RotateCcw size={10} /> Reset
                      </button>
                    )}
                  </label>
                  <input type="range" min="0" max="100" value={settings.sesja_bgOpacity} onChange={e => updateSetting('sesja_bgOpacity', e.target.value)} className="w-full h-1.5 bg-neutral-800 rounded-full appearance-none cursor-pointer accent-emerald-500" />
                </div>
              </div>

              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-4 flex items-center gap-2">
                  <Code size={13} /> Custom CSS
                </h3>
                <textarea
                  value={settings.sesja_customCss}
                  onChange={e => updateSetting('sesja_customCss', e.target.value)}
                  className="w-full h-28 bg-neutral-900 border border-neutral-700 rounded-xl p-3 text-sm font-mono text-neutral-300 outline-none focus:border-emerald-500/50 transition-colors resize-none"
                  placeholder="/* Twój CSS tutaj */"
                  spellCheck={false}
                />
              </div>
            </div>
          )}

          {tab === 'ranga' && (
            <div className="space-y-5 flex flex-col h-full">

              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-5">Generator Linku (Ranga)</h3>
                <div className="space-y-4">
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-semibold text-neutral-400 uppercase tracking-wider">Nick (Epic Games)</label>
                    <input type="text" value={genNick} onChange={e => setGenNick(e.target.value)} className="w-full bg-neutral-800 border border-neutral-700 text-neutral-200 px-3 py-2.5 rounded-xl text-sm outline-none focus:border-emerald-500/50 transition-all" placeholder="TwójNick" />
                  </div>
                  <div className="flex flex-col gap-1.5">
                    <label className="text-xs font-semibold text-neutral-400 uppercase tracking-wider">Tryb (Playlist)</label>
                    <select value={genMode} onChange={e => setGenMode(e.target.value)} className="w-full bg-neutral-800 border border-neutral-700 text-neutral-200 px-3 py-2.5 rounded-xl text-sm outline-none focus:border-emerald-500/50 transition-all cursor-pointer">
                      <option value="1s">1v1</option>
                      <option value="2s">2v2</option>
                      <option value="3s">3v3</option>
                    </select>
                  </div>
                  <div>
                    <label className="text-xs font-semibold text-neutral-400 uppercase tracking-wider mb-2 block">Link dla OBS:</label>
                    <div className="flex bg-neutral-800 border border-neutral-700 rounded-xl overflow-hidden">
                      <input readOnly value={`http://localhost:8080/ranga/${genMode}/${encodeURIComponent(genNick || 'Nick')}`} className="flex-1 bg-transparent text-emerald-400 font-mono text-xs px-3 py-3 outline-none" />
                      <button
                        onClick={handleCopy}
                        className={`px-4 py-3 text-xs font-bold uppercase transition-all flex items-center gap-1.5 ${copied ? 'bg-emerald-600 text-white' : 'bg-neutral-700 hover:bg-neutral-600 text-neutral-300'}`}
                      >
                        {copied ? <><Check size={12} /> Skopiowano</> : 'Kopiuj'}
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl flex-1 flex flex-col min-h-[280px]">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-3 flex items-center gap-2">
                  <Code size={13} /> Custom CSS
                </h3>
                <p className="text-[11px] text-neutral-500 mb-4 leading-relaxed">
                  Wstrzykiwany na żywo do nakładki wybranego motywu.
                </p>
                <div className="flex justify-between items-center bg-neutral-900 px-4 py-2 rounded-t-xl border border-neutral-700 border-b-0">
                  <span className="text-xs font-mono text-neutral-400">custom.css</span>
                  {settings.ranga_customCss !== DEFAULTS.ranga_customCss && (
                    <button onClick={() => updateSetting('ranga_customCss', DEFAULTS.ranga_customCss)} className="text-[10px] text-rose-500 hover:text-rose-400 flex items-center gap-1 font-bold">
                      Wyczyść
                    </button>
                  )}
                </div>
                <textarea
                  value={settings.ranga_customCss}
                  onChange={e => updateSetting('ranga_customCss', e.target.value)}
                  className="flex-1 w-full bg-neutral-900/60 border border-neutral-700 rounded-b-xl p-3 text-sm font-mono text-neutral-300 outline-none focus:border-emerald-500/50 transition-colors resize-none"
                  placeholder="/* Twój CSS tutaj */"
                  spellCheck={false}
                />
              </div>
            </div>
          )}

          {tab === 'ustawienia' && (
            <div className="space-y-5 flex flex-col h-full">

              <div className="bg-neutral-950 p-5 border border-neutral-800 rounded-2xl">
                <h3 className="text-[10px] uppercase tracking-widest text-emerald-500 font-bold mb-4">Ustawienia Systemowe</h3>
                <div className="flex flex-col gap-1.5">
                  <label className="text-xs font-semibold text-neutral-400 uppercase tracking-wider flex justify-between">
                    Port nasłuchu z Rocket League
                    <span className="text-neutral-600 font-mono text-[10px]">Domyślnie: 49123</span>
                  </label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={settings.rlPort || ''}
                      onChange={e => updateSetting('rlPort', e.target.value)}
                      className="w-full bg-neutral-800 border border-neutral-700 text-neutral-200 px-3 py-2.5 rounded-xl text-sm outline-none focus:border-emerald-500/50 transition-all font-mono"
                      placeholder="49123"
                    />
                    {settings.rlPort !== DEFAULTS.rlPort && (
                      <button onClick={() => updateSetting('rlPort', DEFAULTS.rlPort)} className="px-3 bg-neutral-800 hover:bg-neutral-700 rounded-xl text-xs font-bold uppercase transition-colors text-neutral-400 border border-neutral-700 shrink-0">
                        Reset
                      </button>
                    )}
                  </div>
                  <p className="text-[10px] text-neutral-600 mt-1">Zmiana aktywna przy następnej próbie połączenia (~5 sek).</p>
                </div>
              </div>

              <div className="flex-1" />

              <div className="text-center py-5 border-t border-neutral-800">
                <a href="https://github.com/kartosowski/rankrloverlay" target="_blank" rel="noopener noreferrer" className="inline-block hover:opacity-70 transition-opacity">
                  <p className="text-neutral-600 text-[10px] uppercase tracking-widest font-bold">
                    Projekt przez <span className="text-emerald-500">Kartos</span>
                  </p>
                </a>
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="flex-1 bg-neutral-950 flex flex-col overflow-hidden">
        <div className="w-full h-full p-8 flex flex-col gap-8 items-center justify-center overflow-y-auto">

          <div className="w-full max-w-4xl h-52 border border-dashed border-neutral-700/60 rounded-2xl relative flex items-center justify-center group hover:border-emerald-500/40 transition-colors">
            <div className="absolute -top-3 left-4 bg-[#080809] px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-widest text-neutral-500 group-hover:text-emerald-500 transition-colors rounded">
              Nakładka Sesji
            </div>
            <iframe src={iframeUrl} className="w-full h-full border-none bg-transparent" title="Session Overlay Preview" />
          </div>

          <div className="w-full max-w-4xl h-52 border border-dashed border-neutral-700/60 rounded-2xl relative flex items-center justify-center group hover:border-emerald-500/40 transition-colors">
            <div className="absolute -top-3 left-4 bg-[#080809] px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-widest text-neutral-500 group-hover:text-emerald-500 transition-colors rounded">
              Nakładka Rangi (podgląd bez nicku)
            </div>
            <iframe src={iframeRangaUrl} className="w-full h-full border-none bg-transparent" title="Rank Overlay Preview" />
          </div>

        </div>
      </div>

    </div>
  );
}

export default App;
