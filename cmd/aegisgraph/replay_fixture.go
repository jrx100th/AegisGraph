package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const maxReplayFixtureBytes int64 = 8 << 20

type ReplayFixture struct {
	SchemaVersion int `json:"schema_version"`
	Source string `json:"source"`
	Account string `json:"account"`
	Regions []string `json:"regions"`
	RegionPages map[string][]ReplayRegionPage `json:"region_pages"`
	IAMPages []ReplayIAMPage `json:"iam_pages"`
	S3Pages []ReplayS3Page `json:"s3_pages"`
	RDSPages map[string][]ReplayRDSPage `json:"rds_pages"`
	LambdaPages map[string][]ReplayLambdaPage `json:"lambda_pages"`
}

func (f ReplayFixture) Validate() error {
	if f.SchemaVersion != 1 { return fmt.Errorf("unsupported replay fixture schema_version %d", f.SchemaVersion) }
	if strings.TrimSpace(f.Source)=="" { return errors.New("source is required") }
	if strings.TrimSpace(f.Account)=="" { return errors.New("account is required") }
	if len(f.Regions)==0 { return errors.New("at least one region is required") }
	seen:=map[string]bool{}
	for _,region:=range f.Regions {
		if strings.TrimSpace(region)=="" { return errors.New("region cannot be empty") }
		if seen[region] { return fmt.Errorf("duplicate region %q",region) }
		seen[region]=true
	}
	return nil
}

func (f ReplayFixture) Client() (*SimulatedAWS,error) {
	if err:=f.Validate();err!=nil{return nil,err}
	return &SimulatedAWS{
		Account:f.Account,
		Regions:append([]string{},f.Regions...),
		RegionPages:f.RegionPages,
		IAMPages:f.IAMPages,
		S3Pages:f.S3Pages,
		RDSPages:f.RDSPages,
		LambdaPages:f.LambdaPages,
		RDSCursor:map[string]int{},
		LambdaCursor:map[string]int{},
	},nil
}

func LoadReplayFixture(r io.Reader) (*SimulatedAWS,error) {
	if r==nil{return nil,errors.New("fixture reader is nil")}
	dec:=json.NewDecoder(io.LimitReader(r,maxReplayFixtureBytes))
	dec.DisallowUnknownFields()
	var fixture ReplayFixture
	if err:=dec.Decode(&fixture);err!=nil{return nil,fmt.Errorf("decode replay fixture: %w",err)}
	var extra any
	if err:=dec.Decode(&extra);err!=io.EOF{return nil,errors.New("replay fixture contains trailing JSON")}
	return fixture.Client()
}

func RunReplayFixture(ctx context.Context,r io.Reader) (ReplayResult,error) {
	client,err:=LoadReplayFixture(r)
	if err!=nil{return ReplayResult{Status:"FAILED"},err}
	return collectReplay(ctx,client)
}
