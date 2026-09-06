package main

import (
	"testing"
)

func TestPersistenceKeepsCurrentStateAcrossPartialScan(t *testing.T) {
	db,err:=openDB(t.TempDir()+"/lifecycle.db");if err!=nil{t.Fatal(err)}
	defer db.Close()
	server:=&Server{db:db}
	first:=demoSnapshot("attack-path")
	if _,err:=server.persistWithStatus(first,"COMPLETE");err!=nil{t.Fatal(err)}
	var before int
	if err:=db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&before);err!=nil{t.Fatal(err)}
	partial:=Snapshot{Environment:"replay",Account:first.Account,Status:"PARTIAL",Coverage:[]Coverage{{Service:"iam",State:"PARTIAL",Message:"AccessDenied"}}}
	if _,err:=server.persistWithStatus(partial,"PARTIAL");err!=nil{t.Fatal(err)}
	var after,scans int
	if err:=db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&after);err!=nil{t.Fatal(err)}
	if err:=db.QueryRow("SELECT COUNT(*) FROM scans").Scan(&scans);err!=nil{t.Fatal(err)}
	if after!=before||scans!=2{t.Fatalf("partial scan changed current state or history: nodes %d/%d scans %d",before,after,scans)}
}

func TestPersistenceRetiresFindingsOnlyOnCompleteScan(t *testing.T) {
	db,err:=openDB(t.TempDir()+"/finding-lifecycle.db");if err!=nil{t.Fatal(err)}
	defer db.Close()
	server:=&Server{db:db}
	first:=demoSnapshot("attack-path")
	if _,err:=server.persistWithStatus(first,"COMPLETE");err!=nil{t.Fatal(err)}
	var historyCurrent int
	if err:=db.QueryRow("SELECT COUNT(*) FROM finding_history WHERE resolved=0").Scan(&historyCurrent);err!=nil{t.Fatal(err)}
	if historyCurrent==0{t.Fatal("expected current finding history")}
	partial:=Snapshot{Environment:"replay",Account:first.Account,Status:"PARTIAL"}
	if _,err:=server.persistWithStatus(partial,"PARTIAL");err!=nil{t.Fatal(err)}
	if err:=db.QueryRow("SELECT COUNT(*) FROM finding_history WHERE resolved=0").Scan(&historyCurrent);err!=nil{t.Fatal(err)}
	if historyCurrent==0{t.Fatal("partial scan resolved a finding")}
	complete:=Snapshot{Environment:"replay",Account:first.Account,Status:"COMPLETE"}
	if _,err:=server.persistWithStatus(complete,"COMPLETE");err!=nil{t.Fatal(err)}
	if err:=db.QueryRow("SELECT COUNT(*) FROM finding_history WHERE resolved=1").Scan(&historyCurrent);err!=nil{t.Fatal(err)}
	if historyCurrent==0{t.Fatal("complete scan did not resolve missing finding")}
}
