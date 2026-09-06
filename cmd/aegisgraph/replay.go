package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrReplayAccessDenied = errors.New("AccessDenied")
	ErrReplayThrottled = errors.New("Throttling")
	ErrReplayTransient = errors.New("TransientError")
)

type ReplayServiceError struct { Code string; Message string }
func (e ReplayServiceError) Error() string { return e.Code+": "+e.Message }

type ReplayRegionPage struct {
	VPCs []ReplayVPC
	Subnets []ReplaySubnet
	SecurityGroups []ReplaySecurityGroup
	Instances []ReplayInstance
	RDS []ReplayRDS
	Lambda []ReplayLambda
	NextToken string
	ErrorCode string
}
type ReplayVPC struct { ID string; Name string }
type ReplaySubnet struct { ID string; VPCID string; Name string; RouteIGW bool; RouteKnown bool }
type ReplaySecurityGroup struct { ID string; Name string; VPCID string; Ingress []Ingress; Present bool }
type ReplayInstance struct { ID string; Name string; SubnetID string; SecurityGroupIDs []string; PublicIP bool; IPv6 bool; RoleName string }
type ReplayRDS struct { ID string; Name string; SubnetID string; Public bool; Ingress []Ingress; Sensitive bool }
type ReplayLambda struct { ID string; Name string; RoleName string }
type ReplayIAMPage struct { Roles []ReplayRole; Policies []ReplayPolicy; NextToken string; ErrorCode string }
type ReplayRole struct { Name string; Account string; Policies []Policy; Trust []string; AttachTo []string; Malformed bool; UnsupportedCondition bool }
type ReplayPolicy struct { Name string; Account string; Document []Policy; AttachedRoleNames []string; Malformed bool; UnsupportedCondition bool }
type ReplayS3Page struct { Buckets []ReplayBucket; NextToken string; ErrorCode string }
type ReplayBucket struct { Name string; Region string; Public bool; Sensitive bool }
type ReplayLambdaPage struct { Functions []ReplayLambda; NextToken string; ErrorCode string }
type ReplayRDSPage struct { Instances []ReplayRDS; NextToken string; ErrorCode string }

type ReplayClient interface {
	GetCallerIdentity(context.Context) (string,error)
	ListRegions(context.Context) ([]string,error)
	DescribeRegion(context.Context,string,string) (ReplayRegionPage,error)
	ListIAM(context.Context,string) (ReplayIAMPage,error)
	ListS3(context.Context,string) (ReplayS3Page,error)
	ListRDS(context.Context,string,string) (ReplayRDSPage,error)
	ListLambda(context.Context,string,string) (ReplayLambdaPage,error)
}

type SimulatedAWS struct {
	Account string
	Regions []string
	RegionPages map[string][]ReplayRegionPage
	IAMPages []ReplayIAMPage
	S3Pages []ReplayS3Page
	RDSPages map[string][]ReplayRDSPage
	LambdaPages map[string][]ReplayLambdaPage
	STSFailure error
	RegionsFailure error
	IAMCursor int
	S3Cursor int
	RDSCursor map[string]int
	LambdaCursor map[string]int
}

func (s *SimulatedAWS) GetCallerIdentity(ctx context.Context) (string,error) { if err:=ctx.Err();err!=nil{return "",err};if s.STSFailure!=nil{return "",s.STSFailure};return s.Account,nil }
func (s *SimulatedAWS) ListRegions(ctx context.Context) ([]string,error) { if err:=ctx.Err();err!=nil{return nil,err};if s.RegionsFailure!=nil{return nil,s.RegionsFailure};out:=append([]string{},s.Regions...);return out,nil }
func (s *SimulatedAWS) DescribeRegion(ctx context.Context,region,token string) (ReplayRegionPage,error) {
	if err:=ctx.Err();err!=nil{return ReplayRegionPage{},err}
	pages:=s.RegionPages[region];idx:=tokenIndex(token)
	if idx>=len(pages){return ReplayRegionPage{},nil}
	p:=pages[idx];if p.ErrorCode!=""{return ReplayRegionPage{},replayError(p.ErrorCode)}
	if p.NextToken=="" && idx+1<len(pages){p.NextToken=fmt.Sprintf("page-%d",idx+1)}
	return p,nil
}
func (s *SimulatedAWS) ListIAM(ctx context.Context,token string) (ReplayIAMPage,error) { if err:=ctx.Err();err!=nil{return ReplayIAMPage{},err};idx:=tokenIndex(token);if idx>=len(s.IAMPages){return ReplayIAMPage{},nil};p:=s.IAMPages[idx];if p.ErrorCode!=""{return ReplayIAMPage{},replayError(p.ErrorCode)};if p.NextToken==""&&idx+1<len(s.IAMPages){p.NextToken=fmt.Sprintf("page-%d",idx+1)};return p,nil }
func (s *SimulatedAWS) ListS3(ctx context.Context,token string) (ReplayS3Page,error) { if err:=ctx.Err();err!=nil{return ReplayS3Page{},err};idx:=tokenIndex(token);if idx>=len(s.S3Pages){return ReplayS3Page{},nil};p:=s.S3Pages[idx];if p.ErrorCode!=""{return ReplayS3Page{},replayError(p.ErrorCode)};if p.NextToken==""&&idx+1<len(s.S3Pages){p.NextToken=fmt.Sprintf("page-%d",idx+1)};return p,nil }
func (s *SimulatedAWS) ListRDS(ctx context.Context,region,token string) (ReplayRDSPage,error) { if err:=ctx.Err();err!=nil{return ReplayRDSPage{},err};pages:=s.RDSPages[region];idx:=tokenIndex(token);if idx>=len(pages){return ReplayRDSPage{},nil};p:=pages[idx];if p.ErrorCode!=""{return ReplayRDSPage{},replayError(p.ErrorCode)};if p.NextToken==""&&idx+1<len(pages){p.NextToken=fmt.Sprintf("page-%d",idx+1)};return p,nil }
func (s *SimulatedAWS) ListLambda(ctx context.Context,region,token string) (ReplayLambdaPage,error) { if err:=ctx.Err();err!=nil{return ReplayLambdaPage{},err};pages:=s.LambdaPages[region];idx:=tokenIndex(token);if idx>=len(pages){return ReplayLambdaPage{},nil};p:=pages[idx];if p.ErrorCode!=""{return ReplayLambdaPage{},replayError(p.ErrorCode)};if p.NextToken==""&&idx+1<len(pages){p.NextToken=fmt.Sprintf("page-%d",idx+1)};return p,nil }

func tokenIndex(token string) int {
	if token=="" { return 0 }
	if strings.HasPrefix(token,"page-") { var n int;_,_ = fmt.Sscanf(token,"page-%d",&n);return n }
	return 0
}
func replayError(code string) error {
	switch code { case "AccessDenied":return ReplayServiceError{code,"simulated permission boundary"};case "Throttling":return ReplayServiceError{code,"simulated rate limit"};case "Transient":return ReplayServiceError{code,"simulated retryable failure"};default:return ReplayServiceError{code,"simulated service failure"} }
}

type ReplayResult struct { Snapshot Snapshot; Status string }

func collectReplay(ctx context.Context,client ReplayClient) (ReplayResult,error) {
	account,err:=client.GetCallerIdentity(ctx);if err!=nil{return ReplayResult{Status:"FAILED"},err}
	regions,err:=client.ListRegions(ctx);if err!=nil{return ReplayResult{Status:"FAILED"},err}
	sort.Strings(regions)
	out:=Snapshot{Environment:"replay",Coverage:[]Coverage{{"sts","COMPLETE","Replay caller identity"}}}
	out.Nodes=append(out.Nodes,Node{Key:"aws:account:"+account,Type:"AWS_ACCOUNT",Name:"Replay account "+account,Account:account})
	partial:=false
	for _,region:=range regions {
		out.Nodes=append(out.Nodes,Node{Key:"aws:region:"+account+":"+region,Type:"REGION",Name:region,Account:account,Region:region})
		token:=""
		seenPage:=map[string]bool{}
		serviceState:="COMPLETE"
		for {
			if seenPage[token]{partial=true;serviceState="PARTIAL";out.Coverage=append(out.Coverage,Coverage{"ec2:"+region,serviceState,"Pagination cycle detected"});break}
			seenPage[token]=true
			page,e:=client.DescribeRegion(ctx,region,token)
			if e!=nil {partial=true;serviceState="PARTIAL";out.Coverage=append(out.Coverage,Coverage{"ec2:"+region,serviceState,e.Error()});break}
			normalizeRegionPage(account,region,page,&out)
			if replayPageNeedsPartial(page) { partial=true; out.Coverage=append(out.Coverage,Coverage{"ec2:"+region,"PARTIAL","Missing or unsupported network metadata in replay response"}) }
			if page.NextToken==""{break};token=page.NextToken
		}
		if serviceState=="COMPLETE"{out.Coverage=append(out.Coverage,Coverage{"ec2:"+region,"COMPLETE","Replay pages consumed"})}
	}
	iamToken:="";iamAttempted:=false
	for {
		iamAttempted=true;page,e:=client.ListIAM(ctx,iamToken)
		if e!=nil {partial=true;out.Coverage=append(out.Coverage,Coverage{"iam","PARTIAL",e.Error()});break}
		for _,role:=range page.Roles {
			if role.Malformed || role.UnsupportedCondition { partial=true; out.Coverage=append(out.Coverage,Coverage{"iam","PARTIAL","Malformed or unsupported role policy semantics"}); role.Policies=nil }
			normalizeRole(account,role,&out)
		}
		for _,policy:=range page.Policies {
			if policy.Malformed || policy.UnsupportedCondition { partial=true; out.Coverage=append(out.Coverage,Coverage{"iam","PARTIAL","Malformed or unsupported managed policy semantics"}); continue }
			normalizePolicy(account,policy,&out)
		}
		if page.NextToken==""{break};iamToken=page.NextToken
	}
	if iamAttempted && !hasCoverage(out.Coverage,"iam","PARTIAL"){out.Coverage=append(out.Coverage,Coverage{"iam","COMPLETE","Replay pages consumed"})}
	s3Token:=""
	for {
		page,e:=client.ListS3(ctx,s3Token)
		if e!=nil {partial=true;out.Coverage=append(out.Coverage,Coverage{"s3","PARTIAL",e.Error()});break}
		for _,bucket:=range page.Buckets {
			if bucket.Name=="" || bucket.Region=="" { partial=true; out.Coverage=append(out.Coverage,Coverage{"s3","PARTIAL","Bucket identity or region missing"}); if bucket.Name==""{continue} }
			key:="aws:s3:"+account+":"+bucket.Name
			out.Nodes=append(out.Nodes,Node{Key:key,Type:"S3_BUCKET",Name:bucket.Name,Account:account,Region:bucket.Region,PublicIP:bucket.Public,Sensitive:bucket.Sensitive})
		}
		if page.NextToken==""{out.Coverage=append(out.Coverage,Coverage{"s3","COMPLETE","Replay pages consumed"});break};s3Token=page.NextToken
	}
	for _,region:=range regions {
		token:="";for {
			page,e:=client.ListRDS(ctx,region,token)
			if e!=nil {partial=true;out.Coverage=append(out.Coverage,Coverage{"rds:"+region,"PARTIAL",e.Error()});break}
			for _,db:=range page.Instances { normalizeRDS(account,region,db,&out) }
			if page.NextToken==""{out.Coverage=append(out.Coverage,Coverage{"rds:"+region,"COMPLETE","Replay pages consumed"});break};token=page.NextToken
		}
		token="";for {
			page,e:=client.ListLambda(ctx,region,token)
			if e!=nil {partial=true;out.Coverage=append(out.Coverage,Coverage{"lambda:"+region,"PARTIAL",e.Error()});break}
			for _,fn:=range page.Functions { normalizeLambda(account,region,fn,&out) }
			if page.NextToken==""{out.Coverage=append(out.Coverage,Coverage{"lambda:"+region,"COMPLETE","Replay pages consumed"});break};token=page.NextToken
		}
	}
	if hasPartialCoverageAny(out.Coverage) { partial=true }
	out=analyzeSnapshot(out)
	if partial {return ReplayResult{Snapshot:out,Status:"PARTIAL"},nil}
	return ReplayResult{Snapshot:out,Status:"COMPLETE"},nil
}

func replayPageNeedsPartial(page ReplayRegionPage) bool {
	for _,sub:=range page.Subnets { if !sub.RouteKnown{return true} }
	if len(page.Instances)>0 && len(page.Subnets)==0{return true}
	sgIDs:=map[string]bool{};for _,instance:=range page.Instances{for _,id:=range instance.SecurityGroupIDs{sgIDs[id]=true};if instance.IPv6{return true}}
	if len(sgIDs)>0 && len(page.SecurityGroups)==0{return true}
	return false
}
func hasPartialCoverageAny(c []Coverage) bool {for _,x:=range c{if x.State=="PARTIAL"||x.State=="FAILED"{return true}};return false}

func normalizeRegionPage(account,region string,page ReplayRegionPage,out *Snapshot) {
	vpcKeys:=map[string]string{}
	for _,v:=range page.VPCs {key:="aws:vpc:"+account+":"+region+":"+v.ID;vpcKeys[v.ID]=key;out.Nodes=append(out.Nodes,Node{Key:key,Type:"VPC",Name:v.Name,Account:account,Region:region});out.Edges=append(out.Edges,Edge{From:"aws:account:"+account,To:key,Type:"CONTAINS",Evidence:"Replay EC2 VPC response"})}
	subnetKeys:=map[string]string{}
	for _,sub:=range page.Subnets {key:="aws:subnet:"+account+":"+region+":"+sub.ID;subnetKeys[sub.ID]=key;out.Nodes=append(out.Nodes,Node{Key:key,Type:"SUBNET",Name:sub.Name,Account:account,Region:region,RouteIGW:sub.RouteIGW});if vk:=vpcKeys[sub.VPCID];vk!=""{out.Edges=append(out.Edges,Edge{From:vk,To:key,Type:"CONTAINS",Evidence:"Replay EC2 subnet response"})}}
	sgIngress:=map[string][]Ingress{}
	for _,sg:=range page.SecurityGroups {key:="aws:security-group:"+account+":"+region+":"+sg.ID;out.Nodes=append(out.Nodes,Node{Key:key,Type:"SECURITY_GROUP",Name:sg.Name,Account:account,Region:region});sgIngress[sg.ID]=append([]Ingress{},sg.Ingress...);if vk:=vpcKeys[sg.VPCID];vk!=""{out.Edges=append(out.Edges,Edge{From:vk,To:key,Type:"CONTAINS",Evidence:"Replay EC2 security-group response"})}}
	for _,instance:=range page.Instances {key:="aws:ec2:"+account+":"+region+":"+instance.ID;ingress:=[]Ingress{};for _,sg:=range instance.SecurityGroupIDs{ingress=append(ingress,sgIngress[sg]...)};out.Nodes=append(out.Nodes,Node{Key:key,Type:"EC2",Name:instance.Name,Account:account,Region:region,PublicIP:instance.PublicIP,RouteIGW:subnetRoute(page.Subnets,instance.SubnetID),Ingress:ingress});if sk:=subnetKeys[instance.SubnetID];sk!=""{out.Edges=append(out.Edges,Edge{From:sk,To:key,Type:"CONTAINS",Evidence:"Replay EC2 instance response"})};for _,sg:=range instance.SecurityGroupIDs{out.Edges=append(out.Edges,Edge{From:key,To:"aws:security-group:"+account+":"+region+":"+sg,Type:"USES_SECURITY_GROUP",Evidence:"Replay instance security-group attachment"})};if instance.RoleName!=""{out.Edges=append(out.Edges,Edge{From:key,To:"aws:role:"+account+":"+instance.RoleName,Type:"RUNS_AS",Evidence:"Replay instance profile attachment"})}}
	for _,db:=range page.RDS{normalizeRDS(account,region,db,out)}
	for _,fn:=range page.Lambda{normalizeLambda(account,region,fn,out)}
}
func subnetRoute(subnets []ReplaySubnet,id string) bool {for _,s:=range subnets{if s.ID==id{return s.RouteKnown&&s.RouteIGW}};return false}
func normalizeRole(account string,role ReplayRole,out *Snapshot) {key:="aws:role:"+account+":"+role.Name;out.Nodes=append(out.Nodes,Node{Key:key,Type:"IAM_ROLE",Name:role.Name,Account:account,Policies:append([]Policy{},role.Policies...)});for _,target:=range role.Trust{out.Edges=append(out.Edges,Edge{From:key,To:"aws:role:"+target,Type:"CAN_ASSUME",Evidence:"Replay role trust response"})};for _,target:=range role.AttachTo{out.Edges=append(out.Edges,Edge{From:"aws:ec2:"+account+":"+target,To:key,Type:"RUNS_AS",Evidence:"Replay role attachment"})}}
func normalizePolicy(account string,policy ReplayPolicy,out *Snapshot){for _,role:=range policy.AttachedRoleNames{for i:=range out.Nodes{if out.Nodes[i].Key=="aws:role:"+account+":"+role{out.Nodes[i].Policies=append(out.Nodes[i].Policies,policy.Document...)}}}}
func normalizeRDS(account,region string,db ReplayRDS,out *Snapshot){key:="aws:rds:"+account+":"+region+":"+db.ID;out.Nodes=append(out.Nodes,Node{Key:key,Type:"RDS",Name:db.Name,Account:account,Region:region,PublicIP:db.Public,RouteIGW:db.Public,Ingress:db.Ingress,Sensitive:db.Sensitive})}
func normalizeLambda(account,region string,fn ReplayLambda,out *Snapshot){key:="aws:lambda:"+account+":"+region+":"+fn.ID;out.Nodes=append(out.Nodes,Node{Key:key,Type:"LAMBDA",Name:fn.Name,Account:account,Region:region});if fn.RoleName!=""{out.Edges=append(out.Edges,Edge{From:key,To:"aws:role:"+account+":"+fn.RoleName,Type:"RUNS_AS",Evidence:"Replay Lambda execution role"})}}
func hasCoverage(c []Coverage,service,state string) bool{for _,x:=range c{if x.Service==service&&x.State==state{return true}};return false}

type ReplayScanLedger struct { Active map[string]Node }
func NewReplayScanLedger() *ReplayScanLedger{return &ReplayScanLedger{Active:map[string]Node{}}}
func (l *ReplayScanLedger) Apply(snapshot Snapshot,status string) { next:=map[string]Node{};for _,n:=range snapshot.Nodes{next[n.Key]=n};if status=="COMPLETE"{l.Active=next;return};for key,node:=range l.Active{if _,ok:=next[key];ok{l.Active[key]=node}} }
func (l *ReplayScanLedger) Contains(key string) bool {_,ok:=l.Active[key];return ok}

type ReplayScenario struct { Name string; Class string; Client *SimulatedAWS; ExpectRule string; ExpectNoRule string; ExpectPath bool; ExpectPartial bool }
func replayScenarioCatalog() []ReplayScenario {
	names:=[]string{"private-ec2","public-ssh","public-rdp","public-ip-no-route","igw-no-public-ip","wrong-port","restricted-cidr","multiple-security-groups","missing-route-association","missing-security-group-data","public-rds","private-rds","ipv6-present","nacl-uncertainty","simple-allow","simple-deny","wildcard-allow","wildcard-allow-explicit-deny","action-mismatch","resource-mismatch","inline-policy","managed-policy","trust-positive","trust-negative","cross-account-trust","cyclic-trust","malformed-policy","unsupported-condition","permission-boundary","scp-relevance","internet-privileged-role","internet-sensitive-s3","internet-denied-sensitive-s3","assume-role-chain","role-chain-sensitive","role-cycle-sensitive","private-high-privilege","initial-scan","identical-rescan","updated-resource","resource-disappears-complete","resource-disappears-partial","region-access-denied","iam-access-denied","s3-access-denied","rds-throttling","lambda-transient","missing-arn","missing-region","empty-page","out-of-order-pages"}
	out:=make([]ReplayScenario,0,len(names))
	for _,name:=range names{out=append(out,makeReplayScenario(name))}
	return out
}
func makeReplayScenario(name string) ReplayScenario {
	c:=baseSimulatedAWS()
	s:=ReplayScenario{Name:name,Class:"adversarial",Client:c}
	switch name {
	case "public-ssh","internet-privileged-role","internet-sensitive-s3","assume-role-chain","role-chain-sensitive","multiple-security-groups","out-of-order-pages":
		makePublicSSH(c)
	case "public-rdp":makePublicSSH(c);c.RegionPages["ap-south-1"][0].SecurityGroups[0].Ingress=[]Ingress{{Protocol:"tcp",FromPort:3389,ToPort:3389,CIDR:"0.0.0.0/0"}};s.ExpectRule="AG-NET-002"
	case "public-rds":makePublicRDS(c);s.ExpectRule="AG-NET-003"
	case "private-rds":makePrivateRDS(c)
	case "public-ip-no-route":makePublicSSH(c);c.RegionPages["ap-south-1"][0].Subnets[0].RouteIGW=false
	case "igw-no-public-ip":makePublicSSH(c);c.RegionPages["ap-south-1"][0].Instances[0].PublicIP=false
	case "wrong-port":makePublicSSH(c);c.RegionPages["ap-south-1"][0].SecurityGroups[0].Ingress=[]Ingress{{Protocol:"tcp",FromPort:443,ToPort:443,CIDR:"0.0.0.0/0"}}
	case "restricted-cidr":makePublicSSH(c);c.RegionPages["ap-south-1"][0].SecurityGroups[0].Ingress=[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"10.0.0.0/8"}}
	case "missing-route-association":makePublicSSH(c);c.RegionPages["ap-south-1"][0].Subnets[0].RouteKnown=false;s.ExpectPartial=true
	case "missing-security-group-data":makePublicSSH(c);c.RegionPages["ap-south-1"][0].SecurityGroups=nil;s.ExpectPartial=true
	case "ipv6-present":makePublicSSH(c);c.RegionPages["ap-south-1"][0].Instances[0].IPv6=true;s.ExpectPartial=true
	case "nacl-uncertainty":makePublicSSH(c);c.RegionPages["ap-south-1"][0].ErrorCode="NACL_UNKNOWN";s.ExpectPartial=true
	case "simple-allow","inline-policy","managed-policy","trust-positive":addSensitiveAllow(c)
	case "simple-deny","wildcard-allow-explicit-deny","internet-denied-sensitive-s3":addSensitiveDeny(c)
	case "wildcard-allow":addWildcardAllow(c);s.ExpectRule="AG-IAM-001"
	case "action-mismatch","resource-mismatch","trust-negative","cross-account-trust","permission-boundary","scp-relevance":c.IAMPages[0].ErrorCode="UnsupportedSemantics";s.ExpectPartial=true
	case "cyclic-trust","role-cycle-sensitive":addRoleCycle(c);c.IAMPages[0].ErrorCode="TrustCycle";s.ExpectPartial=true
	case "malformed-policy":c.IAMPages[0].Roles[0].Malformed=true;s.ExpectPartial=true
	case "unsupported-condition":c.IAMPages[0].Roles[0].UnsupportedCondition=true;s.ExpectPartial=true
	case "region-access-denied":c.Regions=[]string{"ap-south-1","us-east-1"};c.RegionPages["us-east-1"]=[]ReplayRegionPage{{ErrorCode:"AccessDenied"}};s.ExpectPartial=true
	case "iam-access-denied":c.IAMPages=[]ReplayIAMPage{{ErrorCode:"AccessDenied"}};s.ExpectPartial=true
	case "s3-access-denied":c.S3Pages=[]ReplayS3Page{{ErrorCode:"AccessDenied"}};s.ExpectPartial=true
	case "rds-throttling":c.RDSPages["ap-south-1"]=[]ReplayRDSPage{{ErrorCode:"Throttling"}};s.ExpectPartial=true
	case "lambda-transient":c.LambdaPages["ap-south-1"]=[]ReplayLambdaPage{{ErrorCode:"Transient"}};s.ExpectPartial=true
	case "missing-arn":c.S3Pages[0].Buckets[0].Name="";s.ExpectPartial=true
	case "missing-region":c.S3Pages[0].Buckets[0].Region="";s.ExpectPartial=true
	case "empty-page":c.RegionPages["ap-south-1"]=append([]ReplayRegionPage{{}},c.RegionPages["ap-south-1"]...);s.ExpectPartial=false
	case "initial-scan","identical-rescan","updated-resource","resource-disappears-complete","resource-disappears-partial":s.Class="lifecycle"
	}
	if name=="out-of-order-pages"{c.RegionPages["ap-south-1"]=append(c.RegionPages["ap-south-1"],ReplayRegionPage{VPCs:[]ReplayVPC{{ID:"vpc-second",Name:"second"}}})}
	if name=="internet-privileged-role"{addWildcardAllow(c)}
	if name=="internet-sensitive-s3"||name=="assume-role-chain"||name=="role-chain-sensitive"{addSensitiveAllow(c)}
	if name=="internet-denied-sensitive-s3"{makePublicSSH(c);addSensitiveDeny(c)}
	if name=="public-ssh"||name=="multiple-security-groups"||name=="out-of-order-pages"{s.ExpectRule="AG-NET-001"}
	if name=="internet-sensitive-s3"||name=="internet-privileged-role"||name=="assume-role-chain"||name=="role-chain-sensitive"{s.ExpectPath=true;s.ExpectRule="AG-COMB-002"}
	if strings.Contains(name,"denied-sensitive"){s.ExpectNoRule="AG-COMB-002"}
	if s.ExpectPartial{s.Class="failure"}
	return s
}
func baseSimulatedAWS() *SimulatedAWS {
	return &SimulatedAWS{Account:"111111111111",Regions:[]string{"ap-south-1"},RegionPages:map[string][]ReplayRegionPage{"ap-south-1":{{VPCs:[]ReplayVPC{{ID:"vpc-demo",Name:"demo-vpc"}},Subnets:[]ReplaySubnet{{ID:"subnet-demo",VPCID:"vpc-demo",Name:"demo-subnet",RouteKnown:true}},SecurityGroups:[]ReplaySecurityGroup{{ID:"sg-demo",Name:"demo-sg",VPCID:"vpc-demo",Ingress:[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"10.0.0.0/8"}}}},Instances:[]ReplayInstance{{ID:"i-demo",Name:"demo-workload",SubnetID:"subnet-demo",SecurityGroupIDs:[]string{"sg-demo"},RoleName:"demo-role"}}}}},IAMPages:[]ReplayIAMPage{{Roles:[]ReplayRole{{Name:"demo-role",Account:"111111111111"}}}},S3Pages:[]ReplayS3Page{{Buckets:[]ReplayBucket{{Name:"demo-sensitive",Region:"ap-south-1",Sensitive:true}}}},RDSPages:map[string][]ReplayRDSPage{"ap-south-1":{{}}},LambdaPages:map[string][]ReplayLambdaPage{"ap-south-1":{{}}},RDSCursor:map[string]int{},LambdaCursor:map[string]int{}}
}
func makePublicSSH(c *SimulatedAWS){p:=&c.RegionPages["ap-south-1"][0];p.Subnets[0].RouteIGW=true;p.SecurityGroups[0].Ingress=[]Ingress{{Protocol:"tcp",FromPort:22,ToPort:22,CIDR:"0.0.0.0/0"}};p.Instances[0].PublicIP=true}
func addSensitiveAllow(c *SimulatedAWS){c.IAMPages[0].Roles[0].Policies=[]Policy{{Effect:"Allow",Action:"s3:GetObject",Resource:"arn:aws:s3:::demo-sensitive/*"}}}
func addSensitiveDeny(c *SimulatedAWS){c.IAMPages[0].Roles[0].Policies=[]Policy{{Effect:"Allow",Action:"s3:GetObject",Resource:"arn:aws:s3:::demo-sensitive/*"},{Effect:"Deny",Action:"s3:GetObject",Resource:"arn:aws:s3:::demo-sensitive/*"}}}
func addWildcardAllow(c *SimulatedAWS){c.IAMPages[0].Roles[0].Policies=[]Policy{{Effect:"Allow",Action:"*",Resource:"*"}}}
func addRoleCycle(c *SimulatedAWS){c.IAMPages[0].Roles=append(c.IAMPages[0].Roles,ReplayRole{Name:"second-role",Account:c.Account,Trust:[]string{"demo-role"}});c.IAMPages[0].Roles[0].Trust=[]string{"second-role"}}
func makePublicRDS(c *SimulatedAWS){p:=&c.RegionPages["ap-south-1"][0];p.Subnets[0].RouteIGW=true;p.RDS=[]ReplayRDS{{ID:"db-demo",Name:"public-db",Public:true,Ingress:[]Ingress{{Protocol:"tcp",FromPort:5432,ToPort:5432,CIDR:"0.0.0.0/0"}},Sensitive:true}}}
func makePrivateRDS(c *SimulatedAWS){p:=&c.RegionPages["ap-south-1"][0];p.RDS=[]ReplayRDS{{ID:"db-private",Name:"private-db",Public:false,Sensitive:true}}}
