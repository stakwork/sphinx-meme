package lsat

// storing service caveat information related to services meme server
// cares about (services, capabilities, and constraints)

const (
	// prefix to be attached to any caveats that are service specific
	// e.g. sphinx_meme_timeout
	MemeServerServicePrefix	= "sphinx_meme"
	// CondCapabilitiesSuffix is the condition suffix used for a service's
	// capabilities caveat. For example, the condition of a capabilities
	// caveat for a service named `loop` would be `loop_capabilities`.
	CondCapabilitiesSuffix = "_capabilities"
	MaxUploadCapability = "large_upload"
	CondMaxUploadConstraintSuffix = "_max_mb"
	TimeoutConstraintSuffix = "_timeout"
)
