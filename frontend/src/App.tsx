import { useEffect, useMemo, useState } from "react";

type Stats = { scans:number; assets:number; edges:number; findings:number; attack_paths:number };
type Finding = { rule_id:string; severity:string; title:string; node_key:string; rationale:string; evidence:string[]; remediation:string; risk:number };
type Path = { id:string; category:string; score:number; nodes:string[]; evidence:string[] };
type Graph = { nodes:{key:string;type:string;name:string;account_id?:string;region?:string}[]; edges:{from:string;to:string;type:string;evidence:string}[] };

async function get<T>(url:string):Promise<T> {
  const r = await fetch(url);
  if (!r.ok) throw new Error(await r.text());
  return r.json();
}
async function post(url:string) {
  const r = await fetch(url, {method:"POST"});
  if (!r.ok) throw new Error(await r.text());
  return r.json();
}

export default function App() {
  const [stats,setStats] = useState<Stats>({scans:0,assets:0,edges:0,findings:0,attack_paths:0});
  const [findings,setFindings] = useState<Finding[]>([]);
  const [paths,setPaths] = useState<Path[]>([]);
  const [graph,setGraph] = useState<Graph>({nodes:[],edges:[]});
  const [environment,setEnvironment] = useState("attack-path");
  const [selected,setSelected] = useState<Finding|null>(null);
  const [busy,setBusy] = useState(false);
  const [error,setError] = useState("");

  const refresh = async () => {
    try {
      const [s,f,p,g] = await Promise.all([get<Stats>("/api/stats"),get<Finding[]>("/api/findings"),get<Path[]>("/api/attack-paths"),get<Graph>("/api/graph?limit=80")]);
      setStats(s); setFindings(f); setPaths(p); setGraph(g); setError("");
    } catch (e) { setError(e instanceof Error ? e.message : "Unable to load API"); }
  };
  useEffect(() => { refresh(); }, []);

  const loadDemo = async () => {
    setBusy(true); setError("");
    try { await post("/api/demo/load?environment="+encodeURIComponent(environment)); await refresh(); }
    catch (e) { setError(e instanceof Error ? e.message : "Demo load failed"); }
    finally { setBusy(false); }
  };

  const positions = useMemo(() => {
    const width = 760, height = 300;
    return new Map(graph.nodes.map((n,i) => [n.key,{x:60+(i%5)*165,y:70+Math.floor(i/5)*95}]));
  },[graph.nodes]);

  return <div className="app-shell">
    <header className="topbar">
      <div className="brand"><span className="brand-mark">A</span><div><strong>AegisGraph</strong><small>deterministic cloud security graph</small></div></div>
      <div className="status-pill"><span className="dot"/>LOCAL / NO AI RUNTIME</div>
    </header>
    <main>
      <section className="hero">
        <div><p className="eyebrow">INVESTIGATION CONSOLE</p><h1>See the path. Verify the evidence.</h1><p className="subtitle">A small, transparent AWS-first security graph for authorized environments.</p></div>
        <div className="demo-control"><label>Demo environment<select value={environment} onChange={e=>setEnvironment(e.target.value)}><option value="attack-path">Attack path</option><option value="exposed">Exposed workload</option><option value="secure">Secure</option><option value="false-positive-trap">False-positive trap</option></select></label><button onClick={loadDemo} disabled={busy}>{busy ? "Loading…" : "Load demo"}</button></div>
      </section>
      {error && <div className="error">{error}</div>}
      <section className="metric-grid">
        <Metric label="Assets" value={stats.assets} detail="normalized nodes"/>
        <Metric label="Findings" value={stats.findings} detail="deterministic rules"/>
        <Metric label="Attack paths" value={stats.attack_paths} detail="bounded transitions"/>
        <Metric label="Graph edges" value={stats.edges} detail="evidence-backed"/>
      </section>
      <section className="content-grid">
        <div className="panel graph-panel"><div className="panel-head"><div><p className="eyebrow">SECURITY GRAPH</p><h2>Current environment</h2></div><span className="muted">{graph.nodes.length} visible nodes</span></div><div className="graph-wrap"><svg viewBox="0 0 760 300" role="img" aria-label="Security graph">
          {graph.edges.map((e,i)=>{const a=positions.get(e.from),b=positions.get(e.to);return a&&b?<line key={i} x1={a.x} y1={a.y} x2={b.x} y2={b.y} className="edge"/>:null})}
          {graph.nodes.map(n=>{const p=positions.get(n.key)!;const internet=n.type==="INTERNET";return <g key={n.key}><circle cx={p.x} cy={p.y} r="22" className={"node "+(internet?"internet":"")}/><text x={p.x} y={p.y+4} textAnchor="middle" className="node-icon">{internet?"↗":n.type==="EC2"?"▣":n.type==="IAM_ROLE"?"◈":n.type==="S3_BUCKET"?"◆":"•"}</text><text x={p.x} y={p.y+42} textAnchor="middle" className="node-label">{n.name.slice(0,19)}</text></g>})}
        </svg></div><div className="legend"><span><i className="legend-dot critical"/>critical path</span><span><i className="legend-dot normal"/>asset</span><span><i className="legend-line"/>typed relationship</span></div></div>
        <div className="panel path-panel"><div className="panel-head"><div><p className="eyebrow">PRIORITIZED PATHS</p><h2>Why it matters</h2></div></div>{paths.length===0?<Empty text="No supported attack path in this environment."/>:paths.map(p=><button className="path-card" key={p.id} onClick={()=>setSelected(findings.find(f=>f.rule_id==="AG-COMB-002")||null)}><div className="path-score">{p.score}</div><div><strong>{p.category}</strong><p>{p.nodes.map(n=>n.split(":").pop()).join("  →  ")}</p><small>{p.evidence[0]}</small></div></button>)}</div>
      </section>
      <section className="panel findings-panel"><div className="panel-head"><div><p className="eyebrow">FINDINGS</p><h2>Evidence-backed issues</h2></div><span className="muted">{findings.length} current</span></div>{findings.length===0?<Empty text="No findings were generated."/>:<div className="finding-list">{findings.map(f=><button className="finding-row" key={f.rule_id+f.node_key} onClick={()=>setSelected(f)}><span className={"severity "+f.severity.toLowerCase()}>{f.severity}</span><span className="finding-main"><strong>{f.title}</strong><small>{f.rule_id} · {f.node_key}</small></span><span className="risk">{f.risk}</span><span className="chevron">›</span></button>)}</div>}</section>
      {selected && <aside className="drawer"><button className="close" onClick={()=>setSelected(null)}>×</button><p className="eyebrow">FINDING DETAIL</p><span className={"severity "+selected.severity.toLowerCase()}>{selected.severity}</span><h2>{selected.title}</h2><p>{selected.rationale}</p><h3>Evidence</h3><ul>{selected.evidence.map((e,i)=><li key={i}>{e}</li>)}</ul><h3>Remediation</h3><p>{selected.remediation}</p><div className="score-box"><span>Transparent risk score</span><strong>{selected.risk}/100</strong></div></aside>}
    </main>
    <footer><span>AegisGraph v0.1 · support states are explicit</span><span>Secure by evidence, not inference</span></footer>
  </div>;
}
function Metric({label,value,detail}:{label:string;value:number;detail:string}) { return <div className="metric"><span>{label}</span><strong>{value}</strong><small>{detail}</small></div>; }
function Empty({text}:{text:string}) { return <div className="empty">{text}</div>; }
