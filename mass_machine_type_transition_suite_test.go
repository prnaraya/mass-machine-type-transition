package mass_machine_type_transition_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMassMachineTypeTransition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MassMachineTypeTransition Suite")
}
