package internal

import (
	"path/filepath"
	"strings"
)

const (
	FindingUnidentified = "unidentified_file"
	FindingWrongMatch   = "wrong_match"
	FindingMissingEp    = "missing_episode"
	FindingDuplicate    = "duplicate"
)

// LibraryItem is a catalog row the librarian inspects.
type LibraryItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	FileName  string `json:"file_name,omitempty"`
	Path      string `json:"path,omitempty"`
	TMDBID    int    `json:"tmdb_id,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	SeriesID  string `json:"series_id,omitempty"`
	Season    int    `json:"season,omitempty"`
	Episode   int    `json:"episode,omitempty"`
	Expected  int    `json:"expected_episodes,omitempty"`
}

// Finding is one QA issue.
type Finding struct {
	Kind    string `json:"kind"`
	ItemID  string `json:"item_id,omitempty"`
	Title   string `json:"title,omitempty"`
	Path    string `json:"path,omitempty"`
	Summary string `json:"summary"`
	Repair  string `json:"repair,omitempty"`
}

func scanLibrary(items []LibraryItem) []Finding {
	var out []Finding
	byPath := map[string][]LibraryItem{}
	bySeries := map[string][]LibraryItem{}
	for _, it := range items {
		if it.TMDBID == 0 && strings.TrimSpace(it.Title) == "" {
			out = append(out, Finding{
				Kind: FindingUnidentified, ItemID: it.ID, Path: it.Path,
				Summary: "file has no title or TMDB id", Repair: "identify_file",
			})
		}
		if it.Title != "" && it.FileName != "" && likelyWrongMatch(it.Title, it.FileName) {
			out = append(out, Finding{
				Kind: FindingWrongMatch, ItemID: it.ID, Title: it.Title, Path: it.Path,
				Summary: "filename does not resemble the catalog title", Repair: "rematch_metadata",
			})
		}
		key := it.Path
		if key == "" {
			key = it.FileName
		}
		if key != "" {
			byPath[key] = append(byPath[key], it)
		}
		if it.SeriesID != "" {
			bySeries[it.SeriesID] = append(bySeries[it.SeriesID], it)
		}
	}
	for path, group := range byPath {
		if len(group) > 1 {
			out = append(out, Finding{
				Kind: FindingDuplicate, ItemID: group[0].ID, Path: path,
				Summary: "multiple library rows share the same file path", Repair: "merge_duplicates",
			})
		}
	}
	for sid, eps := range bySeries {
		expected := 0
		have := map[int]bool{}
		title := sid
		for _, e := range eps {
			if e.Expected > expected {
				expected = e.Expected
			}
			if e.Episode > 0 {
				have[e.Episode] = true
			}
			if e.Title != "" {
				title = e.Title
			}
		}
		if expected == 0 {
			continue
		}
		for n := 1; n <= expected; n++ {
			if !have[n] {
				out = append(out, Finding{
					Kind: FindingMissingEp, Title: title,
					Summary: "missing episode in series " + sid, Repair: "request_episode",
				})
			}
		}
	}
	return out
}

func likelyWrongMatch(title, fileName string) bool {
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName)))
	base = strings.ReplaceAll(base, ".", " ")
	base = strings.ReplaceAll(base, "_", " ")
	t := strings.ToLower(title)
	words := strings.Fields(t)
	hits := 0
	for _, w := range words {
		if len(w) < 3 {
			continue
		}
		if strings.Contains(base, w) {
			hits++
		}
	}
	significant := 0
	for _, w := range words {
		if len(w) >= 3 {
			significant++
		}
	}
	if significant == 0 {
		return false
	}
	return hits == 0
}
