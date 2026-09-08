package internal

import "testing"

func TestScanFindsUnidentifiedWrongMatchMissingAndDuplicate(t *testing.T) {
	got := scanLibrary([]LibraryItem{
		{ID: "1", Path: "/lib/foo.mkv"},
		{ID: "2", Title: "Arrival", FileName: "Mad.Max.Fury.Road.2015.mkv", Path: "/lib/madmax.mkv", TMDBID: 76341},
		{ID: "3", Title: "Show", SeriesID: "s1", Episode: 1, Expected: 3},
		{ID: "4", Title: "Show", SeriesID: "s1", Episode: 3, Expected: 3},
		{ID: "5", Title: "Dup A", Path: "/lib/same.mkv", TMDBID: 1},
		{ID: "6", Title: "Dup B", Path: "/lib/same.mkv", TMDBID: 2},
	})
	kinds := map[string]int{}
	for _, f := range got {
		kinds[f.Kind]++
	}
	if kinds[FindingUnidentified] < 1 || kinds[FindingWrongMatch] < 1 || kinds[FindingMissingEp] < 1 || kinds[FindingDuplicate] < 1 {
		t.Fatalf("kinds = %#v findings=%#v", kinds, got)
	}
}

func TestMatchingFilenameIsNotWrongMatch(t *testing.T) {
	got := scanLibrary([]LibraryItem{
		{ID: "1", Title: "Arrival", FileName: "Arrival.2016.mkv", TMDBID: 329865},
	})
	for _, f := range got {
		if f.Kind == FindingWrongMatch {
			t.Fatalf("false positive: %#v", f)
		}
	}
}
