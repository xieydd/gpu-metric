package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/unisound-ail/atlasctl/util"
	"github.com/unisound-ail/atlasctl/cli"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "List all the training jobs",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
			
			client,_,err := cli.GetCliSetNameSpace()
			util.MustE(err)
			releaseMap, err := util.ListReleaseMap()
			// log.Printf("releaseMap %v", releaseMap)
			util.MustE(err)

			allPods, err = acquireAllPods(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			allJobs, err = acquireAllJobs(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			trainers := NewTrainers(client)
			jobs := []TrainingJob{}
			// for

			for name, ns := range releaseMap {
				supportedChart := false
				for _, trainer := range trainers {
					if trainer.IsSupported(name, ns) {
						job, err := trainer.GetTrainingJob(name, ns)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
						jobs = append(jobs, job)
						supportedChart = true
						break
					}
				}

				if !supportedChart {
					log.Debugf("Unknown chart %s\n", name)
				}

			}

			jobs = makeTrainingJobOrderdByAge(jobs)

			displayTrainingJobList(jobs, false)
		},
	}

	return command
}

func displayTrainingJobList(jobInfoList []TrainingJob, displayGPU bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	var (
		totalAllocatedGPUs int64
		totalRequestedGPUs int64
	)

	if displayGPU {
		fmt.Fprintf(w, "NAME\tSTATUS\tTRAINER\tAGE\tNODE\tGPU(Requests)\tGPU(Allocated)\n")
	} else {
		fmt.Fprintf(w, "NAME\tSTATUS\tTRAINER\tAGE\tNODE\n")
	}

	for _, jobInfo := range jobInfoList {
		status := GetJobRealStatus(jobInfo)
		hostIP := jobInfo.HostIPOfChief()
		if displayGPU {
			requestedGPU := jobInfo.RequestedGPU()
			allocatedGPU := jobInfo.AllocatedGPU()
			// status, hostIP := jobInfo.getStatus()
			totalAllocatedGPUs += allocatedGPU
			totalRequestedGPUs += requestedGPU
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", jobInfo.Name(),
				status,
				strings.ToUpper(jobInfo.Trainer()),
				jobInfo.Age(),
				hostIP,
				strconv.FormatInt(requestedGPU, 10),
				strconv.FormatInt(allocatedGPU, 10))

		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", jobInfo.Name(),
				status,
				strings.ToUpper(jobInfo.Trainer()),
				jobInfo.Age(),
				hostIP)

		}
	}

	if displayGPU {
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Total Allocated GPUs of Training Job:\n")
		fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalAllocatedGPUs, 10))
		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "Total Requested GPUs of Training Job:\n")
		fmt.Fprintf(w, "%s \t\n", strconv.FormatInt(totalRequestedGPUs, 10))
	}

	_ = w.Flush()
}
