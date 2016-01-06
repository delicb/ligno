package ligno_test
import (
	"testing"
	"github.com/delicb/ligno"
	"strings"
//	"time"
)

func TestDefaultLogSimpleFormatter(t *testing.T) {
	memoryHandler := ligno.MemoryHandler(ligno.SimpleFormat())
	ligno.SetHandler(memoryHandler)
	ligno.Log(ligno.INFO, "some message")
	ligno.WaitAll()
	messages := memoryHandler.Messages()
	if len(messages) != 1 {
		t.Errorf("Expected only one message, found %d", len(messages))

	}
	if strings.Index(messages[0], "some message") < 0 {
		t.Errorf("Did not found message 'some message' in output.")
	}
}
