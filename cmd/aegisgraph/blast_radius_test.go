package main

import "testing"

func TestBlastRadiusUsesOnlyCapabilityEdges(t *testing.T) {
	s:=Snapshot{Nodes:[]Node{{Key:"workload"},{Key:"role"},{Key:"bucket"},{Key:"vpc"},{Key:"unreachable"}},
		Edges:[]Edge{
			{From:"workload",To:"role",Type:"RUNS_AS",Evidence:"execution role"},
			{From:"role",To:"bucket",Type:"CAN_ACCESS",Evidence:"supported read"},
			{From:"workload",To:"vpc",Type:"CONTAINS",Evidence:"not a capability"},
		}}
	got:=computeBlastRadius(s,"workload",3)
	if len(got.Nodes)!=2 || got.Nodes[0]!="role" || got.Nodes[1]!="bucket" { t.Fatalf("unexpected radius: %#v",got) }
	if len(got.Evidence)!=2 { t.Fatalf("expected evidence for every transition: %#v",got.Evidence) }
}

func TestBlastRadiusIsBoundedAndCycleSafe(t *testing.T) {
	s:=Snapshot{Nodes:[]Node{{Key:"a"},{Key:"b"},{Key:"c"}},
		Edges:[]Edge{
			{From:"a",To:"b",Type:"CAN_ASSUME",Evidence:"a assumes b"},
			{From:"b",To:"c",Type:"CAN_ACCESS",Evidence:"b reaches c"},
			{From:"c",To:"a",Type:"CAN_ASSUME",Evidence:"cycle"},
		}}
	got:=computeBlastRadius(s,"a",2)
	if len(got.Nodes)!=2 { t.Fatalf("expected two bounded nodes, got %#v",got.Nodes) }
	got=computeBlastRadius(s,"a",1)
	if len(got.Nodes)!=1 || got.Nodes[0]!="b" { t.Fatalf("depth bound failed: %#v",got.Nodes) }
}

func TestBlastRadiusUnknownsMissingCapabilityTargets(t *testing.T) {
	got:=computeBlastRadius(Snapshot{Nodes:[]Node{{Key:"role"}},Edges:[]Edge{{From:"role",To:"missing",Type:"CAN_ACCESS",Evidence:"incomplete"}}},"role",2)
	if !got.Unknown || len(got.Nodes)!=0 { t.Fatalf("missing target must be uncertain, got %#v",got) }
}
