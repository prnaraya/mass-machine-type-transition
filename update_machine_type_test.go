package mass_machine_type_transition

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

/*
	updateMachineType() test cases:
		- VM machine type < rhel9.0.0 and not running
		- VM machine type < rhel9.0.0 and running, restartNow=false
		- VM machine type < rhel9.0.0 and running, restartNow=true
		- VM machine type >= rhel9.0.0
		- VM machine type format not pc-q35-rhelx.x.x
		- VM machine type format is default format (q35)
	
	addWarningLabel() test cases:
		- restartNow=false
		- restartNow=true (label will be removed basically as soon as it is applied since it is being restarted immediately)
*/
