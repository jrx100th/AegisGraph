package main

import (
	"fmt"
	"testing"
)

func benchmarkScaleSnapshot() Snapshot {
	nodes:=make([]Node,10000)
	for i:=range nodes { nodes[i]=Node{Key:fmt.Sprintf("node-%05d",i),Type:"ASSET"} }
	edges:=make([]Edge,0,50000)
	for i:=0;i<50000;i++ {
		from:=1000+(i%8000)
		to:=1000+((i*37+11)%8000)
		edges=append(edges,Edge{From:fmt.Sprintf("node-%05d",from),To:fmt.Sprintf("node-%05d",to),Type:"CAN_ACCESS",Evidence:"synthetic scale edge"})
	}
	for i:=0;i<6;i++ { edges=append(edges,Edge{From:fmt.Sprintf("node-%05d",i),To:fmt.Sprintf("node-%05d",i+1),Type:"CAN_ACCESS",Evidence:"reachable scale edge"}) }
	return Snapshot{Nodes:nodes,Edges:edges}
}

func BenchmarkBlastRadius10kNodes50kEdges(b *testing.B) {
	s:=benchmarkScaleSnapshot()
	b.ReportMetric(float64(len(s.Nodes)),"nodes")
	b.ReportMetric(float64(len(s.Edges)),"edges")
	b.ResetTimer()
	for i:=0;i<b.N;i++ { _=computeBlastRadius(s,"node-00000",6) }
}
