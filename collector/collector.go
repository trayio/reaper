package collector

import (
	"log"
	"sync"

	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/aws"
	"github.com/trayio/reaper/Godeps/_workspace/src/github.com/awslabs/aws-sdk-go/service/ec2"
)

func reservations(cfg *aws.Config, result chan []*ec2.Reservation, wg *sync.WaitGroup) {
	defer wg.Done()

	reservations := []*ec2.Reservation{}
	service := ec2.New(cfg)

	describeInstancesOutput, err := service.DescribeInstances(nil)
	if err != nil {
		log.Println("Error in region", cfg.Region, ":", err)
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
			log.Println("Error in region", cfg.Region, ":", err)
			return
		}
		reservations = append(reservations, describeInstancesOutput.Reservations...)
	}
	result <- reservations
}

func Dispatch(credentials aws.CredentialsProvider, regions []string) chan []*ec2.Reservation {
	var wg sync.WaitGroup

	ch := make(chan []*ec2.Reservation)
	go func() {
		for _, region := range regions {
			cfg := &aws.Config{
				Region:      region,
				Credentials: credentials,
			}
			wg.Add(1)
			go reservations(cfg, ch, &wg)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}
