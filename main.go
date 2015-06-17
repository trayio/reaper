package main

import (
	"flag"
	"log"
	"os"
	"sort"

	"github.com/trayio/reaper/candidates"
	"github.com/trayio/reaper/collector"
	"github.com/trayio/reaper/config"

	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/aws"
	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/aws/awsutil"
	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/service/ec2"
)

var regions = []string{
	"ap-northeast-1",
	"ap-southeast-1",
	"ap-southeast-2",
	"eu-central-1",
	"eu-west-1",
	"sa-east-1",
	"us-east-1",
	"us-west-1",
	"us-west-2",
}

func main() {
	groupTag := flag.String("tag", "group", "Tag name to group instances by")
	configFile := flag.String("c", "conf.js", "Configuration file.")
	dryRun := flag.Bool("dry", false, "Enable dry run.")

	accessId := flag.String("access", "", "AWS access ID")
	secretKey := flag.String("secret", "", "AWS secret key")
	region := flag.String("region", "us-west-1", "AWS region")
	flag.Parse()

	cfg, err := config.New(*configFile)
	if err != nil {
		log.Println("Configuration failed:", err)
		os.Exit(1)
	}

	credentials := aws.DetectCreds(*accessId, *secretKey, "")
	service := ec2.New(
		&aws.Config{
			Region:      *region,
			Credentials: credentials,
		},
	)

	params := &ec2.TerminateInstancesInput{
		DryRun: aws.Boolean(*dryRun),
	}

	group := make(candidates.Group)

	reservations := make([]*ec2.Reservation, 0)
	ch := collector.Dispatch(credentials, regions)

	for result := range ch {
		reservations = append(reservations, result...)
	}

	log.Println(awsutil.StringValue(reservations))

	// []reservation -> []instances ->
	//		PublicIpAddress, PrivateIpAddress
	//		[]*tag -> Key, Value
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if *tag.Key == *groupTag && *instance.State.Name == "running" {
					if _, ok := cfg[*tag.Value]; ok {
						info := candidates.Candidate{
							ID:        *instance.InstanceID,
							CreatedAt: *instance.LaunchTime,
						}
						group[*tag.Value] = append(group[*tag.Value], info)
					}
				}
			}
		}
	}

	for tag, hosts := range group {
		oldies := hosts.OlderThan(cfg[tag].Age)

		// oldest instance first
		sort.Sort(oldies)
		sort.Reverse(oldies)

		if cfg[tag].Count >= len(oldies) {
			log.Fatalf("Refusing to terminate all instances from group %s.", tag)
			return
		} else {
			for _, oldie := range oldies[:cfg[tag].Count] {
				log.Printf("Instance %s from %s selected for termination.\n", oldie.ID, tag)
				params.InstanceIDs = append(params.InstanceIDs, aws.String(oldie.ID))
			}
		}
	}

	resp, err := service.TerminateInstances(params)
	if err != nil {
		log.Println("ERROR:", err)
	}
	log.Println(awsutil.StringValue(resp))
}
