package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/trayio/reaper/candidates"
	"github.com/trayio/reaper/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	groupTag   string
	configFile string
	dryRun     bool
)

// map[group]candidates
func getInstances(region string, groups []string) map[string]candidates.Candidates {
	c := make(map[string]candidates.Candidates)

	svc := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", groupTag)),
				Values: make([]*string, len(groups)),
			},
		},
	}

	for index, group := range groups {
		params.Filters[0].Values[index] = aws.String(group)
	}

	output, err := svc.DescribeInstances(params)
	if err != nil {
		log.Printf("error in region %s: %s\n", region, err)
		return nil
	}

	reservations := output.Reservations
	if len(reservations) == 0 {
		return c
	}

	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name == "running" {
				for _, tag := range instance.Tags {
					if *tag.Key == groupTag {
						if _, ok := c[*tag.Value]; !ok {
							c[*tag.Value] = make(candidates.Candidates, 0)
						}

						candidate := candidates.Candidate{
							ID:        *instance.InstanceId,
							CreatedAt: *instance.LaunchTime,
						}

						c[*tag.Value] = append(c[*tag.Value], candidate)
					}
				}
			}
		}
	}

	return c
}

func main() {
	flag.StringVar(&groupTag, "tag", "group", "Tag name to group instances by")
	flag.StringVar(&configFile, "c", "config.js", "Configuration file.")
	flag.BoolVar(&dryRun, "dry", false, "Enable dry run.")
	flag.Parse()

	cfg, err := config.New(configFile)
	if err != nil {
		log.Println("Configuration failed:", err)
		os.Exit(1)
	}

	// region -> groups map
	rg := make(map[string][]string)

	for group, data := range cfg {
		if _, ok := rg[data.Region]; !ok {
			rg[data.Region] = make([]string, 0)
		}

		rg[data.Region] = append(rg[data.Region], group)
	}

	instances := make(map[string]map[string]candidates.Candidates)
	for region, groups := range rg {
		instances[region] = getInstances(region, groups)
	}

	for region, groups := range instances {
		victims := make(candidates.Candidates, 0)

		for group, hosts := range groups {
			oldies := hosts.OlderThan(cfg[group].Age)

			if len(oldies) == len(hosts) && cfg[group].Count >= len(oldies) {
				log.Printf("Refusing to terminate all instances in group %s.\n", group)
				continue
			}

			sort.Sort(oldies)

			if len(oldies) > cfg[group].Count {
				oldies = oldies[:cfg[group].Count]
			}

			victims = append(victims, oldies...)
		}

		terminateParams := &ec2.TerminateInstancesInput{
			DryRun: aws.Bool(dryRun),
		}

		for _, victim := range victims {
			terminateParams.InstanceIds = append(terminateParams.InstanceIds, aws.String(victim.ID))
		}

		svc := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

		if len(terminateParams.InstanceIds) > 0 {
			resp, err := svc.TerminateInstances(terminateParams)
			if err != nil {
				log.Println("ERROR:", err)
			}
			log.Println(awsutil.StringValue(resp))
		} else {
			log.Printf("No instances suitable for termination\n")
		}
	}
}
