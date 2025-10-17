package genericaruba

import (
	"regexp"
	"strings"
)

// ParseShowVersion parses HP/Aruba show version output
func ParseShowVersion(output string) (platform, osVersion, model, serial, uptime string) {
	lines := strings.Split(output, "\n")

	platformRe := regexp.MustCompile(`(?i)(?:HP|ProCurve|Aruba)\s+([\w-]+)`)
	modelRe := regexp.MustCompile(`(?i)((?:HP|ProCurve|Aruba)?\s*[JK]?\d{4}[A-Z]?\s+[\w-]+)`)
	versionRe := regexp.MustCompile(`(?i)(?:revision|software revision|version)\s+([A-Z]{2}\.\d+\.\d+\.\d+)`)
	serialRe := regexp.MustCompile(`(?i)Serial\s+[Nn]umber\s*:?\s*([\w]+)`)
	uptimeRe := regexp.MustCompile(`(?i)(?:Up|uptime is)\s+(.+?)(?:\s*$)`)

	for _, line := range lines {
		if model == "" {
			if match := modelRe.FindStringSubmatch(line); match != nil {
				model = strings.TrimSpace(match[1])
				if platform == "" && strings.Contains(model, " ") {
					parts := strings.Fields(model)
					if len(parts) > 0 {
						platform = parts[0]
					}
				}
			}
		}

		if platform == "" {
			if match := platformRe.FindStringSubmatch(line); match != nil {
				platform = match[1]
			}
		}

		if osVersion == "" {
			if match := versionRe.FindStringSubmatch(line); match != nil {
				osVersion = match[1]
			}
		}

		if serial == "" {
			if match := serialRe.FindStringSubmatch(line); match != nil {
				serial = match[1]
			}
		}

		if uptime == "" {
			if match := uptimeRe.FindStringSubmatch(line); match != nil {
				uptime = strings.TrimSpace(match[1])
			}
		}
	}

	return
}
