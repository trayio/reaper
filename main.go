package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/trayio/reaper/candidates"
	"github.com/trayio/reaper/collector"
	"github.com/trayio/reaper/config"

	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/aws"
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
	instanceTag := flag.String("tag", "group", "Tag name to group instances by")
	configFile := flag.String("c", "conf.js", "Configuration file.")

	accessId := flag.String("access", "", "AWS access ID")
	secretKey := flag.String("secret", "", "AWS secret key")
	flag.Parse()

	c, err := config.New(*configFile)
	if err != nil {
		fmt.Println("Configuration failed:", err)
		os.Exit(1)
	}

	groups := make(map[string]candidates.Candidates)

	credentials := aws.DetectCreds(*accessId, *secretKey, "")

	ch := collector.Dispatch(credentials, regions)

	reservations := []*ec2.Reservation{}

	for result := range ch {
		reservations = append(reservations, result...)
	}

	// []reservation -> []instances ->
	//		PublicIpAddress, PrivateIpAddress
	//		[]*tag -> Key, Value
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				/*
					Instance state codes:
					0 - pending
					16 - running
					32 - shutting down
					64 - stopping
					80 - stopped
					https://godoc.org/github.com/awslabs/aws-sdk-go/service/ec2#InstanceState
				*/
				if *tag.Key == *instanceTag && *instance.State.Code == 16 {
					if _, ok := c[*tag.Value]; ok {
						info := candidates.Candidate{
							ID:        *instance.InstanceID,
							CreatedAt: *instance.LaunchTime,
						}
						groups[*tag.Value] = append(groups[*tag.Value], info)
					}
				}
			}
		}
	}

	for group, hosts := range groups {
		fmt.Println(group)
		oldies := hosts.OlderThan(c[group].Age)
		sort.Sort(oldies)

		if c[group].Count >= len(oldies) {
			for _, oldie := range oldies {
				fmt.Println(oldie.ID, oldie.CreatedAt)
			}
		} else {
			for _, oldie := range oldies[:c[group].Count] {
				fmt.Println(oldie.ID, oldie.CreatedAt)
			}
		}
	}
}
