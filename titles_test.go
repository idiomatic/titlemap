package titlemap

import "testing"

func TestQueueSuffix(t *testing.T) {
	title := Title{}
	if title.QueueSuffix() != "" {
		t.Fail()
	}
}
