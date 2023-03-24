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
	port = ":8080"
)

const viewProcessDetails = `
<html>
	<head>

	<style>
		.buttons {
			margin-bottom: 1rem;
		}
		button {
			background-color: black;
			color: white;
			border: 1px solid black;
			padding: 8px;
			font-size: 16px;
			cursor: pointer;
		}
		table {
			border-collapse: collapse;
			width: 100%;
		}
		th, td {
			border: 1px solid black;
			padding: 8px;
			text-align: left;
		}
		th {
			background-color: black;
			color: white;
		}
	</style>
		<title>Procotor display</title>
	</head>
	<body>
		<div class="container">
		<table>
            <tr>
                <th>Command Name</th>
                <th>Command Path</th>
            </tr>
            <tr>
                <td>{{ .CommandName }}</td>
                <td>{{ .CommandPath }}</td>
            </tr>
			</table>
			{{ . | pDeets }}
		</div>
	</body>
</html>
`

const view = `
<html>
	<head>

	<style>
		.buttons {
			margin-bottom: 1rem;
		}
		button {
			background-color: black;
			color: white;
			border: 1px solid black;
			padding: 8px;
			font-size: 16px;
			cursor: pointer;
		}
		table {
			border-collapse: collapse;
			width: 100%;
		}
		th, td {
			border: 1px solid black;
			padding: 8px;
			text-align: left;
		}
		th {
			background-color: black;
			color: white;
		}
	</style>
		<title>Procotor display</title>
	</head>
	<body>
		<div class="container">
		<div class="status">
		 <p>Last Refreshed: {{ .LastRefresh }}</p>
		</div>
		<div class="buttons">
			<a href="/refresh"><button>Refresh</button></a>
		</div>
		<table>
            <tr>
                <th>PID</th>
                <th>Name</th>
                <th>SHA</th>
            </tr>
			{{range $key, $value := .PS}}
            <tr>
                <td>{{$key}}</td>
				<td><a href="process/{{$key}}">{{.CommandName}}</a></td>
                <td>{{.BinarySHA}}</td>
            </tr>
            {{end}}
			</table>
		</div>
	</body>
</html>
`

type UI struct {
	inspector plib.Inspector
	data      Data
}

type Data struct {
	LastRefresh time.Time
	PS          plib.Processes
}

func New() *UI {
	var err error
	newInspector, err := plib.NewInspector()
	newUI := UI{
		inspector: newInspector,
	}
	if err != nil {
		panic(err)
	}
	return &newUI
}

func (ui *UI) RunUI() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error
		ui.data.PS, err = ui.inspector.GetProcesses()
		ui.data.LastRefresh = ui.inspector.GetLastLoadTime()

		t := template.Must(template.New("map").Parse(view))
		// Render the template with the data
		err = t.Execute(w, ui.data)
		if err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		err := ui.inspector.ClearProcessCache()
		if err != nil {
			panic(err)
		}
		log.Println("refreshed process cache")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/process/", func(w http.ResponseWriter, r *http.Request) {
		pidString := strings.TrimPrefix(r.URL.Path, "/process/")
		pid, err := strconv.Atoi(pidString)
		if err != nil {
			panic(err)
		}

		t := template.Must(template.New("map").Funcs(template.FuncMap{"pDeets": getProcessDetails}).Parse(viewProcessDetails))
		// Render the template with the data
		err = t.Execute(w, ui.data.PS[pid])
		if err != nil {
			panic(err)
		}
	})

	log.Printf("serving at %s", port)
	panic(http.ListenAndServe(port, nil))
}

func getProcessDetails(process plib.Process) string {
	var details string
	t := reflect.TypeOf(process)
	v := reflect.ValueOf(process)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		details += fmt.Sprintf("%s: %v", field.Name, v.Field(i).Interface())
	}
	fmt.Println(details)
	return details
}
