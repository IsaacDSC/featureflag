package main

import (
	"fmt"
	"github.com/IsaacDSC/featureflag"
)

func main() {
	ff := featureflag.NewFeatureFlagSDK("http://localhost:3000")
	isActive, err := ff.WithDefault(true).GetFeatureFlag("teste1")
	fmt.Println("@@@", isActive, err)

	isActive2, err := ff.GetFeatureFlag("teste1")
	fmt.Println("@@@", isActive2, err)
}
