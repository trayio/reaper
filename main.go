package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/trayio/reaper/candidates"
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

func getReservations(cfg *aws.Config, result chan []*ec2.Reservation, wg *sync.WaitGroup) {
	defer wg.Done()

	reservations := []*ec2.Reservation{}
	service := ec2.New(cfg)

	describeInstancesOutput, err := service.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error in region", cfg.Region, ":", err)
		return
	}
	reservations = describeInstancesOutput.Reservations

	// not empty if response is not paged as per docs, but a null pointer
	// https://godoc.org/github.com/awslabs/aws-sdk-go/service/ec2#DescribeInstancesOutput
	for describeInstancesOutput.NextToken != nil {
		describeInstancesOutput, err = service.DescribeInstances(
			&ec2.DescribeInstancesInput{
				NextToken: describeInstancesOutput.NextToken,
			},
		)
		if err != nil {
			fmt.Println("Error in region", cfg.Region, ":", err)
			return
		}
		reservations = append(reservations, describeInstancesOutput.Reservations...)
	}
	result <- reservations
}

func dispatcher(regions []string) chan []*ec2.Reservation {
	var wg sync.WaitGroup

	ch := make(chan []*ec2.Reservation)
	go func() {
		for _, region := range regions {
			cfg := &aws.Config{
				Region:      region,
				Credentials: aws.DetectCreds("", "", ""),
			}
			wg.Add(1)
			go getReservations(cfg, ch, &wg)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

func main() {
	instanceTag := flag.String("tag", "group", "Tag name to group instances by")
	configFile := flag.String("c", "conf.js", "Configuration file.")
	flag.Parse()

	c, err := config.New(*configFile)
	if err != nil {
		fmt.Println("Configuration failed:", err)
		os.Exit(1)
	}

	groups := make(map[string]candidates.Candidates)

	ch := dispatcher(regions)

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
