package v1alpha1

type MariaDBClusterStatus struct {
	Phase          string `json:"phase"`
	CurrentVersion string `json:"currentVersion"`
	TargetVersion  string `json:"targetVersion"`
	// TargetStorage
	// Status - resizing ?
}
