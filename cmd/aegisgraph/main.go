package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Node struct {
	Key string
	Type string
	Name string
	Account string
	Region string
	PublicIP bool
	RouteIGW bool
	Sensitive bool
	Policies []Policy
	Ingress []Ingress
}
type Edge struct {
	From string
	To string
	Type string
	Evidence string
}
type Policy struct {
	Effect string
	Action string
	Resource string
}
type Ingress struct {
	Protocol string
	FromPort int
	ToPort int
	CIDR string
}
type Finding struct {
	RuleID string
	Severity string
	Title string
	NodeKey string
	Rationale string
	Evidence []string
	Remediation string
	Risk int
}
type Path struct {
	ID string
	Category string
	Score int
	Nodes []string
	Evidence []string
}
type Coverage struct {
	Service string
	State string
	Message string
}
type Snapshot struct {
	Environment string
	Nodes []Node
	Edges []Edge
	Findings []Finding
	Paths []Path
	Coverage []Coverage
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "aegisgraph.db", "SQLite database path")
	staticDir := flag.String("static", "frontend/dist", "compiled frontend directory")
	flag.Parse()

	db, err := openDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	server := &Server{db: db, staticDir: *staticDir}
	log.Printf("AegisGraph listening on %s", *addr)
	if err := http.ListenAndServe(*addr, server.routes()); err != nil {
		log.Fatal(err)
	}
}

type Server struct {
	db *sql.DB
	staticDir string
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:"+filepath.Clean(path)+"?mode=rwc&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil { db.Close(); return nil, err }
	return db, nil
}

func migrate(db *sql.DB) error {
	ddl := []string{
		"CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY)",
		"CREATE TABLE IF NOT EXISTS scans(id INTEGER PRIMARY KEY AUTOINCREMENT, environment TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL)",
		"CREATE TABLE IF NOT EXISTS nodes(id INTEGER PRIMARY KEY AUTOINCREMENT, scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE, node_key TEXT NOT NULL, node_type TEXT NOT NULL, name TEXT NOT NULL, account_id TEXT, region TEXT, properties TEXT NOT NULL, UNIQUE(scan_id,node_key))",
		"CREATE TABLE IF NOT EXISTS edges(id INTEGER PRIMARY KEY AUTOINCREMENT, scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE, source_key TEXT NOT NULL, destination_key TEXT NOT NULL, edge_type TEXT NOT NULL, evidence TEXT NOT NULL, UNIQUE(scan_id,source_key,destination_key,edge_type))",
		"CREATE TABLE IF NOT EXISTS findings(id INTEGER PRIMARY KEY AUTOINCREMENT, scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE, rule_id TEXT NOT NULL, severity TEXT NOT NULL, title TEXT NOT NULL, node_key TEXT NOT NULL, rationale TEXT NOT NULL, evidence TEXT NOT NULL, remediation TEXT NOT NULL, risk INTEGER NOT NULL)",
		"CREATE TABLE IF NOT EXISTS attack_paths(id INTEGER PRIMARY KEY AUTOINCREMENT, scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE, path_key TEXT NOT NULL, category TEXT NOT NULL, score INTEGER NOT NULL, nodes TEXT NOT NULL, evidence TEXT NOT NULL, UNIQUE(scan_id,path_key))",
		"CREATE TABLE IF NOT EXISTS scan_coverage(id INTEGER PRIMARY KEY AUTOINCREMENT, scan_id INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE, service TEXT NOT NULL, state TEXT NOT NULL, message TEXT NOT NULL)",
		"CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(node_type)",
		"CREATE INDEX IF NOT EXISTS idx_nodes_key ON nodes(node_key)",
		"CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source_key)",
		"CREATE INDEX IF NOT EXISTS idx_edges_destination ON edges(destination_key)",
		"CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity)",
		"CREATE INDEX IF NOT EXISTS idx_paths_score ON attack_paths(score)",
	}
	for _, statement := range ddl {
		if _, err := db.Exec(statement); err != nil { return err }
	}
	return nil
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/stats", s.stats)
	mux.HandleFunc("/api/scans", s.scans)
	mux.HandleFunc("/api/demo/load", s.demoLoad)
	mux.HandleFunc("/api/assets", s.assets)
	mux.HandleFunc("/api/graph", s.graph)
	mux.HandleFunc("/api/findings", s.findings)
	mux.HandleFunc("/api/attack-paths", s.paths)
	mux.HandleFunc("/api/blast-radius/", s.blastRadius)
	mux.HandleFunc("/api/scans/aws", s.awsScan)
	mux.Handle("/", s.frontend())
	return withSafety(mux)
}

func withSafety(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status":"ok","runtime_ai_dependency":false})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	var scan, nodes, edges, findings, paths int
	_ = s.db.QueryRow("SELECT COUNT(*) FROM scans").Scan(&scan)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&nodes)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM edges").Scan(&edges)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM findings").Scan(&findings)
	_ = s.db.QueryRow("SELECT COUNT(*) FROM attack_paths").Scan(&paths)
	writeJSON(w, http.StatusOK, map[string]any{"scans":scan,"assets":nodes,"edges":edges,"findings":findings,"attack_paths":paths})
}

func (s *Server) scans(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT id,environment,status,created_at FROM scans ORDER BY id DESC LIMIT 50")
	if err != nil { writeError(w, 500, err); return }
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int; var env, status, created string
		if err := rows.Scan(&id,&env,&status,&created); err != nil { writeError(w,500,err); return }
		out = append(out,map[string]any{"id":id,"environment":env,"status":status,"created_at":created})
	}
	writeJSON(w,http.StatusOK,out)
}

func (s *Server) demoLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { writeError(w,405,errors.New("POST required")); return }
	env := r.URL.Query().Get("environment")
	if env == "" { env = "attack-path" }
	snapshot := demoSnapshot(env)
	id, err := s.persist(snapshot)
	if err != nil { writeError(w,500,err); return }
	writeJSON(w,http.StatusOK,map[string]any{"scan_id":id,"environment":env,"status":"COMPLETE","findings":len(snapshot.Findings),"attack_paths":len(snapshot.Paths)})
}

func (s *Server) awsScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { writeError(w,405,errors.New("POST required")); return }
	writeJSON(w,http.StatusNotImplemented,map[string]any{
		"status":"NOT_IMPLEMENTED",
		"message":"AWS collector interface is reserved but live discovery is not enabled in this build. No credentials were accepted or persisted.",
		"next":"Use demo mode. Implement and verify read-only AWS SDK collectors before enabling this endpoint.",
	})
}

func (s *Server) assets(w http.ResponseWriter, r *http.Request) {
	limit := boundedInt(r.URL.Query().Get("limit"),100,1,1000)
	rows, err := s.db.Query("SELECT node_key,node_type,name,account_id,region,properties FROM nodes ORDER BY node_type,name LIMIT ?",limit)
	if err != nil { writeError(w,500,err); return }
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var key, typ, name, account, region, properties string
		if err := rows.Scan(&key,&typ,&name,&account,&region,&properties); err != nil { writeError(w,500,err); return }
		var p map[string]any
		_ = json.Unmarshal([]byte(properties),&p)
		out = append(out,map[string]any{"key":key,"type":typ,"name":name,"account_id":account,"region":region,"properties":p})
	}
	writeJSON(w,http.StatusOK,out)
}

func (s *Server) graph(w http.ResponseWriter, r *http.Request) {
	limit := boundedInt(r.URL.Query().Get("limit"),500,1,2000)
	nodes, err := s.db.Query("SELECT node_key,node_type,name,account_id,region FROM nodes ORDER BY node_type,name LIMIT ?",limit)
	if err != nil { writeError(w,500,err); return }
	defer nodes.Close()
	outNodes := []map[string]any{}
	keys := map[string]bool{}
	for nodes.Next() {
		var key, typ, name, account, region string
		if err := nodes.Scan(&key,&typ,&name,&account,&region); err != nil { writeError(w,500,err); return }
		keys[key]=true
		outNodes=append(outNodes,map[string]any{"key":key,"type":typ,"name":name,"account_id":account,"region":region})
	}
	edges, err := s.db.Query("SELECT source_key,destination_key,edge_type,evidence FROM edges LIMIT ?",limit*3)
	if err != nil { writeError(w,500,err); return }
	defer edges.Close()
	outEdges := []map[string]any{}
	for edges.Next() {
		var from,to,typ,evidence string
		if err := edges.Scan(&from,&to,&typ,&evidence); err != nil { writeError(w,500,err); return }
		if keys[from] && keys[to] { outEdges=append(outEdges,map[string]any{"from":from,"to":to,"type":typ,"evidence":evidence}) }
	}
	writeJSON(w,http.StatusOK,map[string]any{"nodes":outNodes,"edges":outEdges})
}

func (s *Server) findings(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT rule_id,severity,title,node_key,rationale,evidence,remediation,risk FROM findings ORDER BY risk DESC,id")
	if err != nil { writeError(w,500,err); return }
	defer rows.Close()
	out:=[]map[string]any{}
	for rows.Next() {
		var rule,severity,title,node,rationale,evidence,remediation string; var risk int
		if err:=rows.Scan(&rule,&severity,&title,&node,&rationale,&evidence,&remediation,&risk); err!=nil { writeError(w,500,err); return }
		var ev []string; _=json.Unmarshal([]byte(evidence),&ev)
		out=append(out,map[string]any{"rule_id":rule,"severity":severity,"title":title,"node_key":node,"rationale":rationale,"evidence":ev,"remediation":remediation,"risk":risk})
	}
	writeJSON(w,http.StatusOK,out)
}

func (s *Server) paths(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query("SELECT path_key,category,score,nodes,evidence FROM attack_paths ORDER BY score DESC")
	if err != nil { writeError(w,500,err); return }
	defer rows.Close()
	out:=[]map[string]any{}
	for rows.Next() {
		var id,category,nodes,evidence string; var score int
		if err:=rows.Scan(&id,&category,&score,&nodes,&evidence); err!=nil { writeError(w,500,err); return }
		var ns,ev []string; _=json.Unmarshal([]byte(nodes),&ns); _=json.Unmarshal([]byte(evidence),&ev)
		out=append(out,map[string]any{"id":id,"category":category,"score":score,"nodes":ns,"evidence":ev})
	}
	writeJSON(w,http.StatusOK,out)
}

func (s *Server) blastRadius(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path,"/api/blast-radius/")
	if id=="" { writeError(w,400,errors.New("path id required")); return }
	var nodes,evidence string; var score int
	err:=s.db.QueryRow("SELECT nodes,evidence,score FROM attack_paths WHERE path_key=? LIMIT 1",id).Scan(&nodes,&evidence,&score)
	if err!=nil { writeError(w,404,errors.New("path not found")); return }
	var ns,ev []string; _=json.Unmarshal([]byte(nodes),&ns); _=json.Unmarshal([]byte(evidence),&ev)
	writeJSON(w,http.StatusOK,map[string]any{"source":ns[0],"reachable_nodes":ns[1:],"evidence":ev,"depth":len(ns)-1,"score":score,"uncertainty":"Only supported capability transitions are included."})
}

func (s *Server) frontend() http.Handler {
	fs := http.FileServer(http.Dir(s.staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		if strings.HasPrefix(r.URL.Path,"/api/") { http.NotFound(w,r); return }
		rel := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			http.ServeFile(w,r,filepath.Join(s.staticDir,"index.html")); return
		}
		path := filepath.Join(s.staticDir, rel)
		if r.URL.Path=="/" || !fileExists(path) { http.ServeFile(w,r,filepath.Join(s.staticDir,"index.html")); return }
		fs.ServeHTTP(w,r)
	})
}

func (s *Server) persist(snapshot Snapshot) (int64,error) {
	tx,err:=s.db.BeginTx(context.Background(),nil); if err!=nil{return 0,err}
	defer tx.Rollback()
	for _,table:=range []string{"findings","attack_paths","edges","nodes","scan_coverage","scans"} {
		if _,err:=tx.Exec("DELETE FROM "+table);err!=nil{return 0,err}
	}
	now:=time.Now().UTC().Format(time.RFC3339)
	res,err:=tx.Exec("INSERT INTO scans(environment,status,created_at) VALUES(?,?,?)",snapshot.Environment,"COMPLETE",now);if err!=nil{return 0,err}
	scanID,err:=res.LastInsertId();if err!=nil{return 0,err}
	for _,n:=range snapshot.Nodes {
		props,_:=json.Marshal(map[string]any{"public_ip":n.PublicIP,"route_to_internet_gateway":n.RouteIGW,"sensitive":n.Sensitive,"policies":n.Policies,"ingress":n.Ingress})
		if _,err:=tx.Exec("INSERT INTO nodes(scan_id,node_key,node_type,name,account_id,region,properties) VALUES(?,?,?,?,?,?,?)",scanID,n.Key,n.Type,n.Name,n.Account,n.Region,string(props));err!=nil{return 0,err}
	}
	for _,e:=range snapshot.Edges { if _,err:=tx.Exec("INSERT INTO edges(scan_id,source_key,destination_key,edge_type,evidence) VALUES(?,?,?,?,?)",scanID,e.From,e.To,e.Type,e.Evidence);err!=nil{return 0,err} }
	for _,f:=range snapshot.Findings { ev,_:=json.Marshal(f.Evidence); if _,err:=tx.Exec("INSERT INTO findings(scan_id,rule_id,severity,title,node_key,rationale,evidence,remediation,risk) VALUES(?,?,?,?,?,?,?,?,?)",scanID,f.RuleID,f.Severity,f.Title,f.NodeKey,f.Rationale,string(ev),f.Remediation,f.Risk);err!=nil{return 0,err} }
	for _,p:=range snapshot.Paths { ns,_:=json.Marshal(p.Nodes);ev,_:=json.Marshal(p.Evidence);if _,err:=tx.Exec("INSERT INTO attack_paths(scan_id,path_key,category,score,nodes,evidence) VALUES(?,?,?,?,?,?)",scanID,p.ID,p.Category,p.Score,string(ns),string(ev));err!=nil{return 0,err} }
	for _,c:=range snapshot.Coverage { if _,err:=tx.Exec("INSERT INTO scan_coverage(scan_id,service,state,message) VALUES(?,?,?,?)",scanID,c.Service,c.State,c.Message);err!=nil{return 0,err} }
	return scanID,tx.Commit()
}

func demoSnapshot(env string) Snapshot {
	if env!="secure" && env!="exposed" && env!="attack-path" && env!="false-positive-trap" { env="attack-path" }
	nodes:=[]Node{
		{Key:"aws:account:111111111111",Type:"AWS_ACCOUNT",Name:"Demo account",Account:"111111111111"},
		{Key:"aws:vpc:vpc-demo",Type:"VPC",Name:"demo-vpc",Account:"111111111111",Region:"ap-south-1"},
		{Key:"aws:subnet:subnet-demo",Type:"SUBNET",Name:"demo-public-subnet",Account:"111111111111",Region:"ap-south-1"},
		{Key:"aws:ec2:i-demo",Type:"EC2",Name:"demo-workload",Account:"111111111111",Region:"ap-south-1"},
		{Key:"aws:role:demo-role",Type:"IAM_ROLE",Name:"demo-workload-role",Account:"111111111111"},
		{Key:"aws:s3:demo-sensitive",Type:"S3_BUCKET",Name:"demo-sensitive-data",Account:"111111111111",Sensitive:true},
	}
	edges:=[]Edge{
		{From:"aws:account:111111111111",To:"aws:vpc:vpc-demo",Type:"CONTAINS",Evidence:"Account inventory fixture"},
		{From:"aws:vpc:vpc-demo",To:"aws:subnet:subnet-demo",Type:"CONTAINS",Evidence:"VPC/subnet association fixture"},
		{From:"aws:subnet:subnet-demo",To:"aws:ec2:i-demo",Type:"CONTAINS",Evidence:"Subnet instance association fixture"},
		{From:"aws:ec2:i-demo",To:"aws:role:demo-role",Type:"RUNS_AS",Evidence:"Instance profile attachment fixture"},
	}
	ec2:=&nodes[3]; role:=&nodes[4]
	switch env {
	case "secure":
		ec2.PublicIP=false; ec2.RouteIGW=false; ec2.Ingress=[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"10.0.0.0/8"}}
		role.Policies=[]Policy{{Effect:"Allow",Action:"s3:GetObject",Resource:"arn:aws:s3:::demo-sensitive/*"}}
	case "exposed":
		ec2.PublicIP=true; ec2.RouteIGW=true; ec2.Ingress=[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"0.0.0.0/0"}}
	case "false-positive-trap":
		ec2.PublicIP=true; ec2.RouteIGW=false; ec2.Ingress=[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"0.0.0.0/0"}}
	case "attack-path":
		ec2.PublicIP=true; ec2.RouteIGW=true; ec2.Ingress=[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"0.0.0.0/0"}}
		role.Policies=[]Policy{{Effect:"Allow",Action:"s3:GetObject",Resource:"arn:aws:s3:::demo-sensitive/*"}}
		edges=append(edges,
			Edge{From:"external:internet",To:ec2.Key,Type:"EXPOSED_TO",Evidence:"Public IPv4 + IGW route + TCP/22 from 0.0.0.0/0"},
			Edge{From:role.Key,To:"aws:s3:demo-sensitive",Type:"CAN_ACCESS",Evidence:"Supported Allow s3:GetObject on sensitive bucket"},
		)
		nodes=append([]Node{{Key:"external:internet",Type:"INTERNET",Name:"Internet"},},nodes...)
	}
	findings:=[]Finding{}
	reachable,ev:=networkReachable(*ec2)
	if reachable && hasOpenPort(*ec2,22) { findings=append(findings,Finding{"AG-NET-001","HIGH","Public SSH exposure",ec2.Key,"The workload is Internet reachable on TCP/22 under the supported model.",ev,"Remove public SSH exposure; use private access or a tightly scoped administrative path.",75}) }
	if reachable && hasOpenPort(*ec2,3389) { findings=append(findings,Finding{"AG-NET-002","HIGH","Public RDP exposure",ec2.Key,"The workload is Internet reachable on TCP/3389 under the supported model.",ev,"Remove public RDP exposure and use private administrative access.",75}) }
	if hasBroadAdmin(*role) { findings=append(findings,Finding{"AG-IAM-001","CRITICAL","Broad administrative IAM privilege",role.Key,"The role contains a supported Allow for wildcard action and resource without a stronger supported Deny.",[]string{"Role policy explicitly contains Allow * on *"},"Replace wildcard permissions with least-privilege actions and resources.",95}) }
	paths:=[]Path{}
	if reachable && hasOpenPort(*ec2,22) && hasSupportedAccess(*role,"s3:GetObject","arn:aws:s3:::demo-sensitive/*") {
		nodesPath:=[]string{"external:internet",ec2.Key,role.Key,"aws:s3:demo-sensitive"}
		evidence:=[]string{"Internet entry: public IPv4, IGW route, TCP/22 open to 0.0.0.0/0","Execution transition: EC2 runs as demo-workload-role","Capability transition: supported Allow s3:GetObject reaches the marked sensitive bucket"}
		paths=append(paths,Path{"path-internet-sensitive","Internet to sensitive resource",98,nodesPath,evidence})
		findings=append(findings,Finding{"AG-COMB-002","CRITICAL","Internet-exposed workload reaches sensitive resource",ec2.Key,"A bounded supported transition chain reaches a marked sensitive resource.",evidence,"Remove Internet exposure and reduce the role permissions; validate the path after rescanning.",98})
	}
	return Snapshot{Environment:env,Nodes:nodes,Edges:edges,Findings:findings,Paths:paths,Coverage:[]Coverage{{"demo","COMPLETE","Synthetic fixture; no AWS credentials or network required."}}}
}

func networkReachable(n Node) (bool,[]string) {
	ev:=[]string{}
	if !n.PublicIP { return false,append(ev,"No public address prerequisite") }
	ev=append(ev,"Resource has a public address")
	if !n.RouteIGW { return false,append(ev,"No verified 0.0.0.0/0 route to an Internet Gateway") }
	ev=append(ev,"Subnet route table has a verified Internet Gateway route")
	for _,in:=range n.Ingress { if in.CIDR=="0.0.0.0/0" && in.Protocol=="tcp" && (in.FromPort<=22 && in.ToPort>=22 || in.FromPort<=3389 && in.ToPort>=3389) { return true,append(ev,"Security Group permits a supported sensitive TCP port from 0.0.0.0/0") } }
	return false,append(ev,"No supported open sensitive TCP port in the Security Group")
}
func hasOpenPort(n Node,port int) bool { for _,in:=range n.Ingress { if in.Protocol=="tcp" && in.CIDR=="0.0.0.0/0" && in.FromPort<=port && in.ToPort>=port { return true } }; return false }
func hasBroadAdmin(n Node) bool { return hasSupportedAccess(n,"*","*") }
func hasSupportedAccess(n Node,action,resource string) bool {
	allowed:=false
	for _,p:=range n.Policies {
		if p.Effect=="Deny" && wildcard(p.Action,action) && wildcard(p.Resource,resource) { return false }
		if p.Effect=="Allow" && wildcard(p.Action,action) && wildcard(p.Resource,resource) { allowed=true }
	}
	return allowed
}
func wildcard(pattern,value string) bool {
	if pattern=="*" || pattern==value { return true }
	if strings.HasSuffix(pattern,"*") { return strings.HasPrefix(value,strings.TrimSuffix(pattern,"*")) }
	return false
}

func boundedInt(raw string, fallback,min,max int) int { n,err:=strconv.Atoi(raw);if err!=nil||n<min{return fallback};if n>max{return max};return n }
func fileExists(path string) bool { st,err:=os.Stat(path);return err==nil&&!st.IsDir() }
func writeJSON(w http.ResponseWriter,status int,v any) { w.Header().Set("Content-Type","application/json");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(v) }
func writeError(w http.ResponseWriter,status int,err error) { writeJSON(w,status,map[string]any{"error":err.Error()}) }
