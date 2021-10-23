/*
Copyright 2020-2021 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package status

import (
	"fmt"
	"os"

	"github.com/openebs/openebsctl/pkg/client"
	"github.com/openebs/openebsctl/pkg/upgrade"
	batchV1 "k8s.io/api/batch/v1"
)

var WaitFlag bool // For opening wait stream for logs

// Get job with the name -> apply selector to pod
func GetJobStatus() {
	k := client.NewK8sClient()
	namespace := upgrade.OpenebsNs

	// get jiva-upgrade batch jobs
	joblist, err := k.GetBatchJobs(namespace, "cas-type=jiva,name=jiva-upgrade")
	if err != nil {
		fmt.Println("Error getting jiva-upgrade jobs:", err)
		return
	}

	// No jobs found
	if len(joblist.Items) == 0 {
		fmt.Printf("No upgrade-jobs Found in %s namespace", upgrade.OpenebsNs)
	}

	if WaitFlag {
		startLogStream(k, joblist)
	}

	for _, job := range joblist.Items {
		fmt.Println("***************************************")
		fmt.Println("Job Name: ", job.Name)
		getPodLogs(k, job.Name, namespace)
		fmt.Println()
	}
	fmt.Println("***************************************")
}

// Get all the logs from the pods associated with a job
func getPodLogs(k *client.K8sClient, name string, namespace string) {
	// get pods created by the job
	podList, err := k.GetPods(fmt.Sprintf("job-name=%s", name), "", namespace)
	if err != nil {
		fmt.Println("error getting pods of job", name, ": err", err)
		return
	}

	// range over pods to get all the logs
	for _, pod := range podList.Items {
		fmt.Println("From Pod:", pod.Name)
		logs := k.GetPodLogs(pod, namespace)
		if logs == "" {
			fmt.Printf("-> No recent logs from the pod")
			fmt.Println()
			continue
		}
		fmt.Println(logs)
	}

	if len(podList.Items) == 0 {
		fmt.Println("No pods are running for this job")
	}
}

// startLogStream starts opens log stream for a pod
func startLogStream(k *client.K8sClient, jobList *batchV1.JobList) {
	// Stream opens for the first pod in the job
	jobName := jobList.Items[0].Name

	// get pods created by the job
	podList, err := k.GetPods(fmt.Sprintf("job-name=%s", jobName), "", upgrade.OpenebsNs)
	if err != nil {
		fmt.Println("error getting pods of job", jobName, ": err", err)
		return
	}

	// If no pods are running exit silently
	if len(podList.Items) == 0 {
		fmt.Println("No pods are running for the job:", jobName)
		os.Exit(0)
	}

	k.StartPodLogsStream(podList.Items[0], upgrade.OpenebsNs)
}
