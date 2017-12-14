package meshexp

import (
	"testing"
	"github.com/gin-gonic/gin/json"
)

func TestLoad(t *testing.T) {
	tree, err := Default()
	if err != nil {
		t.Error(err)
	}

	_, err = json.MarshalIndent(tree, "", "    ")
	if err != nil {
		t.Error(err)
	}
}

func TestExplode(t *testing.T) {
	tree, err := Default()
	if err != nil {
		t.Error(err)
	}

	t.Log(tree.Explode("neuralgia, postherpetic"))
}
