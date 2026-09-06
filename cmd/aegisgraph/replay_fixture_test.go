package main

import (
	"context"
	"strings"
	"testing"
)

const attackPathFixtureJSON = `{
  "schema_version": 1,
  "source": "sanitized-example",
  "account": "111111111111",
  "regions": ["us-east-1"],
  "region_pages": {
    "us-east-1": [{
      "vpcs": [{"id":"vpc-1","name":"lab"}],
      "subnets": [{"id":"subnet-1","vpc_id":"vpc-1","name":"public","route_igw":true,"route_known":true}],
      "security_groups": [{"id":"sg-1","name":"open-ssh","vpc_id":"vpc-1","present":true,"ingress":[{"protocol":"tcp","from_port":22,"to_port":22,"cidr":"0.0.0.0/0"}]}],
      "instances": [{"id":"i-1","name":"workload","subnet_id":"subnet-1","security_group_ids":["sg-1"],"public_ip":true,"public_known":true,"role_name":"app"}]
    }]
  },
  "iam_pages": [{
    "roles": [{"name":"app","account":"111111111111","policies":[{"effect":"Allow","action":"s3:GetObject","resource":"arn:aws:s3:::sensitive/*"}]}]
  }],
  "s3_pages": [{"buckets":[{"name":"sensitive","region":"us-east-1","sensitive":true}]}],
  "rds_pages": {"us-east-1":[]},
  "lambda_pages": {"us-east-1":[]}
}`

func TestJSONReplayFixtureRunsSharedCollector(t *testing.T) {
	result,err:=RunReplayFixture(context.Background(),strings.NewReader(attackPathFixtureJSON))
	if err!=nil{t.Fatal(err)}
	if result.Status==""||len(result.Snapshot.Nodes)==0{t.Fatalf("fixture did not produce snapshot: %#v",result)}
	if !hasFindingRule(result.Snapshot,"AG-NET-001"){t.Fatalf("fixture did not produce collector-derived SSH finding: %#v",result.Snapshot.Findings)}
}

func TestJSONReplayFixtureRejectsMalformedOrAmbiguousInput(t *testing.T) {
	cases:=[]string{
		`{"schema_version":2,"source":"x","account":"1","regions":["r"]}`,
		`{"schema_version":1,"source":"x","account":"1","regions":["r","r"]}`,
		`{"schema_version":1,"source":"x","account":"1","regions":["r"],"unexpected":true}`,
		attackPathFixtureJSON+" trailing",
	}
	for _,raw:=range cases{if _,err:=LoadReplayFixture(strings.NewReader(raw));err==nil{t.Fatalf("malformed fixture accepted: %s",raw)}}
}

func TestReplayFixtureValidationDoesNotReadPaths(t *testing.T) {
	fixture:=`{"schema_version":1,"source":"x","account":"1","regions":["r"],"region_pages":{"r":[]}}`
	if _,err:=LoadReplayFixture(strings.NewReader(fixture));err!=nil{t.Fatal(err)}
}
