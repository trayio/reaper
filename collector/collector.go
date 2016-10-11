package collector

import (
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func reservations(cfg *aws.Config, result chan []*ec2.Reservation, wg *sync.WaitGroup) {
	defer wg.Done()

	reservations := []*ec2.Reservation{}
	service := ec2.New(session.New(), cfg)

	describeInstancesOutput, err := service.DescribeInstances(nil)
	if err != nil {
		log.Println("Error in region", cfg.Region, ":", err)
		return
	}
	reservations = describeInstancesOutput.Reservations

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

func Dispatch(regions []string) chan []*ec2.Reservation {
	var wg sync.WaitGroup

	ch := make(chan []*ec2.Reservation)
	go func() {
		for _, region := range regions {
			cfg := &aws.Config{
				Region: aws.String(region),
			}
			wg.Add(1)
			go reservations(cfg, ch, &wg)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}
