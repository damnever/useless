package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
)

func WhatTheCommits(ctx context.Context, input string) (output string, err error) {
	var in struct {
		Count int `json:"count"`
	}
	if err = json.Unmarshal([]byte(input), &in); err != nil {
		return
	}

	type commitRes struct {
		commit string
		err    error
	}
	commitc := make(chan commitRes)
	cli := &http.Client{}
	const commitsURL = "http://whatthecommit.com/index.txt"

	wg := sync.WaitGroup{}
	for i := 0; i < in.Count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, err0 := http.NewRequest(http.MethodGet, commitsURL, nil)
			var resp *http.Response
			if err0 == nil {
				resp, err0 = cli.Do(req.WithContext(ctx))
			}

			var res commitRes
			if err0 != nil {
				res.err = err0
			} else {
				defer resp.Body.Close()
				if commit, err0 := ioutil.ReadAll(resp.Body); err0 != nil {
					res.err = err0
				} else {
					res.commit = string(commit)
				}
			}

			select {
			case <-ctx.Done():
				return
			case commitc <- res:
			}
		}()
	}
	go func() {
		wg.Wait()
		close(commitc)
	}()

	commits := []string{}
	for commit := range commitc {
		if err = commit.err; err != nil {
			return
		}
		commits = append(commits, commit.commit)
	}
	if len(commits) != in.Count {
		err = ctx.Err() // Context cancelled, maybe check to ensure it.
		return
	}
	var data []byte
	if data, err = json.Marshal(commits); err == nil {
		output = string(data)
	}
	return
}
