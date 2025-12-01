package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/IsaacDSC/featureflag/sdk/contenthub"
)

func main() {

	ctx := context.Background()
	contenthub := contenthub.NewContenthubSDK("http://localhost:3000")

	go func() {
		_, err := contenthub.Listenner(ctx)
		if err != nil {
			panic(err)
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		content := contenthub.Content("invalid_ff").Val()
		fmt.Println("@@@", string(content))

		content1, err := contenthub.Content("invalid_ff").Err()
		fmt.Println("@@@", string(content1), err)

		content2, err := contenthub.Content("homepage_banner").Err()
		fmt.Println("@@@", string(content2), err)

		sessionsID := []string{"promo-black-friday", "new-users-welcome", "seasonal-collection"}
		random := rand.Intn(len(sessionsID))
		sessionID := sessionsID[random]

		content3, err := contenthub.Content("homepage_banner", sessionID).Err()
		fmt.Println("@@@", random, fmt.Sprintf("session_id:%s", sessionID), string(content3), err)

		w.WriteHeader(http.StatusOK)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
