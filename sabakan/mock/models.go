// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package mock

import (
	"fmt"
	"io"
	"strconv"
)

// BMC represents a Baseboard Management Controller.
type Bmc struct {
	BmcType string `json:"bmcType"`
	Ipv4    string `json:"ipv4"`
}

// Label represents an arbitrary key-value pairs.
type Label struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LabelInput represents a label to search machines.
type LabelInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Machine represents a physical server in a datacenter rack.
type Machine struct {
	Spec   *MachineSpec   `json:"spec"`
	Status *MachineStatus `json:"status"`
}

// MachineParams is a set of input parameters to search machines.
type MachineParams struct {
	Labels              []*LabelInput  `json:"labels,omitempty"`
	Racks               []int          `json:"racks,omitempty"`
	Roles               []string       `json:"roles,omitempty"`
	States              []MachineState `json:"states,omitempty"`
	MinDaysBeforeRetire *int           `json:"minDaysBeforeRetire,omitempty"`
}

// MachineSpec represents specifications of a machine.
type MachineSpec struct {
	Serial       string   `json:"serial"`
	Labels       []*Label `json:"labels,omitempty"`
	Rack         int      `json:"rack"`
	IndexInRack  int      `json:"indexInRack"`
	Role         string   `json:"role"`
	Ipv4         []string `json:"ipv4"`
	RegisterDate string   `json:"registerDate"`
	RetireDate   string   `json:"retireDate"`
	Bmc          *Bmc     `json:"bmc"`
}

// MachineStatus represents status of a Machine.
type MachineStatus struct {
	State     MachineState `json:"state"`
	Timestamp string       `json:"timestamp"`
	Duration  float64      `json:"duration"`
}

type Query struct {
}

// MachineState enumerates machine states.
type MachineState string

const (
	MachineStateUninitialized MachineState = "UNINITIALIZED"
	MachineStateHealthy       MachineState = "HEALTHY"
	MachineStateUnhealthy     MachineState = "UNHEALTHY"
	MachineStateUnreachable   MachineState = "UNREACHABLE"
	MachineStateUpdating      MachineState = "UPDATING"
	MachineStateRetiring      MachineState = "RETIRING"
	MachineStateRetired       MachineState = "RETIRED"
)

var AllMachineState = []MachineState{
	MachineStateUninitialized,
	MachineStateHealthy,
	MachineStateUnhealthy,
	MachineStateUnreachable,
	MachineStateUpdating,
	MachineStateRetiring,
	MachineStateRetired,
}

func (e MachineState) IsValid() bool {
	switch e {
	case MachineStateUninitialized, MachineStateHealthy, MachineStateUnhealthy, MachineStateUnreachable, MachineStateUpdating, MachineStateRetiring, MachineStateRetired:
		return true
	}
	return false
}

func (e MachineState) String() string {
	return string(e)
}

func (e *MachineState) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = MachineState(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid MachineState", str)
	}
	return nil
}

func (e MachineState) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
