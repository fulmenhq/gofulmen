package foundry

// HTTPStatusCode represents an individual HTTP status code with its reason phrase.
type HTTPStatusCode struct {
	// Value is the HTTP status code (e.g., 200, 404).
	Value int

	// Reason is the HTTP reason phrase (e.g., "OK", "Not Found").
	Reason string
}

// HTTPStatusGroup represents a group of related HTTP status codes.
//
// Groups categorize HTTP status codes by their first digit:
//   - informational: 1xx codes (provisional responses)
//   - success: 2xx codes (successful requests)
//   - redirect: 3xx codes (redirection)
//   - client-error: 4xx codes (client errors)
//   - server-error: 5xx codes (server errors)
type HTTPStatusGroup struct {
	// ID is the unique group identifier (e.g., "success", "client-error").
	ID string

	// Name is the human-readable group name.
	Name string

	// Description provides documentation for this group.
	Description string

	// Codes contains the HTTP status codes in this group.
	Codes []HTTPStatusCode
}

// Contains checks if the given status code is in this group.
//
// Example:
//
//	group := &HTTPStatusGroup{
//	    ID: "success",
//	    Codes: []HTTPStatusCode{{Value: 200, Reason: "OK"}},
//	}
//	if group.Contains(200) {
//	    // Status code is in success group
//	}
func (g *HTTPStatusGroup) Contains(statusCode int) bool {
	for _, code := range g.Codes {
		if code.Value == statusCode {
			return true
		}
	}
	return false
}

// GetReason returns the reason phrase for the given status code.
//
// Returns an empty string if the status code is not found in this group.
//
// Example:
//
//	reason := group.GetReason(200) // Returns "OK"
func (g *HTTPStatusGroup) GetReason(statusCode int) string {
	for _, code := range g.Codes {
		if code.Value == statusCode {
			return code.Reason
		}
	}
	return ""
}

// HTTPStatusHelper provides convenient HTTP status code checking utilities.
//
// This helper wraps a catalog's HTTP status groups to provide simple
// boolean checks for status code ranges.
type HTTPStatusHelper struct {
	groups        map[string]*HTTPStatusGroup
	codeToGroupID map[int]string
}

// NewHTTPStatusHelper creates a new HTTP status helper from a catalog.
//
// The helper builds lookup maps for efficient status code checks.
func NewHTTPStatusHelper(groups []*HTTPStatusGroup) *HTTPStatusHelper {
	helper := &HTTPStatusHelper{
		groups:        make(map[string]*HTTPStatusGroup),
		codeToGroupID: make(map[int]string),
	}

	for _, group := range groups {
		helper.groups[group.ID] = group
		for _, code := range group.Codes {
			helper.codeToGroupID[code.Value] = group.ID
		}
	}

	return helper
}

// IsInformational checks if the status code is informational (1xx).
//
// Example:
//
//	if helper.IsInformational(100) {
//	    // Status is informational
//	}
func (h *HTTPStatusHelper) IsInformational(statusCode int) bool {
	return h.isInGroup(statusCode, "informational")
}

// IsSuccess checks if the status code indicates success (2xx).
//
// Example:
//
//	if helper.IsSuccess(200) {
//	    // Request was successful
//	}
func (h *HTTPStatusHelper) IsSuccess(statusCode int) bool {
	return h.isInGroup(statusCode, "success")
}

// IsRedirect checks if the status code indicates redirection (3xx).
//
// Example:
//
//	if helper.IsRedirect(301) {
//	    // Resource has moved
//	}
func (h *HTTPStatusHelper) IsRedirect(statusCode int) bool {
	return h.isInGroup(statusCode, "redirect")
}

// IsClientError checks if the status code indicates a client error (4xx).
//
// Example:
//
//	if helper.IsClientError(404) {
//	    // Client made a bad request
//	}
func (h *HTTPStatusHelper) IsClientError(statusCode int) bool {
	return h.isInGroup(statusCode, "client-error")
}

// IsServerError checks if the status code indicates a server error (5xx).
//
// Example:
//
//	if helper.IsServerError(500) {
//	    // Server encountered an error
//	}
func (h *HTTPStatusHelper) IsServerError(statusCode int) bool {
	return h.isInGroup(statusCode, "server-error")
}

// GetReasonPhrase returns the reason phrase for the given status code.
//
// Returns an empty string if the status code is not recognized.
//
// Example:
//
//	reason := helper.GetReasonPhrase(200) // Returns "OK"
//	reason := helper.GetReasonPhrase(404) // Returns "Not Found"
func (h *HTTPStatusHelper) GetReasonPhrase(statusCode int) string {
	groupID, exists := h.codeToGroupID[statusCode]
	if !exists {
		return ""
	}

	group, exists := h.groups[groupID]
	if !exists {
		return ""
	}

	return group.GetReason(statusCode)
}

// GetGroup returns the HTTP status group for the given status code.
//
// Returns nil if the status code is not recognized.
//
// Example:
//
//	group := helper.GetGroup(200)
//	if group != nil {
//	    // group.ID == "success"
//	}
func (h *HTTPStatusHelper) GetGroup(statusCode int) *HTTPStatusGroup {
	groupID, exists := h.codeToGroupID[statusCode]
	if !exists {
		return nil
	}
	return h.groups[groupID]
}

// isInGroup checks if a status code belongs to a specific group.
func (h *HTTPStatusHelper) isInGroup(statusCode int, groupID string) bool {
	actualGroupID, exists := h.codeToGroupID[statusCode]
	if !exists {
		return false
	}
	return actualGroupID == groupID
}
