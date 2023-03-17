package ui

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/arctir/proctor/plib"
)

const (
	port = ":8080"
)

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
                <td>{{.CommandName}}</td>
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

	log.Printf("serving at %s", port)
	panic(http.ListenAndServe(port, nil))
}
