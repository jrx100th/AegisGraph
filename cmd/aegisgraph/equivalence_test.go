package main

import (
	"context"
	"reflect"
	"testing"
)

type mirrorReplayClient struct{ inner *SimulatedAWS }

func (m mirrorReplayClient) GetCallerIdentity(ctx context.Context)(string,error){return m.inner.GetCallerIdentity(ctx)}
func (m mirrorReplayClient) ListRegions(ctx context.Context)([]string,error){return m.inner.ListRegions(ctx)}
func (m mirrorReplayClient) DescribeRegion(ctx context.Context,region,token string)(ReplayRegionPage,error){return m.inner.DescribeRegion(ctx,region,token)}
func (m mirrorReplayClient) ListIAM(ctx context.Context,token string)(ReplayIAMPage,error){return m.inner.ListIAM(ctx,token)}
func (m mirrorReplayClient) ListS3(ctx context.Context,token string)(ReplayS3Page,error){return m.inner.ListS3(ctx,token)}
func (m mirrorReplayClient) ListRDS(ctx context.Context,region,token string)(ReplayRDSPage,error){return m.inner.ListRDS(ctx,region,token)}
func (m mirrorReplayClient) ListLambda(ctx context.Context,region,token string)(ReplayLambdaPage,error){return m.inner.ListLambda(ctx,region,token)}

func TestLiveReplayAdapterContractProducesEquivalentSemantics(t *testing.T) {
	a:=baseSimulatedAWS()
	b:=baseSimulatedAWS()
	left,err:=collectReplay(context.Background(),a);if err!=nil{t.Fatal(err)}
	right,err:=collectReplay(context.Background(),mirrorReplayClient{inner:b});if err!=nil{t.Fatal(err)}
	if !reflect.DeepEqual(left.Snapshot,right.Snapshot){t.Fatalf("equivalent adapters diverged:\nleft=%#v\nright=%#v",left.Snapshot,right.Snapshot)}
}

func TestEquivalentAccessDeniedProducesEquivalentPartialCoverage(t *testing.T) {
	a:=baseSimulatedAWS();a.IAMPages=[]ReplayIAMPage{{ErrorCode:"AccessDenied"}}
	b:=baseSimulatedAWS();b.IAMPages=[]ReplayIAMPage{{ErrorCode:"AccessDenied"}}
	left,err:=collectReplay(context.Background(),a);if err!=nil{t.Fatal(err)}
	right,err:=collectReplay(context.Background(),mirrorReplayClient{inner:b});if err!=nil{t.Fatal(err)}
	if left.Status!="PARTIAL"||right.Status!="PARTIAL"{t.Fatalf("AccessDenied must be PARTIAL: %s/%s",left.Status,right.Status)}
	if !reflect.DeepEqual(left.Snapshot.Coverage,right.Snapshot.Coverage){t.Fatalf("coverage diverged: %#v/%#v",left.Snapshot.Coverage,right.Snapshot.Coverage)}
}

func TestIAMPolicyDecodingPreservesUncertainty(t *testing.T) {
	encoded:=`{"Statement":{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}}`
	policies,unsupported,err:=parsePolicyDocument(encoded)
	if err!=nil||unsupported||len(policies)!=1{t.Fatalf("plain policy failed: %#v %v %v",policies,unsupported,err)}
	encoded="%7B%22Statement%22%3A%7B%22Effect%22%3A%22Allow%22%2C%22Action%22%3A%22%2A%22%2C%22Resource%22%3A%22%2A%22%7D%7D"
	policies,unsupported,err=parsePolicyDocument(encoded)
	if err!=nil||unsupported||len(policies)!=1{t.Fatalf("URL encoded policy failed: %#v %v %v",policies,unsupported,err)}
	_,unsupported,err=parsePolicyDocument(`{"Statement":{"Effect":"Allow","Action":"*","Resource":"*","Condition":{"StringEquals":{"aws:PrincipalOrgID":"o"}}}}`)
	if err!=nil||!unsupported{t.Fatalf("unsupported condition must remain uncertain: %#v %v",unsupported,err)}
}

func TestCrossRegionIdentityIsolation(t *testing.T) {
	c:=baseSimulatedAWS()
	c.Regions=[]string{"ap-south-1","us-east-1"}
	c.RegionPages["us-east-1"]=c.RegionPages["ap-south-1"]
	result,err:=collectReplay(context.Background(),c);if err!=nil{t.Fatal(err)}
	keys:=map[string]bool{}
	for _,n:=range result.Snapshot.Nodes{if n.Type=="VPC"{keys[n.Key]=true}}
	if len(keys)!=2{t.Fatalf("same native VPC ID across regions collided: %#v",keys)}
}
