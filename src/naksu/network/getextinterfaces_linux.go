package network

import (
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"strings"

	humanize "github.com/dustin/go-humanize"

	"naksu/constants"
	"naksu/log"
	"naksu/mebroutines"
)

// extNicNixLegendRules is a map between regular expressions matching *nix device names
// and user-friendly legends. This is necessary while we don't call lshw or similair
// to description of the network devices
var extNicNixLegendRules = []struct {
	RegExp string
	Legend string
}{
	{"^w", "Wireless"},
	{"^en", "Ethernet"},
	{"^em", "Ethernet"},
	{"^eth", "Ethernet"},
}

func getExtInterfaceSpeed(extInterface string) uint64 {
	carrierPath := fmt.Sprintf("/sys/class/net/%s/carrier", extInterface)
	speedPath := fmt.Sprintf("/sys/class/net/%s/speed", extInterface)

	if !mebroutines.ExistsFile(speedPath) {
		log.Debug(fmt.Sprintf("Network interface speed file '%s' does not exist", speedPath))
		return 0
	}

	/* #nosec */
	carrierFileContent, err := ioutil.ReadFile(carrierPath)
	if err != nil {
		log.Debug(fmt.Sprintf("Could not read network interface carrier from '%s': %v", carrierPath, err))
		return 0
	}
	if strings.TrimSpace(string(carrierFileContent)) == "0" {
		return 0 // Link not up
	}

	/* #nosec */
	speedFileContent, err := ioutil.ReadFile(speedPath)
	if err != nil {
		log.Debug(fmt.Sprintf("Could not read network interface speed from '%s': %v", speedPath, err))
		return 0
	}
	speedFileContentTrimmed := strings.TrimSpace(string(speedFileContent))
	speedInt, errConvert := strconv.ParseUint(speedFileContentTrimmed, 10, 64)
	if errConvert != nil {
		log.Debug(fmt.Sprintf("Could not convert speed string '%s' to integer: %v", speedFileContentTrimmed, errConvert))
		return 0
	}

	return speedInt * 1000000
}

func getExtInterfaceLegend(extInterface string) string {
	for _, table := range extNicNixLegendRules {
		matched, err := regexp.MatchString(table.RegExp, extInterface)
		if err == nil && matched {
			return table.Legend
		}
	}

	return "Unknown device"
}

// GetExtInterfaces returns map of network interfaces. It returns array of
// constants.AvailableSelection where the ConfigValue is system's internal value
// (e.g. "eno1") and Legend human-readable legend (e.g. "Wireless Network 802.11abc")
//
// Currently the Linux returns interface name and speed as a legend
func GetExtInterfaces() []constants.AvailableSelection {
	result := constants.DefaultExtNicArray

	interfaces, err := net.Interfaces()
	if err == nil {
		for n := range interfaces {
			if isIgnoredExtInterfaceLinux(interfaces[n].Name) {
				log.Debug(fmt.Sprintf("Ignoring external network interface '%s'", interfaces[n].Name))
			} else {
				log.Debug(fmt.Sprintf("Adding external network interface '%s' to the list of available devices", interfaces[n].Name))
				var oneInterface constants.AvailableSelection
				oneInterface.ConfigValue = interfaces[n].Name

				// Some day we might use lshw or similar to get more user-friendly description
				guessedLegend := getExtInterfaceLegend(interfaces[n].Name)

				speed := getExtInterfaceSpeed(interfaces[n].Name)
				if speed > 0 {
					oneInterface.Legend = fmt.Sprintf("%s %s (%s)", guessedLegend, interfaces[n].Name, humanize.SI(float64(speed), "bit/s"))
				} else {
					oneInterface.Legend = fmt.Sprintf("%s %s", guessedLegend, interfaces[n].Name)
				}

				result = append(result, oneInterface)
			}
		}
	}

	return result
}
