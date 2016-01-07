package main
import (
	"github.com/delicb/ligno"
	"os"
	"fmt"
)

func main() {
	ligno.SetHandler(ligno.StreamHandler(os.Stdout, ligno.TerminalFormat()))
	err := fmt.Errorf("some error")
	ligno.Error("Something bad happened", "reason", err)
	ligno.WaitAll()
}
