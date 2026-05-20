// Package entities — ID encoding helpers for qualified hierarchical node IDs.
// See graph.go for the full package-level documentation on node ID format.
package entities

// QualifiedNodeID generates a qualified hierarchical ID for a node.
// - System: returns systemID
// - Container: returns systemID/containerID
// - Component: returns systemID/containerID/componentID
func QualifiedNodeID(nodeType, systemID, containerID, nodeID string) string {
	switch nodeType {
	case "system":
		return systemID
	case "container":
		if systemID == "" || containerID == "" {
			return containerID // fallback for backward compatibility
		}
		return systemID + "/" + containerID
	case "component":
		if systemID == "" || containerID == "" || nodeID == "" {
			return nodeID // fallback for backward compatibility
		}
		return systemID + "/" + containerID + "/" + nodeID
	default:
		return nodeID
	}
}

// ParseQualifiedID parses a qualified ID into its component parts and determines node type.
// Returns the parts slice and the inferred node type.
func ParseQualifiedID(qualifiedID string) (parts []string, nodeType string) {
	if qualifiedID == "" {
		return []string{}, ""
	}

	parts = splitID(qualifiedID)

	switch len(parts) {
	case 1:
		nodeType = "system"
	case 2:
		nodeType = "container"
	case 3:
		nodeType = "component"
	default:
		nodeType = "unknown"
	}

	return parts, nodeType
}

// splitID splits a qualified ID by '/' separator.
func splitID(id string) []string {
	if id == "" {
		return []string{}
	}

	parts := []string{}
	current := ""

	for _, ch := range id {
		if ch == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
