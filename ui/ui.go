package ui

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/arctir/proctor/plib"
)

const (
	port          = ":8080"
	refreshPath   = "/refresh"
	processesPath = "/process/"
)

type UI struct {
	inspector plib.Inspector
	data      Data
}

type Data struct {
	LastRefresh time.Time
	PS          plib.Processes
}

type DetailKV struct {
	Field string
	Value string
}

func New() *UI {
	var err error
	newInspector, err := plib.NewInspector()
	newUI := UI{
		inspector: newInspector,
		data:      Data{},
	}
	if err != nil {
		panic(err)
	}
	return &newUI
}

func (ui *UI) RunUI() {
	http.HandleFunc("/", ui.handleAllProcesses)
	http.HandleFunc(refreshPath, ui.handleRefresh)
	http.HandleFunc(processesPath, ui.handleProcessDetails)

	log.Printf("serving at %s", port)
	panic(http.ListenAndServe(port, nil))
}

func (ui *UI) handleAllProcesses(w http.ResponseWriter, r *http.Request) {
	var err error
	ui.data.PS, err = ui.inspector.GetProcesses()
	ui.data.LastRefresh = ui.inspector.GetLastLoadTime()

	t, err := createTemplate(allProcessesView)
	if err != nil {
		// TODO(joshross): do error response
	}
	// Render the template with the data
	err = t.Execute(w, ui.data)
	if err != nil {
		writeFailure(w, err)
	}
}

func (ui *UI) handleRefresh(w http.ResponseWriter, r *http.Request) {
	err := ui.inspector.ClearProcessCache()
	if err != nil {
		panic(err)
	}
	log.Println("refreshed process cache")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (ui *UI) handleProcessDetails(w http.ResponseWriter, r *http.Request) {
	pidString := strings.TrimPrefix(r.URL.Path, processesPath)
	pid, err := strconv.Atoi(pidString)
	if err != nil {
		writeFailure(w, err)
	}

	// Render the template with the data
	if ui.data.PS[pid] == nil {
		writeFailure(w, fmt.Errorf("processes does not exist."))
		return
	}
	t, err := createTemplate(viewProcessDetails)
	if err != nil {
		writeFailure(w, err)
		return
	}
	err = t.Execute(w, ui.data.PS[pid])
	if err != nil {
		writeFailure(w, err)
		return
	}
}

// getProcessDetails returns a slice containing the key and value for each value
// property. It does this by performing reflection and understanding what's
// available on the [plib.Process].
func getProcessDetails(process plib.Process) []DetailKV {
	result := []DetailKV{}
	t := reflect.TypeOf(process)
	v := reflect.ValueOf(process)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == "OSSpecific" {
			continue
		}
		result = append(result, DetailKV{field.Name, fmt.Sprintf("%v", v.Field(i).Interface())})
	}
	t = reflect.TypeOf(process.OSSpecific)
	v = reflect.ValueOf(process.OSSpecific)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		result = append(result, DetailKV{field.Name, fmt.Sprintf("%v", v.Field(i).Interface())})
	}

	return result
}

// createTemplate returns a final template with your template (temp) specified
// and wrapped with [UIHeader] and [UIFooter].
func createTemplate(temp string) (*template.Template, error) {
	t, err := template.New("response").Funcs(template.FuncMap{"pDeets": getProcessDetails}).Parse(uiHeader + temp + uiFooter)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func writeFailure(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	t, _ := createTemplate(errorView)
	t.Execute(w, err.Error())
}
