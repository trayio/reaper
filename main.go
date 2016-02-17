package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/trayio/reaper/candidates"
	"github.com/trayio/reaper/collector"
	"github.com/trayio/reaper/config"

	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/ec2"
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

	region := flag.String("region", "us-west-1", "AWS region")
	flag.Parse()

	cfg, err := config.New(*configFile)
	if err != nil {
		log.Println("Configuration failed:", err)
		os.Exit(1)
	}

	fmt.Printf("%#v\n", cfg)
	for key, value := range cfg {
		fmt.Printf("%#v => %#v\n", key, value)
	}
	os.Exit(1)

	credentialsProvider := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.SharedCredentialsProvider{},
			&credentials.EnvProvider{},
			&credentials.EC2RoleProvider{},
		},
	)

	service := ec2.New(
		&aws.Config{
			Region:      *region,
			Credentials: credentialsProvider,
		},
	)

	params := &ec2.TerminateInstancesInput{
		DryRun: aws.Boolean(*dryRun),
	}

	group := make(candidates.Group)

	reservations := make([]*ec2.Reservation, 0)
	ch := collector.Dispatch(credentialsProvider, regions)

	for result := range ch {
		reservations = append(reservations, result...)
	}

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

		if len(oldies) == len(hosts) && cfg[tag].Count >= len(oldies) {
			log.Fatalf("Refusing to terminate all instances in group %s.\n", tag)
		}

		sort.Sort(oldies)

		if len(oldies) > cfg[tag].Count {
			oldies = oldies[:cfg[tag].Count]
		}

		for _, oldie := range oldies {
			log.Printf("Selected for termination: %s from %s.\n", oldie.ID, tag)
			params.InstanceIDs = append(params.InstanceIDs, aws.String(oldie.ID))
		}
	}

	if len(params.InstanceIDs) > 0 {
		resp, err := service.TerminateInstances(params)
		if err != nil {
			log.Println("ERROR:", err)
		}
		log.Println(awsutil.StringValue(resp))
	} else {
		log.Printf("No instances suitable for termination.")
	}
}
