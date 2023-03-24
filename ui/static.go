package ui

const uiHeader = `
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
`

const uiFooter = `
	</body>
</html>
`

const viewProcessDetails = `
		<div class="container">
		<table>
            <tr>
                <th>Field</th>
                <th>Value</th>
            </tr>
			{{range $idx, $value := . | pDeets }}
            <tr>
                <td>{{ $value.Field }}</td>
                <td>{{ $value.Value }}</td>
            </tr>
			{{ end }}
			</table>
		</div>
`

const allProcessesView = `
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
`

const errorView = `
		<div class="container">
			<div class="status">
			<h1>Failed retrieving page! 404.</h1>
			<p>Error details {{ . }}</p>
			</div>
		</div>
`
