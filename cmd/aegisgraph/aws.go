package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type liveAWSClient struct {
	cfg aws.Config
	iamClient *iam.Client
	s3Client *s3.Client
	ec2Clients map[string]*ec2.Client
	rdsClients map[string]*rds.Client
	lambdaClients map[string]*lambda.Client
	profileRoles map[string]string
	profilesLoaded bool
	iamPages []ReplayIAMPage
	iamLoaded bool
	iamLoadErr error
	s3Pages []ReplayS3Page
	s3Loaded bool
	s3LoadErr error
	regionState map[string]liveRegionState
}

type liveRegionState struct {
	routeIGW map[string]bool
	routeKnown bool
	sgIngress map[string][]Ingress
}

func newLiveAWSClient(cfg aws.Config) *liveAWSClient {
	return &liveAWSClient{
		cfg: cfg,
		iamClient: iam.NewFromConfig(cfg),
		s3Client: s3.NewFromConfig(cfg),
		ec2Clients: map[string]*ec2.Client{},
		rdsClients: map[string]*rds.Client{},
		lambdaClients: map[string]*lambda.Client{},
		profileRoles: map[string]string{},
		regionState: map[string]liveRegionState{},
	}
}

func (c *liveAWSClient) GetCallerIdentity(ctx context.Context) (string, error) {
	out, err := sts.NewFromConfig(c.cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil { return "", err }
	account := aws.ToString(out.Account)
	if account == "" { return "", errors.New("AWS caller identity did not include an account") }
	return account, nil
}

func (c *liveAWSClient) ListRegions(ctx context.Context) ([]string, error) {
	client := ec2.NewFromConfig(c.cfg)
	var regions []string
	var token *string
	for {
		out, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{AllRegions: aws.Bool(false), NextToken: token})
		if err != nil { return nil, err }
		for _, region := range out.Regions {
			if name := aws.ToString(region.RegionName); name != "" { regions = append(regions, name) }
		}
		if aws.ToString(out.NextToken) == "" { break }
		token = out.NextToken
	}
	sort.Strings(regions)
	return regions, nil
}

func (c *liveAWSClient) ec2(region string) *ec2.Client {
	if client := c.ec2Clients[region]; client != nil { return client }
	cfg := c.cfg
	cfg.Region = region
	client := ec2.NewFromConfig(cfg)
	c.ec2Clients[region] = client
	return client
}

func (c *liveAWSClient) rds(region string) *rds.Client {
	if client := c.rdsClients[region]; client != nil { return client }
	cfg := c.cfg
	cfg.Region = region
	client := rds.NewFromConfig(cfg)
	c.rdsClients[region] = client
	return client
}

func (c *liveAWSClient) lambda(region string) *lambda.Client {
	if client := c.lambdaClients[region]; client != nil { return client }
	cfg := c.cfg
	cfg.Region = region
	client := lambda.NewFromConfig(cfg)
	c.lambdaClients[region] = client
	return client
}

func (c *liveAWSClient) DescribeRegion(ctx context.Context, region, token string) (ReplayRegionPage, error) {
	if token != "" { return ReplayRegionPage{}, errors.New("live regional composite pagination must restart from the first page") }
	if err := c.ensureProfiles(ctx); err != nil { return ReplayRegionPage{ErrorCode: classifyAWSError(err)}, nil }
	client := c.ec2(region)
	page := ReplayRegionPage{}
	vpcs, err := collectVPCs(ctx, client)
	if err != nil { return page, err }
	page.VPCs = vpcs
	subnets, err := collectSubnets(ctx, client)
	if err != nil { return page, err }
	page.Subnets = subnets
	routes, err := collectRouteTables(ctx, client)
	if err != nil { return page, err }
	igws, err := collectInternetGateways(ctx, client)
	if err != nil { return page, err }
	securityGroups, err := collectSecurityGroups(ctx, client)
	if err != nil { return page, err }
	page.SecurityGroups = securityGroups
	instances, err := collectInstances(ctx, client)
	if err != nil { return page, err }
	for i := range instances {
		if instances[i].InstanceProfileName != "" {
			if role := c.profileRoles[instances[i].InstanceProfileName]; role != "" { instances[i].RoleName = role }
		}
	}
	page.Instances = instances
	routeIGW, routeKnown := buildRouteState(routes, igws)
	for i := range page.Subnets {
		page.Subnets[i].RouteIGW = routeIGW[page.Subnets[i].ID]
		page.Subnets[i].RouteKnown = routeKnown
	}
	sgIngress := map[string][]Ingress{}
	for _, group := range page.SecurityGroups { sgIngress[group.ID] = append([]Ingress{}, group.Ingress...) }
	c.regionState[region] = liveRegionState{routeIGW:routeIGW, routeKnown:routeKnown, sgIngress:sgIngress}
	return page, nil
}

func collectVPCs(ctx context.Context, client *ec2.Client) ([]ReplayVPC, error) {
	var out []ReplayVPC
	var token *string
	for {
		page, err := client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{NextToken: token})
		if err != nil { return nil, err }
		for _, vpc := range page.Vpcs {
			if id := aws.ToString(vpc.VpcId); id != "" { out = append(out, ReplayVPC{ID:id,Name:tagName(vpc.Tags,id)}) }
		}
		if aws.ToString(page.NextToken) == "" { return out, nil }
		token = page.NextToken
	}
}

func collectSubnets(ctx context.Context, client *ec2.Client) ([]ReplaySubnet, error) {
	var out []ReplaySubnet
	var token *string
	for {
		page, err := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{NextToken: token})
		if err != nil { return nil, err }
		for _, subnet := range page.Subnets {
			if id := aws.ToString(subnet.SubnetId); id != "" { out = append(out, ReplaySubnet{ID:id,VPCID:aws.ToString(subnet.VpcId),Name:tagName(subnet.Tags,id)}) }
		}
		if aws.ToString(page.NextToken) == "" { return out, nil }
		token = page.NextToken
	}
}

func collectRouteTables(ctx context.Context, client *ec2.Client) ([]ec2types.RouteTable, error) {
	var out []ec2types.RouteTable
	var token *string
	for {
		page, err := client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{NextToken: token})
		if err != nil { return nil, err }
		out = append(out, page.RouteTables...)
		if aws.ToString(page.NextToken) == "" { return out, nil }
		token = page.NextToken
	}
}

func collectInternetGateways(ctx context.Context, client *ec2.Client) ([]ec2types.InternetGateway, error) {
	var out []ec2types.InternetGateway
	var token *string
	for {
		page, err := client.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{NextToken: token})
		if err != nil { return nil, err }
		out = append(out, page.InternetGateways...)
		if aws.ToString(page.NextToken) == "" { return out, nil }
		token = page.NextToken
	}
}

func collectSecurityGroups(ctx context.Context, client *ec2.Client) ([]ReplaySecurityGroup, error) {
	var out []ReplaySecurityGroup
	var token *string
	for {
		page, err := client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{NextToken: token})
		if err != nil { return nil, err }
		for _, group := range page.SecurityGroups {
			id := aws.ToString(group.GroupId)
			if id == "" { continue }
			ingress := []Ingress{}
			for _, permission := range group.IpPermissions {
				protocol := aws.ToString(permission.IpProtocol)
				from, to := 0, 65535
				if permission.FromPort != nil { from = int(aws.ToInt32(permission.FromPort)) }
				if permission.ToPort != nil { to = int(aws.ToInt32(permission.ToPort)) }
				for _, ip := range permission.IpRanges { ingress = append(ingress, Ingress{Protocol:protocol,FromPort:from,ToPort:to,CIDR:aws.ToString(ip.CidrIp)}) }
			}
			out = append(out, ReplaySecurityGroup{ID:id,Name:tagName(group.Tags,aws.ToString(group.GroupName)),VPCID:aws.ToString(group.VpcId),Ingress:ingress,Present:true})
		}
		if aws.ToString(page.NextToken) == "" { return out, nil }
		token = page.NextToken
	}
}

func collectInstances(ctx context.Context, client *ec2.Client) ([]ReplayInstance, error) {
	var out []ReplayInstance
	var token *string
	for {
		page, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{NextToken: token})
		if err != nil { return nil, err }
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				id := aws.ToString(instance.InstanceId)
				if id == "" { continue }
				profile := ""
				if instance.IamInstanceProfile != nil { profile = profileNameFromARN(aws.ToString(instance.IamInstanceProfile.Arn)) }
				out = append(out, ReplayInstance{ID:id,Name:tagName(instance.Tags,id),SubnetID:aws.ToString(instance.SubnetId),SecurityGroupIDs:securityGroupIDs(instance.SecurityGroups),PublicIP:aws.ToString(instance.PublicIpAddress)!="",PublicKnown:true,IPv6:len(instance.Ipv6Addresses)>0,InstanceProfileName:profile})
			}
		}
		if aws.ToString(page.NextToken) == "" { return out, nil }
		token = page.NextToken
	}
}

func securityGroupIDs(groups []ec2types.GroupIdentifier) []string {
	out:=make([]string,0,len(groups))
	for _,group:=range groups { if id:=aws.ToString(group.GroupId);id!=""{out=append(out,id)} }
	return out
}

func buildRouteState(routes []ec2types.RouteTable, igws []ec2types.InternetGateway) (map[string]bool,bool) {
	out:=map[string]bool{}
	for _,table:=range routes {
		hasIGW:=false
		for _,route:=range table.Routes {
			if aws.ToString(route.DestinationCidrBlock)=="0.0.0.0/0" && strings.HasPrefix(aws.ToString(route.GatewayId),"igw-") { hasIGW=true }
		}
		if !hasIGW { continue }
		for _,association:=range table.Associations {
			if subnet:=aws.ToString(association.SubnetId);subnet!="" { out[subnet]=true }
		}
	}
	return out,true
}

func (c *liveAWSClient) ListIAM(ctx context.Context, token string) (ReplayIAMPage, error) {
	if !c.iamLoaded { c.iamPages, c.iamLoadErr = c.loadIAM(ctx); c.iamLoaded = true }
	if c.iamLoadErr != nil { return ReplayIAMPage{}, c.iamLoadErr }
	index := tokenIndex(token)
	if index >= len(c.iamPages) { return ReplayIAMPage{}, nil }
	page := c.iamPages[index]
	if index+1 < len(c.iamPages) && page.NextToken=="" { page.NextToken = fmt.Sprintf("page-%d",index+1) }
	return page,nil
}

func (c *liveAWSClient) ensureProfiles(ctx context.Context) error {
	if c.profilesLoaded { return nil }
	var token *string
	for {
		page, err := c.iamClient.ListInstanceProfiles(ctx,&iam.ListInstanceProfilesInput{Marker:token})
		if err != nil { return err }
		for _,profile := range page.InstanceProfiles {
			name:=aws.ToString(profile.InstanceProfileName)
			for _,role:=range profile.Roles { if name!="" { c.profileRoles[name]=aws.ToString(role.RoleName) } }
		}
		if aws.ToString(page.Marker)=="" { break }
		token=page.Marker
	}
	c.profilesLoaded=true
	return nil
}

func (c *liveAWSClient) loadIAM(ctx context.Context) ([]ReplayIAMPage,error) {
	roles, roleErr := c.loadRoles(ctx)
	users, userErr := c.loadUsers(ctx)
	profiles, profileErr := c.loadProfiles(ctx)
	policies, policyErr := c.loadPolicies(ctx)
	page:=ReplayIAMPage{Roles:roles,Users:users,InstanceProfiles:profiles,Policies:policies}
	for _,err:=range []error{roleErr,userErr,profileErr,policyErr}{if err!=nil{page.ErrorCode=classifyAWSError(err);break}}
	return []ReplayIAMPage{page},nil
}

func (c *liveAWSClient) loadRoles(ctx context.Context) ([]ReplayRole,error) {
	var out []ReplayRole
	var token *string
	for {
		page,err:=c.iamClient.ListRoles(ctx,&iam.ListRolesInput{Marker:token})
		if err!=nil{return out,err}
		for _,role:=range page.Roles {
			name:=aws.ToString(role.RoleName)
			if name==""{continue}
			policies,policyErr:=c.rolePolicies(ctx,name)
			trust,trustErr:=parseTrustPolicy(aws.ToString(role.AssumeRolePolicyDocument))
			out=append(out,ReplayRole{Name:name,ARN:aws.ToString(role.Arn),Policies:policies,Trust:trust})
			if policyErr!=nil{return out,policyErr}
			if trustErr!=nil{return out,trustErr}
		}
		if aws.ToString(page.Marker)==""{break};token=page.Marker
	}
	return out,nil
}

func (c *liveAWSClient) rolePolicies(ctx context.Context,roleName string)([]Policy,error) {
	var out []Policy
	var token *string
	for {
		page,err:=c.iamClient.ListRolePolicies(ctx,&iam.ListRolePoliciesInput{RoleName:aws.String(roleName),Marker:token})
		if err!=nil{return out,err}
		for _,name:=range page.PolicyNames {
			policy,err:=c.iamClient.GetRolePolicy(ctx,&iam.GetRolePolicyInput{RoleName:aws.String(roleName),PolicyName:aws.String(name)})
			if err!=nil{return out,err}
			parsed,unsupported,err:=parsePolicyDocument(aws.ToString(policy.PolicyDocument))
			if unsupported||err!=nil{if err!=nil{return out,err};continue}
			out=append(out,parsed...)
		}
		if aws.ToString(page.Marker)==""{break};token=page.Marker
	}
	attached,err:=c.iamClient.ListAttachedRolePolicies(ctx,&iam.ListAttachedRolePoliciesInput{RoleName:aws.String(roleName)})
	if err!=nil{return out,err}
	for _,policy:=range attached.AttachedPolicies {
		document,err:=c.managedPolicyDocument(ctx,aws.ToString(policy.PolicyArn))
		if err!=nil{return out,err}
		out=append(out,document...)
	}
	return out,nil
}

func (c *liveAWSClient) loadUsers(ctx context.Context)([]ReplayUser,error) {
	var out []ReplayUser
	var token *string
	for {
		page,err:=c.iamClient.ListUsers(ctx,&iam.ListUsersInput{Marker:token})
		if err!=nil{return out,err}
		for _,user:=range page.Users {out=append(out,ReplayUser{Name:aws.ToString(user.UserName),ARN:aws.ToString(user.Arn)})}
		if aws.ToString(page.Marker)==""{break};token=page.Marker
	}
	return out,nil
}

func (c *liveAWSClient) loadProfiles(ctx context.Context)([]ReplayInstanceProfile,error) {
	var out []ReplayInstanceProfile
	var token *string
	for {
		page,err:=c.iamClient.ListInstanceProfiles(ctx,&iam.ListInstanceProfilesInput{Marker:token})
		if err!=nil{return out,err}
		for _,profile:=range page.InstanceProfiles {
			name:=aws.ToString(profile.InstanceProfileName)
			roles:=[]string{}
			for _,role:=range profile.Roles{roles=append(roles,aws.ToString(role.RoleName))}
			out=append(out,ReplayInstanceProfile{Name:name,ARN:aws.ToString(profile.Arn),RoleNames:roles})
		}
		if aws.ToString(page.Marker)==""{break};token=page.Marker
	}
	return out,nil
}

func (c *liveAWSClient) loadPolicies(ctx context.Context)([]ReplayPolicy,error) {
	var out []ReplayPolicy
	var token *string
	for {
		page,err:=c.iamClient.ListPolicies(ctx,&iam.ListPoliciesInput{Scope:iamtypes.PolicyScopeTypeLocal,Marker:token})
		if err!=nil{return out,err}
		for _,policy:=range page.Policies {
			arn:=aws.ToString(policy.Arn)
			document,err:=c.managedPolicyDocument(ctx,arn)
			if err!=nil{return out,err}
			out=append(out,ReplayPolicy{Name:aws.ToString(policy.PolicyName),ARN:arn,Document:document})
		}
		if aws.ToString(page.Marker)==""{break};token=page.Marker
	}
	return out,nil
}

func (c *liveAWSClient) managedPolicyDocument(ctx context.Context,arn string)([]Policy,error) {
	policy,err:=c.iamClient.GetPolicy(ctx,&iam.GetPolicyInput{PolicyArn:aws.String(arn)})
	if err!=nil{return nil,err}
	version,err:=c.iamClient.GetPolicyVersion(ctx,&iam.GetPolicyVersionInput{PolicyArn:aws.String(arn),VersionId:policy.Policy.DefaultVersionId})
	if err!=nil{return nil,err}
	parsed,unsupported,err:=parsePolicyDocument(aws.ToString(version.PolicyVersion.Document))
	if unsupported{return nil,nil}
	return parsed,err
}

func parsePolicyDocument(raw string)([]Policy,bool,error) {
	decoded,err:=url.QueryUnescape(raw)
	if err!=nil{return nil,true,err}
	var document struct{Statement json.RawMessage}
	if err:=json.Unmarshal([]byte(decoded),&document);err!=nil{return nil,true,err}
	var statements []json.RawMessage
	if len(document.Statement)>0 && document.Statement[0]=='[' { _=json.Unmarshal(document.Statement,&statements) } else if len(document.Statement)>0 { statements=[]json.RawMessage{document.Statement} }
	var out []Policy
	unsupported:=false
	for _,rawStatement:=range statements {
		var statement struct{Effect string;Action json.RawMessage;Resource json.RawMessage;Condition json.RawMessage}
		if err:=json.Unmarshal(rawStatement,&statement);err!=nil{return nil,true,err}
		if len(statement.Condition)>0 && string(statement.Condition)!="null" {unsupported=true;continue}
		actions:=jsonStrings(statement.Action);resources:=jsonStrings(statement.Resource)
		for _,action:=range actions{for _,resource:=range resources{out=append(out,Policy{Effect:statement.Effect,Action:action,Resource:resource})}}
	}
	return out,unsupported,nil
}

func jsonStrings(raw json.RawMessage)[]string {
	if len(raw)==0{return nil}
	var single string
	if json.Unmarshal(raw,&single)==nil{return []string{single}}
	var many []string
	_ = json.Unmarshal(raw,&many)
	return many
}

func parseTrustPolicy(raw string)([]string,error) {
	decoded,err:=url.QueryUnescape(raw)
	if err!=nil{return nil,err}
	var document struct{Statement json.RawMessage}
	if err:=json.Unmarshal([]byte(decoded),&document);err!=nil{return nil,err}
	var statements []json.RawMessage
	if len(document.Statement)>0 && document.Statement[0]=='[' { _=json.Unmarshal(document.Statement,&statements) } else if len(document.Statement)>0 {statements=[]json.RawMessage{document.Statement}}
	var out []string
	for _,rawStatement:=range statements {
		var statement struct{Principal json.RawMessage}
		if err:=json.Unmarshal(rawStatement,&statement);err!=nil{return nil,err}
		var principal struct{AWS json.RawMessage}
		if json.Unmarshal(statement.Principal,&principal)!=nil{continue}
		for _,arn:=range jsonStrings(principal.AWS){if strings.Contains(arn,":role/"){out=append(out,roleNameFromARN(arn))}}
	}
	sort.Strings(out)
	return uniqueStrings(out),nil
}

func (c *liveAWSClient) ListS3(ctx context.Context, token string)(ReplayS3Page,error) {
	if !c.s3Loaded{c.s3Pages,c.s3LoadErr=c.loadS3(ctx);c.s3Loaded=true}
	if c.s3LoadErr!=nil{return ReplayS3Page{},c.s3LoadErr}
	index:=tokenIndex(token);if index>=len(c.s3Pages){return ReplayS3Page{},nil}
	page:=c.s3Pages[index];if index+1<len(c.s3Pages)&&page.NextToken==""{page.NextToken=fmt.Sprintf("page-%d",index+1)}
	return page,nil
}

func (c *liveAWSClient) loadS3(ctx context.Context)([]ReplayS3Page,error) {
	pager:=s3.NewListBucketsPaginator(c.s3Client,&s3.ListBucketsInput{})
	var buckets []ReplayBucket
	for pager.HasMorePages(){
		page,err:=pager.NextPage(ctx);if err!=nil{return nil,err}
		for _,bucket:=range page.Buckets{
			name:=aws.ToString(bucket.Name);if name==""{continue}
			item:=ReplayBucket{Name:name}
			location,err:=c.s3Client.GetBucketLocation(ctx,&s3.GetBucketLocationInput{Bucket:aws.String(name)})
			if err!=nil{return []ReplayS3Page{{Buckets:append(buckets,item),ErrorCode:classifyAWSError(err)}},nil}
			item.Region=string(location.LocationConstraint);if item.Region==""{item.Region="us-east-1"}
			block,blockErr:=c.s3Client.GetPublicAccessBlock(ctx,&s3.GetPublicAccessBlockInput{Bucket:aws.String(name)})
			if blockErr==nil&&block.PublicAccessBlockConfiguration!=nil{
				item.PublicAccessKnown=true
				cfg:=block.PublicAccessBlockConfiguration
				item.PublicAccessBlocked=aws.ToBool(cfg.BlockPublicAcls)&&aws.ToBool(cfg.IgnorePublicAcls)&&aws.ToBool(cfg.BlockPublicPolicy)&&aws.ToBool(cfg.RestrictPublicBuckets)
			}
			policy,policyErr:=c.s3Client.GetBucketPolicy(ctx,&s3.GetBucketPolicyInput{Bucket:aws.String(name)})
			if policyErr==nil{item.PolicyKnown=true;item.PolicyPublic=policyDocumentPublic(aws.ToString(policy.Policy))}
			enc,encErr:=c.s3Client.GetBucketEncryption(ctx,&s3.GetBucketEncryptionInput{Bucket:aws.String(name)})
			if encErr==nil{item.EncryptionKnown=true;item.Encrypted=enc.ServerSideEncryptionConfiguration!=nil}
			tags,tagErr:=c.s3Client.GetBucketTagging(ctx,&s3.GetBucketTaggingInput{Bucket:aws.String(name)})
			if tagErr==nil{item.Sensitive=tagsContainSensitive(tags.TagSet)}
			if blockErr!=nil||policyErr!=nil||encErr!=nil{item.PublicKnown=false}
			buckets=append(buckets,item)
		}
	}
	return []ReplayS3Page{{Buckets:buckets}},nil
}

func policyDocumentPublic(raw string)bool {
	var document struct{Statement json.RawMessage}
	if json.Unmarshal([]byte(raw),&document)!=nil{return false}
	var statements []json.RawMessage
	if len(document.Statement)>0&&document.Statement[0]=='['{_ = json.Unmarshal(document.Statement,&statements)}else if len(document.Statement)>0{statements=[]json.RawMessage{document.Statement}}
	for _,statement:=range statements{
		var item struct{Principal json.RawMessage;Effect string}
		if json.Unmarshal(statement,&item)!=nil{continue}
		if item.Effect!="Allow"{continue}
		if strings.Contains(string(item.Principal),"\"*\""){return true}
	}
	return false
}

func tagsContainSensitive(tags []s3types.Tag)bool {
	for _,tag:=range tags{
		key:=strings.ToLower(aws.ToString(tag.Key));value:=strings.ToLower(aws.ToString(tag.Value))
		if (key=="sensitivity"||key=="data-classification")&&(value=="sensitive"||value=="critical"||value=="confidential"){return true}
	}
	return false
}

func (c *liveAWSClient) ListRDS(ctx context.Context,region,token string)(ReplayRDSPage,error) {
	client:=c.rds(region)
	page,err:=client.DescribeDBInstances(ctx,&rds.DescribeDBInstancesInput{Marker:aws.String(token)})
	if err!=nil{return ReplayRDSPage{},err}
	state:=c.regionState[region]
	out:=ReplayRDSPage{}
	for _,db:=range page.DBInstances{
		item:=ReplayRDS{ID:aws.ToString(db.DBInstanceIdentifier),Name:aws.ToString(db.DBInstanceIdentifier),Engine:aws.ToString(db.Engine),Public:aws.ToBool(db.PubliclyAccessible),PublicKnown:db.PubliclyAccessible!=nil,Encrypted:aws.ToBool(db.StorageEncrypted),EncryptionKnown:db.StorageEncrypted!=nil,RouteIGW:false,RouteKnown:state.routeKnown}
		if db.DBSubnetGroup!=nil{for _,subnet:=range db.DBSubnetGroup.Subnets{if id:=aws.ToString(subnet.SubnetIdentifier);id!=""{item.SubnetIDs=append(item.SubnetIDs,id);if state.routeIGW[id]{item.RouteIGW=true}}}}
		for _,membership:=range db.VpcSecurityGroups{if id:=aws.ToString(membership.VpcSecurityGroupId);id!=""{item.SecurityGroupIDs=append(item.SecurityGroupIDs,id);item.Ingress=append(item.Ingress,state.sgIngress[id]...)}}
		out.Instances=append(out.Instances,item)
	}
	out.NextToken=aws.ToString(page.Marker)
	return out,nil
}

func (c *liveAWSClient) ListLambda(ctx context.Context,region,token string)(ReplayLambdaPage,error) {
	page,err:=c.lambda(region).ListFunctions(ctx,&lambda.ListFunctionsInput{Marker:aws.String(token)})
	if err!=nil{return ReplayLambdaPage{},err}
	out:=ReplayLambdaPage{NextToken:aws.ToString(page.NextMarker)}
	for _,fn:=range page.Functions{
		id:=aws.ToString(fn.FunctionArn);if id==""{id=aws.ToString(fn.FunctionName)}
		out.Functions=append(out.Functions,ReplayLambda{ID:id,Name:aws.ToString(fn.FunctionName),RoleName:roleNameFromARN(aws.ToString(fn.Role))})
	}
	return out,nil
}

func classifyAWSError(err error)string {
	if err==nil{return ""}
	message:=strings.ToLower(err.Error())
	switch{
	case strings.Contains(message,"accessdenied"),strings.Contains(message,"access denied"),strings.Contains(message,"unauthorized"):return "AccessDenied"
	case strings.Contains(message,"throttl"),strings.Contains(message,"rate exceeded"):return "Throttling"
	default:return "Transient"
	}
}

func profileNameFromARN(arn string)string {if i:=strings.LastIndex(arn,"/");i>=0{return arn[i+1:]};return arn}
func roleNameFromARN(arn string)string {if i:=strings.LastIndex(arn,"/");i>=0{return arn[i+1:]};return arn}
func uniqueStrings(values []string)[]string{seen:=map[string]bool{};out:=[]string{};for _,value:=range values{if value!=""&&!seen[value]{seen[value]=true;out=append(out,value)}};return out}

func collectAWS(ctx context.Context) (Snapshot,error) {
	cfg,err:=config.LoadDefaultConfig(ctx)
	if err!=nil{return Snapshot{},err}
	result,err:=collectReplay(ctx,newLiveAWSClient(cfg))
	if err!=nil{return Snapshot{},err}
	result.Snapshot.Environment="aws"
	return result.Snapshot,nil
}

func tagName(tags []ec2types.Tag,fallback string)string{
	for _,tag:=range tags{if aws.ToString(tag.Key)=="Name"&&aws.ToString(tag.Value)!=""{return aws.ToString(tag.Value)}}
	return fallback
}
