// Copyright 2020 Mike Helmick
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mikehelmick/go-chaff"
)

func randInt(s int) (int64, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(s)))
	if err != nil {
		return 0, err
	}
	return r.Int64(), nil
}

func main() {
	r := mux.NewRouter()

	track := chaff.New()
	defer track.Close()

	{
		// Create a submodule
		sub := r.PathPrefix("").Subrouter()
		// Install the chaff tracker middleware.
		sub.Use(track.Track)
		sub.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sleep, err := randInt(1000)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("error generating random data: %v", err)))
				return
			}
			sleep += 100
			time.Sleep(time.Duration(sleep) * time.Millisecond)

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Slept %v ms, some data below\n", sleep)
			w.Write([]byte(strings.Repeat("a", int(sleep))))
		})).Methods("GET")
	}

	{
		sub := r.PathPrefix("/chaff").Subrouter()
		// The tracker itself is an HTTP handler, so just install on the chaff path.
		sub.Handle("", track).Methods("GET")
	}

	srv := &http.Server{
		Handler: handlers.CombinedLoggingHandler(os.Stdout, r),
		Addr:    "0.0.0.0:8080",
	}
	log.Printf("Listening on :%v", 8080)
	log.Fatal(srv.ListenAndServe())
}
