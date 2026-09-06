package main

import (
	"sort"
	"strings"
)

func analyzeSnapshot(input Snapshot) Snapshot {
	out:=input
	out.Findings=nil
	out.Paths=nil
	nodes:=map[string]Node{}
	for _,n:=range out.Nodes { nodes[n.Key]=n }
	keys:=make([]string,0,len(nodes))
	for key:=range nodes { keys=append(keys,key) }
	sort.Strings(keys)

	for _,key:=range keys {
		n:=nodes[key]
		if n.Type=="EC2" || n.Type=="RDS" {
			reachable,evidence:=networkReachable(n)
			if !reachable { continue }
			if n.Type=="EC2" && hasOpenPort(n,22) { out.Findings=append(out.Findings,Finding{"AG-NET-001","HIGH","Public SSH exposure",n.Key,"The workload is Internet reachable on TCP/22 under the supported model.",evidence,"Remove public SSH exposure; use private access or a tightly scoped administrative path.",75}) }
			if n.Type=="EC2" && hasOpenPort(n,3389) { out.Findings=append(out.Findings,Finding{"AG-NET-002","HIGH","Public RDP exposure",n.Key,"The workload is Internet reachable on TCP/3389 under the supported model.",evidence,"Remove public RDP exposure and use private administrative access.",75}) }
			if n.Type=="RDS" && (hasOpenPort(n,3306) || hasOpenPort(n,5432)) { out.Findings=append(out.Findings,Finding{"AG-NET-003","HIGH","Public database exposure",n.Key,"The database is Internet reachable on a supported database port.",evidence,"Remove public database exposure and use private connectivity.",85}) }
		}
		if n.Type=="IAM_ROLE" && hasBroadAdmin(n) {
			out.Findings=append(out.Findings,Finding{"AG-IAM-001","CRITICAL","Broad administrative IAM privilege",n.Key,"The role contains a supported Allow for wildcard action and resource without a stronger supported Deny.",[]string{"Role policy explicitly contains Allow * on *"},"Replace wildcard permissions with least-privilege actions and resources.",95})
		}
	}

	for _,edge:=range out.Edges {
		if edge.Type!="RUNS_AS" { continue }
		workload,ok:=nodes[edge.From];if !ok||workload.Type!="EC2"{continue}
		role,ok:=nodes[edge.To];if !ok||role.Type!="IAM_ROLE"{continue}
		reachable,networkEvidence:=networkReachable(workload);if !reachable{continue}
		targetKeys:=make([]string,0)
		for key,target:=range nodes {
			if target.Sensitive { targetKeys=append(targetKeys,key) }
		}
		sort.Strings(targetKeys)
		for _,targetKey:=range targetKeys {
			target:=nodes[targetKey]
			if !hasSupportedAccess(role,"s3:GetObject",resourceARN(target)) { continue }
			evidence:=[]string{"Internet entry: "+strings.Join(networkEvidence,"; ")}
			evidence=append(evidence,"Execution transition: "+workload.Key+" RUNS_AS "+role.Key)
			evidence=append(evidence,"Capability transition: supported IAM Allow reaches "+target.Key)
			out.Paths=append(out.Paths,Path{"path:"+workload.Key+":"+target.Key,"Internet to sensitive resource",98,[]string{"external:internet",workload.Key,role.Key,target.Key},evidence})
			out.Findings=append(out.Findings,Finding{"AG-COMB-002","CRITICAL","Internet-exposed workload reaches sensitive resource",workload.Key,"A bounded supported transition chain reaches a marked sensitive resource.",evidence,"Remove Internet exposure and reduce the role permissions; validate the path after rescanning.",98})
		}
	}
	return deduplicateAnalysis(out)
}

func resourceARN(n Node) string {
	if n.Type=="S3_BUCKET" {
		return "arn:aws:s3:::"+n.Name+"/*"
	}
	return n.Key
}

func deduplicateAnalysis(s Snapshot) Snapshot {
	seenFinding:=map[string]bool{};findings:=make([]Finding,0,len(s.Findings))
	for _,f:=range s.Findings { key:=f.RuleID+"|"+f.NodeKey;if seenFinding[key]{continue};seenFinding[key]=true;findings=append(findings,f) }
	s.Findings=findings
	seenPath:=map[string]bool{};paths:=make([]Path,0,len(s.Paths))
	for _,p:=range s.Paths {if seenPath[p.ID]{continue};seenPath[p.ID]=true;paths=append(paths,p)}
	s.Paths=paths
	return s
}

func supportedExposurePort(n Node) (int,bool) {
	for _,port:=range []int{22,3389,3306,5432} { if hasOpenPort(n,port){return port,true} }
	return 0,false
}

