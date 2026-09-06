package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func collectAWS(ctx context.Context) (Snapshot, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil { return Snapshot{}, err }
	identity, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil { return Snapshot{}, err }
	account := aws.ToString(identity.Account)
	if account == "" { return Snapshot{}, errors.New("AWS caller identity did not include an account") }

	snapshot := Snapshot{Environment:"aws", Coverage:[]Coverage{{"sts","COMPLETE","Caller identity resolved through the normal AWS SDK credential chain."}}}
	snapshot.Nodes = append(snapshot.Nodes, Node{Key:"aws:account:"+account,Type:"AWS_ACCOUNT",Name:"AWS account "+account,Account:account})
	regionsOut, err := ec2.NewFromConfig(cfg).DescribeRegions(ctx, &ec2.DescribeRegionsInput{AllRegions:aws.Bool(false)})
	if err != nil {
		snapshot.Coverage=append(snapshot.Coverage,Coverage{"ec2:DescribeRegions","FAILED",err.Error()})
		return snapshot, err
	}
	processed:=0
	for _, region := range regionsOut.Regions {
		if processed >= 50 { snapshot.Coverage=append(snapshot.Coverage,Coverage{"regions","PARTIAL","Region processing capped at 50 enabled regions."}); break }
		name:=aws.ToString(region.RegionName)
		if name=="" { continue }
		processed++
		if err:=collectAWSRegion(ctx,cfg,account,name,&snapshot); err!=nil {
			snapshot.Coverage=append(snapshot.Coverage,Coverage{"ec2:"+name,"PARTIAL",err.Error()})
		} else {
			snapshot.Coverage=append(snapshot.Coverage,Coverage{"ec2:"+name,"COMPLETE","VPC, subnet, route, Internet Gateway, security group, and instance inventory collected."})
		}
	}
	return snapshot,nil
}

func collectAWSRegion(ctx context.Context,cfg aws.Config,account,region string,s *Snapshot) error {
	cfg.Region=region
	client:=ec2.NewFromConfig(cfg)
	vpcsOut,err:=client.DescribeVpcs(ctx,&ec2.DescribeVpcsInput{})
	if err!=nil{return fmt.Errorf("DescribeVpcs: %w",err)}
	subnetsOut,err:=client.DescribeSubnets(ctx,&ec2.DescribeSubnetsInput{})
	if err!=nil{return fmt.Errorf("DescribeSubnets: %w",err)}
	routesOut,err:=client.DescribeRouteTables(ctx,&ec2.DescribeRouteTablesInput{})
	if err!=nil{return fmt.Errorf("DescribeRouteTables: %w",err)}
	igwsOut,err:=client.DescribeInternetGateways(ctx,&ec2.DescribeInternetGatewaysInput{})
	if err!=nil{return fmt.Errorf("DescribeInternetGateways: %w",err)}
	sgOut,err:=client.DescribeSecurityGroups(ctx,&ec2.DescribeSecurityGroupsInput{})
	if err!=nil{return fmt.Errorf("DescribeSecurityGroups: %w",err)}
	instancesOut,err:=client.DescribeInstances(ctx,&ec2.DescribeInstancesInput{})
	if err!=nil{return fmt.Errorf("DescribeInstances: %w",err)}

	vpcKeys:=map[string]string{}
	for _,v:=range vpcsOut.Vpcs {
		id:=aws.ToString(v.VpcId); if id==""{continue}
		key:="aws:vpc:"+account+":"+region+":"+id; vpcKeys[id]=key
		s.Nodes=append(s.Nodes,Node{Key:key,Type:"VPC",Name:tagName(v.Tags,id),Account:account,Region:region})
		s.Edges=append(s.Edges,Edge{From:"aws:account:"+account,To:key,Type:"CONTAINS",Evidence:"EC2 DescribeVpcs"})
	}
	igwByVpc:=map[string]bool{}
	for _,igw:=range igwsOut.InternetGateways {
		for _,a:=range igw.Attachments {
			if aws.ToString(a.State)=="available" && aws.ToString(a.VpcId)!="" { igwByVpc[aws.ToString(a.VpcId)]=true }
		}
	}
	routeBySubnet:=map[string]bool{}
	for _,rt:=range routesOut.RouteTables {
		hasIGW:=false
		for _,route:=range rt.Routes {
			if aws.ToString(route.DestinationCidrBlock)=="0.0.0.0/0" && stringsHasPrefix(aws.ToString(route.GatewayId),"igw-") { hasIGW=true }
		}
		if !hasIGW { continue }
		for _,a:=range rt.Associations {
			if aws.ToString(a.SubnetId)!="" { routeBySubnet[aws.ToString(a.SubnetId)]=true }
		}
	}
	subnetKeys:=map[string]string{}
	for _,sub:=range subnetsOut.Subnets {
		id:=aws.ToString(sub.SubnetId); vpc:=aws.ToString(sub.VpcId); if id==""{continue}
		key:="aws:subnet:"+account+":"+region+":"+id; subnetKeys[id]=key
		s.Nodes=append(s.Nodes,Node{Key:key,Type:"SUBNET",Name:tagName(sub.Tags,id),Account:account,Region:region,RouteIGW:routeBySubnet[id]})
		if vk:=vpcKeys[vpc];vk!=""{s.Edges=append(s.Edges,Edge{From:vk,To:key,Type:"CONTAINS",Evidence:"EC2 DescribeSubnets VPC association"})}
	}
	sgIngress:=map[string][]Ingress{}
	for _,sg:=range sgOut.SecurityGroups {
		id:=aws.ToString(sg.GroupId);if id==""{continue}
		key:="aws:security-group:"+account+":"+region+":"+id
		s.Nodes=append(s.Nodes,Node{Key:key,Type:"SECURITY_GROUP",Name:tagName(sg.Tags,aws.ToString(sg.GroupName)),Account:account,Region:region})
		if vk:=vpcKeys[aws.ToString(sg.VpcId)];vk!=""{s.Edges=append(s.Edges,Edge{From:vk,To:key,Type:"CONTAINS",Evidence:"EC2 DescribeSecurityGroups VPC association"})}
		for _,perm:=range sg.IpPermissions {
			proto:=aws.ToString(perm.IpProtocol);from,to:=0,65535
			if perm.FromPort!=nil{from=int(aws.ToInt32(perm.FromPort))}
			if perm.ToPort!=nil{to=int(aws.ToInt32(perm.ToPort))}
			for _,r:=range perm.IpRanges { sgIngress[id]=append(sgIngress[id],Ingress{Protocol:proto,FromPort:from,ToPort:to,CIDR:aws.ToString(r.CidrIp)}) }
		}
	}
	for _,reservation:=range instancesOut.Reservations {
		for _,instance:=range reservation.Instances {
			id:=aws.ToString(instance.InstanceId);if id==""{continue}
			key:="aws:ec2:"+account+":"+region+":"+id
			ingress:=[]Ingress{};for _,sg:=range instance.SecurityGroups{ingress=append(ingress,sgIngress[aws.ToString(sg.GroupId)]...)}
			n:=Node{Key:key,Type:"EC2",Name:tagName(instance.Tags,id),Account:account,Region:region,PublicIP:aws.ToString(instance.PublicIpAddress)!="",Ingress:ingress}
			n.RouteIGW=routeBySubnet[aws.ToString(instance.SubnetId)] && igwByVpc[aws.ToString(instance.VpcId)]
			s.Nodes=append(s.Nodes,n)
			if sk:=subnetKeys[aws.ToString(instance.SubnetId)];sk!=""{s.Edges=append(s.Edges,Edge{From:sk,To:key,Type:"CONTAINS",Evidence:"EC2 DescribeInstances subnet association"})}
			for _,sg:=range instance.SecurityGroups { sgKey:="aws:security-group:"+account+":"+region+":"+aws.ToString(sg.GroupId);s.Edges=append(s.Edges,Edge{From:key,To:sgKey,Type:"USES_SECURITY_GROUP",Evidence:"EC2 DescribeInstances security group association"}) }
		}
	}
	return nil
}

func tagName(tags []ec2types.Tag,fallback string) string {
	for _,tag:=range tags { if aws.ToString(tag.Key)=="Name" && aws.ToString(tag.Value)!="" { return aws.ToString(tag.Value) } }
	return fallback
}
func stringsHasPrefix(value,prefix string) bool {
	if len(value)<len(prefix){return false}
	return value[:len(prefix)]==prefix
}
