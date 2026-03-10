package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-fiber-stater-kit/pkg/timex"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

func TrimLower(text string) string {
	return strings.TrimSpace(strings.ToLower(text))
}

func F64ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 64)
}

func I64ToString(num int64) string {
	return strconv.FormatInt(num, 64)
}

func F64ToStringDyn(num float64) string {
	return fmt.Sprintf("%v", num)
}

func IsFloat(t reflect.Type) bool {
	if t.Kind() == reflect.Float32 || t == reflect.TypeOf(float64(0)) {
		return true
	}
	return false
}

func IsInt(t reflect.Type) bool {
	if t.Kind() == reflect.Int32 || t.Kind() == reflect.Int64 {
		return true
	}
	return false
}

func Rand(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func IsEmpty(val string) bool {
	if val == "" {
		return true
	} else {
		return false
	}
}

func IsDigit(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func RemoveExtension(filename string) string {
	spl := strings.Split(filename, ".")
	if len(spl) > 1 {
		return spl[0]
	}
	return filename
}

func IsDevelopment() bool {
	return os.Getenv("ENV") == "development" || os.Getenv("ENV") == ""
}

func isNil(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}

func Find(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func FindInt(a []int64, x int64) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func ParseInt(num string) (int, error) {
	return strconv.Atoi(num)
}
func ParseInt64(num string) (int64, error) {
	return strconv.ParseInt(num, 10, 64)
}

func ParseInt64Safety(num string) int64 {
	vi, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return 0
	}
	return vi
}

func ParseFloat64(num string) (float64, error) {
	return strconv.ParseFloat(num, 64)
}

func ParseFloat64Safety(num string) float64 {
	vf, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0
	}
	return vf
}

func Contains(path string, contain ...string) bool {
	found := 0
	for _, c := range contain {
		if strings.Contains(path, c) {
			found++
		}
	}
	return found > 0
}

func IsExpired(jwt string, key string) bool {
	token := strings.Split(jwt, ".")
	payload, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(token[1])
	if err != nil {
		return true
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal(payload, &m); err != nil {
		return true
	}

	if timestamp, ok := m[key].(float64); ok {
		current := timex.Now().Unix()
		if current < int64(timestamp) {
			return false
		}
	}
	return true
}

func ToFloat64(value interface{}) float64 {
	if value == nil {
		return 0.0
	}
	return value.(float64)
}

func RemoveIndex(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}

func HasCode(id string, index int) string {
	return fmt.Sprintf("%sX%d", id, index)
}

func Base64FromByteToByte(b []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(b))
}

func Base64FromByte(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Base64FromString(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
func FloatDecimal(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func TrimToLower(text string) string {
	return strings.ToLower(strings.ReplaceAll(text, " ", ""))
}
func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func IsIn[T comparable](arr []T, val T) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}

func StructToJsonRaw(input any) (json.RawMessage, error) {
	// Convert the struct to JSON bytes
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	// Convert the JSON bytes to json.RawMessage
	var rawJSON json.RawMessage = jsonBytes
	return rawJSON, nil
}

func GetUniqueStoreIds(storeId, storeIdFromStoreGroup string) []string {
	storeIdConcat := ConcatStoreIds(storeId, storeIdFromStoreGroup)
	rs := strings.Split(storeIdConcat, ",")
	storeIdList := []string{}

	// Use a map to track duplicates
	storeIdMap := make(map[string]bool)

	for _, storeIdData := range rs {
		if storeIdData == "" { // Skip empty strings
			continue
		}
		if !storeIdMap[storeIdData] {
			storeIdMap[storeIdData] = true
			storeIdList = append(storeIdList, storeIdData)
		}
	}

	// At this point, storeIdList contains unique store IDs
	return storeIdList
}

func ConcatStoreIds(storeId, storeIdFromStoreGroup string) string {
	if storeId == "" {
		return storeIdFromStoreGroup
	}
	if storeIdFromStoreGroup == "" {
		return storeId
	}
	return storeId + "," + storeIdFromStoreGroup
}
