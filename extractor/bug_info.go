package extractor

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
)

var (
	ErrInvalidVersionInterval = errors.New("invalid version interval")
	ErrInvalidSemver          = errors.New("invalid semver")
	ErrInvalidContent         = errors.New("invalid content")
	ErrVersionGap             = errors.New("missing some versions between affected-version & fixed-version")
	ErrFieldEmpty             = errors.New("field is empty")
)

var githubIssueCommentTemplate = regexp.MustCompile(`<!--(.|\s)*?-->`)
var versionTemplate = regexp.MustCompile(`\[?v?\d+\.\d+\.\d+|master|unplanned|unplaned\]?`)

// parse 3 kinds of inputs
// 1. [v4.0.1:v4.0.11] -> [$version $delimiter $version]
// 2. [:v4.0.11] -> [$delimiter $version]
// 3. v4.0.11 -> $version
// so the whole regexp is: [$version $delimiter $version] | [$delimiter $version] | $version
var versionIntervalTemplate = regexp.MustCompile(`\[v?(\d+\.\d+\.\d+)\s?(:|：|,|，)\s?v?(\d+\.\d+\.\d+)\]|\[\s?(:|：|,|，)\s?v?(\d+\.\d+\.\d+)\]|\[?v?(\d+\.\d+\.\d+|master|unreleased)\]?`)

type BugInfos struct {
	AllTriggerConditions string
	RCA                  string // Root Cause Analysis
	Symptom              string
	Workaround           string
	AffectedVersions     []string
	FixedVersions        []string
}

// getStringInBetween returns s[startIndex+len(startStr): endIndex+1], s[endIndex:]
func getStringInBetween(s string, startStr, endStr string) (string, string) {
	startIndex := strings.Index(s, startStr)
	if startIndex == -1 {
		return "", s
	}

	endIndex := strings.Index(s[startIndex:], endStr)
	if endIndex == -1 {
		endIndex = len(s)
	} else {
		endIndex += startIndex
	}

	newS := s[endIndex:]

	// strip empty lines
	s = s[startIndex+len(startStr) : endIndex]
	lines := strings.Split(s, "\n")
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			values = append(values, line)
		}
	}

	return strings.Join(values, "\n"), newS
}

// cleanupComment removes markdown comment strings from s
func cleanupComment(s string) string {
	return githubIssueCommentTemplate.ReplaceAllString(s, "")
}

var templates = []string{
	"#### 1. Root Cause Analysis (RCA) (optional)",
	"#### 2. Symptom (optional)",
	"#### 3. All Trigger Conditions (optional)",
	`#### 4. Workaround (optional)`,
	"#### 5. Affected versions",
	"#### 6. Fixed versions",
}

// requiredFields of BugInfos
// these fields should not be empty
// key of map should match field name of BugInfos
var requiredFields = map[string]struct{}{
	"AffectedVersions": {},
	"FixedVersions":    {},
}

var replaced = []string{",", "，", " ", "\t", "\n"}

func replace(src string, old []string, new string) string {
	for _, s := range old {
		src = strings.ReplaceAll(src, s, new)
	}

	return src
}

// ParseCommentBody extract BugInfos from githubCommentBody comment
func ParseCommentBody(githubCommentBody string) (*BugInfos, map[string][]error) {
	githubCommentBody = cleanupComment(githubCommentBody)

	info := &BugInfos{}
	errM := make(map[string][]error)
	var startStr, endStr string

	startStr = templates[0]
	endStr = templates[1]
	info.RCA, githubCommentBody = getStringInBetween(githubCommentBody, startStr, endStr)

	startStr = templates[1]
	endStr = templates[2]
	info.Symptom, githubCommentBody = getStringInBetween(githubCommentBody, startStr, endStr)

	startStr = templates[2]
	endStr = templates[3]
	info.AllTriggerConditions, githubCommentBody = getStringInBetween(githubCommentBody, startStr, endStr)

	startStr = templates[3]
	endStr = templates[4]
	info.Workaround, githubCommentBody = getStringInBetween(githubCommentBody, startStr, endStr)

	startStr = templates[4]
	endStr = templates[5]
	versions, githubCommentBody := getStringInBetween(githubCommentBody, startStr, endStr)
	expandedVersions, err := getAffectedVersions(versions)
	if err != nil {
		errM["AffectedVersions"] = append(errM["AffectedVersions"], err)
	} else {
		info.AffectedVersions = append(info.AffectedVersions, expandedVersions...)

		unwanted := versionIntervalTemplate.ReplaceAllString(versions, "")
		if len(replace(unwanted, replaced, "")) > 0 { // has value but didn't match by regexp
			err := fmt.Errorf("%w, got unexpected content: %s", ErrInvalidContent, unwanted)
			errM["AffectedVersions"] = append(errM["AffectedVersions"], err)
		}
	}

	startStr = templates[5]
	endStr = "****end****"
	versions, _ = getStringInBetween(githubCommentBody, startStr, endStr)
	info.FixedVersions = append(info.FixedVersions, versionTemplate.FindAllString(versions, -1)...)
	for i, v := range info.FixedVersions {
		if v == "unplanned" || v == "unplaned" {
			info.FixedVersions[i] = "master"
		}
	}

	unwanted := versionTemplate.ReplaceAllString(versions, "")
	if len(replace(unwanted, replaced, "")) > 0 { // besides delimeters and spaces, there is still unmatched content
		err := fmt.Errorf("%w, got unexpected content: %s", ErrInvalidContent, unwanted)
		errM["FixedVersions"] = append(errM["FixedVersions"], err)
	}

	// 1. if any field's length equals zero, append errM[$fieldname] with ErrFieldEmpty
	v := reflect.ValueOf(*info)
	for fieldName := range requiredFields {
		if v.FieldByName(fieldName).Len() == 0 {
			errM[fieldName] = append(errM[fieldName], ErrFieldEmpty)
		}
	}

	// 2. there should be no gap between affected-versions and fixed-versions
	if hasVersionGap(info) {
		errM["FixedVersions"] = append(errM["FixedVersions"], ErrVersionGap)
	}

	return info, errM
}

func hasVersionGap(info *BugInfos) bool {

	// make sure there is no gap between affected-versions and fix-versions
	// e.g. affect-version = [4.0.1, 4.0.2] fix-version = [4.0.4]
OUTER:
	for _, v := range info.FixedVersions {
		fixed, err := semver.NewVersion(v)
		if err != nil { // ignore "master"
			continue
		}

		fixed.Patch--
		shouldExist := fixed.String()
		for _, affected := range info.AffectedVersions {
			if shouldExist == affected {
				continue OUTER
			}
		}

		return true
	}

	return false
}

func stripEmpty(s []string) []string {
	result := make([]string, 0)
	for i := range s {
		if len(s[i]) > 0 {
			result = append(result, s[i])
		}
	}

	return result
}

func getAffectedVersions(version string) ([]string, error) {
	// this function is highly couple with bug template, kind of messy

	matches := versionIntervalTemplate.FindAllStringSubmatch(version, -1)
	result := make([]string, 0)

	for _, match := range matches {
		match = stripEmpty(match)

		switch len(match) {
		case 0:
			continue

		case 2: // e.g. v4.0.1
			if match[1] == "unreleased" {
				match[1] = "master"
			}
			result = append(result, match[1])

		case 3: // e.g. [:4.0.5] => [4.0.0:4.0.5]
			start, err := semver.NewVersion(match[2])
			if err != nil {
				return nil, ErrInvalidSemver
			}

			start.Patch = 0
			match = append(match[:1], append([]string{start.String()}, match[1:]...)...) // insert
			fallthrough

		case 4: // e.g. [4.0.0:4.0.5]
			start, err := semver.NewVersion(match[1])
			if err != nil {
				return nil, ErrInvalidSemver
			}

			end, err := semver.NewVersion(match[3])
			if err != nil {
				return nil, ErrInvalidSemver
			}

			if start.Major != end.Major ||
				start.Minor != end.Minor ||
				start.Patch > end.Patch {
				return nil, ErrInvalidVersionInterval
			}

			if end.Patch == 99 { // patch == 99 indicates this bug is no gonna be fixed
				result = append(result, fmt.Sprintf("%d.%d", start.Major, start.Minor))
			} else {
				result = append(result, expandVersion(start, end)...)
			}
		}
	}

	return result, nil
}

func expandVersion(start, end *semver.Version) []string {
	if start.Major != end.Major ||
		start.Minor != end.Minor {
		return nil
	}

	result := make([]string, 0, end.Patch-start.Patch+1)
	for start.LessThan(*end) {
		result = append(result, start.String())
		start.Patch++
	}
	result = append(result, end.String())

	start.Patch = end.Patch - int64(len(result)-1) // restore start.Patch
	return result
}

func ContainsBugTemplate(comment string) bool {
	idx := 0
	for _, tem := range templates {
		if idx = strings.Index(comment, tem); idx == -1 {
			return false
		}
		comment = comment[idx+len(tem):]
	}

	return true
}
