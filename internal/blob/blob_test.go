package blob

import "testing"

func TestPackedBlobDataLoads(t *testing.T) {
	if err := ensureBlobs(); err != nil {
		t.Fatal(err)
	}
	if len(blobs) == 0 {
		t.Fatal("packed blob data contained no entries")
	}
}
