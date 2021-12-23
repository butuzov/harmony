package harmony_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/butuzov/harmony"
)

func ExampleFuture_Dogs() {
	// Requests random dogs picture from dog.ceo (dog as service)
	getRandomDogPicture := func(ctx context.Context) string {
		var data struct {
			Message string "json:'message'"
		}

		const API_URL = "https://dog.ceo/api/breeds/image/random"

		if req, err := http.NewRequestWithContext(ctx, http.MethodGet, API_URL, nil); err != nil {
			log.Println(fmt.Errorf("request: %w", err))
			return ""
		} else if res, err := http.DefaultClient.Do(req); err != nil {
			log.Println(fmt.Errorf("request: %w", err))
			return ""
		} else {
			defer res.Body.Close()

			if body, err := ioutil.ReadAll(res.Body); err != nil {
				log.Println(fmt.Errorf("reading body: %w", err))
				return ""
			} else if err := json.Unmarshal(body, &data); err != nil {
				log.Println(fmt.Errorf("unmarshal: %w", err))
				return ""
			}
		}

		return data.Message
	}

	a := harmony.Futute(func() string {
		return getRandomDogPicture(context.Background())
	})
	b := harmony.Futute(func() string {
		return getRandomDogPicture(context.Background())
	})
	fmt.Printf("Rate My Dog: \n\ta) %s\n\tb) %s\n", <-a, <-b)
}

func ExampleFuture() {
	// Requests random dogs picture from dog.ceo (dog as service)

	a := harmony.Futute(func() int { return 1 })
	b := harmony.Futute(func() int { return 0 })
	fmt.Println(<-a, <-b)
	// Output: 1 0
}
