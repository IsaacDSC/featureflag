package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IsaacDSC/featureflag/sdk/featureflag"
)

func main() {

	ctx := context.Background()
	ff := featureflag.NewFeatureFlagSDK("http://localhost:3000")

	go func() {
		_, err := ff.Listenner(ctx)
		if err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		isActive := ff.GetFeatureFlag("invalid_ff").WithDefault(true)
		fmt.Println("@@@", isActive)

		isActiveErr, err := ff.GetFeatureFlag("invalid_ff").Err()
		fmt.Println("@@@", isActiveErr, err)

		isActive2, err := ff.GetFeatureFlag("new_name").Err()
		fmt.Println("@@@", isActive2, err)

		isActive3 := ff.GetFeatureFlag("new_name1").Val()
		fmt.Println("@@@", isActive3)

		test1, err := ff.GetFeatureFlag("teste1").Err()
		fmt.Println("@@@ teste1: ", test1, err)

		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
