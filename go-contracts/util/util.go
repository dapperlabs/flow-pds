package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"text/template"

	"fmt"
	"reflect"

	"github.com/bjartek/overflow/overflow"
	"github.com/onflow/cadence"
	flow "github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

const flowPath = "../flow.json"

var FlowJSON []string = []string{flowPath}

type Addresses struct {
	NonFungibleToken      string
	ExampleNFT            string
	PackNFT               string
	IPackNFT              string
	PDS                   string
	PackNFTName           string
	PackNFTAddress        string
	CollectibleNFTName    string
	CollectibleNFTAddress string
	FungibleToken         string
	MetadataViews         string
	RoyaltyAddress        string
}

type TestEvent struct {
	Name   string
	Fields map[string]interface{}
}

var addresses Addresses

func ParseCadenceTemplate(templatePath string) []byte {
	fb, err := ioutil.ReadFile(templatePath)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("Template").Parse(string(fb))
	if err != nil {
		panic(err)
	}

	// Addresss for emulator are
	// addresses = Addresses{"f8d6e0586b0a20c7", "01cf0e2f2f715450", "01cf0e2f2f715450", "f3fcd2c1a78f5eee", "f3fcd2c1a78f5eee"}
	addresses = Addresses{
		NonFungibleToken:      os.Getenv("NON_FUNGIBLE_TOKEN_ADDRESS"),
		ExampleNFT:            os.Getenv("EXAMPLE_NFT_ADDRESS"),
		PackNFT:               os.Getenv("PACKNFT_ADDRESS"),
		IPackNFT:              os.Getenv("PDS_ADDRESS"),
		PDS:                   os.Getenv("PDS_ADDRESS"),
		PackNFTName:           "PackNFT",
		PackNFTAddress:        os.Getenv("PACKNFT_ADDRESS"),
		CollectibleNFTName:    "ExampleNFT",
		CollectibleNFTAddress: os.Getenv("EXAMPLE_NFT_ADDRESS"),
		FungibleToken:         os.Getenv("FUNGIBLE_TOKEN_ADDRESS"),
		MetadataViews:         os.Getenv("METADATA_VIEWS_ADDRESS"),
		RoyaltyAddress:        os.Getenv("PACKNFT_ADDRESS"),
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, addresses)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func ParseTestEvents(events []flow.Event) (formatedEvents []*overflow.FormatedEvent) {
	for _, e := range events {
		formatedEvents = append(formatedEvents, overflow.ParseEvent(e, uint64(0), time.Now(), nil))
	}
	return
}

func NewExpectedPackNFTEvent(name string) TestEvent {
	return TestEvent{
		Name:   "A." + addresses.PackNFT + ".PackNFT." + name,
		Fields: make(map[string]interface{}),
	}
}

func NewExpectedPDSEvent(name string) TestEvent {
	return TestEvent{
		Name:   "A." + addresses.PDS + ".PDS." + name,
		Fields: make(map[string]interface{}),
	}
}

func (te TestEvent) AddField(fieldName string, fieldValue interface{}) TestEvent {
	te.Fields[fieldName] = fieldValue
	return te
}

func (te TestEvent) AssertHasKey(t *testing.T, event *overflow.FormatedEvent, key string) {
	assert.Equal(t, te.Name, event.Name)
	_, exist := event.Fields[key]
	assert.Equal(t, true, exist)
}

func (te TestEvent) AssertEqual(t *testing.T, event *overflow.FormatedEvent) {
	assert.Equal(t, event.Name, te.Name)
	assert.Equal(t, len(te.Fields), len(event.Fields))
	for k := range te.Fields {
		v := reflect.ValueOf(event.Fields[k])
		switch v.Kind() {
		case reflect.String:
			assert.Equal(t, te.Fields[k], event.Fields[k])
		case reflect.Slice:
			assert.Equal(t, len(te.Fields[k].([]interface{})), v.Len())
			for i := 0; i < v.Len(); i++ {
				// This is the special case we are addressing
				u := te.Fields[k].([]interface{})[i].(uint64)
				assert.Equal(t, strconv.FormatUint(u, 10), v.Interface().([]interface{})[i])
				i++
			}
		case reflect.Map:
			fmt.Printf("map: %v\n", v.Interface())
		case reflect.Chan:
			fmt.Printf("chan %v\n", v.Interface())
		default:
			fmt.Println("Unsupported types")
		}
	}
}

// Gets the address in the format of a hex string from an account name
func GetAccountAddr(g *overflow.Overflow, name string) string {
	address := g.Account(name).Address().String()
	zeroPrefix := "0"
	if string(address[0]) == zeroPrefix {
		address = address[1:]
	}
	return "0x" + address
}

func ReadCadenceCode(ContractPath string) []byte {
	b, err := ioutil.ReadFile(ContractPath)
	if err != nil {
		panic(err)
	}
	return b
}

func GetHash(g *overflow.Overflow, toHash string) (result string, err error) {
	filename := "../cadence-scripts/packNFT/checksum.cdc"
	script := ParseCadenceTemplate(filename)
	r, err := g.Script(string(script)).Args(g.Arguments().String(toHash)).RunReturns()
	result = r.ToGoValue().(string)
	return
}

func GetName(g *overflow.Overflow) (result string, err error) {
	filename := "../../../scripts/contract/get_name.cdc"
	script := ParseCadenceTemplate(filename)
	r, err := g.Script(string(script)).RunReturns()
	result = r.ToGoValue().(string)
	return
}

func GetVersion(g *overflow.Overflow) (result string, err error) {
	filename := "../../../scripts/contract/get_version.cdc"
	script := ParseCadenceTemplate(filename)
	r, err := g.Script(string(script)).RunReturns()
	result = r.ToGoValue().(string)
	return
}

func ConvertCadenceByteArray(a cadence.Value) (b []uint8) {
	// type assertion of interface
	i := a.ToGoValue().([]interface{})

	for _, e := range i {
		// type assertion of uint8
		b = append(b, e.(uint8))
	}
	return

}

func ConvertCadenceStringArray(a cadence.Value) (b []string) {
	// type assertion of interface
	i := a.ToGoValue().([]interface{})

	for _, e := range i {
		b = append(b, e.(string))
	}
	return
}

// Multisig utility functions and type

// Arguement for Multisig functions `Multisig_SignAndSubmit`
// This allows for generic functions to type cast the arguments into
// correct cadence types.
// i.e. for a cadence.UFix64, Arg {V: "12.00", T: "UFix64"}
type Arg struct {
	V interface{}
	T string
}
