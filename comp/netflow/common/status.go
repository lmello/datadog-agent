package common

import (
	"fmt"
)

type FlowListenerData struct {
	FlowType  string
	BindHost  string
	Port      int
	Workers   int
	Namespace string
	FlowCount int
	Error     error
}

type FlowCountUpdate struct {
	Port  int
	Count int
}

var flowDataInstances []*FlowListenerData

func AddFlowDataInstance(flowData FlowListenerData) {
	for _, existingFlowData := range flowDataInstances {
		if existingFlowData.Port == flowData.Port {
			existingFlowData.FlowType = flowData.FlowType
			existingFlowData.BindHost = flowData.BindHost
			existingFlowData.Workers = flowData.Workers
			existingFlowData.Namespace = flowData.Namespace
			return
		}
	}

	// If no existing instance with the same Port is found, add a new instance
	flowDataInstances = append(flowDataInstances, &flowData)
}

func GetAllFlowDataInstances() []*FlowListenerData {
	return flowDataInstances
}

func GetNumListeners() int {
	return len(flowDataInstances)
}

func GetFlowCountByPort(port int) int {
	for _, flowData := range flowDataInstances {
		if flowData.Port == port {
			return flowData.FlowCount
		}
	}
	return 0
}

func UpdateFlowCountByPort(port int, count int) {
	for _, flowData := range flowDataInstances {
		if flowData.Port == port {
			flowData.FlowCount = count
			return
		}
	}
}

func GetTotalFlowCount() int {
	totalCount := 0
	instances := GetAllFlowDataInstances()
	for _, flowData := range instances {
		totalCount += flowData.FlowCount
	}
	return totalCount
}

func UpdateErrorByPort(port int, err error) {
	for _, flowData := range flowDataInstances {
		if flowData.Port == port {
			flowData.Error = err
			fmt.Println("Updated error for port:", port, "Error:", err)
			return
		}
	}
}

func PrintAllFlowDataInstances() {
	fmt.Println("------------------------")
	fmt.Println()
	fmt.Println("Netflow Check Prototype")
	fmt.Println()
	fmt.Println("------------------------")
	instances := GetAllFlowDataInstances()
	fmt.Println()
	numListeners := GetNumListeners()
	if numListeners > 0 {
		fmt.Printf("Listeners Successfully Opened: %d\n", numListeners)
	}
	fmt.Println()

	totalFlowCount := GetTotalFlowCount()
	if totalFlowCount > 0 {
		fmt.Printf("Total Flow Count: %d\n", totalFlowCount)
	}
	fmt.Println()
	fmt.Println()

	for i, flowData := range instances {
		fmt.Printf("Listener %d:\n", i+1)
		fmt.Printf("FlowType: %s\n", flowData.FlowType)
		fmt.Printf("BindHost: %s\n", flowData.BindHost)
		fmt.Printf("Port: %d\n", flowData.Port)
		fmt.Printf("Workers: %d\n", flowData.Workers)
		fmt.Printf("Namespace: %s\n", flowData.Namespace)
		fmt.Printf("Packet Count: %d\n", flowData.FlowCount)
		if flowData.Error != nil {
			fmt.Printf("Error: %s\n", flowData.Error.Error())
		}
		fmt.Println()
	}
}
