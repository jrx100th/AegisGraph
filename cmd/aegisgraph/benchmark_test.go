package main

import "testing"

func BenchmarkDemoSnapshot(b *testing.B) {
	for i:=0;i<b.N;i++ { _=demoSnapshot("attack-path") }
}

func BenchmarkNetworkReachability(b *testing.B) {
	n:=Node{PublicIP:true,RouteIGW:true,Ingress:[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"0.0.0.0/0"}}}
	b.ResetTimer()
	for i:=0;i<b.N;i++ { _,_=networkReachable(n) }
}
