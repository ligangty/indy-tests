package promotetest

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"

	common "github.com/commonjava/indy-tests/pkg/common"
)

// IndyPromoteVars ...
type IndyPromoteVars struct {
	TrackingId     string
	Source         string
	Target         string
	Paths          []string
	Async          bool
	PurgeSource    bool
	DryRun         bool
	FireEvents     bool
	FailWhenExists bool
}

func (promoteVars *IndyPromoteVars) fillDefaults() {
	promoteVars.Async = false
	promoteVars.DryRun = false
	promoteVars.PurgeSource = false
	promoteVars.FireEvents = true
	promoteVars.FailWhenExists = true
}

func createIndyPromoteVars(trackingId, source, target string, paths []string) IndyPromoteVars {
	promoteVars := &IndyPromoteVars{TrackingId: trackingId, Source: source, Target: target, Paths: paths}
	promoteVars.fillDefaults()
	return *promoteVars
}

// IndyPromoteJSONTemplate ...
func IndyPromoteJSONTemplate(indyPromoteVars *IndyPromoteVars) string {
	request := `{
  "trackingId": "{{.TrackingId}}",
  "async": {{.Async}},
  "source": "{{.Source}}",
  "target": "{{.Target}}",
  {{if gt (len .Paths) 0}}
  "paths": [{{range $index,$path := .Paths}}"{{$path}}"{{if isNotLast $index $.Paths}},{{end}}{{end}}],
  {{end}}
  "purgeSource": {{.PurgeSource}},
  "dryRun": {{.DryRun}},
  "fireEvents": {{.FireEvents}},
  "failWhenExists": {{.FailWhenExists}}
}`

	t := template.Must(template.New("promote_request").Funcs(isNotLast).Parse(request))
	var buf bytes.Buffer
	err := t.Execute(&buf, indyPromoteVars)
	if err != nil {
		log.Fatal("executing template:", err)
	}

	return buf.String()
}

var isNotLast = template.FuncMap{
	"isNotLast": func(index int, array []string) bool {
		return index < len(array)-1
	},
}

func promote(indyURL, trackingId, source, target string, paths []string, dryRun bool) (string, int, bool) {
	promoteVars := IndyPromoteVars{
		TrackingId: trackingId,
		Source: source,
		Target: target,
		Paths:  paths,
	}
	promote := IndyPromoteJSONTemplate(&promoteVars)

	URL := fmt.Sprintf("%s/api/promotion/paths/promote", indyURL)

	if dryRun {
		fmt.Printf("Dry run promote request:\n %s\n\n", promote)
		return "", 200, true
	}

	fmt.Printf("Start promote request:\n %s\n\n", promote)
	respText, code, result := common.HTTPRequest(URL, common.MethodPost, nil, true, strings.NewReader(promote), nil, "", false)

	if result {
		fmt.Printf("Promote Done. Result is:\n %s\n\n", respText)
	} else {
		fmt.Printf("Promote Error. Result is:\n %s\n\n", respText)
	}

	return respText, code, result
}

func Rollback(indyURL, promoteResult string, dryRun bool) (string, int, bool) {
	URL := fmt.Sprintf("%s/api/promotion/paths/rollback", indyURL)

	if dryRun {
		fmt.Printf("Dry run rollback request:\n %s\n\n", URL)
		return "", 200, true
	}

	fmt.Printf("Start rollback request: %s\n", URL)
	respText, code, result := common.HTTPRequest(URL, common.MethodPost, nil, true, strings.NewReader(promoteResult), nil, "", false)

	if result {
		fmt.Printf("Rollback Done. code:%d, respText:\n%s\n", code, respText)
	} else {
		fmt.Printf("Rollback Error. code:%d, respText:\n%s\n", code, respText)
	}

	return respText, code, result
}
