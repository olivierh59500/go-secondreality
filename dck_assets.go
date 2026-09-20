package srdata

import "embed"

// DCKAssetPackedFiles shares an embedded resource with the optional DCK version.
func DCKAssetPackedFiles() embed.FS { return packedFiles }
