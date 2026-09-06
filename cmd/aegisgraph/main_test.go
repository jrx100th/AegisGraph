package main

import "testing"

func TestNetworkReachabilityRequiresAllPrerequisites(t *testing.T) {
	n := Node{PublicIP:true,RouteIGW:false,Ingress:[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"0.0.0.0/0"}}}
	ok,_ := networkReachable(n)
	if ok { t.Fatal("public IP without IGW route must not be reachable") }
	n.RouteIGW=true
	ok,_=networkReachable(n)
	if !ok { t.Fatal("all supported exposure prerequisites should be reachable") }
	n.Ingress=[]Ingress{{Protocol:"tcp",FromPort:443,ToPort:443,CIDR:"0.0.0.0/0"}}
	ok,_=networkReachable(n)
	if ok { t.Fatal("wrong port must not be treated as SSH exposure") }
}

func TestExplicitDenyOverridesAllow(t *testing.T) {
	n:=Node{Policies:[]Policy{{Effect:"Allow",Action:"*",Resource:"*"},{Effect:"Deny",Action:"s3:GetObject",Resource:"arn:aws:s3:::sensitive/*"}}}
	if hasSupportedAccess(n,"s3:GetObject","arn:aws:s3:::sensitive/file") { t.Fatal("explicit supported deny must override allow") }
	if !hasSupportedAccess(n,"ec2:DescribeInstances","*") { t.Fatal("unrelated allowed action should remain allowed") }
}

func TestDemoTrapDoesNotProduceCriticalPath(t *testing.T) {
	s:=demoSnapshot("false-positive-trap")
	if len(s.Paths)!=0 { t.Fatalf("false-positive trap generated %d paths",len(s.Paths)) }
}

func TestAttackPathIsEvidenceBacked(t *testing.T) {
	s:=demoSnapshot("attack-path")
	if len(s.Paths)!=1 { t.Fatalf("expected one path, got %d",len(s.Paths)) }
	if len(s.Paths[0].Nodes)!=4 || len(s.Paths[0].Evidence)!=3 { t.Fatal("path must contain four nodes and evidence for each transition") }
}
