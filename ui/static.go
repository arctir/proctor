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
		.tree-wrapper {
			padding-top: 10px;
		  }
		  
		  .tree-list {
			list-style: none;
			padding: 0;
			margin: 0;
		  }
		  .tree-list .tree-item {
			position: relative;
			display: block;
			min-height: 2em;
			line-height: 2em;
			margin-bottom: 10px;
			padding-left: 21px;
		  }
		  .tree-list .tree-item:before, .tree-list .tree-item:after {
			content: "";
			position: absolute;
			display: block;
			background-color: #333;
		  }
		  .tree-list .tree-item:before {
			top: 0;
			left: 10px;
			width: 1px;
			height: calc(100% + 10px);
		  }
		  .tree-list .tree-item:after {
			top: 1em;
			left: 10px;
			width: 11px;
			height: 1px;
		  }
		  .tree-list .tree-item:last-child {
			margin-bottom: 0;
		  }
		  .tree-list .tree-item:last-child:before {
			height: 1em;
		  }
		  .tree-list .tree-item:first-child:before {
			top: -10px;
			height: calc(100% + 20px);
		  }
		  .tree-list .tree-item > span {
			display: inline-block;
			padding: 0 5px;
			border: 1px solid #333;
		  }
		  .tree-list .tree-item > .tree-list {
			padding-top: 10px;
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
		<div class="buttons">
			<a href="/"><button>All Processes</button></a>
			<a href="/tree/{{ .ID }}"><button>Process Hierarchy</button></a>
		</div>
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

const viewTreeDetails = `
		<div class="container">
		<div class="buttons">
			<a href="/"><button>All Processes</button></a>
		</div>
			<div class="tree-wrapper">

		  	    {{ range $value := . }}
				<ul class="tree-list">
					<li class="tree-item has-sub">
						<span><a href="/process/{{ .ID }}">{{ .CommandName }} ({{ .ID }})</a></span>
				{{ end }}
		  	    {{ range . }}
					</ul>
				</li>
				{{ end }}
			</div>
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
			<h1>Failed creating requested page.</h1>
			<p>Error details {{ . }}</p>
			</div>
		</div>
`
