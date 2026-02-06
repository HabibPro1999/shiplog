package session

import "strings"

// FindByQuery searches sessions by query string.
// Priority: exact title match (case-insensitive) -> UUID prefix match -> title substring -> project substring.
// Returns (match, nil) if exactly one match found, or (nil, ambiguous) if multiple matches found.
// Returns (nil, nil) if no match found.
func FindByQuery(sessions []SessionInfo, query string) (*SessionInfo, []SessionInfo) {
	q := strings.ToLower(strings.TrimSpace(query))

	// Exact title match (case-insensitive)
	for i := range sessions {
		if strings.ToLower(sessions[i].Title) == q {
			return &sessions[i], nil
		}
	}

	// UUID prefix match
	for i := range sessions {
		if strings.HasPrefix(strings.ToLower(sessions[i].SessionID), q) {
			return &sessions[i], nil
		}
	}

	// Substring match in title
	var titleMatches []SessionInfo
	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.Title), q) {
			titleMatches = append(titleMatches, s)
		}
	}
	if len(titleMatches) == 1 {
		return &titleMatches[0], nil
	}
	if len(titleMatches) > 1 {
		return nil, titleMatches
	}

	// Substring match in project name
	var projMatches []SessionInfo
	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.Project), q) {
			projMatches = append(projMatches, s)
		}
	}
	if len(projMatches) == 1 {
		return &projMatches[0], nil
	}
	if len(projMatches) > 1 {
		return nil, projMatches
	}

	return nil, nil
}
