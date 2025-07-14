package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("🧪 Multi-vendor Order System Test Suite")
	fmt.Println("=========================================")

	// Run only order-related tests
	tests := []string{
		"TestOrder",
		"TestSubOrder", 
		"TestOrderItem",
		"TestMultiVendorOrderScenario",
		"TestOrderNumberGeneration",
		"TestSubOrderStatusTransitions",
		"TestOrderItemCalculations",
		"TestDeliveryLocationValidation",
	}

	for _, test := range tests {
		fmt.Printf("\n🔍 Running test: %s\n", test)
		fmt.Println("-------------------")
		
		cmd := exec.Command("go", "test", "./internal/model", "-run", test, "-v")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		err := cmd.Run()
		if err != nil {
			fmt.Printf("❌ Test %s failed: %v\n", test, err)
		} else {
			fmt.Printf("✅ Test %s passed!\n", test)
		}
	}
	
	fmt.Println("\n🎯 Test Summary Complete!")
}
