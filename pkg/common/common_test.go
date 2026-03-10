package common_test

import (
	"fmt"
	"go-fiber-stater-kit/pkg/common"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestTrimToLower(t *testing.T) {
	fixedRate := "Fixed Rate"

	actual := common.TrimToLower(fixedRate)

	if actual != "fixedrate" {
		t.Error("Error", actual)
	}
}

func TestIsFloat32(t *testing.T) {
	f := reflect.TypeOf(float64(3.14))

	actual := common.IsFloat(f)

	if !actual {
		t.Error("Is not float")
	}
}

func TestIsFloat64(t *testing.T) {
	f := reflect.TypeOf(float32(3.14))

	actual := common.IsFloat(f)

	if !actual {
		t.Error("Is not float")
	}
}

func TestF64ToString(t *testing.T) {
	f := 11.5200000000186265332323434343545
	expected := "11.5200000000186265"

	actual := common.F64ToString(f)

	if actual != expected {
		t.Error("Convert error", actual)
	}
}

func TestF64ToStringDyn(t *testing.T) {
	f := 11.5200186265332323434343545325657697832
	expected := "11.520000"

	actual := common.F64ToStringDyn(f)

	if actual != expected {
		t.Error("Convert error", actual)
	}
}

func TestParseNumEToNumber(t *testing.T) {
	// define a float with a large number of digits
	f := 123456789.12345678987876543245754255657

	// convert float to string using FormatFloat
	s := strconv.FormatFloat(f, 'f', -1, 64)

	// print the resulting string
	fmt.Println(s)
}

func TestTrimSpace(t *testing.T) {
	text := " Hello "

	actual := strings.TrimSpace(text)

	if actual != "Hello" {
		t.Error("Cannot trim space", actual)
	}
}
