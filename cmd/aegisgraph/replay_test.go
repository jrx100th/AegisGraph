package main

import (
	"context"
	"strings"
	"testing"
)

func hasFindingRule(s Snapshot,rule string) bool {for _,f:=range s.Findings{if f.RuleID==rule{return true}};return false}
func hasCoverageState(s Snapshot,state string) bool {for _,c:=range s.Coverage{if c.State==state{return true}};return false}

func TestReplayCatalogHasBroadScenarioCoverage(t *testing.T) {
	catalog:=replayScenarioCatalog()
	if len(catalog)<30 {t.Fatalf("scenario corpus has %d cases; need at least 30",len(catalog))}
	classes:=map[string]int{}
	for _,scenario:=range catalog{classes[scenario.Class]++}
	if classes["lifecycle"]<5 {t.Fatalf("lifecycle corpus too small: %d",classes["lifecycle"])}
}

func TestReplayScenariosEndToEndThroughCollectorAndAnalysis(t *testing.T) {
	for _,scenario:=range replayScenarioCatalog() {
		scenario:=scenario
		t.Run(scenario.Name,func(t *testing.T){
			result,err:=collectReplay(context.Background(),scenario.Client)
			if err!=nil {t.Fatalf("replay failed: %v",err)}
			if scenario.ExpectRule!=""&&!hasFindingRule(result.Snapshot,scenario.ExpectRule){t.Fatalf("expected finding %s; got %#v",scenario.ExpectRule,result.Snapshot.Findings)}
			if scenario.ExpectNoRule!=""&&hasFindingRule(result.Snapshot,scenario.ExpectNoRule){t.Fatalf("forbidden finding %s was generated",scenario.ExpectNoRule)}
			if scenario.ExpectPath&&len(result.Snapshot.Paths)==0{t.Fatal("expected evidence-backed attack path")}
			if !scenario.ExpectPath&&scenario.ExpectRule==""&&len(result.Snapshot.Paths)>0{t.Fatalf("unexpected path: %#v",result.Snapshot.Paths)}
			if scenario.ExpectPartial&&!hasCoverageState(result.Snapshot,"PARTIAL"){t.Fatalf("expected PARTIAL coverage; got %#v",result.Snapshot.Coverage)}
		})
	}
}

func TestReplayPaginationAndEmptyPages(t *testing.T) {
	scenario:=makeReplayScenario("out-of-order-pages")
	result,err:=collectReplay(context.Background(),scenario.Client)
	if err!=nil{t.Fatal(err)}
	if len(result.Snapshot.Nodes)<6{t.Fatalf("expected resources from multiple pages, got %d nodes",len(result.Snapshot.Nodes))}
	if !hasCoverageState(result.Snapshot,"COMPLETE"){t.Fatal("pagination replay should complete")}
	empty:=makeReplayScenario("empty-page")
	result,err=collectReplay(context.Background(),empty.Client)
	if err!=nil{t.Fatal(err)}
	if !hasCoverageState(result.Snapshot,"COMPLETE"){t.Fatal("empty page followed by data should remain complete")}
}

func TestReplayErrorsBecomePartialOrFailed(t *testing.T) {
	partial:=makeReplayScenario("iam-access-denied")
	result,err:=collectReplay(context.Background(),partial.Client)
	if err!=nil{t.Fatal(err)}
	if result.Status!="PARTIAL"||!hasCoverageState(result.Snapshot,"PARTIAL"){t.Fatalf("expected partial IAM scan, got %s %#v",result.Status,result.Snapshot.Coverage)}
	client:=baseSimulatedAWS();client.STSFailure=ReplayServiceError{Code:"AccessDenied",Message:"sts denied"}
	result,err=collectReplay(context.Background(),client)
	if err==nil||result.Status!="FAILED"{t.Fatalf("expected failed STS scan, got status=%s err=%v",result.Status,err)}
}

func TestReplayEvidenceIsProducedByAnalysis(t *testing.T) {
	result,err:=collectReplay(context.Background(),makeReplayScenario("internet-sensitive-s3").Client)
	if err!=nil{t.Fatal(err)}
	var path Path
	if len(result.Snapshot.Paths)!=1{t.Fatalf("expected one path, got %d",len(result.Snapshot.Paths))}
	path=result.Snapshot.Paths[0]
	if len(path.Nodes)!=4||len(path.Evidence)!=3{t.Fatalf("path evidence incomplete: %#v",path)}
	if !strings.Contains(strings.Join(path.Evidence," "),"public")&&!strings.Contains(strings.Join(path.Evidence," "),"Public"){t.Fatal("path lacks network evidence")}
}

func TestReplayPartialScanDoesNotRetireResources(t *testing.T) {
	first,err:=collectReplay(context.Background(),makeReplayScenario("public-ssh").Client);if err!=nil{t.Fatal(err)}
	ledger:=NewReplayScanLedger();ledger.Apply(first.Snapshot,first.Status)
	key:="aws:ec2:111111111111:ap-south-1:i-demo"
	if !ledger.Contains(key){t.Fatal("initial resource missing from ledger")}
	partial:=makeReplayScenario("iam-access-denied");partial.Client.RegionPages["ap-south-1"]=[]ReplayRegionPage{{}}
	next,err:=collectReplay(context.Background(),partial.Client);if err!=nil{t.Fatal(err)}
	ledger.Apply(next.Snapshot,next.Status)
	if !ledger.Contains(key){t.Fatal("partial scan incorrectly retired prior resource")}
}

func TestReplayIdentityIncludesAccountAndRegion(t *testing.T) {
	one,err:=collectReplay(context.Background(),baseSimulatedAWS());if err!=nil{t.Fatal(err)}
	twoClient:=baseSimulatedAWS();twoClient.Account="222222222222";two,err:=collectReplay(context.Background(),twoClient);if err!=nil{t.Fatal(err)}
	oneKey,twoKey:="",""
	for _,n:=range one.Snapshot.Nodes{if n.Type=="VPC"{oneKey=n.Key}}
	for _,n:=range two.Snapshot.Nodes{if n.Type=="VPC"{twoKey=n.Key}}
	if oneKey==""||twoKey==""||oneKey==twoKey{t.Fatalf("account collision: %q %q",oneKey,twoKey)}
}

func TestReplayPersistsNormalizedResults(t *testing.T) {
	db,err:=openDB(t.TempDir()+"/replay.db");if err!=nil{t.Fatal(err)};defer db.Close()
	result,err:=collectReplay(context.Background(),makeReplayScenario("internet-sensitive-s3").Client);if err!=nil{t.Fatal(err)}
	if _,err:=(&Server{db:db}).persist(result.Snapshot);err!=nil{t.Fatal(err)}
	var nodes,edges,findings,paths int
	for _,q:=range []struct{name string;dst *int}{{"nodes",&nodes},{"edges",&edges},{"findings",&findings},{"attack_paths",&paths}}{if err:=db.QueryRow("SELECT COUNT(*) FROM "+q.name).Scan(q.dst);err!=nil{t.Fatal(err)}}
	if nodes==0||edges==0||findings==0||paths==0{t.Fatalf("persistence lost replay output: nodes=%d edges=%d findings=%d paths=%d",nodes,edges,findings,paths)}
}


func TestReplaySupportedRiskSuiteHasTenPositiveCases(t *testing.T) {
	expected := map[string]string{
		"public-ssh": "AG-NET-001",
		"public-rdp": "AG-NET-002",
		"public-rds": "AG-NET-003",
		"multiple-security-groups": "AG-NET-001",
		"out-of-order-pages": "AG-NET-001",
		"wildcard-allow": "AG-IAM-001",
		"internet-privileged-role": "AG-COMB-002",
		"internet-sensitive-s3": "AG-COMB-002",
		"assume-role-chain": "AG-COMB-002",
		"role-chain-sensitive": "AG-COMB-002",
	}
	for name, rule := range expected {
		result, err := collectReplay(context.Background(), makeReplayScenario(name).Client)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !hasFindingRule(result.Snapshot, rule) {
			t.Fatalf("%s: expected supported-risk finding %s", name, rule)
		}
	}
}

func TestReplayFalsePositiveTrapSuite(t *testing.T) {
	traps := []string{
		"private-ec2",
		"public-ip-no-route",
		"igw-no-public-ip",
		"wrong-port",
		"restricted-cidr",
		"private-rds",
		"simple-deny",
		"wildcard-allow-explicit-deny",
		"action-mismatch",
		"resource-mismatch",
		"trust-negative",
		"cross-account-trust",
		"internet-denied-sensitive-s3",
	}
	for _, name := range traps {
		result, err := collectReplay(context.Background(), makeReplayScenario(name).Client)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(result.Snapshot.Paths) != 0 || hasFindingRule(result.Snapshot, "AG-COMB-002") {
			t.Fatalf("%s: unsupported compound conclusion was generated: paths=%d findings=%#v", name, len(result.Snapshot.Paths), result.Snapshot.Findings)
		}
	}
}
