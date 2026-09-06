package main

import "sort"

// BlastRadiusResult is the bounded, evidence-backed capability set reachable from a
// compromised node. Containment and presentation-only edges are intentionally excluded.
type BlastRadiusResult struct {
	Source string
	Nodes []string
	Evidence []string
	Unknown bool
}

func computeBlastRadius(snapshot Snapshot, source string, maxDepth int) BlastRadiusResult {
	if maxDepth < 1 { maxDepth = 1 }
	result:=BlastRadiusResult{Source:source}
	known:=map[string]bool{}
	for _,n:=range snapshot.Nodes { known[n.Key]=true }
	if !known[source] { result.Unknown=true; return result }
	type step struct{ key string; depth int }
	queue:=[]step{{source,0}}
	visited:=map[string]bool{source:true}
	for len(queue)>0 {
		current:=queue[0]; queue=queue[1:]
		if current.depth>=maxDepth { continue }
		edges:=make([]Edge,0)
		for _,edge:=range snapshot.Edges {
			if edge.From!=current.key { continue }
			switch edge.Type {
			case "RUNS_AS","CAN_ACCESS","CAN_ASSUME":
				edges=append(edges,edge)
			}
		}
		sort.Slice(edges,func(i,j int)bool {
			if edges[i].To!=edges[j].To { return edges[i].To<edges[j].To }
			if edges[i].Type!=edges[j].Type { return edges[i].Type<edges[j].Type }
			return edges[i].Evidence<edges[j].Evidence
		})
		for _,edge:=range edges {
			if !known[edge.To] { result.Unknown=true; continue }
			if visited[edge.To] { continue }
			visited[edge.To]=true
			result.Nodes=append(result.Nodes,edge.To)
			result.Evidence=append(result.Evidence,edge.Type+": "+edge.Evidence)
			queue=append(queue,step{edge.To,current.depth+1})
		}
	}
	return result
}
