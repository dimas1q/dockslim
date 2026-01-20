package analysis

const (
	largeLayerThresholdBytes = 200 * 1024 * 1024
	manyLayersThreshold      = 40
	largeImageThresholdBytes = 1024 * 1024 * 1024
	maxLargestLayers         = 3

	smallLayerThresholdBytes = 1 * 1024 * 1024
	manySmallLayersThreshold = 20

	hugeImageThresholdBytes          = 2 * 1024 * 1024 * 1024
	manyMediumLayerMinBytes          = 10 * 1024 * 1024
	manyMediumLayerMaxBytes          = 50 * 1024 * 1024
	manyMediumLayersThreshold        = 25
	tooFewLayersThreshold            = 5
	hugeFewLayersImageThresholdBytes = 800 * 1024 * 1024
	vendoredLayerThresholdBytes      = 300 * 1024 * 1024
	vendoredLayerMaxCountThreshold   = 15
	cacheCleanupLayerThresholdBytes  = 150 * 1024 * 1024
	rebuildFrequentlyThresholdBytes  = 500 * 1024 * 1024
	pullTimeRiskTotalThresholdBytes  = int64(1.5 * 1024 * 1024 * 1024)
	pullTimeRiskLayerThresholdBytes  = 400 * 1024 * 1024
)
