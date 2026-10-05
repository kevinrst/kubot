package main

import (
	"sort"
	"strings"

	"github.com/kevinrst/kubot/internal/k8s"
)

func suggestWorkload(snap *k8s.Snapshot, query string) []string {
	seen := map[string]bool{}
	var names []string
	add := func(n string) {
		if n != "" && !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	for _, d := range snap.Deployments {
		add(d.Name)
	}
	for _, s := range snap.Services {
		add(s.Name)
	}
	for owner := range snap.PodsByTopOwner() {
		add(owner)
	}
	var out []string
	for _, n := range names {
		if n != query && nearby(n, query) {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}

func nearby(a, b string) bool {
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}
	return levenshtein(a, b) <= 2
}

func levenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i, ca := range ar {
		cur := make([]int, len(br)+1)
		cur[0] = i + 1
		for j, cb := range br {
			cost := 0
			if ca != cb {
				cost = 1
			}
			cur[j+1] = min(prev[j+1]+1, cur[j]+1, prev[j]+cost)
		}
		prev = cur
	}
	return prev[len(br)]
}
