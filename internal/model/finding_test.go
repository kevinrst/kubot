package model

import "testing"

func TestRankSeverity_unknownSortsLast(t *testing.T) {
	if RankSeverity("bogus") <= RankSeverity(SeverityNote) {
		t.Fatal("unknown severity must sort after note, never as critical")
	}
	if RankSeverity(SeverityCritical) != 0 {
		t.Fatal("critical must rank first")
	}
}
