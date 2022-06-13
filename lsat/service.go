package lsat

// storing service caveat information related to services meme server
// cares about (services, capabilities, and constraints)

const (
	// CondCapabilitiesSuffix is the condition suffix used for a service's
	// capabilities caveat. For example, the condition of a capabilities
	// caveat for a service named `loop` would be `loop_capabilities`.
	CondCapabilitiesSuffix = "_capabilities"
	
	MaxUploadCapability = "large_upload"
	CondMaxUploadConstraintSuffix = "_max_mb"
)