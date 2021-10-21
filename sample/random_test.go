package sample

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestRandomFloat64(t *testing.T) {
	//ret := randomFloat64(2.0, 5.6)
	rets := uuid.New().String()
	fmt.Println(strings.Replace(rets, "-", "", -1))
}
