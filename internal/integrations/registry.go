package integrations

func IsSupportedTool(toolID string) bool {
	_, ok := GetAdapter(toolID)
	return ok
}
